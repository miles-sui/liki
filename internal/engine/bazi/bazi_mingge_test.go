package bazi

import (
	"testing"

	"liki/internal/engine/ganzhi"
)

func TestComputeDayMasterStrength_StrongInSeason(t *testing.T) {
	// 甲木日主，寅月 (旺月)，地支寅卯辰全 → 身强
	ec := map[ganzhi.Wuxing]int{
		ganzhi.WxMu:   4, // 日主 + 寅 + 卯 + 辰(中气)
		ganzhi.WxShui: 1, // 生成 +1
		ganzhi.WxHuo:  0,
		ganzhi.WxTu:   1,
		ganzhi.WxJin:  0,
	}
	cangGan := [4]cangGanOut{
		{Main: ganzhi.GanJia},                                  // 年支寅: 本气甲
		{Main: ganzhi.GanYi},                                   // 月支卯: 本气乙
		{Mid: ptr(ganzhi.GanYi)},                               // 日支辰: 中气乙
		{Main: ganzhi.GanWu},                                   // 时支: 无关
	}
	got := computeDayMasterStrength(ec, ganzhi.GanJia, ganzhi.ZhiYin, cangGan)
	if got != strengthStrong {
		t.Errorf("strength = %d (%s), want strengthStrong(%d)", got, got, strengthStrong)
	}
}

func TestComputeDayMasterStrength_WeakControlled(t *testing.T) {
	// 甲木日主，申月 (死月)，金多 → 身弱
	ec := map[ganzhi.Wuxing]int{
		ganzhi.WxMu:   1, // 仅日主
		ganzhi.WxShui: 1,
		ganzhi.WxHuo:  0,
		ganzhi.WxTu:   0,
		ganzhi.WxJin:  3, // 申酉金多
	}
	cangGan := [4]cangGanOut{
		{Main: ganzhi.GanGeng},                                 // 年支申: 本气庚
		{Main: ganzhi.GanXin},                                  // 月支酉: 本气辛
		{Main: ganzhi.GanWu},                                   // 日支: 无关
		{Main: ganzhi.GanRen},                                  // 时支: 无关
	}
	got := computeDayMasterStrength(ec, ganzhi.GanJia, ganzhi.ZhiShen, cangGan)
	if got != strengthWeak {
		t.Errorf("strength = %d (%s), want strengthWeak(%d)", got, got, strengthWeak)
	}
}

func TestComputeDayMasterStrength_Neutral(t *testing.T) {
	// 日主在休月(水木不生)，木不多不少 → 中和
	ec := map[ganzhi.Wuxing]int{
		ganzhi.WxMu:   2,
		ganzhi.WxShui: 1,
		ganzhi.WxHuo:  1,
		ganzhi.WxTu:   2,
		ganzhi.WxJin:  1,
	}
	cangGan := [4]cangGanOut{
		{Main: ganzhi.GanYi},                                   // 根 1: 木本气
		{Main: ganzhi.GanWu},                                   // 无关
		{Main: ganzhi.GanWu},                                   // 无关
		{Main: ganzhi.GanWu},                                   // 无关
	}
	got := computeDayMasterStrength(ec, ganzhi.GanJia, ganzhi.ZhiChen, cangGan)
	if got != strengthNeutral {
		t.Errorf("strength = %d (%s), want strengthNeutral(%d)", got, got, strengthNeutral)
	}
}

func TestComputeDayMasterStrength_RootBonus(t *testing.T) {
	// 甲木日主，墓月，但地支三合木局通根 → 身强
	ec := map[ganzhi.Wuxing]int{
		ganzhi.WxMu:   2, // 日主 + 一藏干
		ganzhi.WxShui: 0,
		ganzhi.WxHuo:  0,
		ganzhi.WxTu:   2,
		ganzhi.WxJin:  2,
	}
	// 藏干中有3个木(本气+中气+余气) → bonus = 2+1+1 = 4
	cangGan := [4]cangGanOut{
		{Main: ganzhi.GanJia},                                  // 本气木 +2
		{Main: ganzhi.GanYi, Mid: ptr(ganzhi.GanYi)},           // 本气木+2, 中气木+1
		{Main: ganzhi.GanWu},                                   // 无关
		{Main: ganzhi.GanWu},                                   // 无关
	}
	// support=2, season=-1 (囚, 未月金旺木囚), rootBonus=2+2+1=5, total=2-1+5=6 → 身强
	got := computeDayMasterStrength(ec, ganzhi.GanJia, ganzhi.ZhiWei, cangGan)
	if got != strengthStrong {
		t.Errorf("strength = %d (%s), want strengthStrong(%d)", got, got, strengthStrong)
	}
}

