package bazi

import (
	"testing"

	"liki/internal/engine/ganzhi"
)

// ── 干支交互准确性测试 ──
// 全部基于命理口诀独立验证。

// ── 天干五合 ──
// 口诀：甲己合化土，乙庚合化金，丙辛合化水，丁壬合化木，戊癸合化火

func TestInteraction_GanHe(t *testing.T) {
	tests := []struct {
		name   string
		a, b   ganzhi.Gan
		wantHe bool
	}{
		{"甲己合", ganzhi.GanJia, ganzhi.GanJi, true},
		{"乙庚合", ganzhi.GanYi, ganzhi.GanGeng, true},
		{"丙辛合", ganzhi.GanBing, ganzhi.GanXin, true},
		{"丁壬合", ganzhi.GanDing, ganzhi.GanRen, true},
		{"戊癸合", ganzhi.GanWu, ganzhi.GanGui, true},
		// 反向
		{"己甲合", ganzhi.GanJi, ganzhi.GanJia, true},
		{"庚乙合", ganzhi.GanGeng, ganzhi.GanYi, true},
		// 不合
		{"甲庚不合", ganzhi.GanJia, ganzhi.GanGeng, false},
		{"乙丙不合", ganzhi.GanYi, ganzhi.GanBing, false},
		{"丙丁不合", ganzhi.GanBing, ganzhi.GanDing, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ganzhi.IsGanHe(tt.a, tt.b)
			if got != tt.wantHe {
				t.Errorf("IsGanHe(%s,%s) = %v, want %v",
					ganzhi.GanName(tt.a), ganzhi.GanName(tt.b), got, tt.wantHe)
			}
		})
	}
}

// ── 地支六合 ──
// 口诀：子丑合土，寅亥合木，卯戌合火，辰酉合金，巳申合水，午未合日月

func TestInteraction_ZhiLiuHe(t *testing.T) {
	tests := []struct {
		name   string
		a, b   ganzhi.Zhi
		wantHe bool
	}{
		{"子丑合", ganzhi.ZhiZi, ganzhi.ZhiChou, true},
		{"寅亥合", ganzhi.ZhiYin, ganzhi.ZhiHai, true},
		{"卯戌合", ganzhi.ZhiMao, ganzhi.ZhiXu, true},
		{"辰酉合", ganzhi.ZhiChen, ganzhi.ZhiYou, true},
		{"巳申合", ganzhi.ZhiSi, ganzhi.ZhiShen, true},
		{"午未合", ganzhi.ZhiWu, ganzhi.ZhiWei, true},
		// 反向
		{"丑子合", ganzhi.ZhiChou, ganzhi.ZhiZi, true},
		// 不合
		{"子午不合", ganzhi.ZhiZi, ganzhi.ZhiWu, false},
		{"寅申不合", ganzhi.ZhiYin, ganzhi.ZhiShen, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ganzhi.IsZhiHe(tt.a, tt.b)
			if got != tt.wantHe {
				t.Errorf("IsZhiHe(%s,%s) = %v, want %v",
					ganzhi.ZhiName(tt.a), ganzhi.ZhiName(tt.b), got, tt.wantHe)
			}
		})
	}
}

// ── 地支六冲 ──
// 口诀：子午冲，丑未冲，寅申冲，卯酉冲，辰戌冲，巳亥冲

func TestInteraction_LiuChong(t *testing.T) {
	tests := []struct {
		name     string
		a, b     ganzhi.Zhi
		wantChong bool
	}{
		{"子午冲", ganzhi.ZhiZi, ganzhi.ZhiWu, true},
		{"丑未冲", ganzhi.ZhiChou, ganzhi.ZhiWei, true},
		{"寅申冲", ganzhi.ZhiYin, ganzhi.ZhiShen, true},
		{"卯酉冲", ganzhi.ZhiMao, ganzhi.ZhiYou, true},
		{"辰戌冲", ganzhi.ZhiChen, ganzhi.ZhiXu, true},
		{"巳亥冲", ganzhi.ZhiSi, ganzhi.ZhiHai, true},
		// 反向
		{"午子冲", ganzhi.ZhiWu, ganzhi.ZhiZi, true},
		// 不冲
		{"子丑不冲", ganzhi.ZhiZi, ganzhi.ZhiChou, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ganzhi.IsLiuChong(tt.a, tt.b)
			if got != tt.wantChong {
				t.Errorf("IsLiuChong(%s,%s) = %v, want %v",
					ganzhi.ZhiName(tt.a), ganzhi.ZhiName(tt.b), got, tt.wantChong)
			}
		})
	}
}

