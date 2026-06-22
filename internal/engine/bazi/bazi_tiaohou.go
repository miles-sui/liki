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

// computeTiaohou returns the 穷通宝鉴 climate-adjustment result for a given
// day-master and month-branch. Returns (TiaoHou, true) on match, or
// (zero, false) if no entry exists.
func computeTiaohou(riYuan ganzhi.Gan, monthBranch ganzhi.Zhi) (TiaoHou, bool) {
	e, ok := lookupTiaohou[tiaohouKey{int(riYuan), int(monthBranch)}]
	if !ok {
		return TiaoHou{}, false
	}

	yongElem := ganzhi.GanWuxing(e.primary)
	xiElem := ganzhi.GanWuxing(e.secondary)

	jiElem := pickJiElement(ganzhi.GanWuxing(riYuan), e.primary, e.secondary)

	season := ganzhi.ZhiSeasonLabel(monthBranch)
	detail := fmt.Sprintf("%s月%s，用%s调候，%s辅之",
		ganzhi.ZhiName(monthBranch)+"月",
		ganzhi.GanName(riYuan)+ganzhi.GanWuxing(riYuan).String(),
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

// elementYangStem maps each element to its yang (阳) stem. Loaded from shensha.json.
var elementYangStem map[ganzhi.Wuxing]ganzhi.Gan

// elementYinStem maps each element to its yin (阴) stem. Loaded from shensha.json.
var elementYinStem map[ganzhi.Wuxing]ganzhi.Gan

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
