package http

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"liki/internal/product"
	"liki/internal/payment"
)

// -- mocks --

type mockPaymentProvider struct {
	checkoutResult *payment.CheckoutResult
	checkoutErr    error
	webhookEvent   *payment.WebhookEvent
	webhookErr     error
}

func (m *mockPaymentProvider) CreateCheckout(_ context.Context, _ product.Product, _ int, _, _, _ string) (*payment.CheckoutResult, error) {
	return m.checkoutResult, m.checkoutErr
}

func (m *mockPaymentProvider) VerifyWebhook(_ []byte, _ http.Header) (*payment.WebhookEvent, error) {
	return m.webhookEvent, m.webhookErr
}

type mockEmailClient struct {
	err error
}

func (m *mockEmailClient) SendReport(_ context.Context, _, _, _ string) error {
	return m.err
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

func newTestPaymentService(t *testing.T, db *sql.DB, dodoCli, xunhuCli *mockPaymentProvider, emailCli *mockEmailClient) *payment.Service {
	t.Helper()
	store, err := payment.NewStore(db)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	return payment.NewService(dodoCli, xunhuCli, emailCli, store, "https://example.com", "admin@example.com", context.Background())
}

func createTestOrder(t *testing.T, db *sql.DB, orderID string, product product.Product, status payment.OrderStatus, email, chartJSON, llmJSON string) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO orders (order_id, product, amount, currency, provider, email, chart_json, llm_json, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		orderID, product, 990, "CNY", "", email, chartJSON, llmJSON, status,
	)
	if err != nil {
		t.Fatalf("create test order: %v", err)
	}
}

// -- handleCreateOrder --

func TestHandleCreateOrder_Success(t *testing.T) {
	db := openPaymentTestDB(t)
	defer db.Close()

	store, err := payment.NewStore(db)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	h := handleCreateOrder(store)
	body := `{"email":"a@b.co","product":"naming","currency":"USD"}`
	r := httptest.NewRequest("POST", "/api/orders", strings.NewReader(body))
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", w.Code, w.Body.String())
	}
	var env struct {
		Data struct {
			OrderID string `json:"order_id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if env.Data.OrderID == "" {
		t.Fatal("order_id is empty")
	}

	// Verify order exists in DB with correct data
	order, err := store.GetOrder(r.Context(), env.Data.OrderID)
	if err != nil {
		t.Fatalf("GetOrder: %v", err)
	}
	if order.Email != "a@b.co" {
		t.Errorf("email = %q, want a@b.co", order.Email)
	}
	if order.Product != product.ProductNaming {
		t.Errorf("product = %q, want naming", order.Product)
	}
	if order.Currency != "USD" {
		t.Errorf("currency = %q, want USD", order.Currency)
	}
	if order.Status != payment.OrderPending {
		t.Errorf("status = %q, want pending", order.Status)
	}
}

func TestHandleCreateOrder_CNY(t *testing.T) {
	db := openPaymentTestDB(t)
	defer db.Close()

	store, err := payment.NewStore(db)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	h := handleCreateOrder(store)
	body := `{"email":"a@b.co","product":"naming","currency":"CNY"}`
	r := httptest.NewRequest("POST", "/api/orders", strings.NewReader(body))
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", w.Code, w.Body.String())
	}
	var env struct {
		Data struct {
			OrderID string `json:"order_id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	order, err := store.GetOrder(r.Context(), env.Data.OrderID)
	if err != nil {
		t.Fatalf("GetOrder: %v", err)
	}
	if order.Amount != 9900 {
		t.Errorf("amount = %d, want 9900 (¥99.00)", order.Amount)
	}
}

func TestHandleCreateOrder_ValidationErrors(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantErr    string
	}{
		{"missing email", `{"product":"naming","currency":"USD"}`, http.StatusUnprocessableEntity, "Invalid request parameters"},
		{"missing product", `{"email":"a@b.co","currency":"USD"}`, http.StatusUnprocessableEntity, "Invalid request parameters"},
		{"missing currency", `{"email":"a@b.co","product":"naming"}`, http.StatusUnprocessableEntity, "Invalid request parameters"},
		{"Invalid request parameters", `{"email":"not-an-email","product":"naming","currency":"USD"}`, http.StatusUnprocessableEntity, "Invalid request parameters"},
		{"invalid product", `{"email":"a@b.co","product":"chart","currency":"USD"}`, http.StatusUnprocessableEntity, "Invalid request parameters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := openPaymentTestDB(t)
			defer db.Close()

			store, err := payment.NewStore(db)
			if err != nil {
				t.Fatalf("new store: %v", err)
			}

			h := handleCreateOrder(store)
			r := httptest.NewRequest("POST", "/api/orders", strings.NewReader(tt.body))
			w := httptest.NewRecorder()
			h(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if !strings.Contains(w.Body.String(), tt.wantErr) {
				t.Errorf("body = %s, want contain %q", w.Body.String(), tt.wantErr)
			}
		})
	}
}

