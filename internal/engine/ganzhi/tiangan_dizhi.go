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

// Gan is a heavenly stem (天干). 1=甲 .. 10=癸.
type Gan int

const (
	GanJia  Gan = 1  // 甲
	GanYi   Gan = 2  // 乙
	GanBing Gan = 3  // 丙
	GanDing Gan = 4  // 丁
	GanWu   Gan = 5  // 戊
	GanJi   Gan = 6  // 己
	GanGeng Gan = 7  // 庚
	GanXin  Gan = 8  // 辛
	GanRen  Gan = 9  // 壬
	GanGui  Gan = 10 // 癸
)

var ganWuxingBiao = [11]int{0, 1, 1, 2, 2, 3, 3, 4, 4, 5, 5} // 1=木..5=水
var ganYinYangBiao = [11]bool{false, true, false, true, false, true, false, true, false, true, false}

// Zhi is an earthly branch (地支). 1=子 .. 12=亥.
type Zhi int

const (
	ZhiZi   Zhi = 1  // 子
	ZhiChou Zhi = 2  // 丑
	ZhiYin  Zhi = 3  // 寅
	ZhiMao  Zhi = 4  // 卯
	ZhiChen Zhi = 5  // 辰
	ZhiSi   Zhi = 6  // 巳
	ZhiWu   Zhi = 7  // 午
	ZhiWei  Zhi = 8  // 未
	ZhiShen Zhi = 9  // 申
	ZhiYou  Zhi = 10 // 酉
	ZhiXu   Zhi = 11 // 戌
	ZhiHai  Zhi = 12 // 亥
)

var zhiWuxingBiao = [13]int{0, 5, 3, 1, 1, 3, 2, 2, 3, 4, 4, 3, 5}

// Wuxing is the five-phase element (五行). 1=木 2=火 3=土 4=金 5=水.
type Wuxing int

const (
	WxMu   Wuxing = 1
	WxHuo  Wuxing = 2
	WxTu   Wuxing = 3
	WxJin  Wuxing = 4
	WxShui Wuxing = 5
)

// YinYang distinguishes yin from yang.
type YinYang bool

const (
	Yin  YinYang = false
	Yang YinYang = true
)

// Zhu is one heavenly-stem / earthly-branch pair (一柱).
type Zhu struct {
	Gan Gan `json:"gan"`
	Zhi Zhi `json:"zhi"`
}

// Bazi holds the four named pillars of a BaZi chart (八字).
type Bazi struct {
	Nian Zhu `json:"nian"`
	Yue  Zhu `json:"yue"`
	Ri   Zhu `json:"ri"`
	Shi  Zhu `json:"shi"`
}

// Slice returns the four pillars as an indexable array.
func (bz Bazi) Slice() [4]Zhu {
	return [4]Zhu{bz.Nian, bz.Yue, bz.Ri, bz.Shi}
}

// Validate checks that all four pillars have valid gan (1-10) and zhi (1-12) values.
func (bz Bazi) Validate() error {
	for i, p := range bz.Slice() {
		if p.Gan < 1 || p.Gan > 10 {
			return fmt.Errorf("bazi[%d].gan must be 1-10", i)
		}
		if p.Zhi < 1 || p.Zhi > 12 {
			return fmt.Errorf("bazi[%d].zhi must be 1-12", i)
		}
	}
	return nil
}

func (g Gan) MarshalJSON() ([]byte, error)    { return json.Marshal(GanName(g)) }
func (z Zhi) MarshalJSON() ([]byte, error)    { return json.Marshal(ZhiName(z)) }
func (g Gan) String() string                  { return GanName(g) }
func (z Zhi) String() string                  { return ZhiName(z) }
func (w Wuxing) MarshalJSON() ([]byte, error) { return json.Marshal(w.String()) }

func (w *Wuxing) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return err
	}
	for i := WxMu; i <= WxShui; i++ {
		if i.String() == name {
			*w = i
			return nil
		}
	}
	return fmt.Errorf("unknown wuxing: %s", name)
}

func (g *Gan) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err == nil {
		for i := 1; i <= 10; i++ {
			if GanNames[i] == name {
				*g = Gan(i)
				return nil
			}
		}
		return &json.UnmarshalTypeError{Value: "string", Type: nil, Field: name}
	}
	var i int
	if err := json.Unmarshal(data, &i); err == nil {
		*g = Gan(i)
		return nil
	}
	return &json.UnmarshalTypeError{Value: string(data), Type: nil}
}

func (z *Zhi) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err == nil {
		for i := 1; i <= 12; i++ {
			if ZhiNames[i] == name {
				*z = Zhi(i)
				return nil
			}
		}
		return &json.UnmarshalTypeError{Value: "string", Type: nil, Field: name}
	}
	var i int
	if err := json.Unmarshal(data, &i); err == nil {
		*z = Zhi(i)
		return nil
	}
	return &json.UnmarshalTypeError{Value: string(data), Type: nil}
}

func (z Zhu) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Gan string `json:"gan"`
		Zhi string `json:"zhi"`
	}{Gan: GanName(z.Gan), Zhi: ZhiName(z.Zhi)})
}

// -- primitive lookups --

func GanWuxing(g Gan) Wuxing   { return Wuxing(ganWuxingBiao[g]) }
func GanYinYang(g Gan) YinYang { return YinYang(ganYinYangBiao[g]) }
func ZhiWuxing(z Zhi) Wuxing   { return Wuxing(zhiWuxingBiao[z]) }

