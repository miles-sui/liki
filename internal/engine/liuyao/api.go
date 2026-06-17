// Package liuyao provides 六爻 (纳甲筮法) computation.
//
// Types
//   Chart, Line, YaoType,
//   LiuQin, LiuShou, YongShen,
//   YongShenResult, FuShen,
//   WangShuai, DayRelation, YingQi
//
// Constants
//   YaoType: LaoYin, ShaoYang, ShaoYin, LaoYang
//   LiuQin: QinFumu, QinXiongDi, QinGuanGui, QinQiCai, QinZiSun
//   LiuShou: ShouQingLong, ShouZhuQue, ShouGouChen, ShouTengShe, ShouBaiHu, ShouXuanWu
//   YongShen: YongFumu, YongXiongDi, YongGuanGui, YongQiCai, YongZiSun, YongShiYao
//   WangShuai: WSWang, WSXiang, WSXiu, WSQiu, WSSi
//
// Functions
//   ComputeChart(solarTime, yongShen, fixed) → Chart
package liuyao
