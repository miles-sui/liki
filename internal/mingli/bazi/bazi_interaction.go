package bazi

import "fmt"

// StemRelation describes a single stem-to-stem relationship.
type StemRelation struct {
	StemA    Stem   `json:"stem_a"`
	StemB    Stem   `json:"stem_b"`
	Type     string `json:"type"`
	Relation string `json:"relation"`
}

// BranchRelation describes a single branch-to-branch relationship.
type BranchRelation struct {
	BranchA  Branch `json:"branch_a"`
	BranchB  Branch `json:"branch_b"`
	Type     string `json:"type"`
	Detail   string `json:"detail"`
}

// PillarInteraction holds stem and branch relations for one pillar against the bazi chart.
type PillarInteraction struct {
	PillarLabel string           `json:"pillar_label"`
	StemRels    []StemRelation   `json:"stem_rels"`
	BranchRels  []BranchRelation `json:"branch_rels"`
}

// AnalyzeStemRelation checks the relationship between two stems.
func AnalyzeStemRelation(a, b Stem) StemRelation {
	r := StemRelation{StemA: a, StemB: b}
	if a == b {
		r.Type = "相同"
		r.Relation = fmt.Sprintf("%s%s同气", stemNameStr(a), stemNameStr(b))
		return r
	}

	e := eng()
	if e != nil {
		for _, p := range defaultData.StemHePairs {
			if (a == p.A && b == p.B) || (a == p.B && b == p.A) {
				r.Type = "天干五合"
				r.Relation = fmt.Sprintf("%s%s合化%s", stemNameStr(a), stemNameStr(b), p.Result.String())
				return r
			}
		}
	}

	aElem, bElem := StemElement(a), StemElement(b)
	if aElem == bElem {
		r.Type = "相同"
		r.Relation = fmt.Sprintf("%s%s同气", stemNameStr(a), stemNameStr(b))
	} else if Sheng(aElem, bElem) {
		r.Type = "相生"
		r.Relation = fmt.Sprintf("%s生%s", stemNameStr(a), stemNameStr(b))
	} else if Sheng(bElem, aElem) {
		r.Type = "相生"
		r.Relation = fmt.Sprintf("%s生%s", stemNameStr(b), stemNameStr(a))
	} else if Ke(aElem, bElem) {
		r.Type = "相克"
		r.Relation = fmt.Sprintf("%s克%s", stemNameStr(a), stemNameStr(b))
	} else if Ke(bElem, aElem) {
		r.Type = "相克"
		r.Relation = fmt.Sprintf("%s克%s", stemNameStr(b), stemNameStr(a))
	} else {
		r.Type = "无"
		r.Relation = "无特殊关系"
	}
	return r
}

// Branch relation extra types — hardcoded per traditional theory.
var (
	// 暗合: hidden harmony via hidden stems. Not in config YAML.
	anhePairs = []BranchPair{
		{A: 3, B: 2},  // 寅丑
		{A: 4, B: 9},  // 卯申
		{A: 7, B: 12}, // 午亥
		{A: 1, B: 11}, // 子戌
	}
	// 破: mutual destruction.
	poPairs = []BranchPair{
		{A: 1, B: 10},  // 子酉
		{A: 3, B: 12},  // 寅亥
		{A: 5, B: 2},   // 辰丑
		{A: 7, B: 4},   // 午卯
		{A: 9, B: 6},   // 申巳
		{A: 11, B: 8},  // 戌未
	}
)

