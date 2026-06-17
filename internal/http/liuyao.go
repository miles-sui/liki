package handler

import (
	"net/http"

	"liki/internal/engine/liuyao"
	"liki/internal/engine/tianwen"
)

var yongShenMap = map[string]liuyao.YongShen{
	"父母": liuyao.YongFumu,
	"兄弟": liuyao.YongXiongDi,
	"官鬼": liuyao.YongGuanGui,
	"妻财": liuyao.YongQiCai,
	"子孙": liuyao.YongZiSun,
	"世爻": liuyao.YongShiYao,
}

type liuyaoRequest struct {
	SolarTime tianwen.SolarTime `json:"solar_time"`
	YongShen  string            `json:"yong_shen"`
	Fixed     [6]int            `json:"fixed,omitempty"`
}

func handleLiuyaoChart(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeJSON[liuyaoRequest](w, r)
	if !ok {
		return
	}
	if req.YongShen == "" {
		req.YongShen = "世爻"
	}
	ys, ok := yongShenMap[req.YongShen]
	if !ok {
		respondInvalidRequest(w, "yong_shen must be one of: 父母/兄弟/官鬼/妻财/子孙/世爻")
		return
	}

	chart := liuyao.ComputeChart(req.SolarTime, ys, req.Fixed)
	respondJSON(w, http.StatusOK, chart)
}
