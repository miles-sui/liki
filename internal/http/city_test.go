package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleCity_MissingName(t *testing.T) {
	r := httptest.NewRequest("GET", "/api/city", nil)
	w := httptest.NewRecorder()
	handleCity(w, r)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", w.Code)
	}

	var env envelope
	json.NewDecoder(w.Body).Decode(&env)
	if env.Error == nil || env.Error.Message == "" {
		t.Error("error response missing message")
	}
}

func TestHandleCity_ValidRequest(t *testing.T) {
	// This test hits the real Nominatim API — skip in CI.
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	r := httptest.NewRequest("GET", "/api/city?name=Beijing", nil)
	w := httptest.NewRecorder()
	handleCity(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var env envelope
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}
	var result struct {
		Name      string  `json:"name"`
		Longitude float64 `json:"longitude"`
		Latitude  float64 `json:"latitude"`
	}
	b, _ := json.Marshal(env.Data)
	json.Unmarshal(b, &result)

	if result.Name == "" {
		t.Error("name is empty")
	}
	if result.Longitude == 0 || result.Latitude == 0 {
		t.Error("coordinates are zero")
	}
}

func TestHandleCity_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	r := httptest.NewRequest("GET", "/api/city?name=xyznotacity999", nil)
	w := httptest.NewRecorder()
	handleCity(w, r)

	if w.Code < 400 {
		t.Errorf("status = %d, want error", w.Code)
	}
}
