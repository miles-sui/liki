package bazi

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"liki/internal/engine/ganzhi"
	"liki/internal/engine/tianwen"
)

var updateGolden = os.Getenv("UPDATE_GOLDEN") == "1"

func TestGoldenComputeChart(t *testing.T) {
	st := tianwen.GregorianToSolar(
		time.Date(1984, 2, 15, 8, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	chart := ComputeChart(st, ganzhi.Male)

	got, err := json.MarshalIndent(chart, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	golden := filepath.Join("testdata", "chart_golden.json")
	if updateGolden {
		if err := os.MkdirAll("testdata", 0755); err != nil {
			t.Fatalf("mkdir testdata: %v", err)
		}
		if err := os.WriteFile(golden, got, 0644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		return
	}

	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("read golden: %v — run with -update to regenerate", err)
	}
	if string(got) != string(want) {
		t.Errorf("chart output differs from golden file.\nGot:\n%s\n\nWant:\n%s", got, want)
	}
}

func TestGoldenComputeLiuNian(t *testing.T) {
	st := tianwen.GregorianToSolar(
		time.Date(1984, 2, 15, 8, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	chart := ComputeChart(st, ganzhi.Male)
	ln, err := ComputeLiuNian(chart.ChartBase, 2026)
	if err != nil {
		t.Fatalf("ComputeLiuNian: %v", err)
	}

	got, err := json.MarshalIndent(ln, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	golden := filepath.Join("testdata", "liunian_golden.json")
	if updateGolden {
		if err := os.MkdirAll("testdata", 0755); err != nil {
			t.Fatalf("mkdir testdata: %v", err)
		}
		if err := os.WriteFile(golden, got, 0644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		return
	}

	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("read golden: %v — run with -update to regenerate", err)
	}
	if string(got) != string(want) {
		t.Errorf("liunian output differs from golden file.\nGot:\n%s\n\nWant:\n%s", got, want)
	}
}

func TestGoldenSolarToLunar(t *testing.T) {
	dates := []struct {
		name string
		gt   tianwen.GregorianTime
	}{
		{"epoch1970", tianwen.GregorianTime(time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC))},
		{"spring2026", tianwen.GregorianTime(time.Date(2026, 2, 17, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)))},
		{"leap2025", tianwen.GregorianTime(time.Date(2025, 7, 25, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)))},
		{"solstice2024", tianwen.GregorianTime(time.Date(2024, 12, 21, 18, 0, 0, 0, time.FixedZone("CST", 8*3600)))},
	}

	for _, d := range dates {
		t.Run(d.name, func(t *testing.T) {
			lt := tianwen.SolarToLunar(d.gt)
			got, err := json.MarshalIndent(lt, "", "  ")
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}

			golden := filepath.Join("testdata", "lunar_"+d.name+"_golden.json")
			if updateGolden {
				if err := os.MkdirAll("testdata", 0755); err != nil {
					t.Fatalf("mkdir testdata: %v", err)
				}
				if err := os.WriteFile(golden, got, 0644); err != nil {
					t.Fatalf("write golden: %v", err)
				}
				return
			}

			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("read golden: %v — run with -update to regenerate", err)
			}
			if string(got) != string(want) {
				t.Errorf("lunar %s output differs from golden file.\nGot:\n%s\n\nWant:\n%s", d.name, got, want)
			}
		})
	}
}
