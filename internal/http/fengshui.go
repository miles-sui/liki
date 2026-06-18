package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-ozzo/ozzo-validation/v4"

	"liki/internal/engine/bazhai"
	"liki/internal/engine/ganzhi"
	"liki/internal/engine/xuankong"
)

func xuankongSanYuan(w http.ResponseWriter, r *http.Request) {
	year := 2024
	if ys := r.URL.Query().Get("year"); ys != "" {
		if y, err := strconv.Atoi(ys); err == nil {
			year = y
		}
	}
	result := struct {
		Current xuankong.SanYuanYun `json:"current"`
	}{Current: xuankong.ComputeSanYuanYun(year)}
	respondJSON(w, http.StatusOK, result)
}

type mingGuaRequest struct {
	Gender    ganzhi.Gender `json:"gender"`
	BirthYear int           `json:"birth_year"`
}

func (r mingGuaRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Gender, validation.Required, validation.In(validGenders...)),
		validation.Field(&r.BirthYear, validation.Required, validation.Min(1900), validation.Max(2100)),
	)
}

func bazhaiMingGua(w http.ResponseWriter, r *http.Request) {
	q, ok := decodeAndValidate[mingGuaRequest](w, r)
	if !ok {
		return
	}
	result := bazhai.ComputeMingGua(q.Gender, q.BirthYear)
	respondJSON(w, http.StatusOK, result)
}

type bazhaiChartRequest struct {
	Birth  timePoint     `json:"birth"`
	Gender ganzhi.Gender `json:"gender"`
}

func (r bazhaiChartRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Birth, validation.By(validateTimePoint)),
		validation.Field(&r.Gender, validation.Required, validation.In(validGenders...)),
	)
}

func bazhaiChart(w http.ResponseWriter, r *http.Request) {
	q, ok := decodeAndValidate[bazhaiChartRequest](w, r)
	if !ok {
		return
	}
	ts, err := q.Birth.Timeset()
	if err != nil {
		respondInvalidRequest(w, err.Error())
		return
	}
	result := bazhai.ComputeChart(ts.Solar, q.Gender)
	respondJSON(w, http.StatusOK, result)
}

type xuankongChartRequest struct {
	Birth        timePoint `json:"birth"`
	SitMountain  *int       `json:"sit_mountain"`
	FaceMountain *int       `json:"face_mountain"`
}

func (r xuankongChartRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Birth, validation.By(validateTimePoint)),
		validation.Field(&r.SitMountain, validation.By(func(value any) error {
			if value.(*int) == nil {
				return errors.New("required")
			}
			return nil
		}), validation.Min(0), validation.Max(23)),
		validation.Field(&r.FaceMountain, validation.By(func(value any) error {
			if value.(*int) == nil {
				return errors.New("required")
			}
			return nil
		}), validation.Min(0), validation.Max(23)),
	)
}

func xuankongChart(w http.ResponseWriter, r *http.Request) {
	q, ok := decodeAndValidate[xuankongChartRequest](w, r)
	if !ok {
		return
	}
	ts, err := q.Birth.Timeset()
	if err != nil {
		respondInvalidRequest(w, err.Error())
		return
	}
	result := xuankong.ComputeChart(ts.Solar, *q.SitMountain, *q.FaceMountain)
	respondJSON(w, http.StatusOK, result)
}
