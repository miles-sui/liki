package payment

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"testing"
	"time"

	"liki/internal/dodo"
	

	"liki/internal/agent"

	_ "modernc.org/sqlite"
)

func TestEmailSubject_KnownUnique(t *testing.T) {
	products := []agent.Product{agent.ProductChart, agent.ProductBond, agent.ProductNaming}
	seen := make(map[agent.Product]string)
	for _, p := range products {
		s := p.EmailSubject()
		if s == "" {
			t.Errorf("EmailSubject(%s) is empty", p)
		}
		seen[p] = s
	}
	if seen[agent.ProductChart] == seen[agent.ProductBond] || seen[agent.ProductChart] == seen[agent.ProductNaming] || seen[agent.ProductBond] == seen[agent.ProductNaming] {
		t.Errorf("products must have distinct subjects: %v", seen)
	}
}

func TestEmailSubject_Default(t *testing.T) {
	def := agent.ProductChart.EmailSubject()
	for _, p := range []agent.Product{"unknown", "", "x"} {
		if got := p.EmailSubject(); got == def {
			t.Errorf("EmailSubject(%q) unexpectedly equals chart subject", p)
		}
	}
}

// -- mocks --

type mockDodoClient struct {
	createResult *dodo.CheckoutResult
	createErr    error
	verifyEvent  *dodo.WebhookEvent
	verifyErr    error
}

func (m *mockDodoClient) CreateCheckout(ctx context.Context, productID string, amount int, orderID, email, returnURL string) (*dodo.CheckoutResult, error) {
	return m.createResult, m.createErr
}

func (m *mockDodoClient) VerifyWebhook(rawBody []byte, headers http.Header) (*dodo.WebhookEvent, error) {
	return m.verifyEvent, m.verifyErr
}

type mockEmailClient struct {
	mu      sync.Mutex
	sentTo  []string
	sendErr error
	sent    chan struct{} // receives once per SendReport call
}

func (m *mockEmailClient) SendReport(ctx context.Context, to, subject, htmlBody string) error {
	m.mu.Lock()
	m.sentTo = append(m.sentTo, to)
	m.mu.Unlock()
	if m.sent != nil {
		m.sent <- struct{}{}
	}
	return m.sendErr
}

func (m *mockEmailClient) sentToCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.sentTo)
}

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	db.SetMaxOpenConns(1)
	return db
}

func newTestSvc(t *testing.T) (*Service, *Store, *mockDodoClient, *mockEmailClient) {
	t.Helper()
	db := newTestDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	// Seed a pending order.
	if err := store.CreateOrder(context.Background(), "order-1", agent.ProductChart, 990, "USD", `{"chart":"data"}`, "", "zh-Hans"); err != nil {
		t.Fatalf("seed order: %v", err)
	}

	dodoMock := &mockDodoClient{
		createResult: &dodo.CheckoutResult{CheckoutURL: "https://pay.example.com/checkout"},
	}
	emailMock := &mockEmailClient{}
	svc := NewService(dodoMock, emailMock, store,
		map[agent.Product]string{agent.ProductChart: "prod-chart"},
		"https://liki.hk", "admin@liki.hk", nil, context.Background(),
	)
	return svc, store, dodoMock, emailMock
}

// -- CreateCheckout --

func TestCreateCheckout_Success(t *testing.T) {
	svc, _, _, _ := newTestSvc(t)
	result, err := svc.CreateCheckout(context.Background(), "order-1", "")
	if err != nil {
		t.Fatalf("CreateCheckout: %v", err)
	}
	if result.CheckoutURL != "https://pay.example.com/checkout" {
		t.Errorf("CheckoutURL = %q", result.CheckoutURL)
	}
}

func TestCreateCheckout_NotFound(t *testing.T) {
	svc, _, _, _ := newTestSvc(t)
	_, err := svc.CreateCheckout(context.Background(), "nonexistent", "")
	if err == nil {
		t.Fatal("expected error for missing order")
	}
	if !errors.Is(err, ErrOrderNotFound) {
		t.Errorf("expected ErrOrderNotFound, got %v", err)
	}
}

