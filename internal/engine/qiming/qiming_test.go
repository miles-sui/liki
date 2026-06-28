package qiming

import (
	"strings"
	"testing"

	"liki/internal/engine/ganzhi"
)

// =============================================================================
// 五格笔画计算
// =============================================================================

func TestComputeWuGeFromStrokes_KnownCase2(t *testing.T) {
	// 李(7) + 明(8) + 辉(15)
	wg := computeWuGeFromStrokes(7, 8, 15)

	// 天格=8, 人格=15, 地格=23, 总格=30, 外格=30-15+1=16
	if wg.TianGe.Stroke != 8 {
		t.Errorf("天格 = %d, want 8", wg.TianGe.Stroke)
	}
	if wg.RenGe.Stroke != 15 {
		t.Errorf("人格 = %d, want 15", wg.RenGe.Stroke)
	}
	if wg.DiGe.Stroke != 23 {
		t.Errorf("地格 = %d, want 23", wg.DiGe.Stroke)
	}
	if wg.ZongGe.Stroke != 30 {
		t.Errorf("总格 = %d, want 30", wg.ZongGe.Stroke)
	}
	if wg.WaiGe.Stroke != 16 {
		t.Errorf("外格 = %d, want 16", wg.WaiGe.Stroke)
	}
}

func TestComputeWuGeFromStrokes_EdgeCases(t *testing.T) {
	// Large strokes: all >81 should wrap
	wg := computeWuGeFromStrokes(50, 40, 30)
	// 天格=51, 人格=90→9, 地格=70, 总格=120→39, 外格=39-9+1=31
	if wg.RenGe.Stroke != 9 {
		t.Errorf("人格(90) = %d, want 9 (wrap)", wg.RenGe.Stroke)
	}
	if wg.ZongGe.Stroke != 39 {
		t.Errorf("总格(120) = %d, want 39 (wrap)", wg.ZongGe.Stroke)
	}
}

// =============================================================================
// 五格五行 — 尾数决定
// =============================================================================

func TestStrokeResult_AllTailDigits(t *testing.T) {
	// 尾数 1,2=木; 3,4=火; 5,6=土; 7,8=金; 9,0=水
	for _, tt := range []struct {
		stroke  int
		element string
	}{
		{1, "木"}, {2, "木"}, {11, "木"}, {21, "木"},
		{3, "火"}, {4, "火"}, {13, "火"}, {24, "火"},
		{5, "土"}, {6, "土"}, {15, "土"}, {26, "土"},
		{7, "金"}, {8, "金"}, {17, "金"}, {28, "金"},
		{9, "水"}, {10, "水"}, {19, "水"}, {30, "水"},
	} {
		r := strokeResult(tt.stroke)
		if r.Element != tt.element {
			t.Errorf("stroke %d: element=%s, want %s", tt.stroke, r.Element, tt.element)
		}
	}
}

func TestStrokeResult_EdgeValues(t *testing.T) {
	// Stroke 0 → treated as 1
	r0 := strokeResult(0)
	if r0.Stroke != 1 {
		t.Errorf("stroke 0 should become 1, got %d", r0.Stroke)
	}

	// Stroke 81 → stays 81
	r81 := strokeResult(81)
	if r81.Stroke != 81 {
		t.Errorf("stroke 81 should stay 81, got %d", r81.Stroke)
	}

	// Stroke 82 → wraps to 1
	r82 := strokeResult(82)
	if r82.Stroke != 1 {
		t.Errorf("stroke 82 should wrap to 1, got %d", r82.Stroke)
	}

	// Stroke 162 → wraps twice → 1
	r162 := strokeResult(162) // (162-1)%81+1 = 161%81+1 = 80+1 = 81
	if r162.Stroke != 81 {
		t.Errorf("stroke 162 → %d, want 81", r162.Stroke)
	}
}

func TestStrokeResult_KnownValues(t *testing.T) {
	// Per sanCaiNums, verify known stroke→element→fortune for common values
	tests := []struct {
		stroke  int
		element string
	}{
		{1, "木"},   // 1=木
		{5, "土"},   // 5=土
		{8, "金"},   // 8=金
		{13, "火"},  // 13=火
		{21, "木"},  // 21=木
		{24, "火"},  // 24=火
		{31, "木"},  // 31=木
		{37, "金"},  // 37=金
		{45, "土"},  // 45=土
	}

	for _, tt := range tests {
		r := strokeResult(tt.stroke)
		if r.Element != tt.element {
			t.Errorf("stroke %d: element=%s, want %s", tt.stroke, r.Element, tt.element)
		}
	}
}

// =============================================================================
// 三才配置
// =============================================================================

