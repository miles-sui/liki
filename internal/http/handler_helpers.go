package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"liki/internal/product"
)

// detectCurrency infers currency from Cloudflare IP country header.
// Defaults to CNY when neither CF header nor geo API identifies the country.
func detectCurrency(r *http.Request, geoCountry string) product.Currency {
	if r.Header.Get("CF-IPCountry") == "CN" || geoCountry == "CN" {
		return product.CNY
	}
	if r.Header.Get("CF-IPCountry") != "" || geoCountry != "" {
		return product.USD
	}
	return product.CNY
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
		respondValidationError(w, err)
		return req, false
	}
	return req, true
}

func respondInvalidRequest(w http.ResponseWriter, msg string) {
	respondError(w, http.StatusBadRequest, "invalid_request", msg)
}

func respondValidationError(w http.ResponseWriter, err error) {
	slog.Warn("validation failed", "err", err)
	respondError(w, http.StatusUnprocessableEntity, "validation_error", "Invalid request parameters")
}


