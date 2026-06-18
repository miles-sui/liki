package handler

import (
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"liki/internal/engine/qiming"
)

var wuxingSet = []any{"木", "火", "土", "金", "水"}

func handleWuge(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeWith(w, r, func(p struct {
		Surname  string   `json:"surname"`
		YongShen string   `json:"yong_shen"`
		XiShen   []string `json:"xi_shen"`
	}) error {
		return validation.ValidateStruct(&p,
			validation.Field(&p.Surname, validation.Required, validation.RuneLength(1, 2)),
			validation.Field(&p.YongShen, validation.Required, validation.In(wuxingSet...)),
			validation.Field(&p.XiShen, validation.Each(validation.In(wuxingSet...))),
		)
	})
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

func handleCompose(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeWith(w, r, func(p struct {
		Surname   string                     `json:"surname"`
		Combos    []qiming.StrokeCombo       `json:"combos"`
		YongChars map[int][]qiming.CharLite  `json:"yong_chars"`
		XiChars   map[int][]qiming.CharLite  `json:"xi_chars"`
	}) error {
		return validation.ValidateStruct(&p,
			validation.Field(&p.Surname, validation.Required, validation.RuneLength(1, 2)),
		)
	})
	if !ok {
		return
	}
	names := qiming.ComposeNames(req.Surname, req.Combos, req.YongChars, req.XiChars)
	if names == nil {
		names = []string{}
	}
	respondJSON(w, http.StatusOK, names)
}

func handleDetail(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeWith(w, r, func(p struct {
		Surname string   `json:"surname"`
		Names   []string `json:"names"`
	}) error {
		return validation.ValidateStruct(&p,
			validation.Field(&p.Surname, validation.Required, validation.RuneLength(1, 2)),
			validation.Field(&p.Names, validation.Required, validation.Length(1, 50)),
		)
	})
	if !ok {
		return
	}
	results, err := qiming.DetailNames(req.Surname, req.Names)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid_surname", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, results)
}

func handleEvaluate(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeWith(w, r, func(p struct {
		Surname   string `json:"surname"`
		GivenName string `json:"given_name"`
		YongShen  string `json:"yong_shen"`
	}) error {
		return validation.ValidateStruct(&p,
			validation.Field(&p.Surname, validation.Required, validation.RuneLength(1, 2)),
			validation.Field(&p.GivenName, validation.Required, validation.RuneLength(1, 2)),
			validation.Field(&p.YongShen, validation.Required, validation.In(wuxingSet...)),
		)
	})
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
