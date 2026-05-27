package qiming

// EvaluateName evaluates a single given name against the surname, yong_shen and zodiac.
func EvaluateName(surname, givenName, yongShen string, zodiac Branch) NameEvaluation {
	runes := []rune(givenName)
	if len(runes) < 1 || len(runes) > 2 {
		runes = runes[:min(len(runes), 2)]
	}

	var givenStrs []string
	var charEntries []CharacterEntry
	for _, r := range runes {
		ch := string(r)
		givenStrs = append(givenStrs, ch)
		if ce, ok := defaultEngine.CharByRune[r]; ok {
			charEntries = append(charEntries, ce)
		} else {
			charEntries = append(charEntries, CharacterEntry{Char: ch})
		}
	}

	wg := ComputeWuGe(surname, givenStrs)
	sc := ComputeSanCai(wg.TianGe.Element, wg.RenGe.Element, wg.DiGe.Element)
	phon := AnalyzePhonetic(charEntries)

	wuxingMatch := false
	if yongShen != "" {
		for _, ce := range charEntries {
			if ce.Element.String() == yongShen {
				wuxingMatch = true
				break
			}
		}
	}

	var zodiacNotes []string
	if zodiac >= 1 && zodiac <= 12 {
		zh := ZodiacFromYearBranch(zodiac)
		for _, ce := range charEntries {
			for _, pref := range zh.PreferredStems {
				if ce.Radical == pref {
					zodiacNotes = append(zodiacNotes, ce.Char+"含宜用部首"+pref)
				}
			}
			for _, forb := range zh.ForbiddenStems {
				if ce.Radical == forb {
					zodiacNotes = append(zodiacNotes, ce.Char+"含忌用部首"+forb)
				}
			}
		}
	}

	return NameEvaluation{
		Surname:     surname,
		GivenName:   CharacterName(charEntries),
		Characters:  charEntries,
		WuGe:        wg,
		SanCai:      sc,
		Phonetic:    phon,
		WuxingMatch: wuxingMatch,
		ZodiacNotes: zodiacNotes,
	}
}
