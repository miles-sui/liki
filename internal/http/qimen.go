package handler

import (
	"errors"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"liki/internal/engine/qimen"
)

type qimenRequest struct {
	Birth timePoint `json:"birth"`
	Kind  string     `json:"kind"`
}

func (r qimenRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Birth, validation.By(validateTimePoint)),
		validation.Field(&r.Kind, validation.By(validateQimenKind)),
	)
}

func validateQimenKind(value any) error {
	s, ok := value.(string)
	if !ok {
		return nil
	}
	if s == "" || s == "shi" || s == "ri" || s == "yue" || s == "nian" {
		return nil
	}
	return errors.New("kind must be shi/ri/yue/nian")
}

func handleQimenPan(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeAndValidate[qimenRequest](w, r)
	if !ok {
		return
	}
	if req.Kind == "" {
		req.Kind = "shi"
	}
	ts, ok := timesetOrRespond(w, req.Birth)
	if !ok {
		return
	}
	chart := qimen.ComputeChart(ts.Solar, req.Kind)
	respondJSON(w, http.StatusOK, chart)
}
