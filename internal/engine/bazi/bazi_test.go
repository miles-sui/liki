package bazi

import (
	"encoding/json"
	"testing"
	"time"

	"liki/internal/engine/ganzhi"
	"liki/internal/engine/tianwen"
)

// =============================================================================
// xunIndex — 旬首
// =============================================================================

func TestXunIndex_AllSixXun(t *testing.T) {
	// 六十甲子分为六旬：甲子旬(0), 甲戌旬(1), 甲申旬(2), 甲午旬(3), 甲辰旬(4), 甲寅旬(5)
	tests := []struct {
		name     string
		gan      ganzhi.Gan
		zhi      ganzhi.Zhi
		wantXun  int
	}{
		{"甲子→旬0", ganzhi.GanJia, ganzhi.ZhiZi, 0},
		{"甲戌→旬1", ganzhi.GanJia, ganzhi.ZhiXu, 1},
		{"甲申→旬2", ganzhi.GanJia, ganzhi.ZhiShen, 2},
		{"甲午→旬3", ganzhi.GanJia, ganzhi.ZhiWu, 3},
		{"甲辰→旬4", ganzhi.GanJia, ganzhi.ZhiChen, 4},
		{"甲寅→旬5", ganzhi.GanJia, ganzhi.ZhiYin, 5},
		// 非旬首 (corrected expectations based on SixtyCycleName)
		{"乙丑→旬0", ganzhi.GanYi, ganzhi.ZhiChou, 0},
		{"癸酉→旬0", ganzhi.GanGui, ganzhi.ZhiYou, 0},
		{"乙亥→旬1", ganzhi.GanYi, ganzhi.ZhiHai, 1},
		{"癸未→旬1", ganzhi.GanGui, ganzhi.ZhiWei, 1},
		{"戊寅→旬1", ganzhi.GanWu, ganzhi.ZhiYin, 1}, // idx=14, xun=1
		{"壬辰→旬2", ganzhi.GanRen, ganzhi.ZhiChen, 2}, // idx=28, xun=2
		{"癸亥→旬5", ganzhi.GanGui, ganzhi.ZhiHai, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := xunIndex(ganzhi.Zhu{Gan: tt.gan, Zhi: tt.zhi})
			if got != tt.wantXun {
				t.Errorf("xunIndex(%s%s)=%d, want %d",
					ganzhi.GanName(tt.gan), ganzhi.ZhiName(tt.zhi), got, tt.wantXun)
			}
		})
	}
}

// =============================================================================
// ComputeXiaoXian — 小限
// =============================================================================

func TestComputeXiaoXian_Male(t *testing.T) {
	// 男命从寅(3)顺行，1岁=寅, 2岁=卯, ...
	xs := ComputeXiaoXian(ganzhi.Male, 13)
	if len(xs) != 13 {
		t.Fatalf("len=%d, want 13", len(xs))
	}
	want := []ganzhi.Zhi{
		ganzhi.ZhiYin, ganzhi.ZhiMao, ganzhi.ZhiChen, ganzhi.ZhiSi,
		ganzhi.ZhiWu, ganzhi.ZhiWei, ganzhi.ZhiShen, ganzhi.ZhiYou,
		ganzhi.ZhiXu, ganzhi.ZhiHai, ganzhi.ZhiZi, ganzhi.ZhiChou,
		ganzhi.ZhiYin, // 13岁回到寅
	}
	for i, w := range want {
		if xs[i].Age != i+1 {
			t.Errorf("age[%d]=%d, want %d", i, xs[i].Age, i+1)
		}
		if xs[i].Zhi != w {
			t.Errorf("age %d: Zhi=%s, want %s", i+1, ganzhi.ZhiName(xs[i].Zhi), ganzhi.ZhiName(w))
		}
	}
}

func TestComputeXiaoXian_Female(t *testing.T) {
	// 女命从申(9)逆行，1岁=申, 2岁=未, ...
	xs := ComputeXiaoXian(ganzhi.Female, 13)
	if len(xs) != 13 {
		t.Fatalf("len=%d, want 13", len(xs))
	}
	want := []ganzhi.Zhi{
		ganzhi.ZhiShen, ganzhi.ZhiWei, ganzhi.ZhiWu, ganzhi.ZhiSi,
		ganzhi.ZhiChen, ganzhi.ZhiMao, ganzhi.ZhiYin, ganzhi.ZhiChou,
		ganzhi.ZhiZi, ganzhi.ZhiHai, ganzhi.ZhiXu, ganzhi.ZhiYou,
		ganzhi.ZhiShen, // 13岁回到申
	}
	for i, w := range want {
		if xs[i].Zhi != w {
			t.Errorf("age %d: Zhi=%s, want %s", i+1, ganzhi.ZhiName(xs[i].Zhi), ganzhi.ZhiName(w))
		}
	}
}

func TestComputeXiaoXian_DefaultMaxAge(t *testing.T) {
	// maxAge <= 0 → default to 12
	xs := ComputeXiaoXian(ganzhi.Male, 0)
	if len(xs) != 12 {
		t.Fatalf("len=%d, want 12", len(xs))
	}
	xs2 := ComputeXiaoXian(ganzhi.Male, -5)
	if len(xs2) != 12 {
		t.Fatalf("len=%d, want 12", len(xs2))
	}
}

// =============================================================================
// ToBazi — ChartBase → ganzhi.Bazi
// =============================================================================

func TestToBazi(t *testing.T) {
	cb := ChartBase{
		Year:  zhuInfo{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi},
		Month: zhuInfo{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiYin},
		Day:   zhuInfo{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiWu},
		Hour:  zhuInfo{Gan: ganzhi.GanGeng, Zhi: ganzhi.ZhiShen},
	}
	bz := cb.ToBazi()
	if bz.Nian.Gan != ganzhi.GanJia || bz.Nian.Zhi != ganzhi.ZhiZi {
		t.Errorf("year: %s%s, want 甲子", ganzhi.GanName(bz.Nian.Gan), ganzhi.ZhiName(bz.Nian.Zhi))
	}
	if bz.Yue.Gan != ganzhi.GanBing || bz.Yue.Zhi != ganzhi.ZhiYin {
		t.Errorf("month: %s%s, want 丙寅", ganzhi.GanName(bz.Yue.Gan), ganzhi.ZhiName(bz.Yue.Zhi))
	}
	if bz.Ri.Gan != ganzhi.GanWu || bz.Ri.Zhi != ganzhi.ZhiWu {
		t.Errorf("day: %s%s, want 戊午", ganzhi.GanName(bz.Ri.Gan), ganzhi.ZhiName(bz.Ri.Zhi))
	}
	if bz.Shi.Gan != ganzhi.GanGeng || bz.Shi.Zhi != ganzhi.ZhiShen {
		t.Errorf("hour: %s%s, want 庚申", ganzhi.GanName(bz.Shi.Gan), ganzhi.ZhiName(bz.Shi.Zhi))
	}
}

// =============================================================================
// NaYinArray — ChartBase → 纳音数组
// =============================================================================

func TestNaYinArray(t *testing.T) {
	cb := ChartBase{
		Year:  zhuInfo{NaYin: "海中金"},
		Month: zhuInfo{NaYin: "炉中火"},
		Day:   zhuInfo{NaYin: "大林木"},
		Hour:  zhuInfo{NaYin: "路旁土"},
	}
	na := cb.NaYinArray()
	want := [4]string{"海中金", "炉中火", "大林木", "路旁土"}
	if na != want {
		t.Errorf("NaYinArray=%v, want %v", na, want)
	}
}

// =============================================================================
// elementThatDrains — 泄 (生我者泄我)
// =============================================================================

func TestElementThatDrains(t *testing.T) {
	// 木生火→木泄火, 火生土→火泄土, ...
	tests := []struct {
		e    ganzhi.Wuxing
		want ganzhi.Wuxing
	}{
		{ganzhi.WxMu, ganzhi.WxHuo},
		{ganzhi.WxHuo, ganzhi.WxTu},
		{ganzhi.WxTu, ganzhi.WxJin},
		{ganzhi.WxJin, ganzhi.WxShui},
		{ganzhi.WxShui, ganzhi.WxMu},
	}
	for _, tt := range tests {
		got := elementThatDrains(tt.e)
		if got != tt.want {
			t.Errorf("elementThatDrains(%s)=%s, want %s", tt.e, got, tt.want)
		}
	}
}

// =============================================================================
// elementThatGenerates / elementThatControls — 生/克
// =============================================================================

func TestElementThatGenerates(t *testing.T) {
	tests := []struct {
		e    ganzhi.Wuxing
		want ganzhi.Wuxing
	}{
		{ganzhi.WxMu, ganzhi.WxShui},   // 木→水生木
		{ganzhi.WxHuo, ganzhi.WxMu},    // 火→木生火
		{ganzhi.WxTu, ganzhi.WxHuo},    // 土→火生土
		{ganzhi.WxJin, ganzhi.WxTu},    // 金→土生金
		{ganzhi.WxShui, ganzhi.WxJin},  // 水→金生水
	}
	for _, tt := range tests {
		got := elementThatGenerates(tt.e)
		if got != tt.want {
			t.Errorf("elementThatGenerates(%s)=%s, want %s", tt.e, got, tt.want)
		}
	}
}

