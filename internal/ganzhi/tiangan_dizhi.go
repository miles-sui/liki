package ganzhi

import (
	"encoding/json"
	"fmt"
)

// Gender represents male or female, used in BaZi and FengShui calculations.
type Gender string

const (
	Male   Gender = "male"
	Female Gender = "female"
)

// Stem is a heavenly stem (天干). 1=甲 .. 10=癸.
type Stem int

const (
	StemJia  Stem = 1  // 甲
	StemYi   Stem = 2  // 乙
	StemBing Stem = 3  // 丙
	StemDing Stem = 4  // 丁
	StemWu   Stem = 5  // 戊
	StemJi   Stem = 6  // 己
	StemGeng Stem = 7  // 庚
	StemXin  Stem = 8  // 辛
	StemRen  Stem = 9  // 壬
	StemGui  Stem = 10 // 癸
)

var StemElements = [11]int{0, 1, 1, 2, 2, 3, 3, 4, 4, 5, 5}     // 1=木..5=水
var StemYinYangTable = [11]bool{false, true, false, true, false, true, false, true, false, true, false}

// Branch is an earthly branch (地支). 1=子 .. 12=亥.
type Branch int

const (
	BranchZi   Branch = 1  // 子
	BranchChou Branch = 2  // 丑
	BranchYin  Branch = 3  // 寅
	BranchMao  Branch = 4  // 卯
	BranchChen Branch = 5  // 辰
	BranchSi   Branch = 6  // 巳
	BranchWu   Branch = 7  // 午
	BranchWei  Branch = 8  // 未
	BranchShen Branch = 9  // 申
	BranchYou  Branch = 10 // 酉
	BranchXu   Branch = 11 // 戌
	BranchHai  Branch = 12 // 亥
)

var BranchElements = [13]int{0, 5, 3, 1, 1, 3, 2, 2, 3, 4, 4, 3, 5}

// Element is the five-phase element (五行). 1=木 2=火 3=土 4=金 5=水.
type Element int

const (
	ElemWood  Element = 1
	ElemFire  Element = 2
	ElemEarth Element = 3
	ElemMetal Element = 4
	ElemWater Element = 5
)

// YinYang distinguishes yin from yang.
type YinYang bool

const (
	Yin  YinYang = false
	Yang YinYang = true
)

// Pillar is one heavenly-stem / earthly-branch pair (一柱).
type Pillar struct {
	Stem   Stem   `json:"stem"`
	Branch Branch `json:"branch"`
}

// Bazi holds the four named pillars of a BaZi chart (八字).
type Bazi struct {
	Year  Pillar `json:"year"`
	Month Pillar `json:"month"`
	Day   Pillar `json:"day"`
	Hour  Pillar `json:"hour"`
}

// Slice returns the four pillars as an indexable array.
func (bz Bazi) Slice() [4]Pillar {
	return [4]Pillar{bz.Year, bz.Month, bz.Day, bz.Hour}
}

// Validate checks that all four pillars have valid stem (1-10) and branch (1-12) values.
func (bz Bazi) Validate() error {
	for i, p := range bz.Slice() {
		if p.Stem < 1 || p.Stem > 10 {
			return fmt.Errorf("bazi[%d].stem must be 1-10", i)
		}
		if p.Branch < 1 || p.Branch > 12 {
			return fmt.Errorf("bazi[%d].branch must be 1-12", i)
		}
	}
	return nil
}

func (s Stem) MarshalJSON() ([]byte, error)     { return json.Marshal(StemName(s)) }
func (b Branch) MarshalJSON() ([]byte, error)  { return json.Marshal(BranchName(b)) }
func (e Element) MarshalJSON() ([]byte, error) { return json.Marshal(e.String()) }

func (e *Element) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return err
	}
	for i := ElemWood; i <= ElemWater; i++ {
		if i.String() == name {
			*e = i
			return nil
		}
	}
	return fmt.Errorf("unknown element: %s", name)
}

func (s *Stem) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err == nil {
		for i := 1; i <= 10; i++ {
			if StemNames[i] == name {
				*s = Stem(i)
				return nil
			}
		}
		return &json.UnmarshalTypeError{Value: "string", Type: nil, Field: name}
	}
	var i int
	if err := json.Unmarshal(data, &i); err == nil {
		*s = Stem(i)
		return nil
	}
	return &json.UnmarshalTypeError{Value: string(data), Type: nil}
}

func (b *Branch) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err == nil {
		for i := 1; i <= 12; i++ {
			if BranchNames[i] == name {
				*b = Branch(i)
				return nil
			}
		}
		return &json.UnmarshalTypeError{Value: "string", Type: nil, Field: name}
	}
	var i int
	if err := json.Unmarshal(data, &i); err == nil {
		*b = Branch(i)
		return nil
	}
	return &json.UnmarshalTypeError{Value: string(data), Type: nil}
}

