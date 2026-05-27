package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/25types/25types/internal/app/application/commerce"
	"github.com/25types/25types/internal/app/db"
	"github.com/25types/25types/internal/app/sqlite"
)

// fakeDodoClient implements commerce.PaymentProvider for tests.
type fakeDodoClient struct {
	createCheckoutFn func(ctx context.Context, productID string, amount int, userID int64, userEmail string, returnURL string) (*commerce.CheckoutResult, error)
	verifyWebhookFn  func(rawBody []byte, headers http.Header) (*commerce.WebhookEvent, error)
	getPaymentFn     func(ctx context.Context, paymentID string) (*commerce.PaymentResult, error)
}

func (f *fakeDodoClient) CreateCheckout(ctx context.Context, productID string, amount int, userID int64, userEmail string, returnURL string) (*commerce.CheckoutResult, error) {
	if f.createCheckoutFn != nil {
		return f.createCheckoutFn(ctx, productID, amount, userID, userEmail, returnURL)
	}
	return &commerce.CheckoutResult{SessionID: "cs_test", CheckoutURL: "https://checkout.example.com/pay/cs_test"}, nil
}

func (f *fakeDodoClient) GetPayment(ctx context.Context, paymentID string) (*commerce.PaymentResult, error) {
	if f.getPaymentFn != nil {
		return f.getPaymentFn(ctx, paymentID)
	}
	return &commerce.PaymentResult{PaymentID: paymentID, Status: "succeeded", Amount: 1990}, nil
}

func (f *fakeDodoClient) VerifyWebhook(rawBody []byte, headers http.Header) (*commerce.WebhookEvent, error) {
	if f.verifyWebhookFn != nil {
		return f.verifyWebhookFn(rawBody, headers)
	}
	// Default: return a payment.succeeded event for user 1
	return &commerce.WebhookEvent{
		Type: "payment.succeeded",
		Data: commerce.WebhookEventData{
			UserID:    1,
			Amount:    1990,
			Email:     "donor@example.com",
			PaymentID: "pay_default",
		},
	}, nil
}

type fakeThankYouSender struct {
	sent []thankYouCall
}

type thankYouCall struct {
	to     string
	locale string
}

func (f *fakeThankYouSender) SendThankYouEmail(ctx context.Context, to, locale string) error {
	f.sent = append(f.sent, thankYouCall{to: to, locale: locale})
	return nil
}

// newPaymentTestServer creates a test server with payment dependencies wired.
func newPaymentTestServer(t *testing.T) (*httptest.Server, *fakeDodoClient, *fakeThankYouSender) {
	t.Helper()

	database, err := db.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { database.Close() })

	userRepo := sqlite.NewUserRepo(database)
	assRepo := sqlite.NewAssessmentRepo(database)
	rlRepo := sqlite.NewReviewLinkRepo(database)
	profileRepo := sqlite.NewProfileRepo(userRepo, assRepo)
	matchLinkRepo := sqlite.NewMatchLinkRepo(database)

	dodoClient := &fakeDodoClient{}
	donationRepo := sqlite.NewDonationRepo(database)
	thankYou := &fakeThankYouSender{}

	cfg := ServerConfig{
		UserRepo:          userRepo,
		UserHasher:        sqlite.PasswordHasher{},
		AssRepo:           assRepo,
		LinkRepo:          rlRepo,
		SubRepo:           rlRepo,
		Profiles:          assRepo,
		ProfileRepo:       profileRepo,
		ProfileUsers:      profileRepo,
		BondStore:         profileRepo,
		MatchLinkRepo:     matchLinkRepo,
		UserEmailSender:   nil,
		TokenValidator:    userRepo,
		UserLookup:        userRepo,
		ExportRepo:        userRepo,
		DB:                database,

		DodoClient:     dodoClient,
		DonationRepo:   donationRepo,
		ThankYouSender: thankYou,
		DodoProductID:  "pdt_test_donation",
		UserEmailFn: func(ctx context.Context, userID int64) (string, bool) {
			u, err := userRepo.FindByID(ctx, userID)
			if err != nil || u == nil {
				return "", false
			}
			if u.Email != "" {
				return u.Email, true
			}
			if u.PendingEmail != nil && *u.PendingEmail != "" {
				return *u.PendingEmail, true
			}
			return "", false
		},
	}

	mux := http.NewServeMux()
	RegisterRoutes(mux, cfg)
	return httptest.NewServer(mux), dodoClient, thankYou
}

