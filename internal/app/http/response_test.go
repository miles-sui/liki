package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// =============================================================================
// respondJSON — single object envelope
// =============================================================================

func TestRespondJSON_SingleObject(t *testing.T) {
	rec := httptest.NewRecorder()
	respondJSON(rec, http.StatusOK, map[string]string{"name": "alice"})

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	ct := rec.Header().Get("Content-Type")
	if ct != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q, want application/json; charset=utf-8", ct)
	}

	var env Envelope
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if env.Data == nil {
		t.Fatal("expected data field")
	}
	if env.Error != nil {
		t.Error("expected no error field")
	}
}

func TestRespondJSON_Created(t *testing.T) {
	rec := httptest.NewRecorder()
	respondJSON(rec, http.StatusCreated, map[string]int{"id": 1})

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", rec.Code)
	}
}

// =============================================================================
// respondList — list envelope
// =============================================================================

func TestRespondList(t *testing.T) {
	rec := httptest.NewRecorder()
	respondList(rec, []string{"a", "b"}, 2)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}

	var env Envelope
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	data, ok := env.Data.(map[string]interface{})
	if !ok {
		t.Fatal("expected data to be an object")
	}
	items, ok := data["items"].([]interface{})
	if !ok {
		t.Fatal("expected items array")
	}
	if len(items) != 2 {
		t.Errorf("items len = %d, want 2", len(items))
	}
	total, ok := data["total"].(float64)
	if !ok || int(total) != 2 {
		t.Errorf("total = %v, want 2", data["total"])
	}
}

func TestRespondList_Empty(t *testing.T) {
	rec := httptest.NewRecorder()
	respondList(rec, []string{}, 0)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	var env Envelope
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	data, _ := env.Data.(map[string]interface{})
	items, _ := data["items"].([]interface{})
	if len(items) != 0 {
		t.Errorf("expected empty items, got %d", len(items))
	}
}

// =============================================================================
// respondError — error envelope
// =============================================================================

func TestRespondError(t *testing.T) {
	rec := httptest.NewRecorder()
	respondError(rec, http.StatusNotFound, "not_found", "User not found")

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}

	var env Envelope
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if env.Data != nil {
		t.Error("expected no data field on error")
	}
	if env.Error == nil {
		t.Fatal("expected error field")
	}
	if env.Error.Code != "not_found" {
		t.Errorf("error.code = %q, want not_found", env.Error.Code)
	}
	if env.Error.Message != "User not found" {
		t.Errorf("error.message = %q, want User not found", env.Error.Message)
	}
}

func TestRespondError_AllStandardCodes(t *testing.T) {
	// Verify all standard error codes from docs/appendix/errors.md
	tests := []struct {
		status  int
		code    string
		message string
	}{
		{http.StatusBadRequest, "invalid_request", "answers is required"},
		{http.StatusUnauthorized, "unauthorized", "Authentication required"},
		{http.StatusUnauthorized, "token_expired", "Token expired"},
		{http.StatusForbidden, "forbidden", "This user's profile is private"},
		{http.StatusNotFound, "not_found", "User not found"},
		{http.StatusConflict, "conflict", "Username already exists"},
		{http.StatusTooManyRequests, "rate_limited", "Rate limit exceeded"},
		{http.StatusInternalServerError, "internal", "An unexpected error occurred"},
	}
	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			rec := httptest.NewRecorder()
			respondError(rec, tt.status, tt.code, tt.message)

			if rec.Code != tt.status {
				t.Errorf("status = %d, want %d", rec.Code, tt.status)
			}
			ct := rec.Header().Get("Content-Type")
			if ct != "application/json; charset=utf-8" {
				t.Errorf("Content-Type = %q", ct)
			}
			var env Envelope
			json.NewDecoder(rec.Body).Decode(&env)
			if env.Error.Code != tt.code {
				t.Errorf("code = %q, want %q", env.Error.Code, tt.code)
			}
		})
	}
}

// =============================================================================
// respondStatus — { "data": { "status": "..." } }
// =============================================================================

func TestRespondStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	respondStatus(rec, http.StatusOK, "logged_out")

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	var env Envelope
	json.NewDecoder(rec.Body).Decode(&env)
	data, _ := env.Data.(map[string]interface{})
	if data["status"] != "logged_out" {
		t.Errorf("status = %q, want logged_out", data["status"])
	}
}

// =============================================================================
// Envelope: no 500 message leaks internals
// =============================================================================

func TestRespondError_NoInternalDetails(t *testing.T) {
	// Per docs: 500 message must not contain stack traces, SQL, or file paths.
	rec := httptest.NewRecorder()
	respondError(rec, http.StatusInternalServerError, "internal", "An unexpected error occurred")

	var env Envelope
	json.NewDecoder(rec.Body).Decode(&env)
	msg := env.Error.Message
	if msg == "" {
		t.Error("500 error should have a message")
	}
	// Generic message — should NOT contain file paths, SQL keywords, or stack traces.
	for _, leak := range []string{".go:", "SELECT ", "sqlite", "panic:", "goroutine"} {
		if strings.Contains(msg, leak) {
			t.Errorf("500 message leaks internals: %q contains %q", msg, leak)
		}
	}
}

