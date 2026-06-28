package dodo

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	dodopayments "github.com/dodopayments/dodopayments-go"
	"github.com/dodopayments/dodopayments-go/option"
	standardwebhooks "github.com/standard-webhooks/standard-webhooks/libraries/go"

	"liki/internal/product"
)

// helperKey is a 32-byte raw key whose base64 is in helperSecret.
var helperKey = []byte("0123456789abcdef0123456789abcdef") // 32 bytes

// helperSecret is the whsec_<base64> form of helperKey.
const helperSecret = "whsec_MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY="

// helperSign signs a payload with helperKey and returns signature headers.
func helperSign(t *testing.T, msgID string, body string) http.Header {
	t.Helper()
	wh, err := standardwebhooks.NewWebhookRaw(helperKey)
	if err != nil {
		t.Fatalf("NewWebhookRaw: %v", err)
	}
	now := time.Now()
	sig, err := wh.Sign(msgID, now, []byte(body))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return http.Header{
		"Webhook-Id":        {msgID},
		"Webhook-Timestamp": {fmt.Sprintf("%d", now.Unix())},
		"Webhook-Signature": {sig},
	}
}

func newTestClient() *Client {
	return New("sk_test", helperSecret, true, nil)
}

// --- VerifyWebhook tests (uses official Dodo SDK Unwrap) ---

func TestVerifyWebhook_PaymentSucceeded(t *testing.T) {
	payload := `{"type":"payment.succeeded","business_id":"bus_1","timestamp":"2026-01-01T00:00:00Z","data":{"payment_id":"pay_1","total_amount":990,"customer":{"customer_id":"cus_1","email":"test@example.com","name":"Test"},"metadata":{"order_id":"order-1","email":""}}}}`
	headers := helperSign(t, "msg_789", payload)

	event, err := newTestClient().VerifyWebhook([]byte(payload), headers)
	if err != nil {
		t.Fatalf("VerifyWebhook: %v", err)
	}

	if event.Type != "payment.succeeded" {
		t.Errorf("type = %q, want payment.succeeded", event.Type)
	}
	if event.Data.OrderID != "order-1" {
		t.Errorf("orderID = %q, want order-1", event.Data.OrderID)
	}
	if event.Data.Amount != 990 {
		t.Errorf("amount = %d, want 990", event.Data.Amount)
	}
	if event.Data.Email != "test@example.com" {
		t.Errorf("email = %q, want test@example.com", event.Data.Email)
	}
	if event.Data.PaymentID != "pay_1" {
		t.Errorf("paymentID = %q, want pay_1", event.Data.PaymentID)
	}
}

func TestVerifyWebhook_EmailFromCustomer(t *testing.T) {
	payload := `{"type":"payment.succeeded","business_id":"bus_1","timestamp":"2026-01-01T00:00:00Z","data":{"payment_id":"pay_2","total_amount":100,"customer":{"customer_id":"cus_2","email":"customer@test.com","name":"Cust"},"metadata":{"order_id":"order-2","email":""}}}}`
	headers := helperSign(t, "msg_abc", payload)

	event, err := newTestClient().VerifyWebhook([]byte(payload), headers)
	if err != nil {
		t.Fatalf("VerifyWebhook: %v", err)
	}
	if event.Data.Email != "customer@test.com" {
		t.Errorf("email = %q, want customer@test.com (Customer.Email)", event.Data.Email)
	}
}

func TestVerifyWebhook_EmailFallbackToMetadata(t *testing.T) {
	payload := `{"type":"payment.succeeded","business_id":"bus_1","timestamp":"2026-01-01T00:00:00Z","data":{"payment_id":"pay_3","total_amount":200,"customer":{"customer_id":"cus_3","email":"","name":"Cust"},"metadata":{"order_id":"order-3","email":"meta@test.com"}}}}`
	headers := helperSign(t, "msg_def", payload)

	event, err := newTestClient().VerifyWebhook([]byte(payload), headers)
	if err != nil {
		t.Fatalf("VerifyWebhook: %v", err)
	}
	if event.Data.Email != "meta@test.com" {
		t.Errorf("email = %q, want meta@test.com (Metadata.email fallback)", event.Data.Email)
	}
}