// ── 地支六害 ──
// 口诀：子未害，丑午害，寅巳害，卯辰害，申亥害，酉戌害

func TestInteraction_LiuHai(t *testing.T) {
	tests := []struct {
		name    string
		a, b    ganzhi.Zhi
		wantHai bool
	}{
		{"子未害", ganzhi.ZhiZi, ganzhi.ZhiWei, true},
		{"丑午害", ganzhi.ZhiChou, ganzhi.ZhiWu, true},
		{"寅巳害", ganzhi.ZhiYin, ganzhi.ZhiSi, true},
		{"卯辰害", ganzhi.ZhiMao, ganzhi.ZhiChen, true},
		{"申亥害", ganzhi.ZhiShen, ganzhi.ZhiHai, true},
		{"酉戌害", ganzhi.ZhiYou, ganzhi.ZhiXu, true},
		// 不害
		{"子午不害", ganzhi.ZhiZi, ganzhi.ZhiWu, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ganzhi.IsHai(tt.a, tt.b)
			if got != tt.wantHai {
				t.Errorf("IsHai(%s,%s) = %v, want %v",
					ganzhi.ZhiName(tt.a), ganzhi.ZhiName(tt.b), got, tt.wantHai)
			}
		})
	}
}

// ── 地支相刑 ──
// 无礼之刑：子卯  无恩之刑：寅巳申  恃势之刑：丑戌未  自刑：辰午酉亥

func TestInteraction_Xing(t *testing.T) {
	tests := []struct {
		name     string
		a, b     ganzhi.Zhi
		wantXing bool
	}{
		// 无礼之刑
		{"子卯刑", ganzhi.ZhiZi, ganzhi.ZhiMao, true},
		{"卯子刑", ganzhi.ZhiMao, ganzhi.ZhiZi, true},
		// 无恩之刑
		{"寅巳刑", ganzhi.ZhiYin, ganzhi.ZhiSi, true},
		{"巳申刑", ganzhi.ZhiSi, ganzhi.ZhiShen, true},
		{"申寅刑", ganzhi.ZhiShen, ganzhi.ZhiYin, true},
		// 恃势之刑
		{"丑戌刑", ganzhi.ZhiChou, ganzhi.ZhiXu, true},
		{"戌未刑", ganzhi.ZhiXu, ganzhi.ZhiWei, true},
		{"未丑刑", ganzhi.ZhiWei, ganzhi.ZhiChou, true},
		// 自刑
		{"辰辰自刑", ganzhi.ZhiChen, ganzhi.ZhiChen, true},
		{"午午自刑", ganzhi.ZhiWu, ganzhi.ZhiWu, true},
		{"酉酉自刑", ganzhi.ZhiYou, ganzhi.ZhiYou, true},
		{"亥亥自刑", ganzhi.ZhiHai, ganzhi.ZhiHai, true},
		// 不刑
		{"子丑不刑", ganzhi.ZhiZi, ganzhi.ZhiChou, false},
		{"寅卯不刑", ganzhi.ZhiYin, ganzhi.ZhiMao, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ganzhi.IsXing(tt.a, tt.b)
			if got != tt.wantXing {
				t.Errorf("IsXing(%s,%s) = %v, want %v",
					ganzhi.ZhiName(tt.a), ganzhi.ZhiName(tt.b), got, tt.wantXing)
			}
		})
	}
}

// ── 暗合 ──
// 口诀：寅丑暗合，卯申暗合，午亥暗合，子戌暗合

func TestInteraction_AnHe(t *testing.T) {
	tests := []struct {
		name     string
		a, b     ganzhi.Zhi
		wantAnHe bool
	}{
		{"寅丑暗合", ganzhi.ZhiYin, ganzhi.ZhiChou, true},
		{"卯申暗合", ganzhi.ZhiMao, ganzhi.ZhiShen, true},
		{"午亥暗合", ganzhi.ZhiWu, ganzhi.ZhiHai, true},
		{"子戌暗合", ganzhi.ZhiZi, ganzhi.ZhiXu, true},
		// 不暗合
		{"子丑不暗合", ganzhi.ZhiZi, ganzhi.ZhiChou, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ganzhi.IsAnHe(tt.a, tt.b)
			if got != tt.wantAnHe {
				t.Errorf("IsAnHe(%s,%s) = %v, want %v",
					ganzhi.ZhiName(tt.a), ganzhi.ZhiName(tt.b), got, tt.wantAnHe)
			}
		})
	}
}