// SixtyCycleName returns the 0-based index in [0,59] for a given gan+zhi.
func SixtyCycleName(gan Gan, zhi Zhi) int {
	idx := (int(gan)*6 - int(zhi)*5 - 1) % 60
	if idx < 0 {
		idx += 60
	}
	return idx
}

var zodiacNames = [13]string{
	"", "鼠", "牛", "虎", "兔", "龙", "蛇",
	"马", "羊", "猴", "鸡", "狗", "猪",
}

// zodiac returns the Chinese zodiac animal name for a given zhi.
func zodiac(z Zhi) (string, bool) {
	if int(z) >= 1 && int(z) <= 12 {
		return zodiacNames[z], true
	}
	return "", false
}

func zhiLabel(s string, ok bool) string {
	if !ok {
		return "未知"
	}
	return s
}

// ZodiacLabel returns the Chinese zodiac animal name, or "未知" if lookup fails.
func ZodiacLabel(z Zhi) string { s, ok := zodiac(z); return zhiLabel(s, ok) }

// zhiSeason returns the season ("春"/"夏"/"秋"/"冬") for a zhi.
func zhiSeason(z Zhi) (string, bool) {
	switch int(z) {
	case 3, 4, 5:
		return "春", true
	case 6, 7, 8:
		return "夏", true
	case 9, 10, 11:
		return "秋", true
	case 12, 1, 2:
		return "冬", true
	}
	return "", false
}

// ZhiSeasonLabel returns the season label, or "未知" if lookup fails.
func ZhiSeasonLabel(z Zhi) string { s, ok := zhiSeason(z); return zhiLabel(s, ok) }

// zhiLunarMonth returns the lunar month label (e.g. "正月") for a zhi.
func zhiLunarMonth(z Zhi) (string, bool) {
	lunarMonths := [13]string{
		"", "十一月", "十二月", "正月", "二月", "三月", "四月",
		"五月", "六月", "七月", "八月", "九月", "十月",
	}
	if int(z) >= 1 && int(z) <= 12 {
		return lunarMonths[z], true
	}
	return "", false
}

// zhiLunarMonthLabel returns the lunar month label, or "未知" if lookup fails.
func zhiLunarMonthLabel(z Zhi) string { s, ok := zhiLunarMonth(z); return zhiLabel(s, ok) }

var HourRanges = [12]string{
	"23:00-01:00", "01:00-03:00", "03:00-05:00", "05:00-07:00",
	"07:00-09:00", "09:00-11:00", "11:00-13:00", "13:00-15:00",
	"15:00-17:00", "17:00-19:00", "19:00-21:00", "21:00-23:00",
}

// zhiHourRange returns the two-hour range label (e.g. "23:00-01:00") for a zhi.
func zhiHourRange(z Zhi) (string, bool) {
	if int(z) >= 1 && int(z) <= 12 {
		return HourRanges[z-1], true
	}
	return "", false
}

// ZhiHourRangeLabel returns the two-hour range label, or "未知" if lookup fails.
func ZhiHourRangeLabel(z Zhi) string { s, ok := zhiHourRange(z); return zhiLabel(s, ok) }

var GanNames = [11]string{"", "甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}
var ZhiNames = [13]string{"", "子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}

// GanName returns the Chinese character for a heavenly stem (e.g. 甲).
func GanName(g Gan) string { return GanNames[g] }

// ZhiName returns the Chinese character for an earthly branch (e.g. 子).
func ZhiName(z Zhi) string { return ZhiNames[z] }

// Sheng returns true if wuxing `from` nourishes wuxing `to` in the five-phase cycle.
func Sheng(from, to Wuxing) bool {
	return (from == WxMu && to == WxHuo) ||
		(from == WxHuo && to == WxTu) ||
		(from == WxTu && to == WxJin) ||
		(from == WxJin && to == WxShui) ||
		(from == WxShui && to == WxMu)
}

func (w Wuxing) String() string {
	switch w {
	case WxMu:
		return "木"
	case WxHuo:
		return "火"
	case WxTu:
		return "土"
	case WxJin:
		return "金"
	case WxShui:
		return "水"
	}
	return "未知"
}

// Ke returns true if wuxing `from` controls wuxing `to` in the five-phase cycle.
func Ke(from, to Wuxing) bool {
	return (from == WxMu && to == WxTu) ||
		(from == WxTu && to == WxShui) ||
		(from == WxShui && to == WxHuo) ||
		(from == WxHuo && to == WxJin) ||
		(from == WxJin && to == WxMu)
}

// WuxingFromChinese converts a Chinese wuxing name to its Wuxing value.
func WuxingFromChinese(s string) Wuxing {
	switch s {
	case "木":
		return WxMu
	case "火":
		return WxHuo
	case "土":
		return WxTu
	case "金":
		return WxJin
	case "水":
		return WxShui
	}
	return 0
}

// WuxingFromString converts a Chinese or English wuxing name to its Wuxing value.
func WuxingFromString(s string) Wuxing {
	switch s {
	case "wood", "木":
		return WxMu
	case "fire", "火":
		return WxHuo
	case "earth", "土":
		return WxTu
	case "metal", "金":
		return WxJin
	case "water", "水":
		return WxShui
	}
	return 0
}

// SixtyToZhu converts a 0-based sixty-cycle index to a Zhu.
func SixtyToZhu(idx int) Zhu {
  g := Gan((idx%10)+1)
  z := Zhi((idx%12)+1)
  return Zhu{Gan: g, Zhi: z}
}
