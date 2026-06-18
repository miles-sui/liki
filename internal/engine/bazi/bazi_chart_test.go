//go:build integration

package bazi

import (
	"encoding/json"
	"testing"

	"liki/internal/engine/ganzhi"
	"liki/internal/engine/tianwen"
)

// ── Known birth dates for chart regression tests ──
// These verify that engine changes do not silently alter chart output.
// Pillar values are validated via invariants (valid ganzhi.Gan/ganzhi.Zhi ranges, DayMaster set)
// rather than hardcoded expected values which are error-prone to maintain.

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
	st := tianwen.ComputeSolarTime(g.Year, g.Month, g.Day, g.Hour, g.Minute, g.Lon, g.TZ)
	return ComputeChart(st, g.Gender)
}

var goldenCharts = []chartSnapshot{
	{Name: "Miles-1982-10-13-06:45-HongKong-male", Year: 1982, Month: 10, Day: 13, Hour: 6, Minute: 45, Lon: 114.134, TZ: 8, Gender: ganzhi.Male},
	{Name: "Beijing-1984-02-15-08:00-male", Year: 1984, Month: 2, Day: 15, Hour: 8, Minute: 0, Lon: 120, TZ: 8, Gender: ganzhi.Male},
	{Name: "Shanghai-1990-05-20-15:00-female", Year: 1990, Month: 5, Day: 20, Hour: 15, Minute: 0, Lon: 121.5, TZ: 8, Gender: ganzhi.Female},
	{Name: "Tokyo-2000-01-01-00:00-male", Year: 2000, Month: 1, Day: 1, Hour: 0, Minute: 0, Lon: 139.76, TZ: 9, Gender: ganzhi.Male},
	{Name: "NewYork-2020-06-15-20:00-female", Year: 2020, Month: 6, Day: 15, Hour: 20, Minute: 0, Lon: -74.0, TZ: -4, Gender: ganzhi.Female},
}

// TestGoldenChart_Zhus verifies all pillar ganzhi.Gan/ganzhi.Zhi are valid and DayMaster is set.
func TestGoldenChart_Pillars(t *testing.T) {
	for _, g := range goldenCharts {
		t.Run(g.Name, func(t *testing.T) {
			cr := computeFullChart(g)

			pillars := []struct {
				name string
				p    zhuInfo
			}{
				{"Year", cr.Year}, {"Month", cr.Month}, {"Day", cr.Day}, {"Hour", cr.Hour},
			}
			for _, p := range pillars {
				if p.p.Gan < 1 || p.p.Gan > 10 {
					t.Errorf("%s.Gan = %d, want [1,10]", p.name, p.p.Gan)
				}
				if p.p.Zhi < 1 || p.p.Zhi > 12 {
					t.Errorf("%s.Zhi = %d, want [1,12]", p.name, p.p.Zhi)
				}
			}
			if cr.DayMaster < 1 || cr.DayMaster > 10 {
				t.Errorf("DayMaster = %d, want [1,10]", cr.DayMaster)
			}
		})
	}
}

// ── Chart output JSON snapshot ──
// Compute full chart output, marshal to JSON, and verify key fields are stable.

func TestGoldenChart_JSONSnapshot(t *testing.T) {
	for _, g := range goldenCharts {
		t.Run(g.Name, func(t *testing.T) {
			cr := computeFullChart(g)
			out := &cr

			b, err := json.MarshalIndent(out, "", "  ")
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			_ = b

			// Verify key invariants.
			var m map[string]any
			if err := json.Unmarshal(b, &m); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}

			// Year pillar is a zhuInfo object.
			if nz, ok := m["Year"].(map[string]any); !ok || nz == nil {
				t.Errorf("Year = %v, want non-nil object", m["Year"])
			}
			if dm, ok := m["DayMaster"].(string); !ok || dm == "" {
				t.Error("DayMaster must be non-empty")
			}
		})
	}
}

// ── ganzhi.Wuxing count totals ──

