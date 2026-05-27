package httputil

import (
	"encoding/json"
	"net/http"
)

// Envelope is the unified response wrapper.
type Envelope struct {
	Data  interface{} `json:"data,omitempty"`
	Error *APIError   `json:"error,omitempty"`
}

// APIError is a structured error.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// RespondJSON writes a single-object success response.
func RespondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(Envelope{Data: data})
}

// RespondList writes a paginated list response.
func RespondList(w http.ResponseWriter, items interface{}, total int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Envelope{Data: map[string]interface{}{
		"items": items,
		"total": total,
	}})
}

// RespondError writes a structured error response.
func RespondError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(Envelope{Error: &APIError{Code: code, Message: message}})
}

// RespondStatus writes a simple status map response.
func RespondStatus(w http.ResponseWriter, status int, s string) {
	RespondJSON(w, status, map[string]string{"status": s})
}
