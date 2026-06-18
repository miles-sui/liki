package dodo

import (
	"context"
	"fmt"
	"net/http"

	dodopayments "github.com/dodopayments/dodopayments-go"
	"github.com/dodopayments/dodopayments-go/option"
)

// Client wraps the Dodo Payments SDK for checkout and webhook operations.
type Client struct {
	checkoutSvc *dodopayments.CheckoutSessionService
	webhookSvc  *dodopayments.WebhookService
}

// CheckoutResult holds the result of creating a checkout session.
type CheckoutResult struct {
	SessionID   string `json:"session_id"`
	CheckoutURL string `json:"checkout_url"`
}

// WebhookEvent is a verified webhook event from Dodo Payments.
type WebhookEvent struct {
	Type string
	Data WebhookEventData
}

// WebhookEventData holds the extracted fields from a webhook event.
type WebhookEventData struct {
	OrderID   string
	Amount    int
	Email     string
	PaymentID string
}

// New creates a Dodo Payments client with the given API key and webhook secret.
func New(apiKey, webhookKey string, testMode bool) *Client {
	opts := []option.RequestOption{option.WithBearerToken(apiKey)}
	if testMode {
		opts = append(opts, option.WithEnvironmentTestMode())
	} else {
		opts = append(opts, option.WithEnvironmentLiveMode())
	}
	return &Client{
		checkoutSvc: dodopayments.NewCheckoutSessionService(opts...),
		webhookSvc:  dodopayments.NewWebhookService(append(opts, option.WithWebhookKey(webhookKey))...),
	}
}

// CreateCheckout creates a Dodo Payments checkout session for the given order.
func (c *Client) CreateCheckout(ctx context.Context, productID string, amount int, orderID, email, returnURL string) (*CheckoutResult, error) {
	params := dodopayments.CheckoutSessionNewParams{
		CheckoutSessionRequest: dodopayments.CheckoutSessionRequestParam{
			ProductCart: dodopayments.F([]dodopayments.ProductItemReqParam{{
				ProductID: dodopayments.F(productID),
				Quantity:  dodopayments.F(int64(1)),
				Amount:    dodopayments.F(int64(amount)),
			}}),
			Metadata: dodopayments.F(map[string]string{
				"order_id": orderID,
				"email":    email,
			}),
			ReturnURL: dodopayments.F(returnURL),
		},
	}
	resp, err := c.checkoutSvc.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("dodo: create checkout: %w", err)
	}
	return &CheckoutResult{SessionID: resp.SessionID, CheckoutURL: resp.CheckoutURL}, nil
}

// VerifyWebhook verifies and parses a Dodo Payments webhook request.
func (c *Client) VerifyWebhook(rawBody []byte, headers http.Header) (*WebhookEvent, error) {
	event, err := c.webhookSvc.Unwrap(rawBody, headers)
	if err != nil {
		return nil, fmt.Errorf("dodo: verify webhook: %w", err)
	}

	union := event.AsUnion()
	paymentEvent, ok := union.(dodopayments.PaymentSucceededWebhookEvent)
	if !ok {
		return &WebhookEvent{Type: string(event.Type)}, nil
	}

	email := paymentEvent.Data.Customer.Email
	if email == "" {
		email = paymentEvent.Data.Metadata["email"]
	}
	return &WebhookEvent{
		Type: string(event.Type),
		Data: WebhookEventData{
			OrderID:   paymentEvent.Data.Metadata["order_id"],
			Amount:    int(paymentEvent.Data.TotalAmount),
			Email:     email,
			PaymentID: paymentEvent.Data.PaymentID,
		},
	}, nil
}
