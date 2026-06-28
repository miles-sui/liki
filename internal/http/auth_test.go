package http

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"liki/internal/payment"
	"liki/internal/product"

	_ "modernc.org/sqlite"
)

func newAuthTestStore(t *testing.T) *payment.Store {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	db.SetMaxOpenConns(1)
	store, err := payment.NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	return store
}

func seedActiveOrder(t *testing.T, store *payment.Store, orderID, email string, prod product.Product, expiresAt string) {
	t.Helper()
	if err := store.CreateOrder(context.Background(), orderID, prod, 2990, "CNY", "", "", "", "dodo"); err != nil {
		t.Fatalf("create order: %v", err)
	}
	if err := store.UpdateEmail(context.Background(), orderID, email); err != nil {
		t.Fatalf("update email: %v", err)
	}
	if _, _, _, err := store.MarkPaidIdempotent(context.Background(), orderID, "pay-"+orderID); err != nil {
		t.Fatalf("mark paid: %v", err)
	}
	if expiresAt == "" {
		expiresAt = "2027-01-01 00:00:00"
	}
	// MarkPaidIdempotent now sets chat_expires_at atomically.
	// Override with the caller's desired value.
	if err := store.OverrideChatExpiresAt(context.Background(), orderID, expiresAt); err != nil {
		t.Fatalf("set chat expiry: %v", err)
	}
}