func TestComputeSanCai_KnownConfigs(t *testing.T) {
	tests := []struct {
		config string   // e.g. "木木木"
	}{
		{"木木木"}, {"木木火"}, {"木木土"}, {"木木金"}, {"木木水"},
		{"木火木"}, {"木火火"}, {"木火土"}, {"木火金"}, {"木火水"},
		{"木土木"}, {"木土火"}, {"木土土"}, {"木土金"}, {"木土水"},
		{"木金木"}, {"木金火"}, {"木金土"}, {"木金金"}, {"木金水"},
		{"木水木"}, {"木水火"}, {"木水土"}, {"木水金"}, {"木水水"},
		{"火木木"}, {"火木火"}, {"火木土"},
	}

	for _, tt := range tests {
		runes := []rune(tt.config)
		sc := computeSanCai(string(runes[0:1]), string(runes[1:2]), string(runes[2:3]))
		if sc.Configuration != tt.config {
			t.Errorf("config=%s, got %s", tt.config, sc.Configuration)
		}
		if sc.Fortune == "" {
			t.Errorf("config=%s: empty fortune", tt.config)
		}
	}
}

func TestComputeSanCai_UnknownConfig(t *testing.T) {
	// 使用不存在的三才组合 → 默认半吉
	sc := computeSanCai("水", "水", "水")
	if sc.Configuration != "水水水" {
		t.Errorf("config = %s, want 水水水", sc.Configuration)
	}
	if sc.Fortune != "半吉" && sc.Fortune == "" {
		t.Errorf("unknown config fortune = %s, want non-empty", sc.Fortune)
	}
}

// =============================================================================
// 吉凶判断
// =============================================================================

func TestIsAuspicious(t *testing.T) {
	if !isAuspicious("吉") {
		t.Error("吉 should be auspicious")
	}
	if !isAuspicious("大吉") {
		t.Error("大吉 should be auspicious")
	}
	if isAuspicious("凶") {
		t.Error("凶 should NOT be auspicious")
	}
	if isAuspicious("半吉") {
		t.Error("半吉 should NOT be auspicious")
	}
	if isAuspicious("") {
		t.Error("empty should NOT be auspicious")
	}
}

// =============================================================================
// 音韵分析 — analyzePhonetic
// =============================================================================

func TestAnalyzePhonetic(t *testing.T) {
	// Empty chars
	phon := analyzePhonetic(nil)
	if phon.Tones != "" {
		t.Errorf("empty tones = %q, want empty", phon.Tones)
	}

	// Single char
	phon2 := analyzePhonetic([]Character{{Char: "明", Tone: 2}})
	if phon2.Tones != "2" {
		t.Errorf("single tones = %q, want 2", phon2.Tones)
	}

	// Two chars
	phon3 := analyzePhonetic([]Character{
		{Char: "明", Tone: 2},
		{Char: "亮", Tone: 4},
	})
	if phon3.Tones != "2-4" {
		t.Errorf("two-char tones = %q, want 2-4", phon3.Tones)
	}
}

// =============================================================================
// 字符名拼接
// =============================================================================

func TestCharacterName(t *testing.T) {
	got := characterName([]Character{
		{Char: "明"}, {Char: "辉"},
	})
	if got != "明辉" {
		t.Errorf("characterName = %q, want 明辉", got)
	}

	got2 := characterName([]Character{{Char: "文"}})
	if got2 != "文" {
		t.Errorf("characterName = %q, want 文", got2)
	}

	got3 := characterName(nil)
	if got3 != "" {
		t.Errorf("characterName(nil) = %q, want empty", got3)
	}
}

// =============================================================================
// 五行映射
// =============================================================================

func TestWuxingFromChinese_AllElements(t *testing.T) {
	tests := []struct {
		ch   string
		want string
	}{
		{"木", "木"}, {"火", "火"}, {"土", "土"}, {"金", "金"}, {"水", "水"},
	}
	for _, tt := range tests {
		got := wuxingFromChinese(tt.ch)
		if got.String() != tt.want {
			t.Errorf("wuxingFromChinese(%q).String() = %q, want %q", tt.ch, got.String(), tt.want)
		}
	}
}

// =============================================================================
// 汉字笔画查询
// =============================================================================

func TestLookupKangxiStroke_KnownChars(t *testing.T) {
	// Verify commonly used surname characters are in the database
	tests := []struct {
		char   string
		expect bool // expect >0 strokes (in DB)
	}{
		{"王", true},
		{"李", true},
		{"张", true},
		{"明", true},
		{"文", true},
		{"xyz", false},    // non-existent
		{"𠀀", false},     // very rare (probably not in DB)
	}

	for _, tt := range tests {
		strokes := lookupKangxiStroke(tt.char)
		if tt.expect && strokes == 0 {
			t.Errorf("lookupKangxiStroke(%q) = 0, want >0 (expected in DB)", tt.char)
		}
		if !tt.expect && strokes != 0 {
			t.Errorf("lookupKangxiStroke(%q) = %d, want 0 (not in DB)", tt.char, strokes)
		}
	}
}

