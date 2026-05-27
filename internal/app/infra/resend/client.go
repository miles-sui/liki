package resend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/25types/25types/internal/app/application/commerce"
	"github.com/25types/25types/internal/app/application/user"
	"github.com/25types/25types/internal/app/infra/emailtpl"
)

var (
	_ user.EmailSender      = (*Client)(nil)
	_ commerce.ThankYouSender = (*Client)(nil)
)

const defaultAPI = "https://api.resend.com"

// Client sends transactional emails via the Resend API.
type Client struct {
	apiKey     string
	from       string
	baseURL    string
	httpClient *http.Client
}

// New creates a Resend client. from is the sender address (e.g. "noreply@25types.com").
func New(apiKey, from string) *Client {
	return &Client{
		apiKey:  apiKey,
		from:    from,
		baseURL: defaultAPI,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Email is a single transactional email.
type Email struct {
	To      string
	Subject string
	Text    string
}

// Send delivers an email. Returns nil on success (Resend returns 200/201).
func (c *Client) Send(ctx context.Context, email Email) error {
	body := struct {
		From    string `json:"from"`
		To      string `json:"to"`
		Subject string `json:"subject"`
		Text    string `json:"text"`
	}{
		From:    c.from,
		To:      email.To,
		Subject: email.Subject,
		Text:    email.Text,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("resend: marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx,
		http.MethodPost, c.baseURL+"/emails", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("resend: new request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("resend: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("resend: %d: %s", resp.StatusCode, string(b))
	}
	return nil
}

// SendVerificationEmail sends an email verification link.
func (c *Client) SendVerificationEmail(ctx context.Context, to, token, locale string) error {
	subject, text := emailtpl.VerificationEmail(locale, token)
	return c.Send(ctx, Email{To: to, Subject: subject, Text: text})
}

// SendPasswordResetEmail sends a password reset link.
func (c *Client) SendPasswordResetEmail(ctx context.Context, to, token, locale string) error {
	subject, text := emailtpl.PasswordResetEmail(locale, token)
	return c.Send(ctx, Email{To: to, Subject: subject, Text: text})
}

// SendBondNotification sends a bond match notification to the link creator.
func (c *Client) SendBondNotification(ctx context.Context, to, otherName, creatorName, locale string) error {
	subject, text := emailtpl.BondNotification(locale, otherName, creatorName)
	return c.Send(ctx, Email{To: to, Subject: subject, Text: text})
}

// SendThankYouEmail sends a donation thank-you email.
func (c *Client) SendThankYouEmail(ctx context.Context, to, locale string) error {
	subject, text := emailtpl.ThankYouEmail(locale)
	return c.Send(ctx, Email{To: to, Subject: subject, Text: text})
}
