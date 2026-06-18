// Package qimen provides 奇门遁甲 computation.
//
// Types
//   Chart, Palace
//   PalaceIndex, StarIndex, DoorIndex, SpiritIndex
//   StemInteraction, DoorInteraction, StarInteraction
//   WangShuai, Pattern, YingQi
//
// Constants
//   PalaceKan .. PalaceLi  (九宫)
//   StarTianPeng .. StarTianYing  (九星)
//   DoorXiu .. DoorKai  (八门)
//   SpiritZhiFu .. SpiritJiuTian  (八神)
//
// Functions
//   ComputeChart(st SolarTime, kind string) → Chart
package qimen

import (
	"liki/internal/engine/tianwen"
)

// ComputeChart computes a complete奇门盘 with all analyses.
// kind: "shi"/"ri"/"yue"/"nian".
func ComputeChart(st tianwen.SolarTime, kind string) Chart {
	bz := tianwen.ComputeBazi(st)
	t := st.Time()
	y, m, d := t.Date()
	return computeChart(bz, kind, y, int(m), d)
}
