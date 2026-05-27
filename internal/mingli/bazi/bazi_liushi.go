package bazi

import (
	"fmt"

	"github.com/25types/25types/internal/ganzhi"
)

// LiushiResult holds the hourly flow (流时) analysis: the current two-hour period's
// pillar and its interactions with the bazi chart.
type LiushiResult struct {
	Time       string           `json:"time"`
	HourStem   Stem             `json:"hour_stem"`
	HourBranch Branch           `json:"hour_branch"`
	HourName   string           `json:"hour_name"`
	TenGod     string           `json:"ten_god"`
	StemRels   []StemRelation   `json:"stem_rels"`
	BranchRels []BranchRelation `json:"branch_rels"`
}

// hourBranchIndex maps date hour to traditional "时辰" branch index (0-11, 0=子).
// 23:00-00:59 → 0, 01:00-02:59 → 1, etc.
func hourBranchIndex(hour int) int {
	switch {
	case hour >= 23 || hour < 1:
		return 0
	default:
		return (hour-1)/2 + 1
	}
}

// ComputeLiushi computes the hour pillar for the given day and hour, and its
// interactions with the bazi chart.
func ComputeLiushi(date string, hour int, dayMaster Stem, bz ganzhi.Bazi) *LiushiResult {
	bazi := bz.Slice()
	y, m, d := 0, 0, 0
	fmt.Sscanf(date, "%d-%d-%d", &y, &m, &d)
	if y == 0 {
		return nil
	}

	dp := DayPillar(y, m, d)
	hbi := hourBranchIndex(hour)
	hourBranch := Branch(hbi + 1)
	hourStem := Stem(((int(dp.Stem)*2 + int(hourBranch) - 2) % 10))
	if hourStem == 0 {
		hourStem = 10
	}

	dmElem := StemElement(dayMaster)
	dmYY := StemYinYang(dayMaster)
	hElem := StemElement(hourStem)
	hYY := StemYinYang(hourStem)
	tgName := TenGodName(TenGodType(dmElem, dmYY, hElem, hYY))

	hourName := stemNameStr(hourStem) + branchNameStr(hourBranch)

	// Hour vs bazi: stem relations (same as liuri logic).
	var stemRels []StemRelation
	var branchRels []BranchRelation
	for _, np := range bazi {
		sr := AnalyzeStemRelation(hourStem, np.Stem)
		if sr.Type != "无" && sr.Type != "相同" {
			stemRels = append(stemRels, sr)
		}
		br := AnalyzeBranchRelation(hourBranch, np.Branch)
		if br.Type != "无" {
			branchRels = append(branchRels, br)
		}
	}

	return &LiushiResult{
		Time:       hourRanges[hbi],
		HourStem:   hourStem,
		HourBranch: hourBranch,
		HourName:   hourName,
		TenGod:     tgName,
		StemRels:   stemRels,
		BranchRels: branchRels,
	}
}
