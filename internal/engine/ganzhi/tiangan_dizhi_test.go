package ganzhi

import (
	"encoding/json"
	"testing"
)

// -- SixtyCycleName (most critical — all pillar indexing depends on this) --

func TestSixtyCycleName_Known(t *testing.T) {
	tests := []struct {
		gan  Gan
		zhi  Zhi
		want int
	}{
		{GanJia, ZhiZi, 0},   // 甲子
		{GanYi, ZhiChou, 1},  // 乙丑
		{GanGui, ZhiHai, 59}, // 癸亥
		{GanJia, ZhiXu, 10},  // 甲戌
		{GanBing, ZhiYin, 2}, // 丙寅
		{GanJia, ZhiZi, 0},   // 甲子 (same as first, idempotent)
	}
	for _, tc := range tests {
		got := SixtyCycleName(tc.gan, tc.zhi)
		if got != tc.want {
			t.Errorf("SixtyCycleName(%s,%s)=%d, want %d", tc.gan, tc.zhi, got, tc.want)
		}
	}
}

func TestSixtyCycleName_Range(t *testing.T) {
	for g := GanJia; g <= GanGui; g++ {
		for z := ZhiZi; z <= ZhiHai; z++ {
			idx := SixtyCycleName(g, z)
			if idx < 0 || idx > 59 {
				t.Errorf("SixtyCycleName(%s,%s)=%d out of [0,59]", g, z, idx)
			}
		}
	}
}

func TestSixtyCycleName_Bijection(t *testing.T) {
	// The 60 JiaZi cycle only includes pairs where gan and zhi have the same
	// yin-yang parity (both odd or both even). All 60 valid pairs must produce
	// distinct indices with no collisions.
	seen := make(map[int]bool)
	count := 0
	for g := GanJia; g <= GanGui; g++ {
		for z := ZhiZi; z <= ZhiHai; z++ {
			// Only valid 60-cycle pairs: same parity (both yang or both yin).
			if int(g)%2 != int(z)%2 {
				continue
			}
			count++
			idx := SixtyCycleName(g, z)
			if seen[idx] {
				t.Errorf("SixtyCycleName collision at index %d (gan=%s, zhi=%s)", idx, g, z)
			}
			seen[idx] = true
		}
	}
	if len(seen) != 60 {
		t.Errorf("expected 60 unique indices, got %d", len(seen))
	}
	if count != 60 {
		t.Errorf("expected 60 valid pairs, counted %d", count)
	}
}

func TestSixtyCycleName_Consecutive(t *testing.T) {
	// In the 60-cycle, consecutive pairs advance both gan and zhi by 1.
	// Starting from 甲子(0), next is 乙丑(1), then 丙寅(2), etc.
	ganVals := []Gan{GanJia, GanYi, GanBing, GanDing, GanWu, GanJi, GanGeng, GanXin, GanRen, GanGui,
		GanJia, GanYi, GanBing, GanDing, GanWu, GanJi, GanGeng, GanXin, GanRen, GanGui,
		GanJia, GanYi, GanBing, GanDing, GanWu, GanJi, GanGeng, GanXin, GanRen, GanGui,
		GanJia, GanYi, GanBing, GanDing, GanWu, GanJi, GanGeng, GanXin, GanRen, GanGui,
		GanJia, GanYi, GanBing, GanDing, GanWu, GanJi, GanGeng, GanXin, GanRen, GanGui,
		GanJia, GanYi, GanBing, GanDing, GanWu, GanJi, GanGeng, GanXin, GanRen, GanGui}
	zhiVals := []Zhi{ZhiZi, ZhiChou, ZhiYin, ZhiMao, ZhiChen, ZhiSi, ZhiWu, ZhiWei, ZhiShen, ZhiYou, ZhiXu, ZhiHai,
		ZhiZi, ZhiChou, ZhiYin, ZhiMao, ZhiChen, ZhiSi, ZhiWu, ZhiWei, ZhiShen, ZhiYou, ZhiXu, ZhiHai,
		ZhiZi, ZhiChou, ZhiYin, ZhiMao, ZhiChen, ZhiSi, ZhiWu, ZhiWei, ZhiShen, ZhiYou, ZhiXu, ZhiHai,
		ZhiZi, ZhiChou, ZhiYin, ZhiMao, ZhiChen, ZhiSi, ZhiWu, ZhiWei, ZhiShen, ZhiYou, ZhiXu, ZhiHai,
		ZhiZi, ZhiChou, ZhiYin, ZhiMao, ZhiChen, ZhiSi, ZhiWu, ZhiWei, ZhiShen, ZhiYou, ZhiXu, ZhiHai}

	for i := 0; i < 59; i++ {
		idx1 := SixtyCycleName(ganVals[i], zhiVals[i])
		idx2 := SixtyCycleName(ganVals[i+1], zhiVals[i+1])
		if idx2 != (idx1+1)%60 {
			t.Errorf("%s%s=%d → %s%s=%d, want +1 mod 60",
				ganVals[i], zhiVals[i], idx1, ganVals[i+1], zhiVals[i+1], idx2)
		}
	}
}

