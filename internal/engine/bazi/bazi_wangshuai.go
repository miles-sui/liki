package bazi

import "liki/internal/engine/ganzhi"

// Wangshuai state constants.
const (
	wang  = "旺"
	xiang = "相"
	xiu   = "休"
	qiu   = "囚"
	si    = "死"
)

// monthWangShuai returns the five-element prosperity state for a given element
// in a given solar month (branch 1-12). Maps the standard 旺相休囚死 table.
//
// Rule per month: 当令者旺 / 我生者相 / 生我者休 / 克我者囚 / 我克者死
func monthWangShuai(elem ganzhi.Wuxing, monthBranch ganzhi.Zhi) string {
	m := int(monthBranch)
	if m < 1 || m > 12 {
		return ""
	}

	mwx := ganzhi.ZhiWuxing(monthBranch)
	if elem == mwx {
		return wang
	}
	if ganzhi.Sheng(elem, mwx) {
		return xiu
	}
	if ganzhi.Sheng(mwx, elem) {
		return xiang
	}
	if ganzhi.Ke(elem, mwx) {
		return qiu
	}
	return si
}