// ── 相破 ──
// 口诀：子酉破，寅亥破，辰丑破，午卯破，申巳破，戌未破

func TestInteraction_Po(t *testing.T) {
	tests := []struct {
		name   string
		a, b   ganzhi.Zhi
		wantPo bool
	}{
		{"子酉破", ganzhi.ZhiZi, ganzhi.ZhiYou, true},
		{"寅亥破", ganzhi.ZhiYin, ganzhi.ZhiHai, true},
		{"辰丑破", ganzhi.ZhiChen, ganzhi.ZhiChou, true},
		{"午卯破", ganzhi.ZhiWu, ganzhi.ZhiMao, true},
		{"申巳破", ganzhi.ZhiShen, ganzhi.ZhiSi, true},
		{"戌未破", ganzhi.ZhiXu, ganzhi.ZhiWei, true},
		// 不破
		{"子丑不破", ganzhi.ZhiZi, ganzhi.ZhiChou, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ganzhi.IsPo(tt.a, tt.b)
			if got != tt.wantPo {
				t.Errorf("IsPo(%s,%s) = %v, want %v",
					ganzhi.ZhiName(tt.a), ganzhi.ZhiName(tt.b), got, tt.wantPo)
			}
		})
	}
}

// ── 三合 / 三会检测 ──

func TestInteraction_TripleHe(t *testing.T) {
	tests := []struct {
		name      string
		a, b      ganzhi.Zhi
		wantTriHe bool
	}{
		// 申子辰水局
		{"申子三合", ganzhi.ZhiShen, ganzhi.ZhiZi, true},
		{"子辰三合", ganzhi.ZhiZi, ganzhi.ZhiChen, true},
		{"申辰三合", ganzhi.ZhiShen, ganzhi.ZhiChen, true},
		// 寅午戌火局
		{"寅午三合", ganzhi.ZhiYin, ganzhi.ZhiWu, true},
		// 巳酉丑金局
		{"巳酉三合", ganzhi.ZhiSi, ganzhi.ZhiYou, true},
		// 亥卯未木局
		{"亥卯三合", ganzhi.ZhiHai, ganzhi.ZhiMao, true},
		// 不三合
		{"子午不三合", ganzhi.ZhiZi, ganzhi.ZhiWu, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ganzhi.IsTripleHe(tt.a, tt.b)
			if got != tt.wantTriHe {
				t.Errorf("IsTripleHe(%s,%s) = %v, want %v",
					ganzhi.ZhiName(tt.a), ganzhi.ZhiName(tt.b), got, tt.wantTriHe)
			}
		})
	}
}

func TestInteraction_TripleHui(t *testing.T) {
	tests := []struct {
		name      string
		a, b      ganzhi.Zhi
		wantHui bool
	}{
		// 寅卯辰会木
		{"寅卯三会", ganzhi.ZhiYin, ganzhi.ZhiMao, true},
		{"卯辰三会", ganzhi.ZhiMao, ganzhi.ZhiChen, true},
		// 巳午未会火
		{"巳午三会", ganzhi.ZhiSi, ganzhi.ZhiWu, true},
		// 申酉戌会金
		{"申酉三会", ganzhi.ZhiShen, ganzhi.ZhiYou, true},
		// 亥子丑会水
		{"亥子三会", ganzhi.ZhiHai, ganzhi.ZhiZi, true},
		// 不三会
		{"子寅不三会", ganzhi.ZhiZi, ganzhi.ZhiYin, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ganzhi.IsTripleHui(tt.a, tt.b)
			if got != tt.wantHui {
				t.Errorf("IsTripleHui(%s,%s) = %v, want %v",
					ganzhi.ZhiName(tt.a), ganzhi.ZhiName(tt.b), got, tt.wantHui)
			}
		})
	}
}

// ── analyzeGanRelation 综合测试 ──
// 验证优先级: 五合 > 同气 > 生克

