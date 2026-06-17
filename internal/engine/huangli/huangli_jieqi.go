package huangli

import (
	"time"

	"liki/internal/engine/ganzhi"
	"liki/internal/engine/tianwen"
)

// jieQiDepth describes the current solar term position for a given date.
type jieQiDepth struct {
	TermName     string `json:"term_name"`      // current solar term name
	DaysIn       int    `json:"days_in"`        // days into current term
	NextTermName string `json:"next_term_name"` // next solar term name
	DaysToNext   int    `json:"days_to_next"`   // days until next term
}

// Solar term names in order (24 terms, starting from 立春).
var jieQiNames = [24]string{
	"立春", "雨水", "惊蛰", "春分", "清明", "谷雨",
	"立夏", "小满", "芒种", "夏至", "小暑", "大暑",
	"立秋", "处暑", "白露", "秋分", "寒露", "霜降",
	"立冬", "小雪", "大雪", "冬至", "小寒", "大寒",
}

// computeJieQiDepth returns which solar term the given date falls in and
// how many days into it we are. termIndex is 0-23.
func computeJieQiDepth(year, month, day int) jieQiDepth {
	date := time.Date(year, time.Month(month), day, 12, 0, 0, 0, time.UTC)

	// Find current and next term.
	var prevIdx, nextIdx int
	var prevDate, nextDate time.Time

	for i := 0; i < 24; i++ {
		t := tianwen.SolarTermTime(year, tianwen.JieQiLongitudes[i])
		tDate := time.Date(t.Year(), t.Month(), t.Day(), 12, 0, 0, 0, time.UTC)
		if tDate.After(date) {
			nextIdx = i
			nextDate = tDate
			if i == 0 {
				prevIdx = 23
				prevDate = time.Date(tianwen.SolarTermTime(year-1, tianwen.JieQiLongitudes[23]).Year(),
					tianwen.SolarTermTime(year-1, tianwen.JieQiLongitudes[23]).Month(),
					tianwen.SolarTermTime(year-1, tianwen.JieQiLongitudes[23]).Day(), 12, 0, 0, 0, time.UTC)
			} else {
				prevIdx = i - 1
				tPrev := tianwen.SolarTermTime(year, tianwen.JieQiLongitudes[i-1])
				prevDate = time.Date(tPrev.Year(), tPrev.Month(), tPrev.Day(), 12, 0, 0, 0, time.UTC)
			}
			break
		}
	}
	if nextDate.IsZero() {
		// Date is after the last term (大寒) — wrap to next year.
		prevIdx = 23
		tPrev := tianwen.SolarTermTime(year, tianwen.JieQiLongitudes[23])
		prevDate = time.Date(tPrev.Year(), tPrev.Month(), tPrev.Day(), 12, 0, 0, 0, time.UTC)
		nextIdx = 0
		tNext := tianwen.SolarTermTime(year+1, tianwen.JieQiLongitudes[0])
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

	return jieQiDepth{
		TermName:     jieQiNames[prevIdx],
		DaysIn:       daysIn,
		NextTermName: jieQiNames[nextIdx],
		DaysToNext:   daysToNext,
	}
}

// renYuanSiLing describes which hidden stem governs during each portion of a month.
type renYuanSiLing struct {
	MonthBranch ganzhi.Zhi                  `json:"month_branch"`
	Phases      []ganzhi.RenYuanPhase `json:"phases"`
	Current     *ganzhi.RenYuanPhase  `json:"current"` // current governing stem, if date provided
}

// computeRenYuanSiLing returns the 人元司令分野 for the given solar month branch.
func computeRenYuanSiLing(solarMonthBranch ganzhi.Zhi) renYuanSiLing {
	phases := ganzhi.RenYuanPhasesForBranch(solarMonthBranch)
	if phases == nil {
		phases = []ganzhi.RenYuanPhase{}
	}
	return renYuanSiLing{
		MonthBranch: solarMonthBranch,
		Phases:      phases,
	}
}

// computeRenYuanSiLingForDate returns the current governing phase for the given
// solar month branch. daysIn is the number of days into the current solar term,
// obtained from computeJieQiDepth.
func computeRenYuanSiLingForDate(monthBranch ganzhi.Zhi, daysIn int) renYuanSiLing {
	result := computeRenYuanSiLing(monthBranch)
	result.Current = findCurrentRenYuanPhase(result.Phases, daysIn)
	return result
}

// findCurrentRenYuanPhase finds the current governing phase given the phase table
// and the number of days into the solar month. Returns nil if phases is empty.
func findCurrentRenYuanPhase(phases []ganzhi.RenYuanPhase, daysIn int) *ganzhi.RenYuanPhase {
	cumulative := 0
	for _, p := range phases {
		cumulative += p.Days
		if daysIn < cumulative {
			cp := p
			return &cp
		}
	}
	if len(phases) > 0 {
		last := phases[len(phases)-1]
		return &last
	}
	return nil
}
