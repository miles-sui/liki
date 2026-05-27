package bazi

import (
	"time"

	"github.com/25types/25types/internal/tianwen"
)

// YearPillar computes the year pillar for a given date, accounting for 立春 boundary.
// If the date is before 立春, the year stem/branch is based on (year-1).
func YearPillar(year, month, day int) Pillar {
	lcM, lcD := LiChunDay(year)
	if month < lcM || (month == lcM && day < lcD) {
		year--
	}
	s := (year - 3) % 10
	if s <= 0 {
		s += 10
	}
	stem := Stem(s)

	b := (year - 3) % 12
	if b <= 0 {
		b += 12
	}
	branch := Branch(b)
	return Pillar{Stem: stem, Branch: branch}
}

// MonthPillar computes the month pillar based on the solar term (节气) at the given birth time.
// solarMonthIndex returns 0-based (0=寅月), but the 五虎遁 formula expects 1-based monthNum (寅=1).
func MonthPillar(birthTime time.Time, yearStem Stem) Pillar {
	mi := SolarMonthIndex(birthTime)            // 0=寅月 .. 11=丑月
	monthNum := mi + 1                             // 1=寅月 .. 12=丑月
	branch := Branch((mi+2)%12 + 1)                // 0→寅(3), 11→丑(2)
	stem := Stem(((int(yearStem)*2 + monthNum) % 10))
	if stem == 0 {
		stem = 10
	}
	return Pillar{Stem: stem, Branch: branch}
}

// DayPillar computes the day pillar using the Julian Day method.
// 1900-01-01 = 甲戌 (0-based index=10).
func DayPillar(year, month, day int) Pillar {
	jd := JulianDay(year, month, day)
	baseJD := JulianDay(1900, 1, 1)
	diff := jd - baseJD
	gzIndex := (10 + diff) % 60
	if gzIndex < 0 {
		gzIndex += 60
	}
	stem := Stem(gzIndex%10 + 1)
	branch := Branch(gzIndex%12 + 1)
	return Pillar{Stem: stem, Branch: branch}
}

// HourPillar computes the hour pillar using 五鼠遁 from day stem and solar time.
func HourPillar(solarTime float64, dayStem Stem) Pillar {
	branch := HourBranchFromSolarTime(solarTime)
	stem := Stem(((int(dayStem)*2 + int(branch) - 2) % 10))
	if stem == 0 {
		stem = 10
	}
	return Pillar{Stem: stem, Branch: branch}
}

// BaziResult is the minimal output from birth info: the four pillars only.
// All other chart data (ten gods, nayin, hidden stems, etc.) is derived from these pillars.
type BaziResult struct {
	Bazi
	SolarTime float64
	SolarDate time.Time
	BaziDate  time.Time
}

// ComputeSolarTime computes the true solar time (真太阳时) in minutes from Gregorian birth info.
func ComputeSolarTime(year, month, day, hour, minute int, longitude, timezone float64, isDST bool) float64 {
	tzDeg := timezone * 15
	return tianwen.ComputeSolarTime(year, month, day, hour, minute, longitude, tzDeg, isDST)
}

// ComputeBazi computes the four pillars (八字) from true solar time and birth date.
func ComputeBazi(solarTime float64, year, month, day, hour, minute int, timezone float64, isDST bool) BaziResult {
	adjYear, adjMonth, adjDay := year, month, day
	if solarTime >= 1380 {
		t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC).AddDate(0, 0, 1)
		adjYear, adjMonth, adjDay = t.Year(), int(t.Month()), t.Day()
	}

	offsetSec := int(timezone * 3600)
	loc := time.FixedZone("birth", offsetSec)
	birthLocal := time.Date(year, time.Month(month), day, hour, minute, 0, 0, loc)

	lstMinutes := float64(hour*60 + minute)
	if isDST {
		lstMinutes -= 60
	}
	solarDate := time.Date(year, time.Month(month), day, 0, 0, 0, 0, loc).Add(time.Duration(solarTime-lstMinutes) * time.Minute)

	baziDate := time.Date(adjYear, time.Month(adjMonth), adjDay, 0, 0, 0, 0, loc)

	yp := YearPillar(year, month, day)
	mp := MonthPillar(birthLocal.UTC(), yp.Stem)
	dp := DayPillar(adjYear, adjMonth, adjDay)
	hp := HourPillar(solarTime, dp.Stem)

	return BaziResult{
		Bazi: Bazi{Year: yp, Month: mp, Day: dp, Hour: hp},
		SolarTime:   solarTime,
		SolarDate:   solarDate,
		BaziDate:    baziDate,
	}
}

// ComputeChartFromBirth is the single entry point for computing a full chart from raw birth data.
// Longitude defaults to 120 (Beijing), timezone to 8 (UTC+8).
func ComputeChartFromBirth(year, month, day, hour, minute int, longitude, timezone float64, gender Gender) ChartResult {
	if longitude == 0 {
		longitude = 120
	}
	if timezone == 0 {
		timezone = 8
	}
	isDST := IsDST(year, month, day)
	ast := ComputeSolarTime(year, month, day, hour, minute, longitude, timezone, isDST)
	bz := ComputeBazi(ast, year, month, day, hour, minute, timezone, isDST)
	return ComputeChart(bz, year, month, day, gender)
}

// ComputeChart produces a full BaZi chart from four pillars and birth metadata.
func ComputeChart(bz BaziResult, year, month, day int, gender Gender) ChartResult {
	dm := bz.Day.Stem

	hs := computeHiddenStems(bz.Bazi)
	ny := computeNaYin(bz.Bazi)
	ls := computeLifeStages(dm)
	bf := computeDayun(year, month, day, bz.Month, gender, bz.Year.Stem)
	ec := computeElementCount(bz.Bazi, hs)

	tgTable := ComputeTenGodsTable(dm, bz.Bazi, hs)
	lsTable := ComputeLifeStageTable(bz.Bazi, hs)
	shensha := ComputeShenSha(bz.Bazi, dm, bz.Month.Branch)
	voidHits := ComputeKongWang(bz.Day, bz.Bazi)
	ps := bz.Slice()

	makePI := func(i int) PillarInfo {
		isVoid := false
		for _, vh := range voidHits {
			if vh == i {
				isVoid = true
				break
			}
		}
		return PillarInfo{
			Stem:        ps[i].Stem,
			Branch:      ps[i].Branch,
			NaYin:       ny[i],
			HiddenStems: hs[i],
			TenGods:     tgTable[i],
			LifeStages:  lsTable[i],
			ShenSha:     shensha[i],
			IsVoid:      isVoid,
			IsSelfHe:    IsSelfHe(ps[i]),
			SelfHeName:  SelfHeName(ps[i]),
			IsKuiGang:   IsKuiGang(ps[i]),
		}
	}

	return ChartResult{
		Year:         makePI(0),
		Month:        makePI(1),
		Day:          makePI(2),
		Hour:         makePI(3),
		SolarTime:    bz.SolarTime,
		SolarDate:    bz.SolarDate,
		BaziDate:     bz.BaziDate,
		LifeStages:   ls,
		Dayun:        bf,
		DayMaster:    dm,
		ElementCount: ec,
	}
}
