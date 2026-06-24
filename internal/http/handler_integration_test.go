//go:build integration

package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"liki/internal/agent"
	"liki/internal/payment"
)

// ── helpers ──

func openHandlerTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "handler-test.db")
	db, err := payment.OpenDB(path)
	if err != nil {
		t.Fatalf("OpenDB: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func newHandlerTestService(t *testing.T, db *sql.DB) *payment.Service {
	t.Helper()
	store, err := payment.NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	return payment.NewService(
		&stubProviderForHandler{},
		&stubProviderForHandler{},
		&stubEmailForHandler{},
		store,
		"https://example.com",
		"admin@example.com",
		nil,
		context.Background(),
	)
}

type stubProviderForHandler struct{}

func (d *stubProviderForHandler) CreateCheckout(_ context.Context, _ agent.Product, _ int, _, _, _ string) (*payment.CheckoutResult, error) {
	return &payment.CheckoutResult{SessionID: "sess-test", CheckoutURL: "https://pay.example.com/checkout"}, nil
}

func (d *stubProviderForHandler) VerifyWebhook(_ []byte, _ http.Header) (*payment.WebhookEvent, error) {
	return &payment.WebhookEvent{
		Type: "payment.succeeded",
		Data: payment.WebhookEventData{OrderID: "hook-1", Amount: 990, PaymentID: "pay-1"},
	}, nil
}

type stubEmailForHandler struct{}

func (e *stubEmailForHandler) SendReport(_ context.Context, _, _, _ string) error {
	return nil
}

func createHandlerTestOrder(t *testing.T, db *sql.DB, orderID string, status payment.OrderStatus, chartJSON, llmJSON string) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO orders (order_id, product, amount, currency, provider, email, chart_json, llm_json, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		orderID, "chart", 990, "CNY", "", "buyer@test.com", chartJSON, llmJSON, status,
	)
	if err != nil {
		t.Fatalf("create test order: %v", err)
	}
}

// ── Checkout → payment return → order status → report: full lifecycle ──

func TestCheckoutToReportLifecycle(t *testing.T) {
	db := openHandlerTestDB(t)
	svc := newHandlerTestService(t, db)

	// 1. Create an order directly in DB (simulating agent's purchase tool).
	createHandlerTestOrder(t, db, "lifecycle-1", payment.OrderPending, `{"chart":{}}`, "")

	// 2. Checkout — creates checkout session.
	checkoutHandler := handleCheckout(svc)
	checkoutReq := httptest.NewRequest("POST", "/api/payments/checkout", strings.NewReader(`{"order_id":"lifecycle-1","email":"buyer@test.com","provider":"dodo"}`))
	checkoutW := httptest.NewRecorder()
	checkoutHandler(checkoutW, checkoutReq)

	if checkoutW.Code != http.StatusOK {
		t.Fatalf("checkout status = %d, want 200", checkoutW.Code)
	}
	var checkoutEnv struct {
		Data struct {
			SessionID   string `json:"session_id"`
			CheckoutURL string `json:"checkout_url"`
		} `json:"data"`
	}
	if err := json.NewDecoder(checkoutW.Body).Decode(&checkoutEnv); err != nil {
		t.Fatalf("decode checkout: %v", err)
	}
	if checkoutEnv.Data.SessionID == "" {
		t.Error("checkout session_id must not be empty")
	}

	// 3. Payment return (succeeded) — redirects to report page.
	returnHandler := handlePaymentReturn()
	returnReq := httptest.NewRequest("GET", "/api/payments/return/lifecycle-1?status=succeeded", nil)
	returnReq.SetPathValue("id", "lifecycle-1")
	returnW := httptest.NewRecorder()
	returnHandler(returnW, returnReq)

	if returnW.Code != http.StatusFound {
		t.Errorf("return status = %d, want 302", returnW.Code)
	}
	if loc := returnW.Header().Get("Location"); loc != "/report/lifecycle-1" {
		t.Errorf("return redirect = %q, want /report/lifecycle-1", loc)
	}

	// 4. Order status — returns pending (not yet paid via webhook).
	statusHandler := handleOrderStatus(svc)
	statusReq := httptest.NewRequest("GET", "/api/orders/lifecycle-1/status", nil)
	statusReq.SetPathValue("id", "lifecycle-1")
	statusW := httptest.NewRecorder()
	statusHandler(statusW, statusReq)

	if statusW.Code != http.StatusOK {
		t.Fatalf("order status = %d, want 200", statusW.Code)
	}
	var statusEnv struct {
		Data struct {
			Status  string `json:"status"`
			Product string `json:"product"`
		} `json:"data"`
	}
	json.NewDecoder(statusW.Body).Decode(&statusEnv)
	if statusEnv.Data.Status != "pending" {
		t.Errorf("status = %q, want pending", statusEnv.Data.Status)
	}

	// 5. Report with pending status — llm_json is hidden.
	reportHandler := handleReport(svc)
	reportReq := httptest.NewRequest("GET", "/api/reports/lifecycle-1", nil)
	reportReq.SetPathValue("id", "lifecycle-1")
	reportW := httptest.NewRecorder()
	reportHandler(reportW, reportReq)

	if reportW.Code != http.StatusOK {
		t.Fatalf("report status = %d, want 200", reportW.Code)
	}
	var reportEnv struct {
		Data struct {
			Status  string `json:"status"`
			LlmJSON string `json:"llm_json"`
		} `json:"data"`
	}
	json.NewDecoder(reportW.Body).Decode(&reportEnv)
	if reportEnv.Data.Status != "pending" {
		t.Errorf("report status = %q, want pending", reportEnv.Data.Status)
	}
	if reportEnv.Data.LlmJSON != "" {
		t.Errorf("pending order must not expose llm_json, got %q", reportEnv.Data.LlmJSON)
	}
}

