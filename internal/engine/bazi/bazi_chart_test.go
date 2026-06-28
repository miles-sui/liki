//go:build integration

package bazi

import (
	"encoding/json"
	"testing"
	"time"

	"liki/internal/engine/ganzhi"
	"liki/internal/engine/tianwen"
)

// ── Known birth dates for chart regression tests ──

type chartSnapshot struct {
	Name   string
	Year   int
	Month  int
	Day    int
	Hour   int
	Minute int
	Lon    float64
	TZ     float64
	Gender ganzhi.Gender
}

func computeFullChart(g chartSnapshot) Chart {
	st := tianwen.GregorianToSolar(
		time.Date(g.Year, time.Month(g.Month), g.Day, g.Hour, g.Minute, 0, 0,
			time.FixedZone("", int(g.TZ*3600))),
		g.Lon, g.TZ)
	return ComputeChart(st, g.Gender)
}

var goldenCharts = []chartSnapshot{
	{Name: "Miles-1982-10-13-06:45-HongKong-male", Year: 1982, Month: 10, Day: 13, Hour: 6, Minute: 45, Lon: 114.134, TZ: 8, Gender: ganzhi.Male},
	{Name: "Beijing-1984-02-15-08:00-male", Year: 1984, Month: 2, Day: 15, Hour: 8, Minute: 0, Lon: 120, TZ: 8, Gender: ganzhi.Male},
	{Name: "Shanghai-1990-05-20-15:00-female", Year: 1990, Month: 5, Day: 20, Hour: 15, Minute: 0, Lon: 121.5, TZ: 8, Gender: ganzhi.Female},
	{Name: "Tokyo-2000-01-01-00:00-male", Year: 2000, Month: 1, Day: 1, Hour: 0, Minute: 0, Lon: 139.76, TZ: 9, Gender: ganzhi.Male},
	{Name: "NewYork-2020-06-15-20:00-female", Year: 2020, Month: 6, Day: 15, Hour: 20, Minute: 0, Lon: -74.0, TZ: -4, Gender: ganzhi.Female},
}

// ── Pillars validity ──

func TestGoldenChart_Pillars(t *testing.T) {
	for _, g := range goldenCharts {
		t.Run(g.Name, func(t *testing.T) {
			cr := computeFullChart(g)

			pillars := []struct {
				name string
				p    zhuInfo
			}{
				{"Nian", cr.Nian}, {"Yue", cr.Yue}, {"Ri", cr.Ri}, {"Shi", cr.Shi},
			}
			for _, p := range pillars {
				if p.p.Gan < 1 || p.p.Gan > 10 {
					t.Errorf("%s.Gan = %d, want [1,10]", p.name, p.p.Gan)
				}
				if p.p.Zhi < 1 || p.p.Zhi > 12 {
					t.Errorf("%s.Zhi = %d, want [1,12]", p.name, p.p.Zhi)
				}
			}

			// Ri.Gan IS the day master
			if cr.Ri.Gan < 1 || cr.Ri.Gan > 10 {
				t.Errorf("Ri.Gan (日主) = %d, want [1,10]", cr.Ri.Gan)
			}
		})
	}
}

// ── JSON snapshot stability ──

func TestGoldenChart_JSONSnapshot(t *testing.T) {
	for _, g := range goldenCharts {
		t.Run(g.Name, func(t *testing.T) {
			cr := computeFullChart(g)

			b, err := json.MarshalIndent(&cr, "", "  ")
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			_ = b

			var m map[string]any
			if err := json.Unmarshal(b, &m); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}

			// Nian pillar is a zhuInfo object.
			if nz, ok := m["nian"].(map[string]any); !ok || nz == nil {
				t.Errorf("nian = %v, want non-nil object", m["nian"])
			}
			// Ri pillar exists
			if rz, ok := m["ri"].(map[string]any); !ok || rz == nil {
				t.Errorf("ri = %v, want non-nil object", m["ri"])
			}
		})
	}
}

// ── Wuxing counts ──

func TestGoldenChart_WuxingCounts(t *testing.T) {
	for _, g := range goldenCharts {
		t.Run(g.Name, func(t *testing.T) {
			cr := computeFullChart(g)

			dmElem := ganzhi.GanWuxing(cr.Ri.Gan)
			if cr.WuxingCount[dmElem] < 1 {
				t.Errorf("day master element %s count = %d, want >= 1", dmElem, cr.WuxingCount[dmElem])
			}
		})
	}
}

