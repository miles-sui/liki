package bazi

import (
	"liki/internal/engine/ganzhi"
	"liki/internal/engine/tianwen"
)

// LiuNian holds the annual (流年) analysis output.
type LiuNian struct {
	Year              int                 `json:"year"`
	YearGan           ganzhi.Gan                 `json:"year_stem"`
	YearZhi           ganzhi.Zhi                 `json:"year_branch"`
	YearName          string              `json:"year_name"`
	Element           string              `json:"wuxing"`
	NaYin             string              `json:"nayin"`
	TenGod            string              `json:"shishen"`
	Generates         int                 `json:"generates"`
	Restrains         int                 `json:"restrains"`
	NatalInteractions []pillarInteraction `json:"natal_interactions"`
	DaYunInteractions []pillarInteraction `json:"dayun_interactions"`
	ShenSha           []shenShaEntry      `json:"shensha"`
	FuYinFanYin       []FuYinFanYin  `json:"fuyin_fanyin"`
}

// ComputeLiuNian computes the year pillar for a given year and analyzes its
// relationship to the day master. When bazi and currentDaYun are provided,
// it also computes three-layer interaction analysis.
func ComputeLiuNian(year int, dayMaster ganzhi.Gan, bz ganzhi.Bazi, currentDaYun *DaYunPillar) *LiuNian {
	yp := tianwen.YearPillar(year, 6, 15) // mid-year avoids LiChun edge
	yearStem, yearBranch := yp.Gan, yp.Zhi

	dmElem := ganzhi.GanWuxing(dayMaster)
	yearElem := ganzhi.GanWuxing(yearStem)

	tgName := ganzhi.TenGodFromGan(dayMaster, yearStem)

	gen, rest := countGenRest(yearElem, dmElem)

	naYin := ganzhi.NaYinLabel(yearStem, yearBranch)

	r := &LiuNian{
		Year:      year,
		YearGan:   yearStem,
		YearZhi:   yearBranch,
		YearName:  ganzhi.GanName(yearStem) + ganzhi.ZhiName(yearBranch),
		Element:   yearElem.String(),
		NaYin:     naYin,
		TenGod:    tgName,
		Generates: gen,
		Restrains: rest,
	}

	// Three-layer analysis when bazi chart and current dayun are available.
	liuNianPillar := ganzhi.Zhu{Gan: yearStem, Zhi: yearBranch}
	r.NatalInteractions = make([]pillarInteraction, 1)
	stemRels, branchRels := analyzePillarWithBazi(liuNianPillar, bz)
	r.NatalInteractions[0] = pillarInteraction{
		PillarLabel: r.YearName,
		GanRels:     stemRels,
		ZhiRels:     branchRels,
	}

	if currentDaYun != nil {
		dyPillar := ganzhi.Zhu{Gan: currentDaYun.Gan, Zhi: currentDaYun.Zhi}
		dyStemRels, dyBranchRels := analyzePillarWithBazi(dyPillar, bz)
		r.DaYunInteractions = []pillarInteraction{{
			PillarLabel: currentDaYun.TenGod + "(" + currentDaYun.Name + ")",
			GanRels:     dyStemRels,
			ZhiRels:     dyBranchRels,
		}}
	}

	r.ShenSha = computeDynamicShenSha(yearBranch, bz.Nian.Zhi, dayMaster)
	r.FuYinFanYin = computeFuYinFanYin(liuNianPillar, bz)

	return r
}

func countGenRest(elem, dmElem ganzhi.Wuxing) (gen, rest int) { return 0, 0 } // TODO: implement
