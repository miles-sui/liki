package qiming

import "fmt"

// PrepareWuGe enumerates stroke combos and character candidates for a surname.
func PrepareWuGe(surname, yongShen string, xiShen []string) (WuGeData, error) {
	strokes := lookupKangxiStroke(surname)
	if strokes == 0 {
		return WuGeData{}, fmt.Errorf("qiming: surname %q not found in Kangxi database", surname)
	}

	enum := enumWuGeCombinations(strokes)

	yongElem := wuxingFromChinese(yongShen)
	yongChars := getCharsByElement(yongElem)

	xiChars := make(map[int][]CharLite)
	for _, xs := range xiShen {
		for stroke, chars := range getCharsByElement(wuxingFromChinese(xs)) {
			xiChars[stroke] = append(xiChars[stroke], chars...)
		}
	}

	return WuGeData{
		Surname:   surname,
		Combos:    enum.Combinations,
		YongChars: yongChars,
		XiChars:   xiChars,
	}, nil
}

// hasNegativeChar checks whether a name contains any character with negative meaning.
func hasNegativeChar(name string) bool {
	for _, r := range name {
		if negativeChars[string(r)] {
			return true
		}
	}
	return false
}

// ComposeNames builds name strings from character pools across all combos.
// Only yong+yong, yong+xi, and xi+yong pairs are allowed. Names that fail
// phonetic validation or contain negative characters are filtered out.
func ComposeNames(surname string, combos []StrokeCombo, yongChars, xiChars map[int][]CharLite) []string {
	seen := make(map[string]bool)
	var names []string

	for _, combo := range combos {
		yong1 := yongChars[combo.Stroke1]
		yong2 := yongChars[combo.Stroke2]
		xi1 := xiChars[combo.Stroke1]
		xi2 := xiChars[combo.Stroke2]

		pairs := [][2][]CharLite{
			{yong1, yong2}, // yong+yong
			{yong1, xi2},   // yong+xi
			{xi1, yong2},   // xi+yong
		}
		for _, p := range pairs {
			for _, c1 := range p[0] {
				for _, c2 := range p[1] {
					name := surname + c1.Char + c2.Char
					if seen[name] {
						continue
					}
					if hasNegativeChar(name) {
						continue
					}
					names = append(names, name)
					seen[name] = true
				}
			}
		}
	}
	return names
}

// DetailNames returns the full five-grid, three-talent, and phonetic analysis
// for a batch of given names sharing the same surname.
func DetailNames(surname string, names []string) ([]NameCandidate, error) {
	surnameStrokes := lookupKangxiStroke(surname)
	if surnameStrokes == 0 {
		return nil, fmt.Errorf("detail names: surname %q not found in Kangxi dictionary", surname)
	}

	var results []NameCandidate
	for _, fullName := range names {
		fullRunes := []rune(fullName)
		surnameRunes := len([]rune(surname))
		if len(fullRunes) <= surnameRunes {
			continue
		}
		given := string(fullRunes[surnameRunes:])
		rs := []rune(given)
		if len(rs) != 2 {
			continue
		}
		ce1, ok1 := charByRune[rs[0]]
		ce2, ok2 := charByRune[rs[1]]
		if !ok1 || !ok2 {
			continue
		}
		s1, s2 := ce1.Stroke, ce2.Stroke

		wg := computeWuGeFromStrokes(surnameStrokes, s1, s2)
		sc := computeSanCai(wg.TianGe.Element, wg.RenGe.Element, wg.DiGe.Element)
		phon := Phonetic{
			Tones:    formatTones(ce1.Tone, ce2.Tone),
		}

		results = append(results, NameCandidate{
			Name:       fullName,
			Characters: []Character{ce1, ce2},
			WuGe:       wg,
			SanCai:     sc,
			Phonetic:   phon,
		})
	}
	return results, nil
}

// computeWuGeFromStrokes computes the five-grid analysis from raw stroke counts.
func computeWuGeFromStrokes(surnameStroke, s1, s2 int) WuGe {
	tian := surnameStroke + 1
	ren := surnameStroke + s1
	di := s1 + s2
	if s2 == 0 {
		di = s1 + 1 // 单字名地格 = 名笔画 + 1
	}
	zong := surnameStroke + s1 + s2
	wai := zong - ren + 1
	if wai < 1 {
		wai = 1
	}
	return WuGe{
		TianGe: strokeResult(tian),
		RenGe:  strokeResult(ren),
		DiGe:   strokeResult(di),
		WaiGe:  strokeResult(wai),
		ZongGe: strokeResult(zong),
	}
}



func formatTones(t1, t2 int) string {
	return fmt.Sprintf("%d-%d", t1, t2)
}
