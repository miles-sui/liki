package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRespondJSON(t *testing.T) {
	w := httptest.NewRecorder()
	respondJSON(w, http.StatusOK, map[string]string{"key": "value"})

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	ct := w.Header().Get("Content-Type")
	if ct != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %s, want application/json; charset=utf-8", ct)
	}

	var env envelope
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if env.Error != nil {
		t.Errorf("error = %+v, want nil", env.Error)
	}
	if env.Data == nil {
		t.Error("data is nil")
	}
}

func TestRespondError(t *testing.T) {
	w := httptest.NewRecorder()
	respondError(w, http.StatusBadRequest, "bad_input", "something wrong")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var env envelope
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if env.Error.Code != "bad_input" {
		t.Errorf("code = %s, want bad_input", env.Error.Code)
	}
	if env.Error.Message != "something wrong" {
		t.Errorf("message = %s, want something wrong", env.Error.Message)
	}
}

func TestRespondStatus(t *testing.T) {
	w := httptest.NewRecorder()
	respondStatus(w, http.StatusOK, "ok")

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var env envelope
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	data, ok := env.Data.(map[string]interface{})
	if !ok {
		t.Fatal("data is not a map")
	}
	if data["status"] != "ok" {
		t.Errorf("status = %s, want ok", data["status"])
	}
}

func TestRespondError_StatusCodes(t *testing.T) {
	tests := []struct {
		status int
		code   string
	}{
		{http.StatusNotFound, "not_found"},
		{http.StatusInternalServerError, "internal_error"},
		{http.StatusUnauthorized, "unauthorized"},
		{http.StatusUnprocessableEntity, "validation_error"},
	}

	for _, tc := range tests {
		t.Run(tc.code, func(t *testing.T) {
			w := httptest.NewRecorder()
			respondError(w, tc.status, tc.code, "msg")
			if w.Code != tc.status {
				t.Errorf("status = %d, want %d", w.Code, tc.status)
			}
		})
	}
}
