package huangli

import (
	"time"
	"liki/internal/engine/ganzhi"
)

type BondDay struct {
	Day
	GanRelation    string `json:"gan_relation"`
	ZhiRelation    string `json:"zhi_relation"`
	TaiSuiRelation string `json:"tai_sui_relation"`
}

type BondMonth struct {
	Month  string    `json:"month"`
	Stem   string    `json:"stem"`
	Branch string    `json:"branch"`
	Days   []BondDay `json:"days"`
}

func CrossDate(dayMaster ganzhi.Gan, dayBranch ganzhi.Zhi, dateStr string, eventType string) (BondDay, error) {
	dayEntry, err := QueryDate(dateStr, eventType)
	if err != nil { return BondDay{}, err }
	dayStem := dayEntry.DayPillar.Gan
	dayZhi := dayEntry.DayPillar.Zhi
	t, _ := time.Parse("2006-01-02", dateStr) //nolint:errcheck
	taiSui := taiSui(t.Year())
	relZhi, _, _ := evaluateZhi(dayZhi, dayBranch, "日柱")
	relTS, _, _ := evaluateZhi(dayZhi, taiSui, "太岁")
	return BondDay{Day: dayEntry, GanRelation: ganzhi.TenGodFromGan(dayMaster, dayStem), ZhiRelation: relZhi, TaiSuiRelation: relTS}, nil
}

func CrossMonth(dayMaster ganzhi.Gan, dayBranch ganzhi.Zhi, yearMonth string, eventType string) (BondMonth, error) {
	m, err := QueryMonth(yearMonth, eventType)
	if err != nil { return BondMonth{}, err }
	t, _ := time.Parse("2006-01", yearMonth) //nolint:errcheck
	taiSui := taiSui(t.Year())
	r := BondMonth{Month: m.Month, Stem: m.Stem, Branch: m.Branch, Days: make([]BondDay, len(m.Days))}
	for i, e := range m.Days {
		ds := e.DayPillar.Gan; dz := e.DayPillar.Zhi
		r.Days[i] = BondDay{Day: e}
		r.Days[i].GanRelation = ganzhi.TenGodFromGan(dayMaster, ds)
		r.Days[i].ZhiRelation, _, _ = evaluateZhi(dz, dayBranch, "日柱")
		r.Days[i].TaiSuiRelation, _, _ = evaluateZhi(dz, taiSui, "太岁")
	}
	return r, nil
}
