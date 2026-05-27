package qiming

// GeResult is one of the five-grid (五格) numbers.
type GeResult struct {
	Stroke      int    `json:"stroke"`
	Element     string `json:"element"`
	Fortune     string `json:"fortune"`
	Description string `json:"description"`
}

// WuGe holds the five-grid (五格) analysis.
type WuGe struct {
	TianGe GeResult `json:"tian_ge"`
	RenGe  GeResult `json:"ren_ge"`
	DiGe   GeResult `json:"di_ge"`
	WaiGe  GeResult `json:"wai_ge"`
	ZongGe GeResult `json:"zong_ge"`
}

// ComputeWuGe computes the five-grid analysis for a name.
func ComputeWuGe(surname string, givenChars []string) WuGe {
	surnameStroke := LookupKangxiStroke(surname)
	name1Stroke := 0
	name2Stroke := 0
	if len(givenChars) > 0 {
		name1Stroke = LookupKangxiStroke(givenChars[0])
	}
	if len(givenChars) > 1 {
		name2Stroke = LookupKangxiStroke(givenChars[1])
	}

	tian := surnameStroke + 1
	ren := surnameStroke + name1Stroke
	di := name1Stroke + name2Stroke
	if di == 0 {
		di = name1Stroke + 1
	}
	zong := surnameStroke + name1Stroke + name2Stroke
	wai := zong - ren + 1
	if wai < 1 {
		wai = 1
	}

	return WuGe{
		TianGe: strokeResult(tian),
		RenGe:  strokeResult(ren),
		DiGe:   strokeResult(di),
		WaiGe:  strokeResult(wai),
		ZongGe: strokeResult(zong),
	}
}

func strokeResult(stroke int) GeResult {
	e := defaultEngine
	if stroke < 1 {
		stroke = 1
	}
	if stroke > 81 {
		stroke = ((stroke - 1) % 81) + 1
	}
	if e != nil {
		if v, ok := e.SanCaiNums[stroke]; ok {
			return GeResult{Stroke: stroke, Element: v.Element, Fortune: v.Fortune, Description: v.Desc}
		}
	}
	return GeResult{Stroke: stroke, Element: "土", Fortune: "半吉", Description: ""}
}