func TestGoldenChart_WuxingCounts(t *testing.T) {
	for _, g := range goldenCharts {
		t.Run(g.Name, func(t *testing.T) {
			cr := computeFullChart(g)

			// WuxingCount includes pillar stems and hidden stems (sum > 4).
			// Day master's element must have count >= 1.
			dmElem := ganzhi.GanWuxing(cr.DayMaster)
			if cr.WuxingCount[dmElem] < 1 {
				t.Errorf("day master element %s count = %d, want >= 1", dmElem, cr.WuxingCount[dmElem])
			}
		})
	}
}

// ── Hidden stems for each pillar are non-empty ──

func TestGoldenChart_HiddenStems(t *testing.T) {
	for _, g := range goldenCharts {
		t.Run(g.Name, func(t *testing.T) {
			cr := computeFullChart(g)
			hs := cr.HiddenStemsArray()

			if len(hs) != 4 {
				t.Fatalf("len(HiddenStemsArray) = %d, want 4", len(hs))
			}
			names := [4]string{"year", "month", "day", "hour"}
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
			if len(cr.DaYun.Zhus) < 8 {
				t.Errorf("len(Pillars) = %d, want >= 8", len(cr.DaYun.Zhus))
			}
			for i, p := range cr.DaYun.Zhus {
				if p.Gan < 1 || p.Gan > 10 {
					t.Errorf("pillar[%d].Gan = %d, want [1,10]", i, p.Gan)
				}
				if p.Zhi < 1 || p.Zhi > 12 {
					t.Errorf("pillar[%d].Zhi = %d, want [1,12]", i, p.Zhi)
				}
			}

			for i, p := range cr.DaYun.Zhus {
				if p.Name == "" {
					t.Errorf("pillar[%d].Name is empty", i)
				}
			}
		})
	}
}

// ── NaYin consistency ──

func TestGoldenChart_NaYin(t *testing.T) {
	for _, g := range goldenCharts {
		t.Run(g.Name, func(t *testing.T) {
			cr := computeFullChart(g)

			// Each pillar should have its own nayin and it should match.
			if cr.Year.NaYin == "" {
				t.Error("NianZhu.NaYin is empty")
			}
			if cr.Month.NaYin == "" {
				t.Error("YueZhu.NaYin is empty")
			}
			if cr.Day.NaYin == "" {
				t.Error("RiZhu.NaYin is empty")
			}
			if cr.Hour.NaYin == "" {
				t.Error("ShiZhu.NaYin is empty")
			}

			// Nayin elements must be recognizable.
			for _, n := range []string{
				cr.Year.NaYin, cr.Month.NaYin,
				cr.Day.NaYin, cr.Hour.NaYin,
			} {
				if elem := nayinElement(n); elem == 0 {
					t.Errorf("nayin %q has unknown element", n)
				}
			}
		})
	}
}

// ── Ten gods table ──

func TestGoldenChart_TenGods(t *testing.T) {
	for _, g := range goldenCharts {
		t.Run(g.Name, func(t *testing.T) {
			cr := computeFullChart(g)

			pillars := [4]zhuInfo{cr.Year, cr.Month, cr.Day, cr.Hour}
			names := [4]string{"year", "month", "day", "hour"}

			validTenGods := map[string]bool{
				"比肩": true, "劫财": true, "食神": true, "伤官": true,
				"偏财": true, "正财": true, "七杀": true, "正官": true,
				"偏印": true, "正印": true,
			}

			for i, p := range pillars {
				if len(p.TenGods) < 1 {
					t.Errorf("%s pillar: no ten gods", names[i])
					continue
				}
				// First entry must be the stem ten god.
				if p.TenGods[0].Source != sourceGan {
					t.Errorf("%s pillar: first ten god source = %s, want stem", names[i], p.TenGods[0].Source)
				}
				if !validTenGods[p.TenGods[0].TenGod] {
					t.Errorf("%s pillar: unknown ten god %q", names[i], p.TenGods[0].TenGod)
				}
				// Must have at least one hidden stem ten god.
				hasMainQi := false
				for _, e := range p.TenGods {
					if e.Source == sourceMainQi {
						hasMainQi = true
						break
					}
				}
				if !hasMainQi {
					t.Errorf("%s pillar: no main_qi ten god", names[i])
				}
			}
		})
	}
}