func TestElementThatControls(t *testing.T) {
	tests := []struct {
		e    ganzhi.Wuxing
		want ganzhi.Wuxing
	}{
		{ganzhi.WxMu, ganzhi.WxJin},   // 木→金克木
		{ganzhi.WxHuo, ganzhi.WxShui}, // 火→水克火
		{ganzhi.WxTu, ganzhi.WxMu},    // 土→木克土
		{ganzhi.WxJin, ganzhi.WxHuo},  // 金→火克金
		{ganzhi.WxShui, ganzhi.WxTu},  // 水→土克水
	}
	for _, tt := range tests {
		got := elementThatControls(tt.e)
		if got != tt.want {
			t.Errorf("elementThatControls(%s)=%s, want %s", tt.e, got, tt.want)
		}
	}
}

// =============================================================================
// pickJiElement — 忌神元素
// =============================================================================

func TestPickJiElement_Normal(t *testing.T) {
	// 甲木日主，克木的是金 → yang stem 庚
	ji := pickJiElement(ganzhi.WxMu, ganzhi.GanBing, ganzhi.GanGui)
	if ji != ganzhi.WxJin {
		t.Errorf("甲木忌神=%s, want 金", ji)
	}
}

func TestPickJiElement_YangConflict(t *testing.T) {
	// 当yang stem与yong/xi冲突时, 应fallback到yin stem
	// 甲木日主，用神庚，忌神的yang stem也是庚(冲突) → yin stem 辛
	ji := pickJiElement(ganzhi.WxMu, ganzhi.GanGeng, ganzhi.GanDing)
	if ji != ganzhi.WxJin { // still 金, just the yin stem
		t.Errorf("conflict case: got %s, still want 金", ji)
	}
}

func TestPickJiElement_BothConflict(t *testing.T) {
	// 当yang和yin stem都冲突 → 泄
	// 甲木日主, 用庚辛, 忌神金(yang=庚 yin=辛)都冲突 → 泄(火)
	ji := pickJiElement(ganzhi.WxMu, ganzhi.GanGeng, ganzhi.GanXin)
	if ji != ganzhi.WxHuo {
		t.Errorf("both conflict: got %s, want 火(泄)", ji)
	}
}

// =============================================================================
// ComputeBond — 合盘
// =============================================================================

func TestComputeBond_ValidCharts(t *testing.T) {
	// Construct two ganzhi.WSSi.String()mple charts with enough info for bond computation
	a := ChartBase{
		Year:      zhuInfo{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi, NaYin: "海中金", TenGods: []tenGodEntry{}},
		Month:     zhuInfo{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiYin, NaYin: "炉中火", TenGods: []tenGodEntry{}},
		Day:       zhuInfo{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiWu, NaYin: "天上火", TenGods: []tenGodEntry{}},
		Hour:      zhuInfo{Gan: ganzhi.GanGeng, Zhi: ganzhi.ZhiShen, NaYin: "石榴木", TenGods: []tenGodEntry{}},
		DayMaster: ganzhi.GanWu,
		FuYi:      FuYi{Yong: "火", Ji: "水"},
		WuxingCount: map[ganzhi.Wuxing]int{
			ganzhi.WxMu: 1, ganzhi.WxHuo: 2, ganzhi.WxTu: 1, ganzhi.WxJin: 1, ganzhi.WxShui: 1,
		},
		DaYun: nil, // skip DaYun cross
	}

	b := ChartBase{
		Year:      zhuInfo{Gan: ganzhi.GanYi, Zhi: ganzhi.ZhiChou, NaYin: "海中金", TenGods: []tenGodEntry{}},
		Month:     zhuInfo{Gan: ganzhi.GanDing, Zhi: ganzhi.ZhiMao, NaYin: "炉中火", TenGods: []tenGodEntry{}},
		Day:       zhuInfo{Gan: ganzhi.GanJi, Zhi: ganzhi.ZhiSi, NaYin: "大林木", TenGods: []tenGodEntry{}},
		Hour:      zhuInfo{Gan: ganzhi.GanXin, Zhi: ganzhi.ZhiYou, NaYin: "石榴木", TenGods: []tenGodEntry{}},
		DayMaster: ganzhi.GanJi,
		FuYi:      FuYi{Yong: "土", Ji: "木"},
		WuxingCount: map[ganzhi.Wuxing]int{
			ganzhi.WxMu: 2, ganzhi.WxHuo: 1, ganzhi.WxTu: 1, ganzhi.WxJin: 2, ganzhi.WxShui: 0,
		},
		DaYun: nil,
	}

	bond := ComputeBond(a, b)

	// Verify pillar cross: 4×4 = 16 pairs
	if len(bond.ZhuCross.Pairs) != 16 {
		t.Errorf("ZhuCross.Pairs len=%d, want 16", len(bond.ZhuCross.Pairs))
	}

	// Verify ten god cross
	if len(bond.TenGodCross.AToB) != 4 {
		t.Errorf("TenGodCross.AToB len=%d, want 4", len(bond.TenGodCross.AToB))
	}
	if len(bond.TenGodCross.BToA) != 4 {
		t.Errorf("TenGodCross.BToA len=%d, want 4", len(bond.TenGodCross.BToA))
	}

	// Verify nayin cross
	if len(bond.NayinCross.Pairs) != 16 {
		t.Errorf("NayinCross.Pairs len=%d, want 16", len(bond.NayinCross.Pairs))
	}
	if len(bond.NayinCross.Elements.A) == 0 {
		t.Error("NayinCross.Elements.A is empty")
	}
	if len(bond.NayinCross.Elements.B) == 0 {
		t.Error("NayinCross.Elements.B is empty")
	}

	// Shensha cross: 戊禄在巳(A日柱支巳), 己禄在午(B日柱支午).
	// A的禄巳在B柱中存在(日柱己巳), B的禄午在A柱中存在(日柱戊午).
	if !bond.ShenshaCross.Lu.AInB {
		t.Error("Lu.AInB: 戊禄在巳 should be in B's pillars")
	}
	if !bond.ShenshaCross.Lu.BInA {
		t.Error("Lu.BInA: 己禄在午 should be in A's pillars")
	}
	// 魁罡: 四柱无一为庚辰/庚戌/壬辰/戊戌, KuiGang should be false both ways.
	if bond.ShenshaCross.KuiGang.AInB || bond.ShenshaCross.KuiGang.BInA {
		t.Error("KuiGang: no kui-gang pillars in either chart")
	}

	// Structure: both DaYun are nil → daYun cross entries are zero-valued.
	if bond.Structure.DaYun.ACurrent.Gan != 0 || bond.Structure.DaYun.BCurrent.Gan != 0 {
		t.Error("Structure.DaYun: expected zero-valued entries when DaYun=nil")
	}
	// XunGong: 戊午日(旬1) vs 己巳日(旬0) → different xun, different branch.
	if bond.Structure.XunGong.SameXun {
		t.Error("XunGong.SameXun: 戊午(旬1) and 己巳(旬0) are in different xun")
	}
	if bond.Structure.XunGong.SameGong {
		t.Error("XunGong.SameGong: 午 ≠ 巳")
	}
}

func TestComputeBond_WithDaYun(t *testing.T) {
	drZhus := []DaYunZhu{
		{Name: "丙寅", TenGod: "偏印", Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiYin},
		{Name: "丁卯", TenGod: "正印", Gan: ganzhi.GanDing, Zhi: ganzhi.ZhiMao},
	}

	a := ChartBase{
		Year:      zhuInfo{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi, NaYin: "海中金"},
		Month:     zhuInfo{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiYin, NaYin: "炉中火"},
		Day:       zhuInfo{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiWu, NaYin: "天上火"},
		Hour:      zhuInfo{Gan: ganzhi.GanGeng, Zhi: ganzhi.ZhiShen, NaYin: "石榴木"},
		DayMaster: ganzhi.GanWu,
		FuYi:      FuYi{Yong: "火", Ji: "水"},
		WuxingCount: map[ganzhi.Wuxing]int{
			ganzhi.WxMu: 1, ganzhi.WxHuo: 2, ganzhi.WxTu: 1, ganzhi.WxJin: 1, ganzhi.WxShui: 1,
		},
		DaYun: &DaYun{
			Zhus:            drZhus,
			CurrentZhuIndex: 0,
		},
	}

	b := ChartBase{
		Year:      zhuInfo{Gan: ganzhi.GanYi, Zhi: ganzhi.ZhiChou, NaYin: "海中金"},
		Month:     zhuInfo{Gan: ganzhi.GanDing, Zhi: ganzhi.ZhiMao, NaYin: "炉中火"},
		Day:       zhuInfo{Gan: ganzhi.GanJi, Zhi: ganzhi.ZhiSi, NaYin: "大林木"},
		Hour:      zhuInfo{Gan: ganzhi.GanXin, Zhi: ganzhi.ZhiYou, NaYin: "石榴木"},
		DayMaster: ganzhi.GanJi,
		FuYi:      FuYi{Yong: "土", Ji: "木"},
		WuxingCount: map[ganzhi.Wuxing]int{
			ganzhi.WxMu: 2, ganzhi.WxHuo: 1, ganzhi.WxTu: 1, ganzhi.WxJin: 2, ganzhi.WxShui: 0,
		},
		DaYun: &DaYun{
			Zhus:            drZhus,
			CurrentZhuIndex: 1,
		},
	}

	bond := ComputeBond(a, b)

	// Verify DaYun cross
	if bond.Structure.DaYun.ACurrent.Gan != ganzhi.GanBing {
		t.Errorf("A current DaYun Gan=%s, want 丙", ganzhi.GanName(bond.Structure.DaYun.ACurrent.Gan))
	}
	if bond.Structure.DaYun.BCurrent.Gan != ganzhi.GanDing {
		t.Errorf("B current DaYun Gan=%s, want 丁", ganzhi.GanName(bond.Structure.DaYun.BCurrent.Gan))
	}

	// Verify XunGong (different day branches)
	if bond.Structure.XunGong.SameGong {
		t.Error("SameGong should be false (午 vs 巳)")
	}
}

