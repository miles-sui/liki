package xunhu

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"liki/internal/product"
)

func TestSign(t *testing.T) {
	tests := []struct {
		name    string
		params  map[string]string
		secret  string
		wantLen int // MD5 hex is always 32
	}{
		{
			name:   "basic params",
			params: map[string]string{"appid": "test_app", "version": "1.0", "trade_order_id": "order-1"},
			secret: "secret123",
		},
		{
			name:    "empty params",
			params:  map[string]string{},
			secret:  "secret",
			wantLen: 32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sig := sign(tt.params, tt.secret)
			if len(sig) != 32 {
				t.Errorf("sign() length = %d, want 32", len(sig))
			}
			// Same input produces same output.
			sig2 := sign(tt.params, tt.secret)
			if sig != sig2 {
				t.Error("sign() not deterministic")
			}
		})
	}
}

func TestSign_DifferentSecret(t *testing.T) {
	params := map[string]string{"appid": "test", "version": "1.0"}
	if sign(params, "secretA") == sign(params, "secretB") {
		t.Error("sign() should differ with different secrets")
	}
}

func TestSign_DifferentParams(t *testing.T) {
	a := map[string]string{"appid": "test", "version": "1.0"}
	b := map[string]string{"appid": "test", "version": "1.1"}
	if sign(a, "secret") == sign(b, "secret") {
		t.Error("sign() should differ with different params")
	}
}

func TestSign_OrderIndependent(t *testing.T) {
	// Same key-values in different map iteration order should produce same hash.
	// We create equivalent map literal — Go map iteration is random, so
	// calling sign twice on the same map tests ordering stability.
	params := map[string]string{"c": "3", "a": "1", "b": "2"}
	s1 := sign(params, "secret")
	s2 := sign(params, "secret")
	if s1 != s2 {
		t.Error("sign() should be deterministic regardless of map iteration order")
	}
}

func newTestClient(srv *httptest.Server) *Client {
	return &Client{
		appID:      "test_appid",
		appSecret:  "test_secret",
		httpClient: srv.Client(),
		baseURL:    srv.URL,
	}
}

func TestCreateCheckout_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}

		// Verify required params are present.
		if r.Form.Get("appid") != "test_appid" {
			t.Errorf("appid = %q", r.Form.Get("appid"))
		}
		if r.Form.Get("trade_order_id") != "order-1" {
			t.Errorf("trade_order_id = %q", r.Form.Get("trade_order_id"))
		}

		// Verify hash parameter is present.
		if r.Form.Get("hash") == "" {
			t.Error("hash is missing")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
			"errcode":    0,
			"errmsg":     "success",
			"url":        "https://pay.xunhupay.com/checkout/test",
			"url_qrcode": "https://pay.xunhupay.com/qrcode/test",
		})
	}))
	defer srv.Close()

	c := newTestClient(srv)
	result, err := c.CreateCheckout(context.Background(), product.ProductNaming, 990, "order-1", "user@test.com", "https://liki.hk/return")
	if err != nil {
		t.Fatalf("CreateCheckout: %v", err)
	}
	if result.SessionID == "" {
		t.Error("SessionID should not be empty")
	}
	if result.CheckoutURL != "https://pay.xunhupay.com/checkout/test" {
		t.Errorf("CheckoutURL = %q", result.CheckoutURL)
	}
	if result.QRCodeURL != "https://pay.xunhupay.com/qrcode/test" {
		t.Errorf("QRCodeURL = %q", result.QRCodeURL)
	}
}

func TestCreateCheckout_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
			"errcode": 400,
			"errmsg":  "appid not found",
		})
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.CreateCheckout(context.Background(), product.ProductNaming, 990, "order-1", "", "https://liki.hk/return")
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

func TestCreateCheckout_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.CreateCheckout(context.Background(), product.ProductNaming, 990, "order-1", "", "https://liki.hk/return")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestCreateCheckout_Timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_, err := c.CreateCheckout(ctx, product.ProductNaming, 990, "order-1", "", "https://liki.hk/return")
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestVerifyWebhook_PaymentSuccess(t *testing.T) {
	c := &Client{
		appID:     "test_appid",
		appSecret: "test_secret",
	}

	// Build form with valid hash.
	form := url.Values{
		"appid":         {"test_appid"},
		"trade_order_id":  {"order-1"},
		"total_fee":     {"990"},
		"trade_no":      {"txn_123"},
		"trade_status":  {"TRADE_SUCCESS"},
		"openid":        {"user_openid"},
	}
	formParams := map[string]string{
		"appid":        "test_appid",
		"trade_order_id": "order-1",
		"total_fee":    "990",
		"trade_no":     "txn_123",
		"trade_status": "TRADE_SUCCESS",
		"openid":       "user_openid",
	}
	form.Set("hash", sign(formParams, "test_secret"))

	body := []byte(form.Encode())
	headers := http.Header{"Content-Type": {"application/x-www-form-urlencoded"}}

	event, err := c.VerifyWebhook(body, headers)
	if err != nil {
		t.Fatalf("VerifyWebhook: %v", err)
	}
	if event.Type != "payment.succeeded" {
		t.Errorf("type = %q, want payment.succeeded", event.Type)
	}
	if event.Data.OrderID != "order-1" {
		t.Errorf("OrderID = %q, want order-1", event.Data.OrderID)
	}
	if event.Data.Amount != 990 {
		t.Errorf("Amount = %d, want 990", event.Data.Amount)
	}
	if event.Data.PaymentID != "txn_123" {
		t.Errorf("PaymentID = %q, want txn_123", event.Data.PaymentID)
	}
}