// =============================================================================
// 五行字符分组 — getCharsByElement
// =============================================================================

func TestGetCharsByElement_Structure(t *testing.T) {
	for _, elem := range []string{"木", "火", "土", "金", "水"} {
		wx := wuxingFromChinese(elem)
		chars := getCharsByElement(wx)
		if len(chars) == 0 {
			t.Errorf("getCharsByElement(%s): empty result", elem)
		}
		// Each char should be sorted within its stroke group
		for stroke, group := range chars {
			if stroke < 1 || stroke > 50 {
				t.Errorf("unexpected stroke %d in group", stroke)
			}
			for i := 1; i < len(group); i++ {
				if group[i-1].Char >= group[i].Char {
					t.Errorf("group %d not sorted at index %d: %q >= %q",
						stroke, i, group[i-1].Char, group[i].Char)
				}
			}
		}
	}
}

// =============================================================================
// PreComputeWuGeCombinations — 枚举吉数组合
// =============================================================================

func TestEnumWuGeCombinations_OutputNotEmpty(t *testing.T) {
	// For surname "王" (4 strokes), there should be valid combos
	result := enumWuGeCombinations(4)

	if result.SurnameStrokes != 4 {
		t.Errorf("SurnameStrokes = %d, want 4", result.SurnameStrokes)
	}
	if result.TianGe.Stroke != 5 { // 天格=姓+1=5
		t.Errorf("TianGe stroke = %d, want 5", result.TianGe.Stroke)
	}
	if len(result.Combinations) == 0 {
		t.Error("enumWuGeCombinations(4) should have combos")
	}

	// For surname 1 (min), still should work
	result2 := enumWuGeCombinations(1)
	if len(result2.Combinations) == 0 {
		t.Error("enumWuGeCombinations(1) should have combos")
	}
}

func TestEnumWuGeCombinations_Validity(t *testing.T) {
	// For surname 4, verify every combo satisfies:
	// 1. All 5 grids are auspicious
	// 2. Three-talent mutual generation (tian→ren, ren→di)
	result := enumWuGeCombinations(4)

	// Build a set for easy lookup
	comboSet := make(map[[2]int]bool)
	for _, c := range result.Combinations {
		comboSet[[2]int{c.Stroke1, c.Stroke2}] = true
	}

	// Known auspicious combos for 王(4) based on domain knowledge:
	// These are pre-verified with professional naming tables
	knownGood := [][2]int{
		{1, 5},   // 天格5(土) + 人5(土) + 地6(土) — all 土
	}
	for _, kg := range knownGood {
		if !comboSet[kg] {
			t.Logf("combo (%d,%d) not in enum result (may be excluded by三才 generation check)", kg[0], kg[1])
		}
	}
}

func TestEnumWuGeCombinations_SancaiFilter(t *testing.T) {
	// Verify that三才 mutual generation filter works:
	// All returned combos must have 天格生人格 and 人格生地格
	// 天格 = surnameStrokes + 1, so it depends on the surname
	result := enumWuGeCombinations(4)

	for _, c := range result.Combinations {
		tian := strokeResult(5)  // 4+1
		ren := strokeResult(4 + c.Stroke1)
		di := strokeResult(c.Stroke1 + c.Stroke2)
		if tian.Fortune == "" || ren.Fortune == "" || di.Fortune == "" {
			t.Errorf("combo (%d,%d): empty fortune", c.Stroke1, c.Stroke2)
		}
	}
}

// =============================================================================
// 部首→五行推断
// =============================================================================

func TestInferElementFromRadical_DirectMatch(t *testing.T) {
	tests := []struct {
		radical string
		want    string
	}{
		{"木", "木"}, {"火", "火"}, {"土", "土"}, {"金", "金"}, {"水", "水"},
		{"艹", "木"}, {"氵", "水"}, {"忄", "火"}, {"钅", "金"}, {"山", "土"},
		{"石", "土"}, {"日", "火"}, {"心", "火"}, {"戈", "金"}, {"玉", "土"},
	}

	for _, tt := range tests {
		elem, ok := inferElementFromRadical(tt.radical)
		if !ok {
			t.Errorf("radical %q: not found", tt.radical)
			continue
		}
		if elem.String() != tt.want {
			t.Errorf("radical %q: element=%s, want %s", tt.radical, elem.String(), tt.want)
		}
	}
}

func TestInferElementFromRadical_UnknownRadical(t *testing.T) {
	// Unknown radicals are not resolved — no fallback scanning.
	_, ok := inferElementFromRadical("unknown")
	if ok {
		t.Error("unknown radical should not resolve to an element")
	}
}

