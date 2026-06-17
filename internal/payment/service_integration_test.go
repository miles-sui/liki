//go:build integration

package payment

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"liki/internal/dodo"
	
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

type stubReportAgent struct {
	result string
	err    error
}

func (a *stubReportAgent) GenerateFromData(_ context.Context, _ string, _ agent.Product, _ json.RawMessage) (string, error) {
	return a.result, a.err
}

type stubDodo struct {
	checkoutResult *dodo.CheckoutResult
	checkoutErr    error
	webhookEvent   *dodo.WebhookEvent
	webhookErr     error
}

func (d *stubDodo) CreateCheckout(_ context.Context, _ string, _ int, _, _, _ string) (*dodo.CheckoutResult, error) {
	return d.checkoutResult, d.checkoutErr
}

func (d *stubDodo) VerifyWebhook(_ []byte, _ http.Header) (*dodo.WebhookEvent, error) {
	return d.webhookEvent, d.webhookErr
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

func productIDMap() map[agent.Product]string {
	return map[agent.Product]string{
		agent.ProductChart:  "prod_chart",
		agent.ProductBond:   "prod_bond",
		agent.ProductNaming: "prod_naming",
	}
}

func newTestService(t *testing.T, dodo *stubDodo, email *trackedEmailClient, report *stubReportAgent) *Service {
	t.Helper()
	return NewService(dodo, email, openServiceTestDB(t), productIDMap(), "https://example.com", "admin@example.com", report, context.Background())
}

// ── Webhook → MarkPaid → Email → Report Generation full pipeline ──

func TestWebhookToReportPipeline(t *testing.T) {
	dodoCli := &stubDodo{
		webhookEvent: &dodo.WebhookEvent{
			Type: "payment.succeeded",
			Data: dodo.WebhookEventData{
				OrderID:   "pipe-1",
				Amount:    990,
				PaymentID: "pay-pipe-1",
			},
		},
	}
	emailCli := &trackedEmailClient{}
	reportAgent := &stubReportAgent{result: "<p>generated report</p>"}
	svc := newTestService(t, dodoCli, emailCli, reportAgent)

	ctx := context.Background()

	// 1. Create a pending order.
	if err := svc.Store.CreateOrder(ctx, "pipe-1", agent.ProductChart, 990, "CNY", `{"chart":{"nianzhu":{"gan":"甲","zhi":"子"}}}`, "", "zh-Hans"); err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}
	svc.Store.UpdateEmail(ctx, "pipe-1", "buyer@test.com")

	// 2. Handle webhook — simulates Dodo callback.
	err := svc.HandleWebhook(ctx, []byte(`{"type":"payment.succeeded"}`), http.Header{})
	if err != nil {
		t.Fatalf("HandleWebhook: %v", err)
	}

	// 3. Verify order status changed to paid.
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

	// 4. Email was sent to buyer (goroutine — poll).
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

	// 5. Report generation runs in background — wait for it.
	for i := 0; i < 20; i++ {
		o, _ = svc.Store.GetOrder(ctx, "pipe-1")
		if o.LlmJSON != "" {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if o.LlmJSON != "<p>generated report</p>" {
		t.Errorf("LlmJSON = %q, want <p>generated report</p>", o.LlmJSON)
	}
}

// ── Webhook idempotency — no double email, no double generation ──

func TestWebhookToReport_Idempotent(t *testing.T) {
	dodoCli := &stubDodo{
		webhookEvent: &dodo.WebhookEvent{
			Type: "payment.succeeded",
			Data: dodo.WebhookEventData{
				OrderID:   "idem-1",
				Amount:    990,
				PaymentID: "pay-idem-1",
			},
		},
	}
	emailCli := &trackedEmailClient{}
	reportAgent := &stubReportAgent{result: "<p>ok</p>"}
	svc := newTestService(t, dodoCli, emailCli, reportAgent)

	ctx := context.Background()
	svc.Store.CreateOrder(ctx, "idem-1", agent.ProductChart, 990, "CNY", `{}`, "", "zh-Hans")
	svc.Store.UpdateEmail(ctx, "idem-1", "buyer@test.com")

	// First webhook.
	if err := svc.HandleWebhook(ctx, []byte(`{"type":"payment.succeeded"}`), http.Header{}); err != nil {
		t.Fatalf("first HandleWebhook: %v", err)
	}

	// Wait for background generation.
	for i := 0; i < 20; i++ {
		o, _ := svc.Store.GetOrder(ctx, "idem-1")
		if o.LlmJSON != "" {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	emailCount := emailCli.count()

	// Second webhook — same PaymentID.
	if err := svc.HandleWebhook(ctx, []byte(`{"type":"payment.succeeded"}`), http.Header{}); err != nil {
		t.Fatalf("second HandleWebhook: %v", err)
	}

	// Email count must not increase.
	if emailCli.count() != emailCount {
		t.Errorf("email count = %d, want %d (no new emails)", emailCli.count(), emailCount)
	}

	// LlmJSON must not be overwritten.
	o, _ := svc.Store.GetOrder(ctx, "idem-1")
	if o.LlmJSON != "<p>ok</p>" {
		t.Errorf("LlmJSON = %q, want <p>ok</p>", o.LlmJSON)
	}
}

// ── Webhook with non-existent order returns error ──

func TestWebhook_NonExistentOrder(t *testing.T) {
	dodoCli := &stubDodo{
		webhookEvent: &dodo.WebhookEvent{
			Type: "payment.succeeded",
			Data: dodo.WebhookEventData{
				OrderID:   "no-such-order",
				Amount:    990,
				PaymentID: "pay-unknown",
			},
		},
	}
	svc := newTestService(t, dodoCli, &trackedEmailClient{}, &stubReportAgent{result: "x"})

	err := svc.HandleWebhook(context.Background(), []byte(`{}`), http.Header{})
	if err == nil {
		t.Fatal("expected error for non-existent order")
	}
	if !errors.Is(err, ErrOrderNotFound) {
		t.Errorf("error = %v, want ErrOrderNotFound", err)
	}
}

// ── Checkout flow creates Dodo session and updates email ──

func TestCreateCheckout_Integration(t *testing.T) {
	dodoCli := &stubDodo{
		checkoutResult: &dodo.CheckoutResult{
			SessionID:   "sess-checkout",
			CheckoutURL: "https://pay.example.com/checkout",
		},
	}
	svc := newTestService(t, dodoCli, &trackedEmailClient{}, nil)
	ctx := context.Background()

	svc.Store.CreateOrder(ctx, "co-1", agent.ProductBond, 1990, "CNY", `{}`, "", "zh-Hans")

	result, err := svc.CreateCheckout(ctx, "co-1", "buyer@test.com")
	if err != nil {
		t.Fatalf("CreateCheckout: %v", err)
	}
	if result.SessionID != "sess-checkout" {
		t.Errorf("SessionID = %q, want sess-checkout", result.SessionID)
	}

	// Email must be persisted.
	o, _ := svc.Store.GetOrder(ctx, "co-1")
	if o.Email != "buyer@test.com" {
		t.Errorf("Email = %q, want buyer@test.com", o.Email)
	}
}

// ── RetryReportGeneration for paid orders without llm_json ──

func TestRetryReportGeneration_Integration(t *testing.T) {
	dodoCli := &stubDodo{}
	emailCli := &trackedEmailClient{}
	reportAgent := &stubReportAgent{result: "<p>retry generated</p>"}
	svc := newTestService(t, dodoCli, emailCli, reportAgent)
	ctx := context.Background()

	// Create and manually mark as paid (simulating missed webhook).
	svc.Store.CreateOrder(ctx, "retry-1", agent.ProductChart, 990, "CNY", `{"chart":{}}`, "", "zh-Hans")
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
		t.Errorf("llmJSON = %q, want empty (bg gen not complete)", llmJSON)
	}

	// Wait for background generation.
	for i := 0; i < 20; i++ {
		o, _ := svc.Store.GetOrder(ctx, "retry-1")
		if o.LlmJSON != "" {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	o, _ := svc.Store.GetOrder(ctx, "retry-1")
	if o.LlmJSON != "<p>retry generated</p>" {
		t.Errorf("LlmJSON = %q, want <p>retry generated</p>", o.LlmJSON)
	}
}

// ── GetReport hides llm_json for pending orders ──

func TestGetReport_HidesLlmJSONForPending(t *testing.T) {
	svc := newTestService(t, &stubDodo{}, &trackedEmailClient{}, nil)
	ctx := context.Background()

	svc.Store.CreateOrder(ctx, "pending-rpt", agent.ProductNaming, 2990, "CNY", `{}`, "# secret", "zh-Hans")

	rd, err := svc.GetReport(ctx, "pending-rpt")
	if err != nil {
		t.Fatalf("GetReport: %v", err)
	}
	if rd.LlmJSON != "" {
		t.Errorf("LlmJSON = %q, want empty for pending order", rd.LlmJSON)
	}
}