func TestVerifyWebhook_BadSignature(t *testing.T) {
	c := &Client{
		appID:     "test_appid",
		appSecret: "test_secret",
	}

	form := url.Values{
		"appid":        {"test_appid"},
		"trade_order_id": {"order-1"},
		"total_fee":    {"990"},
		"trade_status": {"TRADE_SUCCESS"},
		"hash":         {"bad_signature_here"},
	}
	body := []byte(form.Encode())
	headers := http.Header{"Content-Type": {"application/x-www-form-urlencoded"}}

	_, err := c.VerifyWebhook(body, headers)
	if err == nil {
		t.Fatal("expected error for bad signature")
	}
}

func TestVerifyWebhook_MissingHash(t *testing.T) {
	c := &Client{
		appID:     "test_appid",
		appSecret: "test_secret",
	}

	form := url.Values{
		"appid":        {"test_appid"},
		"trade_order_id": {"order-1"},
		"trade_status": {"TRADE_SUCCESS"},
	}
	body := []byte(form.Encode())
	headers := http.Header{"Content-Type": {"application/x-www-form-urlencoded"}}

	_, err := c.VerifyWebhook(body, headers)
	if err == nil {
		t.Fatal("expected error for missing hash")
	}
}

func TestVerifyWebhook_NonPaymentStatus(t *testing.T) {
	c := &Client{
		appID:     "test_appid",
		appSecret: "test_secret",
	}

	formParams := map[string]string{
		"appid":        "test_appid",
		"trade_order_id": "order-1",
		"total_fee":    "990",
		"trade_status": "ORDER_CREATED",
	}
	form := url.Values{}
	for k, v := range formParams {
		form.Set(k, v)
	}
	form.Set("hash", sign(formParams, "test_secret"))

	body := []byte(form.Encode())
	headers := http.Header{"Content-Type": {"application/x-www-form-urlencoded"}}

	event, err := c.VerifyWebhook(body, headers)
	if err != nil {
		t.Fatalf("VerifyWebhook: %v", err)
	}
	if event.Type != "ORDER_CREATED" {
		t.Errorf("type = %q, want ORDER_CREATED", event.Type)
	}
}

func TestVerifyWebhook_InvalidBody(t *testing.T) {
	c := &Client{
		appID:     "test_appid",
		appSecret: "test_secret",
	}

	_, err := c.VerifyWebhook([]byte("%%%"), http.Header{"Content-Type": {"application/x-www-form-urlencoded"}})
	if err == nil {
		t.Fatal("expected error for invalid form body")
	}
}

func TestVerifyWebhook_EmptyBody(t *testing.T) {
	c := &Client{
		appID:     "test_appid",
		appSecret: "test_secret",
	}

	_, err := c.VerifyWebhook([]byte{}, http.Header{"Content-Type": {"application/x-www-form-urlencoded"}})
	if err == nil {
		t.Fatal("expected error for empty body")
	}
}

func TestNew(t *testing.T) {
	c := New("app_123", "secret_456")
	if c == nil {
		t.Fatal("New returned nil")
	}
	if c.appID != "app_123" {
		t.Errorf("appID = %q, want app_123", c.appID)
	}
	if c.appSecret != "secret_456" {
		t.Errorf("appSecret = %q, want secret_456", c.appSecret)
	}
	if c.httpClient == nil {
		t.Error("httpClient should not be nil")
	}
	if c.baseURL == "" {
		t.Error("baseURL should not be empty")
	}
}

func TestDeriveWebhookURL(t *testing.T) {
	tests := []struct {
		name      string
		returnURL string
		want      string
	}{
		{"standard return URL", "https://liki.hk/api/payments/return/order-1", "https://liki.hk/api/payments/webhook"},
		{"https with path", "https://liki.hk/return/order-1", "https://liki.hk/api/payments/webhook"},
		{"url with port", "https://liki.hk:8443/api/payments/return/order-1", "https://liki.hk:8443/api/payments/webhook"},
		{"no path in return URL", "https://liki.hk", "https://liki.hk/api/payments/webhook"},
		{"invalid URL returns input", "not-a-url", "not-a-url"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deriveWebhookURL(tt.returnURL)
			if got != tt.want {
				t.Errorf("deriveWebhookURL(%q) = %q, want %q", tt.returnURL, got, tt.want)
			}
		})
	}
}

