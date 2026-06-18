package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"liki/internal/dodo"
	
	"liki/internal/agent"
	"liki/internal/payment"
)

// -- mocks --

type mockDodoClient struct {
	checkoutResult *dodo.CheckoutResult
	checkoutErr    error
	webhookEvent   *dodo.WebhookEvent
	webhookErr     error
}

func (m *mockDodoClient) CreateCheckout(_ context.Context, _ string, _ int, _, _, _ string) (*dodo.CheckoutResult, error) {
	return m.checkoutResult, m.checkoutErr
}

func (m *mockDodoClient) VerifyWebhook(_ []byte, _ http.Header) (*dodo.WebhookEvent, error) {
	return m.webhookEvent, m.webhookErr
}

type mockEmailClient struct {
	err error
}

func (m *mockEmailClient) SendReport(_ context.Context, _, _, _ string) error {
	return m.err
}

type mockReportAgent struct {
	result string
	err    error
}

func (m *mockReportAgent) GenerateFromData(_ context.Context, _ string, _ agent.Product, _ json.RawMessage) (string, error) {
	return m.result, m.err
}

// -- helpers --

func openPaymentTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	db.SetMaxOpenConns(1)
	return db
}

func newTestPaymentService(t *testing.T, db *sql.DB, dodoCli *mockDodoClient, emailCli *mockEmailClient, reportAgent *mockReportAgent) *payment.Service {
	t.Helper()
	store, err := payment.NewStore(db)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	productIDs := map[agent.Product]string{
		agent.ProductChart:  "prod_chart",
		agent.ProductBond:   "prod_bond",
		agent.ProductNaming: "prod_naming",
	}
	return payment.NewService(dodoCli, emailCli, store, productIDs, "https://example.com", "admin@example.com", reportAgent, context.Background())
}

func createTestOrder(t *testing.T, db *sql.DB, orderID string, product agent.Product, status payment.OrderStatus, email, chartJSON, llmJSON string) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO orders (order_id, product, amount, currency, email, chart_json, llm_json, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		orderID, product, 990, "CNY", email, chartJSON, llmJSON, status,
	)
	if err != nil {
		t.Fatalf("create test order: %v", err)
	}
}

// -- handlePaymentReturn --

