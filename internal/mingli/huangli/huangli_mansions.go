package huangli

import "github.com/25types/25types/internal/mingli/bazi"

// MansionForDay returns the 28-mansion entry for a given day pillar.
func MansionForDay(dayPillar Pillar) bazi.DayMansion {
	return bazi.MansionForDay(dayPillar)
}

// AllMansions returns the full 28-mansion table in order.
func AllMansions() [28]bazi.DayMansion {
	return bazi.AllMansions()
}
