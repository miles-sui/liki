// Package bazi provides 八字 computation.
//
// Types
//   Chart, ChartBase,
//   DaYunPillar, DaYun,
//   Bond, XunGong,
//   LiuNian, LiuYue, LiuRi, LiuShi,
//   XiaoYunPillar, XiaoXian,
//   FuYi, TiaoHou, TaiYuanMingGong
//   GanRelation, ZhiRelation,
//   TripleHeFull, GongJia, FuYinFanYin, TaiYuanMingGong
//
// Functions
//   ComputeChart(st, gender) → Chart
//   ComputeBond(a, b ChartBase) → Bond
//   ComputeLiuNian(year, dm, bz, cd) → *LiuNian
//   ComputeLiuYue(year, month, dm, bz) → *LiuYue
//   ComputeLiuRi(date, dm, bz, dp, lp) → *LiuRi
//   ComputeLiuShi(date, hour, dm, bz) → *LiuShi
//   ComputeXiaoYun(gender, dm, maxAge) → []XiaoYunPillar
//   ComputeXiaoXian(gender, maxAge) → []XiaoXian
package bazi