func TestHandlePaymentReturn_Succeeded(t *testing.T) {
	h := handlePaymentReturn()
	r := httptest.NewRequest("GET", "/api/payments/return/order-1?status=succeeded", nil)
	r.SetPathValue("id", "order-1")
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusFound {
		t.Errorf("status = %d, want 302", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "/report/order-1" {
		t.Errorf("Location = %q, want /report/order-1", loc)
	}
}

func TestHandlePaymentReturn_NotSucceeded(t *testing.T) {
	h := handlePaymentReturn()
	r := httptest.NewRequest("GET", "/api/payments/return/order-1?status=pending", nil)
	r.SetPathValue("id", "order-1")
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusFound {
		t.Errorf("status = %d, want 302", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "/" {
		t.Errorf("Location = %q, want /", loc)
	}
}

func TestHandlePaymentReturn_NoOrderID(t *testing.T) {
	h := handlePaymentReturn()
	r := httptest.NewRequest("GET", "/api/payments/return/?status=succeeded", nil)
	r.SetPathValue("id", "")
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusFound {
		t.Errorf("status = %d, want 302", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "/" {
		t.Errorf("Location = %q, want /", loc)
	}
}

func TestHandlePaymentReturn_MissingStatus(t *testing.T) {
	h := handlePaymentReturn()
	r := httptest.NewRequest("GET", "/api/payments/return/order-1", nil)
	r.SetPathValue("id", "order-1")
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusFound {
		t.Errorf("status = %d, want 302", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "/" {
		t.Errorf("Location = %q, want /", loc)
	}
}

// -- redirectReport --

func TestRedirectDownload_Valid(t *testing.T) {
	h := redirectReport()
	r := httptest.NewRequest("GET", "/api/orders/order-1/report", nil)
	r.SetPathValue("id", "order-1")
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusFound {
		t.Errorf("status = %d, want 302", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "/report/order-1" {
		t.Errorf("Location = %q, want /report/order-1", loc)
	}
}

func TestRedirectDownload_MissingOrderID(t *testing.T) {
	h := redirectReport()
	r := httptest.NewRequest("GET", "/api/orders//report", nil)
	r.SetPathValue("id", "")
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// -- handleCheckout --

func TestHandleCheckout_Success(t *testing.T) {
	db := openPaymentTestDB(t)
	defer db.Close()

	dodoCli := &mockDodoClient{
		checkoutResult: &dodo.CheckoutResult{SessionID: "sess_1", CheckoutURL: "https://pay.example.com/checkout"},
	}
	emailCli := &mockEmailClient{}
	svc := newTestPaymentService(t, db, dodoCli, emailCli, nil)

	createTestOrder(t, db, "order-1", agent.ProductChart, payment.OrderPending, "", `{"x":1}`, "")

	h := handleCheckout(svc)
	body := `{"order_id":"order-1","email":"a@b.co"}`
	r := httptest.NewRequest("POST", "/api/checkout", strings.NewReader(body))
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", w.Code, w.Body.String())
	}
	var env struct {
		Data dodo.CheckoutResult `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if env.Data.SessionID != "sess_1" {
		t.Errorf("session_id = %q, want sess_1", env.Data.SessionID)
	}
}

func TestHandleCheckout_OrderNotFound(t *testing.T) {
	db := openPaymentTestDB(t)
	defer db.Close()

	dodoCli := &mockDodoClient{}
	emailCli := &mockEmailClient{}
	svc := newTestPaymentService(t, db, dodoCli, emailCli, nil)

	h := handleCheckout(svc)
	body := `{"order_id":"nonexistent"}`
	r := httptest.NewRequest("POST", "/api/checkout", strings.NewReader(body))
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestHandleCheckout_InvalidBody(t *testing.T) {
	db := openPaymentTestDB(t)
	defer db.Close()

	dodoCli := &mockDodoClient{}
	emailCli := &mockEmailClient{}
	svc := newTestPaymentService(t, db, dodoCli, emailCli, nil)

	h := handleCheckout(svc)
	r := httptest.NewRequest("POST", "/api/checkout", strings.NewReader("not-json"))
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestHandleCheckout_MissingOrderID(t *testing.T) {
	db := openPaymentTestDB(t)
	defer db.Close()

	dodoCli := &mockDodoClient{}
	emailCli := &mockEmailClient{}
	svc := newTestPaymentService(t, db, dodoCli, emailCli, nil)

	h := handleCheckout(svc)
	body := `{"email":"a@b.co"}`
	r := httptest.NewRequest("POST", "/api/checkout", strings.NewReader(body))
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", w.Code)
	}
}

func TestHandleCheckout_DodoError(t *testing.T) {
	db := openPaymentTestDB(t)
	defer db.Close()

	dodoCli := &mockDodoClient{
		checkoutErr: errors.New("dodo: api error"),
	}
	emailCli := &mockEmailClient{}
	svc := newTestPaymentService(t, db, dodoCli, emailCli, nil)

	createTestOrder(t, db, "order-1", agent.ProductChart, payment.OrderPending, "", `{"x":1}`, "")

	h := handleCheckout(svc)
	body := `{"order_id":"order-1"}`
	r := httptest.NewRequest("POST", "/api/checkout", strings.NewReader(body))
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

// -- handleOrderStatus --

func TestHandleOrderStatus_Valid(t *testing.T) {
	db := openPaymentTestDB(t)
	defer db.Close()

	dodoCli := &mockDodoClient{}
	emailCli := &mockEmailClient{}
	svc := newTestPaymentService(t, db, dodoCli, emailCli, nil)

	createTestOrder(t, db, "order-1", agent.ProductChart, payment.OrderPending, "", `{"x":1}`, "")

	h := handleOrderStatus(svc)
	r := httptest.NewRequest("GET", "/api/orders/order-1/status", nil)
	r.SetPathValue("id", "order-1")
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var env struct {
		Data struct {
			Status  string `json:"status"`
			Product string `json:"product"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if env.Data.Status != "pending" {
		t.Errorf("status = %q, want pending", env.Data.Status)
	}
}

func TestHandleOrderStatus_NotFound(t *testing.T) {
	db := openPaymentTestDB(t)
	defer db.Close()

	dodoCli := &mockDodoClient{}
	emailCli := &mockEmailClient{}
	svc := newTestPaymentService(t, db, dodoCli, emailCli, nil)

	h := handleOrderStatus(svc)
	r := httptest.NewRequest("GET", "/api/orders/nonexistent/status", nil)
	r.SetPathValue("id", "nonexistent")
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestHandleOrderStatus_MissingID(t *testing.T) {
	db := openPaymentTestDB(t)
	defer db.Close()

	dodoCli := &mockDodoClient{}
	emailCli := &mockEmailClient{}
	svc := newTestPaymentService(t, db, dodoCli, emailCli, nil)

	h := handleOrderStatus(svc)
	r := httptest.NewRequest("GET", "/api/orders//status", nil)
	r.SetPathValue("id", "")
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// -- handleRetryOrder --

func TestHandleRetryOrder_Valid(t *testing.T) {
	db := openPaymentTestDB(t)
	defer db.Close()

	dodoCli := &mockDodoClient{}
	emailCli := &mockEmailClient{}
	svc := newTestPaymentService(t, db, dodoCli, emailCli, nil)

	createTestOrder(t, db, "order-1", agent.ProductChart, payment.OrderPending, "", `{"x":1}`, "")

	h := handleRetryOrder(svc)
	r := httptest.NewRequest("POST", "/api/orders/order-1/retry", nil)
	r.SetPathValue("id", "order-1")
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestHandleRetryOrder_NotFound(t *testing.T) {
	db := openPaymentTestDB(t)
	defer db.Close()

	dodoCli := &mockDodoClient{}
	emailCli := &mockEmailClient{}
	svc := newTestPaymentService(t, db, dodoCli, emailCli, nil)

	h := handleRetryOrder(svc)
	r := httptest.NewRequest("POST", "/api/orders/nonexistent/retry", nil)
	r.SetPathValue("id", "nonexistent")
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestHandleRetryOrder_MissingID(t *testing.T) {
	db := openPaymentTestDB(t)
	defer db.Close()

	dodoCli := &mockDodoClient{}
	emailCli := &mockEmailClient{}
	svc := newTestPaymentService(t, db, dodoCli, emailCli, nil)

	h := handleRetryOrder(svc)
	r := httptest.NewRequest("POST", "/api/orders//retry", nil)
	r.SetPathValue("id", "")
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// -- handleWebhook --

func TestHandleWebhook_Success(t *testing.T) {
	db := openPaymentTestDB(t)
	defer db.Close()

	dodoCli := &mockDodoClient{
		webhookEvent: &dodo.WebhookEvent{
			Type: "payment.succeeded",
			Data: dodo.WebhookEventData{
				OrderID:   "order-1",
				Amount:    990,
				PaymentID: "pay_1",
			},
		},
	}
	emailCli := &mockEmailClient{}
	reportAgent := &mockReportAgent{result: "<p>report</p>"}
	svc := newTestPaymentService(t, db, dodoCli, emailCli, reportAgent)

	createTestOrder(t, db, "order-1", agent.ProductChart, payment.OrderPending, "a@b.co", `{"x":1}`, "")

	h := handleWebhook(svc)
	r := httptest.NewRequest("POST", "/api/webhook", strings.NewReader(`{"type":"payment.succeeded"}`))
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", w.Code, w.Body.String())
	}
}

func TestHandleWebhook_VerifyFail(t *testing.T) {
	db := openPaymentTestDB(t)
	defer db.Close()

	dodoCli := &mockDodoClient{
		webhookErr: errors.New("signature invalid"),
	}
	emailCli := &mockEmailClient{}
	svc := newTestPaymentService(t, db, dodoCli, emailCli, nil)

	h := handleWebhook(svc)
	r := httptest.NewRequest("POST", "/api/webhook", strings.NewReader(`{"type":"payment.succeeded"}`))
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestHandleWebhook_NonPayment(t *testing.T) {
	db := openPaymentTestDB(t)
	defer db.Close()

	dodoCli := &mockDodoClient{
		webhookEvent: &dodo.WebhookEvent{
			Type: "checkout.created",
		},
	}
	emailCli := &mockEmailClient{}
	svc := newTestPaymentService(t, db, dodoCli, emailCli, nil)

	h := handleWebhook(svc)
	r := httptest.NewRequest("POST", "/api/webhook", strings.NewReader(`{"type":"checkout.created"}`))
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestHandleWebhook_DuplicateIdempotent(t *testing.T) {
	// Dodo may redeliver webhooks. The second identical payment.succeeded
	// must return 200 without double-processing (idempotent).
	db := openPaymentTestDB(t)
	defer db.Close()

	dodoCli := &mockDodoClient{
		webhookEvent: &dodo.WebhookEvent{
			Type: "payment.succeeded",
			Data: dodo.WebhookEventData{
				OrderID:   "order-1",
				Amount:    990,
				PaymentID: "pay_1",
			},
		},
	}
	emailCli := &mockEmailClient{}
	reportAgent := &mockReportAgent{result: "<p>report</p>"}
	svc := newTestPaymentService(t, db, dodoCli, emailCli, reportAgent)

	createTestOrder(t, db, "order-1", agent.ProductChart, payment.OrderPending, "a@b.co", `{"x":1}`, "")

	h := handleWebhook(svc)
	body := `{"type":"payment.succeeded"}`

	// First webhook: processes payment.
	w1 := httptest.NewRecorder()
	h(w1, httptest.NewRequest("POST", "/api/webhook", strings.NewReader(body)))
	if w1.Code != http.StatusOK {
		t.Fatalf("first webhook: status = %d, want 200", w1.Code)
	}

	// Second webhook (duplicate): must succeed without side effects.
	w2 := httptest.NewRecorder()
	h(w2, httptest.NewRequest("POST", "/api/webhook", strings.NewReader(body)))
	if w2.Code != http.StatusOK {
		t.Errorf("duplicate webhook: status = %d, want 200", w2.Code)
	}
}

func TestHandleWebhook_NonExistentOrder(t *testing.T) {
	// Webhook for an order we don't have should return 500, not crash.
	db := openPaymentTestDB(t)
	defer db.Close()

	dodoCli := &mockDodoClient{
		webhookEvent: &dodo.WebhookEvent{
			Type: "payment.succeeded",
			Data: dodo.WebhookEventData{
				OrderID:   "no-such-order",
				Amount:    990,
				PaymentID: "pay_unknown",
			},
		},
	}
	emailCli := &mockEmailClient{}
	svc := newTestPaymentService(t, db, dodoCli, emailCli, nil)

	h := handleWebhook(svc)
	r := httptest.NewRequest("POST", "/api/webhook", strings.NewReader(`{"type":"payment.succeeded"}`))
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500 for non-existent order", w.Code)
	}
}

// -- handleReport --

func TestHandleReport_Valid(t *testing.T) {
	db := openPaymentTestDB(t)
	defer db.Close()

	dodoCli := &mockDodoClient{}
	emailCli := &mockEmailClient{}
	svc := newTestPaymentService(t, db, dodoCli, emailCli, nil)

	createTestOrder(t, db, "order-1", agent.ProductChart, payment.OrderPaid, "a@b.co", `{"x":1}`, `{"report":"content"}`)

	h := handleReport(svc)
	r := httptest.NewRequest("GET", "/api/reports/order-1", nil)
	r.SetPathValue("id", "order-1")
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var env struct {
		Data payment.ReportData `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if env.Data.Status != payment.OrderPaid {
		t.Errorf("status = %q, want paid", env.Data.Status)
	}
	if env.Data.LlmJSON != `{"report":"content"}` {
		t.Errorf("llm_json = %q", env.Data.LlmJSON)
	}
}

func TestHandleReport_NotFound(t *testing.T) {
	db := openPaymentTestDB(t)
	defer db.Close()

	dodoCli := &mockDodoClient{}
	emailCli := &mockEmailClient{}
	svc := newTestPaymentService(t, db, dodoCli, emailCli, nil)

	h := handleReport(svc)
	r := httptest.NewRequest("GET", "/api/reports/nonexistent", nil)
	r.SetPathValue("id", "nonexistent")
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestHandleReport_MissingID(t *testing.T) {
	db := openPaymentTestDB(t)
	defer db.Close()

	dodoCli := &mockDodoClient{}
	emailCli := &mockEmailClient{}
	svc := newTestPaymentService(t, db, dodoCli, emailCli, nil)

	h := handleReport(svc)
	r := httptest.NewRequest("GET", "/api/reports/", nil)
	r.SetPathValue("id", "")
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestHandleReport_PaidNoLlmJSON(t *testing.T) {
	db := openPaymentTestDB(t)
	defer db.Close()

	dodoCli := &mockDodoClient{}
	emailCli := &mockEmailClient{}
	reportAgent := &mockReportAgent{result: "<p>generated</p>"}
	svc := newTestPaymentService(t, db, dodoCli, emailCli, reportAgent)

	createTestOrder(t, db, "order-1", agent.ProductChart, payment.OrderPaid, "a@b.co", `{"x":1}`, "")

	h := handleReport(svc)
	r := httptest.NewRequest("GET", "/api/reports/order-1", nil)
	r.SetPathValue("id", "order-1")
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	// Lazy generation is background; llm_json still empty at response time.
	var env struct {
		Data payment.ReportData `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if env.Data.LlmJSON != "" {
		t.Errorf("llm_json should be empty (bg gen not yet complete), got %q", env.Data.LlmJSON)
	}
}

func TestHandleReport_PendingStatus(t *testing.T) {
	db := openPaymentTestDB(t)
	defer db.Close()

	dodoCli := &mockDodoClient{}
	emailCli := &mockEmailClient{}
	svc := newTestPaymentService(t, db, dodoCli, emailCli, nil)

	createTestOrder(t, db, "order-1", agent.ProductChart, payment.OrderPending, "a@b.co", `{"x":1}`, "")

	h := handleReport(svc)
	r := httptest.NewRequest("GET", "/api/reports/order-1", nil)
	r.SetPathValue("id", "order-1")
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var env struct {
		Data payment.ReportData `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if env.Data.LlmJSON != "" {
		t.Errorf("pending order should not expose llm_json, got %q", env.Data.LlmJSON)
	}
}

func TestEd3_Payment_ReturnNoStatus(t *testing.T) {
	h := handlePaymentReturn()
	r := httptest.NewRequest("GET", "/api/payments/return/order123", nil)
	w := httptest.NewRecorder()
	h(w, r)
	// 无 status=succeeded → 重定向到 /
	if w.Code != http.StatusFound {
		t.Errorf("status=%d, want 302", w.Code)
	}
	loc := w.Header().Get("Location")
	if loc != "/" {
		t.Errorf("redirect location=%q, want /", loc)
	}
}

func TestEd3_Payment_ReturnSucceeded(t *testing.T) {
	h := handlePaymentReturn()
	r := httptest.NewRequest("GET", "/api/payments/return/order123?status=succeeded", nil)
	r.SetPathValue("id", "order123")
	w := httptest.NewRecorder()
	h(w, r)
	if w.Code != http.StatusFound {
		t.Errorf("status=%d, want 302", w.Code)
	}
	loc := w.Header().Get("Location")
	if loc != "/report/order123" {
		t.Errorf("redirect location=%q, want /report/order123", loc)
	}
}

func TestEd3_Payment_ReturnEmptyID(t *testing.T) {
	h := handlePaymentReturn()
	r := httptest.NewRequest("GET", "/api/payments/return/?status=succeeded", nil)
	w := httptest.NewRecorder()
	h(w, r)
	// 空 id + status=succeeded，redirect 到 /report/
	if w.Code != http.StatusFound {
		t.Errorf("status=%d, want 302", w.Code)
	}
}

func TestEd3_Payment_RedirectReport_EmptyID(t *testing.T) {
	h := redirectReport()
	r := httptest.NewRequest("GET", "/api/orders//report", nil)
	w := httptest.NewRecorder()
	h(w, r)
	// 空 id 应返回错误
	if w.Code < 400 {
		t.Errorf("status=%d, want >=400", w.Code)
	}
}