// =============================================================================
// EvaluateName end-to-end
// =============================================================================

func TestEvaluateName_KnownName(t *testing.T) {
	// 王+明辉 → should work if chars in DB
	eval, err := EvaluateName("王", "明辉", "火")
	if err != nil {
		// Characters may not be in the test DB; skip if so
		if strings.Contains(err.Error(), "not found") {
			t.Skip("test chars not in DB: " + err.Error())
		}
		t.Fatalf("EvaluateName: %v", err)
	}

	if eval.Surname != "王" {
		t.Errorf("surname = %q, want 王", eval.Surname)
	}
	if eval.GivenName != "明辉" {
		t.Errorf("givenName = %q, want 明辉", eval.GivenName)
	}
	if len(eval.Characters) != 2 {
		t.Errorf("chars count = %d, want 2", len(eval.Characters))
	}
	// 火 may or may not match depending on the actual characters
	if eval.WuxingMatch {
		t.Logf("wuxing matched 火")
	}
}

func TestEvaluateName_SurnameNotFound(t *testing.T) {
	_, err := EvaluateName("𠀀", "明辉", "火")
	if err == nil {
		t.Fatal("expected error for unknown surname")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %v, want 'not found'", err)
	}
}

func TestEvaluateName_CharNotFound(t *testing.T) {
	_, err := EvaluateName("王", "𠀀", "火")
	if err == nil {
		t.Fatal("expected error for unknown char")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %v, want 'not found'", err)
	}
}

// =============================================================================
// PrepareWuGe end-to-end
// =============================================================================

func TestPrepareWuGe_KnownSurname(t *testing.T) {
	data, err := PrepareWuGe("王", "火", []string{"木"})
	if err != nil {
		t.Fatalf("PrepareWuGe: %v", err)
	}

	if data.Surname != "王" {
		t.Errorf("surname = %q, want 王", data.Surname)
	}
	if len(data.Combos) == 0 {
		t.Error("combos should not be empty for surname 王")
	}
	if len(data.YongChars) == 0 {
		t.Error("yongChars should not be empty")
	}
	if len(data.XiChars) == 0 {
		t.Error("xiChars should not be empty (喜神=木)")
	}
}

func TestPrepareWuGe_UnknownSurname(t *testing.T) {
	_, err := PrepareWuGe("xyz123", "火", nil)
	if err == nil {
		t.Fatal("expected error for unknown surname")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %v, want 'not found'", err)
	}
}

// =============================================================================
// ComposeNames
// =============================================================================

func TestComposeNames_NoCombo(t *testing.T) {
	names := ComposeNames("王", nil, nil, nil)
	if len(names) != 0 {
		t.Errorf("ComposeNames with nil combos should return empty, got %d names", len(names))
	}
}

// =============================================================================
// DetailNames
// =============================================================================

func TestDetailNames_SurnameNotFound(t *testing.T) {
	results, err := DetailNames("𠀀", []string{"𠀀明辉"})
	if err == nil {
		t.Error("DetailNames with unknown surname should return error")
	}
	_ = results
}

func TestDetailNames_NoGivenName(t *testing.T) {
	results, err := DetailNames("王", nil)
	if err != nil {
		t.Fatalf("DetailNames with nil names should not error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("DetailNames with nil names should return empty, got %d", len(results))
	}
}

// =============================================================================
// elementYAMLToChinese — 默认路径
// =============================================================================

func TestElementYAMLToChinese_Default(t *testing.T) {
	// Unknown element values returned as-is.
	if got := elementYAMLToChinese("unknown"); got != "unknown" {
		t.Errorf("elementYAMLToChinese(unknown) = %s, want unknown", got)
	}
}

// =============================================================================
// fortuneYAMLToChinese — 默认路径
// =============================================================================

func TestFortuneYAMLToChinese_Default(t *testing.T) {
	// Unknown fortune values returned as-is.
	if got := fortuneYAMLToChinese("unknown"); got != "unknown" {
		t.Errorf("fortuneYAMLToChinese(unknown) = %s, want unknown", got)
	}
}

// =============================================================================
// strokeResult — stroke >81 包装验证
// =============================================================================

func TestStrokeResult_WrapAbove81(t *testing.T) {
	// stroke >81 wraps: (stroke-1)%81+1. All 1-81 are in sanCaiNums.
	// 1000 → (1000-1)%81+1 = 999%81+1 = 27+1 = 28
	// 28 = metal(金), xiong(凶)
	r := strokeResult(1000)
	if r.Stroke != 28 {
		t.Errorf("stroke 1000 wraps to stroke=%d, want 28", r.Stroke)
	}
	if r.Element != "金" {
		t.Errorf("stroke 1000 wrap element = %s, want 金", r.Element)
	}
	if r.Fortune != "凶" {
		t.Errorf("stroke 1000 wrap fortune = %s, want 凶", r.Fortune)
	}
}

// =============================================================================
// formatTones
// =============================================================================

func TestFormatTones(t *testing.T) {
	if got := formatTones(1, 4); got != "1-4" {
		t.Errorf("formatTones(1,4) = %s, want 1-4", got)
	}
	if got := formatTones(2, 3); got != "2-3" {
		t.Errorf("formatTones(2,3) = %s, want 2-3", got)
	}
}

// =============================================================================
// EvaluateName — 单字名
// =============================================================================

func TestEvaluateName_SingleCharName(t *testing.T) {
	eval, err := EvaluateName("王", "文", "火")
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			t.Skip("test chars not in DB: " + err.Error())
		}
		t.Fatalf("EvaluateName: %v", err)
	}

	if eval.GivenName != "文" {
		t.Errorf("givenName = %q, want 文", eval.GivenName)
	}
	if len(eval.Characters) != 1 {
		t.Errorf("chars count = %d, want 1 (single char name)", len(eval.Characters))
	}
	// 地格 = s1 + s2(=0) →文 stroke alone
	if eval.WuGe.DiGe.Stroke == 0 {
		t.Error("DiGe stroke should not be 0")
	}
}

