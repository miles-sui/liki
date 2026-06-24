package payment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"liki/internal/agent"
)

var (
	ErrOrderNotFound   = errors.New("payment: order not found")
	ErrOrderNotPaid    = errors.New("payment: order not paid")
	ErrWebhookVerify   = errors.New("payment: webhook verification failed")
	ErrUnknownProvider = errors.New("payment: unknown provider")
)

// CheckoutResult holds the result of creating a checkout session.
type CheckoutResult struct {
	SessionID   string `json:"session_id"`
	CheckoutURL string `json:"checkout_url"`
	QRCodeURL   string `json:"qrcode_url,omitempty"`
}

// WebhookEvent is a verified webhook event from a payment provider.
type WebhookEvent struct {
	Type string
	Data WebhookEventData
}

// WebhookEventData holds the extracted fields from a webhook event.
type WebhookEventData struct {
	OrderID   string
	Amount    int
	Email     string
	PaymentID string
}

// paymentProvider abstracts a payment gateway (Dodo, Xunhu, etc.).
type paymentProvider interface {
	CreateCheckout(ctx context.Context, product agent.Product, amount int, orderID, email, returnURL string) (*CheckoutResult, error)
	VerifyWebhook(rawBody []byte, headers http.Header) (*WebhookEvent, error)
}

type emailClient interface {
	SendReport(ctx context.Context, to, subject, htmlBody string) error
}

// Service handles the full payment lifecycle: checkout creation, webhook processing,
// background report generation, and report retrieval.
type Service struct {
	Dodo         paymentProvider
	Xunhu        paymentProvider
	Email        emailClient
	Store        *Store
	ReturnURL    string
	AdminEmail   string
	ReportAgents map[agent.Product]*agent.ReportAgent

	bgCtx        context.Context
	generatingMu sync.Mutex
	generating   map[string]struct{}
}

// NewService creates a payment service with the given dependencies.
func NewService(dodo, xunhu paymentProvider, email emailClient, store *Store, returnURL, adminEmail string, reportAgents map[agent.Product]*agent.ReportAgent, bgCtx context.Context) *Service {
	return &Service{
		Dodo:         dodo,
		Xunhu:        xunhu,
		Email:        email,
		Store:        store,
		ReturnURL:    returnURL,
		AdminEmail:   adminEmail,
		ReportAgents: reportAgents,
		bgCtx:        bgCtx,
		generating:   make(map[string]struct{}),
	}
}

// CreateCheckout creates a checkout session for an existing order via the given provider.
func (s *Service) CreateCheckout(ctx context.Context, provider, orderID, userEmail string) (*CheckoutResult, error) {
	order, err := s.Store.GetOrder(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrOrderNotFound, err)
	}

	p, err := s.provider(provider)
	if err != nil {
		return nil, err
	}

	if userEmail != "" {
		if err := s.Store.UpdateEmail(ctx, orderID, userEmail); err != nil {
			return nil, fmt.Errorf("payment: update email: %w", err)
		}
	}
	if err := s.Store.UpdateProvider(ctx, orderID, provider); err != nil {
		return nil, fmt.Errorf("payment: update provider: %w", err)
	}

	returnURL := s.ReturnURL + "/api/payments/return/" + orderID
	result, err := p.CreateCheckout(ctx, order.Product, order.Amount, orderID, userEmail, returnURL)
	if err != nil {
		return nil, fmt.Errorf("payment: %s checkout: %w", provider, err)
	}
	return result, nil
}

// provider returns the payment provider for the given name.
func (s *Service) provider(name string) (paymentProvider, error) {
	switch name {
	case "dodo":
		return s.Dodo, nil
	case "xunhu":
		return s.Xunhu, nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownProvider, name)
	}
}

// HandleWebhook processes a payment webhook event from any provider.
func (s *Service) HandleWebhook(ctx context.Context, body []byte, headers http.Header) error {
	var lastErr error
	for _, p := range []paymentProvider{s.Dodo, s.Xunhu} {
		event, err := p.VerifyWebhook(body, headers)
		if err != nil {
			lastErr = err
			continue
		}
		if event == nil {
			lastErr = fmt.Errorf("nil event from provider")
			continue
		}
		return s.handleEvent(ctx, event)
	}
	return fmt.Errorf("%w: %w", ErrWebhookVerify, lastErr)
}

