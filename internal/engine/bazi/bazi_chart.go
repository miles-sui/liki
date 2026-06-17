package bazi

import (
	"liki/internal/engine/ganzhi"
	"liki/internal/engine/tianwen"
)

// --- per-pillar data ---

type hiddenStemsOut struct {
	Main  ganzhi.Gan
	Mid   *ganzhi.Gan
	Minor *ganzhi.Gan
}

type pillarInfo struct {
	Gan ganzhi.Gan; Zhi ganzhi.Zhi; NaYin string
	HiddenStems hiddenStemsOut; TenGods []tenGodEntry; LifeStages []lifeStageEntry
	ShenSha []shenShaEntry; IsVoid, IsSelfHe, IsKuiGang bool; SelfHeName string
}
type tenGodEntry struct{  TenGod, Name, Source string; Gan ganzhi.Gan }
type lifeStageEntry struct{ Stage, Name string; Gan ganzhi.Gan }
type stageOut struct{ Name string; Index int }
type naYinRelation struct{ A, B, Relation string }
type daYunPillars struct{ startAge int; direction string; pillars []ganzhi.Zhu }

// Ten god source constants.
const (
	sourceGan    = "stem"
	sourceMainQi = "main_qi"
	sourceMidQi  = "mid_qi"
	sourceMinQi  = "minor_qi"
)

// --- chart ---

type ChartBase struct {
	Year, Month, Day, Hour pillarInfo
	DayMaster       ganzhi.Gan
	FuYi            FuYi; TiaoHou TiaoHou
	WuxingCount     map[ganzhi.Wuxing]int
	LifeStages      [12]stageOut
	DaYun           *DaYun
	TaiYuanMingGong TaiYuanMingGong
}
type Chart struct {
	ChartBase
	HeHui    []TripleHeFull; GongJia []GongJia; SanQiName string
	WangShuai map[string]string; NayinRel []naYinRelation
}
var pillarNames = [4]string{"year", "month", "day", "hour"}
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

func computeDaYunPillars(st tianwen.SolarTime, month ganzhi.Zhu, nianGan ganzhi.Gan, gender ganzhi.Gender) daYunPillars {
	// TODO: implement大运起运算法
	return daYunPillars{}
}