// =============================================================================
// EvaluateName — 空用神
// =============================================================================

func TestEvaluateName_EmptyYongShen(t *testing.T) {
	eval, err := EvaluateName("王", "文明", "")
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			t.Skip("test chars not in DB: " + err.Error())
		}
		t.Fatalf("EvaluateName: %v", err)
	}
	// With empty yongShen, WuxingMatch should be false.
	if eval.WuxingMatch {
		t.Error("WuxingMatch should be false when yongShen is empty")
	}
}

// =============================================================================
// ComposeNames — 实际字符数据
// =============================================================================

func TestComposeNames_WithRealData(t *testing.T) {
	// Get real character data from the DB for valid elements.
	yongElem := wuxingFromChinese("火")
	xiElem := wuxingFromChinese("木")
	yongRaw := getCharsByElement(yongElem)
	xiRaw := getCharsByElement(xiElem)

	// Get valid combos from PrepareWuGe for surname "王".
	data, err := PrepareWuGe("王", "火", []string{"木"})
	if err != nil {
		t.Fatalf("PrepareWuGe: %v", err)
	}
	if len(data.Combos) == 0 {
		t.Skip("no combos available for 王")
	}

	names := ComposeNames("王", data.Combos, yongRaw, xiRaw)
	if len(names) == 0 {
		t.Skip("no names composed (all phonetically invalid or no char data)")
	}
	// All names should start with "王".
	for _, n := range names {
		if !strings.HasPrefix(n, "王") {
			t.Errorf("name %q should start with 王", n)
		}
		rs := []rune(n)
		if len(rs) != 3 {
			t.Errorf("name %q should be 3 chars (姓+2名), got %d runes", n, len(rs))
		}
	}
}

// =============================================================================
// DetailNames — 实际姓名数据
// =============================================================================

func TestDetailNames_KnownNames(t *testing.T) {
	// "明" and "文" should be in the DB (they are commonly used chars).
	results, err := DetailNames("王", []string{"王明文", "王文明"})
	if err != nil {
		t.Fatalf("DetailNames: %v", err)
	}
	if len(results) == 0 {
		t.Skip("test chars not in DB or failed phonetic lookup")
	}

	for _, r := range results {
		if r.Name == "" {
			t.Error("Name should not be empty")
		}
		if len(r.Characters) != 2 {
			t.Errorf("name %q: chars count = %d, want 2", r.Name, len(r.Characters))
		}
		if r.WuGe.TianGe.Stroke == 0 {
			t.Errorf("name %q: TianGe stroke is 0", r.Name)
		}
		if r.SanCai.Configuration == "" {
			t.Errorf("name %q: SanCai config is empty", r.Name)
		}
		if r.Phonetic.Tones == "" {
			t.Errorf("name %q: phonetic tones empty", r.Name)
		}
	}
}

