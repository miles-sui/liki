// Package ziwei provides紫微斗数 computation.
//
// Types
//   Chart, Bond, DaXianStep, LiuNian, LiuYue, LiuRi
//
// Functions
//   ComputeChart(birthYear, lunarMonth, lunarDay, hourZhi, yearGan, yearZhi, gender) → Chart
//   ComputeDaXian(chart) → []DaXianStep
//   ComputeLiuNian(liuYear, chart) → LiuNian
//   ComputeLiuYue(liuYear, lunarMonth, chart) → LiuYue
//   ComputeLiuRi(liuYear, lunarMonth, lunarDay, chart) → LiuRi
//   ComputeBond(a, b Chart) → Bond
package ziwei
