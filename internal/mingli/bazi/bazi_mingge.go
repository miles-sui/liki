package bazi

// YongShenResult 用神分析结果——扶抑与调候两套独立取用体系。
type YongShenResult struct {
	FuYi    FuYiResult    `json:"fuyi"`
	TiaoHou TiaoHouResult `json:"tiaohou"`
}

// FuYiResult 扶抑取用——基于日主旺衰。
type FuYiResult struct {
	Strength string `json:"strength"` // 身强 / 身弱 / 中和
	Pattern  string `json:"pattern"`  // 格局
	Yong     string `json:"yong"`     // 用神
	Xi       string `json:"xi"`       // 喜神
	Ji       string `json:"ji"`       // 忌神
}

// TiaoHouResult 调候取用——基于月令气候，《穷通宝鉴》。
type TiaoHouResult struct {
	Season string `json:"season"`           // 季候
	Yong   string `json:"yong"`             // 调候用神
	Xi     string `json:"xi"`               // 调候喜神
	Ji     string `json:"ji"`               // 调候忌神
	Detail string `json:"detail,omitempty"` // 调候说明
}

// Strength represents the day master's relative strength (身强/身弱/中和).
type Strength int

const (
	StrengthWeak    Strength = iota // 身弱
	StrengthNeutral                 // 中和
	StrengthStrong                  // 身强
)

// String returns the Chinese name for the strength level.
func (s Strength) String() string {
	switch s {
	case StrengthWeak:
		return "身弱"
	case StrengthNeutral:
		return "中和"
	case StrengthStrong:
		return "身强"
	default:
		return ""
	}
}

// ComputeYongShen computes the full yongshen analysis (扶抑 + 调候) from a chart.
func ComputeYongShen(chart ChartResult) YongShenResult {
	strength := ComputeDayMasterStrength(chart.ElementCount, chart.DayMaster, chart.Month.Branch)
	yong, xi, ji := computeYongShenElements(chart.ElementCount, chart.DayMaster, strength)
	monthTenGodStem := ""
	for _, e := range chart.Month.TenGods {
		if e.Source == SourceStem {
			monthTenGodStem = e.TenGod
			break
		}
	}
	pattern := ComputePattern(chart.DayMaster, chart.Month.Branch, monthTenGodStem)

	fuyi := FuYiResult{
		Strength: strength.String(),
		Pattern:  pattern,
		Yong:     yong,
		Xi:       xi,
		Ji:       ji,
	}

	var tiaohou TiaoHouResult
	if th, ok := ComputeTiaohou(chart.DayMaster, chart.Month.Branch); ok {
		tiaohou = th
	}

	return YongShenResult{FuYi: fuyi, TiaoHou: tiaohou}
}

// ComputeDayMasterStrength determines if the day master is 身强, 身弱, or 中和.
func ComputeDayMasterStrength(elementCount map[Element]int, dayMaster Stem, monthBranch Branch) Strength {
	dmElem := StemElement(dayMaster)
	monthElem := BranchElement(monthBranch)

	genElem := ElementThatGenerates(dmElem)

	support := elementCount[dmElem] + elementCount[genElem]
	monthSupports := monthElem == dmElem || monthElem == genElem

	if support >= 5 {
		return StrengthStrong
	}
	if support <= 2 {
		return StrengthWeak
	}
	if monthSupports {
		return StrengthStrong
	}
	if support == 3 {
		return StrengthWeak
	}
	return StrengthNeutral
}

// computeYongShenElements determines the favorable (用神), supporting (喜神), and
// unfavorable (忌神) elements based on day master strength.
func computeYongShenElements(elementCount map[Element]int, dayMaster Stem, strength Strength) (yongShen, xiShen, jiShen string) {
	dmElem := StemElement(dayMaster)

	switch strength {
	case StrengthStrong:
		ctrlElem := ElementThatControls(dmElem)
		yongShen = ctrlElem.String()
		xiShen = ElementThatGenerates(ctrlElem).String()
		genElem := ElementThatGenerates(dmElem)
		if elementCount[genElem] >= elementCount[dmElem] {
			jiShen = genElem.String()
		} else {
			jiShen = dmElem.String()
		}

	case StrengthWeak, StrengthNeutral:
		genElem := ElementThatGenerates(dmElem)
		yongShen = genElem.String()
		xiShen = dmElem.String()
		ctrlElem := ElementThatControls(dmElem)
		jiShen = ctrlElem.String()
	}
	return
}

// ComputePattern determines the chart pattern (格局).
func ComputePattern(dayMaster Stem, monthBranch Branch, monthTenGodStem string) string {
	dmElem := StemElement(dayMaster)
	monthElem := BranchElement(monthBranch)

	if monthElem == dmElem {
		if StemYinYang(dayMaster) == Yang {
			return "建禄格"
		}
		return "月刃格"
	}

	switch monthTenGodStem {
	case "正官":
		return "正官格"
	case "七杀":
		return "七杀格"
	case "正财":
		return "正财格"
	case "偏财":
		return "偏财格"
	case "正印":
		return "正印格"
	case "偏印":
		return "偏印格"
	case "食神":
		return "食神格"
	case "伤官":
		return "伤官格"
	}
	return "杂格"
}

// DayMasterNameString returns the Chinese name for a day master stem.
func DayMasterNameString(s Stem) string {
	names := map[Stem]string{
		StemJia: "甲木", StemYi: "乙木", StemBing: "丙火", StemDing: "丁火",
		StemWu: "戊土", StemJi: "己土", StemGeng: "庚金", StemXin: "辛金",
		StemRen: "壬水", StemGui: "癸水",
	}
	if n, ok := names[s]; ok {
		return n
	}
	return "未知"
}

// MonthBranchNameString returns Chinese name for a month branch.
func MonthBranchNameString(monthBranch Branch) string {
	names := map[Branch]string{
		BranchYin: "寅月", BranchMao: "卯月", BranchChen: "辰月", BranchSi: "巳月",
		BranchWu: "午月", BranchWei: "未月", BranchShen: "申月", BranchYou: "酉月",
		BranchXu: "戌月", BranchHai: "亥月", BranchZi: "子月", BranchChou: "丑月",
	}
	if n, ok := names[monthBranch]; ok {
		return n
	}
	return "未知月"
}

var (
	elementGenerator = map[Element]Element{
		ElemWood:  ElemWater,
		ElemFire:  ElemWood,
		ElemEarth: ElemFire,
		ElemMetal: ElemEarth,
		ElemWater: ElemMetal,
	}
	elementController = map[Element]Element{
		ElemWood:  ElemMetal,
		ElemFire:  ElemWater,
		ElemEarth: ElemWood,
		ElemMetal: ElemFire,
		ElemWater: ElemEarth,
	}
)

// ElementThatGenerates returns the element that generates (生) the given element.
func ElementThatGenerates(e Element) Element { return elementGenerator[e] }

// ElementThatControls returns the element that controls (克) the given element.
func ElementThatControls(e Element) Element { return elementController[e] }
