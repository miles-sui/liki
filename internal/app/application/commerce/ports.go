package commerce

import (
	"context"
	"net/http"
)

// DonationRepository defines persistence for donation records.
type DonationRepository interface {
	CreateDonation(ctx context.Context, userID int64, amount int, paymentID string) error
}

// ThankYouSender defines the thank-you email delivery interface.
type ThankYouSender interface {
	SendThankYouEmail(ctx context.Context, to, locale string) error
}

// CheckoutResult is the result of creating a checkout session.
type CheckoutResult struct {
	SessionID   string
	CheckoutURL string
}

// CheckoutCreator defines the checkout session creation interface.
type CheckoutCreator interface {
	CreateCheckout(ctx context.Context, productID string, amount int, userID int64, userEmail string, returnURL string) (*CheckoutResult, error)
}

// WebhookVerifier defines the webhook signature verification interface.
type WebhookVerifier interface {
	VerifyWebhook(rawBody []byte, headers http.Header) (*WebhookEvent, error)
}

// PaymentResult is the result of a payment lookup from the provider.
type PaymentResult struct {
	PaymentID string
	Status    string
	Amount    int
	Metadata  map[string]string
}

// PaymentLookup defines the payment lookup interface.
type PaymentLookup interface {
	GetPayment(ctx context.Context, paymentID string) (*PaymentResult, error)
}

// PaymentProvider combines checkout creation, webhook verification, and payment lookup.
type PaymentProvider interface {
	CheckoutCreator
	WebhookVerifier
	PaymentLookup
}

// WebhookEvent is a verified Dodo webhook event.
type WebhookEvent struct {
	Type string
	Data WebhookEventData
}

// WebhookEventData holds the relevant payment data from a webhook.
type WebhookEventData struct {
	UserID    int64
	Amount    int
	Email     string
	PaymentID string
}
