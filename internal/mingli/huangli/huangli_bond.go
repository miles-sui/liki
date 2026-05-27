package huangli

import (
	"time"
)

// BondDayEntry is a DayEntry with personal cross-reference annotations.
type BondDayEntry struct {
	DayEntry
	GanRelation    string `json:"gan_relation"`
	ZhiRelation    string `json:"zhi_relation"`
	TaiSuiRelation string `json:"tai_sui_relation"`
}

// CrossDate returns a single-day bond result.
// Theory requires only dayMaster and dayBranch — full birth chart is not needed.
func CrossDate(dayMaster Stem, dayBranch Branch, dateStr string, eventType string) (BondDayEntry, error) {
	dayEntry, err := QueryDate(dateStr, eventType)
	if err != nil {
		return BondDayEntry{}, err
	}

	dayStem := Stem(dayEntry.DayPillar.Stem)
	dayZhi := Branch(dayEntry.DayPillar.Branch)

	t, _ := time.Parse("2006-01-02", dateStr)
	taiSui := TaiSui(t.Year())

	result := BondDayEntry{DayEntry: dayEntry}
	result.GanRelation = EvaluateGan(dayStem, dayMaster)
	result.ZhiRelation, _, _ = evaluateZhiWithLabels(dayZhi, dayBranch, "日柱")
	result.TaiSuiRelation, _, _ = evaluateZhiWithLabels(dayZhi, taiSui, "太岁")

	return result, nil
}

// CrossMonth returns bond results for every day in a month.
func CrossMonth(dayMaster Stem, dayBranch Branch, yearMonth string, eventType string) ([]BondDayEntry, error) {
	entries, err := QueryMonth(yearMonth, eventType)
	if err != nil {
		return nil, err
	}

	t, _ := time.Parse("2006-01", yearMonth)
	taiSui := TaiSui(t.Year())

	result := make([]BondDayEntry, len(entries))
	for i, e := range entries {
		dayStem := Stem(e.DayPillar.Stem)
		dayZhi := Branch(e.DayPillar.Branch)

		result[i] = BondDayEntry{DayEntry: e}
		result[i].GanRelation = EvaluateGan(dayStem, dayMaster)
		result[i].ZhiRelation, _, _ = evaluateZhiWithLabels(dayZhi, dayBranch, "日柱")
		result[i].TaiSuiRelation, _, _ = evaluateZhiWithLabels(dayZhi, taiSui, "太岁")
	}
	return result, nil
}

// evaluateZhiWithLabels is EvaluateZhi but the label prefix is always the reference (日柱/太岁).
func evaluateZhiWithLabels(dayZhi, refZhi Branch, label string) (relation string, marks []string, warnings []string) {
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
