// Daily handler provides user-facing daily suggestions and questions.
// This is the "application" layer over the huangli domain — it uses the
// user's birth info to produce personalized daily content. The underlying
// calendar/day-pillar computations live in mingli/huangli_*.go and are
// exposed via huangli_handler.go as a separate, user-independent API.
package http

import (
	"net/http"

	"github.com/25types/25types/internal/ganzhi"
	minglihttp "github.com/25types/25types/internal/mingli/http"
)

// DailyHandler serves daily suggestion endpoints.
type DailyHandler struct {
	Users BirthInfoLookup
}

// GET /api/daily/suggestion
func (h *DailyHandler) Suggestion(w http.ResponseWriter, r *http.Request) {
	var dayMaster ganzhi.Stem
	if h.Users != nil {
		userID, ok := UserID(r.Context())
		if ok {
			u, err := h.Users.FindByID(r.Context(), userID)
			if err == nil && u != nil && u.BirthInfo != nil {
				bi := u.BirthInfo
				if bi.Longitude == 0 {
					bi.Longitude = 120.0
				}
				if bi.Timezone == 0 {
					bi.Timezone = 8.0
				}
				dayMaster = minglihttp.DayMasterFromBirthInfo(
					bi.Year, bi.Month, bi.Day,
					bi.Hour, bi.Minute,
					bi.Longitude, bi.Timezone,
				)
			}
		}
	}
	ds := minglihttp.ComputeDailySuggestion(dayMaster)
	respondJSON(w, http.StatusOK, ds)
}

// GET /api/daily/question
func (h *DailyHandler) Question(w http.ResponseWriter, r *http.Request) {
	var dayMaster ganzhi.Stem
	if h.Users != nil {
		userID, ok := UserID(r.Context())
		if ok {
			u, err := h.Users.FindByID(r.Context(), userID)
			if err == nil && u != nil && u.BirthInfo != nil {
				bi := u.BirthInfo
				if bi.Longitude == 0 {
					bi.Longitude = 120.0
				}
				if bi.Timezone == 0 {
					bi.Timezone = 8.0
				}
				dayMaster = minglihttp.DayMasterFromBirthInfo(
					bi.Year, bi.Month, bi.Day,
					bi.Hour, bi.Minute,
					bi.Longitude, bi.Timezone,
				)
			}
		}
	}
	ds := minglihttp.ComputeDailySuggestion(dayMaster)
	respondJSON(w, http.StatusOK, map[string]string{
		"date":     ds.Date,
		"question": ds.Question,
	})
}
