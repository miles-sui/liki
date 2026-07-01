package qiming

// NameCandidate is a single name proposal.
type NameCandidate struct {
	Name       string           `json:"name"`
	Characters []Character `json:"characters"`
	WuGe       WuGe             `json:"wu_ge"`
	SanCai     SanCai           `json:"san_cai"`
	Phonetic   Phonetic     `json:"phonetic"`
}

// Evaluation is the output of evaluating a single name.
type Evaluation struct {
	Name        string           `json:"name,omitempty"`
	Surname     string           `json:"surname"`
	GivenName   string           `json:"given_name"`
	Characters  []Character `json:"characters"`
	WuGe        WuGe             `json:"wu_ge"`
	SanCai      SanCai           `json:"san_cai"`
	Phonetic    Phonetic     `json:"phonetic"`
	WuxingMatch bool             `json:"wuxing_match"`
	Wuxing      *struct {
		Yong bool `json:"yong"`
		Xi   bool `json:"xi,omitempty"`
		Ji   bool `json:"ji,omitempty"`
	} `json:"wuxing,omitempty"`
}

// StrokeCombo is one auspicious stroke1+stroke2 pair.
type StrokeCombo struct {
	Stroke1 int    `json:"stroke1"`
	Stroke2 int    `json:"stroke2"`
	SanCai  string `json:"san_cai"`
	Fortune string `json:"fortune"`
}

// wuGeEnumerationResult is the output of enumerating all auspicious
// stroke combinations for a given surname stroke count.
type wuGeEnumerationResult struct {
	SurnameStrokes int                 `json:"surname_strokes"`
	TianGe         tianGeBrief         `json:"tian_ge"`
	Combinations   []StrokeCombo `json:"combinations"`
}

// tianGeBrief is a compact view of TianGe.
type tianGeBrief struct {
	Stroke int    `json:"stroke"`
	Wuxing string `json:"wuxing"`
}

// WuGeData is the output of PrepareWuGe.
type WuGeData struct {
	Surname   string                     `json:"surname"`
	Combos    []StrokeCombo        `json:"combos"`
	YongChars map[int][]CharLite        `json:"yong_chars"`
	XiChars   map[int][]CharLite        `json:"xi_chars"`
}