func TestInteraction_AnalyzeGanRelation(t *testing.T) {
	tests := []struct {
		name       string
		a, b       ganzhi.Gan
		wantType   string
	}{
		{"甲己→五合", ganzhi.GanJia, ganzhi.GanJi, relGanHe},
		{"乙庚→五合", ganzhi.GanYi, ganzhi.GanGeng, relGanHe},
		{"甲甲→相同", ganzhi.GanJia, ganzhi.GanJia, relSame},
		{"甲乙→同气(木)", ganzhi.GanJia, ganzhi.GanYi, relSame},
		{"甲丙→相生(木生火)", ganzhi.GanJia, ganzhi.GanBing, relSheng},
		{"丙甲→相生(木生火)", ganzhi.GanBing, ganzhi.GanJia, relSheng},
		{"甲戊→相克(木克土)", ganzhi.GanJia, ganzhi.GanWu, relKe},
		{"戊甲→相克(木克土)", ganzhi.GanWu, ganzhi.GanJia, relKe},
		// 庚甲 → 庚(金)克甲(木) → 相克
		{"庚甲→相克", ganzhi.GanGeng, ganzhi.GanJia, relKe},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := analyzeGanRelation(tt.a, tt.b)
			if r.Type != tt.wantType {
				t.Errorf("analyzeGanRelation(%s,%s).Type = %q, want %q",
					ganzhi.GanName(tt.a), ganzhi.GanName(tt.b), r.Type, tt.wantType)
			}
		})
	}
}

// ── analyzeZhiRelation 综合测试 ──
// 验证优先级: 六合 > 三合 > 三会 > 六冲 > 相刑 > 六害 > 暗合 > 破

func TestInteraction_AnalyzeZhiRelation(t *testing.T) {
	tests := []struct {
		name     string
		a, b     ganzhi.Zhi
		wantType string
	}{
		// 相同
		{"子子→相同", ganzhi.ZhiZi, ganzhi.ZhiZi, relSame},
		// 六合 > 三合
		{"子丑→六合", ganzhi.ZhiZi, ganzhi.ZhiChou, relLiuHe},
		// 三合（子辰既是三合也是三会? 不是，亥子丑才是三会。子辰是申子辰三合）
		{"子辰→三合", ganzhi.ZhiZi, ganzhi.ZhiChen, relSanHe},
		// 三会
		{"亥子→三会", ganzhi.ZhiHai, ganzhi.ZhiZi, relSanHui},
		// 六冲（子午不是六合也不是三合三会，直接到六冲）
		{"子午→六冲", ganzhi.ZhiZi, ganzhi.ZhiWu, relLiuChong},
		// 相刑（卯辰: 卯辰六害 > 但卯刑子，卯辰没有刑。卯辰是六害）
		// 卯辰：既是六害也在三会寅卯辰组，六害是完整pair关系优先
		{"卯辰→六害", ganzhi.ZhiMao, ganzhi.ZhiChen, relLiuHai},
		// 酉戌：既是六害也在三会申酉戌组，六害是完整pair关系优先
		{"酉戌→六害", ganzhi.ZhiYou, ganzhi.ZhiXu, relLiuHai},
		// 相刑: 子卯
		// 子卯: 不是六合，不是三合，不是三会，不是六冲 → 相刑
		{"子卯→相刑", ganzhi.ZhiZi, ganzhi.ZhiMao, relXing},
		// 六害: 子未
		// 子未: 不是六合，不是三合，不是三会，不是六冲，不是相刑 → 六害
		{"子未→六害", ganzhi.ZhiZi, ganzhi.ZhiWei, relLiuHai},
		// 暗合: 寅丑
		// 寅丑: 不是六合，不是三合，不是三会，不是六冲，不是相刑，不是六害 → 暗合
		{"寅丑→暗合", ganzhi.ZhiYin, ganzhi.ZhiChou, relAnHe},
		// 破: 子酉
		// 子酉: 不是六合，不是三合/三会，不是六冲，不是相刑，不是六害，不是暗合 → 破
		{"子酉→破", ganzhi.ZhiZi, ganzhi.ZhiYou, relPo},
		// 无关系
		{"寅酉→无", ganzhi.ZhiYin, ganzhi.ZhiYou, relNone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := analyzeZhiRelation(tt.a, tt.b)
			if r.Type != tt.wantType {
				t.Errorf("analyzeZhiRelation(%s,%s).Type = %q, want %q",
					ganzhi.ZhiName(tt.a), ganzhi.ZhiName(tt.b), r.Type, tt.wantType)
			}
		})
	}
}

// ── 拱夹检测 ──

