// Package liuyao provides 六爻 (纳甲筮法) computation.
//
// Types
//   Chart, Line, YaoType,
//   LiuQin, LiuShou, YongShen,
//   YongShenResult, FuShen,
//   ganzhi.WangShuai, DayRelation, YingQi
//
// Constants
//   YaoType: LaoYin, ShaoYang, ShaoYin, LaoYang
//   LiuQin: QinFumu, QinXiongDi, QinGuanGui, QinQiCai, QinZiSun
//   LiuShou: ShouQingLong, ShouZhuQue, ShouGouChen, ShouTengShe, ShouBaiHu, ShouXuanWu
//   YongShen: YongFumu, YongXiongDi, YongGuanGui, YongQiCai, YongZiSun, YongShiYao
//   ganzhi: WSWang, WSXiang, WSXiu, WSQiu, WSSi
//
// Functions
//   ComputeChart(st SolarTime, yongShen YongShen, fixed [6]int) → Chart
package liuyao

import "liki/internal/engine/tianwen"

// ComputeChart computes a complete 六爻 chart from solar time, question type, and optional fixed yaos.
func ComputeChart(st tianwen.SolarTime, yongShen YongShen, fixed [6]int) Chart {
	return computeChart(tianwen.ComputeBazi(st), yongShen, fixed)
}
