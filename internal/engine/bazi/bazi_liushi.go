package bazi

import (
	"fmt"

	"liki/internal/engine/ganzhi"
	"liki/internal/engine/tianwen"
)

// LiuShi holds the hourly flow (流时) analysis: the current two-hour period's
// pillar and its interactions with the bazi chart.
type LiuShi struct {
	Time     string        `json:"time"`
	HourGan  ganzhi.Gan           `json:"hour_stem"`
	HourZhi  ganzhi.Zhi           `json:"hour_branch"`
	HourName string        `json:"hour_name"`
	TenGod   string        `json:"shishen"`
	GanRels  []GanRelation `json:"gan_rels"`
	ZhiRels  []ZhiRelation `json:"branch_rels"`
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

// ComputeLiuShi computes the hour pillar for the given day and hour, and its
// interactions with the bazi chart.
func ComputeLiuShi(date string, hour int, dayMaster ganzhi.Gan, bz ganzhi.Bazi) *LiuShi {
	y, m, d := 0, 0, 0
	if n, _ := fmt.Sscanf(date, "%d-%d-%d", &y, &m, &d); n != 3 { //nolint:errcheck
		return nil
	}

	dp := tianwen.DayPillar(y, m, d)
	hbi := hourBranchIndex(hour)
	hourBranch := ganzhi.Zhi(hbi + 1)
	hourStem := ganzhi.Gan(((int(dp.Gan)*2 + int(hourBranch) - 2) % 10))
	if hourStem == 0 {
		hourStem = 10
	}

	tgName := ganzhi.TenGodFromGan(dayMaster, hourStem)

	hourName := ganzhi.GanName(hourStem) + ganzhi.ZhiName(hourBranch)

	// Hour vs bazi: all 4 pillars, consistent with liunian.
	stemRels, branchRels := analyzePillarWithBazi(ganzhi.Zhu{Gan: hourStem, Zhi: hourBranch}, bz)

	return &LiuShi{
		Time:     ganzhi.HourRanges[hbi],
		HourGan:  hourStem,
		HourZhi:  hourBranch,
		HourName: hourName,
		TenGod:   tgName,
		GanRels:  stemRels,
		ZhiRels:  branchRels,
	}
}
