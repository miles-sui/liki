package email

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	resend "github.com/resend/resend-go/v3"
)

// Client sends emails via the Resend API with retry support.
type Client struct {
	client *resend.Client
	from   string
}

// New creates a Resend email client with the given API key and from address.
func New(apiKey, from string) *Client {
	return &Client{
		client: resend.NewClient(apiKey),
		from:   from,
	}
}

// SendReport sends an HTML email with up to 3 retries on rate limit or transient errors.
func (c *Client) SendReport(ctx context.Context, to, subject, htmlBody string) error {
	params := &resend.SendEmailRequest{
		From:    c.from,
		To:      []string{to},
		Subject: subject,
		Html:    htmlBody,
	}

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		_, err := c.client.Emails.SendWithContext(ctx, params)
		cancel()
		if err == nil {
			return nil
		}
		lastErr = err

		var rateLimitErr *resend.RateLimitError
		if errors.As(err, &rateLimitErr) {
			if retryAfter, e := strconv.Atoi(rateLimitErr.RetryAfter); e == nil {
				slog.Warn("email: rate limited, waiting", "seconds", retryAfter)
				time.Sleep(time.Duration(retryAfter) * time.Second)
				continue
			}
		}

		if attempt < 2 {
			slog.Warn("email: send attempt failed", "attempt", attempt+1, "err", err)
		}
	}
	return fmt.Errorf("email: %w", lastErr)
}
