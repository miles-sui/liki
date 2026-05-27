package minglihttp

import (
	"net/http"
	"time"

	"github.com/25types/25types/internal/httputil"
	"github.com/25types/25types/internal/tianwen"
)

// SolarTerms returns the solar term calendar for the current year.
func SolarTerms(w http.ResponseWriter, r *http.Request) {
	year := time.Now().Year()
	entries := tianwen.PrecomputeSolarTerms(year)
	currentID := tianwen.GetCurrentSolarMonth(time.Now())

	months := []SolarTermMonth{}
	var current SolarTermMonth
	for i, e := range entries {
		nextIdx := (i + 1) % 12
		sm := SolarTermMonth{
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

	httputil.RespondJSON(w, http.StatusOK, SolarTermsResponse{
		Year:    year,
		Current: current,
		Months:  months,
	})
}
