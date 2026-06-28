package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDetectCurrency_CN(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("CF-IPCountry", "CN")
	if got := detectCurrency(r, ""); got != "CNY" {
		t.Errorf("detectCurrency(CF=CN, geo=) = %s, want CNY", got)
	}
}

func TestDetectCurrency_GeoCountryCN(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	if got := detectCurrency(r, "CN"); got != "CNY" {
		t.Errorf("detectCurrency(CF=, geo=CN) = %s, want CNY", got)
	}
}

func TestDetectCurrency_NonCN(t *testing.T) {
	tests := []struct {
		country, want string
	}{
		{"US", "USD"},
		{"JP", "USD"},
		{"GB", "USD"},
		{"", "CNY"},
	}
	for _, tc := range tests {
		t.Run("country="+tc.country, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			if tc.country != "" {
				r.Header.Set("CF-IPCountry", tc.country)
			}
			if got := detectCurrency(r, ""); string(got) != tc.want {
				t.Errorf("detectCurrency(%s) = %s, want %s", tc.country, got, tc.want)
			}
		})
	}
}

func TestCORSMiddleware_Headers(t *testing.T) {
	handler := CORSMiddleware(false, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Origin", "https://liki.hk")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://liki.hk" {
		t.Errorf("Allow-Origin = %s, want https://liki.hk", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Methods"); got != "GET, POST, OPTIONS" {
		t.Errorf("Allow-Methods = %s, want GET, POST, OPTIONS", got)
	}
}

func TestCORSMiddleware_NoOrigin(t *testing.T) {
	handler := CORSMiddleware(false, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://liki.hk" {
		t.Errorf("Allow-Origin without Origin header = %s, want https://liki.hk", got)
	}
}

func TestCORSMiddleware_UnknownOrigin(t *testing.T) {
	handler := CORSMiddleware(false, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Origin", "https://evil.com")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://liki.hk" {
		t.Errorf("Allow-Origin for unknown origin = %s, want https://liki.hk", got)
	}
}

func TestCORSMiddleware_OPTIONS(t *testing.T) {
	handler := CORSMiddleware(false, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for OPTIONS")
	}))

	r := httptest.NewRequest("OPTIONS", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNoContent)
	}
}

func TestBodyLimit_WithinLimit(t *testing.T) {
	handler := BodyLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	body := strings.NewReader(`{"key":"value"}`)
	r := httptest.NewRequest("POST", "/", body)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestBodyLimit_ExceedsLimit(t *testing.T) {
	readErr := make(chan error, 1)
	handler := BodyLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := r.Body.Read(make([]byte, 3<<20))
		readErr <- err
	}))

	body := strings.NewReader(strings.Repeat("x", 3<<20)) // 3 MB
	r := httptest.NewRequest("POST", "/", body)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if err := <-readErr; err == nil {
		t.Error("MaxBytesReader should reject oversized body, but read succeeded")
	}
}

func TestSecurityHeaders(t *testing.T) {
	handler := SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if got := w.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("X-Content-Type-Options = %q, want nosniff", got)
	}
	if got := w.Header().Get("X-Frame-Options"); got != "DENY" {
		t.Errorf("X-Frame-Options = %q, want DENY", got)
	}
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestCORSMiddleware_DevMode(t *testing.T) {
	handler := CORSMiddleware(true, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Origin", "http://localhost:8080")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:8080" {
		t.Errorf("Allow-Origin = %q, want http://localhost:8080", got)
	}
}