// -- sheng / ke --

func TestSheng_Known(t *testing.T) {
	pairs := []struct{ from, to Wuxing }{
		{WxMu, WxHuo},
		{WxHuo, WxTu},
		{WxTu, WxJin},
		{WxJin, WxShui},
		{WxShui, WxMu},
	}
	for _, p := range pairs {
		if !Sheng(p.from, p.to) {
			t.Errorf("Sheng(%s,%s)=false, want true", p.from, p.to)
		}
	}
}

func TestSheng_NonPairs(t *testing.T) {
	nonPairs := []struct{ from, to Wuxing }{
		{WxMu, WxTu}, {WxMu, WxJin}, {WxMu, WxShui}, {WxMu, WxMu},
		{WxHuo, WxMu}, {WxHuo, WxShui}, {WxHuo, WxHuo},
	}
	for _, p := range nonPairs {
		if Sheng(p.from, p.to) {
			t.Errorf("Sheng(%s,%s)=true, want false", p.from, p.to)
		}
	}
}

func TestKe_Known(t *testing.T) {
	pairs := []struct{ from, to Wuxing }{
		{WxMu, WxTu},
		{WxTu, WxShui},
		{WxShui, WxHuo},
		{WxHuo, WxJin},
		{WxJin, WxMu},
	}
	for _, p := range pairs {
		if !Ke(p.from, p.to) {
			t.Errorf("Ke(%s,%s)=false, want true", p.from, p.to)
		}
	}
}

func TestKe_NonPairs(t *testing.T) {
	// sheng pairs and self-same are not ke.
	nonPairs := []struct{ from, to Wuxing }{
		{WxMu, WxHuo},   // sheng
		{WxHuo, WxTu},   // sheng
		{WxTu, WxJin},   // sheng
		{WxJin, WxShui}, // sheng
		{WxShui, WxMu},  // sheng
		{WxMu, WxMu},    // self
		{WxHuo, WxHuo},  // self
		{WxMu, WxShui},  // reverse sheng: 水生木, not ke
		{WxHuo, WxMu},   // reverse sheng: 木生火, not ke
	}
	for _, p := range nonPairs {
		if Ke(p.from, p.to) {
			t.Errorf("Ke(%s,%s)=true, want false", p.from, p.to)
		}
	}
}

// -- ganWuxing / zhiWuxing --

func TestGanWuxing_All(t *testing.T) {
	want := map[Gan]Wuxing{
		GanJia: WxMu, GanYi: WxMu,
		GanBing: WxHuo, GanDing: WxHuo,
		GanWu: WxTu, GanJi: WxTu,
		GanGeng: WxJin, GanXin: WxJin,
		GanRen: WxShui, GanGui: WxShui,
	}
	for g, w := range want {
		if got := GanWuxing(g); got != w {
			t.Errorf("GanWuxing(%s)=%s, want %s", g, got, w)
		}
	}
}