func TestCreateCheckout_NoProductID(t *testing.T) {
	svc, _, _, _ := newTestSvc(t)
	// Bond has no product ID configured.
	store := svc.Store
	if err := store.CreateOrder(context.Background(), "order-2", agent.ProductBond, 1990, "USD", `{}`, "", "zh-Hans"); err != nil {
		t.Fatalf("seed order: %v", err)
	}
	_, err := svc.CreateCheckout(context.Background(), "order-2", "")
	if err == nil {
		t.Fatal("expected error for missing product ID")
	}
}

func TestCreateCheckout_WithEmail(t *testing.T) {
	svc, store, _, _ := newTestSvc(t)
	_, err := svc.CreateCheckout(context.Background(), "order-1", "user@example.com")
	if err != nil {
		t.Fatalf("CreateCheckout: %v", err)
	}
	order, err := store.GetOrder(context.Background(), "order-1")
	if err != nil {
		t.Fatalf("GetOrder: %v", err)
	}
	if order.Email != "user@example.com" {
		t.Errorf("email = %q, want user@example.com", order.Email)
	}
}

// -- HandleWebhook --

func TestHandleWebhook_VerifyFailure(t *testing.T) {
	svc, _, dodoMock, _ := newTestSvc(t)
	dodoMock.verifyErr = errors.New("bad signature")
	err := svc.HandleWebhook(context.Background(), []byte(`{}`), http.Header{})
	if err == nil {
		t.Fatal("expected verification error")
	}
	if !errors.Is(err, ErrWebhookVerify) {
		t.Errorf("expected ErrWebhookVerify, got %v", err)
	}
}

func TestHandleWebhook_NonPaymentEvent(t *testing.T) {
	svc, _, dodoMock, _ := newTestSvc(t)
	dodoMock.verifyEvent = &dodo.WebhookEvent{Type: "checkout.created", Data: dodo.WebhookEventData{}}
	err := svc.HandleWebhook(context.Background(), []byte(`{}`), http.Header{})
	if err != nil {
		t.Fatalf("HandleWebhook non-payment: %v", err)
	}
}

func TestHandleWebhook_PaymentSucceeded(t *testing.T) {
	svc, store, dodoMock, emailMock := newTestSvc(t)

	// Give the order an email so the customer email path is exercised.
	if err := store.UpdateEmail(context.Background(), "order-1", "user@example.com"); err != nil {
		t.Fatalf("UpdateEmail: %v", err)
	}

	emailMock.sent = make(chan struct{}, 4)

	dodoMock.verifyEvent = &dodo.WebhookEvent{
		Type: "payment.succeeded",
		Data: dodo.WebhookEventData{OrderID: "order-1", PaymentID: "pay-1", Amount: 990},
	}
	err := svc.HandleWebhook(context.Background(), []byte(`{}`), http.Header{})
	if err != nil {
		t.Fatalf("HandleWebhook: %v", err)
	}

	<-emailMock.sent // wait for customer email goroutine
	<-emailMock.sent // wait for admin copy goroutine

	order, err := store.GetOrder(context.Background(), "order-1")
	if err != nil {
		t.Fatalf("GetOrder: %v", err)
	}
	if err != nil { t.Fatal(err) }
	if err != nil { t.Fatal(err) }
	if order.Status != OrderPaid {
		t.Errorf("status = %s, want paid", order.Status)
	}
	if order.PaymentID != "pay-1" {
		t.Errorf("paymentID = %s, want pay-1", order.PaymentID)
	}
	if emailMock.sentToCount() == 0 {
		t.Error("customer email was not sent")
	}
}

func TestHandleWebhook_SecondPayment(t *testing.T) {
	// Two different PaymentIDs for the same order — both should succeed.
	svc, _, dodoMock, _ := newTestSvc(t)

	dodoMock.verifyEvent = &dodo.WebhookEvent{
		Type: "payment.succeeded",
		Data: dodo.WebhookEventData{OrderID: "order-1", PaymentID: "pay-1", Amount: 990},
	}
	if err := svc.HandleWebhook(context.Background(), []byte(`{}`), http.Header{}); err != nil {
		t.Fatalf("first HandleWebhook: %v", err)
	}

	// Second different payment for the same order.
	dodoMock.verifyEvent.Data.PaymentID = "pay-2"
	if err := svc.HandleWebhook(context.Background(), []byte(`{}`), http.Header{}); err != nil {
		t.Fatalf("second HandleWebhook: %v", err)
	}
}

