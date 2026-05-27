package bazi

import "github.com/25types/25types/internal/ganzhi"

// TripleHeFull describes a complete 三合局 formed by the bazi chart's branches.
type TripleHeFull struct {
	Type    string `json:"type"`    // "三合" or "三会"
	Name    string `json:"name"`    // "申子辰水局"
	Element string `json:"element"` // "水"
}

func branchSet(bz ganzhi.Bazi) [13]bool {
	pillars := bz.Slice()
	var bs [13]bool
	for _, p := range pillars {
		if b := int(p.Branch); b >= 1 && b <= 12 {
			bs[b] = true
		}
	}
	return bs
}

func countBranches(bs [13]bool, targets ...int) int {
	c := 0
	for _, t := range targets {
		if t >= 1 && t <= 12 && bs[t] {
			c++
		}
	}
	return c
}

// ComputeFullTripleHeHui detects complete 三合局 and 三会方 across the bazi pillars.
func ComputeFullTripleHeHui(bz ganzhi.Bazi) []TripleHeFull {
	bs := branchSet(bz)
	var results []TripleHeFull

	// 三合局: 申子辰(水), 亥卯未(木), 寅午戌(火), 巳酉丑(金)
	type triple struct {
		bs      [3]int
		name    string
		element string
	}
	triples := []triple{
		{[3]int{9, 1, 5}, "申子辰水局", "水"},
		{[3]int{12, 4, 8}, "亥卯未木局", "木"},
		{[3]int{3, 7, 11}, "寅午戌火局", "火"},
		{[3]int{6, 10, 2}, "巳酉丑金局", "金"},
	}

	for _, tr := range triples {
		if countBranches(bs, tr.bs[:]...) == 3 {
			results = append(results, TripleHeFull{Type: "三合", Name: tr.name, Element: tr.element})
		}
	}

	// 三会方: 寅卯辰(木), 巳午未(火), 申酉戌(金), 亥子丑(水)
	huis := []triple{
		{[3]int{3, 4, 5}, "寅卯辰木方", "木"},
		{[3]int{6, 7, 8}, "巳午未火方", "火"},
		{[3]int{9, 10, 11}, "申酉戌金方", "金"},
		{[3]int{12, 1, 2}, "亥子丑水方", "水"},
	}

	for _, hu := range huis {
		if countBranches(bs, hu.bs[:]...) == 3 {
			results = append(results, TripleHeFull{Type: "三会", Name: hu.name, Element: hu.element})
		}
	}

	return results
}

// IsStemHe returns true if the two stems form a 天干五合 pair.
func IsStemHe(a, b Stem) bool { return ganzhi.IsStemHe(a, b) }

// IsBranchHe returns true if the two branches form a 地支六合 pair.
func IsBranchHe(a, b Branch) bool { return ganzhi.IsBranchHe(a, b) }

// IsTripleHe returns true if the two branches are part of the same 三合局.
func IsTripleHe(a, b Branch) bool { return ganzhi.IsTripleHe(a, b) }

// IsTripleHui returns true if the two branches are part of the same 三会方.
func IsTripleHui(a, b Branch) bool { return ganzhi.IsTripleHui(a, b) }

// IsLiuChong returns true if the two branches form a 六冲 pair.
func IsLiuChong(a, b Branch) bool { return ganzhi.IsLiuChong(a, b) }

// IsXing returns true if the two branches form a 相刑 pair.
func IsXing(a, b Branch) bool { return ganzhi.IsXing(a, b) }

// IsHai returns true if the two branches form a 六害 pair.
func IsHai(a, b Branch) bool { return ganzhi.IsHai(a, b) }

func ContainsPair(list []int, a, b int) bool {
	hasA, hasB := false, false
	for _, v := range list {
		if v == a {
			hasA = true
		}
		if v == b {
			hasB = true
		}
	}
	return hasA && hasB
}

