package tianwen

import "time"


// BirthTime packs three calendar representations of a birth moment.
type BirthTime struct {
	Gregorian GregorianTime
	Solar     SolarTime
	Lunar     LunarTime
}

// ComputeBirthTime computes all three calendar representations from a
// Gregorian birth time. Longitude is in degrees, timezone is in hours.
func ComputeBirthTime(year, month, day, hour, minute int, longitude, timezone float64) BirthTime {
	st := ComputeSolarTime(year, month, day, hour, minute, longitude, timezone)
	lt := SolarToLunar(year, month, day)
	lt.Shichen = hourBranchFromSolarTime(st.Minutes())
	return BirthTime{
		Gregorian: GregorianTime(
			time.Date(year, time.Month(month), day, hour, minute, 0, 0,
				time.FixedZone("greg", int(timezone*3600))),
		),
		Solar: st,
		Lunar: lt,
	}
}
