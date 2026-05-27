package commerce

import (
	"context"
	"fmt"
	"log"
)

// CreateCheckoutInput holds the parameters for creating a donation checkout.
type CreateCheckoutInput struct {
	UserID    int64
	UserEmail string
	Amount    int // cents: 990, 1990, or 2990
	Locale    string
	ReturnURL string
}

// CreateCheckoutOutput holds the checkout session URL.
type CreateCheckoutOutput struct {
	URL       string `json:"url"`
	SessionID string `json:"session_id"`
}

var validAmounts = map[int]bool{990: true, 1990: true, 2990: true}

// CreateDonationCheckout creates a Dodo checkout session for a donation.
func CreateDonationCheckout(
	ctx context.Context,
	dodoClient CheckoutCreator,
	productID string,
	input CreateCheckoutInput,
) (*CreateCheckoutOutput, error) {
	if !validAmounts[input.Amount] {
		return nil, fmt.Errorf("invalid donation amount: %d", input.Amount)
	}
	result, err := dodoClient.CreateCheckout(ctx, productID, input.Amount, input.UserID, input.UserEmail, input.ReturnURL)
	if err != nil {
		return nil, err
	}
	return &CreateCheckoutOutput{
		URL:       result.CheckoutURL,
		SessionID: result.SessionID,
	}, nil
}

// HandleDonationWebhook processes a verified webhook event.
// Writes donation record and sends best-effort thank-you email.
func HandleDonationWebhook(
	ctx context.Context,
	repo DonationRepository,
	sender ThankYouSender,
	event *WebhookEvent,
	locale string,
) error {
	if event.Type != "payment.succeeded" {
		return nil
	}
	if err := repo.CreateDonation(ctx, event.Data.UserID, event.Data.Amount, event.Data.PaymentID); err != nil {
		return err
	}
	if sender != nil {
		if err := sender.SendThankYouEmail(ctx, event.Data.Email, locale); err != nil {
			log.Printf("[commerce] thank-you email failed for user %d: %v", event.Data.UserID, err)
		}
	}
	return nil
}

// ConfirmDonationInput holds parameters for confirming a payment and recording a donation.
type ConfirmDonationInput struct {
	UserID    int64
	PaymentID string
}

// ConfirmDonationResult holds the result of payment confirmation.
type ConfirmDonationResult struct {
	Confirmed bool `json:"confirmed"`
}

// ConfirmDonation looks up the payment via the Dodo API and, if the status is "succeeded",
// creates a donation record. The payment_id column (UNIQUE) ensures idempotency — if the
// webhook already processed this payment, the INSERT is a silent no-op.
func ConfirmDonation(
	ctx context.Context,
	lookup PaymentLookup,
	repo DonationRepository,
	input ConfirmDonationInput,
) (*ConfirmDonationResult, error) {
	payment, err := lookup.GetPayment(ctx, input.PaymentID)
	if err != nil {
		return nil, fmt.Errorf("confirm donation: lookup payment: %w", err)
	}

	if payment.Status != "succeeded" {
		return &ConfirmDonationResult{Confirmed: false}, nil
	}

	if err := repo.CreateDonation(ctx, input.UserID, payment.Amount, payment.PaymentID); err != nil {
		return nil, fmt.Errorf("confirm donation: create donation: %w", err)
	}

	return &ConfirmDonationResult{Confirmed: true}, nil
}

// ValidateAmount checks that the amount is one of the allowed donation tiers.
func ValidateAmount(amount int) bool {
	return validAmounts[amount]
}
