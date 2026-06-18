package bazi

import "liki/internal/engine/ganzhi"

// FuYi 扶抑取用——基于日主旺衰。
type FuYi struct {
	Strength string `json:"qiangruo"` // 身强 / 身弱 / 中和
	Pattern  string `json:"geju"`     // 格局
	Yong     string `json:"yong"`     // 用神
	Xi       string `json:"xi"`       // 喜神
	Ji       string `json:"ji"`       // 忌神
}

// TiaoHou 调候取用——基于月令气候，《穷通宝鉴》。
type TiaoHou struct {
	Season string `json:"season"`           // 季候
	Yong   string `json:"yong"`             // 调候用神
	Xi     string `json:"xi"`               // 调候喜神
	Ji     string `json:"ji"`               // 调候忌神
	Detail string `json:"detail,omitempty"` // 调候说明
}

// Strength represents the day master's relative strength (身强/身弱/中和).
type strength int

const (
	strengthWeak    strength = iota // 身弱
	strengthNeutral                 // 中和
	strengthStrong                  // 身强
)

// String returns the Chinese name for the strength level.
func (s strength) String() string {
	switch s {
	case strengthWeak:
		return "身弱"
	case strengthNeutral:
		return "中和"
	case strengthStrong:
		return "身强"
	default:
		return ""
	}
}

// computeFuYi computes the FuYi (扶抑) yongshen analysis from a chart.
func computeFuYi(chart Chart) FuYi {
	strength := computeDayMasterStrength(chart.WuxingCount, chart.DayMaster, chart.Month.Zhi, chart.HiddenStemsArray())
	yong, xi, ji := computeYongJiElements(chart.WuxingCount, chart.DayMaster, strength)
	var monthTenGodStem ganzhi.TenGod
	for _, e := range chart.Month.TenGods {
		if e.Source == sourceGan {
			monthTenGodStem = e.TenGod
			break
		}
	}
	pattern := computePattern(chart.DayMaster, chart.Month.Zhi, monthTenGodStem)
	return FuYi{
		Strength: strength.String(),
		Pattern:  pattern,
		Yong:     yong,
		Xi:       xi,
		Ji:       ji,
	}
}

// computeDayMasterStrength determines if the day master is 身强, 身弱, or 中和.
func computeDayMasterStrength(elementCount map[ganzhi.Wuxing]int, dayMaster ganzhi.Gan, monthBranch ganzhi.Zhi, hiddenStems [4]hiddenStemsOut) strength {
	dmElem := ganzhi.GanWuxing(dayMaster)
	monthElem := ganzhi.ZhiWuxing(monthBranch)
	genElem := elementThatGenerates(dmElem)
	ctrlElem := elementThatControls(dmElem)

	// Base support from element counts.
	support := elementCount[dmElem] + elementCount[genElem]

	// Seasonal weighting (旺相休囚死) based on month branch.
	var seasonScore int
	switch {
	case monthElem == dmElem:
		seasonScore = 3 // 旺: day master in season — strongest
	case monthElem == genElem:
		seasonScore = 2 // 相: generating element in season — strong
	case monthElem == ctrlElem:
		seasonScore = -2 // 死: day master controlled by season — weakest
	case elementThatControls(monthElem) == dmElem:
		seasonScore = -1 // 囚: day master controls season — trapped
	default:
		seasonScore = 0 // 休: neutral
	}

	// Weighted root bonus (通根): main qi (本气) = +2, mid qi (中气) = +1, minor qi (余气) = +1.
	rootBonus := 0
	for _, hs := range hiddenStems {
		if ganzhi.GanWuxing(hs.Main) == dmElem {
			rootBonus += 2
		}
		if hs.Mid != nil && ganzhi.GanWuxing(*hs.Mid) == dmElem {
			rootBonus++
		}
		if hs.Minor != nil && ganzhi.GanWuxing(*hs.Minor) == dmElem {
			rootBonus++
		}
	}

	total := support + seasonScore + rootBonus

	switch {
	case total >= 6:
		return strengthStrong
	case total <= 3:
		return strengthWeak
	default:
		return strengthNeutral
	}
}

// computeYongJiElements determines the favorable (用神), supporting (喜神), and
// unfavorable (忌神) elements based on day master strength.
func computeYongJiElements(elementCount map[ganzhi.Wuxing]int, dayMaster ganzhi.Gan, s strength) (yongShen, xiShen, jiShen string) {
	dmElem := ganzhi.GanWuxing(dayMaster)

	switch s {
	case strengthStrong:
		ctrlElem := elementThatControls(dmElem)
		yongShen = ctrlElem.String()
		xiShen = elementThatGenerates(ctrlElem).String()
		genElem := elementThatGenerates(dmElem)
		if elementCount[genElem] >= elementCount[dmElem] {
			jiShen = genElem.String()
		} else {
			jiShen = dmElem.String()
		}

	case strengthWeak, strengthNeutral:
		genElem := elementThatGenerates(dmElem)
		yongShen = genElem.String()
		xiShen = dmElem.String()
		ctrlElem := elementThatControls(dmElem)
		jiShen = ctrlElem.String()
	}
	return
}

// computePattern determines the chart pattern (格局).
func computePattern(dayMaster ganzhi.Gan, monthBranch ganzhi.Zhi, monthTenGodStem ganzhi.TenGod) string {
	dmElem := ganzhi.GanWuxing(dayMaster)
	monthElem := ganzhi.ZhiWuxing(monthBranch)

	if monthElem == dmElem {
		if ganzhi.GanYinYang(dayMaster) == ganzhi.Yang {
			return "建禄格"
		}
		return "月刃格"
	}

	switch monthTenGodStem {
	case ganzhi.TenGodZhengGuan:
		return "正官格"
	case ganzhi.TenGodQiSha:
		return "七杀格"
	case ganzhi.TenGodZhengCai:
		return "正财格"
	case ganzhi.TenGodPianCai:
		return "偏财格"
	case ganzhi.TenGodZhengYin:
		return "正印格"
	case ganzhi.TenGodPianYin:
		return "偏印格"
	case ganzhi.TenGodShiShen:
		return "食神格"
	case ganzhi.TenGodShangGuan:
		return "伤官格"
	}
	return "杂格"
}

// elementThatGenerates returns the element that generates (生) the given element.
func elementThatGenerates(e ganzhi.Wuxing) ganzhi.Wuxing {
	for i := ganzhi.WxMu; i <= ganzhi.WxShui; i++ {
		if ganzhi.Sheng(i, e) {
			return i
		}
	}
	return 0
}

// elementThatControls returns the element that controls (克) the given element.
func elementThatControls(e ganzhi.Wuxing) ganzhi.Wuxing {
	for i := ganzhi.WxMu; i <= ganzhi.WxShui; i++ {
		if ganzhi.Ke(i, e) {
			return i
		}
	}
	return 0
}
