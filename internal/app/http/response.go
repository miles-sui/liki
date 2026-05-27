package http

import (
	"net/http"

	"github.com/25types/25types/internal/httputil"
)

// Re-exported types for backward compatibility.
type Envelope = httputil.Envelope
type APIError = httputil.APIError

// Function wrappers.
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	httputil.RespondJSON(w, status, data)
}
func respondList(w http.ResponseWriter, items interface{}, total int) {
	httputil.RespondList(w, items, total)
}
func respondError(w http.ResponseWriter, status int, code, message string) {
	httputil.RespondError(w, status, code, message)
}
func respondStatus(w http.ResponseWriter, status int, s string) {
	httputil.RespondStatus(w, status, s)
}
