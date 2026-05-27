package bazi

import (
	"fmt"

	"github.com/25types/25types/internal/ganzhi"
)

// tiaohouKey is the internal compound key for the lookup table.
type tiaohouKey struct {
	stem   int
	branch int
}

// lookupTiaohou maps (dayMaster, monthBranch) → {primary, secondary}.
// Based on the standard 《穷通宝鉴》 reference for all 10 stems × 12 months.
var lookupTiaohou = map[tiaohouKey]struct{ primary, secondary string }{
	// 甲木
	{1, 1}: {"丙", "癸"}, {1, 2}: {"丙", "癸"}, {1, 3}: {"庚", "壬"},
	{1, 4}: {"癸", "丁"}, {1, 5}: {"癸", "庚"}, {1, 6}: {"癸", "庚"},
	{1, 7}: {"庚", "丁"}, {1, 8}: {"庚", "丁"}, {1, 9}: {"庚", "壬"},
	{1, 10}: {"庚", "丁"}, {1, 11}: {"丁", "庚"}, {1, 12}: {"丁", "丙"},
	// 乙木
	{2, 1}: {"丙", "癸"}, {2, 2}: {"丙", "癸"}, {2, 3}: {"癸", "丙"},
	{2, 4}: {"癸", "辛"}, {2, 5}: {"癸", "丙"}, {2, 6}: {"癸", "丙"},
	{2, 7}: {"丙", "癸"}, {2, 8}: {"丙", "癸"}, {2, 9}: {"癸", "辛"},
	{2, 10}: {"丙", "戊"}, {2, 11}: {"丙", "戊"}, {2, 12}: {"丙", "戊"},
	// 丙火
	{3, 1}: {"壬", "庚"}, {3, 2}: {"壬", "己"}, {3, 3}: {"壬", "甲"},
	{3, 4}: {"壬", "庚"}, {3, 5}: {"壬", "庚"}, {3, 6}: {"壬", "庚"},
	{3, 7}: {"壬", "戊"}, {3, 8}: {"壬", "癸"}, {3, 9}: {"壬", "甲"},
	{3, 10}: {"甲", "戊"}, {3, 11}: {"壬", "戊"}, {3, 12}: {"壬", "甲"},
	// 丁火
	{4, 1}: {"甲", "庚"}, {4, 2}: {"甲", "庚"}, {4, 3}: {"甲", "庚"},
	{4, 4}: {"甲", "庚"}, {4, 5}: {"壬", "庚"}, {4, 6}: {"甲", "壬"},
	{4, 7}: {"甲", "庚"}, {4, 8}: {"甲", "庚"}, {4, 9}: {"甲", "庚"},
	{4, 10}: {"甲", "庚"}, {4, 11}: {"甲", "庚"}, {4, 12}: {"甲", "庚"},
	// 戊土
	{5, 1}: {"丙", "甲"}, {5, 2}: {"丙", "甲"}, {5, 3}: {"甲", "丙"},
	{5, 4}: {"甲", "丙"}, {5, 5}: {"壬", "甲"}, {5, 6}: {"壬", "甲"},
	{5, 7}: {"丙", "甲"}, {5, 8}: {"丙", "甲"}, {5, 9}: {"甲", "壬"},
	{5, 10}: {"甲", "丙"}, {5, 11}: {"丙", "甲"}, {5, 12}: {"丙", "甲"},
	// 己土
	{6, 1}: {"丙", "甲"}, {6, 2}: {"甲", "丙"}, {6, 3}: {"丙", "甲"},
	{6, 4}: {"癸", "丙"}, {6, 5}: {"癸", "丙"}, {6, 6}: {"癸", "丙"},
	{6, 7}: {"丙", "癸"}, {6, 8}: {"丙", "癸"}, {6, 9}: {"丙", "甲"},
	{6, 10}: {"丙", "甲"}, {6, 11}: {"丙", "甲"}, {6, 12}: {"丙", "甲"},
	// 庚金
	{7, 1}: {"戊", "甲"}, {7, 2}: {"丁", "丙"}, {7, 3}: {"甲", "丁"},
	{7, 4}: {"壬", "戊"}, {7, 5}: {"壬", "己"}, {7, 6}: {"壬", "己"},
	{7, 7}: {"丁", "甲"}, {7, 8}: {"丁", "丙"}, {7, 9}: {"甲", "壬"},
	{7, 10}: {"丁", "丙"}, {7, 11}: {"丁", "丙"}, {7, 12}: {"丁", "丙"},
	// 辛金
	{8, 1}: {"己", "壬"}, {8, 2}: {"壬", "甲"}, {8, 3}: {"壬", "甲"},
	{8, 4}: {"壬", "甲"}, {8, 5}: {"壬", "己"}, {8, 6}: {"壬", "甲"},
	{8, 7}: {"壬", "甲"}, {8, 8}: {"壬", "甲"}, {8, 9}: {"壬", "甲"},
	{8, 10}: {"壬", "丙"}, {8, 11}: {"丙", "戊"}, {8, 12}: {"丙", "壬"},
	// 壬水
	{9, 1}: {"庚", "戊"}, {9, 2}: {"戊", "辛"}, {9, 3}: {"甲", "庚"},
	{9, 4}: {"壬", "辛"}, {9, 5}: {"癸", "庚"}, {9, 6}: {"辛", "甲"},
	{9, 7}: {"戊", "丁"}, {9, 8}: {"甲", "庚"}, {9, 9}: {"甲", "丙"},
	{9, 10}: {"戊", "庚"}, {9, 11}: {"戊", "丙"}, {9, 12}: {"丙", "丁"},
	// 癸水
	{10, 1}: {"辛", "丙"}, {10, 2}: {"庚", "辛"}, {10, 3}: {"丙", "辛"},
	{10, 4}: {"辛", "壬"}, {10, 5}: {"庚", "壬"}, {10, 6}: {"庚", "辛"},
	{10, 7}: {"丁", "辛"}, {10, 8}: {"辛", "丙"}, {10, 9}: {"辛", "甲"},
	{10, 10}: {"庚", "辛"}, {10, 11}: {"丙", "辛"}, {10, 12}: {"丙", "辛"},
}