func TestComputeBond_WithChart(t *testing.T) {
	// Integration with full ComputeChart
	st := tianwen.ComputeSolarTime(1990, 5, 20, 12, 0, 120, 8)
	c1 := ComputeChart(st, ganzhi.Male)
	c2 := ComputeChart(st, ganzhi.Female)

	bond := ComputeBond(c1.ChartBase, c2.ChartBase)

	if len(bond.ZhuCross.Pairs) != 16 {
		t.Errorf("ZhuCross.Pairs len=%d, want 16", len(bond.ZhuCross.Pairs))
	}
	// Verify all pillar pairs have valid stem/branch info
	for _, p := range bond.ZhuCross.Pairs {
		if p.AStem == "" || p.BStem == "" {
			t.Error("pillar cross pair has empty stem")
		}
	}
}

// =============================================================================
// NayinElement — 纳音取五行
// =============================================================================

func TestNayinElement(t *testing.T) {
	tests := []struct {
		nayin string
		want  ganzhi.Wuxing
	}{
		{"海中金", ganzhi.WxJin},
		{"炉中火", ganzhi.WxHuo},
		{"大林木", ganzhi.WxMu},
		{"路旁土", ganzhi.WxTu},
		{"涧下水", ganzhi.WxShui},
	}
	for _, tt := range tests {
		got := ganzhi.NaYinWuxing(tt.nayin)
		if got != tt.want {
			t.Errorf("ganzhi.NaYinWuxing(%s)=%s, want %s", tt.nayin, got, tt.want)
		}
	}
}

func TestNayinElement_Short(t *testing.T) {
	// nayinElement extracts the last character as the wuxing element
	if got := ganzhi.NaYinWuxing("金"); got != ganzhi.WxJin {
		t.Errorf("ganzhi.NaYinWuxing(金)=%d, want %d(金)", got, ganzhi.WxJin)
	}
	// Empty string → last char extraction fails
	if got := ganzhi.NaYinWuxing(""); got != 0 {
		t.Errorf("ganzhi.NaYinWuxing('')=%d, want 0", got)
	}
}

// =============================================================================
// ComputeChart — 完整八字排盘
// =============================================================================

func TestComputeChart_ValidChart(t *testing.T) {
	st := tianwen.ComputeSolarTime(1984, 2, 15, 8, 0, 120, 8)
	c := ComputeChart(st, ganzhi.Male)

	// DayMaster should be set
	if c.DayMaster < 1 || c.DayMaster > 10 {
		t.Errorf("DayMaster=%d, want 1-10", c.DayMaster)
	}

	// All four pillars should have valid gan/zhi
	zhus := []zhuInfo{c.Year, c.Month, c.Day, c.Hour}
	for i, p := range zhus {
		if p.Gan < 1 || p.Gan > 10 {
			t.Errorf("pillar[%d] Gan=%d invalid", i, p.Gan)
		}
		if p.Zhi < 1 || p.Zhi > 12 {
			t.Errorf("pillar[%d] Zhi=%d invalid", i, p.Zhi)
		}
		if p.NaYin == "" {
			t.Errorf("pillar[%d] NaYin empty", i)
		}
	}

	// HeHui may or may not have entries, depending on the chart

	// WuxingCount should sum to 5 elements
	if len(c.WuxingCount) == 0 {
		t.Error("WuxingCount is empty")
	}

	// TaiYuanMingGong should be set
	if c.TaiYuanMingGong.TaiYuan.Gan == 0 {
		t.Error("TaiYuan.Gan is zero")
	}
}

func TestComputeChart_DayMasterConsistency(t *testing.T) {
	// DayMaster should equal Day.Gan
	st := tianwen.ComputeSolarTime(2000, 6, 15, 12, 0, 120, 8)
	c := ComputeChart(st, ganzhi.Female)
	if c.DayMaster != c.Day.Gan {
		t.Errorf("DayMaster(%s) != Day.Gan(%s)",
			ganzhi.GanName(c.DayMaster), ganzhi.GanName(c.Day.Gan))
	}
}

// =============================================================================
// ComputeChart JSON — 序列化验证
// =============================================================================

func TestChart_JSONRoundtrip(t *testing.T) {
	st := tianwen.ComputeSolarTime(1990, 5, 20, 12, 0, 121.5, 8)
	c := ComputeChart(st, ganzhi.Male)

	b, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var c2 Chart
	if err := json.Unmarshal(b, &c2); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	// Verify DayMaster survives roundtrip
	if c2.DayMaster != c.DayMaster {
		t.Errorf("DayMaster after roundtrip: %d, want %d", c2.DayMaster, c.DayMaster)
	}
}

// =============================================================================
// MingGe String — 命格名
// =============================================================================

func TestFuYi_ComputedFields(t *testing.T) {
	st := tianwen.ComputeSolarTime(1990, 5, 20, 12, 0, 121.5, 8)
	c := ComputeChart(st, ganzhi.Male)
	// FuYi should be populated
	if c.FuYi.Strength == "" {
		t.Error("FuYi.Strength is empty")
	}
	if c.FuYi.Yong == "" {
		t.Error("FuYi.Yong is empty")
	}
	if c.FuYi.Ji == "" {
		t.Error("FuYi.Ji is empty")
	}
}

// =============================================================================
// Strength.String — 旺衰字符串
// =============================================================================

func TestStrength_String(t *testing.T) {
	tests := []struct {
		name string
		s    strength
		want string
	}{
		{"身弱", strengthWeak, "身弱"},
		{"中和", strengthNeutral, "中和"},
		{"身强", strengthStrong, "身强"},
		{"invalid→empty", strength(99), ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.String(); got != tt.want {
				t.Errorf("String()=%q, want %q", got, tt.want)
			}
		})
	}
}

// =============================================================================
// monthWangShuai — 旺相休囚死 (all 5 elements × 12 months)
// =============================================================================

func TestMonthWangShuai_AllElements(t *testing.T) {
	// 规则: 当令者旺 / 我生者相 / 生我者休 / 克我者囚 / 我克者死
	// Where "我" = month element (月令五行).
	tests := []struct {
		name        string
		elem        ganzhi.Wuxing
		monthBranch ganzhi.Zhi
		want        string
	}{
		// ── 寅月(木旺) ──
		{"木在寅月→旺", ganzhi.WxMu, ganzhi.ZhiYin, ganzhi.WSWang.String()},
		{"火在寅月→相", ganzhi.WxHuo, ganzhi.ZhiYin, ganzhi.WSXiang.String()}, // 木生火
		{"水在寅月→休", ganzhi.WxShui, ganzhi.ZhiYin, ganzhi.WSXiu.String()}, // 水生木
		{"金在寅月→囚", ganzhi.WxJin, ganzhi.ZhiYin, ganzhi.WSQiu.String()},  // 金克木
		{"土在寅月→死", ganzhi.WxTu, ganzhi.ZhiYin, ganzhi.WSSi.String()},    // 木克土
		// ── 巳月(火旺) ──
		{"火在巳月→旺", ganzhi.WxHuo, ganzhi.ZhiSi, ganzhi.WSWang.String()},
		{"土在巳月→相", ganzhi.WxTu, ganzhi.ZhiSi, ganzhi.WSXiang.String()}, // 火生土
		{"木在巳月→休", ganzhi.WxMu, ganzhi.ZhiSi, ganzhi.WSXiu.String()},   // 木生火
		{"水在巳月→囚", ganzhi.WxShui, ganzhi.ZhiSi, ganzhi.WSQiu.String()}, // 水克火
		{"金在巳月→死", ganzhi.WxJin, ganzhi.ZhiSi, ganzhi.WSSi.String()},   // 火克金
		// ── 申月(金旺) ──
		{"金在申月→旺", ganzhi.WxJin, ganzhi.ZhiShen, ganzhi.WSWang.String()},
		{"水在申月→相", ganzhi.WxShui, ganzhi.ZhiShen, ganzhi.WSXiang.String()}, // 金生水
		{"土在申月→休", ganzhi.WxTu, ganzhi.ZhiShen, ganzhi.WSXiu.String()},     // 土生金
		{"火在申月→囚", ganzhi.WxHuo, ganzhi.ZhiShen, ganzhi.WSQiu.String()},    // 火克金
		{"木在申月→死", ganzhi.WxMu, ganzhi.ZhiShen, ganzhi.WSSi.String()},      // 金克木
		// ── 亥月(水旺) ──
		{"水在亥月→旺", ganzhi.WxShui, ganzhi.ZhiHai, ganzhi.WSWang.String()},
		{"木在亥月→相", ganzhi.WxMu, ganzhi.ZhiHai, ganzhi.WSXiang.String()}, // 水生木
		{"金在亥月→休", ganzhi.WxJin, ganzhi.ZhiHai, ganzhi.WSXiu.String()},  // 金生水
		{"土在亥月→囚", ganzhi.WxTu, ganzhi.ZhiHai, ganzhi.WSQiu.String()},   // 土克水
		{"火在亥月→死", ganzhi.WxHuo, ganzhi.ZhiHai, ganzhi.WSSi.String()},   // 水克火
		// ── 辰月(土旺) ──
		{"土在辰月→旺", ganzhi.WxTu, ganzhi.ZhiChen, ganzhi.WSWang.String()},
		{"金在辰月→相", ganzhi.WxJin, ganzhi.ZhiChen, ganzhi.WSXiang.String()}, // 土生金
		{"火在辰月→休", ganzhi.WxHuo, ganzhi.ZhiChen, ganzhi.WSXiu.String()},   // 火生土
		{"木在辰月→囚", ganzhi.WxMu, ganzhi.ZhiChen, ganzhi.WSQiu.String()},    // 木克土
		{"水在辰月→死", ganzhi.WxShui, ganzhi.ZhiChen, ganzhi.WSSi.String()},   // 土克水
		// ── 午月(火旺)—同巳月 ──
		{"火在午月→旺", ganzhi.WxHuo, ganzhi.ZhiWu, ganzhi.WSWang.String()},
		{"木在午月→休", ganzhi.WxMu, ganzhi.ZhiWu, ganzhi.WSXiu.String()},
		// ── boundary: invalid branch ──
		{"invalid branch→empty", ganzhi.WxMu, ganzhi.Zhi(0), ""},
		{"invalid branch 13→empty", ganzhi.WxHuo, ganzhi.Zhi(13), ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ganzhi.WangShuaiOf(tt.elem, tt.monthBranch).String()
			if got != tt.want {
				t.Errorf("monthWangShuai(%s, %s)=%q, want %q",
					tt.elem, ganzhi.ZhiName(tt.monthBranch), got, tt.want)
			}
		})
	}
}

