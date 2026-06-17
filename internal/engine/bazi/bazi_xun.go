package bazi

import "liki/internal/engine/ganzhi"

// xunIndex returns the xun index (0-5) for a day pillar.
func xunIndex(dayPillar ganzhi.Zhu) int {
	return ganzhi.SixtyCycleName(dayPillar.Gan, dayPillar.Zhi) / 10
}
