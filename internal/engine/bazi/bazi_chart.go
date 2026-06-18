package bazi

import (
	"time"

	"liki/internal/engine/ganzhi"
	"liki/internal/engine/tianwen"
)

// --- per-pillar data ---

type hiddenStemsOut struct {
	Main  ganzhi.Gan
	Mid   *ganzhi.Gan
	Minor *ganzhi.Gan
}

type zhuInfo struct {
	Gan ganzhi.Gan; Zhi ganzhi.Zhi; NaYin string
	HiddenStems hiddenStemsOut; TenGods []tenGodEntry; ChangSheng []changShengEntry
	ShenSha []shenShaEntry; IsVoid, IsSelfHe, IsKuiGang bool; SelfHeName string
}
type tenGodEntry struct{  TenGod ganzhi.TenGod; Name, Source string; Gan ganzhi.Gan }
type changShengEntry struct{ Stage, Name string; Gan ganzhi.Gan }
type stageOut struct{ Name string; Index ganzhi.Zhi }
type naYinRelation struct{ A, B, Relation string }
type daYunZhus struct{ startAge int; direction string; zhus []ganzhi.Zhu }

// Ten god source constants.
const (
	sourceGan    = "stem"
	sourceMainQi = "main_qi"
	sourceMidQi  = "mid_qi"
	sourceMinQi  = "minor_qi"
)

// --- chart ---

// ChartBase holds the core bazi chart data without optional decorations.
type ChartBase struct {
	Year, Month, Day, Hour zhuInfo
	SolarTime       tianwen.SolarTime
	DayMaster       ganzhi.Gan
	FuYi            FuYi; TiaoHou TiaoHou
	WuxingCount     map[ganzhi.Wuxing]int
	ChangSheng      [12]stageOut
	DaYun           *DaYun
	TaiYuanMingGong TaiYuanMingGong
}
// Chart holds a complete bazi chart including he-hui, gong-jia, and wang-shuai analysis.
type Chart struct {
	ChartBase
	HeHui    []TripleHeFull; GongJia []GongJia; SanQiName string
	WangShuai map[string]string; NayinRel []naYinRelation
}
var zhuNames = [4]string{"year", "month", "day", "hour"}
func (cb ChartBase) ToBazi() ganzhi.Bazi {
	return ganzhi.Bazi{
		Nian: ganzhi.Zhu{Gan: cb.Year.Gan, Zhi: cb.Year.Zhi},
		Yue:  ganzhi.Zhu{Gan: cb.Month.Gan, Zhi: cb.Month.Zhi},
		Ri:   ganzhi.Zhu{Gan: cb.Day.Gan, Zhi: cb.Day.Zhi},
		Shi:  ganzhi.Zhu{Gan: cb.Hour.Gan, Zhi: cb.Hour.Zhi},
	}
}
func (cb ChartBase) NaYinArray() [4]string {
	return [4]string{cb.Year.NaYin, cb.Month.NaYin, cb.Day.NaYin, cb.Hour.NaYin}
}

func (cb ChartBase) HiddenStemsArray() [4]hiddenStemsOut {
	return [4]hiddenStemsOut{cb.Year.HiddenStems, cb.Month.HiddenStems, cb.Day.HiddenStems, cb.Hour.HiddenStems}
}

func computeDaYunZhus(st tianwen.SolarTime, month ganzhi.Zhu, nianGan ganzhi.Gan, gender ganzhi.Gender) daYunZhus {
	isYang := int(nianGan)%2 == 1
	isMale := gender == ganzhi.Male
	forward := (isMale && isYang) || (!isMale && !isYang)

	birthTime := st.Time()
	birthYear := birthTime.Year()
	jz := tianwen.JianYue(tianwen.GregorianTime(birthTime))
	mi := (int(jz) + 9) % 12 // 0=寅月..11=丑月

	// The jie index in JieQiLongitudes for month mi is mi*2.
	jieIdx := mi * 2
	var targetJie time.Time
	var dir string
	if forward {
		nextIdx := ((mi + 1) % 12) * 2
		targetJie = tianwen.SolarTermTime(birthYear, tianwen.JieQiLongitudes[nextIdx])
		if !targetJie.After(birthTime) {
			targetJie = tianwen.SolarTermTime(birthYear+1, tianwen.JieQiLongitudes[nextIdx])
		}
		dir = "顺排"
	} else {
		targetJie = tianwen.SolarTermTime(birthYear, tianwen.JieQiLongitudes[jieIdx])
		if targetJie.After(birthTime) {
			targetJie = tianwen.SolarTermTime(birthYear-1, tianwen.JieQiLongitudes[jieIdx])
		}
		dir = "逆排"
	}

	days := targetJie.Sub(birthTime).Hours() / 24
	if days < 0 {
		days = -days
	}
	startAge := int(days/3 + 0.5) // 3 days = 1 year, round to nearest

	// Generate 8 zhus from month pillar.
	monthIdx := ganzhi.SixtyCycleName(month.Gan, month.Zhi)
	zhus := make([]ganzhi.Zhu, 0, 8)
	for i := 1; i <= 8; i++ {
		var idx int
		if forward {
			idx = (monthIdx + i) % 60
		} else {
			idx = (monthIdx - i + 60) % 60
		}
		g := ganzhi.Gan((idx % 10) + 1)
		z := ganzhi.Zhi((idx % 12) + 1)
		zhus = append(zhus, ganzhi.Zhu{Gan: g, Zhi: z})
	}

	return daYunZhus{
		startAge:  startAge,
		direction: dir,
		zhus:   zhus,
	}
}