// =============================================================================
// sanQiType / sanQiName — 三奇贵人
// =============================================================================

func TestSanQiType(t *testing.T) {
	tests := []struct {
		name string
		zhus [4]ganzhi.Zhu
		want   string
	}{
		// 天上三奇: 甲(1)+戊(5)+庚(7)
		{"天上三奇甲戊庚", [4]ganzhi.Zhu{
			{Gan: 1, Zhi: 1}, {Gan: 5, Zhi: 2}, {Gan: 7, Zhi: 3}, {Gan: 8, Zhi: 4},
		}, "天上"},
		// 地下三奇: 乙(2)+丙(3)+丁(4)
		{"地下三奇乙丙丁", [4]ganzhi.Zhu{
			{Gan: 1, Zhi: 1}, {Gan: 2, Zhi: 2}, {Gan: 3, Zhi: 3}, {Gan: 4, Zhi: 4},
		}, "地下"},
		// 人中三奇: 壬(9)+癸(10)+辛(8)
		{"人中三奇壬癸辛", [4]ganzhi.Zhu{
			{Gan: 9, Zhi: 1}, {Gan: 10, Zhi: 2}, {Gan: 8, Zhi: 3}, {Gan: 2, Zhi: 4},
		}, "人中"},
		// 缺一天上三奇(缺庚)
		{"缺庚→不是三奇", [4]ganzhi.Zhu{
			{Gan: 1, Zhi: 1}, {Gan: 5, Zhi: 2}, {Gan: 9, Zhi: 3}, {Gan: 10, Zhi: 4},
		}, ""},
		// 只有两柱同干→不算
		{"普通八字→空", [4]ganzhi.Zhu{
			{Gan: 1, Zhi: 1}, {Gan: 2, Zhi: 2}, {Gan: 8, Zhi: 3}, {Gan: 9, Zhi: 4},
		}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bz := ganzhi.Bazi{
				Nian: tt.zhus[0], Yue: tt.zhus[1],
				Ri: tt.zhus[2], Shi: tt.zhus[3],
			}
			got := sanQiType(bz)
			if got != tt.want {
				t.Errorf("sanQiType()=%q, want %q", got, tt.want)
			}
		})
	}
}

func TestSanQiName(t *testing.T) {
	tests := []struct {
		typ  string
		want string
	}{
		{"天上", "天上三奇（甲戊庚）"},
		{"地下", "地下三奇（乙丙丁）"},
		{"人中", "人中三奇（壬癸辛）"},
		{"", ""},
		{"unknown", ""},
	}
	for _, tt := range tests {
		t.Run(tt.typ, func(t *testing.T) {
			if got := sanQiName(tt.typ); got != tt.want {
				t.Errorf("sanQiName(%q)=%q, want %q", tt.typ, got, tt.want)
			}
		})
	}
}

// =============================================================================
// isSelfHe / selfHeName — 干支自合
// =============================================================================

func TestIsSelfHe(t *testing.T) {
	// Known self-he pairs from domain knowledge:
	// 甲午, 乙巳, 丙戌, 丁亥, 戊子, 庚辰, 辛巳, 壬戌, 癸巳
	selfHePairs := []struct {
		gan ganzhi.Gan
		zhi ganzhi.Zhi
	}{
		{ganzhi.GanJia, ganzhi.ZhiWu},   // 甲午 (甲+午中己)
		{ganzhi.GanYi, ganzhi.ZhiSi},    // 乙巳 (乙+巳中庚)
		{ganzhi.GanBing, ganzhi.ZhiXu},   // 丙戌 (丙+戌中辛)
		{ganzhi.GanDing, ganzhi.ZhiHai},  // 丁亥 (丁+亥中壬)
		{ganzhi.GanWu, ganzhi.ZhiZi},     // 戊子 (戊+子中癸)
		{ganzhi.GanGeng, ganzhi.ZhiChen}, // 庚辰 (庚+辰中乙)
		{ganzhi.GanXin, ganzhi.ZhiSi},    // 辛巳 (辛+巳中丙)
		{ganzhi.GanRen, ganzhi.ZhiXu},    // 壬戌 (壬+戌中丁)
		{ganzhi.GanGui, ganzhi.ZhiSi},    // 癸巳 (癸+巳中戊)
	}
	for _, p := range selfHePairs {
		t.Run(ganzhi.GanName(p.gan)+ganzhi.ZhiName(p.zhi), func(t *testing.T) {
			if !isSelfHe(ganzhi.Zhu{Gan: p.gan, Zhi: p.zhi}) {
				t.Errorf("%s%s should be self-he", ganzhi.GanName(p.gan), ganzhi.ZhiName(p.zhi))
			}
			name := selfHeName(ganzhi.Zhu{Gan: p.gan, Zhi: p.zhi})
			if name == "" {
				t.Error("selfHeName returned empty")
			}
		})
	}

	// Non-self-he pairs
	nonSelfHe := []ganzhi.Zhu{
		{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi},  // 甲子: 子中癸→不配甲
		{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiWu}, // 丙午: 午中己→不配丙
		{Gan: ganzhi.GanRen, Zhi: ganzhi.ZhiZi},      // 壬子: 子中癸→不配壬
	}
	for _, p := range nonSelfHe {
		t.Run(ganzhi.GanName(p.Gan)+ganzhi.ZhiName(p.Zhi)+"→否", func(t *testing.T) {
			if isSelfHe(p) {
				t.Errorf("%s%s should NOT be self-he", ganzhi.GanName(p.Gan), ganzhi.ZhiName(p.Zhi))
			}
		})
	}
}

// =============================================================================
// isKuiGang — 魁罡
// =============================================================================

func TestIsKuiGang(t *testing.T) {
	// 魁罡四日: 庚辰, 庚戌, 壬辰, 戊戌
	kuiGang := []ganzhi.Zhu{
		{Gan: ganzhi.GanGeng, Zhi: ganzhi.ZhiChen}, // 庚辰
		{Gan: ganzhi.GanGeng, Zhi: ganzhi.ZhiXu},   // 庚戌
		{Gan: ganzhi.GanRen, Zhi: ganzhi.ZhiChen},  // 壬辰
		{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiXu},     // 戊戌
	}
	for _, p := range kuiGang {
		t.Run(ganzhi.GanName(p.Gan)+ganzhi.ZhiName(p.Zhi), func(t *testing.T) {
			if !isKuiGang(p) {
				t.Errorf("%s%s should be 魁罡", ganzhi.GanName(p.Gan), ganzhi.ZhiName(p.Zhi))
			}
		})
	}

	// 近似但非魁罡
	nonKuiGang := []ganzhi.Zhu{
		{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiChen},  // 甲辰
		{Gan: ganzhi.GanGeng, Zhi: ganzhi.ZhiShen}, // 庚申
		{Gan: ganzhi.GanRen, Zhi: ganzhi.ZhiZi},    // 壬子
		{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiXu},   // 丙戌
	}
	for _, p := range nonKuiGang {
		t.Run(ganzhi.GanName(p.Gan)+ganzhi.ZhiName(p.Zhi)+"→否", func(t *testing.T) {
			if isKuiGang(p) {
				t.Errorf("%s%s should NOT be 魁罡", ganzhi.GanName(p.Gan), ganzhi.ZhiName(p.Zhi))
			}
		})
	}
}

// =============================================================================
// computeKongWang — 空亡
// =============================================================================