func TestHandleCreateOrder_UnsupportedCurrency(t *testing.T) {
	db := openPaymentTestDB(t)
	defer db.Close()

	store, err := payment.NewStore(db)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	h := handleCreateOrder(store)
	body := `{"email":"a@b.co","product":"naming","currency":"EUR"}`
	r := httptest.NewRequest("POST", "/api/orders", strings.NewReader(body))
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
	if !strings.Contains(w.Body.String(), "unsupported currency") {
		t.Errorf("body = %s, want contain unsupported currency", w.Body.String())
	}
}

func TestHandleCreateOrder_InvalidBody(t *testing.T) {
	db := openPaymentTestDB(t)
	defer db.Close()

	store, err := payment.NewStore(db)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	h := handleCreateOrder(store)
	r := httptest.NewRequest("POST", "/api/orders", strings.NewReader("not-json"))
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// -- handlePaymentReturn --

func TestHandlePaymentReturn_Succeeded(t *testing.T) {
	db := openPaymentTestDB(t)
	store, err := payment.NewStore(db)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	createTestOrder(t, db, "order-1", product.ProductNaming, payment.OrderPaid, "user@example.com", "", "")

	h := handlePaymentReturn(store)
	r := httptest.NewRequest("GET", "/api/payments/return/order-1?status=succeeded", nil)
	r.SetPathValue("id", "order-1")
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusFound {
		t.Errorf("status = %d, want 302", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "/chat?order_id=order-1" {
		t.Errorf("Location = %q, want /chat?order_id=order-1", loc)
	}
}

func TestHandlePaymentReturn_NotSucceeded(t *testing.T) {
	db := openPaymentTestDB(t)
	store, err := payment.NewStore(db)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	h := handlePaymentReturn(store)
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
	db := openPaymentTestDB(t)
	store, err := payment.NewStore(db)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	h := handlePaymentReturn(store)
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
	db := openPaymentTestDB(t)
	store, err := payment.NewStore(db)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	h := handlePaymentReturn(store)
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

	dodoCli := &mockPaymentProvider{
		checkoutResult: &payment.CheckoutResult{SessionID: "sess_1", CheckoutURL: "https://pay.example.com/checkout"},
	}
	xunhuCli := &mockPaymentProvider{}
	emailCli := &mockEmailClient{}
	svc := newTestPaymentService(t, db, dodoCli, xunhuCli, emailCli)

	createTestOrder(t, db, "order-1", product.ProductNaming, payment.OrderPending, "", `{"x":1}`, "")

	h := handleCheckout(svc, &Analytics{})
	body := `{"order_id":"order-1","email":"a@b.co","provider":"dodo"}`
	r := httptest.NewRequest("POST", "/api/checkout", strings.NewReader(body))
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", w.Code, w.Body.String())
	}
	var env struct {
		Data payment.CheckoutResult `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if env.Data.SessionID != "sess_1" {
		t.Errorf("session_id = %q, want sess_1", env.Data.SessionID)
	}
}

func TestHandleCheckout_Xunhu(t *testing.T) {
	db := openPaymentTestDB(t)
	defer db.Close()

	dodoCli := &mockPaymentProvider{}
	xunhuCli := &mockPaymentProvider{
		checkoutResult: &payment.CheckoutResult{SessionID: "sess_x", CheckoutURL: "https://pay.xunhu.com/checkout", QRCodeURL: "https://pay.xunhu.com/qr"},
	}
	emailCli := &mockEmailClient{}
	svc := newTestPaymentService(t, db, dodoCli, xunhuCli, emailCli)

	createTestOrder(t, db, "order-1", product.ProductNaming, payment.OrderPending, "", `{"x":1}`, "")

	h := handleCheckout(svc, &Analytics{})
	body := `{"order_id":"order-1","email":"a@b.co","provider":"xunhu"}`
	r := httptest.NewRequest("POST", "/api/checkout", strings.NewReader(body))
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", w.Code, w.Body.String())
	}
	var env struct {
		Data payment.CheckoutResult `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if env.Data.QRCodeURL != "https://pay.xunhu.com/qr" {
		t.Errorf("qrcode_url = %q, want https://pay.xunhu.com/qr", env.Data.QRCodeURL)
	}
}

func TestHandleCheckout_OrderNotFound(t *testing.T) {
	db := openPaymentTestDB(t)
	defer db.Close()

	dodoCli := &mockPaymentProvider{}
	xunhuCli := &mockPaymentProvider{}
	emailCli := &mockEmailClient{}
	svc := newTestPaymentService(t, db, dodoCli, xunhuCli, emailCli)

	h := handleCheckout(svc, &Analytics{})
	body := `{"order_id":"nonexistent","email":"a@b.co"}`
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

	dodoCli := &mockPaymentProvider{}
	xunhuCli := &mockPaymentProvider{}
	emailCli := &mockEmailClient{}
	svc := newTestPaymentService(t, db, dodoCli, xunhuCli, emailCli)

	h := handleCheckout(svc, &Analytics{})
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

	dodoCli := &mockPaymentProvider{}
	xunhuCli := &mockPaymentProvider{}
	emailCli := &mockEmailClient{}
	svc := newTestPaymentService(t, db, dodoCli, xunhuCli, emailCli)

	h := handleCheckout(svc, &Analytics{})
	body := `{"email":"a@b.co"}`
	r := httptest.NewRequest("POST", "/api/checkout", strings.NewReader(body))
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", w.Code)
	}
}

func TestHandleCheckout_WithoutEmail(t *testing.T) {
	db := openPaymentTestDB(t)
	defer db.Close()

	dodoCli := &mockPaymentProvider{
		checkoutResult: &payment.CheckoutResult{SessionID: "sess_1", CheckoutURL: "https://pay.example.com/checkout"},
	}
	xunhuCli := &mockPaymentProvider{}
	emailCli := &mockEmailClient{}
	svc := newTestPaymentService(t, db, dodoCli, xunhuCli, emailCli)

	// Order has email stored from createOrder, checkout should use it
	createTestOrder(t, db, "order-1", product.ProductNaming, payment.OrderPending, "stored@example.com", `{"x":1}`, "")

	h := handleCheckout(svc, &Analytics{})
	body := `{"order_id":"order-1","provider":"dodo"}`
	r := httptest.NewRequest("POST", "/api/checkout", strings.NewReader(body))
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", w.Code, w.Body.String())
	}
	var env struct {
		Data payment.CheckoutResult `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if env.Data.SessionID != "sess_1" {
		t.Errorf("session_id = %q, want sess_1", env.Data.SessionID)
	}
}

func TestHandleCheckout_ProviderError(t *testing.T) {
	db := openPaymentTestDB(t)
	defer db.Close()

	dodoCli := &mockPaymentProvider{
		checkoutErr: errors.New("api error"),
	}
	xunhuCli := &mockPaymentProvider{}
	emailCli := &mockEmailClient{}
	svc := newTestPaymentService(t, db, dodoCli, xunhuCli, emailCli)

	createTestOrder(t, db, "order-1", product.ProductNaming, payment.OrderPending, "", `{"x":1}`, "")

	h := handleCheckout(svc, &Analytics{})
	body := `{"order_id":"order-1","email":"a@b.co","provider":"dodo"}`
	r := httptest.NewRequest("POST", "/api/checkout", strings.NewReader(body))
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

func TestHandleCheckout_AutoProvider(t *testing.T) {
	tests := []struct {
		name      string
		ipCountry string
		wantCallX bool // xunhu called, dodo not
	}{
		{"CN selects xunhu", "CN", true},
		{"US selects dodo", "US", false},
		{"empty defaults to xunhu", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := openPaymentTestDB(t)
			defer db.Close()

			dodoCli := &mockPaymentProvider{
				checkoutResult: &payment.CheckoutResult{SessionID: "dodo_sess", CheckoutURL: "https://dodo.example.com/checkout"},
			}
			xunhuCli := &mockPaymentProvider{
				checkoutResult: &payment.CheckoutResult{SessionID: "xunhu_sess", CheckoutURL: "https://xunhu.example.com/checkout"},
			}
			emailCli := &mockEmailClient{}
			svc := newTestPaymentService(t, db, dodoCli, xunhuCli, emailCli)

			createTestOrder(t, db, "order-1", product.ProductNaming, payment.OrderPending, "", `{"x":1}`, "")

			h := handleCheckout(svc, &Analytics{})
			body := `{"order_id":"order-1","email":"a@b.co"}` // no provider
			r := httptest.NewRequest("POST", "/api/checkout", strings.NewReader(body))
			if tt.ipCountry != "" {
				r.Header.Set("CF-IPCountry", tt.ipCountry)
			}
			w := httptest.NewRecorder()
			h(w, r)

			if w.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200, body: %s", w.Code, w.Body.String())
			}
			var env struct {
				Data payment.CheckoutResult `json:"data"`
			}
			if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if tt.wantCallX {
				if env.Data.SessionID != "xunhu_sess" {
					t.Errorf("session_id = %q, want xunhu_sess", env.Data.SessionID)
				}
			} else {
				if env.Data.SessionID != "dodo_sess" {
					t.Errorf("session_id = %q, want dodo_sess", env.Data.SessionID)
				}
			}
		})
	}
}

// -- handleOrderStatus --

func TestHandleOrderStatus_Valid(t *testing.T) {
	db := openPaymentTestDB(t)
	defer db.Close()

	store, err := payment.NewStore(db)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	createTestOrder(t, db, "order-1", product.ProductNaming, payment.OrderPending, "", `{"x":1}`, "")

	h := handleOrderStatus(store)
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

	store, err := payment.NewStore(db)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	h := handleOrderStatus(store)
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

	store, err := payment.NewStore(db)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	h := handleOrderStatus(store)
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

	dodoCli := &mockPaymentProvider{}
	xunhuCli := &mockPaymentProvider{}
	emailCli := &mockEmailClient{}
	svc := newTestPaymentService(t, db, dodoCli, xunhuCli, emailCli)

	createTestOrder(t, db, "order-1", product.ProductNaming, payment.OrderPending, "", `{"x":1}`, "")

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

	dodoCli := &mockPaymentProvider{}
	xunhuCli := &mockPaymentProvider{}
	emailCli := &mockEmailClient{}
	svc := newTestPaymentService(t, db, dodoCli, xunhuCli, emailCli)

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

	dodoCli := &mockPaymentProvider{}
	xunhuCli := &mockPaymentProvider{}
	emailCli := &mockEmailClient{}
	svc := newTestPaymentService(t, db, dodoCli, xunhuCli, emailCli)

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

	dodoCli := &mockPaymentProvider{
		webhookEvent: &payment.WebhookEvent{
			Type: "payment.succeeded",
			Data: payment.WebhookEventData{
				OrderID:   "order-1",
				Amount:    990,
				PaymentID: "pay_1",
			},
		},
	}
	xunhuCli := &mockPaymentProvider{}
	emailCli := &mockEmailClient{}
	svc := newTestPaymentService(t, db, dodoCli, xunhuCli, emailCli)

	createTestOrder(t, db, "order-1", product.ProductNaming, payment.OrderPending, "a@b.co", `{"x":1}`, "")

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

	dodoCli := &mockPaymentProvider{
		webhookErr: errors.New("signature invalid"),
	}
	xunhuCli := &mockPaymentProvider{}
	emailCli := &mockEmailClient{}
	svc := newTestPaymentService(t, db, dodoCli, xunhuCli, emailCli)

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

	dodoCli := &mockPaymentProvider{
		webhookEvent: &payment.WebhookEvent{
			Type: "checkout.created",
		},
	}
	xunhuCli := &mockPaymentProvider{}
	emailCli := &mockEmailClient{}
	svc := newTestPaymentService(t, db, dodoCli, xunhuCli, emailCli)

	h := handleWebhook(svc)
	r := httptest.NewRequest("POST", "/api/webhook", strings.NewReader(`{"type":"checkout.created"}`))
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestHandleWebhook_DuplicateIdempotent(t *testing.T) {
	db := openPaymentTestDB(t)
	defer db.Close()

	dodoCli := &mockPaymentProvider{
		webhookEvent: &payment.WebhookEvent{
			Type: "payment.succeeded",
			Data: payment.WebhookEventData{
				OrderID:   "order-1",
				Amount:    990,
				PaymentID: "pay_1",
			},
		},
	}
	xunhuCli := &mockPaymentProvider{}
	emailCli := &mockEmailClient{}
	svc := newTestPaymentService(t, db, dodoCli, xunhuCli, emailCli)

	createTestOrder(t, db, "order-1", product.ProductNaming, payment.OrderPending, "a@b.co", `{"x":1}`, "")

	h := handleWebhook(svc)
	body := `{"type":"payment.succeeded"}`

	w1 := httptest.NewRecorder()
	h(w1, httptest.NewRequest("POST", "/api/webhook", strings.NewReader(body)))
	if w1.Code != http.StatusOK {
		t.Fatalf("first webhook: status = %d, want 200", w1.Code)
	}

	w2 := httptest.NewRecorder()
	h(w2, httptest.NewRequest("POST", "/api/webhook", strings.NewReader(body)))
	if w2.Code != http.StatusOK {
		t.Errorf("duplicate webhook: status = %d, want 200", w2.Code)
	}
}

func TestHandleWebhook_NonExistentOrder(t *testing.T) {
	db := openPaymentTestDB(t)
	defer db.Close()

	dodoCli := &mockPaymentProvider{
		webhookEvent: &payment.WebhookEvent{
			Type: "payment.succeeded",
			Data: payment.WebhookEventData{
				OrderID:   "no-such-order",
				Amount:    990,
				PaymentID: "pay_unknown",
			},
		},
	}
	xunhuCli := &mockPaymentProvider{}
	emailCli := &mockEmailClient{}
	svc := newTestPaymentService(t, db, dodoCli, xunhuCli, emailCli)

	h := handleWebhook(svc)
	r := httptest.NewRequest("POST", "/api/webhook", strings.NewReader(`{"type":"payment.succeeded"}`))
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500 for non-existent order", w.Code)
	}
}

func TestHandleWebhook_Xunhu(t *testing.T) {
	db := openPaymentTestDB(t)
	defer db.Close()

	dodoCli := &mockPaymentProvider{
		webhookErr: errors.New("not dodo format"),
	}
	xunhuCli := &mockPaymentProvider{
		webhookEvent: &payment.WebhookEvent{
			Type: "payment.succeeded",
			Data: payment.WebhookEventData{
				OrderID:   "order-1",
				Amount:    990,
				PaymentID: "xunhu_pay_1",
			},
		},
	}
	emailCli := &mockEmailClient{}
	svc := newTestPaymentService(t, db, dodoCli, xunhuCli, emailCli)

	createTestOrder(t, db, "order-1", product.ProductNaming, payment.OrderPending, "a@b.co", `{"x":1}`, "")

	h := handleWebhook(svc)
	r := httptest.NewRequest("POST", "/api/webhook", strings.NewReader(`trade_status=TRADE_SUCCESS&out_trade_no=order-1`))
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", w.Code, w.Body.String())
	}
}

// -- handleReport --

func TestHandleReport_Valid(t *testing.T) {
	db := openPaymentTestDB(t)
	defer db.Close()

	dodoCli := &mockPaymentProvider{}
	xunhuCli := &mockPaymentProvider{}
	emailCli := &mockEmailClient{}
	svc := newTestPaymentService(t, db, dodoCli, xunhuCli, emailCli)

	createTestOrder(t, db, "order-1", product.ProductNaming, payment.OrderPaid, "a@b.co", `{"x":1}`, `{"report":"content"}`)

	h := handleReport(svc, &Analytics{})
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

	dodoCli := &mockPaymentProvider{}
	xunhuCli := &mockPaymentProvider{}
	emailCli := &mockEmailClient{}
	svc := newTestPaymentService(t, db, dodoCli, xunhuCli, emailCli)

	h := handleReport(svc, &Analytics{})
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

	dodoCli := &mockPaymentProvider{}
	xunhuCli := &mockPaymentProvider{}
	emailCli := &mockEmailClient{}
	svc := newTestPaymentService(t, db, dodoCli, xunhuCli, emailCli)

	h := handleReport(svc, &Analytics{})
	r := httptest.NewRequest("GET", "/api/reports/", nil)
	r.SetPathValue("id", "")
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestHandleReport_PendingStatus(t *testing.T) {
	db := openPaymentTestDB(t)
	defer db.Close()

	dodoCli := &mockPaymentProvider{}
	xunhuCli := &mockPaymentProvider{}
	emailCli := &mockEmailClient{}
	svc := newTestPaymentService(t, db, dodoCli, xunhuCli, emailCli)

	createTestOrder(t, db, "order-1", product.ProductNaming, payment.OrderPending, "a@b.co", `{"x":1}`, "")

	h := handleReport(svc, &Analytics{})
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

func TestHandleWebhook_ConcurrentSameOrder(t *testing.T) {
	db := openPaymentTestDB(t)
	defer db.Close()

	dodoCli := &mockPaymentProvider{
		webhookEvent: &payment.WebhookEvent{
			Type: "payment.succeeded",
			Data: payment.WebhookEventData{
				OrderID:   "concurrent-1",
				Amount:    990,
				PaymentID: "pay_race_1",
			},
		},
	}
	xunhuCli := &mockPaymentProvider{}
	emailCli := &mockEmailClient{}
	svc := newTestPaymentService(t, db, dodoCli, xunhuCli, emailCli)

	createTestOrder(t, db, "concurrent-1", product.ProductNaming, payment.OrderPending, "a@b.co", `{"x":1}`, "")

	h := handleWebhook(svc)
	body := `{"type":"payment.succeeded"}`

	// Fire two concurrent webhooks.
	done := make(chan struct{}, 2)
	var codes [2]int
	for i := 0; i < 2; i++ {
		go func(idx int) {
			w := httptest.NewRecorder()
			h(w, httptest.NewRequest("POST", "/api/webhook", strings.NewReader(body)))
			codes[idx] = w.Code
			done <- struct{}{}
		}(i)
	}
	<-done
	<-done

	if codes[0] != http.StatusOK && codes[1] != http.StatusOK {
		t.Fatalf("both webhooks failed: codes = %v", codes)
	}

	// Verify order is paid exactly once.
	order, err := svc.Store.GetOrder(context.Background(), "concurrent-1")
	if err != nil {
		t.Fatalf("GetOrder: %v", err)
	}
	if order.Status != "paid" {
		t.Errorf("status = %q, want paid", order.Status)
	}
	if order.PaymentID == "" {
		t.Error("payment_id should not be empty")
	}
}

func TestHandleWebhook_ConcurrentDifferentPaymentIDs(t *testing.T) {
	// Two different payment_ids for the same order — only one should win.
	db := openPaymentTestDB(t)
	defer db.Close()

	dodoCli1 := &mockPaymentProvider{
		webhookEvent: &payment.WebhookEvent{
			Type: "payment.succeeded",
			Data: payment.WebhookEventData{
				OrderID:   "concurrent-2",
				Amount:    990,
				PaymentID: "pay_first",
			},
		},
	}
	dodoCli2 := &mockPaymentProvider{
		webhookEvent: &payment.WebhookEvent{
			Type: "payment.succeeded",
			Data: payment.WebhookEventData{
				OrderID:   "concurrent-2",
				Amount:    990,
				PaymentID: "pay_second",
			},
		},
	}
	xunhuCli := &mockPaymentProvider{}
	emailCli := &mockEmailClient{}
	svc1 := newTestPaymentService(t, db, dodoCli1, xunhuCli, emailCli)
	svc2 := newTestPaymentService(t, db, dodoCli2, xunhuCli, emailCli)

	createTestOrder(t, db, "concurrent-2", product.ProductNaming, payment.OrderPending, "b@c.co", `{"x":1}`, "")

	h1 := handleWebhook(svc1)
	h2 := handleWebhook(svc2)
	body := `{"type":"payment.succeeded"}`

	done := make(chan struct{}, 2)
	go func() {
		w := httptest.NewRecorder()
		h1(w, httptest.NewRequest("POST", "/api/webhook", strings.NewReader(body)))
		done <- struct{}{}
	}()
	go func() {
		w := httptest.NewRecorder()
		h2(w, httptest.NewRequest("POST", "/api/webhook", strings.NewReader(body)))
		done <- struct{}{}
	}()
	<-done
	<-done

	order, err := svc1.Store.GetOrder(context.Background(), "concurrent-2")
	if err != nil {
		t.Fatalf("GetOrder: %v", err)
	}
	if order.Status != "paid" {
		t.Errorf("status = %q, want paid", order.Status)
	}
	if order.PaymentID == "" {
		t.Error("payment_id should not be empty")
	}
}

func TestHandleWebhook_EmailFailureDoesNotBlockPayment(t *testing.T) {
	db := openPaymentTestDB(t)
	defer db.Close()

	dodoCli := &mockPaymentProvider{
		webhookEvent: &payment.WebhookEvent{
			Type: "payment.succeeded",
			Data: payment.WebhookEventData{
				OrderID:   "email-fail-1",
				Amount:    990,
				PaymentID: "pay_email_fail",
			},
		},
	}
	xunhuCli := &mockPaymentProvider{}
	emailCli := &mockEmailClient{err: errors.New("SMTP connection refused")}
	svc := newTestPaymentService(t, db, dodoCli, xunhuCli, emailCli)

	createTestOrder(t, db, "email-fail-1", product.ProductNaming, payment.OrderPending, "user@test.com", `{"x":1}`, "")

	h := handleWebhook(svc)
	body := `{"type":"payment.succeeded"}`
	w := httptest.NewRecorder()
	h(w, httptest.NewRequest("POST", "/api/webhook", strings.NewReader(body)))

	if w.Code != http.StatusOK {
		t.Fatalf("webhook status = %d, want 200", w.Code)
	}

	// Payment should succeed despite email failure.
	order, err := svc.Store.GetOrder(context.Background(), "email-fail-1")
	if err != nil {
		t.Fatalf("GetOrder: %v", err)
	}
	if order.Status != "paid" {
		t.Errorf("status = %q, want paid (email failure should not block payment)", order.Status)
	}

	// Report should be accessible.
	reportH := handleReport(svc, &Analytics{})
	reportW := httptest.NewRecorder()
	reportR := httptest.NewRequest("GET", "/api/reports/email-fail-1", nil)
	reportR.SetPathValue("id", "email-fail-1")
	reportH(reportW, reportR)

	if reportW.Code != http.StatusOK {
		t.Fatalf("report status = %d, want 200", reportW.Code)
	}
}
