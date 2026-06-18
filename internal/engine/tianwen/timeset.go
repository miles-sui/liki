package tianwen

import "time"


// Timeset packs three calendar representations of a moment.
type Timeset struct {
	Gregorian GregorianTime
	Solar     SolarTime
	Lunar     LunarTime
}

// ComputeTime converts Gregorian time + coordinates into all three
// calendar representations (Gregorian, Solar, Lunar). Longitude is in
// degrees, timezone is in hours.
func ComputeTime(year, month, day, hour, minute int, longitude, timezone float64) Timeset {
	st := ComputeSolarTime(year, month, day, hour, minute, longitude, timezone)
	// Use solar-time-adjusted date for lunar calendar lookup,
	// consistent with ComputeBazi's date adjustment.
	lt := SolarToLunar(GregorianTime(st.Time()))
	lt.Shichen = hourBranchFromSolarTime(st.Minutes())
	return Timeset{
		Gregorian: GregorianTime(
			time.Date(year, time.Month(month), day, hour, minute, 0, 0,
				time.FixedZone("greg", int(timezone*3600))),
		),
		Solar: st,
		Lunar: lt,
	}
}