func TestComputeKongWang_HitsVoidBranch(t *testing.T) {
	tests := []struct {
		name       string
		bz         ganzhi.Bazi
		wantHits   int
		wantPillar int // which pillar index hits (0=年,1=月,2=日,3=时)
	}{
		{
			// 甲子旬(0) 空戌(11)亥(12), 年柱戌→空亡
			"甲子旬→年柱戌空",
			ganzhi.Bazi{
				Nian: ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiXu},
				Yue:  ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiChen},
				Ri:   ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi},
				Shi:  ganzhi.Zhu{Gan: ganzhi.GanGeng, Zhi: ganzhi.ZhiWu},
			}, 1, 0,
		},
		{
			// 甲寅旬(5) 空子(1)丑(2), 时柱丑→空亡
			"甲寅旬→时柱丑空",
			ganzhi.Bazi{
				Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiYin},
				Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiChen},
				Ri:   ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiYin},
				Shi:  ganzhi.Zhu{Gan: ganzhi.GanYi, Zhi: ganzhi.ZhiChou}, // 丑在甲寅旬为空
			}, 1, 3,
		},
		{
			// 甲申旬(2) 空午(7)未(8), 月柱午+日柱未→双空
			"甲申旬→月午日未双空",
			ganzhi.Bazi{
				Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiShen},
				Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiWu},  // 午在甲申旬为空
				Ri:   ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiWei},  // 未在甲申旬为空
				Shi:  ganzhi.Zhu{Gan: ganzhi.GanYi, Zhi: ganzhi.ZhiMao},
			}, 2, -1, // 多柱命中，不验具体位置
		},
		{
			// 甲子旬(0) 空戌亥, 四柱均不空
			"无空亡",
			ganzhi.Bazi{
				Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi},
				Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiChen},
				Ri:   ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi},
				Shi:  ganzhi.Zhu{Gan: ganzhi.GanYi, Zhi: ganzhi.ZhiChou},
			}, 0, -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hits := computeKongWang(tt.bz)
			if len(hits) != tt.wantHits {
				t.Errorf("hits len=%d, want %d, hits=%v", len(hits), tt.wantHits, hits)
			}
			if tt.wantPillar >= 0 && (len(hits) == 0 || hits[0] != tt.wantPillar) {
				t.Errorf("expected pillar %d to hit, got hits=%v", tt.wantPillar, hits)
			}
		})
	}
}

// =============================================================================
// computeTaiYuanMingGong — 胎元/命宫/身宫
// =============================================================================

func TestComputeTaiYuanMingGong(t *testing.T) {
	// 1990-05-20 12:00, month=丙寅(gan=3,zhi=3), year stem=庚(7), birthMonth=4, hour=午(7)
	// birthMonth=4 (month of 巳, index 4)
	// 胎元: stem=3+1=4(丁), branch=3+3=6(巳) → 丁巳
	// 命宫: monthOnZi=(1-(4-1)+12)%12=10, hourBranch=(11+1)/2%12+1=6+1=7
	//   mgBranch=(10+7-1)%12+1=5(辰), mgMonthIdx=((5-3+12)%12)+1=3
	//   mgStem=(7*2+3)%10=7(庚), → 庚辰
	// 身宫: shenStart=4, sgBranch=(4-7+1+12)%12=10(酉)
	//   sgMonthIdx=((10-3+12)%12)+1=8, sgStem=(7*2+8)%10=2(乙) → 乙酉
	result := computeTaiYuanMingGong(
		ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiYin}, // 丙寅月
		ganzhi.GanGeng, // 庚年
		4,   // birthMonth (巳月=4)
		11,  // birthHour (11=午时)
	)

	// 胎元
	if result.TaiYuan.Gan != ganzhi.GanDing || result.TaiYuan.Zhi != ganzhi.ZhiSi {
		t.Errorf("TaiYuan=%s%s, want 丁巳", ganzhi.GanName(result.TaiYuan.Gan), ganzhi.ZhiName(result.TaiYuan.Zhi))
	}
	// 命宫: monthOnZi=10, hourBranch=7, mgBranch=(10+7-1)%12=4(卯), mgStem=(7*2+2)%10=6(己)
	if result.MingGong.Gan != ganzhi.GanJi || result.MingGong.Zhi != ganzhi.ZhiMao {
		t.Errorf("MingGong=%s%s, want 己卯", ganzhi.GanName(result.MingGong.Gan), ganzhi.ZhiName(result.MingGong.Zhi))
	}
	// 身宫
	if result.ShenGong.Gan != ganzhi.GanYi || result.ShenGong.Zhi != ganzhi.ZhiYou {
		t.Errorf("ShenGong=%s%s, want 乙酉", ganzhi.GanName(result.ShenGong.Gan), ganzhi.ZhiName(result.ShenGong.Zhi))
	}
}

func TestComputeTaiYuanMingGong_January(t *testing.T) {
	// 正月(寅月=1) + 子时(0) + 甲年
	// 命宫: monthOnZi=(1-(1-1)+12)%12=1, hourBranch=(0+1)/2%12+1=1(子)
	//   mgBranch=(1+1-1)%12=1(子), mgMonthIdx=((1-3+12)%12)+1=11
	//   mgStem=(1*2+11)%10=3(丙) → 丙子
	// 身宫: shenStart=1, sgBranch=(1-1+1+12)%12=1(子)
	//   sgMonthIdx=((1-3+12)%12)+1=11, sgStem=(1*2+11)%10=3(丙) → 丙子
	result := computeTaiYuanMingGong(
		ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi}, // 甲子月
		ganzhi.GanJia, // 甲年
		1,  // 正月
		0,  // 子时
	)
	// 胎元: stem=1+1=2(乙), branch=1+3=4(卯) → 乙卯
	if result.TaiYuan.Gan != ganzhi.GanYi || result.TaiYuan.Zhi != ganzhi.ZhiMao {
		t.Errorf("TaiYuan=%s%s, want 乙卯", ganzhi.GanName(result.TaiYuan.Gan), ganzhi.ZhiName(result.TaiYuan.Zhi))
	}
	// 命宫: 丙子
	if result.MingGong.Gan != ganzhi.GanBing || result.MingGong.Zhi != ganzhi.ZhiZi {
		t.Errorf("MingGong=%s%s, want 丙子", ganzhi.GanName(result.MingGong.Gan), ganzhi.ZhiName(result.MingGong.Zhi))
	}
	// 身宫: 丙子
	if result.ShenGong.Gan != ganzhi.GanBing || result.ShenGong.Zhi != ganzhi.ZhiZi {
		t.Errorf("ShenGong=%s%s, want 丙子", ganzhi.GanName(result.ShenGong.Gan), ganzhi.ZhiName(result.ShenGong.Zhi))
	}
}

func TestComputeTaiYuanMingGong_StemWrapAround(t *testing.T) {
	// 癸丑月(stem=10, branch=2): 胎元 stem=10+1=11→1(甲), branch=2+3=5(辰)
	result := computeTaiYuanMingGong(
		ganzhi.Zhu{Gan: ganzhi.GanGui, Zhi: ganzhi.ZhiChou}, // 癸丑
		ganzhi.GanJia, // 甲年
		12, // 十二月
		23, // 子时
	)
	if result.TaiYuan.Gan != ganzhi.GanJia || result.TaiYuan.Zhi != ganzhi.ZhiChen {
		t.Errorf("TaiYuan=%s%s, want 甲辰", ganzhi.GanName(result.TaiYuan.Gan), ganzhi.ZhiName(result.TaiYuan.Zhi))
	}
}

// =============================================================================
// computePattern — 格局
// =============================================================================

func TestComputePattern_AllTypes(t *testing.T) {
	// 格局基于月令和月干十神
	tests := []struct {
		name           string
		dayMaster      ganzhi.Gan
		monthBranch    ganzhi.Zhi
		monthTenGod    ganzhi.TenGod
		want           string
	}{
		// 建禄格: 月令与日主同五行，阳干
		{"甲日寅月→建禄格", ganzhi.GanJia, ganzhi.ZhiYin, ganzhi.TenGodBiJian, "建禄格"},
		// 月刃格: 月令与日主同五行，阴干
		{"乙日卯月→月刃格", ganzhi.GanYi, ganzhi.ZhiMao, ganzhi.TenGodBiJian, "月刃格"},
		// 正官格
		{"甲日酉月→正官格", ganzhi.GanJia, ganzhi.ZhiYou, ganzhi.TenGodZhengGuan, "正官格"},
		// 七杀格
		{"甲日申月→七杀格", ganzhi.GanJia, ganzhi.ZhiShen, ganzhi.TenGodQiSha, "七杀格"},
		// 正财格
		{"甲日丑月→正财格", ganzhi.GanJia, ganzhi.ZhiChou, ganzhi.TenGodZhengCai, "正财格"},
		// 偏财格
		{"丙日酉月→偏财格", ganzhi.GanBing, ganzhi.ZhiYou, ganzhi.TenGodPianCai, "偏财格"},
		// 正印格
		{"丙日子月→正印格", ganzhi.GanBing, ganzhi.ZhiZi, ganzhi.TenGodZhengYin, "正印格"},
		// 偏印格
		{"丙日亥月→偏印格", ganzhi.GanBing, ganzhi.ZhiHai, ganzhi.TenGodPianYin, "偏印格"},
		// 食神格
		{"戊日申月→食神格", ganzhi.GanWu, ganzhi.ZhiShen, ganzhi.TenGodShiShen, "食神格"},
		// 伤官格
		{"戊日亥月→伤官格", ganzhi.GanWu, ganzhi.ZhiHai, ganzhi.TenGodShangGuan, "伤官格"},
		// 杂格 (月干十神不匹配时)
		{"杂格", ganzhi.GanJia, ganzhi.ZhiYou, ganzhi.TenGodBiJian, "杂格"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computePattern(tt.dayMaster, tt.monthBranch, tt.monthTenGod)
			if got != tt.want {
				t.Errorf("computePattern()=%q, want %q", got, tt.want)
			}
		})
	}
}

