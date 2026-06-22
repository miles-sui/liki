package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleOpenAPI_ReturnsJSON(t *testing.T) {
	r := httptest.NewRequest("GET", "/openapi.json", nil)
	w := httptest.NewRecorder()
	handleOpenAPI()(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var parsed map[string]any
	if err := json.NewDecoder(w.Body).Decode(&parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Verify it's a valid OpenAPI doc
	if parsed["openapi"] == nil {
		t.Error("missing openapi version")
	}
	if parsed["info"] == nil {
		t.Error("missing info section")
	}
	if parsed["paths"] == nil {
		t.Error("missing paths section")
	}

	// Verify key endpoints exist
	paths := parsed["paths"].(map[string]any)
	keys := []string{"/api/bazi/chart", "/api/bazi/liunian", "/api/ziwei/chart", "/api/agent/chat"}
	for _, k := range keys {
		if paths[k] == nil {
			t.Errorf("missing path: %s", k)
		}
	}
}

func TestHandleOpenAPI_CORSHeader(t *testing.T) {
	r := httptest.NewRequest("GET", "/openapi.json", nil)
	w := httptest.NewRecorder()
	handleOpenAPI()(w, r)

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("missing CORS header")
	}
}

func TestHandleOpenAPI_Idempotent(t *testing.T) {
	// Multiple calls return same content.
	r1 := httptest.NewRequest("GET", "/openapi.json", nil)
	w1 := httptest.NewRecorder()
	handleOpenAPI()(w1, r1)

	r2 := httptest.NewRequest("GET", "/openapi.json", nil)
	w2 := httptest.NewRecorder()
	handleOpenAPI()(w2, r2)

	if w1.Body.String() != w2.Body.String() {
		t.Error("non-idempotent responses")
	}
}