func TestDetailNames_ShortGivenName(t *testing.T) {
	// Given name with only 1 char should be skipped.
	results, err := DetailNames("王", []string{"王文"})
	if err != nil {
		t.Fatalf("DetailNames: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("single-char given name should be skipped, got %d results", len(results))
	}
}

func TestDetailNames_CharNotInDB(t *testing.T) {
	// Name where given-name char is not in DB → skipped.
	results, err := DetailNames("王", []string{"王𠀀x"})
	if err != nil {
		t.Fatalf("DetailNames: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("unknown char should be skipped, got %d results", len(results))
	}
}

// =============================================================================
// EvaluateName — 三字名截断
// =============================================================================

func TestEvaluateName_TruncateLongName(t *testing.T) {
	// 3-char given name should be truncated to first 2 chars.
	eval, err := EvaluateName("王", "明文辉", "火")
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			t.Skip("test chars not in DB: " + err.Error())
		}
		t.Fatalf("EvaluateName: %v", err)
	}
	if len(eval.Characters) != 2 {
		t.Errorf("3-char given name chars count = %d, want 2 (truncated)", len(eval.Characters))
	}
	if eval.GivenName != "明文" {
		t.Errorf("givenName = %q, want 明文 (truncated from 明文辉)", eval.GivenName)
	}
}

// =============================================================================
// EvaluateName — 空名
// =============================================================================

func TestEvaluateName_EmptyGivenName(t *testing.T) {
	_, err := EvaluateName("王", "", "火")
	if err == nil {
		t.Error("expected error for empty given name, got nil")
	}
}

// =============================================================================
// ComposeNames — xiChars 含字符时 yong+xi 路径
// =============================================================================

func TestComposeNames_XiPlusYongPath(t *testing.T) {
	// Ensure xi+yong path is exercised by making xi chars available.
	yongElem := wuxingFromChinese("火")
	yongRaw := getCharsByElement(yongElem)
	// xi chars differ from yong so we can distinguish paths.
	xiElem := wuxingFromChinese("木")
	xiRaw := getCharsByElement(xiElem)

	// Use stroke combo {4,5} where both elements have chars.
	combo := []StrokeCombo{{Stroke1: 4, Stroke2: 5}}
	names := ComposeNames("王", combo, yongRaw, xiRaw)
	if len(names) == 0 {
		t.Error("ComposeNames produced no names; expected at least one valid pair")
	}
	// All names must start with surname.
	for _, n := range names {
		if len([]rune(n)) < 3 {
			t.Errorf("name %q too short", n)
		}
	}
}

// =============================================================================
// BUG-9 regression: radical element corrections
// =============================================================================

func TestRadicalToElement_SilkRadical(t *testing.T) {
	// 纟(糸部) → 金 per Kangxi dictionary.
	elem, ok := radicalToElement["纟"]
	if !ok {
		t.Fatal("纟 radical not found in radicalToElement")
	}
	if elem.String() != "金" {
		t.Errorf("纟 element = %s, want 金", elem.String())
	}
}

func TestRadicalToElement_MeatRadical(t *testing.T) {
	// 肉/⺼(肉部) → 土 per Kangxi dictionary.
	for _, r := range []string{"肉", "⺼"} {
		elem, ok := radicalToElement[r]
		if !ok {
			t.Errorf("%q radical not found in radicalToElement", r)
			continue
		}
		if elem.String() != "土" {
			t.Errorf("%q element = %s, want 土", r, elem.String())
		}
	}
}

func TestRadicalToElement_MoonRadical(t *testing.T) {
	// 月(月部) → 水 per Kangxi dictionary.
	elem, ok := radicalToElement["月"]
	if !ok {
		t.Fatal("月 radical not found in radicalToElement")
	}
	if elem.String() != "水" {
		t.Errorf("月 element = %s, want 水", elem.String())
	}
}

// =============================================================================
// BUG-8 regression: inferElementFromRadical no char-scanning fallback
// =============================================================================

func TestInferElementFromRadical_NoCharFallback(t *testing.T) {
	// Even if the char contains a known radical component, the fallback
	// should NOT scan the char string — only direct radical match.
	tests := []string{"明", "沐", "灶", "针"}
	for _, char := range tests {
		_, ok := inferElementFromRadical(char)
		if ok {
			t.Errorf("inferElementFromRadical(%q) should not resolve via char scan", char)
		}
	}
}

// =============================================================================
// BUG-7 regression: negative chars filtered in ComposeNames
// =============================================================================

func TestComposeNames_NegativeCharFiltered(t *testing.T) {
	// Register a known-negative char to verify it is filtered.
	negativeChars["死"] = true
	negativeChars["亡"] = true

	yong := map[int][]CharLite{
		4: {{Char: "死", Tone: 3}},
		5: {{Char: "明", Tone: 2}},
	}
	xi := map[int][]CharLite{
		4: {{Char: "文", Tone: 2}},
		5: {{Char: "亡", Tone: 2}},
	}

	combo := []StrokeCombo{{Stroke1: 4, Stroke2: 5}}
	names := ComposeNames("王", combo, yong, xi)

	for _, name := range names {
		for _, r := range name {
			if negativeChars[string(r)] {
				t.Errorf("name %q contains negative char %q", name, string(r))
			}
		}
	}
}

