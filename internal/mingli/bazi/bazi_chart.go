package bazi

import (
	"math"
	"time"

	"github.com/25types/25types/internal/ganzhi"
)

// Source constants for TenGodEntry.Source.
const (
	SourceStem    = "stem"
	SourceMainQi  = "main_qi"
	SourceMidQi   = "mid_qi"
	SourceMinorQi = "minor_qi"
)

// TenGodEntry is one ten-god result for a stem in a pillar.
type TenGodEntry struct {
	Stem   Stem   `json:"stem"`
	TenGod string `json:"ten_god"`
	Source string `json:"source"`
}

// LifeStageEntry is one life-stage result for a stem/branch pair.
type LifeStageEntry struct {
	Stem   Stem   `json:"stem"`   // stem
	Branch Branch `json:"branch"` // branch
	Stage  string `json:"stage"`  // "长生"..."养"
}

// PillarInfo collects all computed data for one pillar.
type PillarInfo struct {
	Stem        Stem             `json:"stem"`
	Branch      Branch           `json:"branch"`
	NaYin       string           `json:"nayin"`
	HiddenStems HiddenStemsOut  `json:"hidden_stems"`
	TenGods     []TenGodEntry    `json:"ten_gods"`
	LifeStages  []LifeStageEntry `json:"life_stages"`
	ShenSha     []ShenShaEntry   `json:"shensha"`
	IsVoid      bool             `json:"is_void"`
	IsSelfHe    bool             `json:"is_self_he"`
	SelfHeName  string           `json:"self_he_name,omitempty"`
	IsKuiGang   bool             `json:"is_kui_gang"`
}

// -- output types for ComputeChart --

// HiddenStemsOut mirrors domain.HiddenStems but uses engine types.
type HiddenStemsOut struct {
	Main  Stem `json:"main"`
	Mid   *Stem `json:"mid"`
	Minor *Stem `json:"minor"`
}

// StageOut mirrors domain.Stage.
type StageOut struct {
	Name   string `json:"name"`
	Branch Branch `json:"branch"`
}

// TwelveStages holds the twelve life stages (长生十二宫) with named fields.
type TwelveStages struct {
	ChangSheng StageOut // 长生
	MuYu       StageOut // 沐浴
	GuanDai    StageOut // 冠带
	LinGuan    StageOut // 临官
	DiWang     StageOut // 帝旺
	Shuai      StageOut // 衰
	Bing       StageOut // 病
	Si         StageOut // 死
	Mu         StageOut // 墓
	Jue        StageOut // 绝
	Tai        StageOut // 胎
	Yang       StageOut // 养
}

// Slice returns the twelve stages as an array for API serialization.
func (ts TwelveStages) Slice() [12]StageOut {
	return [12]StageOut{ts.ChangSheng, ts.MuYu, ts.GuanDai, ts.LinGuan, ts.DiWang, ts.Shuai, ts.Bing, ts.Si, ts.Mu, ts.Jue, ts.Tai, ts.Yang}
}

// NewTwelveStages creates a TwelveStages from a [12]StageOut array.
func NewTwelveStages(arr [12]StageOut) TwelveStages {
	return TwelveStages{
		ChangSheng: arr[0],
		MuYu:       arr[1],
		GuanDai:    arr[2],
		LinGuan:    arr[3],
		DiWang:     arr[4],
		Shuai:      arr[5],
		Bing:       arr[6],
		Si:         arr[7],
		Mu:         arr[8],
		Jue:        arr[9],
		Tai:        arr[10],
		Yang:       arr[11],
	}
}

// DayunPillars holds raw big fortune (大运) pillar data.
type DayunPillars struct {
	StartAge  int      `json:"start_age"`
	Direction string   `json:"direction"`
	Pillars   []Pillar `json:"pillars"`
}