// =============================================================================
// 大运排盘 — computeDaYunZhus boundary tests
// =============================================================================

func TestComputeDaYunZhus_ShunPai(t *testing.T) {
	// 阳男: year stem 甲(1, yang), male → 顺排
	// 月柱丙寅: 顺排→丁卯, 戊辰, 己巳...
	st := tianwen.SolarTime(time.Date(2000, 2, 5, 12, 0, 0, 0, time.UTC)) // 立春后，寅月
	month := ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiYin}           // 丙寅
	result := computeDaYunZhus(st, month, ganzhi.GanJia, ganzhi.Male)

	if result.direction != "顺排" {
		t.Errorf("direction=%q, want 顺排", result.direction)
	}
	if len(result.zhus) != 8 {
		t.Fatalf("len=%d, want 8", len(result.zhus))
	}
	// 丙寅→丁卯→戊辰→己巳→庚午→辛未→壬申→癸酉
	expected := []struct {
		gan int
		zhi int
	}{
		{4, 4}, // 丁卯
		{5, 5}, // 戊辰
		{6, 6}, // 己巳
		{7, 7}, // 庚午
		{8, 8}, // 辛未
		{9, 9}, // 壬申
		{10, 10}, // 癸酉
		{1, 11},  // 甲戌
	}
	for i, exp := range expected {
		if int(result.zhus[i].Gan) != exp.gan || int(result.zhus[i].Zhi) != exp.zhi {
			t.Errorf("pillar[%d]=%s%s, want (%d,%d)",
				i, ganzhi.GanName(result.zhus[i].Gan), ganzhi.ZhiName(result.zhus[i].Zhi), exp.gan, exp.zhi)
		}
	}
}

func TestComputeDaYunZhus_NiPai(t *testing.T) {
	// 阳女: year stem 甲(1, yang), female → 逆排
	// 月柱丙寅: 逆排→乙丑, 甲子, 癸亥...
	st := tianwen.SolarTime(time.Date(2000, 2, 5, 12, 0, 0, 0, time.UTC)) // 立春后，寅月
	month := ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiYin}           // 丙寅
	result := computeDaYunZhus(st, month, ganzhi.GanJia, ganzhi.Female)

	if result.direction != "逆排" {
		t.Errorf("direction=%q, want 逆排", result.direction)
	}
	if len(result.zhus) != 8 {
		t.Fatalf("len=%d, want 8", len(result.zhus))
	}
	// 丙寅→乙丑→甲子→癸亥→壬戌→辛酉→庚申→己未
	expected := []struct {
		gan int
		zhi int
	}{
		{2, 2},  // 乙丑
		{1, 1},  // 甲子
		{10, 12}, // 癸亥
		{9, 11},  // 壬戌
		{8, 10},  // 辛酉
		{7, 9},   // 庚申
		{6, 8},   // 己未
		{5, 7},   // 戊午
	}
	for i, exp := range expected {
		if int(result.zhus[i].Gan) != exp.gan || int(result.zhus[i].Zhi) != exp.zhi {
			t.Errorf("pillar[%d]=%s%s, want (%d,%d)",
				i, ganzhi.GanName(result.zhus[i].Gan), ganzhi.ZhiName(result.zhus[i].Zhi), exp.gan, exp.zhi)
		}
	}
}

func TestComputeDaYunZhus_YinNan(t *testing.T) {
	// 阴男: year stem 乙(2, yin), male → 逆排
	st := tianwen.SolarTime(time.Date(2000, 2, 5, 12, 0, 0, 0, time.UTC))
	month := ganzhi.Zhu{Gan: ganzhi.GanDing, Zhi: ganzhi.ZhiSi} // 丁巳
	result := computeDaYunZhus(st, month, ganzhi.GanYi, ganzhi.Male)

	if result.direction != "逆排" {
		t.Errorf("阴男: direction=%q, want 逆排", result.direction)
	}
	// 丁巳→丙辰→乙卯→甲寅...
	if int(result.zhus[0].Gan) != 3 || int(result.zhus[0].Zhi) != 5 {
		t.Errorf("first pillar=%s%s, want 丙辰",
			ganzhi.GanName(result.zhus[0].Gan), ganzhi.ZhiName(result.zhus[0].Zhi))
	}
}

func TestComputeDaYunZhus_YinNv(t *testing.T) {
	// 阴女: year stem 乙(2, yin), female → 顺排
	st := tianwen.SolarTime(time.Date(2000, 2, 5, 12, 0, 0, 0, time.UTC))
	month := ganzhi.Zhu{Gan: ganzhi.GanDing, Zhi: ganzhi.ZhiSi} // 丁巳
	result := computeDaYunZhus(st, month, ganzhi.GanYi, ganzhi.Female)

	if result.direction != "顺排" {
		t.Errorf("阴女: direction=%q, want 顺排", result.direction)
	}
	// 丁巳→戊午→己未...
	if int(result.zhus[0].Gan) != 5 || int(result.zhus[0].Zhi) != 7 {
		t.Errorf("first pillar=%s%s, want 戊午",
			ganzhi.GanName(result.zhus[0].Gan), ganzhi.ZhiName(result.zhus[0].Zhi))
	}
}

// =============================================================================
// daYunTenGodLabel — 大运十神标签
// =============================================================================

func TestDaYunTenGodLabel(t *testing.T) {
	// 甲日主见庚→七杀运
	if got := daYunTenGodLabel(ganzhi.GanJia, ganzhi.GanGeng); got != "七杀运" {
		t.Errorf("甲见庚=%q, want 七杀运", got)
	}
	// 甲日主见甲→比肩运
	if got := daYunTenGodLabel(ganzhi.GanJia, ganzhi.GanJia); got != "比肩运" {
		t.Errorf("甲见甲=%q, want 比肩运", got)
	}
	// 甲日主见乙→劫财运
	if got := daYunTenGodLabel(ganzhi.GanJia, ganzhi.GanYi); got != "劫财运" {
		t.Errorf("甲见乙=%q, want 劫财运", got)
	}
	// 甲日主见己→正财运
	if got := daYunTenGodLabel(ganzhi.GanJia, ganzhi.GanJi); got != "正财运" {
		t.Errorf("甲见己=%q, want 正财运", got)
	}
	// 甲日主见癸→正印运
	if got := daYunTenGodLabel(ganzhi.GanJia, ganzhi.GanGui); got != "正印运" {
		t.Errorf("甲见癸=%q, want 正印运", got)
	}
}

// =============================================================================
// computeFuYinFanYin — 伏吟反吟
// =============================================================================

func TestComputeFuYinFanYin_FuYin(t *testing.T) {
	// 流年与八字某柱完全相同 → 伏吟
	bz := ganzhi.Bazi{
		Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi},
		Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiYin},
		Ri:   ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiWu},
		Shi:  ganzhi.Zhu{Gan: ganzhi.GanGeng, Zhi: ganzhi.ZhiShen},
	}
	// Same as year pillar → 伏吟
	flow := ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi}
	entries := computeFuYinFanYin(flow, bz)
	found := false
	for _, e := range entries {
		if e.Type == "伏吟" && e.NatalIndex == 0 {
			found = true
			break
		}
	}
	if !found {
		t.Error("甲子 vs 甲子年柱: expected 伏吟")
	}
}

func TestComputeFuYinFanYin_FanYin(t *testing.T) {
	// 反吟: 天克地冲 (stem clash + branch clash)
	// 甲子(年) vs 庚午(流): 甲庚克 + 子午冲 → 反吟
	bz := ganzhi.Bazi{
		Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi},
		Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiYin},
		Ri:   ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiWu},
		Shi:  ganzhi.Zhu{Gan: ganzhi.GanGeng, Zhi: ganzhi.ZhiShen},
	}
	flow := ganzhi.Zhu{Gan: ganzhi.GanGeng, Zhi: ganzhi.ZhiWu} // 庚午
	entries := computeFuYinFanYin(flow, bz)
	found := false
	for _, e := range entries {
		if e.Type == "反吟" && e.NatalIndex == 0 {
			found = true
			break
		}
	}
	if !found {
		t.Error("庚午 vs 甲子年柱: expected 反吟(天克地冲)")
	}
}

func TestComputeFuYinFanYin_DiZhiFuYin(t *testing.T) {
	// 地支相同但天干不同 → 地支伏吟
	bz := ganzhi.Bazi{
		Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi},
		Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiYin},
		Ri:   ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiWu},
		Shi:  ganzhi.Zhu{Gan: ganzhi.GanGeng, Zhi: ganzhi.ZhiShen},
	}
	flow := ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiZi} // 丙子
	entries := computeFuYinFanYin(flow, bz)
	found := false
	for _, e := range entries {
		if e.Type == "伏吟" && e.NatalIndex == 0 {
			found = true
			break
		}
	}
	if !found {
		t.Error("丙子 vs 甲子年柱: expected 地支伏吟")
	}
}

// =============================================================================
// computeDayMasterStrength — 旺衰边界组合
// =============================================================================

