package minglihttp

import (
	"encoding/json"
	"fmt"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"net/http"
	"strconv"
	"time"

	"github.com/25types/25types/internal/ganzhi"
	"github.com/25types/25types/internal/httputil"
	"github.com/25types/25types/internal/mingli/bazi"
	"github.com/25types/25types/internal/mingli/fengshui"
)

// FengShuiHandler serves feng shui (风水) endpoints.
type FengShuiHandler struct{}

// GET /api/fengshui/san-yuan?year=2026
func (h *FengShuiHandler) GetSanYuan(w http.ResponseWriter, r *http.Request) {
	nowYear := time.Now().Year()
	if ys := r.URL.Query().Get("year"); ys != "" {
		if y, err := strconv.Atoi(ys); err == nil && y >= 1864 {
			nowYear = y
		}
	}

	current := fengshui.ComputeSanYuanYun(nowYear)
	allPeriods := fengshui.AllSanYuanYun()

	httputil.RespondJSON(w, http.StatusOK, struct {
		Current    fengshui.SanYuanYun   `json:"current"`
		AllPeriods [9]fengshui.SanYuanYun `json:"all_periods"`
	}{Current: current, AllPeriods: allPeriods})
}

// POST /api/fengshui/minggua
func (h *FengShuiHandler) MingGua(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Year   int    `json:"year"`
		Gender string `json:"gender"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}
	if err := validation.ValidateStruct(&req,
		validation.Field(&req.Year, validation.Required, validation.Min(1900), validation.Max(2200)),
		validation.Field(&req.Gender, validation.Required, validation.In("male", "female")),
	); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	mingGua := fengshui.ComputeMingGua(ganzhi.Gender(req.Gender), req.Year)
	allTrigrams := fengshui.AllTrigrams()

	httputil.RespondJSON(w, http.StatusOK, struct {
		MingGua     fengshui.MingGuaResult `json:"ming_gua"`
		AllTrigrams [9]fengshui.Trigram    `json:"all_trigrams"`
	}{MingGua: mingGua, AllTrigrams: allTrigrams})
}

// POST /api/fengshui/hecan
func (h *FengShuiHandler) HeCan(w http.ResponseWriter, r *http.Request) {
	var req struct {
		BirthYear int                `json:"birth_year"`
		Gender    string             `json:"gender"`
		Bazi      ganzhi.Bazi        `json:"bazi"`
		YongShen  bazi.YongShenResult `json:"yong_shen"`
		Year      int                `json:"year"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}
	if err := validation.ValidateStruct(&req,
		validation.Field(&req.BirthYear, validation.Required, validation.Min(1900), validation.Max(2200)),
		validation.Field(&req.Gender, validation.Required, validation.In("male", "female")),
		validation.Field(&req.Year, validation.Required, validation.Min(1864)),
	); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if err := req.Bazi.Validate(); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if err := validateYongShen(req.YongShen); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	result := fengshui.ComputeHeCan(
		req.BirthYear,
		ganzhi.Gender(req.Gender),
		req.Bazi,
		req.YongShen,
		req.Year,
	)
	httputil.RespondJSON(w, http.StatusOK, result)
}

var validWuxing = map[string]bool{"金": true, "木": true, "水": true, "火": true, "土": true}

func validateYongShen(ys bazi.YongShenResult) error {
	if !validWuxing[ys.FuYi.Yong] {
		return validation.Errors{"yong_shen.fuyi.yong": fmt.Errorf("must be a valid wuxing (金/木/水/火/土)")}
	}
	if !validWuxing[ys.FuYi.Xi] {
		return validation.Errors{"yong_shen.fuyi.xi": fmt.Errorf("must be a valid wuxing (金/木/水/火/土)")}
	}
	if !validWuxing[ys.FuYi.Ji] {
		return validation.Errors{"yong_shen.fuyi.ji": fmt.Errorf("must be a valid wuxing (金/木/水/火/土)")}
	}
	if !validWuxing[ys.TiaoHou.Yong] {
		return validation.Errors{"yong_shen.tiaohou.yong": fmt.Errorf("must be a valid wuxing (金/木/水/火/土)")}
	}
	if !validWuxing[ys.TiaoHou.Xi] {
		return validation.Errors{"yong_shen.tiaohou.xi": fmt.Errorf("must be a valid wuxing (金/木/水/火/土)")}
	}
	if !validWuxing[ys.TiaoHou.Ji] {
		return validation.Errors{"yong_shen.tiaohou.ji": fmt.Errorf("must be a valid wuxing (金/木/水/火/土)")}
	}
	return nil
}