func TestHandleLogin_NoOrders(t *testing.T) {
	store := newAuthTestStore(t)
	handler := handleLogin(store)

	body := `{"email":"nobody@example.com"}`
	r := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(body))
	w := httptest.NewRecorder()
	handler(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestHandleLogin_OneOrder(t *testing.T) {
	store := newAuthTestStore(t)
	seedActiveOrder(t, store, "order-1", "user@example.com", product.ProductNaming, "2027-01-01 00:00:00")
	handler := handleLogin(store)

	body := `{"email":"user@example.com"}`
	r := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(body))
	w := httptest.NewRecorder()
	handler(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var resp struct {
		Data loginResponse `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if resp.Data.OrderID != "order-1" {
		t.Errorf("order_id = %q, want order-1", resp.Data.OrderID)
	}
	if resp.Data.Redirect != "/chat" {
		t.Errorf("redirect = %q, want /chat", resp.Data.Redirect)
	}

	// Check JWT cookie is set.
	cookies := w.Result().Cookies()
	var tokenCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "liki_token" {
			tokenCookie = c
			break
		}
	}
	if tokenCookie == nil {
		t.Error("liki_token cookie not set")
	}
}

func TestHandleLogin_MultipleOrders(t *testing.T) {
	store := newAuthTestStore(t)
	seedActiveOrder(t, store, "order-1", "user@example.com", product.ProductNaming, "2027-01-01 00:00:00")
	seedActiveOrder(t, store, "order-2", "user@example.com", product.ProductNaming, "2027-02-01 00:00:00")
	handler := handleLogin(store)

	body := `{"email":"user@example.com"}`
	r := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(body))
	w := httptest.NewRecorder()
	handler(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var resp struct {
		Data loginResponse `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if len(resp.Data.Orders) != 2 {
		t.Fatalf("orders count = %d, want 2", len(resp.Data.Orders))
	}
	if resp.Data.OrderID != "" {
		t.Error("order_id should be empty when multiple orders")
	}
}

func TestHandleLogin_EmptyEmail(t *testing.T) {
	store := newAuthTestStore(t)
	handler := handleLogin(store)

	r := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(`{"email":""}`))
	w := httptest.NewRecorder()
	handler(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestHandleLogin_InvalidJSON(t *testing.T) {
	store := newAuthTestStore(t)
	handler := handleLogin(store)

	r := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(`{bad`))
	w := httptest.NewRecorder()
	handler(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestHandleOrderSelect_Success(t *testing.T) {
	store := newAuthTestStore(t)
	seedActiveOrder(t, store, "order-1", "user@example.com", product.ProductNaming, "2027-01-01 00:00:00")
	handler := handleOrderSelect(store)

	body := `{"order_id":"order-1","email":"user@example.com"}`
	r := httptest.NewRequest("POST", "/api/orders/select", strings.NewReader(body))
	w := httptest.NewRecorder()
	handler(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	// Check JWT cookie is set.
	cookies := w.Result().Cookies()
	var tokenCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "liki_token" {
			tokenCookie = c
			break
		}
	}
	if tokenCookie == nil {
		t.Error("liki_token cookie not set")
	}
}

func TestHandleOrderSelect_WrongEmail(t *testing.T) {
	store := newAuthTestStore(t)
	seedActiveOrder(t, store, "order-1", "user@example.com", product.ProductNaming, "2027-01-01 00:00:00")
	handler := handleOrderSelect(store)

	body := `{"order_id":"order-1","email":"attacker@example.com"}`
	r := httptest.NewRequest("POST", "/api/orders/select", strings.NewReader(body))
	w := httptest.NewRecorder()
	handler(w, r)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", w.Code)
	}
}

func TestHandleOrderSelect_OrderNotFound(t *testing.T) {
	store := newAuthTestStore(t)
	handler := handleOrderSelect(store)

	body := `{"order_id":"nonexistent","email":"user@example.com"}`
	r := httptest.NewRequest("POST", "/api/orders/select", strings.NewReader(body))
	w := httptest.NewRecorder()
	handler(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestHandleOrderSelect_MissingFields(t *testing.T) {
	store := newAuthTestStore(t)
	handler := handleOrderSelect(store)

	tests := []struct {
		name string
		body string
	}{
		{"no email", `{"order_id":"order-1"}`},
		{"no order_id", `{"email":"user@example.com"}`},
		{"both empty", `{"order_id":"","email":""}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("POST", "/api/orders/select", strings.NewReader(tt.body))
			w := httptest.NewRecorder()
			handler(w, r)
			if w.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400", w.Code)
			}
		})
	}
}

func TestJWTAuth_NoCookie(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	_, _, ok := jwtAuth(r)
	if ok {
		t.Error("expected no auth without cookie")
	}
}

func TestJWTAuth_InvalidToken(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{Name: "liki_token", Value: "not-a-valid-jwt"})
	_, _, ok := jwtAuth(r)
	if ok {
		t.Error("expected no auth with invalid token")
	}
}

func TestJWTAuth_ExpiredToken(t *testing.T) {
	// Tokens with an expired exp claim should be rejected.
	claims := jwt.MapClaims{
		"email":    "user@example.com",
		"order_id": "order-1",
		"exp":      time.Now().Add(-1 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(jwtSecret())
	if err != nil {
		t.Fatalf("sign expired token: %v", err)
	}

	r := httptest.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{Name: "liki_token", Value: signed})
	_, _, ok := jwtAuth(r)
	if ok {
		t.Error("expected no auth with expired token")
	}
}

func TestJWTAuth_RoundTrip(t *testing.T) {
	// Set cookie, then verify jwtAuth can read it back.
	w := httptest.NewRecorder()
	setJWTCookie(w, "user@example.com", "order-1")

	cookies := w.Result().Cookies()
	var tokenValue string
	for _, c := range cookies {
		if c.Name == "liki_token" {
			tokenValue = c.Value
			break
		}
	}
	if tokenValue == "" {
		t.Fatal("cookie not set")
	}

	r := httptest.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{Name: "liki_token", Value: tokenValue})
	email, orderID, ok := jwtAuth(r)
	if !ok {
		t.Fatal("jwtAuth failed")
	}
	if email != "user@example.com" {
		t.Errorf("email = %q, want user@example.com", email)
	}
	if orderID != "order-1" {
		t.Errorf("orderID = %q, want order-1", orderID)
	}
}

func TestHandleLogin_HasBirthInfo(t *testing.T) {
	store := newAuthTestStore(t)
	seedActiveOrder(t, store, "order-1", "user@example.com", product.ProductNaming, "2027-01-01 00:00:00")
	// Set birth_info on the order.
	if _, err := store.UpdateBirthInfoIfEmpty(context.Background(), "order-1", `{"raw":{"year":2026}}`); err != nil {
		t.Fatalf("UpdateBirthInfoIfEmpty: %v", err)
	}
	handler := handleLogin(store)

	body := `{"email":"user@example.com"}`
	r := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(body))
	w := httptest.NewRecorder()
	handler(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var resp struct {
		Data loginResponse `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !resp.Data.HasBirthInfo {
		t.Error("HasBirthInfo should be true when birth_info is set")
	}
}

func TestHandleLogin_ExpiredOrderNotReturned(t *testing.T) {
	store := newAuthTestStore(t)
	// Seed an expired paid order — should NOT be returned.
	if err := store.CreateOrder(context.Background(), "order-expired", product.ProductNaming, 2990, "CNY", "", "", "", "dodo"); err != nil {
		t.Fatalf("create order: %v", err)
	}
	if err := store.UpdateEmail(context.Background(), "order-expired", "user@example.com"); err != nil {
		t.Fatalf("update email: %v", err)
	}
	if _, _, _, err := store.MarkPaidIdempotent(context.Background(), "order-expired", "pay-expired"); err != nil {
		t.Fatalf("mark paid: %v", err)
	}
	if err := store.OverrideChatExpiresAt(context.Background(), "order-expired", "2020-01-01 00:00:00"); err != nil {
		t.Fatalf("set chat expiry: %v", err)
	}

	handler := handleLogin(store)
	body := `{"email":"user@example.com"}`
	r := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(body))
	w := httptest.NewRecorder()
	handler(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401 (expired orders should not be returned)", w.Code)
	}
}

func TestHandleOrderSelect_HasBirthInfo(t *testing.T) {
	store := newAuthTestStore(t)
	seedActiveOrder(t, store, "order-1", "user@example.com", product.ProductNaming, "2027-01-01 00:00:00")
	if _, err := store.UpdateBirthInfoIfEmpty(context.Background(), "order-1", `{"raw":{"year":2026}}`); err != nil {
		t.Fatalf("UpdateBirthInfoIfEmpty: %v", err)
	}
	handler := handleOrderSelect(store)

	body := `{"order_id":"order-1","email":"user@example.com"}`
	r := httptest.NewRequest("POST", "/api/orders/select", strings.NewReader(body))
	w := httptest.NewRecorder()
	handler(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var resp struct {
		Data loginResponse `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !resp.Data.HasBirthInfo {
		t.Error("HasBirthInfo should be true when birth_info is set")
	}
}
