package qiming

import (
	"fmt"
	"strings"
)

// Phonetic holds phonetic analysis.
type Phonetic struct {
	Tones    string `json:"tones"`
	IsPingZe bool   `json:"is_ping_ze"`
}

// analyzePhonetic analyzes tones and ping-ze for a character sequence.
func analyzePhonetic(chars []Character) Phonetic {
	if len(chars) == 0 {
		return Phonetic{}
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

	return Phonetic{
		Tones:    toneStr,
		IsPingZe: isAlternating,
	}
}
