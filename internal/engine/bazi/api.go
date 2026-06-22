// Package bazi provides 八字 computation.
//
// Types
//   Chart, ChartBase,
//   DaYunZhu, DaYun,
//   Bond, XunGong,
//   LiuNian, LiuYue, LiuRi, LiuShi,
//   XiaoYunZhu, XiaoXian,
//   FuYi, TiaoHou, TaiYuanMingGong
//   GanRelation, ZhiRelation,
//   TripleHeFull, GongJia, FuYinFanYin, TaiYuanMingGong
//
// Functions
//   ComputeChart(st, gender) → Chart
//   ComputeBond(a, b ChartBase) → Bond
//   ComputeLiuNian(st, year, cd) → (*LiuNian, error)
//   ComputeLiuYue(st, year, month) → (*LiuYue, error)
//   ComputeLiuRi(st, date, dp, lp) → (*LiuRi, error)
//   ComputeLiuShi(st, date, hour) → (*LiuShi, error)
//   ComputeXiaoYun(st, gender, maxAge) → []XiaoYunZhu
//   ComputeXiaoXian(gender, maxAge) → []XiaoXian
package bazi

import (
	"fmt"

	"liki/internal/engine/ganzhi"
	"liki/internal/engine/tianwen"
)

// ComputeChart produces a full Chart from solar birth time and gender.
func ComputeChart(st tianwen.SolarTime, gender ganzhi.Gender) Chart {
	return computeChart(tianwen.ComputeBazi(st), st, gender)
}

// ComputeLiuNian computes the year pillar and its interactions with the bazi chart.
func ComputeLiuNian(cb ChartBase, year int) (*LiuNian, error) {
	cd := currentDaYunZhu(cb.DaYun)
	return computeLiuNian(cb.ToBazi(), year, cd)
}

// ComputeLiuYue computes the month pillar and its interactions with the bazi chart.
func ComputeLiuYue(cb ChartBase, year, month int) (*LiuYue, error) {
	return computeLiuYue(cb.ToBazi(), year, month)
}

// ComputeLiuRi computes the day pillar and its interactions with the bazi chart.
func ComputeLiuRi(cb ChartBase, year, month, day int) (*LiuRi, error) {
	dz := currentDaYunZhu(cb.DaYun)
	var dzZhu *ganzhi.Zhu
	if dz != nil {
		dzZhu = &ganzhi.Zhu{Gan: dz.Gan, Zhi: dz.Zhi}
	}
	ln, err := ComputeLiuNian(cb, year)
	if err != nil {
		return nil, fmt.Errorf("computeLiuRi: liunian: %w", err)
	}
	var lnZhu *ganzhi.Zhu
	if ln != nil {
		lnZhu = &ganzhi.Zhu{Gan: ln.YearGan, Zhi: ln.YearZhi}
	}
	return computeLiuRi(cb.ToBazi(), year, month, day, dzZhu, lnZhu)
}

// ComputeLiuShi computes the hour pillar and its interactions with the bazi chart.
func ComputeLiuShi(cb ChartBase, year, month, day, hour int) (*LiuShi, error) {
	return computeLiuShi(cb.ToBazi(), year, month, day, hour)
}

// currentDaYunZhu returns the current DaYun Zhu or nil.
func currentDaYunZhu(dy *DaYun) *DaYunZhu {
	if dy == nil || dy.CurrentZhuIndex < 0 || dy.CurrentZhuIndex >= len(dy.Zhu) {
		return nil
	}
	return &dy.Zhu[dy.CurrentZhuIndex]
}

// ComputeXiaoYun computes the minor fortune (小运) pillars.
func ComputeXiaoYun(st tianwen.SolarTime, gender ganzhi.Gender, maxAge int) []XiaoYunZhu {
	return computeXiaoYun(tianwen.ComputeBazi(st), gender, maxAge)
}
