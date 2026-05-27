package bazi

import (
	"time"

	"github.com/25types/25types/internal/ganzhi"
)

// LiuyueResult holds the monthly (流月) analysis output.
type LiuyueResult struct {
	Year        int              `json:"year"`
	Month       int              `json:"month"`
	MonthStem   Stem             `json:"month_stem"`
	MonthBranch Branch           `json:"month_branch"`
	MonthName   string           `json:"month_name"`
	Element     string           `json:"element"`
	TenGod      string           `json:"ten_god"`
	Generates   int              `json:"generates"`
	Restrains   int              `json:"restrains"`
	StemRels    []StemRelation   `json:"stem_rels"`
	BranchRels  []BranchRelation `json:"branch_rels"`
	ShenSha     []ShenShaEntry   `json:"shensha"`
}

// ComputeLiuyue computes the month pillar for a given year+month and analyzes
// its interactions with the bazi chart.
func ComputeLiuyue(year, month int, dayMaster Stem, bz ganzhi.Bazi) *LiuyueResult {
	bazi := bz.Slice()
	// Use mid-month (15th) for stable solar term calculation.
	birthTime := time.Date(year, time.Month(month), 15, 12, 0, 0, 0, time.UTC)

	// Year pillar for the WuHuDun formula.
	yp := YearPillar(year, month, 15)
	mp := MonthPillar(birthTime, yp.Stem)

	dmElem := StemElement(dayMaster)
	monthElem := StemElement(mp.Stem)
	dmYY := StemYinYang(dayMaster)
	monthYY := StemYinYang(mp.Stem)

	tgName := TenGodName(TenGodType(dmElem, dmYY, monthElem, monthYY))

	gen, rest := 0, 0
	if Sheng(monthElem, dmElem) {
		gen = 1
	}
	if Ke(monthElem, dmElem) {
		rest = 1
	}

	// Month vs bazi: stem + branch relations.
	var stemRels []StemRelation
	var branchRels []BranchRelation
	for _, np := range bazi {
		sr := AnalyzeStemRelation(mp.Stem, np.Stem)
		if sr.Type != "无" && sr.Type != "相同" {
			stemRels = append(stemRels, sr)
		}
		br := AnalyzeBranchRelation(mp.Branch, np.Branch)
		if br.Type != "无" {
			branchRels = append(branchRels, br)
		}
	}

	shensha := ComputeDynamicShenSha(mp.Branch, bazi[0].Branch, dayMaster)

	monthElemStr := monthElem.String()

	return &LiuyueResult{
		Year:        year,
		Month:       month,
		MonthStem:   mp.Stem,
		MonthBranch: mp.Branch,
		MonthName:   branchNameStr(mp.Branch) + "月",
		Element:     monthElemStr,
		TenGod:      tgName,
		Generates:   gen,
		Restrains:   rest,
		StemRels:    stemRels,
		BranchRels:  branchRels,
		ShenSha:     shensha,
	}
}
