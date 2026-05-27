package tianwen

import (
	"testing"
	"time"
)

func TestJulianDay_KnownDates(t *testing.T) {
	tests := []struct {
		y, m, d int
		want    int
	}{
		{2000, 1, 1, 2451545},  // J2000.0 epoch
		{2000, 1, 2, 2451546},
		{1999, 1, 1, 2451180},
		{2020, 2, 29, 2458909},
	}
	for _, tc := range tests {
		got := JulianDay(tc.y, tc.m, tc.d)
		if got != tc.want {
			t.Errorf("JulianDay(%d,%d,%d) = %d, want %d", tc.y, tc.m, tc.d, got, tc.want)
		}
	}
}

func TestSolarLongitude_Equinox(t *testing.T) {
	// March equinox 2024: solar longitude ≈ 0°
	equinox := time.Date(2024, 3, 20, 3, 6, 0, 0, time.UTC)
	lon := SolarLongitude(equinox)
	if lon > 2 && lon < 358 {
		t.Errorf("March equinox 2024 solar longitude = %.2f, want ~0", lon)
	}
}

func TestSolarTermLongitudes(t *testing.T) {
	if len(SolarTermLongitudes) != 12 {
		t.Errorf("SolarTermLongitudes has %d entries, want 12", len(SolarTermLongitudes))
	}
	// First term (立春) should be 315°
	if SolarTermLongitudes[0] != 315 {
		t.Errorf("SolarTermLongitudes[0] = %.0f, want 315", SolarTermLongitudes[0])
	}
}

func TestPrecomputeSolarTerms(t *testing.T) {
	entries := PrecomputeSolarTerms(2024)
	if len(entries) != 12 {
		t.Fatalf("PrecomputeSolarTerms(2024) = %d entries, want 12", len(entries))
	}
	// Should be sorted by date
	for i := 1; i < len(entries); i++ {
		if !entries[i-1].Date.Before(entries[i].Date) {
			t.Errorf("entries not sorted: %s >= %s", entries[i-1].Date, entries[i].Date)
		}
	}
	// Each entry must have a month ID and name
	for i, e := range entries {
		if e.MonthID == "" {
			t.Errorf("entry %d has empty MonthID", i)
		}
		if e.NameEN == "" {
			t.Errorf("entry %d has empty NameEN", i)
		}
	}
}

func TestGetCurrentSolarMonth(t *testing.T) {
	// Early February should return 寅月
	feb := time.Date(2024, 2, 10, 12, 0, 0, 0, time.UTC)
	month := GetCurrentSolarMonth(feb)
	if month == "" {
		t.Error("GetCurrentSolarMonth returned empty string")
	}
}

func TestLiChunDay(t *testing.T) {
	m, d := LiChunDay(2024)
	// 立春 2024 should be early February
	if m != 2 || d < 1 || d > 15 {
		t.Errorf("LiChunDay(2024) = (%d, %d), want early Feb", m, d)
	}
}
