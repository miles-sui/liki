package commerce

import (
	"context"
	"errors"
	"testing"
)

// =============================================================================
// Stubs
// =============================================================================

type stubCheckoutCreator struct {
	result *CheckoutResult
	err    error
}

func (s *stubCheckoutCreator) CreateCheckout(ctx context.Context, productID string, amount int, userID int64, userEmail string, returnURL string) (*CheckoutResult, error) {
	return s.result, s.err
}

type stubDonationRepo struct {
	donations []donationRecord
	err       error
}

type donationRecord struct {
	userID    int64
	amount    int
	paymentID string
}

func (r *stubDonationRepo) CreateDonation(ctx context.Context, userID int64, amount int, paymentID string) error {
	if r.err != nil {
		return r.err
	}
	r.donations = append(r.donations, donationRecord{userID, amount, paymentID})
	return nil
}

type stubThankYouSender struct {
	sent []thankYouRecord
	err  error
}

type thankYouRecord struct{ to, locale string }

func (s *stubThankYouSender) SendThankYouEmail(ctx context.Context, to, locale string) error {
	if s.err != nil {
		return s.err
	}
	s.sent = append(s.sent, thankYouRecord{to, locale})
	return nil
}

type stubPaymentLookup struct {
	result *PaymentResult
	err    error
}

func (l *stubPaymentLookup) GetPayment(ctx context.Context, paymentID string) (*PaymentResult, error) {
	return l.result, l.err
}

// =============================================================================
// ValidateAmount
// =============================================================================

func TestValidateAmount(t *testing.T) {
	tests := []struct {
		amount int
		want   bool
	}{
		{990, true},
		{1990, true},
		{2990, true},
		{0, false},
		{-990, false},
		{1000, false},
		{5000, false},
	}
	for _, tc := range tests {
		got := ValidateAmount(tc.amount)
		if got != tc.want {
			t.Errorf("ValidateAmount(%d) = %v, want %v", tc.amount, got, tc.want)
		}
	}
}

// =============================================================================
// CreateDonationCheckout
// =============================================================================

func TestCreateDonationCheckout_OK(t *testing.T) {
	client := &stubCheckoutCreator{result: &CheckoutResult{
		SessionID:   "sess-1",
		CheckoutURL: "https://checkout.example.com/pay",
	}}
	out, err := CreateDonationCheckout(context.Background(), client, "pdt_test", CreateCheckoutInput{
		UserID:    1,
		UserEmail: "a@b.com",
		Amount:    990,
		ReturnURL: "https://example.com/donate",
	})
	if err != nil {
		t.Fatalf("CreateDonationCheckout: %v", err)
	}
	if out.URL != "https://checkout.example.com/pay" {
		t.Errorf("URL = %q", out.URL)
	}
	if out.SessionID != "sess-1" {
		t.Errorf("SessionID = %q", out.SessionID)
	}
}

func TestCreateDonationCheckout_InvalidAmount(t *testing.T) {
	client := &stubCheckoutCreator{}
	_, err := CreateDonationCheckout(context.Background(), client, "pdt_test", CreateCheckoutInput{
		Amount: 500,
	})
	if err == nil {
		t.Fatal("expected error for invalid amount")
	}
}

func TestCreateDonationCheckout_ClientError(t *testing.T) {
	client := &stubCheckoutCreator{err: errors.New("dodo: network error")}
	_, err := CreateDonationCheckout(context.Background(), client, "pdt_test", CreateCheckoutInput{
		Amount: 990,
	})
	if err == nil {
		t.Fatal("expected error from client")
	}
}

// =============================================================================
// HandleDonationWebhook
// =============================================================================

func TestHandleDonationWebhook_PaymentSucceeded(t *testing.T) {
	repo := &stubDonationRepo{}
	sender := &stubThankYouSender{}
	event := &WebhookEvent{
		Type: "payment.succeeded",
		Data: WebhookEventData{
			UserID:    42,
			Amount:    1990,
			Email:     "donor@example.com",
			PaymentID: "pay-123",
		},
	}
	err := HandleDonationWebhook(context.Background(), repo, sender, event, "en")
	if err != nil {
		t.Fatalf("HandleDonationWebhook: %v", err)
	}
	if len(repo.donations) != 1 {
		t.Fatalf("expected 1 donation, got %d", len(repo.donations))
	}
	d := repo.donations[0]
	if d.userID != 42 || d.amount != 1990 || d.paymentID != "pay-123" {
		t.Errorf("donation = %+v, want {42, 1990, pay-123}", d)
	}
	if len(sender.sent) != 1 {
		t.Errorf("expected 1 thank-you email, got %d", len(sender.sent))
	}
}