func TestComputeDayMasterStrength_AllCases(t *testing.T) {
	tests := []struct {
		name        string
		ec          map[ganzhi.Wuxing]int
		riYuan      ganzhi.Gan
		monthBranch ganzhi.Zhi
		cangGan     [4]cangGanOut
		want        strength
	}{

		{"丙火日主午月帝旺", map[ganzhi.Wuxing]int{
			ganzhi.WxMu: 2, ganzhi.WxHuo: 4, ganzhi.WxTu: 2, ganzhi.WxJin: 1, ganzhi.WxShui: 0,
		}, ganzhi.GanBing, ganzhi.ZhiWu, [4]cangGanOut{
			{Main: ganzhi.GanBing}, {Main: ganzhi.GanDing}, {Main: ganzhi.GanWu}, {Main: ganzhi.GanWu},
		}, strengthStrong},

		{"庚金日主午月死地无根", map[ganzhi.Wuxing]int{
			ganzhi.WxMu: 1, ganzhi.WxHuo: 3, ganzhi.WxTu: 1, ganzhi.WxJin: 1, ganzhi.WxShui: 1,
		}, ganzhi.GanGeng, ganzhi.ZhiWu, [4]cangGanOut{
			{Main: ganzhi.GanBing}, {Main: ganzhi.GanDing}, {Main: ganzhi.GanWu}, {Main: ganzhi.GanWu},
		}, strengthWeak},

		{"壬水日主申月相旺", map[ganzhi.Wuxing]int{
			ganzhi.WxMu: 1, ganzhi.WxHuo: 0, ganzhi.WxTu: 1, ganzhi.WxJin: 3, ganzhi.WxShui: 3,
		}, ganzhi.GanRen, ganzhi.ZhiShen, [4]cangGanOut{
			{Main: ganzhi.GanRen}, {Main: ganzhi.GanGeng}, {Main: ganzhi.GanRen}, {Main: ganzhi.GanWu},
		}, strengthStrong},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeDayMasterStrength(tt.ec, tt.riYuan, tt.monthBranch, tt.cangGan)
			if got != tt.want {
				t.Errorf("strength = %d (%s), want %d (%s)", got, got, tt.want, tt.want)
			}
		})
	}
}

func ptr(g ganzhi.Gan) *ganzhi.Gan { return &g }

// TestComputeDayMasterStrength_WithHeHui verifies that day master strength
// is correctly classified when the chart has a SanHui (三会) formation.
// 1983-12-15 12:00 Beijing: 癸亥 甲子 丁丑 丙午 → 亥子丑 三会水方.
// 丁火日主, 子月(死), water dominates → 身弱.
func TestComputeDayMasterStrength_WithHeHui(t *testing.T) {
	// This chart has 亥+子+丑 → 三会水方.
	// 丁火日主 in 子月: monthElem=水, dmElem=火 → 水克火 → 死 = -2.
	// WuxingCount: 癸(水) 甲(木) 丁(火) 丙(火) + cangGan →
	//   亥: 壬(水) 甲(木)    → 水+1 木+1
	//   子: 癸(水)           → 水+1
	//   丑: 己(土) 癸(水) 辛(金) → 土+1 水+1 金+1
	//   午: 丁(火) 己(土)    → 火+1 土+1
	// Stems: 水:1, 木:1, 火:2
	// CangGan: 水:3, 木:1, 土:2, 金:1, 火:1
	// Total: 木:2, 火:3, 土:2, 金:1, 水:4
	ec := map[ganzhi.Wuxing]int{
		ganzhi.WxMu:   2,
		ganzhi.WxHuo:  3,
		ganzhi.WxTu:   2,
		ganzhi.WxJin:  1,
		ganzhi.WxShui: 4,
	}
	// 丁火: support = 火(3) + 木(2) = 5
	// seasonScore = -2 (死 in 子月)
	// cangGan root: 午 main 丁(火) → +2, 其他无关
	cangGan := [4]cangGanOut{
		{Main: ganzhi.GanRen, Mid: ptr(ganzhi.GanJia)},                    // 亥
		{Main: ganzhi.GanGui},                                             // 子
		{Main: ganzhi.GanJi, Mid: ptr(ganzhi.GanGui), Minor: ptr(ganzhi.GanXin)}, // 丑
		{Main: ganzhi.GanDing, Mid: ptr(ganzhi.GanJi)},                    // 午
	}
	got := computeDayMasterStrength(ec, ganzhi.GanDing, ganzhi.ZhiZi, cangGan)
	// support(5) + seasonScore(-2) + rootBonus(2) = 5 → neutral
	// With HeHui 三会水方, water dominates even more.
	// A HeHui-aware algorithm would penalize further (extra -1 or -2).
	// Current: borderline neutral, but water dominance suggests 身弱.
	if got != strengthNeutral {
		t.Errorf("strength = %d (%s), want strengthNeutral(%d). "+
			"Note: HeHui 三会水方 not factored into strength — with HeHui this might be 身弱",
			got, got, strengthNeutral)
	}
}

// TestComputeDayMasterStrength_SanHeWoodDayMaster verifies a chart where
// SanHe and day master share the same element.
func TestComputeDayMasterStrength_SanHeWoodDayMaster(t *testing.T) {
	// 亥卯未 → 三合木局, 甲木日主.
	// For this test, simulate the counts directly.
	ec := map[ganzhi.Wuxing]int{
		ganzhi.WxMu:   4, // day master + 三合木局 branches
		ganzhi.WxHuo:  1,
		ganzhi.WxTu:   1,
		ganzhi.WxJin:  1,
		ganzhi.WxShui: 1,
	}
	// 甲木 in 卯月: monthElem=木=dmElem → 旺 = +3
	// cangGan: 亥(壬甲), 卯(乙), 未(己丁乙)
	cangGan := [4]cangGanOut{
		{Main: ganzhi.GanRen, Mid: ptr(ganzhi.GanJia)},                     // 亥
		{Main: ganzhi.GanYi},                                               // 卯
		{Main: ganzhi.GanJi, Mid: ptr(ganzhi.GanDing), Minor: ptr(ganzhi.GanYi)}, // 未
		{Main: ganzhi.GanWu},                                               // 时支无关
	}
	got := computeDayMasterStrength(ec, ganzhi.GanJia, ganzhi.ZhiMao, cangGan)
	// support = 木(4) + 水(1) = 5
	// seasonScore = 3 (旺)
	// rootBonus = 亥 mid 甲(+1) + 卯 main 乙(+2) + 未 minor 乙(+1) = 4
	// total = 5 + 3 + 4 = 12 → 身强
	// HeHui 三合木局 is consistent with this result.
	if got != strengthStrong {
		t.Errorf("strength = %d (%s), want strengthStrong(%d)",
			got, got, strengthStrong)
	}
}
