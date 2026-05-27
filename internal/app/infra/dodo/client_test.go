package dodo

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	standardwebhooks "github.com/standard-webhooks/standard-webhooks/libraries/go"

	"github.com/25types/25types/internal/app/application/commerce"
	dodo "github.com/dodopayments/dodopayments-go"
	"github.com/dodopayments/dodopayments-go/option"
)

func TestNew_TestMode(t *testing.T) {
	c := New("api-key", "wh-key", true)
	if c == nil {
		t.Fatal("expected client, got nil")
	}
	if c.checkoutSvc == nil {
		t.Error("checkoutSvc is nil")
	}
	if c.webhookSvc == nil {
		t.Error("webhookSvc is nil")
	}
}

func TestNew_LiveMode(t *testing.T) {
	c := New("api-key", "wh-key", false)
	if c == nil {
		t.Fatal("expected client, got nil")
	}
}

func TestCreateCheckout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/checkouts" {
			t.Errorf("expected /checkouts, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			t.Errorf("wrong auth header: %s", r.Header.Get("Authorization"))
		}

		var body checkoutRequestJSON
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		cart := body.ProductCart
		if len(cart) != 1 {
			t.Fatalf("expected 1 product, got %d", len(cart))
		}
		if cart[0].Amount != 990 {
			t.Errorf("expected amount 990, got %d", cart[0].Amount)
		}
		if cart[0].ProductID != "pdt_xxx" {
			t.Errorf("expected product_id pdt_xxx, got %s", cart[0].ProductID)
		}
		if body.Metadata["user_id"] != "42" {
			t.Errorf("expected metadata.user_id 42, got %s", body.Metadata["user_id"])
		}
		if body.Metadata["user_email"] != "alice@example.com" {
			t.Errorf("expected metadata.user_email alice@example.com, got %s", body.Metadata["user_email"])
		}

		resp := dodo.CheckoutSessionResponse{
			SessionID:   "cs_test_abc123",
			CheckoutURL: "https://checkout.dodopayments.com/pay/cs_test_abc123",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	c := &Client{
		checkoutSvc: dodo.NewCheckoutSessionService(
			option.WithBearerToken("test-api-key"),
			option.WithBaseURL(ts.URL),
		),
	}

	result, err := c.CreateCheckout(context.Background(), "pdt_xxx", 990, 42, "alice@example.com", "https://25types.com/donate")
	if err != nil {
		t.Fatalf("CreateCheckout: %v", err)
	}
	if result.SessionID != "cs_test_abc123" {
		t.Errorf("session ID mismatch: %s", result.SessionID)
	}
	if result.CheckoutURL != "https://checkout.dodopayments.com/pay/cs_test_abc123" {
		t.Errorf("checkout URL mismatch: %s", result.CheckoutURL)
	}
}

// checkoutRequestJSON mirrors the SDK's JSON shape for inspection in tests.
// CheckoutSessionNewParams.MarshalJSON unwraps to send CheckoutSessionRequest fields directly.
type checkoutRequestJSON struct {
	ProductCart []struct {
		ProductID string `json:"product_id"`
		Quantity  int64  `json:"quantity"`
		Amount    int64  `json:"amount"`
	} `json:"product_cart"`
	Metadata  map[string]string `json:"metadata"`
	ReturnURL string            `json:"return_url"`
}

