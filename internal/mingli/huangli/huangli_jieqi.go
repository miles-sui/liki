package huangli

import "time"

// JieQiDepth describes the current solar term position for a given date.
type JieQiDepth struct {
	TermName     string `json:"term_name"`     // current solar term name
	DaysIn       int    `json:"days_in"`       // days into current term
	NextTermName string `json:"next_term_name"` // next solar term name
	DaysToNext   int    `json:"days_to_next"`  // days until next term
}

// Solar term names in order (24 terms, starting from 立春).
var jieQiNames = [24]string{
	"立春", "雨水", "惊蛰", "春分", "清明", "谷雨",
	"立夏", "小满", "芒种", "夏至", "小暑", "大暑",
	"立秋", "处暑", "白露", "秋分", "寒露", "霜降",
	"立冬", "小雪", "大雪", "冬至", "小寒", "大寒",
}

// ComputeJieQiDepth returns which solar term the given date falls in and
// how many days into it we are. termIndex is 0-23.
func ComputeJieQiDepth(year, month, day int) JieQiDepth {
	date := time.Date(year, time.Month(month), day, 12, 0, 0, 0, time.UTC)

	// Find current and next term.
	var prevIdx, nextIdx int
	var prevDate, nextDate time.Time

	for i := 0; i < 24; i++ {
		t := solarTermDate(year, jieQiLongitudes[i])
		tDate := time.Date(t.Year(), t.Month(), t.Day(), 12, 0, 0, 0, time.UTC)
		if tDate.After(date) {
			nextIdx = i
			nextDate = tDate
			if i == 0 {
				prevIdx = 23
				prevDate = time.Date(solarTermDate(year-1, jieQiLongitudes[23]).Year(),
					solarTermDate(year-1, jieQiLongitudes[23]).Month(),
					solarTermDate(year-1, jieQiLongitudes[23]).Day(), 12, 0, 0, 0, time.UTC)
			} else {
				prevIdx = i - 1
				tPrev := solarTermDate(year, jieQiLongitudes[i-1])
				prevDate = time.Date(tPrev.Year(), tPrev.Month(), tPrev.Day(), 12, 0, 0, 0, time.UTC)
			}
			break
		}
	}
	if nextDate.IsZero() {
		// Date is after the last term (大寒) — wrap to next year.
		prevIdx = 23
		tPrev := solarTermDate(year, jieQiLongitudes[23])
		prevDate = time.Date(tPrev.Year(), tPrev.Month(), tPrev.Day(), 12, 0, 0, 0, time.UTC)
		nextIdx = 0
		tNext := solarTermDate(year+1, jieQiLongitudes[0])
		nextDate = time.Date(tNext.Year(), tNext.Month(), tNext.Day(), 12, 0, 0, 0, time.UTC)
	}

	daysIn := int(date.Sub(prevDate).Hours() / 24)
	if daysIn < 0 {
		daysIn = 0
	}
	daysToNext := int(nextDate.Sub(date).Hours() / 24)
	if daysToNext < 0 {
		daysToNext = 0
	}

	return JieQiDepth{
		TermName:     jieQiNames[prevIdx],
		DaysIn:       daysIn,
		NextTermName: jieQiNames[nextIdx],
		DaysToNext:   daysToNext,
	}
}

// RenYuanPhase is one phase in the 人元司令分野 table.
type RenYuanPhase struct {
	Stem     int    `json:"stem"`
	StemName string `json:"stem_name"`
	Days     int    `json:"days"`
}

// RenYuanSiLing describes which hidden stem governs during each portion of a month.
type RenYuanSiLing struct {
	MonthBranch int            `json:"month_branch"`
	Phases      []RenYuanPhase `json:"phases"`
	Current     *RenYuanPhase  `json:"current"` // current governing stem, if date provided
}

// renYuanTable maps month branch (solar month, 寅=3) to its phases.
var renYuanTable = map[int][]RenYuanPhase{
	3:  {{Stem: 5, StemName: "戊", Days: 7}, {Stem: 3, StemName: "丙", Days: 7}, {Stem: 1, StemName: "甲", Days: 16}},  // 寅月
	4:  {{Stem: 1, StemName: "甲", Days: 10}, {Stem: 2, StemName: "乙", Days: 20}},                                      // 卯月
	5:  {{Stem: 2, StemName: "乙", Days: 9}, {Stem: 10, StemName: "癸", Days: 3}, {Stem: 5, StemName: "戊", Days: 18}},   // 辰月
	6:  {{Stem: 5, StemName: "戊", Days: 5}, {Stem: 7, StemName: "庚", Days: 9}, {Stem: 3, StemName: "丙", Days: 16}},    // 巳月
	7:  {{Stem: 3, StemName: "丙", Days: 10}, {Stem: 6, StemName: "己", Days: 9}, {Stem: 4, StemName: "丁", Days: 11}},   // 午月
	8:  {{Stem: 4, StemName: "丁", Days: 9}, {Stem: 2, StemName: "乙", Days: 3}, {Stem: 6, StemName: "己", Days: 18}},   // 未月
	9:  {{Stem: 6, StemName: "己", Days: 7}, {Stem: 5, StemName: "戊", Days: 3}, {Stem: 9, StemName: "壬", Days: 3}, {Stem: 7, StemName: "庚", Days: 17}}, // 申月
	10: {{Stem: 7, StemName: "庚", Days: 10}, {Stem: 8, StemName: "辛", Days: 20}},                                       // 酉月
	11: {{Stem: 8, StemName: "辛", Days: 9}, {Stem: 4, StemName: "丁", Days: 3}, {Stem: 5, StemName: "戊", Days: 18}},     // 戌月
	12: {{Stem: 5, StemName: "戊", Days: 7}, {Stem: 1, StemName: "甲", Days: 5}, {Stem: 9, StemName: "壬", Days: 18}},     // 亥月
	1:  {{Stem: 9, StemName: "壬", Days: 10}, {Stem: 10, StemName: "癸", Days: 20}},                                      // 子月
	2:  {{Stem: 10, StemName: "癸", Days: 9}, {Stem: 6, StemName: "己", Days: 3}, {Stem: 8, StemName: "辛", Days: 18}},   // 丑月
}

// ComputeRenYuanSiLing returns the 人元司令分野 for the given solar month branch.
// If the date (year/month/day) is provided, also computes the current governing phase.
func ComputeRenYuanSiLing(solarMonthBranch Branch) RenYuanSiLing {
	phases, ok := renYuanTable[int(solarMonthBranch)]
	if !ok {
		phases = []RenYuanPhase{}
	}
	return RenYuanSiLing{
		MonthBranch: int(solarMonthBranch),
		Phases:      phases,
	}
}

// ComputeRenYuanSiLingForDate returns the current governing phase for a given date.
func ComputeRenYuanSiLingForDate(year, month, day int) RenYuanSiLing {
	yp := YearPillar(year, month, day)
	birthTime := time.Date(year, time.Month(month), day, 12, 0, 0, 0, time.UTC)
	result := ComputeRenYuanSiLing(MonthPillar(birthTime, yp.Stem).Branch)

	// Determine days into the solar month.
	jqDepth := ComputeJieQiDepth(year, month, day)
	daysIn := jqDepth.DaysIn

	// Accumulate through phases to find current.
	cumulative := 0
	for _, p := range result.Phases {
		cumulative += p.Days
		if daysIn < cumulative {
			cp := p
			result.Current = &cp
			break
		}
	}
	// If beyond all phases, last phase governs.
	if result.Current == nil && len(result.Phases) > 0 {
		last := result.Phases[len(result.Phases)-1]
		result.Current = &last
	}

	return result
}
