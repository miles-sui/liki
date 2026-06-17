package handler

import (
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"liki/internal/engine/qiming"
)

var wuxingSet = []any{"木", "火", "土", "金", "水"}

type wugeParams struct {
	Surname  string   `json:"surname"`
	YongShen string   `json:"yong_shen"`
	XiShen   []string `json:"xi_shen"`
}

type composeParams struct {
	Surname   string                     `json:"surname"`
	Combos    []qiming.StrokeCombo `json:"combos"`
	YongChars map[int][]string           `json:"yong_chars"`
	XiChars   map[int][]string           `json:"xi_chars"`
}

type detailParams struct {
	Surname string   `json:"surname"`
	Names   []string `json:"names"`
}

type evalParams struct {
	Surname   string `json:"surname"`
	GivenName string `json:"given_name"`
	YongShen  string `json:"yong_shen"`
}

func handleWuge(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeJSON[wugeParams](w, r)
	if !ok {
		return
	}
	if err := validation.ValidateStruct(&req,
		validation.Field(&req.Surname, validation.Required, validation.RuneLength(1, 2)),
		validation.Field(&req.YongShen, validation.Required, validation.In(wuxingSet...)),
		validation.Field(&req.XiShen, validation.Each(validation.In(wuxingSet...))),
	); err != nil {
		respondValidationError(w, err)
		return
	}

	result, err := qiming.PrepareWuGe(req.Surname, req.YongShen, req.XiShen)
	if err != nil {
		respondInvalidRequest(w, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func handleCompose(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeJSON[composeParams](w, r)
	if !ok {
		return
	}
	if err := validation.ValidateStruct(&req,
		validation.Field(&req.Surname, validation.Required, validation.RuneLength(1, 2)),
	); err != nil {
		respondValidationError(w, err)
		return
	}

	names := qiming.ComposeNames(req.Surname, req.Combos, req.YongChars, req.XiChars)
	if names == nil {
		names = []string{}
	}
	respondJSON(w, http.StatusOK, names)
}

func handleDetail(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeJSON[detailParams](w, r)
	if !ok {
		return
	}
	if err := validation.ValidateStruct(&req,
		validation.Field(&req.Surname, validation.Required, validation.RuneLength(1, 2)),
		validation.Field(&req.Names, validation.Required, validation.Length(1, 50)),
	); err != nil {
		respondValidationError(w, err)
		return
	}

	results := qiming.DetailNames(req.Surname, req.Names)
	if results == nil {
		results = []qiming.NameCandidate{}
	}
	respondJSON(w, http.StatusOK, map[string]any{
		"results": results,
	})
}

func handleEvaluate(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeJSON[evalParams](w, r)
	if !ok {
		return
	}
	if err := validation.ValidateStruct(&req,
		validation.Field(&req.Surname, validation.Required, validation.RuneLength(1, 2)),
		validation.Field(&req.GivenName, validation.Required, validation.RuneLength(1, 2)),
		validation.Field(&req.YongShen, validation.Required, validation.In(wuxingSet...)),
	); err != nil {
		respondValidationError(w, err)
		return
	}

	result, err := qiming.EvaluateName(req.Surname, req.GivenName, req.YongShen)
	if err != nil {
		respondInvalidRequest(w, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, result)
}
