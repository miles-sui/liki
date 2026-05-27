package bazi

import (
	"time"

	"github.com/25types/25types/internal/ganzhi"
	"github.com/25types/25types/internal/tianwen"
)

// -- type aliases --

type Stem = ganzhi.Stem
type Branch = ganzhi.Branch
type Element = ganzhi.Element
type YinYang = ganzhi.YinYang
type Pillar = ganzhi.Pillar
type Bazi = ganzhi.Bazi
type Gender = ganzhi.Gender

// -- constant aliases --

const (
	StemJia  = ganzhi.StemJia
	StemYi   = ganzhi.StemYi
	StemBing = ganzhi.StemBing
	StemDing = ganzhi.StemDing
	StemWu   = ganzhi.StemWu
	StemJi   = ganzhi.StemJi
	StemGeng = ganzhi.StemGeng
	StemXin  = ganzhi.StemXin
	StemRen  = ganzhi.StemRen
	StemGui  = ganzhi.StemGui
)

const (
	BranchZi   = ganzhi.BranchZi
	BranchChou = ganzhi.BranchChou
	BranchYin  = ganzhi.BranchYin
	BranchMao  = ganzhi.BranchMao
	BranchChen = ganzhi.BranchChen
	BranchSi   = ganzhi.BranchSi
	BranchWu   = ganzhi.BranchWu
	BranchWei  = ganzhi.BranchWei
	BranchShen = ganzhi.BranchShen
	BranchYou  = ganzhi.BranchYou
	BranchXu   = ganzhi.BranchXu
	BranchHai  = ganzhi.BranchHai
)

const (
	ElemWood  = ganzhi.ElemWood
	ElemFire  = ganzhi.ElemFire
	ElemEarth = ganzhi.ElemEarth
	ElemMetal = ganzhi.ElemMetal
	ElemWater = ganzhi.ElemWater
)

const (
	Yin  = ganzhi.Yin
	Yang = ganzhi.Yang
)

const (
	Male   = ganzhi.Male
	Female = ganzhi.Female
)

// -- variable aliases (unexported) --

var (
	hourRanges          = ganzhi.HourRanges
	solarTermLongitudes = tianwen.SolarTermLongitudes
)

// -- unexported function wrappers --

func sixtyCycleName(st Stem, br Branch) int { return ganzhi.SixtyCycleName(st, br) }
func stemNameStr(s Stem) string             { return ganzhi.StemNameStr(s) }
func branchNameStr(b Branch) string         { return ganzhi.BranchNameStr(b) }

// -- exported function wrappers --

func StemElement(s Stem) Element          { return ganzhi.StemElement(s) }
func StemYinYang(s Stem) YinYang          { return ganzhi.StemYinYang(s) }
func BranchElement(b Branch) Element      { return ganzhi.BranchElement(b) }
func Sheng(from, to Element) bool         { return ganzhi.Sheng(from, to) }
func Ke(from, to Element) bool            { return ganzhi.Ke(from, to) }
func Zodiac(b Branch) string              { return ganzhi.Zodiac(b) }
func BranchHourRange(b Branch) string     { return ganzhi.BranchHourRange(b) }

func solarMonthIndex(t time.Time) int                     { return tianwen.SolarMonthIndex(t) }
func julianDay(year, month, day int) int                  { return tianwen.JulianDay(year, month, day) }
func dayOfYear(year, month, day int) int                  { return tianwen.DayOfYear(year, month, day) }
func solarTermDate(year int, targetLon float64) time.Time { return tianwen.SolarTermTime(year, targetLon) }
func liChunDay(year int) (int, int)                       { return tianwen.LiChunDay(year) }
func solarLongitude(t time.Time) float64                  { return tianwen.SolarLongitude(t) }

func HourBranchFromSolarTime(astMinutes float64) Branch { return tianwen.HourBranchFromSolarTime(astMinutes) }
func IsDST(year, month, day int) bool                   { return tianwen.IsDST(year, month, day) }
func SolarMonthIndex(t time.Time) int                   { return tianwen.SolarMonthIndex(t) }
func JulianDay(year, month, day int) int                { return tianwen.JulianDay(year, month, day) }
func LiChunDay(year int) (int, int)                     { return tianwen.LiChunDay(year) }
func DayOfYear(year, month, day int) int               { return tianwen.DayOfYear(year, month, day) }
func SearchCities(q string) []tianwen.City             { return tianwen.SearchCities(q) }
