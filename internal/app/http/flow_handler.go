package http

import (
	"net/http"
	"time"

	"github.com/25types/25types/internal/app/application/flow"
)

// FlowHandler holds dependencies for flow HTTP handlers.
type FlowHandler struct {
	Profiles flow.ProfileLoader
}

type flowMonthResponse struct {
	MonthID   string `json:"month_id"`
	MonthEN   string `json:"month_en"`
	Generates int    `json:"generates"`
	Restrains int    `json:"restrains"`
}

type flowYearlyResponse struct {
	Months  []flowMonthResponse `json:"months"`
	Current string              `json:"current"`
}

type solarTermMonthOut struct {
	ID     string `json:"id"`
	NameEN string `json:"name_en"`
	Start  string `json:"start"`
	End    string `json:"end"`
}

type solarTermsResponse struct {
	Year    int                 `json:"year"`
	Current solarTermMonthOut   `json:"current"`
	Months  []solarTermMonthOut `json:"months"`
}

// GET /api/flow
func (h *FlowHandler) GetFlow(w http.ResponseWriter, r *http.Request) {
	uid, _ := UserID(r.Context())

	result, err := flow.GetFlow(r.Context(), h.Profiles, uid)
	if err != nil {
		respondError(w, http.StatusNotFound, "not_found", "No profile found. Submit an assessment first.")
		return
	}

	respondJSON(w, http.StatusOK, flowMonthResponse{
		MonthID: result.MonthID, MonthEN: result.MonthEN,
		Generates: result.Generates, Restrains: result.Restrains,
	})
}

// GET /api/flow/yearly
func (h *FlowHandler) GetFlowYearly(w http.ResponseWriter, r *http.Request) {
	uid, _ := UserID(r.Context())

	months, current, err := flow.GetFlowYearly(r.Context(), h.Profiles, uid)
	if err != nil {
		respondError(w, http.StatusNotFound, "not_found", "No profile found. Submit an assessment first.")
		return
	}

	outMonths := make([]flowMonthResponse, 0, len(months))
	for _, m := range months {
		outMonths = append(outMonths, flowMonthResponse{MonthID: m.MonthID, MonthEN: m.MonthEN, Generates: m.Generates, Restrains: m.Restrains})
	}

	respondJSON(w, http.StatusOK, flowYearlyResponse{
		Months:  outMonths,
		Current: current.MonthID,
	})
}

// GET /api/solar-terms
func (h *FlowHandler) GetSolarTerms(w http.ResponseWriter, r *http.Request) {
	entries, currentID := flow.GetSolarTerms()

	months := make([]solarTermMonthOut, 0, 12)
	var current solarTermMonthOut
	for i, e := range entries {
		nextIdx := (i + 1) % 12
		sm := solarTermMonthOut{
			ID:     e.MonthID,
			NameEN: e.NameEN,
			Start:  e.Date.Format("2006-01-02"),
			End:    entries[nextIdx].Date.Format("2006-01-02"),
		}
		if e.MonthID == currentID {
			current = sm
		}
		months = append(months, sm)
	}

	respondJSON(w, http.StatusOK, solarTermsResponse{
		Year:    time.Now().Year(),
		Current: current,
		Months:  months,
	})
}
