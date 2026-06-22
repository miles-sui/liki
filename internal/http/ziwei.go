package handler

import (
	"errors"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"liki/internal/engine/ganzhi"
	"liki/internal/engine/ziwei"
)

func validateZiweiChart(value any) error {
	var chart ziwei.Chart
	switch v := value.(type) {
	case ziwei.Chart:
		chart = v
	case *ziwei.Chart:
		if v == nil {
			return errors.New("required")
		}
		chart = *v
	default:
		return nil
	}
	if chart.JuShu < 2 || chart.JuShu > 6 {
		return errors.New("invalid chart")
	}
	return nil
}

func computeZiweiChart(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeAndValidate[BirthRequest](w, r)
	if !ok {
		return
	}
	ts, ok := timesetOrRespond(w, req.Birth)
	if !ok {
		return
	}
	result := ziwei.ComputeChart(ts.Solar, req.Gender)
	respondJSON(w, http.StatusOK, result)
}

type ziweiDaxianRequest struct {
	Chart  ziwei.Chart   `json:"chart"`
	Gender ganzhi.Gender `json:"gender"`
}

func (r ziweiDaxianRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Gender, validation.Required, validation.In(validGenders...)),
		validation.Field(&r.Chart, validation.By(validateZiweiChart)),
	)
}

func computeZiweiDaxian(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeAndValidate[ziweiDaxianRequest](w, r)
	if !ok {
		return
	}
	steps := ziwei.ComputeDaXian(req.Chart)
	respondJSON(w, http.StatusOK, steps)
}

type ziweiLiunianRequest struct {
	LiuYear int         `json:"liu_year"`
	Chart   ziwei.Chart `json:"chart"`
}

func (r ziweiLiunianRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.LiuYear, validation.Required, validation.Min(1900), validation.Max(2100)),
		validation.Field(&r.Chart, validation.By(validateZiweiChart)),
	)
}

func computeZiweiLiunian(w http.ResponseWriter, r *http.Request) {
	p, ok := decodeAndValidate[ziweiLiunianRequest](w, r)
	if !ok {
		return
	}
	result := ziwei.ComputeLiuNian(p.LiuYear, p.Chart)
	respondJSON(w, http.StatusOK, result)
}

type ziweiLiuyueRequest struct {
	LiuYear    int         `json:"liu_year"`
	LunarMonth int         `json:"lunar_month"`
	Chart      ziwei.Chart `json:"chart"`
}

func (r ziweiLiuyueRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.LiuYear, validation.Required, validation.Min(1900), validation.Max(2100)),
		validation.Field(&r.LunarMonth, validation.Required, validation.Min(1), validation.Max(12)),
		validation.Field(&r.Chart, validation.By(validateZiweiChart)),
	)
}

func computeZiweiLiuyue(w http.ResponseWriter, r *http.Request) {
	p, ok := decodeAndValidate[ziweiLiuyueRequest](w, r)
	if !ok {
		return
	}
	result := ziwei.ComputeLiuYue(p.LiuYear, p.LunarMonth, p.Chart)
	respondJSON(w, http.StatusOK, result)
}

type ziweiLiuriRequest struct {
	LiuYear    int         `json:"liu_year"`
	LunarMonth int         `json:"lunar_month"`
	LunarDay   int         `json:"lunar_day"`
	Chart      ziwei.Chart `json:"chart"`
}

func (r ziweiLiuriRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.LiuYear, validation.Required, validation.Min(1900), validation.Max(2100)),
		validation.Field(&r.LunarMonth, validation.Required, validation.Min(1), validation.Max(12)),
		validation.Field(&r.LunarDay, validation.Required, validation.Min(1), validation.Max(30)),
		validation.Field(&r.Chart, validation.By(validateZiweiChart)),
	)
}

func computeZiweiLiuri(w http.ResponseWriter, r *http.Request) {
	p, ok := decodeAndValidate[ziweiLiuriRequest](w, r)
	if !ok {
		return
	}
	result := ziwei.ComputeLiuRi(p.LiuYear, p.LunarMonth, p.LunarDay, p.Chart)
	respondJSON(w, http.StatusOK, result)
}

type ziweiBondRequest struct {
	A *ziwei.Chart `json:"a"`
	B *ziwei.Chart `json:"b"`
}

func (r ziweiBondRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.A, validation.Required, validation.By(validateZiweiChart)),
		validation.Field(&r.B, validation.Required, validation.By(validateZiweiChart)),
	)
}

func computeZiweiBond(w http.ResponseWriter, r *http.Request) {
	p, ok := decodeAndValidate[ziweiBondRequest](w, r)
	if !ok {
		return
	}
	result := ziwei.ComputeBond(*p.A, *p.B)
	respondJSON(w, http.StatusOK, result)
}