func TestHandleWebhook_DuplicatePaymentIdempotent(t *testing.T) {
	// Same PaymentID sent twice must be idempotent — no double-processing.
	svc, store, dodoMock, emailMock := newTestSvc(t)
	emailMock.sent = make(chan struct{}, 4)

	if err := store.UpdateEmail(context.Background(), "order-1", "user@example.com"); err != nil {
		t.Fatalf("UpdateEmail: %v", err)
	}

	dodoMock.verifyEvent = &dodo.WebhookEvent{
		Type: "payment.succeeded",
		Data: dodo.WebhookEventData{OrderID: "order-1", PaymentID: "pay-1", Amount: 990},
	}

	// First webhook: should trigger emails and report generation.
	if err := svc.HandleWebhook(context.Background(), []byte(`{}`), http.Header{}); err != nil {
		t.Fatalf("first HandleWebhook: %v", err)
	}
	<-emailMock.sent // customer email
	<-emailMock.sent // admin email
	emailCount := emailMock.sentToCount()
	if emailCount < 2 {
		t.Errorf("first webhook: expected >=2 emails, got %d", emailCount)
	}

	// Second webhook with SAME PaymentID: must not re-send emails.
	if err := svc.HandleWebhook(context.Background(), []byte(`{}`), http.Header{}); err != nil {
		t.Fatalf("second HandleWebhook with same PaymentID: %v", err)
	}
	if emailMock.sentToCount() != emailCount {
		t.Errorf("duplicate webhook sent extra emails: %d before, %d after",
			emailCount, emailMock.sentToCount())
	}

	// Verify payment ID is preserved (first one wins).
	order, err := store.GetOrder(context.Background(), "order-1")
	if err != nil {
		t.Fatalf("GetOrder: %v", err)
	}
	if order.PaymentID != "pay-1" {
		t.Errorf("PaymentID = %q, want pay-1 (first payment preserved)", order.PaymentID)
	}
}

// -- GetReport --

func TestGetReport_Pending(t *testing.T) {
	svc, _, _, _ := newTestSvc(t)
	report, err := svc.GetReport(context.Background(), "order-1")
	if err != nil {
		t.Fatalf("GetReport: %v", err)
	}
	if report.Status != OrderPending {
		t.Errorf("status = %s, want pending", report.Status)
	}
	if report.LlmJSON != "" {
		t.Errorf("LlmJSON should be empty for pending order")
	}
}

func TestGetReport_Paid(t *testing.T) {
	svc, store, dodoMock, _ := newTestSvc(t)

	// Mark order as paid with llm_json.
	dodoMock.verifyEvent = &dodo.WebhookEvent{
		Type: "payment.succeeded",
		Data: dodo.WebhookEventData{OrderID: "order-1", PaymentID: "pay-1", Amount: 990},
	}
	if err := svc.HandleWebhook(context.Background(), []byte(`{}`), http.Header{}); err != nil {
		t.Fatalf("HandleWebhook: %v", err)
	}
	if err := store.UpdateLlmJSON(context.Background(), "order-1", `{"report":"content"}`); err != nil {
		t.Fatalf("UpdateLlmJSON: %v", err)
	}

	report, err := svc.GetReport(context.Background(), "order-1")
	if err != nil {
		t.Fatalf("GetReport: %v", err)
	}
	if report.Status != OrderPaid {
		t.Errorf("status = %s, want paid", report.Status)
	}
	if report.LlmJSON != `{"report":"content"}` {
		t.Errorf("LlmJSON = %q", report.LlmJSON)
	}
}

func TestGetReport_NotFound(t *testing.T) {
	svc, _, _, _ := newTestSvc(t)
	_, err := svc.GetReport(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing order")
	}
	if !errors.Is(err, ErrOrderNotFound) {
		t.Errorf("expected ErrOrderNotFound, got %v", err)
	}
}

// -- OrderStatus --

func TestOrderStatus_Found(t *testing.T) {
	svc, _, _, _ := newTestSvc(t)
	status, product, err := svc.OrderStatus(context.Background(), "order-1")
	if err != nil {
		t.Fatalf("OrderStatus: %v", err)
	}
	if status != OrderPending {
		t.Errorf("status = %s, want pending", status)
	}
	if product != agent.ProductChart {
		t.Errorf("product = %s, want chart", product)
	}
}