func (s *Service) handleEvent(ctx context.Context, event *WebhookEvent) error {
	if event.Type != "payment.succeeded" {
		level := slog.Info
		if event.Type == "payment.refunded" || event.Type == "payment.disputed" {
			level = slog.Error
		}
		level("payment: non-payment webhook event", "type", event.Type, "order_id", event.Data.OrderID)
		return nil
	}

	orderID := event.Data.OrderID
	if orderID == "" {
		slog.Error("payment: webhook with empty order_id", "payment_id", event.Data.PaymentID)
		return fmt.Errorf("payment: empty order_id in webhook event")
	}
	newPayment, email, product, chartJSON, err := s.Store.MarkPaidIdempotent(ctx, orderID, event.Data.PaymentID)
	if err != nil {
		return fmt.Errorf("payment: mark paid: %w", err)
	}

	if newPayment {
		if _, ok := s.ReportAgents[product]; ok {
			s.StartReportGeneration(orderID, product, chartJSON)
		}

		if email != "" {
			customerHTML := fmt.Sprintf(`<p>感谢购买！<a href="%s/report/%s">点击查看完整报告</a></p>
				<p>请保存此链接以便日后查阅。如有疑问请回复此邮件。</p>`, s.ReturnURL, orderID)
			go func() { if err := s.Email.SendReport(s.bgCtx, email, product.EmailSubject(), customerHTML); err != nil { slog.Error("send customer report", "err", err) } }()
		}

		if s.AdminEmail != "" {
			adminHTML := fmt.Sprintf(
				`<p>新订单 <strong>%s</strong> | 产品: %s | 金额: ¥%.2f | 用户: %s</p>
				<p><a href="%s/report/%s">查看报告</a></p>`,
				orderID, product, float64(event.Data.Amount)/100, email, s.ReturnURL, orderID,
			)
			go func() {
				if err := s.Email.SendReport(s.bgCtx, s.AdminEmail,
					fmt.Sprintf("[灵机] %s · %s", product, orderID), adminHTML); err != nil {
					slog.Error("send admin report", "err", err)
				}
			}()
		}
	}

	return nil
}

// StartReportGeneration starts background LLM report generation if not already in progress.
func (s *Service) StartReportGeneration(orderID string, product agent.Product, chartJSON string) {
	s.generatingMu.Lock()
	if _, ok := s.generating[orderID]; ok {
		s.generatingMu.Unlock()
		return
	}
	s.generating[orderID] = struct{}{}
	s.generatingMu.Unlock()

	go s.generateFullReport(orderID, product, chartJSON)
}

// generateFullReport runs GenerateFromData in background and caches the result.
func (s *Service) generateFullReport(orderID string, product agent.Product, chartJSON string) {
	defer func() {
		s.generatingMu.Lock()
		delete(s.generating, orderID)
		s.generatingMu.Unlock()
	}()
	defer func() {
		if r := recover(); r != nil {
			slog.Error("payment: panic in generateFullReport", "orderID", orderID, "panic", r)
		}
	}()
	ctx, cancel := context.WithTimeout(s.bgCtx, 120*time.Second)
	defer cancel()

	order, err := s.Store.GetOrder(ctx, orderID)
	var locale string
	if err != nil {
		slog.Error("payment: get order for report generation", "orderID", orderID, "err", err)
	} else {
		locale = order.Locale
	}
	if locale == "" {
		locale = "zh-Hans"
	}

	ra, ok := s.ReportAgents[product]
	if !ok {
		slog.Error("payment: no report agent for product", "product", product)
		return
	}
	content, err := ra.Generate(ctx, locale, json.RawMessage(chartJSON), nil)
	if err != nil {
		slog.Error("payment: generate full report", "orderID", orderID, "err", err)
		return
	}

	if updated, err := s.Store.UpdateLlmJSONIfEmpty(ctx, orderID, content); err != nil {
		slog.Error("payment: cache full report", "orderID", orderID, "err", err)
	} else if updated {
		slog.Info("payment: cached full report", "orderID", orderID)
	}
}

// ReportData holds the full report data for a paid order.
type ReportData struct {
	OrderID   string         `json:"order_id"`
	Product   agent.Product  `json:"product"`
	Status    OrderStatus    `json:"status"`
	Email     string         `json:"email,omitempty"`
	ChartJSON string         `json:"chart_json"`
	LlmJSON   string         `json:"llm_json"`
}

// OrderStatus returns the payment status and product type of an order.
func (s *Service) OrderStatus(ctx context.Context, orderID string) (status OrderStatus, product agent.Product, err error) {
	order, err := s.Store.GetOrder(ctx, orderID)
	if err != nil {
		return "", "", err
	}
	return order.Status, order.Product, nil
}

// RetryReportGeneration triggers report generation for paid orders missing their LLM report.
func (s *Service) RetryReportGeneration(ctx context.Context, orderID string) (OrderStatus, agent.Product, string, error) {
	order, err := s.Store.GetOrder(ctx, orderID)
	if err != nil {
		return "", "", "", err
	}
	if order.Status == OrderPaid && order.LlmJSON == "" {
		s.StartReportGeneration(orderID, order.Product, order.ChartJSON)
	}
	return order.Status, order.Product, order.LlmJSON, nil
}

// GetReport returns the full report data for an order.
func (s *Service) GetReport(ctx context.Context, orderID string) (*ReportData, error) {
	order, err := s.Store.GetOrder(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrOrderNotFound, err)
	}

	rd := &ReportData{
		OrderID:   order.OrderID,
		Product:   order.Product,
		Status:    order.Status,
		Email:     order.Email,
		ChartJSON: order.ChartJSON,
	}

	if order.Status == OrderPaid {
		rd.LlmJSON = order.LlmJSON
	}

	return rd, nil
}
