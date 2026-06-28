package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleHealth(t *testing.T) {
	h := handleHealth()
	r := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"ok"`) {
		t.Errorf("body should contain ok: %s", w.Body.String())
	}
}

func TestHandleVersion(t *testing.T) {
	h := handleVersion("2026-06-14T10:00:00Z")
	r := httptest.NewRequest("GET", "/version", nil)
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "2026-06-14") {
		t.Errorf("body should contain build time: %s", w.Body.String())
	}
}

func TestClientIP_XForwardedFor(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("X-Forwarded-For", "1.2.3.4, 5.6.7.8")
	if got := clientIP(r); got != "1.2.3.4" {
		t.Errorf("clientIP = %q, want 1.2.3.4", got)
	}
}

func TestClientIP_RemoteAddr(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "192.168.1.1:12345"
	if got := clientIP(r); got != "192.168.1.1" {
		t.Errorf("clientIP = %q, want 192.168.1.1", got)
	}
}

func TestClientIP_RemoteAddrNoPort(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "10.0.0.1"
	if got := clientIP(r); got != "10.0.0.1" {
		t.Errorf("clientIP = %q, want 10.0.0.1", got)
	}
}

func TestIsPrivateIP_Loopback(t *testing.T) {
	if !isPrivateIP("127.0.0.1") {
		t.Error("127.0.0.1 should be private")
	}
	if !isPrivateIP("::1") {
		t.Error("::1 should be private")
	}
}

func TestIsPrivateIP_Private(t *testing.T) {
	if !isPrivateIP("192.168.1.1") {
		t.Error("192.168.1.1 should be private")
	}
	if !isPrivateIP("10.0.0.1") {
		t.Error("10.0.0.1 should be private")
	}
	if !isPrivateIP("172.16.0.1") {
		t.Error("172.16.0.1 should be private")
	}
}

func TestIsPrivateIP_Public(t *testing.T) {
	if isPrivateIP("8.8.8.8") {
		t.Error("8.8.8.8 should not be private")
	}
}

func TestIsPrivateIP_Invalid(t *testing.T) {
	if !isPrivateIP("not-an-ip") {
		t.Error("invalid IP should be treated as private")
	}
}

func TestIsPrivateIP_Unspecified(t *testing.T) {
	if !isPrivateIP("0.0.0.0") {
		t.Error("0.0.0.0 should be private")
	}
}

func TestEd3_Health_AllMethods(t *testing.T) {
	h := handleHealth()
	for _, method := range []string{"GET", "POST", "PUT", "DELETE", "PATCH"} {
		t.Run(method, func(t *testing.T) {
			r := httptest.NewRequest(method, "/api/health", nil)
			w := httptest.NewRecorder()
			h(w, r)
			if w.Code >= 500 {
				t.Errorf("%s health: status=%d", method, w.Code)
			}
		})
	}
}

func TestEd3_Health_IgnoresQueryParams(t *testing.T) {
	h := handleHealth()
	r := httptest.NewRequest("GET", "/api/health?foo=bar&baz=qux", nil)
	w := httptest.NewRecorder()
	h(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("status=%d, want 200", w.Code)
	}
}