func TestOrderStatus_NotFound(t *testing.T) {
	svc, _, _, _ := newTestSvc(t)
	_, _, err := svc.OrderStatus(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing order")
	}
}

// -- mock ReportAgent --

type mockReportGen struct {
	mu       sync.Mutex
	count    int
	content  string
	err      error
	panic    bool
	blockCh  chan struct{} // if set, GenerateFromData blocks until closed (for dedup test)
	calledCh chan struct{} // signals entry into GenerateFromData (non-blocking)
	doneCh   chan struct{} // signals GenerateFromData has returned (non-blocking)
}

func (m *mockReportGen) GenerateFromData(_ context.Context, _ string, _ agent.Product, _ json.RawMessage) (string, error) {
	m.mu.Lock()
	m.count++
	m.mu.Unlock()
	if m.calledCh != nil {
		select {
		case m.calledCh <- struct{}{}:
		default:
		}
	}
	if m.blockCh != nil {
		<-m.blockCh
	}
	defer func() {
		if m.doneCh != nil {
			select {
			case m.doneCh <- struct{}{}:
			default:
			}
		}
	}()
	if m.panic {
		panic("test panic in GenerateFromData")
	}
	return m.content, m.err
}

func (m *mockReportGen) callCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.count
}

func newTestSvcWithReportGen(t *testing.T, rg *mockReportGen) (*Service, *Store) {
	t.Helper()
	db := newTestDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	if err := store.CreateOrder(context.Background(), "order-1", agent.ProductChart, 990, "USD", `{"chart":"data"}`, "", "zh-Hans"); err != nil {
		t.Fatalf("seed order: %v", err)
	}
	dodoMock := &mockDodoClient{
		createResult: &dodo.CheckoutResult{CheckoutURL: "https://pay.example.com/checkout"},
	}
	emailMock := &mockEmailClient{}
	svc := NewService(dodoMock, emailMock, store,
		map[agent.Product]string{agent.ProductChart: "prod-chart"},
		"https://liki.hk", "admin@liki.hk", rg, context.Background(),
	)
	return svc, store
}

// -- StartReportGeneration --

func TestStartReportGeneration_Dedup(t *testing.T) {
	// Concurrent calls with the same orderID must only launch ONE goroutine.
	// This prevents wasting LLM tokens on duplicate report generation.
	rg := &mockReportGen{
		content:  "<p>report</p>",
		blockCh:  make(chan struct{}),
		calledCh: make(chan struct{}, 1),
	}
	svc, _ := newTestSvcWithReportGen(t, rg)

	// First call: should start generation (goroutine blocks on blockCh).
	svc.StartReportGeneration("order-1", agent.ProductChart, `{"x":1}`)

	// Wait for GenerateFromData to be entered.
	select {
	case <-rg.calledCh:
	case <-time.After(time.Second):
		t.Fatal("GenerateFromData was not called")
	}

	// Second call: must be no-op (already generating).
	svc.StartReportGeneration("order-1", agent.ProductChart, `{"x":1}`)

	if rg.callCount() != 1 {
		t.Errorf("GenerateFromData called %d times, want 1 (dedup failed)", rg.callCount())
	}

	// Unblock and let it finish.
	close(rg.blockCh)
}

func TestStartReportGeneration_RetriggersAfterCompletion(t *testing.T) {
	// After generation completes, a new call should re-trigger generation.
	// This is the recovery path: if first gen failed, retry must work.
	rg := &mockReportGen{
		content: "<p>report</p>",
		doneCh:  make(chan struct{}, 2),
	}
	svc, _ := newTestSvcWithReportGen(t, rg)

	svc.StartReportGeneration("order-1", agent.ProductChart, `{"x":1}`)
	<-rg.doneCh // wait for first gen to return

	// doneCh fires inside GenerateFromData; the goroutine still needs a moment
	// to run its deferred map cleanup before we can re-trigger.
	time.Sleep(50 * time.Millisecond)

	// Now calling again should re-trigger (entry was cleaned up).
	svc.StartReportGeneration("order-1", agent.ProductChart, `{"x":1}`)
	<-rg.doneCh

	if rg.callCount() != 2 {
		t.Errorf("GenerateFromData called %d times, want 2 (retrigger failed)", rg.callCount())
	}
}

