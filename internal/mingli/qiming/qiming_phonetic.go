package qiming

import (
	"fmt"
	"strings"
)

// PhoneticMark holds phonetic analysis.
type PhoneticMark struct {
	Tones         string `json:"tones"`
	IsPingZe      bool   `json:"is_ping_ze"`
	HasHomophone  bool   `json:"has_homophone"`
	HomophoneNote string `json:"homophone_note,omitempty"`
}

// AnalyzePhonetic analyzes tones and ping-ze for a character sequence.
func AnalyzePhonetic(chars []CharacterEntry) PhoneticMark {
	if len(chars) == 0 {
		return PhoneticMark{}
	}
	var tones []string
	for _, c := range chars {
		tones = append(tones, fmt.Sprintf("%d", c.Tone))
	}
	toneStr := strings.Join(tones, "-")

	isAlternating := true
	if len(chars) >= 2 {
		prevPing := chars[0].Tone <= 2
		for i := 1; i < len(chars); i++ {
			curPing := chars[i].Tone <= 2
			if prevPing == curPing {
				isAlternating = false
				break
			}
			prevPing = curPing
		}
	}

	return PhoneticMark{
		Tones:    toneStr,
		IsPingZe: isAlternating,
	}
}
