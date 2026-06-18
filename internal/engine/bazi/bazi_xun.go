package bazi

import "liki/internal/engine/ganzhi"

// xunIndex returns the xun index (0-5) for a day pillar.
func xunIndex(riZhu ganzhi.Zhu) int {
	return ganzhi.SixtyCycleName(riZhu.Gan, riZhu.Zhi) / 10
}
