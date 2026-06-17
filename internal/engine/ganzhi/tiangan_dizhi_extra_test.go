package ganzhi

import "testing"

// -- ZhiSeasonLabel --

func TestZhiSeasonLabel_All(t *testing.T) {
	tests := []struct {
		z    Zhi
		want string
	}{
		{ZhiYin, "春"}, {ZhiMao, "春"}, {ZhiChen, "春"},
		{ZhiSi, "夏"}, {ZhiWu, "夏"}, {ZhiWei, "夏"},
		{ZhiShen, "秋"}, {ZhiYou, "秋"}, {ZhiXu, "秋"},
		{ZhiHai, "冬"}, {ZhiZi, "冬"}, {ZhiChou, "冬"},
	}
	for _, tc := range tests {
		got := ZhiSeasonLabel(tc.z)
		if got != tc.want {
			t.Errorf("ZhiSeasonLabel(%d)=%s, want %s", tc.z, got, tc.want)
		}
	}
}

func TestZhiSeasonLabel_Invalid(t *testing.T) {
	if got := ZhiSeasonLabel(0); got != "未知" {
		t.Errorf("ZhiSeasonLabel(0)=%s, want 未知", got)
	}
	if got := ZhiSeasonLabel(13); got != "未知" {
		t.Errorf("ZhiSeasonLabel(13)=%s, want 未知", got)
	}
}

// -- zhiLunarMonth complete --

func TestZhiLunarMonth_All(t *testing.T) {
	tests := []struct {
		z    Zhi
		want string
	}{
		{ZhiZi, "十一月"},   // 1
		{ZhiChou, "十二月"}, // 2
		{ZhiYin, "正月"},    // 3
		{ZhiMao, "二月"},    // 4
		{ZhiChen, "三月"},   // 5
		{ZhiSi, "四月"},     // 6
		{ZhiWu, "五月"},     // 7
		{ZhiWei, "六月"},    // 8
		{ZhiShen, "七月"},   // 9
		{ZhiYou, "八月"},    // 10
		{ZhiXu, "九月"},     // 11
		{ZhiHai, "十月"},    // 12
	}
	for _, tc := range tests {
		got, ok := zhiLunarMonth(tc.z)
		if !ok {
			t.Errorf("zhiLunarMonth(%d) not ok", tc.z)
		}
		if got != tc.want {
			t.Errorf("zhiLunarMonth(%d)=%s, want %s", tc.z, got, tc.want)
		}
	}
}

func TestZhiLunarMonth_Invalid(t *testing.T) {
	for _, z := range []Zhi{0, 13} {
		if _, ok := zhiLunarMonth(z); ok {
			t.Errorf("zhiLunarMonth(%d) should not be ok", z)
		}
	}
}

// -- zhiLunarMonthLabel --

func TestZhiLunarMonthLabel_Valid(t *testing.T) {
	if got := zhiLunarMonthLabel(ZhiYin); got != "正月" {
		t.Errorf("zhiLunarMonthLabel(寅)=%s, want 正月", got)
	}
}

func TestZhiLunarMonthLabel_Invalid(t *testing.T) {
	if got := zhiLunarMonthLabel(0); got != "未知" {
		t.Errorf("zhiLunarMonthLabel(0)=%s, want 未知", got)
	}
}

// -- zhiHourRange complete --

func TestZhiHourRange_All(t *testing.T) {
	tests := []struct {
		z    Zhi
		want string
	}{
		{ZhiZi, "23:00-01:00"},
		{ZhiChou, "01:00-03:00"},
		{ZhiYin, "03:00-05:00"},
		{ZhiMao, "05:00-07:00"},
		{ZhiChen, "07:00-09:00"},
		{ZhiSi, "09:00-11:00"},
		{ZhiWu, "11:00-13:00"},
		{ZhiWei, "13:00-15:00"},
		{ZhiShen, "15:00-17:00"},
		{ZhiYou, "17:00-19:00"},
		{ZhiXu, "19:00-21:00"},
		{ZhiHai, "21:00-23:00"},
	}
	for _, tc := range tests {
		got, ok := zhiHourRange(tc.z)
		if !ok {
			t.Errorf("zhiHourRange(%d) not ok", tc.z)
		}
		if got != tc.want {
			t.Errorf("zhiHourRange(%d)=%s, want %s", tc.z, got, tc.want)
		}
	}
}

func TestZhiHourRange_Invalid(t *testing.T) {
	for _, z := range []Zhi{0, 13} {
		if _, ok := zhiHourRange(z); ok {
			t.Errorf("zhiHourRange(%d) should not be ok", z)
		}
	}
}

// -- ZhiHourRangeLabel --

func TestZhiHourRangeLabel_Valid(t *testing.T) {
	if got := ZhiHourRangeLabel(ZhiWu); got != "11:00-13:00" {
		t.Errorf("ZhiHourRangeLabel(午)=%s, want 11:00-13:00", got)
	}
}

func TestZhiHourRangeLabel_Invalid(t *testing.T) {
	if got := ZhiHourRangeLabel(0); got != "未知" {
		t.Errorf("ZhiHourRangeLabel(0)=%s, want 未知", got)
	}
}

// -- GanName / ZhiName —
func TestGanName_Invalid(t *testing.T) {
	if got := GanName(0); got != "" {
		t.Errorf("GanName(0)=%s, want empty", got)
	}
}
func TestZhiName_Invalid(t *testing.T) {
	if got := ZhiName(0); got != "" {
		t.Errorf("ZhiName(0)=%s, want empty", got)
	}
}
