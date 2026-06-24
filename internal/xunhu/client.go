package xunhu

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"liki/internal/agent"
	"liki/internal/payment"
)

const (
	defaultBaseURL = "https://api.xunhupay.com/payment/do.html"
	requestTimeout = 30 * time.Second
)

// Client implements payment.Provider for XunhuPay (虎皮椒).
type Client struct {
	appID      string
	appSecret  string
	httpClient *http.Client
	baseURL    string
}

// New creates a XunhuPay client.
func New(appID, appSecret string) *Client {
	return &Client{
		appID:      appID,
		appSecret:  appSecret,
		httpClient: &http.Client{Timeout: requestTimeout},
		baseURL:    defaultBaseURL,
	}
}

// CreateCheckout creates a XunhuPay checkout session.
func (c *Client) CreateCheckout(ctx context.Context, product agent.Product, amount int, orderID, email, returnURL string) (*payment.CheckoutResult, error) {
	webhookURL := deriveWebhookURL(returnURL)

	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("xunhu: generate nonce: %w", err)
	}

	params := map[string]string{
		"appid":          c.appID,
		"time":           strconv.FormatInt(time.Now().Unix(), 10),
		"version":        "1.1",
		"nonce_str":      hex.EncodeToString(nonce),
		"trade_order_id": orderID,
		"total_fee":      fmt.Sprintf("%d.%02d", amount/100, amount%100),
		"title":          product.EmailSubject(),
		"notify_url":     webhookURL,
		"return_url":     returnURL,
	}
	params["hash"] = sign(params, c.appSecret)

	form := url.Values{}
	for k, v := range params {
		form.Set(k, v)
	}

	body := strings.NewReader(form.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, body)
	if err != nil {
		return nil, fmt.Errorf("xunhu: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("xunhu: request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("xunhu: read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("xunhu: http %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
		URL     string `json:"url"`
		QRCode  string `json:"url_qrcode"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("xunhu: parse response: %w", err)
	}
	if result.ErrCode != 0 {
		return nil, fmt.Errorf("xunhu: %s (errcode=%d)", result.ErrMsg, result.ErrCode)
	}

	return &payment.CheckoutResult{
		SessionID:   orderID,
		CheckoutURL: result.URL,
		QRCodeURL:   result.QRCode,
	}, nil
}

// VerifyWebhook verifies a XunhuPay webhook request.
func (c *Client) VerifyWebhook(rawBody []byte, headers http.Header) (*payment.WebhookEvent, error) {
	if len(rawBody) == 0 {
		return nil, fmt.Errorf("xunhu: empty body")
	}

	values, err := url.ParseQuery(string(rawBody))
	if err != nil {
		return nil, fmt.Errorf("xunhu: parse webhook: %w", err)
	}

	hash := values.Get("hash")
	if hash == "" {
		return nil, fmt.Errorf("xunhu: missing hash in webhook")
	}

	// Recalculate hash from all params except hash itself.
	params := make(map[string]string, len(values))
	for k := range values {
		if k == "hash" {
			continue
		}
		params[k] = values.Get(k)
	}
	expected := sign(params, c.appSecret)
	if hash != expected {
		return nil, fmt.Errorf("xunhu: signature mismatch")
	}

	amount, err := strconv.Atoi(values.Get("total_fee"))
	if err != nil {
		amount = 0
	}
	eventType := values.Get("trade_status")
	if eventType == "TRADE_SUCCESS" {
		eventType = "payment.succeeded"
	}

	orderID := values.Get("trade_order_id")
	if orderID == "" {
		orderID = values.Get("out_trade_no")
	}
	return &payment.WebhookEvent{
		Type: eventType,
		Data: payment.WebhookEventData{
			OrderID:   orderID,
			Amount:    amount,
			Email:     values.Get("openid"),
			PaymentID: values.Get("trade_no"),
		},
	}, nil
}

// sign generates an MD5 signature for XunhuPay parameters.
func sign(params map[string]string, secret string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var raw string
	for _, k := range keys {
		raw += k + "=" + params[k] + "&"
	}
	if len(raw) > 0 {
		raw = raw[:len(raw)-1] // remove trailing &
	}
	raw += secret

	h := md5.Sum([]byte(raw))
	return hex.EncodeToString(h[:])
}

// deriveWebhookURL extracts the base URL from returnURL and appends the webhook path.
func deriveWebhookURL(returnURL string) string {
	u, err := url.Parse(returnURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return returnURL
	}
	return u.Scheme + "://" + u.Host + "/api/webhook"
}
