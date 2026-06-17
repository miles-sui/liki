package tianwen

import (
	"math"
	"time"
)

var JieQiLongitudes = [24]float64{315,330,345,0,15,30,45,60,75,90,105,120,135,150,165,180,195,210,225,240,255,270,285,300}

var solarTermLongitudes = func() [12]float64 {
	var a [12]float64
	for i := 0; i < 12; i++ { a[i] = JieQiLongitudes[i*2] }
	return a
}()

func SolarMonthIndex(t time.Time) int {
	lon := solarLongitude(t)
	for i := 0; i < 12; i++ {
		cl, nl := solarTermLongitudes[i], solarTermLongitudes[(i+1)%12]
		if cl <= nl { if lon >= cl && lon < nl { return i } } else { if lon >= cl || lon < nl { return i } }
	}
	return 0
}

func julianDay(year, month, day int) int {
	if month <= 2 { year--; month += 12 }
	A := year/100
	return int(365.25*float64(year+4716)) + int(30.6001*float64(month+1)) + day + (2-A+A/4) - 1524
}

func solarLongitude(t time.Time) float64 {
	jd := float64(julianDay(t.Year(), int(t.Month()), t.Day()))
	T := (jd-2451545.0)/36525.0; T2 := T*T
	L0 := 280.46646 + 36000.76983*T + 0.0003032*T2
	M := 357.52911 + 35999.05029*T - 0.0001537*T2
	Mrad := M * math.Pi / 180.0
	C := (1.914602-0.004817*T-0.000014*T2)*math.Sin(Mrad) + (0.019993-0.000101*T)*math.Sin(2*Mrad) + 0.000289*math.Sin(3*Mrad)
	lon := L0 + C
	lon = math.Mod(lon, 360)
	if lon < 0 { lon += 360 }
	return lon
}

func SolarTermTime(year int, targetLon float64) time.Time {
	ti := 0
	for i, lon := range solarTermLongitudes { if math.Abs(lon-targetLon) < 0.01 { ti = i; break } }
	t := time.Date(year, 1, 1, 12, 0, 0, 0, time.UTC).AddDate(0, 0, 35+ti*15)
	for iter := 0; iter < 20; iter++ {
		lon := solarLongitude(t)
		diff := targetLon - lon
		if diff > 180 { diff -= 360 } else if diff < -180 { diff += 360 }
		if math.Abs(diff) < 0.01 { break }
		step := diff / 0.9856
		if step > 15 { step = 15 } else if step < -15 { step = -15 }
		t = t.Add(time.Duration(step*24*3600) * time.Second)
	}
	return t
}

func solarTermDate(year int, targetLon float64) (month, day int) {
	t := SolarTermTime(year, targetLon)
	return int(t.Month()), t.Day()
}

func liChunDay(year int) (month, day int) { return solarTermDate(year, 315) }

func SolarTermIndex(year, month, day int) int {
	target := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	terms := allSolarTerms(year)
	for i := 0; i < 24; i++ {
		next := terms[(i+1)%24]
		if next.Before(terms[i]) { next = next.AddDate(1, 0, 0) }
		if !target.Before(terms[i]) && target.Before(next) { return i }
	}
	return 23
}

func allSolarTerms(year int) [24]time.Time {
	var terms [24]time.Time
	terms[0] = SolarTermTime(year-1, JieQiLongitudes[21])
	terms[1] = SolarTermTime(year-1, JieQiLongitudes[22])
	terms[2] = SolarTermTime(year-1, JieQiLongitudes[23])
	for i := 3; i < 12; i++ { terms[i] = SolarTermTime(year, JieQiLongitudes[i-3]) }
	for i := 12; i < 24; i++ { terms[i] = SolarTermTime(year, JieQiLongitudes[i-3]) }
	return terms
}
