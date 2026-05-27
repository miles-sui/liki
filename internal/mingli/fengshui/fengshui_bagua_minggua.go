package fengshui

import "github.com/25types/25types/internal/ganzhi"

// -- 八卦纳甲 ----------------------------------------------------------------

// Trigram represents one of the eight Bagua trigrams.
type Trigram struct {
	Index   int    `json:"index"`   // 1-8
	Name    string `json:"name"`    // e.g. "乾"
	Element ganzhi.Element `json:"element"`
	Direction string `json:"direction"` // e.g. "西北"
}

var trigramTable = [9]Trigram{
	{},
	{1, "乾", ganzhi.ElemMetal, "西北"},
	{2, "兑", ganzhi.ElemMetal, "西"},
	{3, "离", ganzhi.ElemFire, "南"},
	{4, "震", ganzhi.ElemWood, "东"},
	{5, "巽", ganzhi.ElemWood, "东南"},
	{6, "坎", ganzhi.ElemWater, "北"},
	{7, "艮", ganzhi.ElemEarth, "东北"},
	{8, "坤", ganzhi.ElemEarth, "西南"},
}

// StemNaJia maps a stem to its Bagua trigram (天干纳甲).
func StemNaJia(stem Stem) Trigram {
	switch stem {
	case 1, 9: // 甲, 壬
		return trigramTable[1] // 乾
	case 2, 10: // 乙, 癸
		return trigramTable[8] // 坤
	case 3: // 丙
		return trigramTable[7] // 艮
	case 4: // 丁
		return trigramTable[2] // 兑
	case 5: // 戊
		return trigramTable[6] // 坎
	case 6: // 己
		return trigramTable[3] // 离
	case 7: // 庚
		return trigramTable[4] // 震
	case 8: // 辛
		return trigramTable[5] // 巽
	}
	return Trigram{}
}

// BranchNaJia maps a branch to its Bagua trigram (地支纳甲/卦宫).
func BranchNaJia(branch Branch) Trigram {
	switch branch {
	case 1: // 子
		return trigramTable[6] // 坎
	case 2, 3: // 丑, 寅
		return trigramTable[7] // 艮
	case 4: // 卯
		return trigramTable[4] // 震
	case 5, 6: // 辰, 巳
		return trigramTable[5] // 巽
	case 7: // 午
		return trigramTable[3] // 离
	case 8, 9: // 未, 申
		return trigramTable[8] // 坤
	case 10: // 酉
		return trigramTable[2] // 兑
	case 11, 12: // 戌, 亥
		return trigramTable[1] // 乾
	}
	return Trigram{}
}

// PillarNaJia returns the trigram for a pillar based on its branch (卦宫).
func PillarNaJia(p Pillar) Trigram {
	return BranchNaJia(p.Branch)
}

// AllTrigrams returns the full Bagua trigram table.
func AllTrigrams() [9]Trigram {
	return trigramTable
}

// -- 命卦 (八宅命卦) ----------------------------------------------------------

// MingGuaResult holds the fate trigram (命卦) and related info.
type MingGuaResult struct {
	Gua       Trigram `json:"gua"`
	GuaNumber int     `json:"gua_number"` // 1-9
	Group     string  `json:"group"`      // "东四命" or "西四命"
}

// east/west four groups
var (
	eastGroup  = map[int]bool{1: true, 3: true, 4: true, 9: true}  // 坎震巽离
	westGroup = map[int]bool{2: true, 6: true, 7: true, 8: true}   // 坤乾兑艮
)

// ComputeMingGua computes the 命卦 (birth fate trigram) from birth year and gender.
//
// Formula (traditional 八宅):
//
//	Male:   n = (100 - year) % 9;  n=0 → 9, n=5 → male→2(坤)
//	Female: n = (year - 4) % 9;    n=0 → 9, n=5 → female→8(艮)
func ComputeMingGua(gender ganzhi.Gender, birthYear int) MingGuaResult {
	shortYear := birthYear % 100
	var n int
	if gender == ganzhi.Male {
		n = (100 - shortYear) % 9
		if n < 0 {
			n += 9
		}
		if n == 0 {
			n = 9
		}
		if n == 5 {
			n = 2 // 坤
		}
	} else {
		n = (shortYear - 4) % 9
		if n < 0 {
			n += 9
		}
		if n == 0 {
			n = 9
		}
		if n == 5 {
			n = 8 // 艮
		}
	}

	group := "东四命"
	if westGroup[n] {
		group = "西四命"
	}

	return MingGuaResult{
		Gua:       trigramTable[n],
		GuaNumber: n,
		Group:     group,
	}
}
