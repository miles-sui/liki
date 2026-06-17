package qiming

import "fmt"

// PrepareWuGe enumerates stroke combos and character candidates for a surname.
func PrepareWuGe(surname, yongShen string, xiShen []string) (WuGeData, error) {
	strokes := lookupKangxiStroke(surname)
	if strokes == 0 {
		return WuGeData{}, fmt.Errorf("surname %q not found in Kangxi database", surname)
	}

	enum := enumWuGeCombinations(strokes)

	yongElem := wuxingFromChinese(yongShen)
	yongChars := getCharsByElement(yongElem)

	var xiChars map[int][]CharLite
	if len(xiShen) > 0 {
		xiChars = getCharsByElement(wuxingFromChinese(xiShen[0]))
	}

	return WuGeData{
		Surname:   surname,
		Combos:    enum.Combinations,
		YongChars: yongChars,
		XiChars:   xiChars,
	}, nil
}

// ComposeNames builds name strings from character pools across all combos.
// Only yong+yong, yong+xi, and xi+yong pairs are allowed. Names that fail
// phonetic validation are filtered out. Returns only name strings.
func ComposeNames(surname string, combos []StrokeCombo, yongChars, xiChars map[int][]string) []string {
	seen := make(map[string]bool)
	var names []string

	for _, combo := range combos {
		yong1 := yongChars[combo.Stroke1]
		yong2 := yongChars[combo.Stroke2]
		xi1 := xiChars[combo.Stroke1]
		xi2 := xiChars[combo.Stroke2]

		pairs := [][2][]string{
			{yong1, yong2}, // yong+yong
			{yong1, xi2},   // yong+xi
			{xi1, yong2},   // xi+yong
		}
		for _, p := range pairs {
			for _, c1 := range p[0] {
				for _, c2 := range p[1] {
					t1 := lookupTone(c1)
					t2 := lookupTone(c2)
					if !isPhoneticValid(t1, t2) {
						continue
					}
					name := surname + c1 + c2
					if !seen[name] {
						names = append(names, name)
						seen[name] = true
					}
				}
			}
		}
	}
	return names
}

// DetailNames returns the full five-grid, three-talent, and phonetic analysis
// for a batch of given names sharing the same surname.
func DetailNames(surname string, names []string) []NameCandidate {
	surnameStrokes := lookupKangxiStroke(surname)
	if surnameStrokes == 0 {
		return nil
	}

	var results []NameCandidate
	for _, fullName := range names {
		given := fullName[len([]rune(surname)):]
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
			IsPingZe: isPhoneticValid(ce1.Tone, ce2.Tone),
		}

		results = append(results, NameCandidate{
			Name:       fullName,
			Characters: []Character{ce1, ce2},
			WuGe:       wg,
			SanCai:     sc,
			Phonetic:   phon,
		})
	}
	return results
}

// computeWuGeFromStrokes computes the five-grid analysis from raw stroke counts.
func computeWuGeFromStrokes(surnameStroke, s1, s2 int) WuGe {
	tian := surnameStroke + 1
	ren := surnameStroke + s1
	di := s1 + s2
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

func lookupTone(char string) int {
	rs := []rune(char)
	if len(rs) > 0 {
		if ce, ok := charByRune[rs[0]]; ok {
			return ce.Tone
		}
	}
	return 0
}

func formatTones(t1, t2 int) string {
	return fmt.Sprintf("%d-%d", t1, t2)
}

func isAllPing(t1, t2 int) bool {
	return (t1 == 1 || t1 == 2) && (t2 == 1 || t2 == 2)
}

func isAllZe(t1, t2 int) bool {
	return (t1 == 3 || t1 == 4) && (t2 == 3 || t2 == 4)
}

func isAdjacentTone3(t1, t2 int) bool {
	return t1 == 3 && t2 == 3
}

func isPhoneticValid(t1, t2 int) bool {
	return !isAllPing(t1, t2) && !isAllZe(t1, t2) && !isAdjacentTone3(t1, t2)
}
