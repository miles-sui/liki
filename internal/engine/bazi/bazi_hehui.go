package bazi

import (
	"strings"

	"liki/internal/engine/ganzhi"
)

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
		if b := int(p.Zhi); b >= 1 && b <= 12 {
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

func tripleName(branches []int, element ganzhi.Wuxing, suffix string) string {
	parts := make([]string, len(branches))
	for i, b := range branches {
		parts[i] = ganzhi.ZhiName(ganzhi.Zhi(b))
	}
	return strings.Join(parts, "") + element.String() + suffix
}

// computeFullTripleHeHui detects complete 三合局 and 三会方 across the bazi pillars.
func computeFullTripleHeHui(bz ganzhi.Bazi) []TripleHeFull {
	bs := branchSet(bz)
	var results []TripleHeFull

	for _, tr := range ganzhi.TripleHeList {
		if countBranches(bs, tr.Branches...) == len(tr.Branches) {
			results = append(results, TripleHeFull{
				Type:    relSanHe,
				Name:    tripleName(tr.Branches, tr.Element, "局"),
				Element: tr.Element.String(),
			})
		}
	}

	for _, tr := range ganzhi.TripleHuiList {
		if countBranches(bs, tr.Branches...) == len(tr.Branches) {
			results = append(results, TripleHeFull{
				Type:    relSanHui,
				Name:    tripleName(tr.Branches, tr.Element, "方"),
				Element: tr.Element.String(),
			})
		}
	}

	return results
}

func containsPair(list []int, a, b int) bool {
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
