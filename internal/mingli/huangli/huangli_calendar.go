package huangli

import (
	"fmt"
	"time"
)

// --- Types ---

// DayPillarInfo holds the stem-branch for a single day.
type DayPillarInfo struct {
	Stem   Stem   `json:"gan"`
	Branch Branch `json:"zhi"`
	NaYin  string `json:"na_yin"`
}

// --- In-memory config ---

type JianchuConfig struct {
	Sequence   []string                     `yaml:"sequence"`
	Suitable   map[string][]string          `yaml:"suitable"`
	Forbidden  map[string][]string          `yaml:"forbidden"`
	EventRules map[string]struct {
		Label     string   `yaml:"label"`
		Suitable  []string `yaml:"suitable"`
		Forbidden []string `yaml:"forbidden"`
	} `yaml:"event_rules"`
	ShenSha map[string]map[string][]string `yaml:"shensha"`
}

// --- Engine Functions ---

// TaiSui returns the year's presiding branch (太岁).
func TaiSui(year int) Branch {
	// Year pillar branch = (year - 3) % 12, with 1=子.
	b := ((year - 3) % 12) + 1
	if b <= 0 {
		b += 12
	}
	return Branch(b)
}

// LookupDayPillar returns the stem-branch and na-yin for a date string (YYYY-MM-DD).
func LookupDayPillar(date string) (DayPillarInfo, error) {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return DayPillarInfo{}, fmt.Errorf("dates: parse date %s: %w", date, err)
	}
	p := DayPillar(t.Year(), int(t.Month()), t.Day())
	return DayPillarInfo{Stem: p.Stem, Branch: p.Branch, NaYin: NaYinString(p.Stem, p.Branch)}, nil
}

// LookupJianChu returns the JianChu (建除) god for a date string.
func LookupJianChu(date string) string {
	e := defaultEngine
	if e == nil {
		return ""
	}
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return ""
	}
	yp := YearPillar(t.Year(), int(t.Month()), t.Day())
	mp := MonthPillar(t, yp.Stem)
	monthBranch := mp.Branch

	dp := DayPillar(t.Year(), int(t.Month()), t.Day())

	jianIdx := int(monthBranch) - 1
	dayIdx := int(dp.Branch) - 1

	offset := (dayIdx - jianIdx + 12) % 12
	return e.JianChuConfig.Sequence[offset]
}

// JianChuSuitable checks if the jianchu god is suitable for the event type.
func JianChuSuitable(jianChu, eventType string) (suitable bool, marks []string, warnings []string) {
	e := defaultEngine
	if e == nil {
		return true, nil, nil
	}
	rule, ok := e.JianChuConfig.EventRules[eventType]
	if !ok {
		return true, nil, nil
	}

	for _, s := range rule.Suitable {
		if jianChu == s {
			suitable = true
			marks = append(marks, jianChu+"日宜"+rule.Label)
			break
		}
	}
	for _, f := range rule.Forbidden {
		if jianChu == f {
			suitable = false
			if f == "破" {
				warnings = append(warnings, "破日，万事不宜")
			} else {
				warnings = append(warnings, jianChu+"日忌"+rule.Label)
			}
			break
		}
	}
	return
}

// EvaluateGan returns the ten god name for day stem vs day master.
func EvaluateGan(dayStem, dayMaster Stem) string {
	dmElem := StemElement(dayMaster)
	dayElem := StemElement(dayStem)
	dmYY := StemYinYang(dayMaster)
	dayYY := StemYinYang(dayStem)

	tg := TenGodType(dmElem, dmYY, dayElem, dayYY)
	return TenGodName(tg)
}

// EvaluateZhi checks the branch relationship and returns marks/warnings.
func EvaluateZhi(dayZhi, refZhi Branch, label string) (relation string, marks []string, warnings []string) {
	switch {
	case IsBranchHe(dayZhi, refZhi):
		return "六合", []string{label + "六合日"}, nil
	case IsTripleHe(dayZhi, refZhi):
		return "三合半", []string{label + "三合"}, nil
	case IsTripleHui(dayZhi, refZhi):
		return "三会半", []string{label + "三会"}, nil
	case IsLiuChong(dayZhi, refZhi):
		return "六冲", nil, []string{"冲" + label}
	case IsXing(dayZhi, refZhi):
		return "相刑", nil, []string{"刑" + label}
	case IsHai(dayZhi, refZhi):
		return "六害", nil, []string{"害" + label}
	}
	return "无", nil, nil
}

// MonthPillarForDate returns the month pillar for a given date.
func MonthPillarForDate(t time.Time) Pillar {
	yp := YearPillar(t.Year(), int(t.Month()), t.Day())
	return MonthPillar(t, yp.Stem)
}


