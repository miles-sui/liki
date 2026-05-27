package qiming

import (
	"sort"
	"strings"
)

// CharacterEntry is a naming character from the ben-hua general standard Chinese table.
type CharacterEntry struct {
	Char       string  `json:"char"`
	Element    Element `json:"element"`
	Stroke     int     `json:"stroke"`
	Radical    string  `json:"radical"`
	Pinyin     string  `json:"pinyin"`
	Tone       int     `json:"tone"`
	Traditional string `json:"traditional,omitempty"`
}

func ElementYAMLToChinese(e string) string {
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

// LookupCharacterElement returns the element of a character by radical lookup.
func LookupCharacterElement(char string) Element {
	e := defaultEngine
	if e == nil {
		return 0
	}
	rs := []rune(char)
	if len(rs) > 0 {
		if ce, ok := e.CharByRune[rs[0]]; ok {
			return ce.Element
		}
	}
	return 0
}

// LookupSurnameElement returns the element of a surname character.
func LookupSurnameElement(surname string) Element {
	return LookupCharacterElement(surname)
}

// LookupKangxiStroke returns the Kangxi dictionary stroke count for a character.
func LookupKangxiStroke(char string) int {
	e := defaultEngine
	if e == nil {
		return 0
	}
	rs := []rune(char)
	if len(rs) > 0 {
		if ce, ok := e.CharByRune[rs[0]]; ok {
			return ce.Stroke
		}
	}
	return 0
}

// GetCharacterDB returns the full character database.
func GetCharacterDB() []CharacterEntry {
	e := defaultEngine
	if e == nil {
		return nil
	}
	return e.CharDB
}

// GetCharactersByElement returns characters filtered by element and stroke range.
func GetCharactersByElement(elem Element, strokeMin, strokeMax, limit int) []CharacterEntry {
	e := defaultEngine
	if e == nil {
		return nil
	}
	chars := e.CharByElement[elem]
	var result []CharacterEntry
	for _, c := range chars {
		if strokeMin > 0 && c.Stroke < strokeMin {
			continue
		}
		if strokeMax > 0 && c.Stroke > strokeMax {
			continue
		}
		result = append(result, c)
	}
	sortCharacters(result)
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result
}

func sortCharacters(chars []CharacterEntry) {
	sort.Slice(chars, func(i, j int) bool {
		if chars[i].Stroke != chars[j].Stroke {
			return chars[i].Stroke < chars[j].Stroke
		}
		return chars[i].Pinyin < chars[j].Pinyin
	})
}

// CharacterName returns the combined given name string from characters.
func CharacterName(chars []CharacterEntry) string {
	var b strings.Builder
	for _, c := range chars {
		b.WriteString(c.Char)
	}
	return b.String()
}