func TestComputeDayMasterStrength_DiverseCases(t *testing.T) {
	tests := []struct {
		name         string
		dayMaster    ganzhi.Gan
		monthBranch  ganzhi.Zhi
		wuxingCount  map[ganzhi.Wuxing]int
		hiddenStems  [4]hiddenStemsOut
		want         strength
	}{
		// 甲木日主，寅月(木旺)，木多 → 身强
		{"甲木寅月木旺→身强", ganzhi.GanJia, ganzhi.ZhiYin,
			map[ganzhi.Wuxing]int{ganzhi.WxMu: 4, ganzhi.WxHuo: 1, ganzhi.WxTu: 1, ganzhi.WxJin: 0, ganzhi.WxShui: 1},
			[4]hiddenStemsOut{
				{Main: ganzhi.GanJia}, {Main: ganzhi.GanBing}, {Main: ganzhi.GanWu}, {Main: ganzhi.GanGeng},
			},
			strengthStrong},
		// 甲木日主，申月(金旺)，金多克木 → 身弱
		{"甲木申月金克→身弱", ganzhi.GanJia, ganzhi.ZhiShen,
			map[ganzhi.Wuxing]int{ganzhi.WxMu: 1, ganzhi.WxHuo: 0, ganzhi.WxTu: 1, ganzhi.WxJin: 3, ganzhi.WxShui: 2},
			[4]hiddenStemsOut{
				{Main: ganzhi.GanBing}, {Main: ganzhi.GanBing}, {Main: ganzhi.GanWu}, {Main: ganzhi.GanGeng},
			},
			strengthWeak},
		// 丙火日主，午月(火旺)，火得令 → 身强
		{"丙火午月火旺→身强", ganzhi.GanBing, ganzhi.ZhiWu,
			map[ganzhi.Wuxing]int{ganzhi.WxMu: 2, ganzhi.WxHuo: 4, ganzhi.WxTu: 1, ganzhi.WxJin: 0, ganzhi.WxShui: 0},
			[4]hiddenStemsOut{
				{Main: ganzhi.GanJia}, {Main: ganzhi.GanBing}, {Main: ganzhi.GanBing}, {Main: ganzhi.GanWu},
			},
			strengthStrong},
		// 庚金日主，子月(水旺), 金少 → 身弱
		{"庚金子月水泄→身弱", ganzhi.GanGeng, ganzhi.ZhiZi,
			map[ganzhi.Wuxing]int{ganzhi.WxMu: 1, ganzhi.WxHuo: 0, ganzhi.WxTu: 1, ganzhi.WxJin: 2, ganzhi.WxShui: 3},
			[4]hiddenStemsOut{
				{Main: ganzhi.GanRen}, {Main: ganzhi.GanJia}, {Main: ganzhi.GanWu}, {Main: ganzhi.GanRen},
			},
			strengthWeak},
		// 日主元素=月令，且有根 → 极强
		{"甲木卯月+通根→身强", ganzhi.GanJia, ganzhi.ZhiMao,
			map[ganzhi.Wuxing]int{ganzhi.WxMu: 3, ganzhi.WxHuo: 1, ganzhi.WxTu: 1, ganzhi.WxJin: 1, ganzhi.WxShui: 1},
			[4]hiddenStemsOut{
				{Main: ganzhi.GanJia}, // 通根本气+2
				{Main: ganzhi.GanBing},
				{Main: ganzhi.GanWu},
				{Main: ganzhi.GanGeng},
			},
			strengthStrong},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeDayMasterStrength(tt.wuxingCount, tt.dayMaster, tt.monthBranch, tt.hiddenStems)
			if got != tt.want {
				t.Errorf("strength=%s, want %s", got, tt.want)
			}
		})
	}
}

// =============================================================================
// computeChart DaYun direction — 完整排盘大运方向验证
// =============================================================================

func TestComputeChart_DaYunDirection(t *testing.T) {
	tests := []struct {
		name    string
		year    int
		month   int
		day     int
		gender  ganzhi.Gender
		wantDir string
	}{
		// 1984=甲子年(阳年) → 男顺/女逆
		{"甲子年男→顺排", 1984, 2, 15, ganzhi.Male, "顺排"},
		{"甲子年女→逆排", 1984, 2, 15, ganzhi.Female, "逆排"},
		// 1985=乙丑年(阴年) → 男逆/女顺
		{"乙丑年男→逆排", 1985, 6, 15, ganzhi.Male, "逆排"},
		{"乙丑年女→顺排", 1985, 6, 15, ganzhi.Female, "顺排"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := tianwen.ComputeSolarTime(tt.year, tt.month, tt.day, 12, 0, 120, 8)
			c := ComputeChart(st, tt.gender)
			if c.DaYun.Direction != tt.wantDir {
				t.Errorf("DaYun.Direction=%q, want %q", c.DaYun.Direction, tt.wantDir)
			}
			if c.DaYun.StartAge < 0 || c.DaYun.StartAge > 12 {
				t.Errorf("StartAge=%d, expected [0,12]", c.DaYun.StartAge)
			}
		})
	}
}

// =============================================================================
// 十神 table completeness — 每柱十神完整
// =============================================================================

// =============================================================================
// Hidden stems — 藏干 domain knowledge verification
// =============================================================================

func TestComputeChart_HiddenStemsConsistency(t *testing.T) {
	// 构造已知藏干的八字: 年甲子(子藏癸), 月丙午(午藏丁己),
	// 日戊辰(辰藏戊乙癸), 时庚申(申藏庚壬戊)
	bz := ganzhi.Bazi{
		Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi},
		Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiWu},
		Ri:   ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiChen},
		Shi:  ganzhi.Zhu{Gan: ganzhi.GanGeng, Zhi: ganzhi.ZhiShen},
	}
	hs := computeHiddenStems(bz)

	// 子藏癸
	if hs[0].Main != ganzhi.GanGui || hs[0].Mid != nil || hs[0].Minor != nil {
		t.Errorf("子: want main=癸, mid=nil, minor=nil; got main=%d, mid=%v, minor=%v", hs[0].Main, hs[0].Mid, hs[0].Minor)
	}
	// 午藏丁己
	if hs[1].Main != ganzhi.GanDing {
		t.Errorf("午 main: want 丁, got %d", hs[1].Main)
	}
	if hs[1].Mid == nil || *hs[1].Mid != ganzhi.GanJi {
		t.Errorf("午 mid: want 己, got %v", hs[1].Mid)
	}
	if hs[1].Minor != nil {
		t.Errorf("午 minor: want nil, got %v", hs[1].Minor)
	}
	// 辰藏戊乙癸
	if hs[2].Main != ganzhi.GanWu {
		t.Errorf("辰 main: want 戊, got %d", hs[2].Main)
	}
	if hs[2].Mid == nil || *hs[2].Mid != ganzhi.GanYi {
		t.Errorf("辰 mid: want 乙, got %v", hs[2].Mid)
	}
	if hs[2].Minor == nil || *hs[2].Minor != ganzhi.GanGui {
		t.Errorf("辰 minor: want 癸, got %v", hs[2].Minor)
	}
	// 申藏庚壬戊
	if hs[3].Main != ganzhi.GanGeng {
		t.Errorf("申 main: want 庚, got %d", hs[3].Main)
	}
	if hs[3].Mid == nil || *hs[3].Mid != ganzhi.GanRen {
		t.Errorf("申 mid: want 壬, got %v", hs[3].Mid)
	}
	if hs[3].Minor == nil || *hs[3].Minor != ganzhi.GanWu {
		t.Errorf("申 minor: want 戊, got %v", hs[3].Minor)
	}
}

// =============================================================================
// countGenRest — 生扶计数
// =============================================================================

func TestCountGenRest(t *testing.T) {
	tests := []struct {
		name     string
		elem     ganzhi.Wuxing
		dmElem   ganzhi.Wuxing
		wantGen  int
		wantRest int
	}{
		{"同元素→生", ganzhi.WxMu, ganzhi.WxMu, 1, 0},
		{"生我→生", ganzhi.WxShui, ganzhi.WxMu, 1, 0},     // 水生木
		{"我生→泄", ganzhi.WxHuo, ganzhi.WxMu, 0, 1},       // 木生火
		{"克我→泄", ganzhi.WxJin, ganzhi.WxMu, 0, 1},       // 金克木
		{"我克→生(非克非生关系)", ganzhi.WxTu, ganzhi.WxMu, 1, 0}, // 木克土，土不克木也不生木
		{"火生→土生", ganzhi.WxHuo, ganzhi.WxTu, 1, 0},     // 火生土
		{"水不克火→泄", ganzhi.WxShui, ganzhi.WxHuo, 0, 1}, // 水克火
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen, rest := countGenRest(tt.elem, tt.dmElem)
			if gen != tt.wantGen || rest != tt.wantRest {
				t.Errorf("countGenRest(%s, %s)=(%d,%d), want (%d,%d)",
					tt.elem, tt.dmElem, gen, rest, tt.wantGen, tt.wantRest)
			}
		})
	}
}

// =============================================================================
// computeGongJia — 拱夹 golden values
// =============================================================================

