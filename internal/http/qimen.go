package handler

import (
	"net/http"
	"time"

	"liki/internal/engine/qimen"
	"liki/internal/engine/tianwen"
)

type qimenRequest struct {
	SolarTime time.Time `json:"solar_time"`
	Kind      string    `json:"kind"`
}

func handleQimenPan(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeJSON[qimenRequest](w, r)
	if !ok {
		return
	}
	if req.Kind == "" {
		req.Kind = "shi"
	}
	if req.Kind != "shi" && req.Kind != "ri" && req.Kind != "yue" && req.Kind != "nian" {
		respondInvalidRequest(w, "kind must be shi/ri/yue/nian")
		return
	}

	chart := qimen.ComputeChart(req.Kind, tianwen.SolarTime(req.SolarTime))
	respondJSON(w, http.StatusOK, chart)
}
