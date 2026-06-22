package handler

import (
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"liki/internal/engine/bazi"
	"liki/internal/engine/ganzhi"
)

type bondRequest struct {
	A BirthRequest `json:"a"`
	B BirthRequest `json:"b"`
}

func (r bondRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.A, validation.By(validateBirthRequest)),
		validation.Field(&r.B, validation.By(validateBirthRequest)),
	)
}

func validateBirthRequest(value any) error {
	c, ok := value.(BirthRequest)
	if !ok {
		return nil
	}
	return c.Validate()
}

type liuNianRequest struct {
	Year   int            `json:"year"`
	Birth  timePoint      `json:"birth"`
	Gender ganzhi.Gender  `json:"gender"`
}

func (r liuNianRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Birth, validation.By(validateTimePoint)),
		validation.Field(&r.Year, validation.Required, validation.Min(1900), validation.Max(2100)),
		validation.Field(&r.Gender, validation.Required, validation.In(validGenders...)),
	)
}

type liuYueRequest struct {
	Year   int            `json:"year"`
	Month  int            `json:"month"`
	Birth  timePoint      `json:"birth"`
	Gender ganzhi.Gender  `json:"gender"`
}

func (r liuYueRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Birth, validation.By(validateTimePoint)),
		validation.Field(&r.Year, validation.Required, validation.Min(1900), validation.Max(2100)),
		validation.Field(&r.Month, validation.Required, validation.Min(1), validation.Max(12)),
		validation.Field(&r.Gender, validation.Required, validation.In(validGenders...)),
	)
}

type liuRiRequest struct {
	Year   int            `json:"year"`
	Month  int            `json:"month"`
	Day    int            `json:"day"`
	Birth  timePoint      `json:"birth"`
	Gender ganzhi.Gender  `json:"gender"`
}

func (r liuRiRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Birth, validation.By(validateTimePoint)),
		validation.Field(&r.Year, validation.Required, validation.Min(1900), validation.Max(2100)),
		validation.Field(&r.Month, validation.Required, validation.Min(1), validation.Max(12)),
		validation.Field(&r.Day, validation.Required, validation.Min(1), validation.Max(31)),
		validation.Field(&r.Gender, validation.Required, validation.In(validGenders...)),
	)
}

type liuShiRequest struct {
	Year   int            `json:"year"`
	Month  int            `json:"month"`
	Day    int            `json:"day"`
	Hour   int            `json:"hour"`
	Birth  timePoint      `json:"birth"`
	Gender ganzhi.Gender  `json:"gender"`
}

func (r liuShiRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Birth, validation.By(validateTimePoint)),
		validation.Field(&r.Year, validation.Required, validation.Min(1900), validation.Max(2100)),
		validation.Field(&r.Month, validation.Required, validation.Min(1), validation.Max(12)),
		validation.Field(&r.Day, validation.Required, validation.Min(1), validation.Max(31)),
		validation.Field(&r.Hour, validation.Min(0), validation.Max(23)),
		validation.Field(&r.Gender, validation.Required, validation.In(validGenders...)),
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
	req, ok := decodeAndValidate[BirthRequest](w, r)
	if !ok {
		return
	}
	ts, ok := timesetOrRespond(w, req.Birth)
	if !ok {
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
	tsA, ok := timesetOrRespond(w, req.A.Birth)
	if !ok {
		return
	}
	tsB, ok := timesetOrRespond(w, req.B.Birth)
	if !ok {
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
	ts, ok := timesetOrRespond(w, req.Birth)
	if !ok {
		return
	}
	chart := bazi.ComputeChart(ts.Solar, req.Gender)
	result, err := bazi.ComputeLiuNian(chart.ChartBase, req.Year)
	if err != nil {
		respondInvalidRequest(w, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func liuYue(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeAndValidate[liuYueRequest](w, r)
	if !ok {
		return
	}
	ts, ok := timesetOrRespond(w, req.Birth)
	if !ok {
		return
	}
	chart := bazi.ComputeChart(ts.Solar, req.Gender)
	result, err := bazi.ComputeLiuYue(chart.ChartBase, req.Year, req.Month)
	if err != nil {
		respondInvalidRequest(w, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func liuRi(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeAndValidate[liuRiRequest](w, r)
	if !ok {
		return
	}
	ts, ok := timesetOrRespond(w, req.Birth)
	if !ok {
		return
	}
	chart := bazi.ComputeChart(ts.Solar, req.Gender)
	result, err := bazi.ComputeLiuRi(chart.ChartBase, req.Year, req.Month, req.Day)
	if err != nil {
		respondInvalidRequest(w, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func liuShi(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeAndValidate[liuShiRequest](w, r)
	if !ok {
		return
	}
	ts, ok := timesetOrRespond(w, req.Birth)
	if !ok {
		return
	}
	chart := bazi.ComputeChart(ts.Solar, req.Gender)
	result, err := bazi.ComputeLiuShi(chart.ChartBase, req.Year, req.Month, req.Day, req.Hour)
	if err != nil {
		respondInvalidRequest(w, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func xiaoYun(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeAndValidate[xiaoYunRequest](w, r)
	if !ok {
		return
	}
	ts, ok := timesetOrRespond(w, req.Birth)
	if !ok {
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
