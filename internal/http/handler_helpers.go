package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"liki/internal/agent"
	"liki/internal/engine/ganzhi"
	"liki/internal/engine/tianwen"
)

// timePoint is an alias for agent.TimePoint, used by HTTP handler validation.
// The actual type and conversion logic live in the agent package.
type timePoint = agent.TimePoint



var validGenders = []any{ganzhi.Male, ganzhi.Female}

// detectCurrency infers currency from Cloudflare IP country header.
func detectCurrency(r *http.Request) string {
	if r.Header.Get("CF-IPCountry") == "CN" {
		return "CNY"
	}
	return "USD"
}

// decodeJSON decodes the JSON request body into T.
// On failure it writes a standard error response and returns false.
func decodeJSON[T any](w http.ResponseWriter, r *http.Request) (T, bool) {
	var req T
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			respondError(w, http.StatusRequestEntityTooLarge, "too_large", "Request body too large")
			return req, false
		}
		respondInvalidRequest(w, "Invalid JSON body")
		return req, false
	}
	return req, true
}

type validator interface {
	Validate() error
}

// decodeAndValidate decodes JSON body and calls req.Validate().
func decodeAndValidate[T validator](w http.ResponseWriter, r *http.Request) (T, bool) {
	req, ok := decodeJSON[T](w, r)
	if !ok {
		return req, false
	}
	if err := req.Validate(); err != nil {
		slog.Warn("validation failed", "err", err)
		respondError(w, http.StatusUnprocessableEntity, "invalid_request", "Invalid request")
		return req, false
	}
	return req, true
}

// decodeAndValidate decodes JSON body and calls the provided validate function.
// Use this when the request type is an anonymous struct that can't have methods.
func decodeWith[T any](w http.ResponseWriter, r *http.Request, validate func(T) error) (T, bool) {
	req, ok := decodeJSON[T](w, r)
	if !ok {
		return req, false
	}
	if err := validate(req); err != nil {
		slog.Warn("validation failed", "err", err)
		respondError(w, http.StatusUnprocessableEntity, "invalid_request", "Invalid request")
		return req, false
	}
	return req, true
}

func respondInvalidRequest(w http.ResponseWriter, msg string) {
	respondError(w, http.StatusBadRequest, "invalid_request", msg)
}

func respondValidationError(w http.ResponseWriter, err error) {
	slog.Warn("validation failed", "err", err)
	respondError(w, http.StatusUnprocessableEntity, "invalid_request", "Invalid request")
}

func validateTimePoint(value any) error {
	tp, ok := value.(timePoint)
	if !ok {
		return errors.New("required")
	}
	if tp.Lunar != nil {
		if tp.Lunar.Year < 1900 || tp.Lunar.Year > 2100 {
			return errors.New("lunar year must be 1900-2100")
		}
		if tp.Lunar.Month < 1 || tp.Lunar.Month > 12 {
			return errors.New("lunar month must be 1-12")
		}
		if tp.Lunar.Day < 1 || tp.Lunar.Day > 30 {
			return errors.New("lunar day must be 1-30")
		}
		if tp.Lunar.Hour < 0 || tp.Lunar.Hour > 23 {
			return errors.New("lunar hour must be 0-23")
		}
		return nil
	}
	if tp.Time == "" {
		return errors.New("time or lunar is required")
	}
	if _, err := time.Parse(time.RFC3339, tp.Time); err != nil {
		return errors.New("time must be RFC3339 format")
	}
	return nil
}

func validateNonZeroGregorianTime(value any) error {
	if gt, ok := value.(tianwen.GregorianTime); ok && gt.Time().IsZero() {
		return errors.New("required")
	}
	return nil
}