// -- generateFullReport (via StartReportGeneration) --

func TestGenerateFullReport_Success(t *testing.T) {
	rg := &mockReportGen{
		content:  "<p>generated report content</p>",
		doneCh: make(chan struct{}, 1),
	}
	svc, store := newTestSvcWithReportGen(t, rg)

	svc.StartReportGeneration("order-1", agent.ProductChart, `{"chart":"data"}`)
	<-rg.doneCh

	order, err := store.GetOrder(context.Background(), "order-1")
	if err != nil {
		t.Fatalf("GetOrder: %v", err)
	}
	if order.LlmJSON != "<p>generated report content</p>" {
		t.Errorf("llm_json = %q, want generated report content (not cached)", order.LlmJSON)
	}
}

func TestGenerateFullReport_Error(t *testing.T) {
	rg := &mockReportGen{
		err:    errors.New("LLM timeout"),
		doneCh: make(chan struct{}, 1),
	}
	svc, store := newTestSvcWithReportGen(t, rg)

	svc.StartReportGeneration("order-1", agent.ProductChart, `{"chart":"data"}`)
	<-rg.doneCh

	order, err := store.GetOrder(context.Background(), "order-1")
	if err != nil {
		t.Fatalf("GetOrder: %v", err)
	}
	if order.LlmJSON != "" {
		t.Errorf("llm_json = %q, want empty (generation failed, should not cache)", order.LlmJSON)
	}
}

func TestGenerateFullReport_PanicRecovery(t *testing.T) {
	// If GenerateFromData panics, the goroutine must recover and not crash the process.
	rg := &mockReportGen{
		panic:  true,
		doneCh: make(chan struct{}, 1),
	}
	svc, store := newTestSvcWithReportGen(t, rg)

	// This must not panic the test.
	svc.StartReportGeneration("order-1", agent.ProductChart, `{"chart":"data"}`)
	<-rg.doneCh

	order, err := store.GetOrder(context.Background(), "order-1")
	if err != nil {
		t.Fatalf("GetOrder: %v", err)
	}
	if order.LlmJSON != "" {
		t.Errorf("llm_json = %q, want empty (panicked, should not cache)", order.LlmJSON)
	}
}

func TestGenerateFullReport_DefaultLocale(t *testing.T) {
	// Order with empty locale defaults to zh-Hans.
	db := newTestDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	if err := store.CreateOrder(context.Background(), "order-no-locale", agent.ProductChart, 990, "USD", `{"x":1}`, "", ""); err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}

	var gotLocale string
	rg := &mockReportGen{
		content:  "<p>ok</p>",
		doneCh: make(chan struct{}, 1),
	}
	// Override GenerateFromData to capture locale.
	localeRG := &localeCapturingReportGen{
		mockReportGen: rg,
		locale:        &gotLocale,
	}
	svc := NewService(&mockDodoClient{}, &mockEmailClient{}, store,
		map[agent.Product]string{agent.ProductChart: "prod-chart"},
		"https://liki.hk", "admin@liki.hk", localeRG, context.Background(),
	)

	svc.StartReportGeneration("order-no-locale", agent.ProductChart, `{"x":1}`)
	<-rg.doneCh

	if gotLocale != "zh-Hans" {
		t.Errorf("locale = %q, want zh-Hans (default)", gotLocale)
	}
}

type localeCapturingReportGen struct {
	*mockReportGen
	locale *string
}

func (m *localeCapturingReportGen) GenerateFromData(ctx context.Context, locale string, product agent.Product, chartJSON json.RawMessage) (string, error) {
	*m.locale = locale
	return m.mockReportGen.GenerateFromData(ctx, locale, product, chartJSON)
}

// -- RetryReportGeneration --