func TestZhiWuxing_All(t *testing.T) {
	want := map[Zhi]Wuxing{
		ZhiYin: WxMu, ZhiMao: WxMu,
		ZhiSi: WxHuo, ZhiWu: WxHuo,
		ZhiChen: WxTu, ZhiXu: WxTu, ZhiChou: WxTu, ZhiWei: WxTu,
		ZhiShen: WxJin, ZhiYou: WxJin,
		ZhiHai: WxShui, ZhiZi: WxShui,
	}
	for z, w := range want {
		if got := ZhiWuxing(z); got != w {
			t.Errorf("ZhiWuxing(%s)=%s, want %s", z, got, w)
		}
	}
}

// -- ganYinYang --

func TestGanYinYang_All(t *testing.T) {
	yang := []Gan{GanJia, GanBing, GanWu, GanGeng, GanRen}
	yin := []Gan{GanYi, GanDing, GanJi, GanXin, GanGui}
	for _, g := range yang {
		if got := GanYinYang(g); got != Yang {
			t.Errorf("GanYinYang(%s)=Yin, want Yang", g)
		}
	}
	for _, g := range yin {
		if got := GanYinYang(g); got != Yin {
			t.Errorf("GanYinYang(%s)=Yang, want Yin", g)
		}
	}
}

// -- Wuxing.String / WuxingFromChinese round-trip --

func TestWuxing_String(t *testing.T) {
	want := map[Wuxing]string{WxMu: "木", WxHuo: "火", WxTu: "土", WxJin: "金", WxShui: "水"}
	for w, s := range want {
		if got := w.String(); got != s {
			t.Errorf("Wuxing(%d).String()=%s, want %s", w, got, s)
		}
	}
}

func TestWuxingFromChinese_Roundtrip(t *testing.T) {
	names := []string{"木", "火", "土", "金", "水"}
	for _, s := range names {
		w := WuxingFromChinese(s)
		if w == 0 {
			t.Errorf("WuxingFromChinese(%s)=0", s)
		}
		if w.String() != s {
			t.Errorf("roundtrip: %s → %d → %s", s, w, w.String())
		}
	}
	if got := WuxingFromChinese("X"); got != 0 {
		t.Errorf("WuxingFromChinese(X)=%d, want 0", got)
	}
}

// -- GanName / ZhiName --