// stemToElement maps a Chinese stem name to its wuxing element.
var stemToElement = map[string]Element{
	"甲": ElemWood, "乙": ElemWood,
	"丙": ElemFire, "丁": ElemFire,
	"戊": ElemEarth, "己": ElemEarth,
	"庚": ElemMetal, "辛": ElemMetal,
	"壬": ElemWater, "癸": ElemWater,
}

func monthBranchSeason(b Branch) string {
	switch b {
	case 1, 2, 3:
		return "春"
	case 4, 5, 6:
		return "夏"
	case 7, 8, 9:
		return "秋"
	case 10, 11, 12:
		return "冬"
	}
	return ""
}

// ComputeTiaohou returns the 穷通宝鉴 climate-adjustment result for a given
// day-master and month-branch. Returns (TiaoHouResult, true) on match, or
// (zero, false) if no entry exists.
func ComputeTiaohou(dayMaster Stem, monthBranch Branch) (TiaoHouResult, bool) {
	e, ok := lookupTiaohou[tiaohouKey{int(dayMaster), int(monthBranch)}]
	if !ok {
		return TiaoHouResult{}, false
	}

	yongElem := stemToElement[e.primary]
	xiElem := stemToElement[e.secondary]
	jiElem := ElementThatControls(yongElem)

	season := monthBranchSeason(monthBranch)
	detail := fmt.Sprintf("%s月%s，用%s调候，%s辅之",
		MonthBranchNameString(monthBranch),
		DayMasterNameString(dayMaster),
		e.primary,
		e.secondary,
	)

	return TiaoHouResult{
		Season: season,
		Yong:   yongElem.String(),
		Xi:     xiElem.String(),
		Ji:     jiElem.String(),
		Detail: detail,
	}, true
}

// Ensure ganzhi import is used (needed by callers that reference Stem/Branch types).
var _ = ganzhi.StemName