func TestCheckout_NoAuth(t *testing.T) {
	srv, _, _ := newPaymentTestServer(t)
	defer srv.Close()

	code, body := doReq(t, "POST", srv.URL+"/api/payments/checkout", `{"amount":990}`, "")
	if code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d: %s", code, body)
	}
}

func TestCheckout_InvalidAmount(t *testing.T) {
	srv, _, _ := newPaymentTestServer(t)
	defer srv.Close()

	token, _ := registerAndLogin(t, srv, "donor-amount", "testpass")

	tests := []struct {
		name   string
		body   string
		errMsg string
	}{
		{"zero", `{"amount":0}`, "must be 990, 1990, or 2990"},
		{"negative", `{"amount":-1}`, "must be 990, 1990, or 2990"},
		{"invalid", `{"amount":500}`, "must be 990, 1990, or 2990"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			code, body := doReq(t, "POST", srv.URL+"/api/payments/checkout", tc.body, token)
			if code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d: %s", code, body)
			}
			if !strings.Contains(body, tc.errMsg) {
				t.Errorf("expected error containing %q, got: %s", tc.errMsg, body)
			}
		})
	}
}

func TestCheckout_ValidAmounts(t *testing.T) {
	srv, _, _ := newPaymentTestServer(t)
	defer srv.Close()

	token, _ := registerAndLogin(t, srv, "donor-valid", "testpass")

	validAmounts := []int{990, 1990, 2990}
	for _, amount := range validAmounts {
		t.Run(string(rune(amount)), func(t *testing.T) {
			reqBody, _ := json.Marshal(map[string]int{"amount": amount})
			code, body := doReq(t, "POST", srv.URL+"/api/payments/checkout", string(reqBody), token)
			if code != http.StatusOK {
				t.Errorf("amount %d: expected 200, got %d: %s", amount, code, body)
			}
			data := envelopeOk(t, body)
			if data["url"] == nil || data["url"].(string) == "" {
				t.Errorf("amount %d: expected non-empty url, got %v", amount, data["url"])
			}
		})
	}
}

func TestCheckout_DodoDisabled(t *testing.T) {
	// Create server without Dodo client to test graceful degradation.
	database, err := db.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	defer database.Close()

	userRepo := sqlite.NewUserRepo(database)
	assRepo := sqlite.NewAssessmentRepo(database)
	rlRepo := sqlite.NewReviewLinkRepo(database)
	profileRepo := sqlite.NewProfileRepo(userRepo, assRepo)
	matchLinkRepo := sqlite.NewMatchLinkRepo(database)

	cfg := ServerConfig{
		UserRepo:          userRepo,
		UserHasher:        sqlite.PasswordHasher{},
		AssRepo:           assRepo,
		LinkRepo:          rlRepo,
		SubRepo:           rlRepo,
		Profiles:          assRepo,
		ProfileRepo:       profileRepo,
		ProfileUsers:      profileRepo,
		BondStore:         profileRepo,
		MatchLinkRepo:     matchLinkRepo,
		TokenValidator:    userRepo,
		UserLookup:        userRepo,
		ExportRepo:        userRepo,
		DB:                database,
		// DodoClient is nil — payment service not configured
	}

	mux := http.NewServeMux()
	RegisterRoutes(mux, cfg)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	token, _ := registerAndLogin(t, srv, "no-dodo", "testpass")
	code, body := doReq(t, "POST", srv.URL+"/api/payments/checkout", `{"amount":990}`, token)
	if code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d: %s", code, body)
	}
}

