package bazi

import (
	"time"

	"liki/internal/engine/ganzhi"
	"liki/internal/engine/tianwen"
)

// LiuYue holds the monthly (流月) analysis output.
type LiuYue struct {
	Year      int            `json:"year"`
	Month     int            `json:"month"`
	MonthGan  ganzhi.Gan            `json:"month_stem"`
	MonthZhi  ganzhi.Zhi            `json:"month_branch"`
	MonthName string         `json:"month_name"`
	Element   string         `json:"wuxing"`
	TenGod    string         `json:"shishen"`
	Generates int            `json:"generates"`
	Restrains int            `json:"restrains"`
	GanRels   []GanRelation  `json:"gan_rels"`
	ZhiRels   []ZhiRelation  `json:"branch_rels"`
	ShenSha   []shenShaEntry `json:"shensha"`
}

// ComputeLiuYue computes the month pillar for a given year+month and analyzes
// its interactions with the bazi chart.
func ComputeLiuYue(year, month int, dayMaster ganzhi.Gan, bz ganzhi.Bazi) *LiuYue {
	bazi := bz.Slice()
	// Use mid-month (15th) for stable solar term calculation.
	birthTime := time.Date(year, time.Month(month), 15, 12, 0, 0, 0, time.UTC)

	// Year pillar for the WuHuDun formula.
	yp := tianwen.YearPillar(year, month, 15)
	mp := tianwen.MonthPillar(birthTime, yp.Gan)

	dmElem := ganzhi.GanWuxing(dayMaster)
	monthElem := ganzhi.GanWuxing(mp.Gan)

	tgName := ganzhi.TenGodFromGan(dayMaster, mp.Gan)

	gen, rest := countGenRest(monthElem, dmElem)

	// Month vs bazi: all 4 pillars, consistent with liunian.
	stemRels, branchRels := analyzePillarWithBazi(mp, bz)

	shensha := computeDynamicShenSha(mp.Zhi, bazi[0].Zhi, dayMaster)

	monthElemStr := monthElem.String()

	return &LiuYue{
		Year:      year,
		Month:     month,
		MonthGan:  mp.Gan,
		MonthZhi:  mp.Zhi,
		MonthName: ganzhi.ZhiName(mp.Zhi) + "月",
		Element:   monthElemStr,
		TenGod:    tgName,
		Generates: gen,
		Restrains: rest,
		GanRels:   stemRels,
		ZhiRels:   branchRels,
		ShenSha:   shensha,
	}
}
