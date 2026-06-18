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
	"liki/internal/engine/ganzhi"
	"liki/internal/engine/tianwen"
)

// ComputeChart produces a full Chart from solar birth time and gender.
func ComputeChart(st tianwen.SolarTime, gender ganzhi.Gender) Chart {
	return computeChart(tianwen.ComputeBazi(st), st, gender)
}

// ComputeLiuNian computes the year pillar and its interactions with the bazi chart.
func ComputeLiuNian(st tianwen.SolarTime, year int, currentDaYun *DaYunZhu) (*LiuNian, error) {
	return computeLiuNian(tianwen.ComputeBazi(st), year, currentDaYun)
}

// ComputeLiuYue computes the month pillar and its interactions with the bazi chart.
func ComputeLiuYue(st tianwen.SolarTime, year, month int) (*LiuYue, error) {
	return computeLiuYue(tianwen.ComputeBazi(st), year, month)
}

// ComputeLiuRi computes the day pillar and its interactions with the bazi chart.
func ComputeLiuRi(st tianwen.SolarTime, date string, daYunZhu *ganzhi.Zhu, liuNianZhu *ganzhi.Zhu) (*LiuRi, error) {
	return computeLiuRi(tianwen.ComputeBazi(st), date, daYunZhu, liuNianZhu)
}

// ComputeLiuShi computes the hour pillar and its interactions with the bazi chart.
func ComputeLiuShi(st tianwen.SolarTime, date string, hour int) (*LiuShi, error) {
	return computeLiuShi(tianwen.ComputeBazi(st), date, hour)
}

// ComputeXiaoYun computes the minor fortune (小运) pillars.
func ComputeXiaoYun(st tianwen.SolarTime, gender ganzhi.Gender, maxAge int) []XiaoYunZhu {
	return computeXiaoYun(tianwen.ComputeBazi(st), gender, maxAge)
}
