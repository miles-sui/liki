package handler

import (
	"errors"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"liki/internal/engine/liuyao"
)


type liuyaoRequest struct {
	Birth    timePoint `json:"birth"`
	YongShen string     `json:"yong_shen"`
	Fixed    [6]int     `json:"fixed"`
}

func (r liuyaoRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Birth, validation.By(validateTimePoint)),
		validation.Field(&r.YongShen, validation.By(validateYongShen)),
		validation.Field(&r.Fixed, validation.By(validateFixedYaos)),
	)
}

var validYaoValues = map[int]bool{0: true, 6: true, 7: true, 8: true, 9: true}

func validateFixedYaos(value any) error {
	yaos, ok := value.([6]int)
	if !ok {
		return nil
	}
	for _, v := range yaos {
		if !validYaoValues[v] {
			return errors.New("fixed yao values must be 0 (auto), 6, 7, 8, or 9")
		}
	}
	return nil
}

func validateYongShen(value any) error {
	s, ok := value.(string)
	if !ok {
		return nil
	}
	if s == "" {
		return nil
	}
	if _, err := liuyao.ParseYongShen(s); err != nil {
		return errors.New("yong_shen must be one of: 父母/兄弟/官鬼/妻财/子孙/世爻")
	}
	return nil
}

func handleLiuyaoChart(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeAndValidate[liuyaoRequest](w, r)
	if !ok {
		return
	}
	if req.YongShen == "" {
		req.YongShen = "世爻"
	}
	ys, err := liuyao.ParseYongShen(req.YongShen)
	if err != nil {
		respondInvalidRequest(w, err.Error())
		return
	}
	ts, err := parseTimeset(req.Birth)
	if err != nil {
		respondInvalidRequest(w, err.Error())
		return
	}
	chart := liuyao.ComputeChart(ts.Solar, ys, req.Fixed)
	respondJSON(w, http.StatusOK, chart)
}
