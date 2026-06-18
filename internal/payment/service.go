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
	"liki/internal/dodo"
	
)

// ReportGenerator generates full LLM reports from computed chart data.
type ReportGenerator interface {
	GenerateFromData(ctx context.Context, locale string, product agent.Product, chartJSON json.RawMessage) (string, error)
}

var (
	ErrOrderNotFound = errors.New("payment: order not found")
	ErrOrderNotPaid  = errors.New("payment: order not paid")
	ErrWebhookVerify = errors.New("payment: webhook verification failed")
)

type dodoClient interface {
	CreateCheckout(ctx context.Context, productID string, amount int, orderID, email, returnURL string) (*dodo.CheckoutResult, error)
	VerifyWebhook(rawBody []byte, headers http.Header) (*dodo.WebhookEvent, error)
}

type emailClient interface {
	SendReport(ctx context.Context, to, subject, htmlBody string) error
}

// Service handles the full payment lifecycle: checkout creation, webhook processing,
// background report generation, and report retrieval.
type Service struct {
	Dodo        dodoClient
	Email       emailClient
	Store       *Store
	ProductIDs  map[agent.Product]string
	ReturnURL   string
	AdminEmail  string
	ReportAgent ReportGenerator

	bgCtx        context.Context
	generatingMu sync.Mutex
	generating   map[string]struct{}
}

// NewService creates a payment service with the given dependencies.
func NewService(dodo dodoClient, email emailClient, store *Store, productIDs map[agent.Product]string, returnURL, adminEmail string, reportAgent ReportGenerator, bgCtx context.Context) *Service {
	return &Service{
		Dodo:        dodo,
		Email:       email,
		Store:       store,
		ProductIDs:  productIDs,
		ReturnURL:   returnURL,
		AdminEmail:  adminEmail,
		ReportAgent: reportAgent,
		bgCtx:       bgCtx,
		generating:  make(map[string]struct{}),
	}
}

// CreateCheckout creates a Dodo Payments checkout session for an existing order.
func (s *Service) CreateCheckout(ctx context.Context, orderID, userEmail string) (*dodo.CheckoutResult, error) {
	order, err := s.Store.GetOrder(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrOrderNotFound, err)
	}

	productID, ok := s.ProductIDs[order.Product]
	if !ok {
		return nil, fmt.Errorf("payment: no product for %s", order.Product)
	}

	if userEmail != "" {
		if err := s.Store.UpdateEmail(ctx, orderID, userEmail); err != nil {
			return nil, fmt.Errorf("payment: update email: %w", err)
		}
	}

	returnURL := s.ReturnURL + "/api/payments/return/" + orderID
	result, err := s.Dodo.CreateCheckout(ctx, productID, order.Amount, orderID, userEmail, returnURL)
	if err != nil {
		return nil, fmt.Errorf("payment: dodo checkout: %w", err)
	}
	return result, nil
}

// HandleWebhook processes a Dodo Payments webhook event and triggers report generation.
func (s *Service) HandleWebhook(ctx context.Context, body []byte, headers http.Header) error {
	event, err := s.Dodo.VerifyWebhook(body, headers)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrWebhookVerify, err)
	}

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
		// Report generation runs in background (one-shot, no retry).
		// If it fails, the report page triggers lazy generation on visit.
		if s.ReportAgent != nil {
			s.StartReportGeneration(orderID, product, chartJSON)
		}

		// Emails run in background. Retry is handled by the email client.
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

	content, err := s.ReportAgent.GenerateFromData(ctx, locale, product, json.RawMessage(chartJSON))
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
	Product   agent.Product `json:"product"`
	Status    OrderStatus    `json:"status"`
	Email     string         `json:"email,omitempty"`
	ChartJSON string         `json:"chart_json"`
	LlmJSON   string         `json:"llm_json"`
}

// OrderStatus returns the status and product of an order.
// OrderStatus returns the payment status and product type of an order.
func (s *Service) OrderStatus(ctx context.Context, orderID string) (status OrderStatus, product agent.Product, err error) {
	order, err := s.Store.GetOrder(ctx, orderID)
	if err != nil {
		return "", "", err
	}
	return order.Status, order.Product, nil
}

// RetryReportGeneration checks order status and triggers LLM report generation
// if the order is paid but has no cached llm_json (missed webhook recovery).
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
