package qimen

import (

	"liki/internal/engine/ganzhi"
	"liki/internal/engine/tianwen"
)

// Chart bundles a complete奇门盘 with all analysis layers.
type Chart struct {
	Pan              pan               `json:"pan"`
	StemInteractions [9]StemInteraction `json:"stem_interactions"`
	DoorInteractions [9]DoorInteraction `json:"door_interactions"`
	StarInteractions [9]StarInteraction `json:"star_interactions"`
	WangShuai        [9]WangShuai       `json:"wang_shuai"`
	MenPo            []PalaceIndex      `json:"men_po"`
	MenZhi           []PalaceIndex      `json:"men_zhi"`
	Patterns         []Pattern          `json:"patterns"`
	YingQi           YingQi             `json:"ying_qi"`
}

// ComputeChart computes a complete奇门盘 with all analyses.
// kind: "shi"/"ri"/"yue"/"nian".
func ComputeChart(kind string, st tianwen.SolarTime) Chart {
	bz := tianwen.ComputeBazi(st)
	t := st.Time()
	y, m, d := t.Date()
	ju := determineJuShu(y, int(m), d, int(bz.Ri.Gan), int(bz.Ri.Zhi))

	var driveGan ganzhi.Gan
	var driveZhi ganzhi.Zhi
	switch kind {
	case "ri":
		driveGan, driveZhi = bz.Ri.Gan, bz.Ri.Zhi
	case "yue":
		driveGan, driveZhi = bz.Yue.Gan, bz.Yue.Zhi
	case "nian":
		driveGan, driveZhi = bz.Nian.Gan, bz.Nian.Zhi
	default: // "shi"
		driveGan, driveZhi = bz.Shi.Gan, bz.Shi.Zhi
	}

	p := computePan(ju, driveGan, driveZhi)
	return Chart{
		Pan:              p,
		StemInteractions: computeStemInteractions(p),
		DoorInteractions: computeDoorInteractions(p),
		StarInteractions: computeStarInteractions(p),
		WangShuai:        computeWangShuai(p),
		MenPo:            findMenPo(p),
		MenZhi:           findMenZhi(p),
		Patterns:         findPatterns(p),
		YingQi:           computeYingQi(p),
	}
}
