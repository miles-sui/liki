package handler

import (
	"errors"
	"net/http"
	"regexp"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"liki/internal/engine/bazi"
	"liki/internal/engine/ganzhi"
	"liki/internal/engine/tianwen"
)

// YMDRegexp matches YYYY-MM-DD date strings.
var YMDRegexp = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)



func validGan(p ganzhi.Zhu) bool { return p.Gan >= 1 && p.Gan <= 10 }

func isValidDate(y, m, d int) bool {
	if m < 1 || m > 12 || d < 1 {
		return false
	}
	return d <= time.Date(y, time.Month(m)+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func applySolarDefaults(p *SolarTime) {
	if p.Longitude == 0 {
		p.Longitude = 120
	}
	if p.Timezone == 0 {
		p.Timezone = 8
	}
}
func validateSolarParams(p *SolarTime) error {
	e := validation.ValidateStruct(p,
		validation.Field(&p.Year, validation.Required, validation.Min(1900), validation.Max(2200)),
		validation.Field(&p.Month, validation.Required, validation.Min(1), validation.Max(12)),
		validation.Field(&p.Day, validation.Required, validation.Min(1), validation.Max(31)),
		validation.Field(&p.Hour, validation.Min(0), validation.Max(23)),
		validation.Field(&p.Minute, validation.Min(0), validation.Max(59)),
		validation.Field(&p.Longitude, validation.Min(-180.0), validation.Max(180.0)),
		validation.Field(&p.Timezone, validation.Min(-12.0), validation.Max(14.0)),
		validation.Field(&p.Gender, validation.Required, validation.In(genderSet...)),
	)
	if e != nil { return e }
	if !isValidDate(p.Year, p.Month, p.Day) { return errors.New("invalid date") }
	return nil
}

type SolarTime struct {
	Year, Month, Day, Hour, Minute int
	Longitude, Timezone            float64
	Gender                         string
}
type bondRequest struct{ A, B *SolarTime }
type liuNianRequest  struct { Bazi ganzhi.Bazi `json:"bazi"`; Year int; CurrentDaYun ganzhi.Zhu `json:"current_dayun"` }
type liuYueRequest   struct { Bazi ganzhi.Bazi `json:"bazi"`; Year, Month int }
type liuRiRequest    struct { Bazi ganzhi.Bazi `json:"bazi"`; Date string; DaYunPillar, LiuNianPillar ganzhi.Zhu }
type liuShiRequest   struct { Bazi ganzhi.Bazi `json:"bazi"`; Date string; Hour int }
type xiaoYunRequest  struct { Gender string `json:"gender"`; DayMaster int `json:"day_master"`; Count int `json:"count"` }
type xiaoXianRequest struct { Gender string; Count int }

func computeChart(w http.ResponseWriter, r *http.Request) {
	bp, ok := decodeJSON[SolarTime](w, r)
	if !ok { return }
	if err := validateSolarParams(&bp); err != nil { respondValidationError(w, err); return }
	bt := tianwen.ComputeBirthTime(bp.Year, bp.Month, bp.Day, bp.Hour, bp.Minute, bp.Longitude, bp.Timezone)
	cr := bazi.ComputeChart(bt.Solar, ganzhi.Gender(bp.Gender))
	respondJSON(w, http.StatusOK, cr)
}

func bondCharts(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeJSON[bondRequest](w, r)
	if !ok { return }
	if req.A == nil || req.B == nil { respondInvalidRequest(w, "Both a and b are required"); return }
	if e := validateSolarParams(req.A); e != nil { respondValidationError(w, e); return }
	if e := validateSolarParams(req.B); e != nil { respondValidationError(w, e); return }
	bta := tianwen.ComputeBirthTime(req.A.Year, req.A.Month, req.A.Day, req.A.Hour, req.A.Minute, req.A.Longitude, req.A.Timezone)
	btb := tianwen.ComputeBirthTime(req.B.Year, req.B.Month, req.B.Day, req.B.Hour, req.B.Minute, req.B.Longitude, req.B.Timezone)
	cra := bazi.ComputeChart(bta.Solar, ganzhi.Gender(req.A.Gender))
	crb := bazi.ComputeChart(btb.Solar, ganzhi.Gender(req.B.Gender))
	result := bazi.ComputeBond(cra.ChartBase, crb.ChartBase)
	respondJSON(w, http.StatusOK, result)
}

func liuNian(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeJSON[liuNianRequest](w, r)
	if !ok { return }
	if e := validation.ValidateStruct(&req, validation.Field(&req.Year, validation.Required, validation.Min(1900), validation.Max(2100))); e != nil { respondValidationError(w, e); return }
	if e := req.Bazi.Validate(); e != nil { respondValidationError(w, e); return }
var cd *bazi.DaYunPillar
	if validGan(req.CurrentDaYun) { cd = &bazi.DaYunPillar{Gan: req.CurrentDaYun.Gan, Zhi: req.CurrentDaYun.Zhi} }
	result := bazi.ComputeLiuNian(req.Year, req.Bazi.Ri.Gan, req.Bazi, cd)
	respondJSON(w, http.StatusOK, result)
}

func liuYue(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeJSON[liuYueRequest](w, r)
	if !ok { return }
	if e := validation.ValidateStruct(&req, validation.Field(&req.Year, validation.Required, validation.Min(1900), validation.Max(2200)), validation.Field(&req.Month, validation.Required, validation.Min(1), validation.Max(12))); e != nil { respondValidationError(w, e); return }
	if e := req.Bazi.Validate(); e != nil { respondValidationError(w, e); return }
	result := bazi.ComputeLiuYue(req.Year, req.Month, req.Bazi.Ri.Gan, req.Bazi)
	respondJSON(w, http.StatusOK, result)
}

func liuRi(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeJSON[liuRiRequest](w, r)
	if !ok { return }
	if e := validation.ValidateStruct(&req, validation.Field(&req.Date, validation.Required, validation.Match(YMDRegexp))); e != nil { respondValidationError(w, e); return }
	if e := req.Bazi.Validate(); e != nil { respondValidationError(w, e); return }
	var dp, lp *ganzhi.Zhu
	if validGan(req.DaYunPillar) { dp = &ganzhi.Zhu{Gan: req.DaYunPillar.Gan, Zhi: req.DaYunPillar.Zhi} }
	if validGan(req.LiuNianPillar) { lp = &ganzhi.Zhu{Gan: req.LiuNianPillar.Gan, Zhi: req.LiuNianPillar.Zhi} }
	result := bazi.ComputeLiuRi(req.Date, req.Bazi.Ri.Gan, req.Bazi, dp, lp)
	if result == nil { respondError(w, http.StatusInternalServerError, "compute_error", "failed"); return }
	respondJSON(w, http.StatusOK, result)
}

func liuShi(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeJSON[liuShiRequest](w, r)
	if !ok { return }
	if e := validation.ValidateStruct(&req, validation.Field(&req.Date, validation.Required, validation.Match(YMDRegexp)), validation.Field(&req.Hour, validation.Required, validation.Min(0), validation.Max(23))); e != nil { respondValidationError(w, e); return }
	if e := req.Bazi.Validate(); e != nil { respondValidationError(w, e); return }
	result := bazi.ComputeLiuShi(req.Date, req.Hour, req.Bazi.Ri.Gan, req.Bazi)
	if result == nil { respondError(w, http.StatusInternalServerError, "compute_error", "failed"); return }
	respondJSON(w, http.StatusOK, result)
}

func xiaoYun(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeJSON[xiaoYunRequest](w, r)
	if !ok { return }
	if e := validation.ValidateStruct(&req, validation.Field(&req.Gender, validation.Required, validation.In(genderSet...)), validation.Field(&req.DayMaster, validation.Required, validation.Min(1), validation.Max(10)), validation.Field(&req.Count, validation.Required, validation.Min(1), validation.Max(120))); e != nil { respondValidationError(w, e); return }
	pillars := bazi.ComputeXiaoYun(ganzhi.Gender(req.Gender), ganzhi.Gan(req.DayMaster), req.Count)
	respondJSON(w, http.StatusOK, pillars)
}

func xiaoXian(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeJSON[xiaoXianRequest](w, r)
	if !ok { return }
	if e := validation.ValidateStruct(&req, validation.Field(&req.Gender, validation.Required, validation.In(genderSet...)), validation.Field(&req.Count, validation.Required, validation.Min(1), validation.Max(120))); e != nil { respondValidationError(w, e); return }
	entries := bazi.ComputeXiaoXian(ganzhi.Gender(req.Gender), req.Count)
	respondJSON(w, http.StatusOK, entries)
}