// ChartResult is the complete output of ComputeChart.
type ChartResult struct {
	Year         PillarInfo
	Month        PillarInfo
	Day          PillarInfo
	Hour         PillarInfo
	SolarTime  float64
	SolarDate  time.Time
	BaziDate   time.Time
	LifeStages TwelveStages
	Dayun        DayunPillars
	DayMaster    Stem
	ElementCount map[Element]int
}

// ToBazi returns the four pillars as a Bazi value.
func (cr ChartResult) ToBazi() Bazi {
	return Bazi{
		Year:  Pillar{Stem: cr.Year.Stem, Branch: cr.Year.Branch},
		Month: Pillar{Stem: cr.Month.Stem, Branch: cr.Month.Branch},
		Day:   Pillar{Stem: cr.Day.Stem, Branch: cr.Day.Branch},
		Hour:  Pillar{Stem: cr.Hour.Stem, Branch: cr.Hour.Branch},
	}
}

// NaYinArray returns the four nayin strings as a [4]string.
func (cr ChartResult) NaYinArray() [4]string {
	return [4]string{cr.Year.NaYin, cr.Month.NaYin, cr.Day.NaYin, cr.Hour.NaYin}
}

// HiddenStemsArray returns the four hidden stems as a [4]HiddenStemsOut.
func (cr ChartResult) HiddenStemsArray() [4]HiddenStemsOut {
	return [4]HiddenStemsOut{cr.Year.HiddenStems, cr.Month.HiddenStems, cr.Day.HiddenStems, cr.Hour.HiddenStems}
}

// TenGodsArray returns the four ten-god pairs as a [4][2]string.
func (cr ChartResult) TenGodsArray() [4][2]string {
	pick := func(pi PillarInfo) (stem, branch string) {
		for _, e := range pi.TenGods {
			switch e.Source {
			case SourceStem:
				stem = e.TenGod
			case SourceMainQi:
				branch = e.TenGod
			}
		}
		return
	}
	s0, b0 := pick(cr.Year)
	s1, b1 := pick(cr.Month)
	s2, b2 := pick(cr.Day)
	s3, b3 := pick(cr.Hour)
	return [4][2]string{{s0, b0}, {s1, b1}, {s2, b2}, {s3, b3}}
}

// -- hidden stems --

func computeHiddenStems(bz ganzhi.Bazi) [4]HiddenStemsOut {
	pillars := bz.Slice()
	var out [4]HiddenStemsOut
	for i, p := range pillars {
		hs := HiddenStemsForBranch(p.Branch)
		out[i] = HiddenStemsOut{
			Main:  intPtrToStemVal(hs.Main),
			Mid:   intPtrToStem(hs.Mid),
			Minor: intPtrToStem(hs.Minor),
		}
	}
	return out
}

func intPtrToStemVal(p *int) Stem {
	if p == nil {
		return 0
	}
	return Stem(*p)
}

func intPtrToStem(p *int) *Stem {
	if p == nil {
		return nil
	}
	s := Stem(*p)
	return &s
}

// HiddenStemsForBranch returns the hidden stems (藏干) for a branch.
func HiddenStemsForBranch(b Branch) HiddenStemsQi {
	if hs, ok := defaultEngine.HiddenStemsTable[int(b)]; ok {
		return hs
	}
	return HiddenStemsQi{}
}

// -- ten gods --

// TenGod type enumeration.
const (
	TenGodBiJian  = 0 // 比肩
	TenGodJieCai  = 1 // 劫财
	TenGodShiShen = 2 // 食神
	TenGodShangGuan = 3 // 伤官
	TenGodPianCai = 4 // 偏财
	TenGodZhengCai = 5 // 正财
	TenGodQiSha   = 6 // 七杀
	TenGodZhengGuan = 7 // 正官
	TenGodPianYin = 8 // 偏印
	TenGodZhengYin = 9 // 正印
)

var tenGodNamesZH = [10]string{
	"比肩", "劫财", "食神", "伤官", "偏财",
	"正财", "七杀", "正官", "偏印", "正印",
}

// TenGodName returns the Chinese name for a ten god type.
func TenGodName(tg int) string {
	if tg >= 0 && tg < 10 {
		return tenGodNamesZH[tg]
	}
	return ""
}

