package qiming

import "fmt"

// EvaluateName evaluates a single given name against the surname and yong_shen.
func EvaluateName(surname, givenName, yongShen string) (Evaluation, error) {
	runes := []rune(givenName)
	if len(runes) < 1 || len(runes) > 2 {
		runes = runes[:min(len(runes), 2)]
	}

	var charEntries []Character
	for _, r := range runes {
		if ce, ok := charByRune[r]; ok {
			charEntries = append(charEntries, ce)
		} else {
			return Evaluation{}, fmt.Errorf("character %q not found in Kangxi stroke database", string(r))
		}
	}

	surnameStrokes := lookupKangxiStroke(surname)
	if surnameStrokes == 0 {
		return Evaluation{}, fmt.Errorf("surname %q not found in Kangxi stroke database", surname)
	}

	s1, s2 := 0, 0
	if len(charEntries) > 0 {
		s1 = charEntries[0].Stroke
	}
	if len(charEntries) > 1 {
		s2 = charEntries[1].Stroke
	}
	wg := computeWuGeFromStrokes(surnameStrokes, s1, s2)
	sc := computeSanCai(wg.TianGe.Element, wg.RenGe.Element, wg.DiGe.Element)
	phon := analyzePhonetic(charEntries)

	wuxingMatch := false
	if yongShen != "" {
		for _, ce := range charEntries {
			if ce.Element.String() == yongShen {
				wuxingMatch = true
				break
			}
		}
	}

	return Evaluation{
		Surname:     surname,
		GivenName:   characterName(charEntries),
		Characters:  charEntries,
		WuGe:        wg,
		SanCai:      sc,
		Phonetic:    phon,
		WuxingMatch: wuxingMatch,
	}, nil
}
