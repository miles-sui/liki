package payment

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"sync"
	"testing"

	"liki/internal/product"

	_ "modernc.org/sqlite"
)

// -- mocks --

type mockPaymentProvider struct {
	createResult  *CheckoutResult
	createErr     error
	verifyEvent   *WebhookEvent
	verifyErr     error
	lastEmailSent string // email passed to CreateCheckout
}

func (m *mockPaymentProvider) CreateCheckout(_ context.Context, _ product.Product, _ int, _, email, _, _ string) (*CheckoutResult, error) {
	m.lastEmailSent = email
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

func newTestSvc(t *testing.T) (*Service, *Store, *mockPaymentProvider, *mockEmailClient) {
	t.Helper()
	db := newTestDB(t)
	store, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	if err := store.CreateOrder(context.Background(), "order-1", product.ProductNaming, 990, "CNY", "", `{"chart":"data"}`, "", ""); err != nil {
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
		"https://liki.hk", "admin@liki.hk", context.Background(),
	)
	return svc, store, dodoMock, emailMock
}

// -- CreateCheckout --

func TestCreateCheckout_Success(t *testing.T) {
	svc, _, _, _ := newTestSvc(t)
	result, err := svc.CreateCheckout(context.Background(), "dodo", "order-1", "", "")
	if err != nil {
		t.Fatalf("CreateCheckout: %v", err)
	}
	if result.CheckoutURL != "https://pay.example.com/checkout" {
		t.Errorf("CheckoutURL = %q", result.CheckoutURL)
	}
}

func TestCreateCheckout_Xunhu(t *testing.T) {
	svc, _, _, _ := newTestSvc(t)
	result, err := svc.CreateCheckout(context.Background(), "xunhu", "order-1", "", "")
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
	_, err := svc.CreateCheckout(context.Background(), "dodo", "nonexistent", "", "")
	if err == nil {
		t.Fatal("expected error for missing order")
	}
	if !errors.Is(err, ErrOrderNotFound) {
		t.Errorf("expected ErrOrderNotFound, got %v", err)
	}
}

func TestCreateCheckout_UnknownProvider(t *testing.T) {
	svc, _, _, _ := newTestSvc(t)
	_, err := svc.CreateCheckout(context.Background(), "unknown", "order-1", "", "")
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestCreateCheckout_ProviderError(t *testing.T) {
	svc, _, dodoMock, _ := newTestSvc(t)
	dodoMock.createErr = errors.New("api error")
	_, err := svc.CreateCheckout(context.Background(), "dodo", "order-1", "", "")
	if err == nil {
		t.Fatal("expected error from provider")
	}
}

func TestCreateCheckout_WithEmail(t *testing.T) {
	svc, store, _, _ := newTestSvc(t)
	_, err := svc.CreateCheckout(context.Background(), "dodo", "order-1", "user@example.com", "")
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

func TestCreateCheckout_EmailFallback(t *testing.T) {
	svc, store, dodoMock, _ := newTestSvc(t)
	if err := store.UpdateEmail(context.Background(), "order-1", "stored@example.com"); err != nil {
		t.Fatalf("UpdateEmail: %v", err)
	}

	result, err := svc.CreateCheckout(context.Background(), "dodo", "order-1", "", "")
	if err != nil {
		t.Fatalf("CreateCheckout: %v", err)
	}
	if result.CheckoutURL != "https://pay.example.com/checkout" {
		t.Errorf("CheckoutURL = %q", result.CheckoutURL)
	}
	if dodoMock.lastEmailSent != "stored@example.com" {
		t.Errorf("provider email = %q, want stored@example.com (fallback to order.Email)", dodoMock.lastEmailSent)
	}
}

// -- HandleWebhook --

func TestHandleWebhook_VerifyFailure(t *testing.T) {
	svc, _, dodoMock, _ := newTestSvc(t)
	xunhuMock := svc.xunhu.(*mockPaymentProvider)
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
	svc, store, dodoMock, emailMock := newTestSvc(t)
	emailMock.sent = make(chan struct{}, 4)

	if err := store.UpdateEmail(context.Background(), "order-1", "user@example.com"); err != nil {
		t.Fatalf("UpdateEmail: %v", err)
	}

	dodoMock.verifyErr = errors.New("dodo: bad signature")
	xunhuMock := svc.xunhu.(*mockPaymentProvider)
	xunhuMock.verifyEvent = &WebhookEvent{
		Type: EventPaymentSucceeded,
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
	xunhuMock := svc.xunhu.(*mockPaymentProvider)
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
		Type: EventPaymentSucceeded,
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

	if err := store.UpdateEmail(context.Background(), "order-1", "user@example.com"); err != nil {
		t.Fatalf("UpdateEmail: %v", err)
	}

	emailMock.sent = make(chan struct{}, 4)

	dodoMock.verifyEvent = &WebhookEvent{
		Type: EventPaymentSucceeded,
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
	svc, store, dodoMock, emailMock := newTestSvc(t)
	emailMock.sent = make(chan struct{}, 4)

	if err := store.UpdateEmail(context.Background(), "order-1", "user@example.com"); err != nil {
		t.Fatalf("UpdateEmail: %v", err)
	}

	dodoMock.verifyEvent = &WebhookEvent{
		Type: EventPaymentSucceeded,
		Data: WebhookEventData{OrderID: "order-1", PaymentID: "pay-1", Amount: 990},
	}
	if err := svc.HandleWebhook(context.Background(), []byte(`{}`), http.Header{}); err != nil {
		t.Fatalf("first HandleWebhook: %v", err)
	}
	<-emailMock.sent // customer
	<-emailMock.sent // admin
	emailCount := emailMock.sentToCount()

	dodoMock.verifyEvent.Data.PaymentID = "pay-2"
	if err := svc.HandleWebhook(context.Background(), []byte(`{}`), http.Header{}); err != nil {
		t.Fatalf("second HandleWebhook: %v", err)
	}

	if emailMock.sentToCount() != emailCount {
		t.Errorf("second payment sent extra emails: %d before, %d after",
			emailCount, emailMock.sentToCount())
	}

	order, err := store.GetOrder(context.Background(), "order-1")
	if err != nil {
		t.Fatalf("GetOrder: %v", err)
	}
	if order.PaymentID != "pay-1" {
		t.Errorf("PaymentID = %q, want pay-1 (first payment preserved)", order.PaymentID)
	}
}

func TestHandleWebhook_DuplicatePaymentIdempotent(t *testing.T) {
	svc, store, dodoMock, emailMock := newTestSvc(t)
	emailMock.sent = make(chan struct{}, 4)

	if err := store.UpdateEmail(context.Background(), "order-1", "user@example.com"); err != nil {
		t.Fatalf("UpdateEmail: %v", err)
	}

	dodoMock.verifyEvent = &WebhookEvent{
		Type: EventPaymentSucceeded,
		Data: WebhookEventData{OrderID: "order-1", PaymentID: "pay-1", Amount: 990},
	}

	if err := svc.HandleWebhook(context.Background(), []byte(`{}`), http.Header{}); err != nil {
		t.Fatalf("first HandleWebhook: %v", err)
	}
	<-emailMock.sent // customer email
	<-emailMock.sent // admin email
	emailCount := emailMock.sentToCount()
	if emailCount < 2 {
		t.Errorf("first webhook: expected >=2 emails, got %d", emailCount)
	}

	if err := svc.HandleWebhook(context.Background(), []byte(`{}`), http.Header{}); err != nil {
		t.Fatalf("second HandleWebhook with same PaymentID: %v", err)
	}
	if emailMock.sentToCount() != emailCount {
		t.Errorf("duplicate webhook sent extra emails: %d before, %d after",
			emailCount, emailMock.sentToCount())
	}

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

	dodoMock.verifyEvent = &WebhookEvent{
		Type: EventPaymentSucceeded,
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

// -- GetOrderData (status queries) --

func TestGetOrderData_StatusFound(t *testing.T) {
	svc, _, _, _ := newTestSvc(t)
	status, prod, _, err := svc.GetOrderData(context.Background(), "order-1")
	if err != nil {
		t.Fatalf("GetOrderData: %v", err)
	}
	if status != OrderPending {
		t.Errorf("status = %s, want pending", status)
	}
	if prod != product.ProductNaming {
		t.Errorf("product = %s, want naming", prod)
	}
}

func TestGetOrderData_StatusNotFound(t *testing.T) {
	svc, _, _, _ := newTestSvc(t)
	_, _, _, err := svc.GetOrderData(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing order")
	}
}

func TestShutdown_CompletesWithoutGoroutines(t *testing.T) {
	svc, _, _, _ := newTestSvc(t)
	// Shutdown without any in-flight goroutines should return immediately.
	if err := svc.Shutdown(context.Background()); err != nil {
		t.Errorf("Shutdown: %v", err)
	}
}

func TestShutdown_WaitsForInFlightEmails(t *testing.T) {
	svc, store, dodoMock, emailMock := newTestSvc(t)

	if err := store.UpdateEmail(context.Background(), "order-1", "user@example.com"); err != nil {
		t.Fatalf("UpdateEmail: %v", err)
	}

	emailMock.sent = make(chan struct{}, 4)

	dodoMock.verifyEvent = &WebhookEvent{
		Type: EventPaymentSucceeded,
		Data: WebhookEventData{OrderID: "order-1", PaymentID: "pay-1", Amount: 990},
	}
	if err := svc.HandleWebhook(context.Background(), []byte(`{}`), http.Header{}); err != nil {
		t.Fatalf("HandleWebhook: %v", err)
	}
	// Drain both email signals — after goroutines finish, Shutdown returns.
	<-emailMock.sent
	<-emailMock.sent

	if err := svc.Shutdown(context.Background()); err != nil {
		t.Errorf("Shutdown: %v", err)
	}
}

func TestGetOrderData_Found(t *testing.T) {
	svc, _, _, _ := newTestSvc(t)
	status, prod, llmJSON, err := svc.GetOrderData(context.Background(), "order-1")
	if err != nil {
		t.Fatalf("GetOrderData: %v", err)
	}
	if status != OrderPending {
		t.Errorf("status = %s, want pending", status)
	}
	if prod != product.ProductNaming {
		t.Errorf("product = %s, want naming", prod)
	}
	if llmJSON != "" {
		t.Errorf("llmJSON should be empty for fresh order, got %q", llmJSON)
	}
}

func TestGetOrderData_NotFound(t *testing.T) {
	svc, _, _, _ := newTestSvc(t)
	_, _, _, err := svc.GetOrderData(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing order")
	}
}
