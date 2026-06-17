package tianwen

import (
	"time"

	"liki/internal/engine/ganzhi"
)

// DayPillar computes the day pillar for a given date using the Julian Day method.
// 1900-01-01 = 甲戌 (0-based index=10).
func DayPillar(year, month, day int) ganzhi.Zhu {
	jd := julianDay(year, month, day)
	baseJD := julianDay(1900, 1, 1)
	diff := jd - baseJD
	gzIndex := (10 + diff) % 60
	if gzIndex < 0 {
		gzIndex += 60
	}
	return ganzhi.Zhu{Gan: ganzhi.Gan(gzIndex%10 + 1), Zhi: ganzhi.Zhi(gzIndex%12 + 1)}
}

// YearPillar computes the year pillar for a given date, accounting for 立春 boundary.
// If the date is before 立春, the year stem/branch is based on (year-1).
func YearPillar(year, month, day int) ganzhi.Zhu {
	lcM, lcD := liChunDay(year)
	if month < lcM || (month == lcM && day < lcD) {
		year--
	}
	s := (year - 3) % 10
	if s <= 0 {
		s += 10
	}
	b := (year - 3) % 12
	if b <= 0 {
		b += 12
	}
	return ganzhi.Zhu{Gan: ganzhi.Gan(s), Zhi: ganzhi.Zhi(b)}
}

// MonthPillar computes the month pillar based on the solar term at the given time.
func MonthPillar(birthTime time.Time, yearGan ganzhi.Gan) ganzhi.Zhu {
	mi := SolarMonthIndex(birthTime) // 0=寅月 .. 11=丑月
	monthNum := mi + 1               // 1=寅月 .. 12=丑月
	branch := ganzhi.Zhi((mi+2)%12 + 1)
	stem := ganzhi.Gan(((int(yearGan)*2 + monthNum) % 10))
	if stem == 0 {
		stem = 10
	}
	return ganzhi.Zhu{Gan: stem, Zhi: branch}
}

// HourPillar computes the hour pillar from solar time and day stem.
func HourPillar(solarTime float64, dayGan ganzhi.Gan) ganzhi.Zhu {
	branch := hourBranchFromSolarTime(solarTime)
	stem := ganzhi.Gan(((int(dayGan)*2 + int(branch) - 2) % 10))
	if stem == 0 {
		stem = 10
	}
	return ganzhi.Zhu{Gan: stem, Zhi: branch}
}


func ComputeBazi(st SolarTime) ganzhi.Bazi {
	t := st.Time()
	y, m, d := t.Date()
	ast := float64(t.Hour()*60 + t.Minute())
	if ast >= 1380 { t = t.AddDate(0, 0, 1); y, m, d = t.Date() }
	yp := YearPillar(y, int(m), d)
	mp := MonthPillar(t.UTC(), yp.Gan)
	dp := DayPillar(y, int(m), d)
	hp := HourPillar(ast, dp.Gan)
	return ganzhi.Bazi{Nian: yp, Yue: mp, Ri: dp, Shi: hp}
}
