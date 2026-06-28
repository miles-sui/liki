package bazi

import (
	"testing"

	"liki/internal/engine/ganzhi"
)

func TestComputeFullTripleHeHui_SanHeWater(t *testing.T) {
	// 申子辰 → 三合水局
	bz := ganzhi.Bazi{
		Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiShen}, // 甲申
		Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiZi},  // 丙子
		Ri:   ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiChen},  // 戊辰
		Shi:  ganzhi.Zhu{Gan: ganzhi.GanGeng, Zhi: ganzhi.ZhiWu},  // 庚午
	}
	got := computeFullTripleHeHui(bz)
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].Type != relSanHe {
		t.Errorf("Type = %q, want %q", got[0].Type, relSanHe)
	}
	if got[0].Element != "水" {
		t.Errorf("Element = %q, want 水", got[0].Element)
	}
}

func TestComputeFullTripleHeHui_SanHeFire(t *testing.T) {
	// 寅午戌 → 三合火局
	bz := ganzhi.Bazi{
		Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiYin},  // 甲寅
		Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiWu},  // 丙午
		Ri:   ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiXu},    // 戊戌
		Shi:  ganzhi.Zhu{Gan: ganzhi.GanGeng, Zhi: ganzhi.ZhiShen}, // 庚申
	}
	got := computeFullTripleHeHui(bz)
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].Type != relSanHe {
		t.Errorf("Type = %q, want %q", got[0].Type, relSanHe)
	}
	if got[0].Element != "火" {
		t.Errorf("Element = %q, want 火", got[0].Element)
	}
}

func TestComputeFullTripleHeHui_SanHeWood(t *testing.T) {
	// 亥卯未 → 三合木局
	bz := ganzhi.Bazi{
		Nian: ganzhi.Zhu{Gan: ganzhi.GanYi, Zhi: ganzhi.ZhiHai}, // 乙亥
		Yue:  ganzhi.Zhu{Gan: ganzhi.GanDing, Zhi: ganzhi.ZhiMao}, // 丁卯
		Ri:   ganzhi.Zhu{Gan: ganzhi.GanJi, Zhi: ganzhi.ZhiWei},  // 己未
		Shi:  ganzhi.Zhu{Gan: ganzhi.GanXin, Zhi: ganzhi.ZhiYou}, // 辛酉
	}
	got := computeFullTripleHeHui(bz)
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].Type != relSanHe {
		t.Errorf("Type = %q, want %q", got[0].Type, relSanHe)
	}
	if got[0].Element != "木" {
		t.Errorf("Element = %q, want 木", got[0].Element)
	}
}

func TestComputeFullTripleHeHui_SanHeMetal(t *testing.T) {
	// 巳酉丑 → 三合金局
	bz := ganzhi.Bazi{
		Nian: ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiSi},  // 丙巳
		Yue:  ganzhi.Zhu{Gan: ganzhi.GanJi, Zhi: ganzhi.ZhiYou},   // 己酉
		Ri:   ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiChou},  // 戊丑
		Shi:  ganzhi.Zhu{Gan: ganzhi.GanGeng, Zhi: ganzhi.ZhiShen}, // 庚申
	}
	got := computeFullTripleHeHui(bz)
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].Type != relSanHe {
		t.Errorf("Type = %q, want %q", got[0].Type, relSanHe)
	}
	if got[0].Element != "金" {
		t.Errorf("Element = %q, want 金", got[0].Element)
	}
}

func TestComputeFullTripleHeHui_SanHuiWood(t *testing.T) {
	// 寅卯辰 → 三会木方 (东方)
	bz := ganzhi.Bazi{
		Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiYin},  // 甲寅
		Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiMao},  // 丙卯
		Ri:   ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiChen},   // 戊辰
		Shi:  ganzhi.Zhu{Gan: ganzhi.GanGeng, Zhi: ganzhi.ZhiWu},   // 庚午
	}
	got := computeFullTripleHeHui(bz)
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].Type != relSanHui {
		t.Errorf("Type = %q, want %q", got[0].Type, relSanHui)
	}
	if got[0].Element != "木" {
		t.Errorf("Element = %q, want 木", got[0].Element)
	}
}

func TestComputeFullTripleHeHui_SanHuiWater(t *testing.T) {
	// 亥子丑 → 三会水方 (北方)
	bz := ganzhi.Bazi{
		Nian: ganzhi.Zhu{Gan: ganzhi.GanYi, Zhi: ganzhi.ZhiHai},  // 乙亥
		Yue:  ganzhi.Zhu{Gan: ganzhi.GanDing, Zhi: ganzhi.ZhiZi},  // 丁子
		Ri:   ganzhi.Zhu{Gan: ganzhi.GanJi, Zhi: ganzhi.ZhiChou},  // 己丑
		Shi:  ganzhi.Zhu{Gan: ganzhi.GanXin, Zhi: ganzhi.ZhiWei},  // 辛未
	}
	got := computeFullTripleHeHui(bz)
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].Type != relSanHui {
		t.Errorf("Type = %q, want %q", got[0].Type, relSanHui)
	}
	if got[0].Element != "水" {
		t.Errorf("Element = %q, want 水", got[0].Element)
	}
}

func TestComputeFullTripleHeHui_NoMatch(t *testing.T) {
	// No He or Hui pattern present.
	bz := ganzhi.Bazi{
		Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi},   // 甲子
		Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiYin},  // 丙寅
		Ri:   ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiWu},     // 戊午
		Shi:  ganzhi.Zhu{Gan: ganzhi.GanGeng, Zhi: ganzhi.ZhiShen}, // 庚申
	}
	got := computeFullTripleHeHui(bz)
	if len(got) != 0 {
		t.Errorf("len = %d, want 0", len(got))
	}
}

func TestComputeFullTripleHeHui_DualPattern(t *testing.T) {
	// 寅午戌 (三合火) + 巳午未 (三会火方) — 午 repeats in both.
	bz := ganzhi.Bazi{
		Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiYin}, // 甲寅
		Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiWu}, // 丙午
		Ri:   ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiXu},   // 戊戌
		Shi:  ganzhi.Zhu{Gan: ganzhi.GanDing, Zhi: ganzhi.ZhiSi}, // 丁巳
	}
	got := computeFullTripleHeHui(bz)
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1 (fire sanhe only, 巳午未 needs 未)", len(got))
	}
}