func TestGongJia(t *testing.T) {
	// 拱：两支相差2 → 中间地支为拱
	// 寅辰拱卯，巳未拱午，申戌拱酉，亥丑拱子
	tests := []struct {
		name       string
		branches   [4]ganzhi.Zhi
		wantGong   bool
		wantMidZhi ganzhi.Zhi
	}{
		{"寅辰拱卯", [4]ganzhi.Zhi{ganzhi.ZhiYin, ganzhi.ZhiZi, ganzhi.ZhiChen, ganzhi.ZhiWu}, true, ganzhi.ZhiMao},
		{"申戌拱酉", [4]ganzhi.Zhi{ganzhi.ZhiShen, ganzhi.ZhiZi, ganzhi.ZhiXu, ganzhi.ZhiWu}, true, ganzhi.ZhiYou},
		{"亥丑拱子(回绕)", [4]ganzhi.Zhi{ganzhi.ZhiHai, ganzhi.ZhiYin, ganzhi.ZhiChou, ganzhi.ZhiWu}, true, ganzhi.ZhiZi},
		{"无拱", [4]ganzhi.Zhi{ganzhi.ZhiZi, ganzhi.ZhiMao, ganzhi.ZhiWu, ganzhi.ZhiYou}, false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bz := ganzhi.Bazi{
				Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: tt.branches[0]},
				Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: tt.branches[1]},
				Ri:   ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: tt.branches[2]},
				Shi:  ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: tt.branches[3]},
			}
			gj := computeGongJia(bz)
			if tt.wantGong && len(gj) == 0 {
				t.Error("want 拱 but none found")
				return
			}
			if !tt.wantGong {
				if len(gj) > 0 {
					for _, g := range gj {
						t.Logf("unexpected 拱: %s", ganzhi.ZhiName(g.Zhi))
					}
					t.Error("want no 拱 but found some")
				}
				return
			}
			found := false
			for _, g := range gj {
				if g.Zhi == tt.wantMidZhi {
					found = true
					break
				}
			}
			if !found {
				for _, g := range gj {
					t.Logf("found 拱: %s", ganzhi.ZhiName(g.Zhi))
				}
				t.Errorf("拱 %s not found in results", ganzhi.ZhiName(tt.wantMidZhi))
			}
		})
	}
}

// ── 三合局完整检测 ──

func TestFullTripleHeHui(t *testing.T) {
	// 地支全申子辰 → 三合水局
	bz := ganzhi.Bazi{
		Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiShen},
		Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiZi},
		Ri:   ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiChen},
		Shi:  ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiXu},
	}
	results := computeFullTripleHeHui(bz)
	if len(results) == 0 {
		t.Fatal("want at least one 三合/三会")
	}
	found := false
	for _, r := range results {
		if r.Type == relSanHe && r.Element == "水" {
			found = true
		}
	}
	if !found {
		t.Errorf("want 三合水局, got: %+v", results)
	}
}

// ── 伏吟反吟 ──

func TestFuYinFanYin(t *testing.T) {
	// 日柱己卯，遇上己卯 → 伏吟
	bz := ganzhi.Bazi{
		Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi},
		Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiYin},
		Ri:   ganzhi.Zhu{Gan: ganzhi.GanJi, Zhi: ganzhi.ZhiMao},
		Shi:  ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiChen},
	}
	// 流年己卯 → 与日柱相同 → 伏吟
	flow := ganzhi.Zhu{Gan: ganzhi.GanJi, Zhi: ganzhi.ZhiMao}
	ff := computeFuYinFanYin(flow, bz)

	hasFuYin := false
	for _, f := range ff {
		if f.Type == "伏吟" && f.NatalIndex == 2 {
			hasFuYin = true
		}
	}
	if !hasFuYin {
		t.Errorf("己卯流年遇己卯日柱，应有伏吟。got: %+v", ff)
	}

	// 反吟：天克地冲 → 乙酉 vs 己卯 → 乙克己(木克土) + 卯酉冲
	flow2 := ganzhi.Zhu{Gan: ganzhi.GanYi, Zhi: ganzhi.ZhiYou}
	ff2 := computeFuYinFanYin(flow2, bz)

	hasFanYin := false
	for _, f := range ff2 {
		if f.Type == "反吟" && f.NatalIndex == 2 {
			hasFanYin = true
		}
	}
	if !hasFanYin {
		t.Errorf("乙酉遇己卯(天克地冲)，应有反吟。got: %+v", ff2)
	}
}

