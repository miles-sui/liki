package qimen

import (
	"liki/internal/engine/ganzhi"
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

// computeChart computes a complete奇门盘 with all analyses.
// kind: "shi"/"ri"/"yue"/"nian".
func computeChart(bz ganzhi.Bazi, kind string, y, m, d int) Chart {
	ju := determineJuShu(y, m, d, bz.Ri.Gan, bz.Ri.Zhi)

	var driveZhu ganzhi.Zhu
	switch kind {
	case "ri":
		driveZhu = bz.Ri
	case "yue":
		driveZhu = bz.Yue
	case "nian":
		driveZhu = bz.Nian
	default: // "shi"
		driveZhu = bz.Shi
	}

	p := computePan(ju, driveZhu)
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
