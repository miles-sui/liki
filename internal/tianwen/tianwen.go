package tianwen

import (
	"math"
	"sort"
	"time"

	"github.com/25types/25types/internal/ganzhi"
)

// -- solar term longitudes ---------------------------------------------------

// JieQiLongitudes holds the 24 solar term longitudes in order.
// Index 0=立春 315°, 1=雨水 330°, ..., 23=大寒 300°.
var JieQiLongitudes = [24]float64{
	315, 330, 345, 0, 15, 30,
	45, 60, 75, 90, 105, 120,
	135, 150, 165, 180, 195, 210,
	225, 240, 255, 270, 285, 300,
}

// SolarTermLongitudes holds the 12 major solar terms (节) that mark solar month boundaries.
var SolarTermLongitudes = func() [12]float64 {
	var a [12]float64
	for i := 0; i < 12; i++ {
		a[i] = JieQiLongitudes[i*2]
	}
	return a
}()

// -- solar month index -------------------------------------------------------

// SolarMonthIndex returns 0-based solar month index (0=寅月 .. 11=丑月) at the given time.
func SolarMonthIndex(t time.Time) int {
	lon := SolarLongitude(t)
	for i := 0; i < 12; i++ {
		curLon := SolarTermLongitudes[i]
		nextLon := SolarTermLongitudes[(i+1)%12]
		if curLon <= nextLon {
			if lon >= curLon && lon < nextLon {
				return i
			}
		} else {
			if lon >= curLon || lon < nextLon {
				return i
			}
		}
	}
	return 0
}

// -- Julian day --------------------------------------------------------------

// JulianDay returns the Julian Day Number at noon UTC for the given date.
func JulianDay(year, month, day int) int {
	if month <= 2 {
		year--
		month += 12
	}
	A := year / 100
	B := 2 - A + A/4
	return int(365.25*float64(year+4716)) + int(30.6001*float64(month+1)) + day + B - 1524
}

// DayOfYear returns the day-of-year (1=Jan 1).
func DayOfYear(year, month, day int) int {
	return dayOfYear(year, month, day)
}

// julianDay returns the Julian Day Number with fractional time for the given time.
func julianDay(t time.Time) float64 {
	y, m, d := t.Year(), int(t.Month()), t.Day()
	if m <= 2 {
		y--
		m += 12
	}
	A := y / 100
	B := 2 - A + A/4
	dayFrac := float64(t.Hour()*3600+t.Minute()*60+t.Second()) / 86400.0
	return float64(int(365.25*float64(y+4716))) + float64(int(30.6001*float64(m+1))) + float64(d) + float64(B) - 1524.5 + dayFrac
}

// -- solar longitude ---------------------------------------------------------

// SolarLongitude returns the apparent solar longitude in degrees at the given time.
func SolarLongitude(t time.Time) float64 {
	jd := julianDay(t)
	T := (jd - 2451545.0) / 36525.0
	T2 := T * T

	L0 := 280.46646 + 36000.76983*T + 0.0003032*T2
	M := 357.52911 + 35999.05029*T - 0.0001537*T2
	Mrad := M * math.Pi / 180.0

	C := (1.914602-0.004817*T-0.000014*T2)*math.Sin(Mrad) +
		(0.019993-0.000101*T)*math.Sin(2*Mrad) +
		0.000289*math.Sin(3*Mrad)

	lon := L0 + C
	lon = math.Mod(lon, 360)
	if lon < 0 {
		lon += 360
	}
	return lon
}

// -- solar term computation --------------------------------------------------

// SolarTermTime returns the UTC time when the sun reaches targetLon degrees in the given year.
func SolarTermTime(year int, targetLon float64) time.Time {
	termIdx := 0
	for i, lon := range SolarTermLongitudes {
		if math.Abs(lon-targetLon) < 0.01 {
			termIdx = i
			break
		}
	}
	approxDay := 35 + termIdx*15
	t := time.Date(year, 1, 1, 12, 0, 0, 0, time.UTC).AddDate(0, 0, approxDay)

	for iter := 0; iter < 20; iter++ {
		lon := SolarLongitude(t)
		diff := targetLon - lon
		if diff > 180 {
			diff -= 360
		} else if diff < -180 {
			diff += 360
		}
		if math.Abs(diff) < 0.01 {
			break
		}
		step := diff / 0.9856
		if step > 15 {
			step = 15
		} else if step < -15 {
			step = -15
		}
		t = t.Add(time.Duration(step*24*3600) * time.Second)
	}
	return t
}

// SolarTermDate returns the UTC date when the sun reaches targetLon degrees.
func SolarTermDate(year int, targetLon float64) (month, day int) {
	t := SolarTermTime(year, targetLon)
	return int(t.Month()), t.Day()
}

// LiChunDay returns the (month, day) of 立春 (315°) for the given year.
func LiChunDay(year int) (month, day int) {
	return SolarTermDate(year, 315)
}

// GetCurrentSolarMonth returns the current solar month ID (e.g. "寅月").
func GetCurrentSolarMonth(t time.Time) string {
	return ganzhi.SolarMonthOrder[SolarMonthIndex(t)]
}

// SolarTermEntry represents one solar term with its month ID and date.
type SolarTermEntry struct {
	MonthID string    `json:"month_id"`
	NameEN  string    `json:"name_en"`
	Date    time.Time `json:"date"`
}

// PrecomputeSolarTerms returns the 12 solar terms for the given year, sorted by date.
func PrecomputeSolarTerms(year int) []SolarTermEntry {
	var entries []SolarTermEntry
	for i, termLon := range SolarTermLongitudes {
		entries = append(entries, SolarTermEntry{
			MonthID: ganzhi.SolarMonthOrder[i],
			NameEN:  ganzhi.MonthNamesEN[ganzhi.SolarMonthOrder[i]],
			Date:    SolarTermTime(year, termLon),
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Date.Before(entries[j].Date)
	})
	return entries
}
