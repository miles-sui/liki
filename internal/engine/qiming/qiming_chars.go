package qiming

import (
	"sort"
	"strings"
)

// Character is a naming character from the ben-hua general standard Chinese table.
type Character struct {
	Char        string `json:"char"`
	Element     Wuxing `json:"wuxing"`
	Stroke      int    `json:"stroke"`
	Radical     string `json:"radical"`
	Pinyin      string `json:"pinyin"`
	Tone        int    `json:"tone"`
	Traditional string `json:"traditional,omitempty"`
}

// CharLite is a lightweight character view for the HTTP chars endpoint.
type CharLite struct {
	Char string `json:"char"`
	Tone int    `json:"tone"`
}

func elementYAMLToChinese(e string) string {
	switch e {
	case "wood":
		return "木"
	case "fire":
		return "火"
	case "earth":
		return "土"
	case "metal":
		return "金"
	case "water":
		return "水"
	}
	return e
}

// lookupKangxiStroke returns the Kangxi dictionary stroke count for a character.
func lookupKangxiStroke(char string) int {
	rs := []rune(char)
	if len(rs) > 0 {
		if ce, ok := charByRune[rs[0]]; ok {
			return ce.Stroke
		}
	}
	return 0
}

// getCharsByElement returns all characters of the given element, grouped by stroke.
func getCharsByElement(elem Wuxing) map[int][]CharLite {
	chars := charByElement[elem]
	result := make(map[int][]CharLite)
	for _, c := range chars {
		result[c.Stroke] = append(result[c.Stroke], CharLite{Char: c.Char, Tone: c.Tone})
	}
	for _, v := range result {
		sort.Slice(v, func(i, j int) bool { return v[i].Char < v[j].Char })
	}
	return result
}

// characterName returns the combined given name string from characters.
func characterName(chars []Character) string {
	var b strings.Builder
	for _, c := range chars {
		b.WriteString(c.Char)
	}
	return b.String()
}