func TestWebhook_PaymentSucceeded(t *testing.T) {
	srv, dodo, thankYou := newPaymentTestServer(t)
	defer srv.Close()

	// Register a user — the default VerifyWebhook returns event for userID=1.
	token, _ := registerAndLogin(t, srv, "webhook-donor", "testpass")
	_ = token

	code, body := doReq(t, "POST", srv.URL+"/api/payments/webhook", `{"event_type":"payment.succeeded"}`, "")
	if code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", code, body)
	}

	// Verify thank-you email was sent.
	if len(thankYou.sent) != 1 {
		t.Errorf("expected 1 thank-you email, got %d", len(thankYou.sent))
	} else {
		if thankYou.sent[0].to != "donor@example.com" {
			t.Errorf("expected email to donor@example.com, got %s", thankYou.sent[0].to)
		}
	}

	_ = dodo
}

func TestWebhook_InvalidSignature(t *testing.T) {
	srv, dodo, _ := newPaymentTestServer(t)
	defer srv.Close()

	dodo.verifyWebhookFn = func(rawBody []byte, headers http.Header) (*commerce.WebhookEvent, error) {
		return nil, fmt.Errorf("invalid signature")
	}

	code, body := doReq(t, "POST", srv.URL+"/api/payments/webhook", `{"event_type":"payment.succeeded"}`, "")
	if code != http.StatusUnauthorized {
		t.Errorf("expected 401 for invalid webhook, got %d: %s", code, body)
	}
}

func TestCheckout_ReturnURLUsesHost(t *testing.T) {
	srv, dodo, _ := newPaymentTestServer(t)
	defer srv.Close()

	token, _ := registerAndLogin(t, srv, "donor-url", "testpass")

	var capturedReturnURL string
	dodo.createCheckoutFn = func(ctx context.Context, productID string, amount int, userID int64, userEmail string, returnURL string) (*commerce.CheckoutResult, error) {
		capturedReturnURL = returnURL
		return &commerce.CheckoutResult{SessionID: "cs_test", CheckoutURL: "https://example.com"}, nil
	}

	reqBody := `{"amount": 1990}`
	code, _ := doReq(t, "POST", srv.URL+"/api/payments/checkout", reqBody, token)
	if code != http.StatusOK {
		t.Fatalf("expected 200, got %d", code)
	}

	// The return URL must use the test server's host, not a hardcoded domain.
	if !strings.HasPrefix(capturedReturnURL, srv.URL+"/") {
		t.Errorf("expected returnURL to start with %q, got %q", srv.URL+"/", capturedReturnURL)
	}
}

func TestCheckout_UsesUserEmail(t *testing.T) {
	srv, dodo, _ := newPaymentTestServer(t)
	defer srv.Close()

	token, _ := registerAndLogin(t, srv, "donor-email", "testpass")

	var capturedEmail string
	dodo.createCheckoutFn = func(ctx context.Context, productID string, amount int, userID int64, userEmail string, returnURL string) (*commerce.CheckoutResult, error) {
		capturedEmail = userEmail
		return &commerce.CheckoutResult{SessionID: "cs_test", CheckoutURL: "https://example.com"}, nil
	}

	reqBody := `{"amount": 1990}`
	code, _ := doReq(t, "POST", srv.URL+"/api/payments/checkout", reqBody, token)
	if code != http.StatusOK {
		t.Fatalf("expected 200, got %d", code)
	}

	// User registered with "donor-email@test.com" by registerAndLogin
	expectedEmail := "donor-email@test.com"
	if capturedEmail != expectedEmail {
		t.Errorf("expected user email %q, got %q", expectedEmail, capturedEmail)
	}
}
