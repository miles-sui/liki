package tencent

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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
	_ user.EmailSender       = (*Client)(nil)
	_ commerce.ThankYouSender = (*Client)(nil)
)

const (
	endpoint = "ses.tencentcloudapi.com"
	service  = "ses"
)

// Client sends transactional emails via Tencent Cloud SES.
type Client struct {
	secretID  string
	secretKey string
	from      string
	region    string
	http      *http.Client
}

// New creates a Tencent SES client.
func New(secretID, secretKey, from, region string) *Client {
	if region == "" {
		region = "ap-hongkong"
	}
	return &Client{
		secretID:  secretID,
		secretKey: secretKey,
		from:      from,
		region:    region,
		http:      &http.Client{Timeout: 10 * time.Second},
	}
}

type emailRequest struct {
	FromEmailAddress string         `json:"FromEmailAddress"`
	Destination      []string       `json:"Destination"`
	Subject          string         `json:"Subject"`
	ReplyToAddresses string         `json:"ReplyToAddresses,omitempty"`
	Simple           *simpleContent `json:"Simple,omitempty"`
}

type simpleContent struct {
	HTML string `json:"Html,omitempty"`
	Text string `json:"Text,omitempty"`
}

// Send delivers an email via Tencent SES.
func (c *Client) Send(ctx context.Context, to, subject, text string) error {
	body := emailRequest{
		FromEmailAddress: c.from,
		Destination:      []string{to},
		Subject:          subject,
		Simple:           &simpleContent{Text: text},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("tencent: marshal: %w", err)
	}

	reader := bytes.NewReader(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://"+endpoint, reader)
	if err != nil {
		return fmt.Errorf("tencent: new request: %w", err)
	}

	timestamp := time.Now().Unix()
	c.sign(req, payload, timestamp, "SendEmail")

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-TC-Action", "SendEmail")
	req.Header.Set("X-TC-Version", "2020-10-02")
	req.Header.Set("X-TC-Timestamp", fmt.Sprintf("%d", timestamp))
	req.Header.Set("X-TC-Region", c.region)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("tencent: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("tencent: %d: %s", resp.StatusCode, string(b))
	}

	var result struct {
		Response struct {
			Error *struct {
				Code    string `json:"Code"`
				Message string `json:"Message"`
			} `json:"Error"`
		} `json:"Response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("tencent: decode response: %w", err)
	}
	if result.Response.Error != nil {
		return fmt.Errorf("tencent: %s — %s", result.Response.Error.Code, result.Response.Error.Message)
	}
	return nil
}

// TC3-HMAC-SHA256 signing.
func (c *Client) sign(req *http.Request, payload []byte, timestamp int64, action string) {
	date := time.Unix(timestamp, 0).UTC().Format("2006-01-02")

	// Canonical headers
	ct := req.Header.Get("Content-Type")
	canonicalHeaders := fmt.Sprintf("content-type:%s\nhost:%s\n", ct, endpoint)
	signedHeaders := "content-type;host"

	hashedPayload := sha256Hex(payload)

	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		"POST", "/", "", canonicalHeaders, signedHeaders, hashedPayload)

	credentialScope := fmt.Sprintf("%s/%s/tc3_request", date, service)
	stringToSign := fmt.Sprintf("%s\n%d\n%s\n%s",
		"TC3-HMAC-SHA256", timestamp, credentialScope, sha256Hex([]byte(canonicalRequest)))

	secretDate := hmacSHA256([]byte("TC3"+c.secretKey), []byte(date))
	secretService := hmacSHA256(secretDate, []byte(service))
	secretSigning := hmacSHA256(secretService, []byte("tc3_request"))
	signature := hex.EncodeToString(hmacSHA256(secretSigning, []byte(stringToSign)))

	authorization := fmt.Sprintf(
		"TC3-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		c.secretID, credentialScope, signedHeaders, signature)
	req.Header.Set("Authorization", authorization)
}

func sha256Hex(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func hmacSHA256(key, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return mac.Sum(nil)
}

// SendVerificationEmail sends an email verification link.
func (c *Client) SendVerificationEmail(ctx context.Context, to, token, locale string) error {
	subject, text := emailtpl.VerificationEmail(locale, token)
	return c.Send(ctx, to, subject, text)
}

// SendPasswordResetEmail sends a password reset link.
func (c *Client) SendPasswordResetEmail(ctx context.Context, to, token, locale string) error {
	subject, text := emailtpl.PasswordResetEmail(locale, token)
	return c.Send(ctx, to, subject, text)
}

// SendBondNotification sends a bond match notification to the link creator.
func (c *Client) SendBondNotification(ctx context.Context, to, otherName, creatorName, locale string) error {
	subject, text := emailtpl.BondNotification(locale, otherName, creatorName)
	return c.Send(ctx, to, subject, text)
}

// SendThankYouEmail sends a donation thank-you email.
func (c *Client) SendThankYouEmail(ctx context.Context, to, locale string) error {
	subject, text := emailtpl.ThankYouEmail(locale)
	return c.Send(ctx, to, subject, text)
}
