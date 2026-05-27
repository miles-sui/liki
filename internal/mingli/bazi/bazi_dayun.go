package bazi

// DayunPillar is a single big fortune pillar with human-readable labels.
type DayunPillar struct {
	Stem     Stem   `json:"stem"`
	Branch   Branch `json:"branch"`
	AgeStart int    `json:"age_start"`
	AgeEnd   int    `json:"age_end"`
	Name     string `json:"name"`
	Element  string `json:"element"`
	TenGod   string `json:"ten_god"`
}

// DayunResult holds the formatted big fortune output.
type DayunResult struct {
	StartAge            int                `json:"start_age"`
	Direction           string             `json:"direction"`
	Pillars             []DayunPillar      `json:"pillars"`
	CurrentPillarIndex  int                `json:"current_pillar_index"`
	DayunInteractions   []PillarInteraction `json:"dayun_interactions"`
	CurrentShenSha        []ShenShaEntry     `json:"current_shensha"`
}

// ComputeDayunResult converts raw DayunPillars into labeled DayunPillar slice.
// Each pillar gets a name (干支组合), element, and the ten god relationship
// between the pillar stem and the day master.
func ComputeDayunResult(bf DayunPillars, dayMaster Stem, birthYear, currentYear int, bz Bazi) *DayunResult {
	r := &DayunResult{
		StartAge:  bf.StartAge,
		Direction: bf.Direction,
	}

	dmElem := StemElement(dayMaster)
	dmYY := StemYinYang(dayMaster)
	for i, p := range bf.Pillars {
		ageStart := bf.StartAge + i*10
		ageEnd := ageStart + 9
		name := stemNameStr(p.Stem) + branchNameStr(p.Branch)
		elem := StemElement(p.Stem).String()
		tg := tenGodName(dmElem, dmYY, p.Stem)

		r.Pillars = append(r.Pillars, DayunPillar{
			Stem:     p.Stem,
			Branch:   p.Branch,
			AgeStart: ageStart,
			AgeEnd:   ageEnd,
			Name:     name,
			Element:  elem,
			TenGod:   tg,
		})
	}

	r.CurrentPillarIndex = computeCurrentPillarIndex(birthYear, currentYear, r.Pillars)
	r.DayunInteractions = ComputeDayunInteractions(r.Pillars, bz)

	if r.CurrentPillarIndex >= 0 && r.CurrentPillarIndex < len(r.Pillars) {
		cp := r.Pillars[r.CurrentPillarIndex]
		r.CurrentShenSha = ComputeDynamicShenSha(cp.Branch, bz.Year.Branch, dayMaster)
	}

	return r
}

// computeCurrentPillarIndex returns the index (0-7) of the dayun pillar that
// covers the current year, or -1 if before the first or after the last.
func computeCurrentPillarIndex(birthYear, currentYear int, pillars []DayunPillar) int {
	currentAge := currentYear - birthYear
	for i, p := range pillars {
		if currentAge >= p.AgeStart && currentAge <= p.AgeEnd {
			return i
		}
	}
	return -1
}

func tenGodName(dmElem Element, dmYY YinYang, other Stem) string {
	otherElem := StemElement(other)
	otherYY := StemYinYang(other)
	tgName := TenGodName(TenGodType(dmElem, dmYY, otherElem, otherYY))
	if tgName != "" {
		return tgName + "运"
	}
	return "未知运"
}
