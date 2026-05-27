package minglihttp

import (
	"github.com/25types/25types/internal/mingli/bazi"
)

// ---- BaZi chart response types ----

// ChartResponse is the typed response for POST /api/bazi/chart.
type ChartResponse = bazi.ChartOutput

// BondResponse is the typed response for POST /api/bazi/bond.
type BondResponse = bazi.BondOutput

// ---- Solar terms response ----

// SolarTermMonth is one solar term month in the calendar.
type SolarTermMonth struct {
	ID     string `json:"id"`
	NameEN string `json:"name_en"`
	Start  string `json:"start"`
	End    string `json:"end"`
}

// SolarTermsResponse is the typed response for GET /api/solar-terms.
type SolarTermsResponse struct {
	Year    int              `json:"year"`
	Current SolarTermMonth   `json:"current"`
	Months  []SolarTermMonth `json:"months"`
}

// ---- Zodiac reference response ----

// ZodiacPairEntry is a pair of zodiac branches.
type ZodiacPairEntry struct {
	A     int    `json:"a"`
	B     int    `json:"b"`
	AName string `json:"a_name"`
	BName string `json:"b_name"`
}

// ZodiacTripleEntry is a triple of zodiac branches.
type ZodiacTripleEntry struct {
	Branches []int    `json:"branches"`
	Names    []string `json:"names"`
	Element  string   `json:"element"`
}

// ZodiacXingEntry is a xing (punishment) group.
type ZodiacXingEntry struct {
	Type     string   `json:"type"`
	Branches []int    `json:"branches"`
	Names    []string `json:"names"`
}

// ZodiacResponse is the typed response for GET /api/reference/zodiac.
type ZodiacResponse struct {
	SixHe     []ZodiacPairEntry   `json:"six_he"`
	TripleHe  []ZodiacTripleEntry `json:"triple_he"`
	TripleHui []ZodiacTripleEntry `json:"triple_hui"`
	SixChong  []ZodiacPairEntry   `json:"six_chong"`
	SixHai    []ZodiacPairEntry   `json:"six_hai"`
	Xing      []ZodiacXingEntry   `json:"xing"`
}