// ── Webhook marks order paid, report shows llm_json ──

func TestWebhookThenReport(t *testing.T) {
	db := openHandlerTestDB(t)
	svc := newHandlerTestService(t, db)

	createHandlerTestOrder(t, db, "hook-rpt-1", payment.OrderPending, `{"chart":{}}`, "")
	// Pre-fill llm_json as if background generation completed.
	svc.Store.MarkPaidIdempotent(context.Background(), "hook-rpt-1", "pay-hook-1")
	svc.Store.UpdateLlmJSON(context.Background(), "hook-rpt-1", "<p>full report</p>")

	// Report after payment: llm_json is exposed.
	reportHandler := handleReport(svc)
	reportReq := httptest.NewRequest("GET", "/api/reports/hook-rpt-1", nil)
	reportReq.SetPathValue("id", "hook-rpt-1")
	reportW := httptest.NewRecorder()
	reportHandler(reportW, reportReq)

	if reportW.Code != http.StatusOK {
		t.Fatalf("report status = %d, want 200", reportW.Code)
	}
	var reportEnv struct {
		Data struct {
			Status  string `json:"status"`
			LlmJSON string `json:"llm_json"`
		} `json:"data"`
	}
	json.NewDecoder(reportW.Body).Decode(&reportEnv)
	if reportEnv.Data.Status != "paid" {
		t.Errorf("status = %q, want paid", reportEnv.Data.Status)
	}
	if reportEnv.Data.LlmJSON != "<p>full report</p>" {
		t.Errorf("llm_json = %q, want <p>full report</p>", reportEnv.Data.LlmJSON)
	}
}

// ── Retry order re-triggers generation for paid orders without llm_json ──

func TestRetryOrder_Integration(t *testing.T) {
	db := openHandlerTestDB(t)
	svc := newHandlerTestService(t, db)

	// Manually mark paid without llm_json (missed webhook scenario).
	createHandlerTestOrder(t, db, "retry-int-1", payment.OrderPaid, `{"chart":{}}`, "")
	db.Exec(`UPDATE orders SET status='paid', payment_id='pay-retry-int' WHERE order_id='retry-int-1'`)

	retryHandler := handleRetryOrder(svc)
	retryReq := httptest.NewRequest("POST", "/api/orders/retry-int-1/retry", nil)
	retryReq.SetPathValue("id", "retry-int-1")
	retryW := httptest.NewRecorder()
	retryHandler(retryW, retryReq)

	if retryW.Code != http.StatusOK {
		t.Fatalf("retry status = %d, want 200", retryW.Code)
	}
}

// ── Missing order ID returns 400/404 ──

func TestReport_MissingID_Integration(t *testing.T) {
	db := openHandlerTestDB(t)
	svc := newHandlerTestService(t, db)

	reportHandler := handleReport(svc)
	reportReq := httptest.NewRequest("GET", "/api/reports/", nil)
	reportReq.SetPathValue("id", "")
	reportW := httptest.NewRecorder()
	reportHandler(reportW, reportReq)

	if reportW.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", reportW.Code)
	}
}

func TestOrderStatus_NotFound_Integration(t *testing.T) {
	db := openHandlerTestDB(t)
	svc := newHandlerTestService(t, db)

	statusHandler := handleOrderStatus(svc)
	statusReq := httptest.NewRequest("GET", "/api/orders/no-such-order/status", nil)
	statusReq.SetPathValue("id", "no-such-order")
	statusW := httptest.NewRecorder()
	statusHandler(statusW, statusReq)

	if statusW.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", statusW.Code)
	}
}

// ── Payment return redirects ──

func TestPaymentReturn_Failed_Integration(t *testing.T) {
	handler := handlePaymentReturn()
	req := httptest.NewRequest("GET", "/api/payments/return/order-x?status=failed", nil)
	req.SetPathValue("id", "order-x")
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("status = %d, want 302", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "/" {
		t.Errorf("Location = %q, want /", loc)
	}
}

// ── Download redirect ──

func TestDownloadRedirect_Integration(t *testing.T) {
	handler := redirectReport()
	req := httptest.NewRequest("GET", "/api/orders/dl-1/report", nil)
	req.SetPathValue("id", "dl-1")
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("status = %d, want 302", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "/report/dl-1" {
		t.Errorf("Location = %q, want /report/dl-1", loc)
	}
}

// ── Stale order cleanup ──

func TestCleanStale_Integration(t *testing.T) {
	db := openHandlerTestDB(t)
	svc := newHandlerTestService(t, db)

	ctx := context.Background()
	svc.Store.CreateOrder(ctx, "stale-1", agent.ProductChart, 990, "CNY", `{}`, "", "zh-Hans", "")
	// Backdate to 2020.
	db.ExecContext(ctx, `UPDATE orders SET created_at = '2020-01-01 00:00:00' WHERE order_id = 'stale-1'`)

	if err := svc.Store.CleanStale(ctx, 24*time.Hour); err != nil {
		t.Fatalf("CleanStale: %v", err)
	}

	_, err := svc.Store.GetOrder(ctx, "stale-1")
	if err != payment.ErrOrderNotFound {
		t.Errorf("stale order should be cleaned, got %v", err)
	}
}
