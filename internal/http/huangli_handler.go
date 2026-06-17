package handler

import (
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"liki/internal/engine/huangli"
	"liki/internal/engine/tianwen"
)

func queryHuangli(w http.ResponseWriter, r *http.Request) {
	type params struct{ Event, Date, Month string }
	p := params{
		Event: r.URL.Query().Get("event"),
		Date:  r.URL.Query().Get("date"),
		Month: r.URL.Query().Get("month"),
	}
	if err := validation.ValidateStruct(&p,
		validation.Field(&p.Event, validation.Required),
		validation.Field(&p.Date, validation.When(p.Month == "", validation.Required)),
	); err != nil {
		respondValidationError(w, err)
		return
	}
	if p.Date != "" {
		entry, err := huangli.QueryDate(p.Date, p.Event)
		if err != nil {
			respondInvalidRequest(w, "invalid date format, use YYYY-MM-DD")
			return
		}
		respondJSON(w, http.StatusOK, entry)
		return
	}
	entries, err := huangli.QueryMonth(p.Month, p.Event)
	if err != nil {
		respondInvalidRequest(w, "invalid month format, use YYYY-MM")
		return
	}
		respondJSON(w, http.StatusOK, entries)
}

type huangliBondRequest struct {
	Birth     SolarTime `json:"birth_info"`
	Month     string      `json:"month"`
	Date      string      `json:"date"`
	EventType string      `json:"event_type"`
}

func (r huangliBondRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.EventType, validation.Required),
		validation.Field(&r.Month, validation.When(r.Date == "", validation.Required)),
	)
}

func bondHuangli(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeJSON[huangliBondRequest](w, r)
	if !ok { return }
	if err := validateSolarParams(&req.Birth); err != nil { respondValidationError(w, err); return }
	if err := req.Validate(); err != nil { respondValidationError(w, err); return }

	bt := tianwen.ComputeBirthTime(req.Birth.Year, req.Birth.Month, req.Birth.Day, req.Birth.Hour, req.Birth.Minute, req.Birth.Longitude, req.Birth.Timezone)
	bz := tianwen.ComputeBazi(bt.Solar)

	if req.Date != "" {
		entry, err := huangli.CrossDate(bz.Ri.Gan, bz.Ri.Zhi, req.Date, req.EventType)
		if err != nil { respondValidationError(w, err); return }
		respondJSON(w, http.StatusOK, entry)
		return
	}
	entries, err := huangli.CrossMonth(bz.Ri.Gan, bz.Ri.Zhi, req.Month, req.EventType)
	if err != nil { respondValidationError(w, err); return }
		respondJSON(w, http.StatusOK, entries)
}
