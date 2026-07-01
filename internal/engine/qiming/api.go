// Package qiming provides起名 computation.
//
// Types
//   Wuxing
//   WuGeData
//   NameCandidate, Evaluation
//   WuGe, SanCai
//   StrokeCombo
//   Character, CharLite, Phonetic
//
// Functions
//   EnumerateSancai(surname) → ([]StrokeCombo, error)
//   GetChars(wuxing, strokeMin, strokeMax) → (map[int][]CharLite, error)
//   PrepareWuGe(surname, yongShen, xiShen) → (WuGeData, error)
//   ComposeNames(surname, combos, yongChars, xiChars) → []string
//   DetailNames(surname, names) → ([]NameCandidate, error)
//   EvaluateName(surname, given, yong) → (Evaluation, error)
//   EvaluateNames(surname, names, yong, xi, ji) → ([]Evaluation, error)
package qiming

import (
	"fmt"
)

// EnumerateSancai returns all auspicious stroke combos for a surname.
func EnumerateSancai(surname string) ([]StrokeCombo, error) {
	strokes := lookupKangxiStroke(surname)
	if strokes == 0 {
		return nil, fmt.Errorf("sancai: surname %q not found in Kangxi dictionary", surname)
	}
	enum := enumWuGeCombinations(strokes)
	return enum.Combinations, nil
}

// GetChars returns naming characters of the given element, grouped by stroke.
// strokeMin and strokeMax are inclusive bounds. 0 means unbounded.
func GetChars(wuxingName string, strokeMin, strokeMax int) (map[int][]CharLite, error) {
	elem := wuxingFromChinese(wuxingName)
	if elem == 0 {
		return nil, fmt.Errorf("chars: invalid wuxing %q", wuxingName)
	}
	raw := getCharsByElement(elem)
	if strokeMin <= 0 && strokeMax <= 0 {
		return raw, nil
	}
	filtered := make(map[int][]CharLite)
	for stroke, chars := range raw {
		if (strokeMin <= 0 || stroke >= strokeMin) && (strokeMax <= 0 || stroke <= strokeMax) {
			filtered[stroke] = chars
		}
	}
	return filtered, nil
}

// EvaluateNames evaluates a batch of names with full wuxing analysis.
func EvaluateNames(surname string, names []string, yongShen string, xiShen, jiShen []string) ([]Evaluation, error) {
	surnameStrokes := lookupKangxiStroke(surname)
	if surnameStrokes == 0 {
		return nil, fmt.Errorf("evaluate names: surname %q not found in Kangxi dictionary", surname)
	}
	surnameRunes := len([]rune(surname))

	yongElem := Wuxing(0)
	if yongShen != "" {
		yongElem = wuxingFromChinese(yongShen)
	}
	xiElems := make([]Wuxing, len(xiShen))
	for i, xs := range xiShen {
		xiElems[i] = wuxingFromChinese(xs)
	}
	jiElems := make([]Wuxing, len(jiShen))
	for i, js := range jiShen {
		jiElems[i] = wuxingFromChinese(js)
	}

	var results []Evaluation
	for _, fullName := range names {
		fullRunes := []rune(fullName)
		if len(fullRunes) <= surnameRunes {
			continue
		}
		given := string(fullRunes[surnameRunes:])
		rs := []rune(given)
		if len(rs) < 1 || len(rs) > 2 {
			continue
		}

		var charEntries []Character
		for _, r := range rs {
			ce, ok := charByRune[r]
			if !ok {
				continue
			}
			charEntries = append(charEntries, ce)
		}
		if len(charEntries) != len(rs) {
			continue
		}

		s1, s2 := charEntries[0].Stroke, 0
		if len(charEntries) > 1 {
			s2 = charEntries[1].Stroke
		}
		wg := computeWuGeFromStrokes(surnameStrokes, s1, s2)
		sc := computeSanCai(wg.TianGe.Element, wg.RenGe.Element, wg.DiGe.Element)
		phon := analyzePhonetic(charEntries)

		wuxingMatch := false
		if yongElem != 0 {
			for _, ce := range charEntries {
				if ce.Element == yongElem {
					wuxingMatch = true
					break
				}
			}
		}

		ev := Evaluation{
			Name:        surname + given,
			Surname:     surname,
			GivenName:   given,
			Characters:  charEntries,
			WuGe:        wg,
			SanCai:      sc,
			Phonetic:    phon,
			WuxingMatch: wuxingMatch,
		}

		if yongShen != "" {
			wx := &struct {
				Yong bool `json:"yong"`
				Xi   bool `json:"xi,omitempty"`
				Ji   bool `json:"ji,omitempty"`
			}{}
			for _, ce := range charEntries {
				if ce.Element == yongElem {
					wx.Yong = true
				}
			}
			if len(xiElems) > 0 {
				for _, ce := range charEntries {
					for _, xe := range xiElems {
						if ce.Element == xe {
							wx.Xi = true
							break
						}
					}
					if wx.Xi {
						break
					}
				}
			}
			if len(jiElems) > 0 {
				for _, ce := range charEntries {
					for _, je := range jiElems {
						if ce.Element == je {
							wx.Ji = true
							break
						}
					}
					if wx.Ji {
						break
					}
				}
			}
			ev.Wuxing = wx
		}

		results = append(results, ev)
	}

	return results, nil
}