// ── 空亡 ──

func TestKongWang(t *testing.T) {
	// 甲子日 → 甲子旬(0-9), 空戌亥
	bz := ganzhi.Bazi{
		Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiXu},  // 戌=空亡
		Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiYin},
		Ri:   ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi},  // 甲子日
		Shi:  ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiChen},
	}
	hits := computeKongWang(bz)
	if len(hits) != 1 || hits[0] != 0 {
		t.Errorf("甲子日空戌亥, 年柱戌应空亡。hits=%v", hits)
	}

	// 甲寅日 → 甲寅旬(50-59), 空子丑
	bz2 := ganzhi.Bazi{
		Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi},  // 子=空亡
		Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiChou}, // 丑=空亡
		Ri:   ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiYin},  // 甲寅日
		Shi:  ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiMao},
	}
	hits2 := computeKongWang(bz2)
	if len(hits2) != 2 {
		t.Errorf("甲寅日空子丑, 年柱子+月柱丑应空亡。hits=%v", hits2)
	}
}

// ── 十神 ──

func TestTenGod(t *testing.T) {
	// 日主 = 甲(木,阳)
	// 甲见甲=比肩, 甲见乙=劫财, 甲见丙=食神, 甲见丁=伤官
	// 甲见戊=偏财, 甲见己=正财, 甲见庚=七杀, 甲见辛=正官
	// 甲见壬=偏印, 甲见癸=正印
	tests := []struct {
		name     string
		dm, other ganzhi.Gan
		want     ganzhi.TenGod
	}{
		{"甲见甲→比肩", ganzhi.GanJia, ganzhi.GanJia, ganzhi.TenGodBiJian},
		{"甲见乙→劫财", ganzhi.GanJia, ganzhi.GanYi, ganzhi.TenGodJieCai},
		{"甲见丙→食神", ganzhi.GanJia, ganzhi.GanBing, ganzhi.TenGodShiShen},
		{"甲见丁→伤官", ganzhi.GanJia, ganzhi.GanDing, ganzhi.TenGodShangGuan},
		{"甲见戊→偏财", ganzhi.GanJia, ganzhi.GanWu, ganzhi.TenGodPianCai},
		{"甲见己→正财", ganzhi.GanJia, ganzhi.GanJi, ganzhi.TenGodZhengCai},
		{"甲见庚→七杀", ganzhi.GanJia, ganzhi.GanGeng, ganzhi.TenGodQiSha},
		{"甲见辛→正官", ganzhi.GanJia, ganzhi.GanXin, ganzhi.TenGodZhengGuan},
		{"甲见壬→偏印", ganzhi.GanJia, ganzhi.GanRen, ganzhi.TenGodPianYin},
		{"甲见癸→正印", ganzhi.GanJia, ganzhi.GanGui, ganzhi.TenGodZhengYin},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ganzhi.TenGodFromGan(tt.dm, tt.other)
			if got != tt.want {
				t.Errorf("TenGodFromGan(%s日主见%s) = %q, want %q",
					ganzhi.GanName(tt.dm), ganzhi.GanName(tt.other), got, tt.want)
			}
		})
	}
}

// ── 纳音 ──

func TestNaYin(t *testing.T) {
	// 甲子乙丑=海中金, 丙寅丁卯=炉中火, 戊辰己巳=大林木
	tests := []struct {
		name string
		gan  ganzhi.Gan
		zhi  ganzhi.Zhi
		want string
	}{
		{"甲子→海中金", ganzhi.GanJia, ganzhi.ZhiZi, "海中金"},
		{"乙丑→海中金", ganzhi.GanYi, ganzhi.ZhiChou, "海中金"},
		{"丙寅→炉中火", ganzhi.GanBing, ganzhi.ZhiYin, "炉中火"},
		{"丁卯→炉中火", ganzhi.GanDing, ganzhi.ZhiMao, "炉中火"},
		{"戊辰→大林木", ganzhi.GanWu, ganzhi.ZhiChen, "大林木"},
		{"己巳→大林木", ganzhi.GanJi, ganzhi.ZhiSi, "大林木"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ganzhi.NaYinLabel(tt.gan, tt.zhi)
			if got != tt.want {
				t.Errorf("NaYinLabel(%s%s) = %q, want %q",
					ganzhi.GanName(tt.gan), ganzhi.ZhiName(tt.zhi), got, tt.want)
			}
		})
	}
}