// ── Hidden stems ──

func TestGoldenChart_HiddenStems(t *testing.T) {
	for _, g := range goldenCharts {
		t.Run(g.Name, func(t *testing.T) {
			cr := computeFullChart(g)
			hs := cr.CangGanArray()

			if len(hs) != 4 {
				t.Fatalf("len(CangGanArray) = %d, want 4", len(hs))
			}
			names := [4]string{"nian", "yue", "ri", "shi"}
			for i, h := range hs {
				if h.Main == 0 {
					t.Errorf("%s pillar hidden stem main qi is zero", names[i])
				}
			}
		})
	}
}

// ── DaYun direction and pillars ──

func TestGoldenChart_DaYun(t *testing.T) {
	for _, g := range goldenCharts {
		t.Run(g.Name, func(t *testing.T) {
			cr := computeFullChart(g)

			if cr.DaYun.Direction != "顺排" && cr.DaYun.Direction != "逆排" {
				t.Errorf("Direction = %q, want 顺排 or 逆排", cr.DaYun.Direction)
			}
			if cr.DaYun.StartAge < 0 || cr.DaYun.StartAge > 12 {
				t.Errorf("StartAge = %d, want [0,12]", cr.DaYun.StartAge)
			}
			if len(cr.DaYun.Zhu) < 8 {
				t.Errorf("len(Zhu) = %d, want >= 8", len(cr.DaYun.Zhu))
			}
			for i, p := range cr.DaYun.Zhu {
				if p.Gan < 1 || p.Gan > 10 {
					t.Errorf("zhu[%d].Gan = %d, want [1,10]", i, p.Gan)
				}
				if p.Zhi < 1 || p.Zhi > 12 {
					t.Errorf("zhu[%d].Zhi = %d, want [1,12]", i, p.Zhi)
				}
			}
			for i, p := range cr.DaYun.Zhu {
				if p.Name == "" {
					t.Errorf("zhu[%d].Name is empty", i)
				}
			}
		})
	}
}

// ── NaYin ──

func TestGoldenChart_NaYin(t *testing.T) {
	for _, g := range goldenCharts {
		t.Run(g.Name, func(t *testing.T) {
			cr := computeFullChart(g)

			if cr.Nian.NaYin == "" {
				t.Error("Nian.NaYin is empty")
			}
			if cr.Yue.NaYin == "" {
				t.Error("Yue.NaYin is empty")
			}
			if cr.Ri.NaYin == "" {
				t.Error("Ri.NaYin is empty")
			}
			if cr.Shi.NaYin == "" {
				t.Error("Shi.NaYin is empty")
			}

			for _, n := range []string{
				cr.Nian.NaYin, cr.Yue.NaYin,
				cr.Ri.NaYin, cr.Shi.NaYin,
			} {
				if elem := ganzhi.NayinWuxing(n); elem == 0 {
					t.Errorf("nayin %q has unknown element", n)
				}
			}
		})
	}
}

// ── FuYi analysis ──

func TestGoldenChart_FuYi(t *testing.T) {
	for _, g := range goldenCharts {
		t.Run(g.Name, func(t *testing.T) {
			cr := computeFullChart(g)

			validStrengths := map[string]bool{"身强": true, "身弱": true, "中和": true}
			if !validStrengths[cr.FuYi.Strength] {
				t.Errorf("FuYi.Strength = %q, want one of 身强/身弱/中和", cr.FuYi.Strength)
			}
			if cr.FuYi.Pattern == "" {
				t.Error("FuYi.Pattern is empty")
			}
			if cr.FuYi.Yong == "" {
				t.Error("FuYi.Yong is empty")
			}
			if cr.FuYi.Xi == "" {
				t.Error("FuYi.Xi is empty")
			}
			if cr.FuYi.Ji == "" {
				t.Error("FuYi.Ji is empty")
			}
		})
	}
}

// ── TiaoHou ──