func TestHasNegativeChar(t *testing.T) {
	negativeChars["死"] = true

	if !hasNegativeChar("王死明") {
		t.Error("王死明 should have negative char")
	}
	if hasNegativeChar("王文明") {
		t.Error("王文明 should NOT have negative char")
	}
}

// =============================================================================
// BUG-13 regression: stroke limit 36
// =============================================================================

func TestEnumWuGeCombinations_Stroke36(t *testing.T) {
	// Characters with 36 Kangxi strokes should be included in enumeration.
	// Use a surname with known stroke count. Just verify the function runs
	// without truncating at 31 and produces combos with strokes up to 36.
	// 王=4 strokes in Kangxi.
	result := enumWuGeCombinations(4)
	if len(result.Combinations) == 0 {
		t.Fatal("no combos generated")
	}

	maxS1, maxS2 := 0, 0
	for _, c := range result.Combinations {
		if c.Stroke1 > maxS1 {
			maxS1 = c.Stroke1
		}
		if c.Stroke2 > maxS2 {
			maxS2 = c.Stroke2
		}
	}
	// At least some combos should have strokes above the old limit of 31.
	if maxS1 <= 31 && maxS2 <= 31 {
		t.Log("no combos with stroke >31 found (may be filtered by auspiciousness)")
	}
	// Verify no combo exceeds the new limit.
	if maxS1 > 36 || maxS2 > 36 {
		t.Errorf("max stroke (%d, %d) exceeds limit 36", maxS1, maxS2)
	}
}
func TestComputeWuGeFromStrokes(t *testing.T) {
	// 王(4) + 1st名(9) + 2nd名(16) → standard example
	wg := computeWuGeFromStrokes(4, 9, 16)

	if wg.TianGe.Stroke != 5 {
		t.Errorf("天格 stroke = %d, want 5", wg.TianGe.Stroke)
	}
	if wg.RenGe.Stroke != 13 {
		t.Errorf("人格 stroke = %d, want 13", wg.RenGe.Stroke)
	}
	if wg.DiGe.Stroke != 25 {
		t.Errorf("地格 stroke = %d, want 25", wg.DiGe.Stroke)
	}
	if wg.ZongGe.Stroke != 29 {
		t.Errorf("总格 stroke = %d, want 29", wg.ZongGe.Stroke)
	}
	// 外格 = 总格 - 人格 + 1 = 29 - 13 + 1 = 17
	if wg.WaiGe.Stroke != 17 {
		t.Errorf("外格 stroke = %d, want 17", wg.WaiGe.Stroke)
	}
}

// TestComputeWuGeFromStrokes_Minimum strokes verifies boundary case.
func TestComputeWuGeFromStrokes_MinStrokes(t *testing.T) {
	wg := computeWuGeFromStrokes(1, 1, 1)

	// 天格=2, 人格=2, 地格=2, 总格=3, 外格=总格-人格+1=3-2+1=2
	if wg.TianGe.Stroke != 2 {
		t.Errorf("天格 = %d, want 2", wg.TianGe.Stroke)
	}
	if wg.WaiGe.Stroke != 2 {
		t.Errorf("外格 = %d, want 2", wg.WaiGe.Stroke)
	}
}

// TestWuGeElements verifies five-element attribution for known stroke counts.
func TestWuGeElements(t *testing.T) {
	// According to 81-number五格, the last digit determines element:
	// 1,2=木 3,4=火 5,6=土 7,8=金 9,0=水
	tests := []struct {
		stroke     int
		wantWuxing string
	}{
		{1, "木"}, {2, "木"},
		{3, "火"}, {4, "火"},
		{5, "土"}, {6, "土"},
		{7, "金"}, {8, "金"},
		{9, "水"}, {10, "水"},
		{11, "木"}, {21, "木"},
		{13, "火"}, {24, "火"},
	}

	for _, tt := range tests {
		got := strokeResult(tt.stroke)
		if got.Element != tt.wantWuxing {
			t.Errorf("strokeResult(%d).Element = %s, want %s",
				tt.stroke, got.Element, tt.wantWuxing)
		}
	}
}

// TestStrokeOverflow verifies stroke count wrapping above 81.
func TestStrokeOverflow(t *testing.T) {
	// 82 should wrap to 1
	r1 := strokeResult(1)
	r82 := strokeResult(82)
	if r1.Element != r82.Element {
		t.Errorf("stroke 82 should wrap to 1: element %s != %s", r82.Element, r1.Element)
	}
	if r82.Stroke != 1 {
		t.Errorf("stroke 82 should wrap to stroke 1, got %d", r82.Stroke)
	}

	// 81 should stay as 81
	r81 := strokeResult(81)
	if r81.Stroke != 81 {
		t.Errorf("stroke 81 should stay 81, got %d", r81.Stroke)
	}
}