func TestHandleDonationWebhook_NonPaymentEvent(t *testing.T) {
	repo := &stubDonationRepo{}
	event := &WebhookEvent{Type: "payment.refunded"}
	err := HandleDonationWebhook(context.Background(), repo, nil, event, "en")
	if err != nil {
		t.Fatalf("HandleDonationWebhook: %v", err)
	}
	if len(repo.donations) != 0 {
		t.Error("expected no donation for non-payment event")
	}
}

func TestHandleDonationWebhook_RepoError(t *testing.T) {
	repo := &stubDonationRepo{err: errors.New("db down")}
	event := &WebhookEvent{Type: "payment.succeeded", Data: WebhookEventData{
		UserID: 1, PaymentID: "pay-1",
	}}
	err := HandleDonationWebhook(context.Background(), repo, nil, event, "en")
	if err == nil {
		t.Fatal("expected error from repo")
	}
}

func TestHandleDonationWebhook_NilSender(t *testing.T) {
	repo := &stubDonationRepo{}
	event := &WebhookEvent{Type: "payment.succeeded", Data: WebhookEventData{
		UserID: 1, Amount: 990, PaymentID: "pay-2",
	}}
	err := HandleDonationWebhook(context.Background(), repo, nil, event, "en")
	if err != nil {
		t.Fatalf("HandleDonationWebhook with nil sender: %v", err)
	}
	if len(repo.donations) != 1 {
		t.Error("expected donation even with nil sender")
	}
}

func TestHandleDonationWebhook_SenderError(t *testing.T) {
	repo := &stubDonationRepo{}
	sender := &stubThankYouSender{err: errors.New("smtp down")}
	event := &WebhookEvent{Type: "payment.succeeded", Data: WebhookEventData{
		UserID: 1, Amount: 990, PaymentID: "pay-3",
	}}
	// Sender error is best-effort — should not fail the webhook.
	err := HandleDonationWebhook(context.Background(), repo, sender, event, "en")
	if err != nil {
		t.Fatalf("HandleDonationWebhook with sender error: %v", err)
	}
	if len(repo.donations) != 1 {
		t.Error("expected donation even when thank-you email fails")
	}
}

// =============================================================================
// ConfirmDonation
// =============================================================================

func TestConfirmDonation_Succeeded(t *testing.T) {
	lookup := &stubPaymentLookup{result: &PaymentResult{
		PaymentID: "pay-1",
		Status:    "succeeded",
		Amount:    2990,
	}}
	repo := &stubDonationRepo{}
	result, err := ConfirmDonation(context.Background(), lookup, repo, ConfirmDonationInput{
		UserID: 7, PaymentID: "pay-1",
	})
	if err != nil {
		t.Fatalf("ConfirmDonation: %v", err)
	}
	if !result.Confirmed {
		t.Error("expected Confirmed=true")
	}
	if len(repo.donations) != 1 {
		t.Errorf("expected 1 donation, got %d", len(repo.donations))
	}
}

func TestConfirmDonation_NotSucceeded(t *testing.T) {
	lookup := &stubPaymentLookup{result: &PaymentResult{
		PaymentID: "pay-2", Status: "pending", Amount: 990,
	}}
	repo := &stubDonationRepo{}
	result, err := ConfirmDonation(context.Background(), lookup, repo, ConfirmDonationInput{
		UserID: 7, PaymentID: "pay-2",
	})
	if err != nil {
		t.Fatalf("ConfirmDonation: %v", err)
	}
	if result.Confirmed {
		t.Error("expected Confirmed=false for pending payment")
	}
	if len(repo.donations) != 0 {
		t.Error("expected no donation for non-succeeded payment")
	}
}

func TestConfirmDonation_LookupError(t *testing.T) {
	lookup := &stubPaymentLookup{err: errors.New("dodo: not found")}
	_, err := ConfirmDonation(context.Background(), lookup, &stubDonationRepo{}, ConfirmDonationInput{
		UserID: 7, PaymentID: "pay-missing",
	})
	if err == nil {
		t.Fatal("expected error from payment lookup")
	}
}

func TestConfirmDonation_RepoError(t *testing.T) {
	lookup := &stubPaymentLookup{result: &PaymentResult{
		PaymentID: "pay-3", Status: "succeeded", Amount: 1990,
	}}
	repo := &stubDonationRepo{err: errors.New("db down")}
	_, err := ConfirmDonation(context.Background(), lookup, repo, ConfirmDonationInput{
		UserID: 7, PaymentID: "pay-3",
	})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}
