package bazi

import (
	"fmt"

	"liki/internal/engine/ganzhi"
)

// tiaohouKey is the internal compound key for the lookup table.
type tiaohouKey struct {
	stem   int
	branch int
}

// lookupTiaohou maps (dayMaster, monthBranch) → {primary, secondary}.
// Based on the standard 《穷通宝鉴》 reference for all 10 stems × 12 months.
var lookupTiaohou = map[tiaohouKey]struct{ primary, secondary ganzhi.Gan }{
	// 甲木
	{1, 1}: {ganzhi.GanBing, ganzhi.GanGui}, {1, 2}: {ganzhi.GanBing, ganzhi.GanGui}, {1, 3}: {ganzhi.GanGeng, ganzhi.GanRen},
	{1, 4}: {ganzhi.GanGui, ganzhi.GanDing}, {1, 5}: {ganzhi.GanGui, ganzhi.GanGeng}, {1, 6}: {ganzhi.GanGui, ganzhi.GanGeng},
	{1, 7}: {ganzhi.GanGeng, ganzhi.GanDing}, {1, 8}: {ganzhi.GanGeng, ganzhi.GanDing}, {1, 9}: {ganzhi.GanGeng, ganzhi.GanRen},
	{1, 10}: {ganzhi.GanGeng, ganzhi.GanDing}, {1, 11}: {ganzhi.GanDing, ganzhi.GanGeng}, {1, 12}: {ganzhi.GanDing, ganzhi.GanBing},
	// 乙木
	{2, 1}: {ganzhi.GanBing, ganzhi.GanGui}, {2, 2}: {ganzhi.GanBing, ganzhi.GanGui}, {2, 3}: {ganzhi.GanGui, ganzhi.GanBing},
	{2, 4}: {ganzhi.GanGui, ganzhi.GanXin}, {2, 5}: {ganzhi.GanGui, ganzhi.GanBing}, {2, 6}: {ganzhi.GanGui, ganzhi.GanBing},
	{2, 7}: {ganzhi.GanBing, ganzhi.GanGui}, {2, 8}: {ganzhi.GanBing, ganzhi.GanGui}, {2, 9}: {ganzhi.GanGui, ganzhi.GanXin},
	{2, 10}: {ganzhi.GanBing, ganzhi.GanWu}, {2, 11}: {ganzhi.GanBing, ganzhi.GanWu}, {2, 12}: {ganzhi.GanBing, ganzhi.GanWu},
	// 丙火
	{3, 1}: {ganzhi.GanRen, ganzhi.GanGeng}, {3, 2}: {ganzhi.GanRen, ganzhi.GanJi}, {3, 3}: {ganzhi.GanRen, ganzhi.GanJia},
	{3, 4}: {ganzhi.GanRen, ganzhi.GanGeng}, {3, 5}: {ganzhi.GanRen, ganzhi.GanGeng}, {3, 6}: {ganzhi.GanRen, ganzhi.GanGeng},
	{3, 7}: {ganzhi.GanRen, ganzhi.GanWu}, {3, 8}: {ganzhi.GanRen, ganzhi.GanGui}, {3, 9}: {ganzhi.GanRen, ganzhi.GanJia},
	{3, 10}: {ganzhi.GanJia, ganzhi.GanWu}, {3, 11}: {ganzhi.GanRen, ganzhi.GanWu}, {3, 12}: {ganzhi.GanRen, ganzhi.GanJia},
	// 丁火
	{4, 1}: {ganzhi.GanJia, ganzhi.GanGeng}, {4, 2}: {ganzhi.GanJia, ganzhi.GanGeng}, {4, 3}: {ganzhi.GanJia, ganzhi.GanGeng},
	{4, 4}: {ganzhi.GanJia, ganzhi.GanGeng}, {4, 5}: {ganzhi.GanRen, ganzhi.GanGeng}, {4, 6}: {ganzhi.GanJia, ganzhi.GanRen},
	{4, 7}: {ganzhi.GanJia, ganzhi.GanGeng}, {4, 8}: {ganzhi.GanJia, ganzhi.GanGeng}, {4, 9}: {ganzhi.GanJia, ganzhi.GanGeng},
	{4, 10}: {ganzhi.GanJia, ganzhi.GanGeng}, {4, 11}: {ganzhi.GanJia, ganzhi.GanGeng}, {4, 12}: {ganzhi.GanJia, ganzhi.GanGeng},
	// 戊土
	{5, 1}: {ganzhi.GanBing, ganzhi.GanJia}, {5, 2}: {ganzhi.GanBing, ganzhi.GanJia}, {5, 3}: {ganzhi.GanJia, ganzhi.GanBing},
	{5, 4}: {ganzhi.GanJia, ganzhi.GanBing}, {5, 5}: {ganzhi.GanRen, ganzhi.GanJia}, {5, 6}: {ganzhi.GanRen, ganzhi.GanJia},
	{5, 7}: {ganzhi.GanBing, ganzhi.GanJia}, {5, 8}: {ganzhi.GanBing, ganzhi.GanJia}, {5, 9}: {ganzhi.GanJia, ganzhi.GanRen},
	{5, 10}: {ganzhi.GanJia, ganzhi.GanBing}, {5, 11}: {ganzhi.GanBing, ganzhi.GanJia}, {5, 12}: {ganzhi.GanBing, ganzhi.GanJia},
	// 己土
	{6, 1}: {ganzhi.GanBing, ganzhi.GanJia}, {6, 2}: {ganzhi.GanJia, ganzhi.GanBing}, {6, 3}: {ganzhi.GanBing, ganzhi.GanJia},
	{6, 4}: {ganzhi.GanGui, ganzhi.GanBing}, {6, 5}: {ganzhi.GanGui, ganzhi.GanBing}, {6, 6}: {ganzhi.GanGui, ganzhi.GanBing},
	{6, 7}: {ganzhi.GanBing, ganzhi.GanGui}, {6, 8}: {ganzhi.GanBing, ganzhi.GanGui}, {6, 9}: {ganzhi.GanBing, ganzhi.GanJia},
	{6, 10}: {ganzhi.GanBing, ganzhi.GanJia}, {6, 11}: {ganzhi.GanBing, ganzhi.GanJia}, {6, 12}: {ganzhi.GanBing, ganzhi.GanJia},
	// 庚金
	{7, 1}: {ganzhi.GanWu, ganzhi.GanJia}, {7, 2}: {ganzhi.GanDing, ganzhi.GanBing}, {7, 3}: {ganzhi.GanJia, ganzhi.GanDing},
	{7, 4}: {ganzhi.GanRen, ganzhi.GanWu}, {7, 5}: {ganzhi.GanRen, ganzhi.GanJi}, {7, 6}: {ganzhi.GanRen, ganzhi.GanJi},
	{7, 7}: {ganzhi.GanDing, ganzhi.GanJia}, {7, 8}: {ganzhi.GanDing, ganzhi.GanBing}, {7, 9}: {ganzhi.GanJia, ganzhi.GanRen},
	{7, 10}: {ganzhi.GanDing, ganzhi.GanBing}, {7, 11}: {ganzhi.GanDing, ganzhi.GanBing}, {7, 12}: {ganzhi.GanDing, ganzhi.GanBing},
	// 辛金
	{8, 1}: {ganzhi.GanJi, ganzhi.GanRen}, {8, 2}: {ganzhi.GanRen, ganzhi.GanJia}, {8, 3}: {ganzhi.GanRen, ganzhi.GanJia},
	{8, 4}: {ganzhi.GanRen, ganzhi.GanJia}, {8, 5}: {ganzhi.GanRen, ganzhi.GanJi}, {8, 6}: {ganzhi.GanRen, ganzhi.GanJia},
	{8, 7}: {ganzhi.GanRen, ganzhi.GanJia}, {8, 8}: {ganzhi.GanRen, ganzhi.GanJia}, {8, 9}: {ganzhi.GanRen, ganzhi.GanJia},
	{8, 10}: {ganzhi.GanRen, ganzhi.GanBing}, {8, 11}: {ganzhi.GanBing, ganzhi.GanWu}, {8, 12}: {ganzhi.GanBing, ganzhi.GanRen},
	// 壬水
	{9, 1}: {ganzhi.GanGeng, ganzhi.GanWu}, {9, 2}: {ganzhi.GanWu, ganzhi.GanXin}, {9, 3}: {ganzhi.GanJia, ganzhi.GanGeng},
	{9, 4}: {ganzhi.GanRen, ganzhi.GanXin}, {9, 5}: {ganzhi.GanGui, ganzhi.GanGeng}, {9, 6}: {ganzhi.GanXin, ganzhi.GanJia},
	{9, 7}: {ganzhi.GanWu, ganzhi.GanDing}, {9, 8}: {ganzhi.GanJia, ganzhi.GanGeng}, {9, 9}: {ganzhi.GanJia, ganzhi.GanBing},
	{9, 10}: {ganzhi.GanWu, ganzhi.GanGeng}, {9, 11}: {ganzhi.GanWu, ganzhi.GanBing}, {9, 12}: {ganzhi.GanBing, ganzhi.GanDing},
	// 癸水
	{10, 1}: {ganzhi.GanXin, ganzhi.GanBing}, {10, 2}: {ganzhi.GanGeng, ganzhi.GanXin}, {10, 3}: {ganzhi.GanBing, ganzhi.GanXin},
	{10, 4}: {ganzhi.GanXin, ganzhi.GanRen}, {10, 5}: {ganzhi.GanGeng, ganzhi.GanRen}, {10, 6}: {ganzhi.GanGeng, ganzhi.GanXin},
	{10, 7}: {ganzhi.GanDing, ganzhi.GanXin}, {10, 8}: {ganzhi.GanXin, ganzhi.GanBing}, {10, 9}: {ganzhi.GanXin, ganzhi.GanJia},
	{10, 10}: {ganzhi.GanGeng, ganzhi.GanXin}, {10, 11}: {ganzhi.GanBing, ganzhi.GanXin}, {10, 12}: {ganzhi.GanBing, ganzhi.GanXin},
}