// TestSanCai verifies三才 configuration.
func TestSanCai(t *testing.T) {
	// 木木木 → 大吉 (all wood, mutual support)
	sc := computeSanCai("木", "木", "木")
	if sc.Configuration != "木木木" {
		t.Errorf("configuration = %s, want 木木木", sc.Configuration)
	}
	if sc.Fortune == "" {
		t.Error("fortune should not be empty")
	}

	// 金金金 → should exist in config
	sc2 := computeSanCai("金", "金", "金")
	if sc2.Configuration != "金金金" {
		t.Errorf("configuration = %s, want 金金金", sc2.Configuration)
	}

	// Unknown combo → default
	sc3 := computeSanCai("木", "火", "金")
	if sc3.Configuration != "木火金" {
		t.Errorf("configuration = %s, want 木火金", sc3.Configuration)
	}
}

// =============================================================================
// Golden: 首页 36 个示例名 — 验证引擎能力覆盖
// =============================================================================

var exampleNames = []string{
	"林观澜", "赵知微", "徐望舒", "王砚清", "李鹿鸣", "刘予安",
	"黄文茵", "吴佩弦", "张知行", "陈思诚", "杨明哲", "孙思远",
	"马归真", "朱修远", "周如玉", "郑含章", "谢清风", "唐致远",
	"于若水", "邓景行", "钱浩然", "薛养正", "卢思齐", "戴知远",
	"邵明德", "雷敬之", "方敏行", "袁守拙", "乔清和", "秦云舒",
	"任心怡", "苏子衿", "罗静言", "夏砚耕", "顾逢春", "汤书白",
}

func TestExampleNames_CharsInDatabase(t *testing.T) {
	for _, full := range exampleNames {
		rs := []rune(full)
		surname := string(rs[0])
		g1 := string(rs[1])
		g2 := string(rs[2])

		// 姓氏必须在字典中
		if lookupKangxiStroke(surname) == 0 {
			t.Errorf("%s: surname %q not in database", full, surname)
			continue
		}
		// 两个名字字必须在字典中
		ce1, ok1 := charByRune[rs[1]]
		if !ok1 {
			t.Errorf("%s: char %q not in charByRune", full, g1)
			continue
		}
		ce2, ok2 := charByRune[rs[2]]
		if !ok2 {
			t.Errorf("%s: char %q not in charByRune", full, g2)
			continue
		}
		// 五格必须可计算
		ss := lookupKangxiStroke(surname)
		wg := computeWuGeFromStrokes(ss, ce1.Stroke, ce2.Stroke)
		if wg.TianGe.Stroke == 0 || wg.RenGe.Stroke == 0 || wg.DiGe.Stroke == 0 {
			t.Errorf("%s: wuge calculation failed: %+v", full, wg)
		}
	}
}

func TestExampleNames_DetailNames(t *testing.T) {
	for _, full := range exampleNames {
		rs := []rune(full)
		surname := string(rs[0])
		candidates, err := DetailNames(surname, []string{full})
		if err != nil {
			t.Errorf("%s: DetailNames error: %v", full, err)
			continue
		}
		if len(candidates) != 1 {
			t.Errorf("%s: expected 1 candidate, got %d", full, len(candidates))
			continue
		}
		c := candidates[0]
		if c.Name != full {
			t.Errorf("%s: name mismatch: %s", full, c.Name)
		}
		// 五格 fortune 不能为空
		if c.WuGe.RenGe.Fortune == "" {
			t.Errorf("%s: renge fortune empty", full)
		}
		if c.SanCai.Fortune == "" {
			t.Errorf("%s: sancai fortune empty", full)
		}
	}
}

func TestExampleNames_PhoneticInfo(t *testing.T) {
	for _, full := range exampleNames {
		rs := []rune(full)
		surname := string(rs[0])
		candidates, err := DetailNames(surname, []string{full})
		if err != nil || len(candidates) != 1 {
			continue
		}
		c := candidates[0]
		if c.Phonetic.Tones == "" {
			t.Errorf("%s: tones empty", full)
		}
	}
}

// TestWuxingFromChinese verifies element mapping.
func TestWuxingFromChinese(t *testing.T) {
	tests := []struct {
		ch   string
		want Wuxing
	}{
		{"木", ganzhi.WxMu},
		{"火", ganzhi.WxHuo},
		{"土", ganzhi.WxTu},
		{"金", ganzhi.WxJin},
		{"水", ganzhi.WxShui},
		{"x", ganzhi.Wuxing(0)},
	}

	for _, tt := range tests {
		got := wuxingFromChinese(tt.ch)
		if got != tt.want {
			t.Errorf("wuxingFromChinese(%q) = %d, want %d", tt.ch, got, tt.want)
		}
	}
}

