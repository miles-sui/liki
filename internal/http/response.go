package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// envelope is the unified response wrapper.
type envelope struct {
	Data  any       `json:"data,omitempty"`
	Error *apiError `json:"error,omitempty"`
}

// apiError is a structured error.
type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(envelope{Data: data}); err != nil {
		slog.Warn("respondJSON encode", "err", err)
	}
}

func respondError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(envelope{Error: &apiError{Code: code, Message: message}}); err != nil {
		slog.Warn("respondError encode", "err", err)
	}
}

func respondStatus(w http.ResponseWriter, status int, s string) {
	respondJSON(w, status, map[string]string{"status": s})
}
