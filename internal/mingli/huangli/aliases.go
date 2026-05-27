package huangli

import (
	"time"

	"github.com/25types/25types/internal/ganzhi"
	"github.com/25types/25types/internal/mingli/bazi"
	"github.com/25types/25types/internal/tianwen"
)

// -- type aliases --

type Stem = ganzhi.Stem
type Branch = ganzhi.Branch
type Pillar = ganzhi.Pillar

// -- variable aliases --

var jieQiLongitudes = tianwen.JieQiLongitudes

// -- unexported function wrappers --

func sixtyCycleName(st Stem, br Branch) int { return ganzhi.SixtyCycleName(st, br) }

// -- exported function wrappers --

func StemElement(s Stem) ganzhi.Element   { return ganzhi.StemElement(s) }
func StemYinYang(s Stem) ganzhi.YinYang   { return ganzhi.StemYinYang(s) }

func solarTermDate(year int, targetLon float64) time.Time { return tianwen.SolarTermTime(year, targetLon) }

// -- bazi function wrappers --

func YearPillar(year, month, day int) Pillar          { return bazi.YearPillar(year, month, day) }
func MonthPillar(t time.Time, yearStem Stem) Pillar   { return bazi.MonthPillar(t, yearStem) }
func DayPillar(year, month, day int) Pillar           { return bazi.DayPillar(year, month, day) }
func TenGodName(tg int) string                        { return bazi.TenGodName(tg) }
func TenGodType(dmElem ganzhi.Element, dmYY ganzhi.YinYang, otherElem ganzhi.Element, otherYY ganzhi.YinYang) int {
	return bazi.TenGodType(dmElem, dmYY, otherElem, otherYY)
}
func IsBranchHe(b1, b2 Branch) bool  { return ganzhi.IsBranchHe(b1, b2) }
func IsTripleHe(b1, b2 Branch) bool  { return ganzhi.IsTripleHe(b1, b2) }
func IsTripleHui(b1, b2 Branch) bool { return ganzhi.IsTripleHui(b1, b2) }
func IsLiuChong(b1, b2 Branch) bool  { return ganzhi.IsLiuChong(b1, b2) }
func IsXing(b1, b2 Branch) bool      { return ganzhi.IsXing(b1, b2) }
func IsHai(b1, b2 Branch) bool       { return ganzhi.IsHai(b1, b2) }
func NaYinString(s Stem, b Branch) string { return bazi.NaYinString(s, b) }
