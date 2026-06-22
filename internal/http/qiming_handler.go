package handler

import (
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"liki/internal/engine/qiming"
)

var wuxingSet = []any{"木", "火", "土", "金", "水"}

type wugeRequest struct {
	Surname  string   `json:"surname"`
	YongShen string   `json:"yong_shen"`
	XiShen   []string `json:"xi_shen"`
}

func (r wugeRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Surname, validation.Required, validation.RuneLength(1, 2)),
		validation.Field(&r.YongShen, validation.Required, validation.In(wuxingSet...)),
		validation.Field(&r.XiShen, validation.Each(validation.In(wuxingSet...))),
	)
}

func handleWuge(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeAndValidate[wugeRequest](w, r)
	if !ok {
		return
	}
	result, err := qiming.PrepareWuGe(req.Surname, req.YongShen, req.XiShen)
	if err != nil {
		respondInvalidRequest(w, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, result)
}

type composeRequest struct {
	Surname   string                    `json:"surname"`
	Combos    []qiming.StrokeCombo      `json:"combos"`
	YongChars map[int][]qiming.CharLite `json:"yong_chars"`
	XiChars   map[int][]qiming.CharLite `json:"xi_chars"`
}

func (r composeRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Surname, validation.Required, validation.RuneLength(1, 2)),
	)
}

func handleCompose(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeAndValidate[composeRequest](w, r)
	if !ok {
		return
	}
	names := qiming.ComposeNames(req.Surname, req.Combos, req.YongChars, req.XiChars)
	if names == nil {
		names = []string{}
	}
	respondJSON(w, http.StatusOK, names)
}

type detailRequest struct {
	Surname string   `json:"surname"`
	Names   []string `json:"names"`
}

func (r detailRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Surname, validation.Required, validation.RuneLength(1, 2)),
		validation.Field(&r.Names, validation.Required, validation.Length(1, 50)),
	)
}

func handleDetail(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeAndValidate[detailRequest](w, r)
	if !ok {
		return
	}
	results, err := qiming.DetailNames(req.Surname, req.Names)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, results)
}

type evaluateRequest struct {
	Surname   string `json:"surname"`
	GivenName string `json:"given_name"`
	YongShen  string `json:"yong_shen"`
}

func (r evaluateRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Surname, validation.Required, validation.RuneLength(1, 2)),
		validation.Field(&r.GivenName, validation.Required, validation.RuneLength(1, 2)),
		validation.Field(&r.YongShen, validation.Required, validation.In(wuxingSet...)),
	)
}

func handleEvaluate(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeAndValidate[evaluateRequest](w, r)
	if !ok {
		return
	}
	result, err := qiming.EvaluateName(req.Surname, req.GivenName, req.YongShen)
	if err != nil {
		respondInvalidRequest(w, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, result)
}