// computeTiaohou returns the 穷通宝鉴 climate-adjustment result for a given
// day-master and month-branch. Returns (TiaoHou, true) on match, or
// (zero, false) if no entry exists.
func computeTiaohou(dayMaster ganzhi.Gan, monthBranch ganzhi.Zhi) (TiaoHou, bool) {
	e, ok := lookupTiaohou[tiaohouKey{int(dayMaster), int(monthBranch)}]
	if !ok {
		return TiaoHou{}, false
	}

	yongElem := ganzhi.GanWuxing(e.primary)
	xiElem := ganzhi.GanWuxing(e.secondary)

	jiElem := pickJiElement(ganzhi.GanWuxing(dayMaster), e.primary, e.secondary)

	season := ganzhi.ZhiSeasonLabel(monthBranch)
	detail := fmt.Sprintf("%s月%s，用%s调候，%s辅之",
		ganzhi.ZhiName(monthBranch)+"月",
		ganzhi.GanName(dayMaster)+ganzhi.GanWuxing(dayMaster).String(),
		ganzhi.GanName(e.primary),
		ganzhi.GanName(e.secondary),
	)

	return TiaoHou{
		Season: season,
		Yong:   yongElem.String(),
		Xi:     xiElem.String(),
		Ji:     jiElem.String(),
		Detail: detail,
	}, true
}