func TestGanName_All(t *testing.T) {
	names := []string{"", "甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}
	for i := 1; i <= 10; i++ {
		if got := GanName(Gan(i)); got != names[i] {
			t.Errorf("GanName(%d)=%s, want %s", i, got, names[i])
		}
	}
}

func TestZhiName_All(t *testing.T) {
	names := []string{"", "子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}
	for i := 1; i <= 12; i++ {
		if got := ZhiName(Zhi(i)); got != names[i] {
			t.Errorf("ZhiName(%d)=%s, want %s", i, got, names[i])
		}
	}
}

// -- zodiac / zhiSeason / zhiLunarMonth / zhiHourRange --

func TestZodiac_All(t *testing.T) {
	want := []string{"", "鼠", "牛", "虎", "兔", "龙", "蛇", "马", "羊", "猴", "鸡", "狗", "猪"}
	for i := 1; i <= 12; i++ {
		s, ok := zodiac(Zhi(i))
		if !ok {
			t.Errorf("zodiac(%d) not found", i)
		}
		if s != want[i] {
			t.Errorf("zodiac(%d)=%s, want %s", i, s, want[i])
		}
	}
}

func TestZodiac_Invalid(t *testing.T) {
	if _, ok := zodiac(0); ok {
		t.Error("zodiac(0) should not be ok")
	}
	if _, ok := zodiac(13); ok {
		t.Error("zodiac(13) should not be ok")
	}
	if got := ZodiacLabel(0); got != "未知" {
		t.Errorf("ZodiacLabel(0)=%s, want 未知", got)
	}
}

func TestZhiSeason_Known(t *testing.T) {
	spring := []Zhi{ZhiYin, ZhiMao, ZhiChen}
	summer := []Zhi{ZhiSi, ZhiWu, ZhiWei}
	autumn := []Zhi{ZhiShen, ZhiYou, ZhiXu}
	winter := []Zhi{ZhiHai, ZhiZi, ZhiChou}
	for _, z := range spring {
		s, ok := zhiSeason(z)
		if !ok || s != "春" {
			t.Errorf("zhiSeason(%s)=(%s,%v), want (春,true)", z, s, ok)
		}
	}
	for _, z := range summer {
		s, ok := zhiSeason(z)
		if !ok || s != "夏" {
			t.Errorf("zhiSeason(%s)=(%s,%v), want (夏,true)", z, s, ok)
		}
	}
	for _, z := range autumn {
		s, ok := zhiSeason(z)
		if !ok || s != "秋" {
			t.Errorf("zhiSeason(%s)=(%s,%v), want (秋,true)", z, s, ok)
		}
	}
	for _, z := range winter {
		s, ok := zhiSeason(z)
		if !ok || s != "冬" {
			t.Errorf("zhiSeason(%s)=(%s,%v), want (冬,true)", z, s, ok)
		}
	}
}

func TestZhiLunarMonth_Known(t *testing.T) {
	// 正月 = 寅(3)
	s, ok := zhiLunarMonth(ZhiYin)
	if !ok || s != "正月" {
		t.Errorf("zhiLunarMonth(寅)=(%s,%v), want (正月,true)", s, ok)
	}
	// 十一月 = 子(1)
	s, ok = zhiLunarMonth(ZhiZi)
	if !ok || s != "十一月" {
		t.Errorf("zhiLunarMonth(子)=(%s,%v), want (十一月,true)", s, ok)
	}
}

func TestZhiHourRange_Known(t *testing.T) {
	// 子 = 23:00-01:00
	s, ok := zhiHourRange(ZhiZi)
	if !ok || s != "23:00-01:00" {
		t.Errorf("zhiHourRange(子)=(%s,%v), want (23:00-01:00,true)", s, ok)
	}
	// 午 = 11:00-13:00
	s, ok = zhiHourRange(ZhiWu)
	if !ok || s != "11:00-13:00" {
		t.Errorf("zhiHourRange(午)=(%s,%v), want (11:00-13:00,true)", s, ok)
	}
}

// -- JSON serialization --

func TestZhu_MarshalJSON(t *testing.T) {
	z := Zhu{Gan: GanJia, Zhi: ZhiZi}
	b, err := json.Marshal(z)
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}
	var m map[string]string
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if m["gan"] != "甲" || m["zhi"] != "子" {
		t.Errorf("marshaled Zhu = %v, want {甲 子}", m)
	}
}

func TestGan_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		input string
		want  Gan
	}{
		{`"甲"`, GanJia},
		{`"丙"`, GanBing},
		{`"癸"`, GanGui},
		{`1`, Gan(1)},
		{`10`, Gan(10)},
	}
	for _, tc := range tests {
		var g Gan
		if err := json.Unmarshal([]byte(tc.input), &g); err != nil {
			t.Errorf("UnmarshalJSON(%s): %v", tc.input, err)
		}
		if g != tc.want {
			t.Errorf("UnmarshalJSON(%s)=%d, want %d", tc.input, g, tc.want)
		}
	}
}

func TestZhi_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		input string
		want  Zhi
	}{
		{`"子"`, ZhiZi},
		{`"午"`, ZhiWu},
		{`"亥"`, ZhiHai},
		{`1`, Zhi(1)},
		{`12`, Zhi(12)},
	}
	for _, tc := range tests {
		var z Zhi
		if err := json.Unmarshal([]byte(tc.input), &z); err != nil {
			t.Errorf("UnmarshalJSON(%s): %v", tc.input, err)
		}
		if z != tc.want {
			t.Errorf("UnmarshalJSON(%s)=%d, want %d", tc.input, z, tc.want)
		}
	}
}

