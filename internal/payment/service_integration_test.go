//go:build integration

package payment

import (
	"context"
	"errors"
	"net/http"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"liki/internal/agent"
)

// trackedEmailClient records every SendReport call for inspection.
type trackedEmailClient struct {
	mu      sync.Mutex
	reports []emailReport
	err     error
}

type emailReport struct {
	To      string
	Subject string
}

func (m *trackedEmailClient) SendReport(_ context.Context, to, subject, _ string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.reports = append(m.reports, emailReport{To: to, Subject: subject})
	return m.err
}

func (m *trackedEmailClient) count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.reports)
}

type stubProvider struct {
	checkoutResult *CheckoutResult
	checkoutErr    error
	webhookEvent   *WebhookEvent
	webhookErr     error
}

func (p *stubProvider) CreateCheckout(_ context.Context, _ agent.Product, _ int, _, _, _ string) (*CheckoutResult, error) {
	return p.checkoutResult, p.checkoutErr
}

func (p *stubProvider) VerifyWebhook(_ []byte, _ http.Header) (*WebhookEvent, error) {
	return p.webhookEvent, p.webhookErr
}

func openServiceTestDB(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "service-test.db")
	db, err := OpenDB(path)
	if err != nil {
		t.Fatalf("OpenDB: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	s, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	return s
}

func newTestSvcInt(t *testing.T, dodo, xunhu paymentProvider, email emailClient) *Service {
	t.Helper()
	return NewService(dodo, xunhu, email, openServiceTestDB(t), "https://example.com", "admin@example.com", nil, context.Background())
}

// ── Webhook → MarkPaid → Email ──

func TestWebhookToReportPipeline(t *testing.T) {
	dodoCli := &stubProvider{
		webhookEvent: &WebhookEvent{
			Type: "payment.succeeded",
			Data: WebhookEventData{
				OrderID:   "pipe-1",
				Amount:    990,
				PaymentID: "pay-pipe-1",
			},
		},
	}
	emailCli := &trackedEmailClient{}
	svc := newTestSvcInt(t, dodoCli, &stubProvider{}, emailCli)

	ctx := context.Background()

	if err := svc.Store.CreateOrder(ctx, "pipe-1", agent.ProductChart, 990, "CNY", `{"chart":{"nianzhu":{"gan":"甲","zhi":"子"}}}`, "", "zh-Hans", ""); err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}
	svc.Store.UpdateEmail(ctx, "pipe-1", "buyer@test.com")

	err := svc.HandleWebhook(ctx, []byte(`{"type":"payment.succeeded"}`), http.Header{})
	if err != nil {
		t.Fatalf("HandleWebhook: %v", err)
	}

	o, err := svc.Store.GetOrder(ctx, "pipe-1")
	if err != nil {
		t.Fatalf("GetOrder: %v", err)
	}
	if o.Status != OrderPaid {
		t.Errorf("Status = %q, want paid", o.Status)
	}
	if o.PaymentID != "pay-pipe-1" {
		t.Errorf("PaymentID = %q, want pay-pipe-1", o.PaymentID)
	}

	// Email sent to buyer (goroutine — poll).
	for i := 0; i < 20; i++ {
		if emailCli.count() >= 1 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if emailCli.count() < 1 {
		t.Fatal("expected at least 1 email (buyer)")
	}
	foundBuyer := false
	for _, r := range emailCli.reports {
		if r.To == "buyer@test.com" {
			foundBuyer = true
		}
	}
	if !foundBuyer {
		t.Errorf("expected buyer email to buyer@test.com, got %v", emailCli.reports)
	}
}

// ── Webhook idempotency — no double email ──

func TestWebhookToReport_Idempotent(t *testing.T) {
	dodoCli := &stubProvider{
		webhookEvent: &WebhookEvent{
			Type: "payment.succeeded",
			Data: WebhookEventData{
				OrderID:   "idem-1",
				Amount:    990,
				PaymentID: "pay-idem-1",
			},
		},
	}
	emailCli := &trackedEmailClient{}
	svc := newTestSvcInt(t, dodoCli, &stubProvider{}, emailCli)

	ctx := context.Background()
	svc.Store.CreateOrder(ctx, "idem-1", agent.ProductChart, 990, "CNY", `{}`, "", "zh-Hans", "")
	svc.Store.UpdateEmail(ctx, "idem-1", "buyer@test.com")

	if err := svc.HandleWebhook(ctx, []byte(`{"type":"payment.succeeded"}`), http.Header{}); err != nil {
		t.Fatalf("first HandleWebhook: %v", err)
	}

	// Allow goroutines to complete.
	time.Sleep(100 * time.Millisecond)
	emailCount := emailCli.count()

	// Second webhook — same PaymentID.
	if err := svc.HandleWebhook(ctx, []byte(`{"type":"payment.succeeded"}`), http.Header{}); err != nil {
		t.Fatalf("second HandleWebhook: %v", err)
	}

	if emailCli.count() != emailCount {
		t.Errorf("email count = %d, want %d (no new emails)", emailCli.count(), emailCount)
	}
}

// ── Webhook with non-existent order returns error ──

func TestWebhook_NonExistentOrder(t *testing.T) {
	dodoCli := &stubProvider{
		webhookEvent: &WebhookEvent{
			Type: "payment.succeeded",
			Data: WebhookEventData{
				OrderID:   "no-such-order",
				Amount:    990,
				PaymentID: "pay-unknown",
			},
		},
	}
	svc := newTestSvcInt(t, dodoCli, &stubProvider{}, &trackedEmailClient{})

	err := svc.HandleWebhook(context.Background(), []byte(`{}`), http.Header{})
	if err == nil {
		t.Fatal("expected error for non-existent order")
	}
	if !errors.Is(err, ErrOrderNotFound) {
		t.Errorf("error = %v, want ErrOrderNotFound", err)
	}
}

// ── Checkout flow sets provider ──

func TestCreateCheckout_Integration(t *testing.T) {
	dodoCli := &stubProvider{
		checkoutResult: &CheckoutResult{
			SessionID:   "sess-checkout",
			CheckoutURL: "https://pay.example.com/checkout",
		},
	}
	svc := newTestSvcInt(t, dodoCli, &stubProvider{}, &trackedEmailClient{})
	ctx := context.Background()

	svc.Store.CreateOrder(ctx, "co-1", agent.ProductBond, 1990, "CNY", `{}`, "", "zh-Hans", "")

	result, err := svc.CreateCheckout(ctx, "dodo", "co-1", "")
	if err != nil {
		t.Fatalf("CreateCheckout: %v", err)
	}
	if result.SessionID != "sess-checkout" {
		t.Errorf("SessionID = %q, want sess-checkout", result.SessionID)
	}

	o, _ := svc.Store.GetOrder(ctx, "co-1")
	if o.Provider != "dodo" {
		t.Errorf("Provider = %q, want dodo", o.Provider)
	}
}

// ── RetryReportGeneration for paid orders without llm_json ──

func TestRetryReportGeneration_Integration(t *testing.T) {
	svc := newTestSvcInt(t, &stubProvider{}, &stubProvider{}, &trackedEmailClient{})
	ctx := context.Background()

	svc.Store.CreateOrder(ctx, "retry-1", agent.ProductChart, 990, "CNY", `{"chart":{}}`, "", "zh-Hans", "")
	svc.Store.MarkPaidIdempotent(ctx, "retry-1", "pay-retry-1")

	status, product, llmJSON, err := svc.RetryReportGeneration(ctx, "retry-1")
	if err != nil {
		t.Fatalf("RetryReportGeneration: %v", err)
	}
	if status != OrderPaid {
		t.Errorf("status = %q, want paid", status)
	}
	if product != agent.ProductChart {
		t.Errorf("product = %q, want chart", product)
	}
	if llmJSON != "" {
		t.Errorf("llmJSON = %q, want empty", llmJSON)
	}
}

// ── GetReport hides llm_json for pending orders ──

func TestGetReport_HidesLlmJSONForPending(t *testing.T) {
	svc := newTestSvcInt(t, &stubProvider{}, &stubProvider{}, &trackedEmailClient{})
	ctx := context.Background()

	svc.Store.CreateOrder(ctx, "pending-rpt", agent.ProductNaming, 2990, "CNY", `{}`, "# secret", "zh-Hans", "")

	rd, err := svc.GetReport(ctx, "pending-rpt")
	if err != nil {
		t.Fatalf("GetReport: %v", err)
	}
	if rd.LlmJSON != "" {
		t.Errorf("LlmJSON = %q, want empty for pending order", rd.LlmJSON)
	}
}
