package handler

import (
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"liki/internal/engine/huangli"
)


func huangliDate(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	event := r.URL.Query().Get("event")
	if date == "" || event == "" {
		respondInvalidRequest(w, "date and event query params are required")
		return
	}
	entry, err := huangli.QueryDate(date, event)
	if err != nil {
		respondInvalidRequest(w, "invalid date format, use YYYY-MM-DD")
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"entry": entry})
}
func huangliMonth(w http.ResponseWriter, r *http.Request) {
	month := r.URL.Query().Get("month")
	event := r.URL.Query().Get("event")
	if month == "" || event == "" {
		respondInvalidRequest(w, "month and event query params are required")
		return
	}
	entries, err := huangli.QueryMonth(month, event)
	if err != nil {
		respondInvalidRequest(w, "invalid month format, use YYYY-MM")
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"entries": entries})
}

type huangliBondDateRequest struct {
	Birth     timePoint `json:"birth"`
	EventType string     `json:"event_type"`
	Date      string     `json:"date"`
}

func (r huangliBondDateRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Birth, validation.By(validateTimePoint)),
		validation.Field(&r.EventType, validation.Required),
		validation.Field(&r.Date, validation.Required),
	)
}

func huangliBondDate(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeAndValidate[huangliBondDateRequest](w, r)
	if !ok {
		return
	}
	ts, err := req.Birth.Timeset()
	if err != nil {
		respondInvalidRequest(w, err.Error())
		return
	}
	entry, err := huangli.ComputeBondDay(ts.Solar, req.EventType, req.Date)
	if err != nil {
		respondValidationError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"entry": entry})
}

type huangliBondMonthRequest struct {
	Birth     timePoint `json:"birth"`
	EventType string     `json:"event_type"`
	Month     string     `json:"month"`
}

func (r huangliBondMonthRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Birth, validation.By(validateTimePoint)),
		validation.Field(&r.EventType, validation.Required),
		validation.Field(&r.Month, validation.Required),
	)
}

func huangliBondMonth(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeAndValidate[huangliBondMonthRequest](w, r)
	if !ok {
		return
	}
	ts, err := req.Birth.Timeset()
	if err != nil {
		respondInvalidRequest(w, err.Error())
		return
	}
	entries, err := huangli.ComputeBondMonth(ts.Solar, req.EventType, req.Month)
	if err != nil {
		respondValidationError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"entries": entries})
}