// pickJiElement returns the Ji (忌神) element for the TiaoHou result.
// Ji is the element that controls (克) the day master. Prefer the yang stem;
// if it collides with the given yong/xi stems, try the yin stem. If both
// collide, fall back to the element that drains (泄) the day master.
func pickJiElement(dmElem ganzhi.Wuxing, yong, xi ganzhi.Gan) ganzhi.Wuxing {
	ctrlElem := elementThatControls(dmElem)
	ctrlYang := elementYangStem[ctrlElem]
	ctrlYin := elementYinStem[ctrlElem]

	// Prefer yang stem, then yin, then drain element.
	if ctrlYang != yong && ctrlYang != xi {
		return ctrlElem
	}
	if ctrlYin != yong && ctrlYin != xi {
		return ctrlElem
	}
	return elementThatDrains(dmElem)
}

// elementYangStem maps each element to its yang (阳) stem.
var elementYangStem = map[ganzhi.Wuxing]ganzhi.Gan{
	ganzhi.WxMu:   ganzhi.GanJia,
	ganzhi.WxHuo:  ganzhi.GanBing,
	ganzhi.WxTu:   ganzhi.GanWu,
	ganzhi.WxJin:  ganzhi.GanGeng,
	ganzhi.WxShui: ganzhi.GanRen,
}

// elementYinStem maps each element to its yin (阴) stem.
var elementYinStem = map[ganzhi.Wuxing]ganzhi.Gan{
	ganzhi.WxMu:   ganzhi.GanYi,
	ganzhi.WxHuo:  ganzhi.GanDing,
	ganzhi.WxTu:   ganzhi.GanJi,
	ganzhi.WxJin:  ganzhi.GanXin,
	ganzhi.WxShui: ganzhi.GanGui,
}

// elementThatDrains returns the element that the given element generates
// (生 = produces/drains). E.g., Wood drains to Fire (木生火).
func elementThatDrains(e ganzhi.Wuxing) ganzhi.Wuxing {
	switch e {
	case ganzhi.WxMu:
		return ganzhi.WxHuo
	case ganzhi.WxHuo:
		return ganzhi.WxTu
	case ganzhi.WxTu:
		return ganzhi.WxJin
	case ganzhi.WxJin:
		return ganzhi.WxShui
	case ganzhi.WxShui:
		return ganzhi.WxMu
	}
	return ganzhi.WxMu
}
