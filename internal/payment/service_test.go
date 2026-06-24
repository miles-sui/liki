package payment

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"liki/internal/agent"
	"liki/internal/llm"

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

type mockPaymentProvider struct {
	createResult *CheckoutResult
	createErr    error
	verifyEvent  *WebhookEvent
	verifyErr    error
}

func (m *mockPaymentProvider) CreateCheckout(_ context.Context, _ agent.Product, _ int, _, _, _ string) (*CheckoutResult, error) {
	return m.createResult, m.createErr
}

func (m *mockPaymentProvider) VerifyWebhook(_ []byte, _ http.Header) (*WebhookEvent, error) {
	return m.verifyEvent, m.verifyErr
}

type mockEmailClient struct {
	mu      sync.Mutex
	sentTo  []string
	sendErr error
	sent    chan struct{} // receives once per SendReport call
}

func (m *mockEmailClient) SendReport(_ context.Context, to, _, _ string) error {
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

// waitForCleanup polls the generating map until the goroutine finishes
// cleanup, avoiding time.Sleep which is flaky on slow CI runners.
func waitForCleanup(t *testing.T, svc *Service, orderID string) {
	t.Helper()
	for i := 0; i < 100; i++ {
		svc.generatingMu.Lock()
		_, ok := svc.generating[orderID]
		svc.generatingMu.Unlock()
		if !ok {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Errorf("generating map entry %q not cleaned up after 100ms", orderID)
}

func newTestSvc(t *testing.T) (*Service, *Store, *mockPaymentProvider, *mockEmailClient) {
	t.Helper()
	db := newTestDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	// Seed a pending order.
	if err := store.CreateOrder(context.Background(), "order-1", agent.ProductChart, 990, "CNY", `{"chart":"data"}`, "", "zh-Hans", ""); err != nil {
		t.Fatalf("seed order: %v", err)
	}

	dodoMock := &mockPaymentProvider{
		createResult: &CheckoutResult{CheckoutURL: "https://pay.example.com/checkout"},
	}
	xunhuMock := &mockPaymentProvider{
		createResult: &CheckoutResult{CheckoutURL: "https://pay.xunhu.com/checkout", QRCodeURL: "https://pay.xunhu.com/qr"},
	}
	emailMock := &mockEmailClient{}
	svc := NewService(dodoMock, xunhuMock, emailMock, store,
		"https://liki.hk", "admin@liki.hk", nil, context.Background(),
	)
	return svc, store, dodoMock, emailMock
}

// -- CreateCheckout --

func TestCreateCheckout_Success(t *testing.T) {
	svc, _, _, _ := newTestSvc(t)
	result, err := svc.CreateCheckout(context.Background(), "dodo", "order-1", "")
	if err != nil {
		t.Fatalf("CreateCheckout: %v", err)
	}
	if result.CheckoutURL != "https://pay.example.com/checkout" {
		t.Errorf("CheckoutURL = %q", result.CheckoutURL)
	}
}

func TestCreateCheckout_Xunhu(t *testing.T) {
	svc, _, _, _ := newTestSvc(t)
	result, err := svc.CreateCheckout(context.Background(), "xunhu", "order-1", "")
	if err != nil {
		t.Fatalf("CreateCheckout xunhu: %v", err)
	}
	if result.CheckoutURL != "https://pay.xunhu.com/checkout" {
		t.Errorf("CheckoutURL = %q", result.CheckoutURL)
	}
	if result.QRCodeURL != "https://pay.xunhu.com/qr" {
		t.Errorf("QRCodeURL = %q", result.QRCodeURL)
	}
}

func TestCreateCheckout_NotFound(t *testing.T) {
	svc, _, _, _ := newTestSvc(t)
	_, err := svc.CreateCheckout(context.Background(), "dodo", "nonexistent", "")
	if err == nil {
		t.Fatal("expected error for missing order")
	}
	if !errors.Is(err, ErrOrderNotFound) {
		t.Errorf("expected ErrOrderNotFound, got %v", err)
	}
}

func TestCreateCheckout_UnknownProvider(t *testing.T) {
	svc, _, _, _ := newTestSvc(t)
	_, err := svc.CreateCheckout(context.Background(), "unknown", "order-1", "")
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestCreateCheckout_ProviderError(t *testing.T) {
	svc, _, dodoMock, _ := newTestSvc(t)
	dodoMock.createErr = errors.New("api error")
	_, err := svc.CreateCheckout(context.Background(), "dodo", "order-1", "")
	if err == nil {
		t.Fatal("expected error from provider")
	}
}

func TestCreateCheckout_WithEmail(t *testing.T) {
	svc, store, _, _ := newTestSvc(t)
	_, err := svc.CreateCheckout(context.Background(), "dodo", "order-1", "user@example.com")
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
	xunhuMock := svc.Xunhu.(*mockPaymentProvider)
	dodoMock.verifyErr = errors.New("bad signature")
	xunhuMock.verifyErr = errors.New("bad hash")
	err := svc.HandleWebhook(context.Background(), []byte(`{}`), http.Header{})
	if err == nil {
		t.Fatal("expected verification error")
	}
	if !errors.Is(err, ErrWebhookVerify) {
		t.Errorf("expected ErrWebhookVerify, got %v", err)
	}
}

func TestHandleWebhook_XunhuFallback(t *testing.T) {
	// Dodo verification fails, xunhu succeeds — dispatch falls back to xunhu.
	svc, store, dodoMock, emailMock := newTestSvc(t)
	emailMock.sent = make(chan struct{}, 4)

	if err := store.UpdateEmail(context.Background(), "order-1", "user@example.com"); err != nil {
		t.Fatalf("UpdateEmail: %v", err)
	}

	dodoMock.verifyErr = errors.New("dodo: bad signature")
	xunhuMock := svc.Xunhu.(*mockPaymentProvider)
	xunhuMock.verifyEvent = &WebhookEvent{
		Type: "payment.succeeded",
		Data: WebhookEventData{OrderID: "order-1", PaymentID: "xunhu-pay-1", Amount: 990},
	}

	err := svc.HandleWebhook(context.Background(), []byte(`trade_status=TRADE_SUCCESS&trade_order_id=order-1`), http.Header{})
	if err != nil {
		t.Fatalf("HandleWebhook xunhu fallback: %v", err)
	}
	<-emailMock.sent // customer email
	<-emailMock.sent // admin email

	order, err := store.GetOrder(context.Background(), "order-1")
	if err != nil {
		t.Fatalf("GetOrder: %v", err)
	}
	if order.Status != OrderPaid {
		t.Errorf("status = %s, want paid", order.Status)
	}
	if order.PaymentID != "xunhu-pay-1" {
		t.Errorf("PaymentID = %q, want xunhu-pay-1", order.PaymentID)
	}
}

func TestHandleWebhook_BothProvidersFail(t *testing.T) {
	svc, _, dodoMock, _ := newTestSvc(t)
	xunhuMock := svc.Xunhu.(*mockPaymentProvider)
	dodoMock.verifyErr = errors.New("dodo: bad signature")
	xunhuMock.verifyErr = errors.New("xunhu: bad hash")

	err := svc.HandleWebhook(context.Background(), []byte(`{}`), http.Header{})
	if err == nil {
		t.Fatal("expected error when both providers fail")
	}
	if !errors.Is(err, ErrWebhookVerify) {
		t.Errorf("expected ErrWebhookVerify, got %v", err)
	}
}

func TestHandleWebhook_EmptyOrderID(t *testing.T) {
	svc, _, dodoMock, _ := newTestSvc(t)
	dodoMock.verifyEvent = &WebhookEvent{
		Type: "payment.succeeded",
		Data: WebhookEventData{OrderID: "", PaymentID: "pay-1", Amount: 990},
	}
	err := svc.HandleWebhook(context.Background(), []byte(`{}`), http.Header{})
	if err == nil {
		t.Fatal("expected error for empty order_id")
	}
}

func TestHandleWebhook_NonPaymentEvent(t *testing.T) {
	svc, _, dodoMock, _ := newTestSvc(t)
	dodoMock.verifyEvent = &WebhookEvent{Type: "checkout.created", Data: WebhookEventData{}}
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

	dodoMock.verifyEvent = &WebhookEvent{
		Type: "payment.succeeded",
		Data: WebhookEventData{OrderID: "order-1", PaymentID: "pay-1", Amount: 990},
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

func TestHandleWebhook_SecondPaymentIgnored(t *testing.T) {
	// Second payment with a different PaymentID must be silently ignored
	// because MarkPaidIdempotent only updates when status='pending'.
	svc, store, dodoMock, emailMock := newTestSvc(t)
	emailMock.sent = make(chan struct{}, 4)

	if err := store.UpdateEmail(context.Background(), "order-1", "user@example.com"); err != nil {
		t.Fatalf("UpdateEmail: %v", err)
	}

	dodoMock.verifyEvent = &WebhookEvent{
		Type: "payment.succeeded",
		Data: WebhookEventData{OrderID: "order-1", PaymentID: "pay-1", Amount: 990},
	}
	if err := svc.HandleWebhook(context.Background(), []byte(`{}`), http.Header{}); err != nil {
		t.Fatalf("first HandleWebhook: %v", err)
	}
	<-emailMock.sent // customer
	<-emailMock.sent // admin
	emailCount := emailMock.sentToCount()

	// Second different payment for the same order — silently ignored.
	dodoMock.verifyEvent.Data.PaymentID = "pay-2"
	if err := svc.HandleWebhook(context.Background(), []byte(`{}`), http.Header{}); err != nil {
		t.Fatalf("second HandleWebhook: %v", err)
	}

	// Must not re-send emails.
	if emailMock.sentToCount() != emailCount {
		t.Errorf("second payment sent extra emails: %d before, %d after",
			emailCount, emailMock.sentToCount())
	}

	// PaymentID must still be the first one (first payment wins).
	order, err := store.GetOrder(context.Background(), "order-1")
	if err != nil {
		t.Fatalf("GetOrder: %v", err)
	}
	if order.PaymentID != "pay-1" {
		t.Errorf("PaymentID = %q, want pay-1 (first payment preserved)", order.PaymentID)
	}
}

func TestHandleWebhook_DuplicatePaymentIdempotent(t *testing.T) {
	// Same PaymentID sent twice must be idempotent — no double-processing.
	svc, store, dodoMock, emailMock := newTestSvc(t)
	emailMock.sent = make(chan struct{}, 4)

	if err := store.UpdateEmail(context.Background(), "order-1", "user@example.com"); err != nil {
		t.Fatalf("UpdateEmail: %v", err)
	}

	dodoMock.verifyEvent = &WebhookEvent{
		Type: "payment.succeeded",
		Data: WebhookEventData{OrderID: "order-1", PaymentID: "pay-1", Amount: 990},
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
	dodoMock.verifyEvent = &WebhookEvent{
		Type: "payment.succeeded",
		Data: WebhookEventData{OrderID: "order-1", PaymentID: "pay-1", Amount: 990},
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

// -- controllable LLM mock for report generation tests --

type controllableLLM struct {
	mu               sync.Mutex
	count            int
	content          string
	err              error
	panic            bool
	blockCh          chan struct{}
	calledCh         chan struct{}
	doneCh           chan struct{}
	lastSystemPrompt string
}

func (m *controllableLLM) ChatStreamWithTools(ctx context.Context, messages []llm.Message, tools []llm.ToolDef) (<-chan llm.StreamEvent, error) {
	m.mu.Lock()
	m.count++
	if len(messages) > 0 {
		m.lastSystemPrompt = messages[0].Content
	}
	m.mu.Unlock()

	if m.calledCh != nil {
		select {
		case m.calledCh <- struct{}{}:
		default:
		}
	}
	if m.blockCh != nil {
		select {
		case <-m.blockCh:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
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
		panic("test panic in ChatStreamWithTools")
	}
	if m.err != nil {
		return nil, m.err
	}
	ch := make(chan llm.StreamEvent, 1)
	ch <- llm.StreamEvent{Content: m.content, FinishReason: "stop"}
	close(ch)
	return ch, nil
}

func (m *controllableLLM) ChatStream(ctx context.Context, systemPrompt, userMessage string) (<-chan string, error) {
	return nil, errors.New("ChatStream not implemented")
}

func (m *controllableLLM) callCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.count
}

func newTestSvcWithReportGen(t *testing.T, cllm *controllableLLM) (*Service, *Store) {
	t.Helper()
	db := newTestDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	if err := store.CreateOrder(context.Background(), "order-1", agent.ProductChart, 990, "CNY", `{"chart":"data"}`, "", "zh-Hans", ""); err != nil {
		t.Fatalf("seed order: %v", err)
	}
	dodoMock := &mockPaymentProvider{
		createResult: &CheckoutResult{CheckoutURL: "https://pay.example.com/checkout"},
	}
	xunhuMock := &mockPaymentProvider{}
	emailMock := &mockEmailClient{}
	tools := &agent.MockToolRegistry{}
	ra := agent.NewReportAgent(cllm, tools, "", "")
	reportAgents := map[agent.Product]*agent.ReportAgent{
		agent.ProductChart: ra,
	}
	svc := NewService(dodoMock, xunhuMock, emailMock, store,
		"https://liki.hk", "admin@liki.hk", reportAgents, context.Background(),
	)
	return svc, store
}

// -- StartReportGeneration --

func TestStartReportGeneration_Dedup(t *testing.T) {
	// Concurrent calls with the same orderID must only launch ONE goroutine.
	cllm := &controllableLLM{
		content:  "<p>report</p>",
		blockCh:  make(chan struct{}),
		calledCh: make(chan struct{}, 1),
	}
	svc, _ := newTestSvcWithReportGen(t, cllm)

	// First call: should start generation (goroutine blocks on blockCh).
	svc.StartReportGeneration("order-1", agent.ProductChart, `{"x":1}`)

	// Wait for Generate to be entered.
	select {
	case <-cllm.calledCh:
	case <-time.After(time.Second):
		t.Fatal("Generate was not called")
	}

	// Second call: must be no-op (already generating).
	svc.StartReportGeneration("order-1", agent.ProductChart, `{"x":1}`)

	if cllm.callCount() != 1 {
		t.Errorf("Generate called %d times, want 1 (dedup failed)", cllm.callCount())
	}

	// Unblock and let it finish.
	close(cllm.blockCh)
}

func TestStartReportGeneration_RetriggersAfterCompletion(t *testing.T) {
	// After generation completes, a new call should re-trigger generation.
	cllm := &controllableLLM{
		content: "<p>report</p>",
		doneCh:  make(chan struct{}, 2),
	}
	svc, _ := newTestSvcWithReportGen(t, cllm)

	svc.StartReportGeneration("order-1", agent.ProductChart, `{"x":1}`)
	<-cllm.doneCh // wait for first gen LLM call to return

	// Poll generating map until the goroutine finishes cleanup.
	waitForCleanup(t, svc, "order-1")

	svc.StartReportGeneration("order-1", agent.ProductChart, `{"x":1}`)
	<-cllm.doneCh // wait for second gen

	if cllm.callCount() != 2 {
		t.Errorf("Generate called %d times, want 2 (retrigger failed)", cllm.callCount())
	}
}

// -- generateFullReport (via StartReportGeneration) --

func TestGenerateFullReport_Success(t *testing.T) {
	cllm := &controllableLLM{
		content: "<p>generated report content</p>",
		doneCh:  make(chan struct{}, 1),
	}
	svc, store := newTestSvcWithReportGen(t, cllm)

	svc.StartReportGeneration("order-1", agent.ProductChart, `{"chart":"data"}`)
	<-cllm.doneCh
	waitForCleanup(t, svc, "order-1")

	order, err := store.GetOrder(context.Background(), "order-1")
	if err != nil {
		t.Fatalf("GetOrder: %v", err)
	}
	if order.LlmJSON != "<p>generated report content</p>" {
		t.Errorf("llm_json = %q, want generated report content (not cached)", order.LlmJSON)
	}
}

func TestGenerateFullReport_Error(t *testing.T) {
	cllm := &controllableLLM{
		err:    errors.New("LLM timeout"),
		doneCh: make(chan struct{}, 1),
	}
	svc, store := newTestSvcWithReportGen(t, cllm)

	svc.StartReportGeneration("order-1", agent.ProductChart, `{"chart":"data"}`)
	<-cllm.doneCh

	order, err := store.GetOrder(context.Background(), "order-1")
	if err != nil {
		t.Fatalf("GetOrder: %v", err)
	}
	if order.LlmJSON != "" {
		t.Errorf("llm_json = %q, want empty (generation failed, should not cache)", order.LlmJSON)
	}
}

func TestGenerateFullReport_PanicRecovery(t *testing.T) {
	// If Generate panics, the goroutine must recover and not crash the process.
	cllm := &controllableLLM{
		panic:  true,
		doneCh: make(chan struct{}, 1),
	}
	svc, store := newTestSvcWithReportGen(t, cllm)

	// This must not panic the test.
	svc.StartReportGeneration("order-1", agent.ProductChart, `{"chart":"data"}`)
	<-cllm.doneCh

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
	if err := store.CreateOrder(context.Background(), "order-no-locale", agent.ProductChart, 990, "CNY", `{"x":1}`, "", "", ""); err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}

	cllm := &controllableLLM{
		content: "<p>ok</p>",
		doneCh:  make(chan struct{}, 1),
	}
	tools := &agent.MockToolRegistry{}
	ra := agent.NewReportAgent(cllm, tools, "{locale}", "")
	reportAgents := map[agent.Product]*agent.ReportAgent{
		agent.ProductChart: ra,
	}
	svc := NewService(&mockPaymentProvider{}, &mockPaymentProvider{}, &mockEmailClient{}, store,
		"https://liki.hk", "admin@liki.hk", reportAgents, context.Background(),
	)

	svc.StartReportGeneration("order-no-locale", agent.ProductChart, `{"x":1}`)
	<-cllm.doneCh

	if !strings.Contains(cllm.lastSystemPrompt, "zh-Hans") {
		t.Errorf("lastSystemPrompt = %q, want containing 'zh-Hans' (default locale)", cllm.lastSystemPrompt)
	}
}

// -- RetryReportGeneration --

func TestRetryReportGeneration_PaidNoJSON(t *testing.T) {
	cllm := &controllableLLM{
		content: "<p>recovered</p>",
		doneCh:  make(chan struct{}, 1),
	}
	svc, store := newTestSvcWithReportGen(t, cllm)

	// Mark the order as paid via webhook, but prevent report generation
	// so we can test the retry path (paid + empty llm_json).
	dodoMock := svc.Dodo.(*mockPaymentProvider)
	dodoMock.verifyEvent = &WebhookEvent{
		Type: "payment.succeeded",
		Data: WebhookEventData{OrderID: "order-1", PaymentID: "pay-1", Amount: 990},
	}
	origRA := svc.ReportAgents
	svc.ReportAgents = nil
	if err := svc.HandleWebhook(context.Background(), []byte(`{}`), http.Header{}); err != nil {
		t.Fatalf("HandleWebhook: %v", err)
	}
	svc.ReportAgents = origRA

	// Verify paid, no llm_json.
	order, err := store.GetOrder(context.Background(), "order-1")
	if err != nil {
		t.Fatalf("setup GetOrder: %v", err)
	}
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

	// Wait for background generation and DB write to complete.
	<-cllm.doneCh
	waitForCleanup(t, svc, "order-1")

	order, orderErr := store.GetOrder(context.Background(), "order-1")
	if orderErr != nil {
		t.Fatalf("GetOrder: %v", orderErr)
	}
	if order.LlmJSON != "<p>recovered</p>" {
		t.Errorf("llm_json = %q, want recovered (retry should trigger generation)", order.LlmJSON)
	}
}

func TestRetryReportGeneration_PaidWithJSON(t *testing.T) {
	cllm := &controllableLLM{
		content: "<p>existing</p>",
		doneCh:  make(chan struct{}, 2),
	}
	svc, _ := newTestSvcWithReportGen(t, cllm)

	// Mark as paid with llm_json via webhook.
	dodoMock := svc.Dodo.(*mockPaymentProvider)
	dodoMock.verifyEvent = &WebhookEvent{
		Type: "payment.succeeded",
		Data: WebhookEventData{OrderID: "order-1", PaymentID: "pay-1", Amount: 990},
	}
	if err := svc.HandleWebhook(context.Background(), []byte(`{}`), http.Header{}); err != nil {
		t.Fatalf("HandleWebhook: %v", err)
	}
	<-cllm.doneCh // wait for webhook-triggered generation
	waitForCleanup(t, svc, "order-1")

	// Now retry: must NOT trigger new generation since llm_json exists.
	genCountBefore := cllm.callCount()
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
	if cllm.callCount() != genCountBefore {
		t.Errorf("Generate called %d extra times, want 0 (should not regenerate when llm_json exists)", cllm.callCount()-genCountBefore)
	}
}

func TestRetryReportGeneration_Pending(t *testing.T) {
	cllm := &controllableLLM{content: "<p>nope</p>"}
	svc, _ := newTestSvcWithReportGen(t, cllm)

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
	if cllm.callCount() > 0 {
		t.Errorf("Generate called %d times, want 0 (pending order should not trigger)", cllm.callCount())
	}
}

func TestRetryReportGeneration_NotFound(t *testing.T) {
	cllm := &controllableLLM{}
	svc, _ := newTestSvcWithReportGen(t, cllm)

	_, _, _, err := svc.RetryReportGeneration(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing order")
	}
}