func (p Pillar) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Stem   string `json:"stem"`
		Branch string `json:"branch"`
	}{Stem: StemName(p.Stem), Branch: BranchName(p.Branch)})
}

// -- primitive lookups --

func StemElement(s Stem) Element     { return Element(StemElements[s]) }
func StemYinYang(s Stem) YinYang     { return YinYang(StemYinYangTable[s]) }
func BranchElement(b Branch) Element { return Element(BranchElements[b]) }

// SixtyCycleName returns the 0-based index in [0,59] for a given stem+branch.
func SixtyCycleName(stem Stem, branch Branch) int {
	idx := (int(stem)*6 - int(branch)*5 - 1) % 60
	if idx < 0 {
		idx += 60
	}
	return idx
}

var ZodiacNames = [13]string{
	"", "鼠", "牛", "虎", "兔", "龙", "蛇",
	"马", "羊", "猴", "鸡", "狗", "猪",
}

// Zodiac returns the Chinese zodiac animal name for a given branch.
func Zodiac(b Branch) string {
	if int(b) >= 1 && int(b) <= 12 {
		return ZodiacNames[b]
	}
	return ""
}

// BranchSeason returns the season ("春"/"夏"/"秋"/"冬") for a branch.
func BranchSeason(b Branch) string {
	switch int(b) {
	case 3, 4, 5:
		return "春"
	case 6, 7, 8:
		return "夏"
	case 9, 10, 11:
		return "秋"
	case 12, 1, 2:
		return "冬"
	}
	return ""
}

// BranchLunarMonth returns the lunar month label (e.g. "正月") for a branch.
func BranchLunarMonth(b Branch) string {
	lunarMonths := [13]string{
		"", "十一月", "十二月", "正月", "二月", "三月", "四月",
		"五月", "六月", "七月", "八月", "九月", "十月",
	}
	if int(b) >= 1 && int(b) <= 12 {
		return lunarMonths[b]
	}
	return ""
}

var HourRanges = [12]string{
	"23:00-01:00", "01:00-03:00", "03:00-05:00", "05:00-07:00",
	"07:00-09:00", "09:00-11:00", "11:00-13:00", "13:00-15:00",
	"15:00-17:00", "17:00-19:00", "19:00-21:00", "21:00-23:00",
}

// BranchHourRange returns the two-hour range label (e.g. "23:00-01:00") for a branch.
func BranchHourRange(b Branch) string {
	if int(b) >= 1 && int(b) <= 12 {
		return HourRanges[b-1]
	}
	return ""
}

var StemNames = [11]string{"", "甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}
var BranchNames = [13]string{"", "子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}

// StemNameStr returns the Chinese character for a heavenly stem.
func StemNameStr(s Stem) string { return StemNames[s] }

// BranchNameStr returns the Chinese character for an earthly branch.
func BranchNameStr(b Branch) string { return BranchNames[b] }

// StemName returns the Chinese character for a heavenly stem (e.g. 甲).
func StemName(s Stem) string { return StemNameStr(s) }

// BranchName returns the Chinese character for an earthly branch (e.g. 子).
func BranchName(b Branch) string { return BranchNameStr(b) }

// Sheng returns true if element `from` nourishes element `to` in the five-phase cycle.
func Sheng(from, to Element) bool {
	return (from == ElemWood && to == ElemFire) ||
		(from == ElemFire && to == ElemEarth) ||
		(from == ElemEarth && to == ElemMetal) ||
		(from == ElemMetal && to == ElemWater) ||
		(from == ElemWater && to == ElemWood)
}

func (e Element) String() string {
	switch e {
	case ElemWood:
		return "木"
	case ElemFire:
		return "火"
	case ElemEarth:
		return "土"
	case ElemMetal:
		return "金"
	case ElemWater:
		return "水"
	}
	return "未知"
}

// Ke returns true if element `from` controls element `to` in the five-phase cycle.
func Ke(from, to Element) bool {
	return (from == ElemWood && to == ElemEarth) ||
		(from == ElemEarth && to == ElemWater) ||
		(from == ElemWater && to == ElemFire) ||
		(from == ElemFire && to == ElemMetal) ||
		(from == ElemMetal && to == ElemWood)
}

// ElementFromChinese converts a Chinese element name to its Element value.
func ElementFromChinese(s string) Element {
	switch s {
	case "木":
		return ElemWood
	case "火":
		return ElemFire
	case "土":
		return ElemEarth
	case "金":
		return ElemMetal
	case "水":
		return ElemWater
	}
	return 0
}
