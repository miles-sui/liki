package bazi

// LiunianResult holds the annual (流年) analysis output.
type LiunianResult struct {
	Year              int                `json:"year"`
	YearStem          Stem               `json:"year_stem"`
	YearBranch        Branch             `json:"year_branch"`
	YearName          string             `json:"year_name"`
	Element           string             `json:"element"`
	NaYin             string             `json:"nayin"`
	TenGod            string             `json:"ten_god"`
	Generates         int                `json:"generates"`
	Restrains         int                `json:"restrains"`
	NatalInteractions []PillarInteraction `json:"natal_interactions"`
	DayunInteractions []PillarInteraction `json:"dayun_interactions"`
	ShenSha           []ShenShaEntry      `json:"shensha"`
	FuYinFanYin       []FuYinFanYinEntry  `json:"fuyin_fanyin"`
}

// ComputeLiunian computes the year pillar for a given year and analyzes its
// relationship to the day master. When bazi and currentDayun are provided,
// it also computes three-layer interaction analysis.
func ComputeLiunian(year int, dayMaster Stem, bz Bazi, currentDayun *DayunPillar) *LiunianResult {
	yp := YearPillar(year, 6, 15) // mid-year avoids LiChun edge
	yearStem, yearBranch := yp.Stem, yp.Branch

	dmElem := StemElement(dayMaster)
	yearElem := StemElement(yearStem)
	dmYY := StemYinYang(dayMaster)
	yearYY := StemYinYang(yearStem)

	tgName := TenGodName(TenGodType(dmElem, dmYY, yearElem, yearYY))

	gen, rest := 0, 0
	if Sheng(yearElem, dmElem) {
		gen = 1
	}
	if Ke(yearElem, dmElem) {
		rest = 1
	}

	naYin := NaYinString(yearStem, yearBranch)
	if naYin == "" {
		naYin = "未知"
	}

	r := &LiunianResult{
		Year:       year,
		YearStem:   yearStem,
		YearBranch: yearBranch,
		YearName:   stemNameStr(yearStem) + branchNameStr(yearBranch),
		Element:    yearElem.String(),
		NaYin:      naYin,
		TenGod:     tgName,
		Generates:  gen,
		Restrains:  rest,
	}

	// Three-layer analysis when bazi chart and current dayun are available.
	liunianPillar := Pillar{Stem: yearStem, Branch: yearBranch}
	r.NatalInteractions = make([]PillarInteraction, 1)
	stemRels, branchRels := AnalyzePillarWithBazi(liunianPillar, bz)
	r.NatalInteractions[0] = PillarInteraction{
		PillarLabel: r.YearName,
		StemRels:    stemRels,
		BranchRels:  branchRels,
	}

	if currentDayun != nil {
		dyPillar := Pillar{Stem: currentDayun.Stem, Branch: currentDayun.Branch}
		dyStemRels, dyBranchRels := AnalyzePillarWithBazi(dyPillar, bz)
		r.DayunInteractions = []PillarInteraction{{
			PillarLabel: currentDayun.TenGod + "(" + currentDayun.Name + ")",
			StemRels:    dyStemRels,
			BranchRels:  dyBranchRels,
		}}
	}

	r.ShenSha = ComputeDynamicShenSha(yearBranch, bz.Year.Branch, dayMaster)
	r.FuYinFanYin = ComputeFuYinFanYin(liunianPillar, bz)

	return r
}

