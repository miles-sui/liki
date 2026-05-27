package bazi

import (
	"fmt"

	"github.com/25types/25types/internal/ganzhi"
)

// LiuriResult holds the daily flow (流日) analysis: today's pillar
// and its interactions with the bazi chart, dayun, and liunian pillars.
type LiuriResult struct {
	Date         string           `json:"date"`
	DayStem      Stem             `json:"day_stem"`
	DayBranch    Branch           `json:"day_branch"`
	DayName      string           `json:"day_name"`
	DayNaYin     string           `json:"day_nayin"`
	TenGod       string           `json:"ten_god"`
	StemRels     []StemRelation   `json:"stem_rels"`
	BranchRels   []BranchRelation `json:"branch_rels"`
	DayunRels    []BranchRelation `json:"dayun_rels"`
	LiunianRels  []BranchRelation `json:"liunian_rels"`
	ShenSha      []ShenShaEntry   `json:"shensha"`
}

// ComputeLiuri computes the day pillar for the given date and its full
// interactions with the bazi chart, current dayun, and current liunian.
func ComputeLiuri(date string, dayMaster Stem, bz ganzhi.Bazi, dayunPillar *Pillar, liunianPillar *Pillar) *LiuriResult {
	bazi := bz.Slice()
	// Parse date.
	y, m, d := 0, 0, 0
	fmt.Sscanf(date, "%d-%d-%d", &y, &m, &d)
	if y == 0 {
		return nil
	}

	dp := DayPillar(y, m, d)
	dmElem := StemElement(dayMaster)
	dmYY := StemYinYang(dayMaster)
	dayElem := StemElement(dp.Stem)
	dayYY := StemYinYang(dp.Stem)

	tgName := TenGodName(TenGodType(dmElem, dmYY, dayElem, dayYY))

	dayName := stemNameStr(dp.Stem) + branchNameStr(dp.Branch)

	// Day vs bazi (stem + branch relations).
	var stemRels []StemRelation
	var branchRels []BranchRelation
	for _, np := range bazi {
		sr := AnalyzeStemRelation(dp.Stem, np.Stem)
		if sr.Type != "无" && sr.Type != "相同" {
			stemRels = append(stemRels, sr)
		}
		br := AnalyzeBranchRelation(dp.Branch, np.Branch)
		if br.Type != "无" {
			branchRels = append(branchRels, br)
		}
	}

	// Day vs dayun.
	var dayunRels []BranchRelation
	if dayunPillar != nil {
		br := AnalyzeBranchRelation(dp.Branch, dayunPillar.Branch)
		if br.Type != "无" {
			dayunRels = append(dayunRels, br)
		}
	}

	// Day vs liunian.
	var liunianRels []BranchRelation
	if liunianPillar != nil {
		br := AnalyzeBranchRelation(dp.Branch, liunianPillar.Branch)
		if br.Type != "无" {
			liunianRels = append(liunianRels, br)
		}
	}

	// Na yin.
	naYin := NaYinString(dp.Stem, dp.Branch)
	if naYin == "" {
		naYin = "未知"
	}

	// Daily shensha: day stem/branch vs bazi.
	var shensha []ShenShaEntry
	// 天乙贵人 on day stem.
	if targets, ok := tianYiLookup[int(dp.Stem)]; ok {
		for _, tb := range targets {
			for _, np := range bazi {
				if int(np.Branch) == tb {
					shensha = append(shensha, ShenShaEntry{Name: "天乙贵人", Category: "吉", Description: "流日天乙贵人日"})
				}
			}
		}
	}
	// 文昌 on day stem.
	if targets, ok := wenChangLookup[int(dp.Stem)]; ok {
		for _, tb := range targets {
			for _, np := range bazi {
				if int(np.Branch) == tb {
					shensha = append(shensha, ShenShaEntry{Name: "文昌", Category: "吉", Description: "流日文昌日，利学业文书"})
				}
			}
		}
	}
	// 驿马/桃花/华盖 from year branch triad → day branch check.
	yBranch := int(bazi[0].Branch)
	triadMaps := []struct {
		m    map[int]int
		name string
		cat  string
		desc string
	}{
		{yimaBranchMap, "驿马", "中性", "流日驿马，动象"},
		{taohuaBranchMap, "桃花", "中性", "流日桃花，异性缘佳"},
		{huagaiBranchMap, "华盖", "中性", "流日华盖，宜静思"},
	}
	for _, tm := range triadMaps {
		if tb, ok := tm.m[yBranch]; ok && int(dp.Branch) == tb {
			shensha = append(shensha, ShenShaEntry{Name: tm.name, Category: tm.cat, Description: tm.desc})
		}
	}

	return &LiuriResult{
		Date:        date,
		DayStem:     dp.Stem,
		DayBranch:   dp.Branch,
		DayName:     dayName,
		DayNaYin:    naYin,
		TenGod:      tgName,
		StemRels:    stemRels,
		BranchRels:  branchRels,
		DayunRels:   dayunRels,
		LiunianRels: liunianRels,
		ShenSha:     shensha,
	}
}