func TenGodType(dmElem Element, dmYY YinYang, otherElem Element, otherYY YinYang) int {
	switch {
	case dmElem == otherElem:
		if dmYY == otherYY {
			return TenGodBiJian
		}
		return TenGodJieCai
	case Sheng(dmElem, otherElem):
		if dmYY == otherYY {
			return TenGodShiShen
		}
		return TenGodShangGuan
	case Sheng(otherElem, dmElem):
		if dmYY == otherYY {
			return TenGodPianYin
		}
		return TenGodZhengYin
	case Ke(dmElem, otherElem):
		if dmYY == otherYY {
			return TenGodPianCai
		}
		return TenGodZhengCai
	default:
		if dmYY == otherYY {
			return TenGodQiSha
		}
		return TenGodZhengGuan
	}
}

// -- na yin --

func computeNaYin(bz ganzhi.Bazi) [4]string {
	pillars := bz.Slice()
	var out [4]string
	for i, p := range pillars {
		if s := NaYinString(p.Stem, p.Branch); s != "" {
			out[i] = s
		} else {
			out[i] = "未知"
		}
	}
	return out
}

// -- life stages --

var stageNamesZH = [12]string{
	"长生", "沐浴", "冠带", "临官", "帝旺",
	"衰", "病", "死", "墓", "绝", "胎", "养",
}

func computeLifeStages(dayMaster Stem) TwelveStages {
	e := defaultEngine
	if e == nil {
		return TwelveStages{}
	}
	stages, ok := defaultData.LifeStagesTable[int(dayMaster)]
	if !ok {
		return TwelveStages{}
	}
	var out [12]StageOut
	for i, b := range stages {
		out[i] = StageOut{Name: stageNamesZH[i], Branch: Branch(b)}
	}
	return NewTwelveStages(out)
}

// -- big fortune --

func computeDayun(year, month, day int, monthPillar Pillar, gender Gender, yearStem Stem) DayunPillars {
	isYang := StemYinYang(yearStem) == Yang
	isMale := gender == Male
	forward := (isMale && isYang) || (!isMale && !isYang)

	direction := "forward"
	if !forward {
		direction = "backward"
	}

	startAge := fortuneStartAge(year, month, day, forward)
	pillars := fortunePillars(monthPillar, forward, 8)

	return DayunPillars{
		StartAge:  startAge,
		Direction: direction,
		Pillars:   pillars,
	}
}

// fortuneStartAge calculates the starting age for big fortune (大运起运岁数).
func fortuneStartAge(year, month, day int, forward bool) int {
	// Find the nearest "节" (major solar term).
	birthTime := time.Date(year, time.Month(month), day, 12, 0, 0, 0, time.UTC)

	var targetDate time.Time
	if forward {
		// Next 节 after birth.
		for i := 0; i < 12; i++ {
			lon := solarTermLongitudes[i]
			t := solarTermDate(year, lon)
			tDate := time.Date(t.Year(), t.Month(), t.Day(), 12, 0, 0, 0, time.UTC)
			if tDate.After(birthTime) || tDate.Equal(birthTime) {
				targetDate = tDate
				break
			}
		}
		// If all 节 this year are before birth, use first 节 of next year.
		if targetDate.IsZero() {
			t := solarTermDate(year+1, solarTermLongitudes[0])
			targetDate = time.Date(t.Year(), t.Month(), t.Day(), 12, 0, 0, 0, time.UTC)
		}
	} else {
		// Previous 节 before birth.
		for i := 11; i >= 0; i-- {
			lon := solarTermLongitudes[i]
			t := solarTermDate(year, lon)
			tDate := time.Date(t.Year(), t.Month(), t.Day(), 12, 0, 0, 0, time.UTC)
			if tDate.Before(birthTime) {
				targetDate = tDate
				break
			}
		}
		if targetDate.IsZero() {
			t := solarTermDate(year-1, solarTermLongitudes[11])
			targetDate = time.Date(t.Year(), t.Month(), t.Day(), 12, 0, 0, 0, time.UTC)
		}
	}

	days := int(math.Abs(targetDate.Sub(birthTime).Hours()) / 24)
	return days / 3
}

