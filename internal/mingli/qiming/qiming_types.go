package qiming

// ZodiacHint provides zodiac-based naming advice.
type ZodiacHint struct {
	Animal           string   `json:"animal"`
	PreferredStems   []string `json:"preferred_radicals"`
	ForbiddenStems   []string `json:"forbidden_radicals"`
}

// NamingAnalysis holds the BaZi-driven naming analysis input.
type NamingAnalysis struct {
	Surname    string     `json:"surname"`
	YongShen   string     `json:"yong_shen"`
	XiShen     []string   `json:"xi_shen"`
	ZodiacHint ZodiacHint `json:"zodiac_hint"`
}

// NameCandidate is a single name proposal.
type NameCandidate struct {
	Name       string                `json:"name"`
	Characters []CharacterEntry `json:"characters"`
	WuGe       WuGe           `json:"wu_ge"`
	SanCai     SanCai         `json:"san_cai"`
	Phonetic   PhoneticMark   `json:"phonetic"`
	Highlights []string              `json:"highlights"`
}

// NamingResult is the complete naming analysis output.
type NamingResult struct {
	Analysis   NamingAnalysis  `json:"analysis"`
	Candidates []NameCandidate `json:"candidates"`
}

// NameEvaluation is the output of evaluating a single name.
type NameEvaluation struct {
	Surname     string                `json:"surname"`
	GivenName   string                `json:"given_name"`
	Characters  []CharacterEntry `json:"characters"`
	WuGe        WuGe           `json:"wu_ge"`
	SanCai      SanCai         `json:"san_cai"`
	Phonetic    PhoneticMark   `json:"phonetic"`
	WuxingMatch bool                  `json:"wuxing_match"`
	ZodiacNotes []string              `json:"zodiac_notes"`
}
