package payment

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"liki/internal/product"
)

// Webhook event types from payment providers.
const (
	EventPaymentSucceeded = "payment.succeeded"
	EventPaymentRefunded  = "payment.refunded"
	EventPaymentDisputed  = "payment.disputed"
)

// Provider names.
const (
	ProviderDodo  = "dodo"
	ProviderXunhu = "xunhu"
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
	CreateCheckout(ctx context.Context, product product.Product, amount int, orderID, email, returnURL string) (*CheckoutResult, error)
	VerifyWebhook(rawBody []byte, headers http.Header) (*WebhookEvent, error)
}

type emailClient interface {
	SendReport(ctx context.Context, to, subject, htmlBody string) error
}

// Service handles the full payment lifecycle: checkout creation, webhook processing,
// and report retrieval.
type Service struct {
	dodo       paymentProvider
	xunhu      paymentProvider
	email      emailClient
	Store      *Store
	returnURL  string
	adminEmail string

	bgCtx context.Context
	wg    sync.WaitGroup
}

// NewService creates a payment service with the given dependencies.
func NewService(dodo, xunhu paymentProvider, email emailClient, store *Store, returnURL, adminEmail string, bgCtx context.Context) *Service {
	return &Service{
		dodo:       dodo,
		xunhu:      xunhu,
		email:      email,
		Store:      store,
		returnURL:  returnURL,
		adminEmail: adminEmail,
		bgCtx:      bgCtx,
	}
}

// Shutdown waits for in-flight background goroutines (e.g. email sends) to finish,
// or returns when ctx is cancelled. Derive from the shutdown deadline so a stuck
// SMTP connection doesn't block forever.
func (s *Service) Shutdown(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// CreateCheckout creates a checkout session for an existing order via the given provider.
func (s *Service) CreateCheckout(ctx context.Context, provider, orderID, userEmail string) (*CheckoutResult, error) {
	order, err := s.Store.GetOrder(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("payment: checkout: %w", err)
	}

	p, err := s.provider(provider)
	if err != nil {
		return nil, err
	}

	if userEmail != "" {
		if err := s.Store.UpdateEmail(ctx, orderID, userEmail); err != nil {
			return nil, fmt.Errorf("payment: update email: %w", err)
		}
	} else {
		userEmail = order.Email
	}
	if err := s.Store.UpdateProvider(ctx, orderID, provider); err != nil {
		return nil, fmt.Errorf("payment: update provider: %w", err)
	}

	returnURL := s.returnURL + "/api/payments/return/" + orderID
	result, err := p.CreateCheckout(ctx, order.Product, order.Amount, orderID, userEmail, returnURL)
	if err != nil {
		return nil, fmt.Errorf("payment: %s checkout: %w", provider, err)
	}
	return result, nil
}

// provider returns the payment provider for the given name.
func (s *Service) provider(name string) (paymentProvider, error) {
	switch name {
	case ProviderDodo:
		return s.dodo, nil
	case ProviderXunhu:
		return s.xunhu, nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownProvider, name)
	}
}

// HandleWebhook processes a payment webhook event from any provider.
func (s *Service) HandleWebhook(ctx context.Context, body []byte, headers http.Header) error {
	var errs []error
	for _, p := range []paymentProvider{s.dodo, s.xunhu} {
		event, err := p.VerifyWebhook(body, headers)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if event == nil {
			errs = append(errs, fmt.Errorf("nil event from provider"))
			continue
		}
		return s.handleEvent(ctx, event)
	}
	return fmt.Errorf("payment: webhook verify: %w: %w", ErrWebhookVerify, errors.Join(errs...))
}

func (s *Service) handleEvent(ctx context.Context, event *WebhookEvent) error {
	if event.Type != EventPaymentSucceeded {
		level := slog.Info
		if event.Type == EventPaymentRefunded || event.Type == EventPaymentDisputed {
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
	newPayment, email, product, err := s.Store.MarkPaidIdempotent(ctx, orderID, event.Data.PaymentID)
	if err != nil {
		return fmt.Errorf("payment: mark paid: %w", err)
	}

	if newPayment {
		if email != "" {
			customerHTML := fmt.Sprintf(`<p>感谢购买！<a href="%s/chat">点击开始起名</a></p>
				<p>7 天内可随时回来继续磋商，输入此邮箱即可恢复对话。</p>
				<p>如有疑问请回复此邮件。</p>`, s.returnURL)
			s.wg.Add(1)
			go func() {
				defer s.wg.Done()
				if err := s.email.SendReport(s.bgCtx, email, product.EmailSubject(), customerHTML); err != nil {
					slog.Error("send customer report", "err", err)
				}
			}()
		}

		if s.adminEmail != "" {
			adminHTML := fmt.Sprintf(
				`<p>新订单 <strong>%s</strong> | %s</p>
				<p>产品: %s | 金额: ¥%.2f | 用户: %s</p>
				<p><a href="%s/report/%s">查看报告</a></p>`,
				orderID, time.Now().UTC().Format(time.DateTime),
				product, float64(event.Data.Amount)/100, email,
				s.returnURL, orderID,
			)
			s.wg.Add(1)
			go func() {
				defer s.wg.Done()
				if err := s.email.SendReport(s.bgCtx, s.adminEmail,
					fmt.Sprintf("[灵机] %s · %s", product, orderID), adminHTML); err != nil {
					slog.Error("send admin report", "err", err)
				}
			}()
		}
	}

	return nil
}

// ReportData holds the full report data for a paid order.
type ReportData struct {
	OrderID   string         `json:"order_id"`
	Product   product.Product  `json:"product"`
	Status    OrderStatus    `json:"status"`
	Email     string         `json:"email,omitempty"`
	ChartJSON string         `json:"chart_json"`
	LlmJSON   string         `json:"llm_json"`
}

// GetOrderData returns order status, product, and llm_json for the retry endpoint.
func (s *Service) GetOrderData(ctx context.Context, orderID string) (OrderStatus, product.Product, string, error) {
	order, err := s.Store.GetOrder(ctx, orderID)
	if err != nil {
		return "", "", "", err
	}
	return order.Status, order.Product, order.LlmJSON, nil
}

// GetReport returns the full report data for an order.
func (s *Service) GetReport(ctx context.Context, orderID string) (*ReportData, error) {
	order, err := s.Store.GetOrder(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("payment: get report: %w", err)
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