func TestWuxing_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		input string
		want  Wuxing
	}{
		{`"木"`, WxMu},
		{`"火"`, WxHuo},
		{`"土"`, WxTu},
		{`"金"`, WxJin},
		{`"水"`, WxShui},
	}
	for _, tc := range tests {
		var w Wuxing
		if err := json.Unmarshal([]byte(tc.input), &w); err != nil {
			t.Errorf("UnmarshalJSON(%s): %v", tc.input, err)
		}
		if w != tc.want {
			t.Errorf("UnmarshalJSON(%s)=%d, want %d", tc.input, w, tc.want)
		}
	}
}

// -- Bazi.Validate --

func TestBazi_Validate_Valid(t *testing.T) {
	bz := Bazi{
		Nian: Zhu{Gan: GanJia, Zhi: ZhiZi},
		Yue:  Zhu{Gan: GanYi, Zhi: ZhiChou},
		Ri:   Zhu{Gan: GanBing, Zhi: ZhiYin},
		Shi:  Zhu{Gan: GanDing, Zhi: ZhiMao},
	}
	if err := bz.Validate(); err != nil {
		t.Errorf("valid Bazi should not error: %v", err)
	}
}

func TestBazi_Validate_Invalid(t *testing.T) {
	tests := []struct {
		name string
		bz   Bazi
	}{
		{"gan=0", Bazi{Nian: Zhu{Gan: 0, Zhi: ZhiZi}, Yue: Zhu{Gan: GanYi, Zhi: ZhiChou}, Ri: Zhu{Gan: GanBing, Zhi: ZhiYin}, Shi: Zhu{Gan: GanDing, Zhi: ZhiMao}}},
		{"gan=11", Bazi{Nian: Zhu{Gan: 11, Zhi: ZhiZi}, Yue: Zhu{Gan: GanYi, Zhi: ZhiChou}, Ri: Zhu{Gan: GanBing, Zhi: ZhiYin}, Shi: Zhu{Gan: GanDing, Zhi: ZhiMao}}},
		{"zhi=0", Bazi{Nian: Zhu{Gan: GanJia, Zhi: 0}, Yue: Zhu{Gan: GanYi, Zhi: ZhiChou}, Ri: Zhu{Gan: GanBing, Zhi: ZhiYin}, Shi: Zhu{Gan: GanDing, Zhi: ZhiMao}}},
		{"zhi=13", Bazi{Nian: Zhu{Gan: GanJia, Zhi: 13}, Yue: Zhu{Gan: GanYi, Zhi: ZhiChou}, Ri: Zhu{Gan: GanBing, Zhi: ZhiYin}, Shi: Zhu{Gan: GanDing, Zhi: ZhiMao}}},
	}
	for _, tc := range tests {
		if err := tc.bz.Validate(); err == nil {
			t.Errorf("%s: expected error, got nil", tc.name)
		}
	}
}

func TestBazi_Slice(t *testing.T) {
	bz := Bazi{
		Nian: Zhu{Gan: GanJia, Zhi: ZhiZi},
		Yue:  Zhu{Gan: GanYi, Zhi: ZhiChou},
		Ri:   Zhu{Gan: GanBing, Zhi: ZhiYin},
		Shi:  Zhu{Gan: GanDing, Zhi: ZhiMao},
	}
	s := bz.Slice()
	if len(s) != 4 {
		t.Fatalf("Slice() len=%d, want 4", len(s))
	}
	names := []string{"年", "月", "日", "时"}
	for i, name := range names {
		if s[i].Gan == 0 || s[i].Zhi == 0 {
			t.Errorf("%s pillar is empty", name)
		}
	}
}
