package bazi

import "github.com/25types/25types/internal/ganzhi"

// FuYinFanYinEntry describes a 伏吟 or 反吟 hit between a flow pillar and a bazi pillar.
type FuYinFanYinEntry struct {
	NatalIndex int    `json:"natal_index"` // 0-3 (年/月/日/时)
	Type       string `json:"type"`       // "伏吟" or "反吟"
	Detail     string `json:"detail"`
}

// ComputeFuYinFanYin checks a flow pillar (e.g., liunian or dayun) against each
// bazi pillar for 伏吟 (same stem+branch) and 反吟 (stem clash + branch clash —
// 天克地冲).
func ComputeFuYinFanYin(flow Pillar, bz ganzhi.Bazi) []FuYinFanYinEntry {
	bazi := bz.Slice()
	var entries []FuYinFanYinEntry

	for i, np := range bazi {
		sameStem := flow.Stem == np.Stem
		sameBranch := flow.Branch == np.Branch

		if sameStem && sameBranch {
			entries = append(entries, FuYinFanYinEntry{
				NatalIndex: i,
				Type:       "伏吟",
				Detail:     stemNameStr(flow.Stem) + branchNameStr(flow.Branch) + "伏吟",
			})
		} else if sameBranch && !sameStem {
			entries = append(entries, FuYinFanYinEntry{
				NatalIndex: i,
				Type:       "伏吟",
				Detail:     branchNameStr(flow.Branch) + "地支伏吟",
			})
		}

		// 反吟: 天克地冲 (stem clash AND branch clash)
		sr := AnalyzeStemRelation(flow.Stem, np.Stem)
		br := AnalyzeBranchRelation(flow.Branch, np.Branch)
		if sr.Type == "相克" && br.Type == "六冲" {
			entries = append(entries, FuYinFanYinEntry{
				NatalIndex: i,
				Type:       "反吟",
				Detail:     stemNameStr(flow.Stem) + branchNameStr(flow.Branch) +
					"与" + stemNameStr(np.Stem) + branchNameStr(np.Branch) + "天克地冲",
			})
		}
	}

	return entries
}