// AnalyzeBranchRelation checks all relationship types between two branches.
// Priority: 六合 > 三合 > 三会 > 六冲 > 相刑 > 六害 > 暗合 > 破.
func AnalyzeBranchRelation(a, b Branch) BranchRelation {
	r := BranchRelation{BranchA: a, BranchB: b}
	if a == b {
		r.Type = "相同"
		r.Detail = fmt.Sprintf("%s%s同气", branchNameStr(a), branchNameStr(b))
		return r
	}

	e := eng()
	if e == nil {
		r.Type = "无"
		r.Detail = "无特殊关系"
		return r
	}

	// 六合
	for _, p := range defaultData.BranchHePairs {
		if (a == p.A && b == p.B) || (a == p.B && b == p.A) {
			r.Type = "六合"
			r.Detail = fmt.Sprintf("%s%s合化%s", branchNameStr(a), branchNameStr(b), p.Result.String())
			return r
		}
	}

	// 三合
	for _, th := range defaultData.TripleHeList {
		if ContainsPair(th.Branches, int(a), int(b)) {
			r.Type = "三合"
			r.Detail = fmt.Sprintf("%s%s三合%s局", branchNameStr(a), branchNameStr(b), Element(th.Element).String())
			return r
		}
	}

	// 三会
	for _, th := range defaultData.TripleHuiList {
		if ContainsPair(th.Branches, int(a), int(b)) {
			r.Type = "三会"
			r.Detail = fmt.Sprintf("%s%s三会%s方", branchNameStr(a), branchNameStr(b), Element(th.Element).String())
			return r
		}
	}

	// 六冲
	for _, p := range defaultData.ChongPairs {
		if (a == p.A && b == p.B) || (a == p.B && b == p.A) {
			r.Type = "六冲"
			r.Detail = fmt.Sprintf("%s%s相冲", branchNameStr(a), branchNameStr(b))
			return r
		}
	}

	// 相刑
	for _, x := range defaultData.XingGroups {
		if ContainsPair(x.Branches, int(a), int(b)) {
			r.Type = "相刑"
			r.Detail = fmt.Sprintf("%s%s%s", branchNameStr(a), branchNameStr(b), xingTypeLabel(x.Type))
			return r
		}
	}

	// 六害
	for _, p := range defaultData.HaiPairs {
		if (a == p.A && b == p.B) || (a == p.B && b == p.A) {
			r.Type = "六害"
			r.Detail = fmt.Sprintf("%s%s相害", branchNameStr(a), branchNameStr(b))
			return r
		}
	}

	// 暗合
	for _, p := range anhePairs {
		if (a == p.A && b == p.B) || (a == p.B && b == p.A) {
			r.Type = "暗合"
			r.Detail = fmt.Sprintf("%s%s暗合", branchNameStr(a), branchNameStr(b))
			return r
		}
	}

	// 破
	for _, p := range poPairs {
		if (a == p.A && b == p.B) || (a == p.B && b == p.A) {
			r.Type = "破"
			r.Detail = fmt.Sprintf("%s%s相破", branchNameStr(a), branchNameStr(b))
			return r
		}
	}

	r.Type = "无"
	r.Detail = "无特殊关系"
	return r
}

func xingTypeLabel(t string) string {
	switch t {
	case "wuli":
		return "无礼之刑"
	case "wuen":
		return "无恩之刑"
	case "shishi":
		return "恃势之刑"
	case "zi":
		return "自刑"
	}
	return "刑"
}

// AnalyzePillarWithBazi analyzes one pillar against all 4 bazi chart pillars.
func AnalyzePillarWithBazi(pillar Pillar, bz Bazi) ([]StemRelation, []BranchRelation) {
	stemRels := make([]StemRelation, 4)
	branchRels := make([]BranchRelation, 4)
	for i, np := range bz.Slice() {
		stemRels[i] = AnalyzeStemRelation(pillar.Stem, np.Stem)
		branchRels[i] = AnalyzeBranchRelation(pillar.Branch, np.Branch)
	}
	return stemRels, branchRels
}

// ComputeDayunInteractions computes structured interactions for all dayun pillars vs bazi chart.
func ComputeDayunInteractions(dayunPillars []DayunPillar, bz Bazi) [
]PillarInteraction {
	out := make([]PillarInteraction, len(dayunPillars))
	for i, dp := range dayunPillars {
		pillar := Pillar{Stem: dp.Stem, Branch: dp.Branch}
		stemRels, branchRels := AnalyzePillarWithBazi(pillar, bz)
		out[i] = PillarInteraction{
			PillarLabel: fmt.Sprintf("%s运（%d-%d岁，%s）", dp.Name, dp.AgeStart, dp.AgeEnd, dp.TenGod),
			StemRels:    stemRels,
			BranchRels:  branchRels,
		}
	}
	return out
}

