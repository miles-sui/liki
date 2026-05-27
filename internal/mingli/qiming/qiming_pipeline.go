package qiming

import (
	"sort"
)

const maxCandidatePool = 2000

// GenerateCandidates runs the full naming pipeline and returns top candidates.
func GenerateCandidates(surname string, analysis NamingAnalysis, limit int) []NameCandidate {
	if limit <= 0 {
		limit = 20
	}

	// Step 1: Build candidate pool from yong_shen + xi_shen elements.
	pool := buildCandidatePool(analysis)

	// Step 2: Zodiac filter — remove characters with forbidden radicals.
	forbidden := analysis.ZodiacHint.ForbiddenStems
	if len(forbidden) > 0 {
		pool = filterByRadicals(pool, forbidden, true)
	}

	// Step 3: Generate combinations (姓 + 字1 + 字2).
	type candidate struct {
		Name       string
		Chars      []CharacterEntry
		WuGe       WuGe
		SanCai     SanCai
		Phonetic   PhoneticMark
		Highlights []string
		sortKey    int
	}
	var candidates []candidate

	for i, c1 := range pool {
		for j, c2 := range pool {
			name := surname + c1.Char + c2.Char
			chars := []CharacterEntry{c1, c2}
			wg := ComputeWuGe(surname, []string{c1.Char, c2.Char})

			sancai := ComputeSanCai(wg.TianGe.Element, wg.RenGe.Element, wg.DiGe.Element)

			phon := AnalyzePhonetic(chars)
			hl := buildHighlights(chars, wg, sancai, phon, analysis)

			key := sancaiSortKey(sancai.Fortune)*1000 + wugeJiCount(wg)*10
			if i == j {
				key += 1
			}

			candidates = append(candidates, candidate{
				Name: name, Chars: chars, WuGe: wg, SanCai: sancai,
				Phonetic: phon, Highlights: hl, sortKey: key,
			})
		}
		if len(candidates) > maxCandidatePool {
			break
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].sortKey < candidates[j].sortKey
	})

	if limit > len(candidates) {
		limit = len(candidates)
	}

	result := make([]NameCandidate, limit)
	for i := 0; i < limit; i++ {
		c := candidates[i]
		result[i] = NameCandidate{
			Name: c.Name, Characters: c.Chars, WuGe: c.WuGe,
			SanCai: c.SanCai, Phonetic: c.Phonetic, Highlights: c.Highlights,
		}
	}
	return result
}

func buildCandidatePool(analysis NamingAnalysis) []CharacterEntry {
	var pool []CharacterEntry
	seen := make(map[string]bool)

	yongElem := ElementFromChinese(analysis.YongShen)
	if yongElem != 0 {
		for _, c := range GetCharactersByElement(yongElem, 0, 0, 0) {
			if !seen[c.Char] {
				pool = append(pool, c)
				seen[c.Char] = true
			}
		}
	}
	for _, xs := range analysis.XiShen {
		elem := ElementFromChinese(xs)
		if elem == 0 {
			continue
		}
		for _, c := range GetCharactersByElement(elem, 0, 0, 0) {
			if !seen[c.Char] {
				pool = append(pool, c)
				seen[c.Char] = true
			}
		}
	}
	// Sort: common chars first.
	sortCharacters(pool)
	return pool
}

func filterByRadicals(pool []CharacterEntry, radicals []string, exclude bool) []CharacterEntry {
	radSet := make(map[string]bool)
	for _, r := range radicals {
		radSet[r] = true
	}
	var result []CharacterEntry
	for _, c := range pool {
		match := radSet[c.Radical]
		if exclude == !match {
			result = append(result, c)
		}
	}
	return result
}

func sancaiSortKey(fortune string) int {
	switch fortune {
	case "吉":
		return 0
	case "半吉":
		return 1
	default:
		return 2
	}
}

func wugeJiCount(wg WuGe) int {
	count := 0
	for _, g := range []GeResult{wg.TianGe, wg.RenGe, wg.DiGe, wg.WaiGe, wg.ZongGe} {
		if g.Fortune == "吉" {
			count++
		}
	}
	return count
}

func buildHighlights(chars []CharacterEntry, wg WuGe, sc SanCai, phon PhoneticMark, analysis NamingAnalysis) []string {
	var hl []string
	if sc.Fortune == "吉" {
		hl = append(hl, "三才全吉")
	} else if sc.Fortune == "半吉" {
		hl = append(hl, "三才一吉")
	}

	jiCount := wugeJiCount(wg)
	if jiCount >= 4 {
		hl = append(hl, "数理大吉")
	} else if jiCount >= 2 {
		hl = append(hl, "数理较吉")
	}

	if phon.IsPingZe {
		hl = append(hl, "平仄交替")
	}

	for _, c := range chars {
		if c.Element.String() == analysis.YongShen {
			hl = append(hl, "五行补益到位")
			break
		}
	}

	if len(hl) == 0 {
		hl = append(hl, "综合平衡")
	}
	return hl
}

// ZodiacFromYearBranch returns the zodiac hint for the given year branch.
// Data is loaded from data/zodiac.yaml at startup.
func ZodiacFromYearBranch(branch Branch) ZodiacHint {
	e := defaultEngine
	if e == nil {
		return ZodiacHint{}
	}
	if entry, ok := e.Zodiac[int(branch)]; ok {
		return ZodiacHint{
			Animal:         entry.Animal,
			PreferredStems: entry.Preferred,
			ForbiddenStems: entry.Forbidden,
		}
	}
	return ZodiacHint{}
}
