package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"liki/internal/agent"
	"liki/internal/engine/ganzhi"
	"liki/internal/engine/tianwen"
)

// timePoint is an alias for agent.TimePoint, used by HTTP handler validation.
// The actual type and conversion logic live in the agent package.
type timePoint = agent.TimePoint

// BirthRequest is the common HTTP body for chart endpoints that accept
// birth time + gender. The handler converts timePoint to tianwen.Timeset,
// then passes SolarTime + Gender to the engine.
type BirthRequest struct {
	Birth  timePoint     `json:"birth"`
	Gender ganzhi.Gender `json:"gender"`
}

func (r BirthRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Birth, validation.By(validateTimePoint)),
		validation.Field(&r.Gender, validation.Required, validation.In(validGenders...)),
	)
}

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
	respondError(w, http.StatusUnprocessableEntity, "validation_error", err.Error())
}

func validateTimePoint(value any) error {
	tp, ok := value.(timePoint)
	if !ok {
		return errors.New("required")
	}
	if tp.Time == "" {
		return errors.New("time is required")
	}
	if _, err := time.Parse(time.RFC3339, tp.Time); err != nil {
		return errors.New("time must be RFC3339 format")
	}
	return nil
}

// timesetOrRespond converts timePoint to tianwen.Timeset, writing an error
// response and returning false on failure. Callers should return immediately.
func timesetOrRespond(w http.ResponseWriter, tp timePoint) (tianwen.Timeset, bool) {
	ts, err := tp.Timeset()
	if err != nil {
		respondInvalidRequest(w, err.Error())
		return tianwen.Timeset{}, false
	}
	return ts, true
}

func validateNonZeroGregorianTime(value any) error {
	if gt, ok := value.(tianwen.GregorianTime); ok && gt.Time().IsZero() {
		return errors.New("required")
	}
	return nil
}