func TestCreateCheckout_Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid product"})
	}))
	defer ts.Close()

	c := &Client{
		checkoutSvc: dodo.NewCheckoutSessionService(
			option.WithBearerToken("test-api-key"),
			option.WithBaseURL(ts.URL),
		),
	}

	_, err := c.CreateCheckout(context.Background(), "bad_product", 990, 1, "", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestVerifyWebhook_ValidSignature(t *testing.T) {
	webhookKey := "whsec_test1234"
	event := map[string]interface{}{
		"type": "payment.succeeded",
		"data": map[string]interface{}{
			"payment_id":  "pay_test123",
			"total_amount": 1990,
			"currency":     "usd",
			"metadata": map[string]string{
				"user_id":    "7",
				"user_email": "donor@example.com",
			},
		},
	}
	rawBody, _ := json.Marshal(event)

	webhookID := "msg_2cLvN5R8pQ"
	now := time.Now()
	webhookTimestamp := strconv.FormatInt(now.Unix(), 10)

	wh, err := standardwebhooks.NewWebhook(webhookKey)
	if err != nil {
		t.Fatalf("NewWebhook: %v", err)
	}
	sig, err := wh.Sign(webhookID, now, rawBody)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	headers := http.Header{}
	headers.Set("webhook-id", webhookID)
	headers.Set("webhook-timestamp", webhookTimestamp)
	headers.Set("webhook-signature", sig)

	c := &Client{
		webhookSvc: dodo.NewWebhookService(
			option.WithWebhookKey(webhookKey),
		),
	}

	evt, err := c.VerifyWebhook(rawBody, headers)
	if err != nil {
		t.Fatalf("VerifyWebhook: %v", err)
	}
	if evt.Type != "payment.succeeded" {
		t.Errorf("expected payment.succeeded, got %s", evt.Type)
	}
	if evt.Data.UserID != 7 {
		t.Errorf("expected user_id 7, got %d", evt.Data.UserID)
	}
	if evt.Data.Amount != 1990 {
		t.Errorf("expected amount 1990, got %d", evt.Data.Amount)
	}
	if evt.Data.Email != "donor@example.com" {
		t.Errorf("expected donor@example.com, got %s", evt.Data.Email)
	}
}

func TestVerifyWebhook_InvalidSignature(t *testing.T) {
	rawBody := []byte(`{"type":"payment.succeeded","data":{"metadata":{"user_id":"1"}}}`)
	webhookID := "msg_test"
	webhookTimestamp := strconv.FormatInt(time.Now().Unix(), 10)

	headers := http.Header{}
	headers.Set("webhook-id", webhookID)
	headers.Set("webhook-timestamp", webhookTimestamp)
	headers.Set("webhook-signature", "v1,badsignature1234")

	c := &Client{
		webhookSvc: dodo.NewWebhookService(
			option.WithWebhookKey("whsec_real_key"),
		),
	}

	_, err := c.VerifyWebhook(rawBody, headers)
	if err == nil {
		t.Fatal("expected error for invalid signature, got nil")
	}
}

func TestVerifyWebhook_MissingHeaders(t *testing.T) {
	c := &Client{
		webhookSvc: dodo.NewWebhookService(
			option.WithWebhookKey("whsec_test"),
		),
	}

	_, err := c.VerifyWebhook([]byte(`{}`), http.Header{})
	if err == nil {
		t.Fatal("expected error for missing headers, got nil")
	}
}

func TestVerifyWebhook_UnparsableUserID(t *testing.T) {
	webhookKey := "whsec_test"
	event := map[string]interface{}{
		"type": "payment.succeeded",
		"data": map[string]interface{}{
			"payment_id":  "pay_bad",
			"total_amount": 990,
			"currency":     "usd",
			"metadata": map[string]string{
				"user_id":    "not-a-number",
				"user_email": "x@y.com",
			},
		},
	}
	rawBody, _ := json.Marshal(event)

	webhookID := "msg_bad"
	now := time.Now()
	webhookTimestamp := strconv.FormatInt(now.Unix(), 10)

	wh, _ := standardwebhooks.NewWebhook(webhookKey)
	sig, _ := wh.Sign(webhookID, now, rawBody)

	headers := http.Header{}
	headers.Set("webhook-id", webhookID)
	headers.Set("webhook-timestamp", webhookTimestamp)
	headers.Set("webhook-signature", sig)

	c := &Client{
		webhookSvc: dodo.NewWebhookService(
			option.WithWebhookKey(webhookKey),
		),
	}

	_, err := c.VerifyWebhook(rawBody, headers)
	if err == nil {
		t.Fatal("expected error for invalid user_id, got nil")
	}
}

func TestVerifyWebhook_NonPaymentEvent(t *testing.T) {
	webhookKey := "whsec_test"
	event := map[string]interface{}{
		"type": "payment.failed",
		"data": map[string]interface{}{
			"payment_id": "pay_fail",
			"metadata":   map[string]string{},
		},
	}
	rawBody, _ := json.Marshal(event)

	webhookID := "msg_nonpay"
	now := time.Now()
	webhookTimestamp := strconv.FormatInt(now.Unix(), 10)

	wh, _ := standardwebhooks.NewWebhook(webhookKey)
	sig, _ := wh.Sign(webhookID, now, rawBody)

	headers := http.Header{}
	headers.Set("webhook-id", webhookID)
	headers.Set("webhook-timestamp", webhookTimestamp)
	headers.Set("webhook-signature", sig)

	c := &Client{
		webhookSvc: dodo.NewWebhookService(
			option.WithWebhookKey(webhookKey),
		),
	}

	evt, err := c.VerifyWebhook(rawBody, headers)
	if err != nil {
		t.Fatalf("VerifyWebhook: %v", err)
	}
	// Non-payment.succeeded events should return Type but zero Data
	if evt.Type != "payment.failed" {
		t.Errorf("expected payment.failed, got %s", evt.Type)
	}
}

// Compile-time interface checks
var _ commerce.CheckoutCreator = (*Client)(nil)
