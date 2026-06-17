package qiming

import "liki/internal/engine/ganzhi"

// ge is one of the five-grid (五格) numbers.
type ge struct {
	Stroke      int    `json:"stroke"`
	Element     string `json:"wuxing"`
	Fortune     string `json:"fortune"`
	Description string `json:"description"`
}

// WuGe holds the five-grid (五格) analysis.
type WuGe struct {
	TianGe ge `json:"tian_ge"`
	RenGe  ge `json:"ren_ge"`
	DiGe   ge `json:"di_ge"`
	WaiGe  ge `json:"wai_ge"`
	ZongGe ge `json:"zong_ge"`
}

// enumWuGeCombinations enumerates all auspicious stroke1+stroke2 pairs for a given
// surname Kangxi stroke count. A combination is included only when the three-talent
// mutual-generation condition holds AND all five grids are individually auspicious.
func enumWuGeCombinations(surnameStrokes int) wuGeEnumerationResult {
	tianRaw := surnameStrokes + 1
	tian := strokeResult(tianRaw)

	var combos []StrokeCombo
	for s1 := 1; s1 <= 31; s1++ {
		for s2 := 1; s2 <= 31; s2++ {
			renRaw := surnameStrokes + s1
			diRaw := s1 + s2
			zongRaw := surnameStrokes + s1 + s2
			waiRaw := zongRaw - renRaw + 1
			if waiRaw < 1 {
				waiRaw = 1
			}

			ren := strokeResult(renRaw)
			di := strokeResult(diRaw)
			zong := strokeResult(zongRaw)
			wai := strokeResult(waiRaw)

			if !isAuspicious(ren.Fortune) || !isAuspicious(di.Fortune) ||
				!isAuspicious(zong.Fortune) || !isAuspicious(wai.Fortune) {
				continue
			}

			tianElem := wuxingFromChinese(tian.Element)
			renElem := wuxingFromChinese(ren.Element)
			diElem := wuxingFromChinese(di.Element)
			if !ganzhi.Sheng(tianElem, renElem) || !ganzhi.Sheng(renElem, diElem) {
				continue
			}

			sc := computeSanCai(tian.Element, ren.Element, di.Element)
			combos = append(combos, StrokeCombo{
				Stroke1: s1,
				Stroke2: s2,
				SanCai:  sc.Configuration,
				Fortune: sc.Fortune,
			})
		}
	}

	return wuGeEnumerationResult{
		SurnameStrokes: surnameStrokes,
		TianGe:         tianGeBrief{Stroke: tian.Stroke, Wuxing: tian.Element},
		Combinations:   combos,
	}
}

func isAuspicious(fortune string) bool {
	return fortune == "吉" || fortune == "大吉"
}

func strokeResult(stroke int) ge {
	if stroke < 1 {
		stroke = 1
	}
	if stroke > 81 {
		stroke = ((stroke - 1) % 81) + 1
	}
	if v, ok := sanCaiNums[stroke]; ok {
		return ge{Stroke: stroke, Element: v.Element, Fortune: v.Fortune, Description: v.Desc}
	}
	return ge{Stroke: stroke, Element: "土", Fortune: "半吉", Description: ""}
}