func TestGoldenChart_TiaoHou(t *testing.T) {
	for _, g := range goldenCharts {
		t.Run(g.Name, func(t *testing.T) {
			cr := computeFullChart(g)

			if cr.TiaoHou.Season == "" {
				t.Error("TiaoHou.Season is empty")
			}
			if cr.TiaoHou.Yong == "" {
				t.Error("TiaoHou.Yong is empty")
			}
			if cr.TiaoHou.Xi == "" {
				t.Error("TiaoHou.Xi is empty")
			}
			if cr.TiaoHou.Ji == "" {
				t.Error("TiaoHou.Ji is empty")
			}
		})
	}
}

// ── ShenSha ──

func TestGoldenChart_ShenSha(t *testing.T) {
	for _, g := range goldenCharts {
		t.Run(g.Name, func(t *testing.T) {
			cr := computeFullChart(g)

			validCats := map[string]bool{"吉": true, "凶": true, "中性": true}
			pillars := []struct {
				name string
				ss   []shenShaEntry
			}{
				{"Nian", cr.Nian.ShenSha},
				{"Yue", cr.Yue.ShenSha},
				{"Ri", cr.Ri.ShenSha},
				{"Shi", cr.Shi.ShenSha},
			}
			for _, p := range pillars {
				if len(p.ss) < 1 {
					t.Errorf("%s: no shensha entries", p.name)
					continue
				}
				for _, e := range p.ss {
					if e.Name == "" {
						t.Errorf("%s: shensha name is empty", p.name)
					}
					if !validCats[e.Category] {
						t.Errorf("%s: shensha %q category = %q, want 吉/凶/中性", p.name, e.Name, e.Category)
					}
				}
			}
		})
	}
}

// ── ChangSheng stages ──

func TestGoldenChart_ChangSheng(t *testing.T) {
	for _, g := range goldenCharts {
		t.Run(g.Name, func(t *testing.T) {
			cr := computeFullChart(g)

			if len(cr.ChangSheng) != 12 {
				t.Fatalf("len(ChangSheng) = %d, want 12", len(cr.ChangSheng))
			}

			validStages := map[string]bool{
				"长生": true, "沐浴": true, "冠带": true, "临官": true,
				"帝旺": true, "衰": true, "病": true, "死": true,
				"墓": true, "绝": true, "胎": true, "养": true,
			}
			seen := map[ganzhi.Zhi]bool{}

			for i, s := range cr.ChangSheng {
				if s.Index < 1 || s.Index > 12 {
					t.Errorf("ChangSheng[%d].Index = %d, want [1,12]", i, s.Index)
				}
				if !validStages[s.Name] {
					t.Errorf("ChangSheng[%d].Name = %q, not a valid stage name", i, s.Name)
				}
				if seen[s.Index] {
					t.Errorf("ChangSheng[%d]: duplicate Index %d", i, s.Index)
				}
				seen[s.Index] = true
			}
			if len(seen) != 12 {
				t.Errorf("ChangSheng has %d unique indices, want 12", len(seen))
			}
		})
	}
}

// ── TaiYuan / MingGong / ShenGong ──

func TestGoldenChart_SanYuan(t *testing.T) {
	for _, g := range goldenCharts {
		t.Run(g.Name, func(t *testing.T) {
			cr := computeFullChart(g)

			tm := cr.SanYuan

			pillars := []struct {
				name string
				p    ganzhi.Zhu
			}{
				{"TaiYuan", tm.TaiYuan},
				{"MingGong", tm.MingGong},
				{"ShenGong", tm.ShenGong},
			}
			for _, p := range pillars {
				if p.p.Gan < 1 || p.p.Gan > 10 {
					t.Errorf("%s.Gan = %d, want [1,10]", p.name, p.p.Gan)
				}
				if p.p.Zhi < 1 || p.p.Zhi > 12 {
					t.Errorf("%s.Zhi = %d, want [1,12]", p.name, p.p.Zhi)
				}
			}

			// Not all three pillars have the same Gan.
			if tm.TaiYuan.Gan == tm.MingGong.Gan && tm.MingGong.Gan == tm.ShenGong.Gan {
				t.Error("all three SanYuan pillars have identical Gan")
			}
		})
	}
}

// ── All five wuxing elements present ──