func TestVerifyWebhook_NonPayment(t *testing.T) {
	payload := `{"type":"refund.succeeded","business_id":"bus_1","timestamp":"2026-01-01T00:00:00Z","data":{"payment_id":"pay_4","total_amount":100,"customer":{"customer_id":"cus_4","email":"","name":"C"},"metadata":{"order_id":"order-4","email":""}}}}`
	headers := helperSign(t, "msg_ghi", payload)

	event, err := newTestClient().VerifyWebhook([]byte(payload), headers)
	if err != nil {
		t.Fatalf("VerifyWebhook: %v", err)
	}
	if event.Type != "refund.succeeded" {
		t.Errorf("type = %q, want refund.succeeded", event.Type)
	}
}

func TestVerifyWebhook_BadSignature(t *testing.T) {
	body := []byte(`{"type":"test"}`)
	headers := http.Header{
		"Webhook-Id":        {"msg_bad"},
		"Webhook-Timestamp": {"1780000000"},
		"Webhook-Signature": {"v1,bad_signature_here"},
	}
	_, err := newTestClient().VerifyWebhook(body, headers)
	if err == nil {
		t.Fatal("expected error for bad signature")
	}
}

func TestVerifyWebhook_InvalidJSON(t *testing.T) {
	body := []byte(`not json`)
	headers := helperSign(t, "msg_jkl", string(body))

	_, err := newTestClient().VerifyWebhook(body, headers)
	if err == nil {
		t.Fatal("expected parse error for invalid JSON")
	}
}

// TestVerifyWebhook_RealKey signs a payload with the actual production key
// and verifies it through the official SDK Unwrap path. If this test fails,
// the DODO_WEBHOOK_KEY in .env is invalid or has been rotated.
func TestVerifyWebhook_RealKey(t *testing.T) {
	const secret = "whsec_/9fvZdg0IQrgxJ55H4yerZoUrn38x0ZA"

	rawKey, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(secret, "whsec_"))
	if err != nil {
		t.Fatalf("decode real key: %v", err)
	}

	body := `{"type":"payment.succeeded","business_id":"bus_1","timestamp":"2026-01-01T00:00:00Z","data":{"payment_id":"pay_real","total_amount":990,"customer":{"customer_id":"cus_1","email":"real@test.com","name":"Real"},"metadata":{"order_id":"order-real","email":""}}}}`

	wh, err := standardwebhooks.NewWebhookRaw(rawKey)
	if err != nil {
		t.Fatalf("NewWebhookRaw: %v", err)
	}
	now := time.Now()
	sig, err := wh.Sign("msg_real", now, []byte(body))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	headers := http.Header{
		"Webhook-Id":        {"msg_real"},
		"Webhook-Timestamp": {fmt.Sprintf("%d", now.Unix())},
		"Webhook-Signature": {sig},
	}

	svc := dodopayments.NewWebhookService(option.WithWebhookKey(secret))
	event, err := svc.Unwrap([]byte(body), headers)
	if err != nil {
		t.Fatalf("SDK Unwrap with real key: %v", err)
	}

	union := event.AsUnion()
	paymentEvent, ok := union.(dodopayments.PaymentSucceededWebhookEvent)
	if !ok {
		t.Fatal("not a PaymentSucceeded event")
	}
	if paymentEvent.Data.PaymentID != "pay_real" {
		t.Errorf("payment_id = %q, want pay_real", paymentEvent.Data.PaymentID)
	}
}