func TestRetryReportGeneration_PaidNoJSON(t *testing.T) {
	rg := &mockReportGen{
		content: "<p>recovered</p>",
		doneCh:  make(chan struct{}, 1),
	}
	svc, store := newTestSvcWithReportGen(t, rg)

	// Mark the order as paid via webhook, but prevent report generation
	// so we can test the retry path (paid + empty llm_json).
	dodoMock := svc.Dodo.(*mockDodoClient)
	dodoMock.verifyEvent = &dodo.WebhookEvent{
		Type: "payment.succeeded",
		Data: dodo.WebhookEventData{OrderID: "order-1", PaymentID: "pay-1", Amount: 990},
	}
	origRA := svc.ReportAgent
	svc.ReportAgent = nil
	if err := svc.HandleWebhook(context.Background(), []byte(`{}`), http.Header{}); err != nil {
		t.Fatalf("HandleWebhook: %v", err)
	}
	svc.ReportAgent = origRA

	// Verify paid, no llm_json.
	order, err := store.GetOrder(context.Background(), "order-1")
	if err != nil { t.Fatal(err) }
	if err != nil { t.Fatal(err) }
	if order.Status != OrderPaid || order.LlmJSON != "" {
		t.Fatalf("setup: status=%s llm_json=%q, want paid+empty", order.Status, order.LlmJSON)
	}

	// Retry should trigger generation.
	status, product, llmJSON, err := svc.RetryReportGeneration(context.Background(), "order-1")
	if err != nil {
		t.Fatalf("RetryReportGeneration: %v", err)
	}
	if status != OrderPaid {
		t.Errorf("status = %s, want paid", status)
	}
	if product != agent.ProductChart {
		t.Errorf("product = %s", product)
	}
	if llmJSON != "" {
		t.Errorf("llmJSON = %q, want empty (bg generation not yet complete)", llmJSON)
	}

	// Wait for background generation.
	<-rg.doneCh

	order, orderErr := store.GetOrder(context.Background(), "order-1")
	if orderErr != nil {
		t.Fatalf("GetOrder: %v", orderErr)
	}
	if order.LlmJSON != "<p>recovered</p>" {
		t.Errorf("llm_json = %q, want recovered (retry should trigger generation)", order.LlmJSON)
	}
}

func TestRetryReportGeneration_PaidWithJSON(t *testing.T) {
	rg := &mockReportGen{
		content: "<p>existing</p>",
		doneCh:  make(chan struct{}, 2),
	}
	svc, _ := newTestSvcWithReportGen(t, rg)

	// Mark as paid with llm_json via webhook.
	dodoMock := svc.Dodo.(*mockDodoClient)
	dodoMock.verifyEvent = &dodo.WebhookEvent{
		Type: "payment.succeeded",
		Data: dodo.WebhookEventData{OrderID: "order-1", PaymentID: "pay-1", Amount: 990},
	}
	if err := svc.HandleWebhook(context.Background(), []byte(`{}`), http.Header{}); err != nil {
		t.Fatalf("HandleWebhook: %v", err)
	}
	<-rg.doneCh // wait for webhook-triggered generation

	// Now retry: must NOT trigger new generation since llm_json exists.
	genCountBefore := rg.callCount()
	status, _, llmJSON, err := svc.RetryReportGeneration(context.Background(), "order-1")
	if err != nil {
		t.Fatalf("RetryReportGeneration: %v", err)
	}
	if status != OrderPaid {
		t.Errorf("status = %s, want paid", status)
	}
	if llmJSON == "" {
		t.Error("llmJSON should not be empty for paid order with report")
	}
	if rg.callCount() != genCountBefore {
		t.Errorf("GenerateFromData called %d extra times, want 0 (should not regenerate when llm_json exists)", rg.callCount()-genCountBefore)
	}
}

func TestRetryReportGeneration_Pending(t *testing.T) {
	rg := &mockReportGen{content: "<p>nope</p>"}
	svc, _ := newTestSvcWithReportGen(t, rg)

	status, _, llmJSON, err := svc.RetryReportGeneration(context.Background(), "order-1")
	if err != nil {
		t.Fatalf("RetryReportGeneration: %v", err)
	}
	if status != OrderPending {
		t.Errorf("status = %s, want pending", status)
	}
	if llmJSON != "" {
		t.Errorf("llmJSON = %q, want empty for pending", llmJSON)
	}
	if rg.callCount() > 0 {
		t.Errorf("GenerateFromData called %d times, want 0 (pending order should not trigger)", rg.callCount())
	}
}

func TestRetryReportGeneration_NotFound(t *testing.T) {
	rg := &mockReportGen{}
	svc, _ := newTestSvcWithReportGen(t, rg)

	_, _, _, err := svc.RetryReportGeneration(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing order")
	}
}
