package bazi

import (
	"sort"

	"github.com/25types/25types/internal/ganzhi"
)

// GongJiaEntry describes a 拱 or 夹 between two bazi pillars.
type GongJiaEntry struct {
	PillarA   int    `json:"pillar_a"`   // index 0-3 of first pillar
	PillarB   int    `json:"pillar_b"`   // index 0-3 of second pillar
	Type      string `json:"type"`       // "拱" or "夹"
	Branch    Branch `json:"branch"`     // the hidden branch between them
}

// ComputeGongJia detects 拱 (gap=2) and 夹 (gap=1) between branches of bazi pillars.
// When two pillar branches differ by 2 (mod 12), the midpoint = 拱.
// When two pillar branches differ by 1 (mod 12), there's a 夹 at gap 1.
func ComputeGongJia(bz ganzhi.Bazi) []GongJiaEntry {
	pillars := bz.Slice()
	// Collect unique branch indices.
	bs := make([]int, 0, 4)
	seen := [13]bool{}
	for _, p := range pillars {
		b := int(p.Branch)
		if b >= 1 && b <= 12 && !seen[b] {
			seen[b] = true
			bs = append(bs, b)
		}
	}
	sort.Ints(bs)

	var results []GongJiaEntry

	for i := 0; i < len(bs); i++ {
		for j := i + 1; j < len(bs); j++ {
			a, bb := bs[i], bs[j]
			forward := (bb - a + 12) % 12
			backward := (a - bb + 12) % 12
			gap := forward
			if backward < forward {
				gap = backward
			}

			switch gap {
			case 2: // 拱
				mid := ((a + bb) / 2)
				if (bb-a+12)%12 == 4 { // 顺时针隔1
					mid = (a + 1)
					if mid > 12 {
						mid -= 12
					}
				}
				// The hidden branch = (a + 1) mod 12 clockwise
				midB := a%12 + 1
				if midB > 12 {
					midB = 1
				}
				// Find which pillars correspond to a and bb
				pA, pB := pillarIndexForBranch(bz, a), pillarIndexForBranch(bz, bb)
				if pA >= 0 && pB >= 0 {
					results = append(results, GongJiaEntry{
						PillarA:    pA,
						PillarB:    pB,
						Type:       "拱",
						Branch:     Branch(midB),
						
					})
				}
			case 1: // 夹
				pA, pB := pillarIndexForBranch(bz, a), pillarIndexForBranch(bz, bb)
				if pA >= 0 && pB >= 0 {
					// The branch between them doesn't fit a "拱" case, it's adjacent
					gapB := a%12 + 1
					if gapB > 12 {
						gapB = 1
					}
					results = append(results, GongJiaEntry{
						PillarA:    pA,
						PillarB:    pB,
						Type:       "夹",
						Branch:     Branch(gapB),
						
					})
				}
			}
		}
	}

	return results
}

func pillarIndexForBranch(bz ganzhi.Bazi, b int) int {
	pillars := bz.Slice()
	for i, p := range pillars {
		if int(p.Branch) == b {
			return i
		}
	}
	return -1
}
