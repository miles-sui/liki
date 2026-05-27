package dodo

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/25types/25types/internal/app/application/commerce"
	dodopayments "github.com/dodopayments/dodopayments-go"
	"github.com/dodopayments/dodopayments-go/option"
)

// Client wraps the Dodo Payments official Go SDK for checkout creation, webhook
// verification, and payment lookup. It implements commerce.PaymentProvider.
type Client struct {
	checkoutSvc *dodopayments.CheckoutSessionService
	webhookSvc  *dodopayments.WebhookService
	paymentSvc  *dodopayments.PaymentService
}

// New creates a Dodo Payments client. testMode selects test or live environment.
func New(apiKey, webhookKey string, testMode bool) *Client {
	opts := []option.RequestOption{
		option.WithBearerToken(apiKey),
	}
	if testMode {
		opts = append(opts, option.WithEnvironmentTestMode())
	} else {
		opts = append(opts, option.WithEnvironmentLiveMode())
	}

	return &Client{
		checkoutSvc: dodopayments.NewCheckoutSessionService(opts...),
		webhookSvc: dodopayments.NewWebhookService(
			append(opts, option.WithWebhookKey(webhookKey))...,
		),
		paymentSvc: dodopayments.NewPaymentService(opts...),
	}
}

// CreateCheckout creates a Dodo checkout session for a one-time donation.
func (c *Client) CreateCheckout(ctx context.Context, productID string, amount int, userID int64, userEmail string, returnURL string) (*commerce.CheckoutResult, error) {
	params := dodopayments.CheckoutSessionNewParams{
		CheckoutSessionRequest: dodopayments.CheckoutSessionRequestParam{
			ProductCart: dodopayments.F([]dodopayments.ProductItemReqParam{{
				ProductID: dodopayments.F(productID),
				Quantity:  dodopayments.F(int64(1)),
				Amount:    dodopayments.F(int64(amount)),
			}}),
			Metadata: dodopayments.F(map[string]string{
				"user_id":    strconv.FormatInt(userID, 10),
				"user_email": userEmail,
			}),
			ReturnURL: dodopayments.F(returnURL),
		},
	}

	resp, err := c.checkoutSvc.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("dodo: create checkout: %w", err)
	}

	return &commerce.CheckoutResult{
		SessionID:   resp.SessionID,
		CheckoutURL: resp.CheckoutURL,
	}, nil
}

// VerifyWebhook verifies the Dodo webhook signature (Standard Webhooks HMAC-SHA256)
// and returns the parsed event. Only payment.succeeded events contain parsed user data;
// other events return the event type with zero-value Data.
func (c *Client) VerifyWebhook(rawBody []byte, headers http.Header) (*commerce.WebhookEvent, error) {
	event, err := c.webhookSvc.Unwrap(rawBody, headers)
	if err != nil {
		return nil, fmt.Errorf("dodo: verify webhook: %w", err)
	}

	union := event.AsUnion()
	paymentEvent, ok := union.(dodopayments.PaymentSucceededWebhookEvent)
	if !ok {
		return &commerce.WebhookEvent{
			Type: string(event.Type),
		}, nil
	}

	userID, err := strconv.ParseInt(paymentEvent.Data.Metadata["user_id"], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("dodo: invalid user_id in webhook metadata: %w", err)
	}

	return &commerce.WebhookEvent{
		Type: string(event.Type),
		Data: commerce.WebhookEventData{
			UserID:    userID,
			Amount:    int(paymentEvent.Data.TotalAmount),
			Email:     paymentEvent.Data.Metadata["user_email"],
			PaymentID: paymentEvent.Data.PaymentID,
		},
	}, nil
}

// GetPayment retrieves a payment by ID from Dodo.
func (c *Client) GetPayment(ctx context.Context, paymentID string) (*commerce.PaymentResult, error) {
	payment, err := c.paymentSvc.Get(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("dodo: get payment: %w", err)
	}
	return &commerce.PaymentResult{
		PaymentID: payment.PaymentID,
		Status:    string(payment.Status),
		Amount:    int(payment.TotalAmount),
		Metadata:  payment.Metadata,
	}, nil
}

// Compile-time interface checks
var _ commerce.CheckoutCreator = (*Client)(nil)
var _ commerce.WebhookVerifier = (*Client)(nil)
var _ commerce.PaymentLookup = (*Client)(nil)
