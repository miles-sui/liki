package minglihttp

import (
	"encoding/json"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/25types/25types/internal/ganzhi"
	"github.com/25types/25types/internal/httputil"
	"github.com/25types/25types/internal/mingli/bazi"
)

// MingliHandler serves BaZi (八字) computation endpoints — all POST with explicit birth params.
type MingliHandler struct{}

type birthParams struct {
	Year      int     `json:"year"`
	Month     int     `json:"month"`
	Day       int     `json:"day"`
	Hour      int     `json:"hour"`
	Minute    int     `json:"minute"`
	Longitude float64 `json:"longitude"`
	Timezone  float64 `json:"timezone"`
	Gender    string  `json:"gender"`
}

func (p birthParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.Year, validation.Required, validation.Min(1900), validation.Max(2200)),
		validation.Field(&p.Month, validation.Required, validation.Min(1), validation.Max(12)),
		validation.Field(&p.Day, validation.Required, validation.Min(1), validation.Max(31)),
		validation.Field(&p.Hour, validation.Min(0), validation.Max(23)),
		validation.Field(&p.Minute, validation.Min(0), validation.Max(59)),
		validation.Field(&p.Longitude,
			validation.Min(-180.0).Error("longitude must be -180 to 180"),
			validation.Max(180.0).Error("longitude must be -180 to 180"),
		),
		validation.Field(&p.Timezone, validation.Min(-12.0), validation.Max(14.0)),
		validation.Field(&p.Gender, validation.Required, validation.In("male", "female")),
	)
}

func (p *birthParams) computeChart() bazi.ChartResult {
	return bazi.ComputeChartFromBirth(p.Year, p.Month, p.Day, p.Hour, p.Minute, p.Longitude, p.Timezone, bazi.Gender(p.Gender))
}

// POST /api/bazi/chart
func (h *MingliHandler) ComputeChart(w http.ResponseWriter, r *http.Request) {
	var bp birthParams
	if err := json.NewDecoder(r.Body).Decode(&bp); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "invalid_birth_info", "Invalid JSON body")
		return
	}

	if err := bp.Validate(); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "invalid_birth_info", err.Error())
		return
	}

	ch := bp.computeChart()
	httputil.RespondJSON(w, http.StatusOK, bazi.BuildChartOutput(ch, bp.Year, bp.Month, bp.Hour))
}

// POST /api/bazi/bond
func (h *MingliHandler) BondCharts(w http.ResponseWriter, r *http.Request) {
	var req struct {
		A *birthParams `json:"a"`
		B *birthParams `json:"b"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "invalid_birth_info", "Invalid JSON body")
		return
	}

	if req.A == nil || req.B == nil {
		httputil.RespondError(w, http.StatusBadRequest, "chart_required", "Both a and b birth info are required")
		return
	}

	if err := req.A.Validate(); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "invalid_birth_info", err.Error())
		return
	}
	if err := req.B.Validate(); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "invalid_birth_info", err.Error())
		return
	}

	ca := req.A.computeChart()
	cb := req.B.computeChart()

	bond := bazi.ComputeBond(ca, cb, req.A.Year, req.A.Month, req.A.Hour, req.B.Year, req.B.Month, req.B.Hour)

	httputil.RespondJSON(w, http.StatusOK, bazi.BondOutput{
		ChartA: bazi.BuildChartOutput(ca, req.A.Year, req.A.Month, req.A.Hour),
		ChartB: bazi.BuildChartOutput(cb, req.B.Year, req.B.Month, req.B.Hour),
		Bond:   bond,
	})
}

func (h *MingliHandler) LiuNian(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Bazi         ganzhi.Bazi `json:"bazi"`
		Year         int         `json:"year"`
		CurrentDayun *ganzhi.Pillar `json:"current_dayun,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}
	if err := validation.ValidateStruct(&req,
		validation.Field(&req.Year, validation.Required, validation.Min(1900), validation.Max(2100)),
	); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	var cd *bazi.DayunPillar
	if req.CurrentDayun != nil {
		cd = &bazi.DayunPillar{Stem: req.CurrentDayun.Stem, Branch: req.CurrentDayun.Branch}
	}

	result := bazi.ComputeLiunian(req.Year, req.Bazi.Day.Stem, req.Bazi, cd)
	httputil.RespondJSON(w, http.StatusOK, result)
}

// POST /api/bazi/liuyue
func (h *MingliHandler) LiuYue(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Bazi  ganzhi.Bazi `json:"bazi"`
		Year  int         `json:"year"`
		Month int         `json:"month"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}
	if err := validation.ValidateStruct(&req,
		validation.Field(&req.Month, validation.Required, validation.Min(1), validation.Max(12)),
	); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	result := bazi.ComputeLiuyue(req.Year, req.Month, req.Bazi.Day.Stem, req.Bazi)
	httputil.RespondJSON(w, http.StatusOK, result)
}

// POST /api/bazi/liuri
func (h *MingliHandler) LiuRi(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Bazi          ganzhi.Bazi  `json:"bazi"`
		Date          string       `json:"date"`
		DayunPillar   *ganzhi.Pillar `json:"dayun_pillar,omitempty"`
		LiunianPillar *ganzhi.Pillar `json:"liunian_pillar,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	result := bazi.ComputeLiuri(req.Date, req.Bazi.Day.Stem, req.Bazi, req.DayunPillar, req.LiunianPillar)
	if result == nil {
		httputil.RespondError(w, http.StatusInternalServerError, "compute_error", "failed to compute liuri")
		return
	}
	httputil.RespondJSON(w, http.StatusOK, result)
}

// POST /api/bazi/liushi
func (h *MingliHandler) LiuShi(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Bazi ganzhi.Bazi `json:"bazi"`
		Date string      `json:"date"`
		Hour int         `json:"hour"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}
	if err := validation.ValidateStruct(&req,
		validation.Field(&req.Hour, validation.Required, validation.Min(0), validation.Max(23)),
	); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	result := bazi.ComputeLiushi(req.Date, req.Hour, req.Bazi.Day.Stem, req.Bazi)
	if result == nil {
		httputil.RespondError(w, http.StatusInternalServerError, "compute_error", "failed to compute liushi")
		return
	}
	httputil.RespondJSON(w, http.StatusOK, result)
}

// POST /api/bazi/xiao-yun
func (h *MingliHandler) XiaoYun(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Birth birthParams `json:"birth"`
		Count int         `json:"count"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}
		if err := req.Birth.Validate(); err != nil {
			httputil.RespondError(w, http.StatusBadRequest, "invalid_birth_info", err.Error())
			return
		}
	chart := req.Birth.computeChart()

	pillars := bazi.ComputeXiaoYun(bazi.Gender(req.Birth.Gender), chart.DayMaster, req.Count)
	httputil.RespondJSON(w, http.StatusOK, pillars)
}

// POST /api/bazi/xiao-xian
func (h *MingliHandler) XiaoXian(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Gender string `json:"gender"`
		Count  int    `json:"count"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}
	if err := validation.ValidateStruct(&req,
		validation.Field(&req.Gender, validation.Required, validation.In("male", "female")),
	); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	entries := bazi.ComputeXiaoXian(bazi.Gender(req.Gender), req.Count)
	httputil.RespondJSON(w, http.StatusOK, entries)
}

