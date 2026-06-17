package handler

import (
	"net/http"

	"liki/internal/engine/ziwei"
	"liki/internal/engine/ganzhi"
	"liki/internal/engine/tianwen"
)

// ziweiChartRequest is the input for POST /api/ziwei/chart.
type ziweiChartRequest struct {
	Year      int     `json:"year"`
	Month     int     `json:"month"`
	Day       int     `json:"day"`
	Hour      int     `json:"hour"`
	Minute    int     `json:"minute"`
	Longitude float64 `json:"longitude"`
	Timezone  float64 `json:"timezone"`
	Gender    string  `json:"gender"`

	// Ziwei-specific: lunar month/day (required).
	LunarMonth int `json:"lunar_month"`
	LunarDay   int `json:"lunar_day"`
}

// computeZiweiChart handles POST /api/ziwei/chart.
func computeZiweiChart(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeJSON[ziweiChartRequest](w, r)
	if !ok { return }
	bp := SolarTime{Year: req.Year, Month: req.Month, Day: req.Day, Hour: req.Hour, Minute: req.Minute,
		Longitude: req.Longitude, Timezone: req.Timezone, Gender: req.Gender}
	if err := validateSolarParams(&bp); err != nil { respondValidationError(w, err); return }

	bt := tianwen.ComputeBirthTime(req.Year, req.Month, req.Day, req.Hour, req.Minute, req.Longitude, req.Timezone)
	bz := tianwen.ComputeBazi(bt.Solar)
	lm, ld := req.LunarMonth, req.LunarDay
	if lm == 0 {
		lm, ld = bt.Lunar.Month, bt.Lunar.Day
	}
	result := ziwei.ComputeChart(req.Year, lm, ld, bz.Shi.Zhi, bz.Nian.Gan, bz.Nian.Zhi, ganzhi.Gender(req.Gender))
respondJSON(w, http.StatusOK, result)
}

// ziweiDaxianRequest is POST /api/ziwei/daxian.
type ziweiDaxianRequest struct {
	Chart  ziwei.Chart `json:"chart"`
	Gender string      `json:"gender"`
}

func computeZiweiDaxian(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeJSON[ziweiDaxianRequest](w, r)
	if !ok {
		return
	}
	if req.Gender != "male" && req.Gender != "female" {
		respondValidationError(w, nil)
		return
	}
	steps := ziwei.ComputeDaXian(req.Chart)
	respondJSON(w, http.StatusOK, map[string]any{"da_xian": steps})
}

// ziweiLiunianRequest is POST /api/ziwei/liunian.
type ziweiLiunianRequest struct {
	Chart     ziwei.Chart `json:"chart"`
	BirthYear int         `json:"birth_year"`
	LiuYear   int         `json:"liu_year"`
}

func computeZiweiLiunian(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeJSON[ziweiLiunianRequest](w, r)
	if !ok {
		return
	}
	result := ziwei.ComputeLiuNian(req.LiuYear, req.Chart)
	respondJSON(w, http.StatusOK, result)
}

func computeZiweiLiuyue(w http.ResponseWriter, r *http.Request) {
	type req struct{ LiuYear, LunarMonth int; Chart ziwei.Chart }
	p, ok := decodeJSON[req](w, r)
	if !ok { return }
	result := ziwei.ComputeLiuYue(p.LiuYear, p.LunarMonth, p.Chart)
	respondJSON(w, http.StatusOK, result)
}

func computeZiweiLiuri(w http.ResponseWriter, r *http.Request) {
	type req struct{ LiuYear, LunarMonth, LunarDay int; Chart ziwei.Chart }
	p, ok := decodeJSON[req](w, r)
	if !ok { return }
	result := ziwei.ComputeLiuRi(p.LiuYear, p.LunarMonth, p.LunarDay, p.Chart)
	respondJSON(w, http.StatusOK, result)
}

func computeZiweiBond(w http.ResponseWriter, r *http.Request) {
	type req struct{ A, B ziwei.Chart }
	p, ok := decodeJSON[req](w, r)
	if !ok { return }
	result := ziwei.ComputeBond(p.A, p.B)
	respondJSON(w, http.StatusOK, result)
}
