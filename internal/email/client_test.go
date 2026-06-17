package email

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	resend "github.com/resend/resend-go/v3"
)

func TestSendReport_InvalidAPIKey(t *testing.T) {
	c := New("bad-key", "from@test.com")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := c.SendReport(ctx, "to@example.com", "Test", "<p>Hello</p>")
	if err == nil {
		t.Fatal("expected error with bad API key")
	}
	t.Logf("expected error: %v", err)
}

// TestSendReport_Integration sends a real email via Resend.
// Run with: RESEND_API_KEY=... EMAIL_FROM=... go test -tags integration -run TestSendReport_Integration ./internal/email/
func TestSendReport_Integration(t *testing.T) {
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		t.Skip("RESEND_API_KEY not set")
	}
	from := os.Getenv("EMAIL_FROM")
	if from == "" {
		from = "noreply@notify.liki.hk"
	}

	to := "suiqiang@foxmail.com"
	subject := "[灵机测试] 邮件发送集成测试"
	htmlBody := `<html><body>
	<h1>灵机邮件测试</h1>
	<p>这是一封来自灵机系统的测试邮件。</p>
	<p>如果你收到这封邮件，说明 Resend 邮件服务配置正确。</p>
	<hr>
	<p><small>发送时间: ` + time.Now().UTC().Format(time.RFC3339) + `</small></p>
	</body></html>`

	c := New(apiKey, from)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := c.SendReport(ctx, to, subject, htmlBody); err != nil {
		t.Fatalf("SendReport failed: %v", err)
	}
	t.Logf("email sent successfully to %s", to)
}

func newTestClientWithBaseURL(t *testing.T, apiKey, from, baseURL string) *Client {
	t.Helper()
	tc := resend.NewClient(apiKey)
	var err error
	tc.BaseURL, err = url.Parse(baseURL)
	if err != nil {
		t.Fatalf("url.Parse: %v", err)
	}
	return &Client{client: tc, from: from}
}

func TestSendReport_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/emails" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"email_test123"}`)) //nolint:errcheck
	}))
	defer srv.Close()

	c := newTestClientWithBaseURL(t, "re_test", "from@test.com", srv.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := c.SendReport(ctx, "to@test.com", "Test Subject", "<p>Hello</p>")
	if err != nil {
		t.Fatalf("SendReport: %v", err)
	}
}

func TestSendReport_RateLimitRetry(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("retry-after", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"message":"rate limited"}`)) //nolint:errcheck
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"email_retry_ok"}`)) //nolint:errcheck
	}))
	defer srv.Close()

	c := newTestClientWithBaseURL(t, "re_test", "from@test.com", srv.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := c.SendReport(ctx, "to@test.com", "Test", "<p>Hello</p>")
	if err != nil {
		t.Fatalf("SendReport should succeed after rate limit retry: %v", err)
	}
	if attempts != 2 {
		t.Errorf("attempts = %d, want 2", attempts)
	}
}

func TestSendReport_AllRetriesFail(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"internal error"}`)) //nolint:errcheck
	}))
	defer srv.Close()

	c := newTestClientWithBaseURL(t, "re_test", "from@test.com", srv.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := c.SendReport(ctx, "to@test.com", "Test", "<p>Hello</p>")
	if err == nil {
		t.Fatal("expected error after all retries fail")
	}
}
