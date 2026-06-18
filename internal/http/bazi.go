package handler

import (
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"liki/internal/engine/bazi"
	"liki/internal/engine/ganzhi"
	"liki/internal/engine/tianwen"
)

type chartRequest struct {
	Birth  timePoint     `json:"birth"`
	Gender ganzhi.Gender `json:"gender"`
}

func (r chartRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Birth, validation.By(validateTimePoint)),
		validation.Field(&r.Gender, validation.Required, validation.In(validGenders...)),
	)
}

type bondRequest struct {
	A chartRequest `json:"a"`
	B chartRequest `json:"b"`
}

func (r bondRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.A, validation.By(validateChartRequest)),
		validation.Field(&r.B, validation.By(validateChartRequest)),
	)
}

func validateChartRequest(value any) error {
	c, ok := value.(chartRequest)
	if !ok {
		return nil
	}
	return c.Validate()
}

type liuNianRequest struct {
	Year  int        `json:"year"`
	Birth timePoint `json:"birth"`
}

func (r liuNianRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Birth, validation.By(validateTimePoint)),
		validation.Field(&r.Year, validation.Required, validation.Min(1900), validation.Max(2100)),
	)
}

type liuYueRequest struct {
	Year  int        `json:"year"`
	Month int        `json:"month"`
	Birth timePoint `json:"birth"`
}

func (r liuYueRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Birth, validation.By(validateTimePoint)),
		validation.Field(&r.Year, validation.Required, validation.Min(1900), validation.Max(2100)),
		validation.Field(&r.Month, validation.Required, validation.Min(1), validation.Max(12)),
	)
}

type liuRiRequest struct {
	Date  tianwen.GregorianTime `json:"date"`
	Birth timePoint            `json:"birth"`
}

func (r liuRiRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Birth, validation.By(validateTimePoint)),
		validation.Field(&r.Date, validation.By(validateNonZeroGregorianTime)),
	)
}

type liuShiRequest struct {
	Date  tianwen.GregorianTime `json:"date"`
	Hour  int                   `json:"hour"`
	Birth timePoint            `json:"birth"`
}

func (r liuShiRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Birth, validation.By(validateTimePoint)),
		validation.Field(&r.Date, validation.By(validateNonZeroGregorianTime)),
		validation.Field(&r.Hour, validation.Min(0), validation.Max(23)),
	)
}

type xiaoYunRequest struct {
	Birth  timePoint     `json:"birth"`
	Gender ganzhi.Gender `json:"gender"`
	Count  int            `json:"count"`
}

func (r xiaoYunRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Birth, validation.By(validateTimePoint)),
		validation.Field(&r.Gender, validation.Required, validation.In(validGenders...)),
		validation.Field(&r.Count, validation.Required, validation.Min(1), validation.Max(120)),
	)
}

type xiaoXianRequest struct {
	Gender ganzhi.Gender `json:"gender"`
	Count  int           `json:"count"`
}

func (r xiaoXianRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Gender, validation.Required, validation.In(validGenders...)),
		validation.Field(&r.Count, validation.Required, validation.Min(1), validation.Max(120)),
	)
}

func computeChart(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeAndValidate[chartRequest](w, r)
	if !ok {
		return
	}
	ts, err := req.Birth.Timeset()
	if err != nil {
		respondInvalidRequest(w, err.Error())
		return
	}
	cr := bazi.ComputeChart(ts.Solar, req.Gender)
	respondJSON(w, http.StatusOK, cr)
}

func bondCharts(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeAndValidate[bondRequest](w, r)
	if !ok {
		return
	}
	tsA, err := req.A.Birth.Timeset()
	if err != nil {
		respondInvalidRequest(w, err.Error())
		return
	}
	tsB, err := req.B.Birth.Timeset()
	if err != nil {
		respondInvalidRequest(w, err.Error())
		return
	}
	cra := bazi.ComputeChart(tsA.Solar, req.A.Gender)
	crb := bazi.ComputeChart(tsB.Solar, req.B.Gender)
	result := bazi.ComputeBond(cra.ChartBase, crb.ChartBase)
	respondJSON(w, http.StatusOK, result)
}

func liuNian(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeAndValidate[liuNianRequest](w, r)
	if !ok {
		return
	}
	ts, err := req.Birth.Timeset()
	if err != nil {
		respondInvalidRequest(w, err.Error())
		return
	}
	result, err := bazi.ComputeLiuNian(ts.Solar, req.Year, nil)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "compute_error", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func liuYue(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeAndValidate[liuYueRequest](w, r)
	if !ok {
		return
	}
	ts, err := req.Birth.Timeset()
	if err != nil {
		respondInvalidRequest(w, err.Error())
		return
	}
	result, err := bazi.ComputeLiuYue(ts.Solar, req.Year, req.Month)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "compute_error", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func liuRi(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeAndValidate[liuRiRequest](w, r)
	if !ok {
		return
	}
	ts, err := req.Birth.Timeset()
	if err != nil {
		respondInvalidRequest(w, err.Error())
		return
	}
	dateStr := req.Date.Time().Format("2006-01-02")
	result, err := bazi.ComputeLiuRi(ts.Solar, dateStr, nil, nil)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "compute_error", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func liuShi(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeAndValidate[liuShiRequest](w, r)
	if !ok {
		return
	}
	ts, err := req.Birth.Timeset()
	if err != nil {
		respondInvalidRequest(w, err.Error())
		return
	}
	dateStr := req.Date.Time().Format("2006-01-02")
	result, err := bazi.ComputeLiuShi(ts.Solar, dateStr, req.Hour)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "compute_error", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func xiaoYun(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeAndValidate[xiaoYunRequest](w, r)
	if !ok {
		return
	}
	ts, err := req.Birth.Timeset()
	if err != nil {
		respondInvalidRequest(w, err.Error())
		return
	}
	pillars := bazi.ComputeXiaoYun(ts.Solar, req.Gender, req.Count)
	respondJSON(w, http.StatusOK, pillars)
}

func xiaoXian(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeAndValidate[xiaoXianRequest](w, r)
	if !ok {
		return
	}
	entries := bazi.ComputeXiaoXian(req.Gender, req.Count)
	respondJSON(w, http.StatusOK, entries)
}