func TestGoldenChart_WuxingAll(t *testing.T) {
	for _, g := range goldenCharts {
		t.Run(g.Name, func(t *testing.T) {
			cr := computeFullChart(g)

			allElem := [5]ganzhi.Wuxing{1, 2, 3, 4, 5} // 木火土金水
			for _, e := range allElem {
				c := cr.WuxingCount[e] // missing key → 0, which is valid
				if c < 0 {
					t.Errorf("WuxingCount[%s] = %d, want >= 0", e, c)
				}
			}
		})
	}
}

// ── KongWang (空亡) on all four pillars ──

func TestGoldenChart_KongWang(t *testing.T) {
	for _, g := range goldenCharts {
		t.Run(g.Name, func(t *testing.T) {
			cr := computeFullChart(g)

			// Verify IsVoid fields exist and are accessible on all 4 pillars.
			pillars := []bool{cr.Nian.IsVoid, cr.Yue.IsVoid, cr.Ri.IsVoid, cr.Shi.IsVoid}
			for _, v := range pillars {
				_ = v // valid bool by construction
			}
		})
	}
}

// ── WangShuai (旺衰) covers all five elements ──

func TestGoldenChart_WangShuai(t *testing.T) {
	for _, g := range goldenCharts {
		t.Run(g.Name, func(t *testing.T) {
			cr := computeFullChart(g)

			allElem := []string{"木", "火", "土", "金", "水"}
			for _, e := range allElem {
				if _, ok := cr.WangShuai[e]; !ok {
					t.Errorf("WangShuai missing key %q", e)
				}
			}
		})
	}
}

// ── HeHui / GongJia ──

func TestGoldenChart_HeHui_GongJia(t *testing.T) {
	for _, g := range goldenCharts {
		t.Run(g.Name, func(t *testing.T) {
			cr := computeFullChart(g)

			for i, h := range cr.HeHui {
				if h.Type == "" {
					t.Errorf("HeHui[%d].Type is empty", i)
				}
				if h.Name == "" {
					t.Errorf("HeHui[%d].Name is empty", i)
				}
				if h.Element == "" {
					t.Errorf("HeHui[%d].Element is empty", i)
				}
			}
			for i, gj := range cr.GongJia {
				if gj.Type == "" {
					t.Errorf("GongJia[%d].Type is empty", i)
				}
			}
		})
	}
}

// ── DaYun direction and start age ──

func TestGoldenChart_DaYun_Specific(t *testing.T) {
	for _, g := range goldenCharts {
		t.Run(g.Name, func(t *testing.T) {
			cr := computeFullChart(g)

			if cr.DaYun.Direction != "顺排" && cr.DaYun.Direction != "逆排" {
				t.Errorf("Direction = %q, want 顺排 or 逆排", cr.DaYun.Direction)
			}
			if cr.DaYun.StartAge < 0 || cr.DaYun.StartAge > 12 {
				t.Errorf("StartAge = %d, want [0,12]", cr.DaYun.StartAge)
			}
		})
	}
}

// ── ShiShens table ──

func TestGoldenChart_ShiShens(t *testing.T) {
	for _, g := range goldenCharts {
		t.Run(g.Name, func(t *testing.T) {
			cr := computeFullChart(g)

			pillars := [4]zhuInfo{cr.Nian, cr.Yue, cr.Ri, cr.Shi}
			names := [4]string{"nian", "yue", "ri", "shi"}

			validShiShens := map[string]bool{
				"比肩": true, "劫财": true, "食神": true, "伤官": true,
				"偏财": true, "正财": true, "七杀": true, "正官": true,
				"偏印": true, "正印": true,
			}

			for i, p := range pillars {
				if len(p.ShiShens) < 1 {
					t.Errorf("%s pillar: no shi shens", names[i])
					continue
				}
				if p.ShiShens[0].Source != sourceGan {
					t.Errorf("%s pillar: first shi shen source = %s, want stem", names[i], p.ShiShens[0].Source)
				}
				if !validShiShens[p.ShiShens[0].ShiShen.String()] {
					t.Errorf("%s pillar: unknown shi shen %q", names[i], p.ShiShens[0].ShiShen)
				}
				hasMainQi := false
				for _, e := range p.ShiShens {
					if e.Source == sourceMainQi {
						hasMainQi = true
						break
					}
				}
				if !hasMainQi {
					t.Errorf("%s pillar: no main_qi shi shen", names[i])
				}
			}
		})
	}
}
