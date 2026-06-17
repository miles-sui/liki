package handler

import (
	"net/http"
	"time"

	"liki/internal/engine/ganzhi"
	"liki/internal/engine/tianwen"
)

func computeSolarTime(w http.ResponseWriter, r *http.Request) {
	bp, ok := decodeJSON[SolarTime](w, r)
	if !ok {
		return
	}
	if err := validateSolarParams(&bp); err != nil {
		respondValidationError(w, err)
		return
	}
	applySolarDefaults(&bp)
	st := tianwen.ComputeSolarTime(bp.Year, bp.Month, bp.Day, bp.Hour, bp.Minute, bp.Longitude, bp.Timezone)
	br := ganzhi.Zhi((int(st.Minutes()+60)/120)%12 + 1)
	respondJSON(w, http.StatusOK, map[string]any{
		"solar_time":       st.Time().Format(time.RFC3339),
		"hour_branch":      int(br),
		"hour_branch_name": br.String(),
	})
}