func fortunePillars(monthPillar Pillar, forward bool, steps int) []Pillar {
	pillars := make([]Pillar, 0, steps)
	currentIdx := sixtyCycleName(monthPillar.Stem, monthPillar.Branch)

	for i := 0; i < steps; i++ {
		if forward {
			currentIdx = (currentIdx + 1) % 60
		} else {
			currentIdx = (currentIdx - 1 + 60) % 60
		}
		stem := Stem(currentIdx%10 + 1)
		branch := Branch(currentIdx%12 + 1)
		pillars = append(pillars, Pillar{Stem: stem, Branch: branch})
	}
	return pillars
}

// ComputeTenGodsTable returns all ten-god entries for each pillar,
// including stem, main hidden stem, mid hidden stem, and minor hidden stem
// vs the day master.
func ComputeTenGodsTable(dayMaster Stem, bz ganzhi.Bazi, hs [4]HiddenStemsOut) [4][]TenGodEntry {
	pillars := bz.Slice()
	var out [4][]TenGodEntry
	dmElem := StemElement(dayMaster)
	dmYY := StemYinYang(dayMaster)

	for pi, p := range pillars {
		var entries []TenGodEntry

		sElem := StemElement(p.Stem)
		sYY := StemYinYang(p.Stem)
		entries = append(entries, TenGodEntry{
			Stem:   p.Stem,
			TenGod: TenGodName(TenGodType(dmElem, dmYY, sElem, sYY)),
			Source: SourceStem,
		})

		addHidden := func(s Stem, source string) {
			hElem := StemElement(s)
			hYY := StemYinYang(s)
			entries = append(entries, TenGodEntry{
				Stem:   s,
				TenGod: TenGodName(TenGodType(dmElem, dmYY, hElem, hYY)),
				Source: source,
			})
		}
		if hs[pi].Main != 0 {
			addHidden(hs[pi].Main, SourceMainQi)
		}
		if hs[pi].Mid != nil {
			addHidden(*hs[pi].Mid, SourceMidQi)
		}
		if hs[pi].Minor != nil {
			addHidden(*hs[pi].Minor, SourceMinorQi)
		}

		out[pi] = entries
	}
	return out
}

// ComputeLifeStageTable returns the life-stage of every stem that appears
// in the chart (pillar stems + all hidden stems) against each pillar's branch.
func ComputeLifeStageTable(bz ganzhi.Bazi, hiddenStems [4]HiddenStemsOut) [4][]LifeStageEntry {
	pillars := bz.Slice()
	e := defaultEngine
	if e == nil {
		return [4][]LifeStageEntry{}
	}

	type stemAt struct {
		stem    Stem
		pillarI int
	}
	var stems []stemAt
	for pi := range pillars {
		stems = append(stems, stemAt{pillars[pi].Stem, pi})
		hs := hiddenStems[pi]
		if hs.Main != 0 {
			stems = append(stems, stemAt{hs.Main, pi})
		}
		if hs.Mid != nil {
			stems = append(stems, stemAt{*hs.Mid, pi})
		}
		if hs.Minor != nil {
			stems = append(stems, stemAt{*hs.Minor, pi})
		}
	}

	var out [4][]LifeStageEntry

	for _, sa := range stems {
		stageRow, ok := defaultData.LifeStagesTable[int(sa.stem)]
		if !ok {
			continue
		}
		for bi, b := range pillars {
			bn := int(b.Branch)
			if bn < 1 || bn > 12 {
				continue
			}
			stageIdx := -1
			for si, sb := range stageRow {
				if sb == bn {
					stageIdx = si
					break
				}
			}
			if stageIdx < 0 || stageIdx >= 12 {
				continue
			}
			out[bi] = append(out[bi], LifeStageEntry{
				Stem:   sa.stem,
				Branch: Branch(bn),
				Stage:  stageNamesZH[stageIdx],
			})
		}
	}
	return out
}

