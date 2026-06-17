package handler

import (
	"net/http"

	"github.com/go-ozzo/ozzo-validation/v4"

	"liki/internal/engine/bazhai"
	"liki/internal/engine/ganzhi"
	"liki/internal/engine/tianwen"
	"liki/internal/engine/xuankong"
)

func getSanYuan(w http.ResponseWriter, r *http.Request) {
	type req struct{ Year int }
	p, _ := decodeJSON[req](w, r)
	if p.Year == 0 {
		p.Year = 2024
	}
	result := struct {
		Current xuankong.SanYuanYun `json:"current"`
	}{Current: xuankong.ComputeSanYuanYun(p.Year)}
	respondJSON(w, http.StatusOK, result)
}

func mingGua(w http.ResponseWriter, r *http.Request) {
	type req struct {
		Gender    string `json:"gender"`
		BirthYear int    `json:"birth_year"`
	}
	q, _ := decodeJSON[req](w, r)
	g := ganzhi.Male
	if q.Gender == "female" {
		g = ganzhi.Female
	}
	result := bazhai.ComputeMingGua(g, q.BirthYear)
	respondJSON(w, http.StatusOK, result)
}

func fengshuiChart(w http.ResponseWriter, r *http.Request) {
	type req struct {
		SolarTime tianwen.SolarTime `json:"solar_time"`
		Gender    string            `json:"gender"`
	}
	q, ok := decodeJSON[req](w, r)
	if !ok {
		return
	}
	if e := validation.ValidateStruct(&q,
		validation.Field(&q.SolarTime, validation.Required),
		validation.Field(&q.Gender, validation.Required, validation.In(genderSet...)),
	); e != nil {
		respondValidationError(w, e)
		return
	}
	g := ganzhi.Male
	if q.Gender == "female" {
		g = ganzhi.Female
	}
	result := bazhai.ComputeChart(q.SolarTime, g)
	respondJSON(w, http.StatusOK, result)
}

func xuankongChart(w http.ResponseWriter, r *http.Request) {
	type req struct {
		SolarTime    tianwen.SolarTime `json:"solar_time"`
		SitMountain  int               `json:"sit_mountain"`
		FaceMountain int               `json:"face_mountain"`
	}
	q, ok := decodeJSON[req](w, r)
	if !ok {
		return
	}
	if e := validation.ValidateStruct(&q,
		validation.Field(&q.SolarTime, validation.Required),
		validation.Field(&q.SitMountain, validation.Required, validation.Min(0), validation.Max(23)),
		validation.Field(&q.FaceMountain, validation.Required, validation.Min(0), validation.Max(23)),
	); e != nil {
		respondValidationError(w, e)
		return
	}
	result := xuankong.ComputeChart(q.SolarTime, q.SitMountain, q.FaceMountain)
	respondJSON(w, http.StatusOK, result)
}