func TestComputeGongJia_GoldenValues(t *testing.T) {
	tests := []struct {
		name     string
		bz       ganzhi.Bazi
		wantNum  int
		wantZhis []int // expected拱 branch zhi values
	}{
		{
			// 子(1) 寅(3): gap=2 → 拱丑(2); fillers: 午(7)酉(10) are 3 apart → no gap-2
			name: "子寅拱丑",
			bz: ganzhi.Bazi{
				Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi},  // 1
				Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiYin}, // 3
				Ri:   ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiWu},    // 7 (filler)
				Shi:  ganzhi.Zhu{Gan: ganzhi.GanGeng, Zhi: ganzhi.ZhiYou}, // 10 (filler)
			},
			wantNum:  1,
			wantZhis: []int{2}, // 丑
		},
		{
			// 寅(3) 辰(5): gap=2 → 拱卯(4); fillers: 丑(2)未(8) — no gap-2 with 3 or 5
			name: "寅辰拱卯",
			bz: ganzhi.Bazi{
				Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiYin},  // 3
				Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiChen}, // 5
				Ri:   ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiChou},   // 2 (filler)
				Shi:  ganzhi.Zhu{Gan: ganzhi.GanGeng, Zhi: ganzhi.ZhiWei},  // 8 (filler)
			},
			wantNum:  1,
			wantZhis: []int{4}, // 卯
		},
		{
			// 亥(12) 丑(2): gap=2 backward → 拱子(1); fillers: 辰(5)未(8) are 3 apart
			name: "亥丑拱子",
			bz: ganzhi.Bazi{
				Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiHai},  // 12
				Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiChou}, // 2
				Ri:   ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiChen},   // 5 (filler)
				Shi:  ganzhi.Zhu{Gan: ganzhi.GanGeng, Zhi: ganzhi.ZhiWei},  // 8 (filler)
			},
			wantNum:  1,
			wantZhis: []int{1}, // 子
		},
		{
			// 酉(10) 亥(12): gap=2 → 拱戌(11); fillers: 子(1)午(7) — no gap-2 with 10 or 12
			name: "酉亥拱戌",
			bz: ganzhi.Bazi{
				Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiYou},  // 10
				Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiHai}, // 12
				Ri:   ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiZi},    // 1 (filler)
				Shi:  ganzhi.Zhu{Gan: ganzhi.GanGeng, Zhi: ganzhi.ZhiWu},  // 7 (filler)
			},
			wantNum:  1,
			wantZhis: []int{11}, // 戌
		},
		{
			// 相邻地支(子丑): gap=1 → 不拱; fillers: 辰(5)未(8) are 3 apart
			name: "子丑相邻不拱",
			bz: ganzhi.Bazi{
				Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi},   // 1
				Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiChou}, // 2
				Ri:   ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiChen},   // 5 (filler)
				Shi:  ganzhi.Zhu{Gan: ganzhi.GanGeng, Zhi: ganzhi.ZhiWei},  // 8 (filler)
			},
			wantNum:  0,
			wantZhis: nil,
		},
		{
			// 重复地支去重: 寅(3)辰(5)拱卯(4); filler: 酉(10)
			name: "重复地支去重",
			bz: ganzhi.Bazi{
				Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiYin},  // 3
				Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiYin}, // 3 dup
				Ri:   ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiChen},  // 5
				Shi:  ganzhi.Zhu{Gan: ganzhi.GanGeng, Zhi: ganzhi.ZhiYou}, // 10 (filler)
			},
			wantNum:  1,
			wantZhis: []int{4}, // 寅(3)辰(5)拱卯(4)
		},
		{
			// 多对拱: 子寅(拱丑) + 酉亥(拱戌) → 2对
			name: "多对拱",
			bz: ganzhi.Bazi{
				Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi},   // 1
				Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiYin}, // 3
				Ri:   ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiYou},   // 10
				Shi:  ganzhi.Zhu{Gan: ganzhi.GanGeng, Zhi: ganzhi.ZhiHai}, // 12
			},
			wantNum:  2,
			wantZhis: []int{2, 11}, // 丑+戌
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeGongJia(tt.bz)
			if len(got) != tt.wantNum {
				t.Errorf("Gong count=%d, want %d: %+v", len(got), tt.wantNum, got)
			}
			if tt.wantZhis != nil {
				found := make(map[int]bool)
				for _, g := range got {
					found[int(g.Zhi)] = true
				}
				for _, z := range tt.wantZhis {
					if !found[z] {
						t.Errorf("misganzhi.WSSi.String()ng Gong branch %s", ganzhi.ZhiName(ganzhi.Zhi(z)))
					}
				}
			}
		})
	}
}

// =============================================================================
// computeFullTripleHeHui — 三合三会 golden values
// =============================================================================

func TestComputeFullTripleHeHui_GoldenValues(t *testing.T) {
	tests := []struct {
		name       string
		bz         ganzhi.Bazi
		wantNum    int
		wantElems  []string // expected elements
	}{
		{
			// 申子辰水局 (branches 9,1,5)
			name: "申子辰完整三合水局",
			bz: ganzhi.Bazi{
				Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiShen}, // 申=9
				Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiZi},   // 子=1
				Ri:   ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiChen},   // 辰=5
				Shi:  ganzhi.Zhu{Gan: ganzhi.GanGeng, Zhi: ganzhi.ZhiWu},   // 午=7
			},
			wantNum:   1,
			wantElems: []string{"水"},
		},
		{
			// 亥卯未三合木局 (branches 12,4,8)
			name: "亥卯未完整三合木局",
			bz: ganzhi.Bazi{
				Nian: ganzhi.Zhu{Gan: ganzhi.GanYi, Zhi: ganzhi.ZhiHai},  // 亥=12
				Yue:  ganzhi.Zhu{Gan: ganzhi.GanDing, Zhi: ganzhi.ZhiMao}, // 卯=4
				Ri:   ganzhi.Zhu{Gan: ganzhi.GanJi, Zhi: ganzhi.ZhiWei},   // 未=8
				Shi:  ganzhi.Zhu{Gan: ganzhi.GanXin, Zhi: ganzhi.ZhiYou},  // 酉=10
			},
			wantNum:   1,
			wantElems: []string{"木"},
		},
		{
			// 寅卯辰三会木方 (branches 3,4,5)
			name: "寅卯辰完整三会木方",
			bz: ganzhi.Bazi{
				Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiYin},  // 寅=3
				Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiMao}, // 卯=4
				Ri:   ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiChen},  // 辰=5
				Shi:  ganzhi.Zhu{Gan: ganzhi.GanGeng, Zhi: ganzhi.ZhiWu},  // 午=7
			},
			wantNum:   1,
			wantElems: []string{"木"},
		},
		{
			// 只两支(半合) → 不检测 (complete only)
			name: "半合不检测",
			bz: ganzhi.Bazi{
				Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiShen}, // 申
				Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiZi},   // 子
				Ri:   ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiWu},     // 午 (unrelated)
				Shi:  ganzhi.Zhu{Gan: ganzhi.GanGeng, Zhi: ganzhi.ZhiShen}, // 申 (dup)
			},
			wantNum:   0,
			wantElems: nil,
		},
		{
			// 无合无会
			name: "无合无会",
			bz: ganzhi.Bazi{
				Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi},   // 子
				Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiWu},  // 午
				Ri:   ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiMao},   // 卯
				Shi:  ganzhi.Zhu{Gan: ganzhi.GanGeng, Zhi: ganzhi.ZhiYou}, // 酉
			},
			wantNum:   0,
			wantElems: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeFullTripleHeHui(tt.bz)
			if len(got) != tt.wantNum {
				t.Errorf("HeHui count=%d, want %d: %+v", len(got), tt.wantNum, got)
			}
			if tt.wantElems != nil {
				found := make(map[string]bool)
				for _, h := range got {
					found[h.Element] = true
				}
				for _, e := range tt.wantElems {
					if !found[e] {
						t.Errorf("misganzhi.WSSi.String()ng HeHui element %s", e)
					}
				}
			}
		})
	}
}

// =============================================================================
// computeNaYinRelations — 纳音关系 golden values
// =============================================================================

func TestComputeNaYinRelations_GoldenValues(t *testing.T) {
	tests := []struct {
		name       string
		nayins     [4]string
		wantNum    int         // 6 pairs always
		wantRelation string    // check first relation
	}{
		{
			// 海中金 + 炉中火 = 火克金 → 相克
			name:        "金火相克",
			nayins:      [4]string{"海中金", "炉中火", "大林木", "路旁土"},
			wantNum:     6,
			wantRelation: "相克", // 金 vs 火
		},
		{
			// 海中金 + 涧下水 = 金生水 → 相生
			name:        "金水相生",
			nayins:      [4]string{"海中金", "涧下水", "大林木", "路旁土"},
			wantNum:     6,
			wantRelation: "相生", // 金 vs 水
		},
		{
			// All same → all "相同"
			name:        "全部相同",
			nayins:      [4]string{"海中金", "海中金", "海中金", "海中金"},
			wantNum:     6,
			wantRelation: "相同",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeNaYinRelations(tt.nayins)
			if len(got) != tt.wantNum {
				t.Errorf("len=%d, want %d", len(got), tt.wantNum)
			}
			if len(got) > 0 && got[0].Relation != tt.wantRelation {
				t.Errorf("first relation=%q, want %q (nayins=%v)", got[0].Relation, tt.wantRelation, tt.nayins)
			}
		})
	}
}

// =============================================================================
// analyzeZhiRelation — priority: 三会 should precede 六害
// 卯辰: both 六害 AND 三会(寅卯辰), 三会更significant → should return 三会
// =============================================================================

// verify LiuHe priority (no change needed, just confirm)
func TestAnalyzeZhiRelation_LiuHeFirst(t *testing.T) {
	// 子丑六合 → 六合
	r := analyzeZhiRelation(ganzhi.ZhiZi, ganzhi.ZhiChou)
	if r.Type != relLiuHe {
		t.Errorf("子丑: got %q, want %q (六合)", r.Type, relLiuHe)
	}
}
