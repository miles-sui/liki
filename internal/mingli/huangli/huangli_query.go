package huangli

import (
	"fmt"
	"time"

	"github.com/25types/25types/internal/mingli/bazi"
)

// DayEntry is the huangli query result for one day.
type DayEntry struct {
	Date       string            `json:"date"`
	DayPillar  DayPillarInfo     `json:"day_pillar"`
	JianChu    string            `json:"jian_chu"`
	Suitable   bool              `json:"suitable"`
	Marks      []string          `json:"marks"`
	Warnings   []string          `json:"warnings"`
	HuangDao   HuangDaoStar      `json:"huangdao"`
	Directions DayStemDirections `json:"directions"`
	Taboos     DayTaboos         `json:"taboos"`
	Mansion    bazi.DayMansion   `json:"mansion"`
}

// QueryDate returns huangli info for a single date.
func QueryDate(dateStr string, eventType string) (DayEntry, error) {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return DayEntry{}, fmt.Errorf("huangli: parse date %s: %w", dateStr, err)
	}

	dp, err := LookupDayPillar(dateStr)
	if err != nil {
		return DayEntry{}, err
	}

	monthBranch := MonthPillarForDate(t).Branch

	entry := DayEntry{
		Date:       dateStr,
		DayPillar:  dp,
		JianChu:    LookupJianChu(dateStr),
		HuangDao:   HuangDaoForDay(monthBranch, Branch(dp.Branch)),
		Directions: ComputeDayDirections(Stem(dp.Stem)),
		Taboos:     ComputeDayTaboos(Pillar{Stem: Stem(dp.Stem), Branch: Branch(dp.Branch)}),
		Mansion:    MansionForDay(Pillar{Stem: Stem(dp.Stem), Branch: Branch(dp.Branch)}),
	}

	if eventType != "" {
		entry.Suitable, entry.Marks, entry.Warnings = JianChuSuitable(entry.JianChu, eventType)
	}

	return entry, nil
}

// QueryMonth returns huangli entries for every day in the given month.
func QueryMonth(yearMonth string, eventType string) ([]DayEntry, error) {
	t, err := time.Parse("2006-01", yearMonth)
	if err != nil {
		return nil, fmt.Errorf("huangli: parse year-month %s: %w", yearMonth, err)
	}

	year, month := t.Year(), int(t.Month())
	daysInMonth := time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.UTC).Day()

	var entries []DayEntry
	for d := 1; d <= daysInMonth; d++ {
		dateStr := fmt.Sprintf("%04d-%02d-%02d", year, month, d)
		entry, err := QueryDate(dateStr, eventType)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}
