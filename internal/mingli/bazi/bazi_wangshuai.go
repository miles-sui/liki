package bazi

// MonthWangShuai returns the five-element prosperity state for a given element
// in a given solar month (branch 1-12). Maps the standard 旺相休囚死 table.
//
// Rule per month: 当令者旺 / 我生者相 / 生我者休 / 克我者囚 / 我克者死
func MonthWangShuai(elem Element, monthBranch Branch) string {
	m := int(monthBranch)
	if m < 1 || m > 12 {
		return ""
	}

	wang := BranchElement(monthBranch)
	if elem == wang {
		return "旺"
	}
	if Sheng(elem, wang) {
		return "休"
	}
	if Sheng(wang, elem) {
		return "相"
	}
	if Ke(elem, wang) {
		return "囚"
	}
	return "死"
}