// -- element count --

func computeElementCount(bz ganzhi.Bazi, hs [4]HiddenStemsOut) map[Element]int {
	pillars := bz.Slice()
	count := make(map[Element]int)
	for _, p := range pillars {
		count[StemElement(p.Stem)]++
		count[BranchElement(p.Branch)]++
	}
	for _, h := range hs {
		count[StemElement(h.Main)]++
		if h.Mid != nil {
			count[StemElement(*h.Mid)]++
		}
		if h.Minor != nil {
			count[StemElement(*h.Minor)]++
		}
	}
	return count
}

// NaYinRelation describes the five-element relationship between two pillars' nayin.
type NaYinRelation struct {
	FromPillar int    `json:"from_pillar"`
	ToPillar   int    `json:"to_pillar"`
	Relation   string `json:"relation"` // "相生"|"相克"|"相同"
	Detail     string `json:"detail"`   // "海中金生大林木"
}

// DayMansion describes one of the 28 mansions (二十八宿).
type DayMansion struct {
	Index    int    `json:"index"`
	Name     string `json:"name"`
	Animal   string `json:"animal"`
	Element  string `json:"element"`
	Group    string `json:"group"`
	GroupIdx int    `json:"group_idx"`
}

// ComputeNaYinRelations computes nayin inter-pillar element relations.
func ComputeNaYinRelations(nayin [4]string) []NaYinRelation {
	elements := [4]Element{}
	for i, n := range nayin {
		elements[i] = nayinElement(n)
	}

	var rels []NaYinRelation
	for i := 0; i < 4; i++ {
		for j := i + 1; j < 4; j++ {
			ei, ej := elements[i], elements[j]
			if ei == 0 || ej == 0 {
				continue
			}
			var rel, detail string
			switch {
			case ei == ej:
				rel = "相同"
				detail = nayin[i] + "与" + nayin[j] + "同气"
			case Sheng(ei, ej):
				rel = "相生"
				detail = nayin[i] + "生" + nayin[j]
			case Sheng(ej, ei):
				rel = "相生"
				detail = nayin[j] + "生" + nayin[i]
			case Ke(ei, ej):
				rel = "相克"
				detail = nayin[i] + "克" + nayin[j]
			case Ke(ej, ei):
				rel = "相克"
				detail = nayin[j] + "克" + nayin[i]
			}
			rels = append(rels, NaYinRelation{
				FromPillar: i, ToPillar: j, Relation: rel, Detail: detail,
			})
		}
	}
	return rels
}

// ElementCountStrings converts an element count map to string-keyed form.
func ElementCountStrings(ec map[Element]int) map[string]int {
	m := make(map[string]int, len(ec))
	for e, c := range ec {
		m[e.String()] = c
	}
	return m
}

// ComputeWangShuaiMap returns the seasonal strength of all five elements for a month branch.
func ComputeWangShuaiMap(monthBranch Branch) map[string]string {
	return map[string]string{
		ElemWood.String():  MonthWangShuai(ElemWood, monthBranch),
		ElemFire.String():  MonthWangShuai(ElemFire, monthBranch),
		ElemEarth.String(): MonthWangShuai(ElemEarth, monthBranch),
		ElemMetal.String(): MonthWangShuai(ElemMetal, monthBranch),
		ElemWater.String(): MonthWangShuai(ElemWater, monthBranch),
	}
}