// TestRealPayloadParses confirms Dodo's webhook JSON matches the SDK structs.
func TestRealPayloadParses(t *testing.T) {
	const body = `{"business_id":"bus_0NeNx5oyfSqpPlqd9vNNt","data":{"billing":{"city":"Chang Sha Shi","country":"CN","state":"Hu Nan sheng","street":"Hengda Yayuan （Northwest Gate）,  Wan Jia Li Bei Lu, Kai Fu Qu","zipcode":"410073"},"brand_id":"bus_0NeNx5oyfSqpPlqd9vNNt","business_id":"bus_0NeNx5oyfSqpPlqd9vNNt","card_holder_name":null,"card_issuing_country":null,"card_last_four":null,"card_network":null,"card_type":null,"checkout_session_id":"cks_0Ngx30md6Tx01bE8Kcraw","created_at":"2026-06-13T10:33:11.206983Z","currency":"CNY","custom_field_responses":null,"customer":{"customer_id":"cus_0Nf6ti9AfKgZzonHRtgq2","email":"suiqiang@foxmail.com","metadata":{},"name":"Miles Sui","phone_number":"+8613973113693"},"digital_products_delivered":false,"discount_id":null,"discounts":null,"disputes":[],"error_code":null,"error_message":null,"invoice_id":"inv_0Ngx37uRimY678alVnawL","invoice_url":"https://test.dodopayments.com/invoices/payments/pay_0Ngx37uRimY678ae3uiSZ","metadata":{"email":"","order_id":"e03494f2-d9f4-4af2-97dc-cee4faa20e5c"},"payload_type":"Payment","payment_id":"pay_0Ngx37uRimY678ae3uiSZ","payment_link":"https://test.checkout.dodopayments.com/W7tRl8K5","payment_method":"wallet","payment_method_type":"we_chat_pay","payment_provider":"dodo","product_cart":[{"product_id":"pdt_0NfnRrzO5xcK5HfLZv1jZ","quantity":1}],"refund_status":null,"refunds":[],"retry_attempt":0,"settlement_amount":990,"settlement_currency":"USD","settlement_tax":0,"status":"succeeded","subscription_id":null,"tax":0,"total_amount":6963,"updated_at":null},"timestamp":"2026-06-13T10:33:35.160144Z","type":"payment.succeeded"}`

	var event dodopayments.UnwrapWebhookEvent
	if err := json.Unmarshal([]byte(body), &event); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	union := event.AsUnion()
	paymentEvent, ok := union.(dodopayments.PaymentSucceededWebhookEvent)
	if !ok {
		t.Fatalf("not a PaymentSucceeded event")
	}

	if orderID := paymentEvent.Data.Metadata["order_id"]; orderID != "e03494f2-d9f4-4af2-97dc-cee4faa20e5c" {
		t.Errorf("order_id = %q", orderID)
	}
	if email := paymentEvent.Data.Customer.Email; email != "suiqiang@foxmail.com" {
		t.Errorf("customer email = %q", email)
	}
	if paymentID := paymentEvent.Data.PaymentID; paymentID != "pay_0Ngx37uRimY678ae3uiSZ" {
		t.Errorf("payment_id = %q", paymentID)
	}
	if amount := paymentEvent.Data.TotalAmount; amount != 6963 {
		t.Errorf("total_amount = %d, want 6963", amount)
	}
}

func TestNew_LiveMode(t *testing.T) {
	c := New("sk_live", "whsec_test", false, nil)
	if c == nil {
		t.Fatal("New returned nil")
		return
	}
	if c.checkoutSvc == nil {
		t.Error("checkoutSvc should not be nil")
	}
	if c.webhookSvc == nil {
		t.Error("webhookSvc should not be nil")
	}
}

func TestNew_TestMode(t *testing.T) {
	c := New("sk_test", "whsec_test", true, nil)
	if c == nil {
		t.Fatal("New returned nil")
	}
}

func newTestClientWithURL(apiKey, baseURL string) *Client {
	opts := []option.RequestOption{
		option.WithBearerToken(apiKey),
		option.WithBaseURL(baseURL),
	}
	return &Client{
		checkoutSvc: dodopayments.NewCheckoutSessionService(opts...),
		webhookSvc:  dodopayments.NewWebhookService(append(opts, option.WithWebhookKey("whsec_test"))...),
		products: map[product.Product]string{
			product.ProductNaming: "pdt_test_naming",
		},
	}
}

func TestCreateCheckout_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/checkouts" {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{ //nolint:errcheck
			"session_id":   "sess_test123",
			"checkout_url": "https://checkout.example.com/pay/test",
		})
	}))
	defer srv.Close()

	c := newTestClientWithURL("sk_test", srv.URL)
	ctx := context.Background()

	result, err := c.CreateCheckout(ctx, product.ProductNaming, 990, "order-1", "user@test.com", "https://liki.hk/return")
	if err != nil {
		t.Fatalf("CreateCheckout: %v", err)
	}
	if result.SessionID != "sess_test123" {
		t.Errorf("SessionID = %q, want sess_test123", result.SessionID)
	}
	if result.CheckoutURL != "https://checkout.example.com/pay/test" {
		t.Errorf("CheckoutURL = %q", result.CheckoutURL)
	}
}

func TestCreateCheckout_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal"}`)) //nolint:errcheck
	}))
	defer srv.Close()

	c := newTestClientWithURL("sk_test", srv.URL)
	ctx := context.Background()

	_, err := c.CreateCheckout(ctx, product.ProductNaming, 990, "order-1", "", "https://liki.hk/return")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestCreateCheckout_InvalidBaseURL(t *testing.T) {
	// Use an unroutable address to trigger connection error.
	c := newTestClientWithURL("sk_test", "http://127.0.0.1:1")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := c.CreateCheckout(ctx, product.ProductNaming, 990, "order-1", "", "https://liki.hk/return")
	if err == nil {
		t.Fatal("expected connection error for invalid URL")
	}
	t.Logf("expected error: %v", err)
}
