package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleLocation_CFCountryCN(t *testing.T) {
	r := httptest.NewRequest("GET", "/api/location", nil)
	r.Header.Set("CF-IPCountry", "CN")
	r.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()
	handleLocation(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var env struct {
		Data locationResult `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if env.Data.Country != "CN" {
		t.Errorf("country = %q, want CN", env.Data.Country)
	}
	if env.Data.Currency != "CNY" {
		t.Errorf("currency = %q, want CNY", env.Data.Currency)
	}
	if env.Data.City != "" {
		t.Errorf("city = %q, want empty (private IP skips geo API)", env.Data.City)
	}
}

func TestHandleLocation_CFCountryUS(t *testing.T) {
	r := httptest.NewRequest("GET", "/api/location", nil)
	r.Header.Set("CF-IPCountry", "US")
	r.RemoteAddr = "10.0.0.1:12345"
	w := httptest.NewRecorder()
	handleLocation(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var env struct {
		Data locationResult `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if env.Data.Country != "US" {
		t.Errorf("country = %q, want US", env.Data.Country)
	}
	if env.Data.Currency != "USD" {
		t.Errorf("currency = %q, want USD (non-CN defaults to USD)", env.Data.Currency)
	}
}

func TestHandleLocation_NoCFHeader_PrivateIP(t *testing.T) {
	r := httptest.NewRequest("GET", "/api/location", nil)
	r.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()
	handleLocation(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var env struct {
		Data locationResult `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if env.Data.Country != "unknown" {
		t.Errorf("country = %q, want unknown", env.Data.Country)
	}
	if env.Data.Currency != "USD" {
		t.Errorf("currency = %q, want USD (default)", env.Data.Currency)
	}
}

func TestHandleLocation_InvalidIP(t *testing.T) {
	r := httptest.NewRequest("GET", "/api/location", nil)
	r.RemoteAddr = "not-an-ip"
	w := httptest.NewRecorder()
	handleLocation(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var env struct {
		Data locationResult `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if env.Data.Country != "unknown" {
		t.Errorf("country = %q, want unknown", env.Data.Country)
	}
}

// TestHandleLocation_GeoAPI tests the ip-api.com integration path.
// Must mock the HTTP transport to avoid real external calls.
func TestHandleLocation_GeoAPI(t *testing.T) {
	orig := locHTTPClient
	locHTTPClient = &http.Client{
		Transport: &mockLocationTransport{
			body: `{"countryCode":"JP","city":"Tokyo"}`,
		},
	}
	defer func() { locHTTPClient = orig }()

	r := httptest.NewRequest("GET", "/api/location", nil)
	// Use public IP to trigger geo API lookup.
	r.RemoteAddr = "8.8.8.8:12345"
	w := httptest.NewRecorder()
	handleLocation(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var env struct {
		Data locationResult `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if env.Data.Country != "JP" {
		t.Errorf("country = %q, want JP (from geo API fallback)", env.Data.Country)
	}
	if env.Data.City != "Tokyo" {
		t.Errorf("city = %q, want Tokyo", env.Data.City)
	}
}

// TestHandleLocation_CFOverridesGeoAPI: CF-IPCountry header takes priority over geo API.
func TestHandleLocation_CFOverridesGeoAPI(t *testing.T) {
	orig := locHTTPClient
	locHTTPClient = &http.Client{
		Transport: &mockLocationTransport{
			body: `{"countryCode":"JP","city":"Tokyo"}`,
		},
	}
	defer func() { locHTTPClient = orig }()

	r := httptest.NewRequest("GET", "/api/location", nil)
	r.Header.Set("CF-IPCountry", "CN")
	r.RemoteAddr = "8.8.8.8:12345"
	w := httptest.NewRecorder()
	handleLocation(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var env struct {
		Data locationResult `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if env.Data.Country != "CN" {
		t.Errorf("country = %q, want CN (CF header overrides geo API)", env.Data.Country)
	}
	// City still comes from geo API even when CF header is set.
	if env.Data.City != "Tokyo" {
		t.Errorf("city = %q, want Tokyo (geo API city used alongside CF country)", env.Data.City)
	}
}

// TestHandleLocation_GeoAPIError: geo API failure should not break the handler.
func TestHandleLocation_GeoAPIError(t *testing.T) {
	orig := locHTTPClient
	locHTTPClient = &http.Client{
		Transport: &mockLocationTransport{
			status: http.StatusInternalServerError,
			body:   "error",
		},
	}
	defer func() { locHTTPClient = orig }()

	r := httptest.NewRequest("GET", "/api/location", nil)
	r.RemoteAddr = "8.8.8.8:12345"
	w := httptest.NewRecorder()
	handleLocation(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (should not fail on geo API error)", w.Code)
	}
	var env struct {
		Data locationResult `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if env.Data.Country != "unknown" {
		t.Errorf("country = %q, want unknown (fallback when no CF and geo API fails)", env.Data.Country)
	}
}

type mockLocationTransport struct {
	status int
	body   string
}

func (m *mockLocationTransport) RoundTrip(_ *http.Request) (*http.Response, error) {
	if m.status == 0 {
		m.status = http.StatusOK
	}
	return &http.Response{
		StatusCode: m.status,
		Body:       io.NopCloser(strings.NewReader(m.body)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}, nil
}