// ChartOutput is the unified bazi chart output shared by HTTP and MCP.
type ChartOutput struct {
	YearPillar       PillarInfo      `json:"year_pillar"`
	MonthPillar      PillarInfo      `json:"month_pillar"`
	DayPillar        PillarInfo      `json:"day_pillar"`
	HourPillar       PillarInfo      `json:"hour_pillar"`
	DayMaster        string          `json:"day_master"`
	LifeStages       [12]StageOut    `json:"life_stages"`
	Dayun            *DayunResult    `json:"dayun"`
	ElementCount     map[string]int  `json:"element_count"`
	SolarTimeMinutes float64         `json:"solar_time_minutes"`
	SolarDatetime    string          `json:"solar_datetime"`
	BaziDatetime     string          `json:"bazi_datetime"`
	FullHeHui        []TripleHeFull  `json:"full_he_hui"`
	GongJia          []GongJiaEntry  `json:"gong_jia"`
	TaiYuanMingGong  TaiYuanMingGong `json:"tai_yuan_ming_gong"`
	NayinRelations   []NaYinRelation `json:"nayin_relations"`
	SanQiName        string          `json:"sanqi_name"`
	Zodiac           string          `json:"zodiac"`
	Season           string          `json:"season"`
	LunarMonth       string          `json:"lunar_month"`
	HourRange        string          `json:"hour_range"`
	XunName          string          `json:"xun_name"`
	WangShuai        map[string]string `json:"wang_shuai"`
	DayMansion       DayMansion      `json:"day_mansion"`
	YongShen         YongShenResult  `json:"yong_shen"`
}

// BuildChartOutput converts a ChartResult into the unified API output.
func BuildChartOutput(chart ChartResult, birthYear, birthMonth, birthHour int) ChartOutput {
	bz := chart.ToBazi()
	ys := ComputeYongShen(chart)
	dayun := ComputeDayunResult(chart.Dayun, chart.DayMaster, birthYear, time.Now().Year(), bz)

	return ChartOutput{
		YearPillar:       chart.Year,
		MonthPillar:      chart.Month,
		DayPillar:        chart.Day,
		HourPillar:       chart.Hour,
		DayMaster:        ganzhi.StemName(chart.DayMaster),
		LifeStages:       chart.LifeStages.Slice(),
		Dayun:            dayun,
		ElementCount:     ElementCountStrings(chart.ElementCount),
		SolarTimeMinutes: chart.SolarTime,
		SolarDatetime:    chart.SolarDate.Format(time.RFC3339),
		BaziDatetime:     chart.BaziDate.Format("2006-01-02") + " " + ganzhi.BranchName(chart.Hour.Branch) + "时",
		FullHeHui:        ComputeFullTripleHeHui(bz),
		GongJia:          ComputeGongJia(bz),
		TaiYuanMingGong:  ComputeTaiYuanMingGong(Pillar{Stem: chart.Month.Stem, Branch: chart.Month.Branch}, chart.Year.Stem, birthMonth, birthHour),
		NayinRelations:   ComputeNaYinRelations(chart.NaYinArray()),
		SanQiName:        SanQiName(SanQiType(bz)),
		Zodiac:           Zodiac(chart.Year.Branch),
		Season:           ganzhi.BranchSeason(chart.Month.Branch),
		LunarMonth:       ganzhi.BranchLunarMonth(chart.Month.Branch),
		HourRange:        BranchHourRange(chart.Hour.Branch),
		XunName:          XunName(Pillar{Stem: chart.Day.Stem, Branch: chart.Day.Branch}),
		WangShuai:        ComputeWangShuaiMap(chart.Month.Branch),
		DayMansion:       MansionForDay(Pillar{Stem: chart.Day.Stem, Branch: chart.Day.Branch}),
		YongShen:         ys,
	}
}

// BondOutput is the unified bond (合盘) output shared by HTTP and MCP.
type BondOutput struct {
	ChartA ChartOutput `json:"chart_a"`
	ChartB ChartOutput `json:"chart_b"`
	Bond   BondResult  `json:"bond"`
}

// NaYinString returns the NaYin name for a stem-branch combination.
func NaYinString(s Stem, b Branch) string {
	e := defaultEngine
	if e == nil {
		return ""
	}
	idx := sixtyCycleName(s, b)
	if int(idx) < 60 && e.NayinTable != nil {
		if name, ok := e.NayinTable[int(idx)]; ok {
			return name
		}
	}
	return ""
}
