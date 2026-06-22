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
				if elem := ganzhi.NaYinWuxing(n); elem == 0 {
					t.Errorf("nayin %q has unknown element", n)
				}
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