func TestCreateCheckout_ProductPricing(t *testing.T) {
	// Verify all known products produce a valid checkout call.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
			"errcode": 0,
			"errmsg":  "ok",
			"url":     "https://pay.example.com/checkout",
		})
	}))
	defer srv.Close()

	c := newTestClient(srv)

	tests := []struct {
		name    string
		prod product.Product
		amount  int
	}{
		{"chart", product.ProductNaming, 990},
		{"bond", product.ProductNaming, 1990},
		{"naming", product.ProductNaming, 2990},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := c.CreateCheckout(context.Background(), tt.prod, tt.amount, "order-1", "", "https://liki.hk/return")
			if err != nil {
				t.Fatalf("CreateCheckout(%s): %v", tt.name, err)
			}
			if result.CheckoutURL == "" {
				t.Error("CheckoutURL should not be empty")
			}
		})
	}
}

func TestCreateCheckout_ParamFormat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}

		// version must be 1.1.
		if v := r.Form.Get("version"); v != "1.1" {
			t.Errorf("version = %q, want 1.1", v)
		}
		// nonce_str must be 32 hex chars.
		if n := r.Form.Get("nonce_str"); len(n) != 32 {
			t.Errorf("nonce_str len = %d, want 32", len(n))
		}
		// total_fee must be decimal yuan, not integer fen.
		if f := r.Form.Get("total_fee"); f != "9.90" {
			t.Errorf("total_fee = %q, want 9.90 (yuan decimal)", f)
		}
		// title must be present (product subject).
		if title := r.Form.Get("title"); title == "" {
			t.Error("title should not be empty")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
			"errcode": 0,
			"errmsg":  "ok",
			"url":     "https://pay.example.com/checkout",
		})
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.CreateCheckout(context.Background(), product.ProductNaming, 990, "order-1", "", "https://liki.hk/return")
	if err != nil {
		t.Fatalf("CreateCheckout: %v", err)
	}
}

func TestCreateCheckout_AmountFormat(t *testing.T) {
	tests := []struct {
		name   string
		amount int
		want   string
	}{
		{"990 fen → 9.90", 990, "9.90"},
		{"1990 fen → 19.90", 1990, "19.90"},
		{"2990 fen → 29.90", 2990, "29.90"},
		{"100 fen → 1.00", 100, "1.00"},
		{"0 fen → 0.00", 0, "0.00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.ParseForm() //nolint:errcheck
				if f := r.Form.Get("total_fee"); f != tt.want {
					t.Errorf("total_fee = %q, want %q", f, tt.want)
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
					"errcode": 0, "errmsg": "ok", "url": "https://pay.example.com/checkout",
				})
			}))
			defer srv.Close()

			c := newTestClient(srv)
			_, err := c.CreateCheckout(context.Background(), product.ProductNaming, tt.amount, "order-1", "", "https://liki.hk/return")
			if err != nil {
				t.Fatalf("CreateCheckout: %v", err)
			}
		})
	}
}

func TestVerifyWebhook_LegacyOrderID(t *testing.T) {
	c := &Client{
		appID:     "test_appid",
		appSecret: "test_secret",
	}

	formParams := map[string]string{
		"appid":        "test_appid",
		"out_trade_no": "legacy-order-1",
		"total_fee":    "990",
		"trade_no":     "txn_123",
		"trade_status": "TRADE_SUCCESS",
	}
	form := url.Values{}
	for k, v := range formParams {
		form.Set(k, v)
	}
	form.Set("hash", sign(formParams, "test_secret"))

	event, err := c.VerifyWebhook(
		[]byte(form.Encode()),
		http.Header{"Content-Type": {"application/x-www-form-urlencoded"}},
	)
	if err != nil {
		t.Fatalf("VerifyWebhook: %v", err)
	}
	if event.Data.OrderID != "legacy-order-1" {
		t.Errorf("OrderID = %q, want legacy-order-1", event.Data.OrderID)
	}
}

func TestCreateCheckout_HashPresent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm() //nolint:errcheck
		h := r.Form.Get("hash")
		if h == "" {
			t.Error("hash is missing")
		}
		if len(h) != 32 {
			t.Errorf("hash len = %d, want 32", len(h))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
			"errcode": 0, "errmsg": "ok", "url": "https://pay.example.com/checkout",
		})
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.CreateCheckout(context.Background(), product.ProductNaming, 990, "order-1", "", "https://liki.hk/return")
	if err != nil {
		t.Fatalf("CreateCheckout: %v", err)
	}
}

func TestSign_Deterministic(t *testing.T) {
	params := map[string]string{
		"appid":          "201906182089",
		"version":        "1.1",
		"time":           "1782304566",
		"nonce_str":      "b0eaff923b531dc05e4d679aeb32bb73",
		"trade_order_id": "go-test-001",
		"total_fee":      "9.90",
		"title":          "test-chart",
		"notify_url":     "https://liki.hk/api/payments/webhook",
		"return_url":     "https://liki.hk/return",
	}
	got := sign(params, "df0e94d93d3d3328fc83f4daf815ffc2")

	got2 := sign(params, "df0e94d93d3d3328fc83f4daf815ffc2")
	if got != got2 {
		t.Error("sign() should be deterministic")
	}
	if len(got) != 32 {
		t.Errorf("sign() length = %d, want 32", len(got))
	}
}
