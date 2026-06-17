package tianwen

import (
	"time"

	"liki/internal/engine/ganzhi"
)

// LunarTime is a Chinese lunar calendar date with shichen.
type LunarTime struct {
	Year, Month, Day int
	Leap             bool
	Shichen          ganzhi.Zhi
}

// lunarInfo packs one lunar year (1900-2100) into a uint32:
//
//	bits 0-11:  12 bits — 1=30 days, 0=29 days for months 1-12
//	bits 12-15: leap month index (0=no leap, 1-12=which month doubled)
//	bit  16:    leap month days (0=29, 1=30) — only meaningful if leap>0
//
// Base: lunar 1900-01-01 = solar 1900-01-31.
var lunarInfo = [201]uint32{
	// 1900
	0x04bd8, // 1901
	0x04ae0, 0x0a570, 0x054d5, 0x0d260, 0x0d950, // →1906
	0x16554, 0x056a0, 0x09ad0, 0x055d2, // →1910
	0x04ae0, 0x0a5b6, 0x0a4d0, 0x0d250, 0x1d255, // →1915
	0x0b540, 0x0d6a0, 0x0ada2, 0x095b0, 0x14977, // →1920
	0x04970, 0x0a4b0, 0x0b4b5, 0x06a50, 0x06d40, // →1925
	0x1ab54, 0x02b60, 0x09570, 0x052f2, 0x04970, // →1930
	0x06566, 0x0d4a0, 0x0ea50, 0x06e95, 0x05ad0, // →1935
	0x02b60, 0x186e3, 0x092e0, 0x1c8d7, 0x0c950, // →1940
	0x0d4a0, 0x1d8a6, 0x0b550, 0x056a0, 0x1a5b4, // →1945
	0x025d0, 0x092d0, 0x0d2b2, 0x0a950, 0x0b557, // →1950
	0x06ca0, 0x0b550, 0x15355, 0x04da0, 0x0a5b0, // →1955
	0x14573, 0x052b0, 0x0a9a8, 0x0e950, 0x06aa0, // →1960
	0x0aea6, 0x0ab50, 0x04b60, 0x0aae4, 0x0a570, // →1965
	0x05260, 0x0f263, 0x0d950, 0x05b57, 0x056a0, // →1970
	0x096d0, 0x04dd5, 0x04ad0, 0x0a4d0, 0x0d4d4, // →1975
	0x0d250, 0x0d558, 0x0b540, 0x0b6a0, 0x195a6, // →1980
	0x095b0, 0x049b0, 0x0a974, 0x0a4b0, 0x0b27a, // →1985
	0x06a50, 0x06d40, 0x0af46, 0x0ab60, 0x09570, // →1990
	0x04af5, 0x04970, 0x064b0, 0x074a3, 0x0ea50, // →1995
	0x06b58, 0x05ac0, 0x0ab60, 0x096d5, 0x092e0, // →2000
	0x0c960, 0x0d954, 0x0d4a0, 0x0da50, 0x07552, // →2005
	0x056a0, 0x0abb7, 0x025d0, 0x092d0, 0x0cab5, // →2010
	0x0a950, 0x0b4a0, 0x0baa4, 0x0ad50, 0x055d9, // →2015
	0x04ba0, 0x0a5b0, 0x15176, 0x052b0, 0x0a930, // →2020
	0x07954, 0x06aa0, 0x0ad50, 0x05b52, 0x04b60, // →2025
	0x0a6e6, 0x0a4e0, 0x0d260, 0x0ea65, 0x0d530, // →2030
	0x05aa0, 0x076a3, 0x096d0, 0x04afb, 0x04ad0, // →2035
	0x0a4d0, 0x1d0b6, 0x0d250, 0x0d520, 0x0dd45, // →2040
	0x0b5a0, 0x056d0, 0x055b2, 0x049b0, 0x0a577, // →2045
	0x0a4b0, 0x0aa50, 0x1b255, 0x06d20, 0x0ada0, // →2050
	0x14b63, 0x09370, 0x049f8, 0x04970, 0x064b0, // →2055
	0x168a6, 0x0ea50, 0x06b20, 0x1a6c4, 0x0aae0, // →2060
	0x0a2e0, 0x0d2e3, 0x0c960, 0x0d557, 0x0d4a0, // →2065
	0x0da50, 0x05d55, 0x056a0, 0x0a6d0, 0x055d4, // →2070
	0x052d0, 0x0a9b8, 0x0a950, 0x0b4a0, 0x0b6a6, // →2075
	0x0ad50, 0x055a0, 0x0aba4, 0x0a5b0, 0x052b0, // →2080
	0x0b273, 0x06930, 0x07337, 0x06aa0, 0x0ad50, // →2085
	0x14b55, 0x04b60, 0x0a570, 0x054e4, 0x0d160, // →2090
	0x0e968, 0x0d520, 0x0daa0, 0x16aa6, 0x056d0, // →2095
	0x04ae0, 0x0a9d4, 0x0a4d0, 0x0d150, 0x0f252, // →2100
	0x0d520, // 2101
}

// SolarToLunar converts a Gregorian date to Chinese lunar date.
func SolarToLunar(solarYear, solarMonth, solarDay int) LunarTime {
	// Days from 1900-01-31
	base := time.Date(1900, 1, 31, 0, 0, 0, 0, time.UTC)
	target := time.Date(solarYear, time.Month(solarMonth), solarDay, 0, 0, 0, 0, time.UTC)
	offset := int(target.Sub(base).Hours() / 24)
	if offset < 0 {
		return LunarTime{}
	}

	// Walk years
	ly := 1900
	for ly <= 2100 {
		yd := lunarYearDays(ly - 1900)
		if offset < yd {
			break
		}
		offset -= yd
		ly++
	}
	if ly > 2100 {
		return LunarTime{}
	}

	// Walk months
	info := lunarInfo[ly-1900]
	leap := int((info >> 12) & 0xf) // 0=no leap, 1-12=which month
	lm := 1
	isLeap := false
	for lm <= 12 {
		md := lunarMonthDays(info, lm, false)
		if offset < md {
			break
		}
		offset -= md
		if leap > 0 && lm == leap {
			lmd := lunarMonthDays(info, lm, true)
			if offset < lmd {
				isLeap = true
				break
			}
			offset -= lmd
		}
		lm++
	}

	return LunarTime{Year: ly, Month: lm, Day: offset + 1, Leap: isLeap}
}

func lunarYearDays(idx int) int {
	if idx < 0 || idx >= len(lunarInfo) {
		return 0
	}
	info := lunarInfo[idx]
	sum := 0
	for m := 1; m <= 12; m++ {
		sum += lunarMonthDays(info, m, false)
	}
	leap := int((info >> 12) & 0xf)
	if leap > 0 {
		sum += lunarMonthDays(info, leap, true)
	}
	return sum
}

func lunarMonthDays(info uint32, month int, leap bool) int {
	if leap {
		if (info>>16)&1 == 1 {
			return 30
		}
		return 29
	}
	if (info>>uint(month-1))&1 == 1 {
		return 30
	}
	return 29
}
