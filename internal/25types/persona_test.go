package persona

import (
	"encoding/json"
	"math"
	"testing"
	"time"

	"github.com/25types/25types/internal/ganzhi"
)

func TestDeviation_Relu(t *testing.T) {
	d := Deviation{0.3, -0.1, 0.0, 0.5, -0.2}
	r := d.Relu()
	want := Deviation{0.3, 0, 0.0, 0.5, 0}
	if r != want {
		t.Errorf("Relu() = %v, want %v", r, want)
	}
}

func TestDeviation_Add(t *testing.T) {
	a := Deviation{0.1, 0.2, 0.3, 0.4, 0.5}
	b := Deviation{0.5, 0.4, 0.3, 0.2, 0.1}
	got := a.Add(b)
	want := Deviation{0.6, 0.6, 0.6, 0.6, 0.6}
	for i := range want {
		if math.Abs(got[i]-want[i]) > 1e-12 {
			t.Errorf("Add[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestDeviation_Dot(t *testing.T) {
	d := Deviation{0.1, 0.2, 0.3, 0.4, 0.5}
	o := [5]float64{1, 2, 3, 4, 5}
	got := d.Dot(o)
	want := 0.1*1 + 0.2*2 + 0.3*3 + 0.4*4 + 0.5*5
	if math.Abs(got-want) > 1e-12 {
		t.Errorf("Dot = %v, want %v", got, want)
	}
}

func TestDeviation_Sum(t *testing.T) {
	d := Deviation{0.1, 0.2, -0.1, -0.15, -0.05}
	got := d.Sum()
	if math.Abs(got) > 1e-12 {
		t.Errorf("Sum = %v, want 0", got)
	}
}

func TestDeviation_ToProportion(t *testing.T) {
	d := Deviation{0.5, -0.5, 0.0, 0.0, 0.0}
	p := d.ToProportion()
	// p[i] = 0.2 + d[i]/5
	want := Proportion{0.3, 0.1, 0.2, 0.2, 0.2}
	for i := range want {
		if math.Abs(p[i]-want[i]) > 1e-12 {
			t.Errorf("ToProportion[%d] = %v, want %v", i, p[i], want[i])
		}
	}
}

func TestProportion_ToDeviation(t *testing.T) {
	p := Proportion{0.3, 0.1, 0.2, 0.2, 0.2}
	d := p.ToDeviation()
	want := Deviation{0.5, -0.5, 0.0, 0.0, 0.0}
	for i := range want {
		if math.Abs(d[i]-want[i]) > 1e-12 {
			t.Errorf("ToDeviation[%d] = %v, want %v", i, d[i], want[i])
		}
	}
}

func TestProportion_Sum(t *testing.T) {
	p := Proportion{0.3, 0.1, 0.2, 0.2, 0.2}
	got := p.Sum()
	if math.Abs(got-1.0) > 1e-12 {
		t.Errorf("Sum = %v, want 1.0", got)
	}
}

func TestDeviation_Roundtrip(t *testing.T) {
	d := Deviation{0.3, -0.1, 0.2, -0.15, -0.25}
	p := d.ToProportion()
	d2 := p.ToDeviation()
	for i := range d {
		if math.Abs(d[i]-d2[i]) > 1e-12 {
			t.Errorf("roundtrip[%d]: %v → %v", i, d[i], d2[i])
		}
	}
}

func TestDeviation_JSON_Object(t *testing.T) {
	d := Deviation{0.3, -0.1, 0.2, -0.25, -0.15}
	b, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	var d2 Deviation
	if err := json.Unmarshal(b, &d2); err != nil {
		t.Fatal(err)
	}
	if d != d2 {
		t.Errorf("JSON roundtrip: %v → %v", d, d2)
	}
}

func TestDeviation_JSON_Array(t *testing.T) {
	raw := `[0.3, -0.1, 0.2, -0.25, -0.15]`
	var d Deviation
	if err := json.Unmarshal([]byte(raw), &d); err != nil {
		t.Fatal(err)
	}
	want := Deviation{0.3, -0.1, 0.2, -0.25, -0.15}
	if d != want {
		t.Errorf("JSON array: got %v, want %v", d, want)
	}
}

func TestProportion_JSON(t *testing.T) {
	p := Proportion{0.3, 0.1, 0.2, 0.2, 0.2}
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	// Unmarshal back
	var p2 Proportion
	if err := json.Unmarshal(b, &p2); err != nil {
		t.Fatal(err)
	}
	// Marshaled values are rounded to 2 decimal places
	for i := range p {
		if math.Abs(p[i]-p2[i]) > 0.01 {
			t.Errorf("Proportion JSON roundtrip[%d]: %v → %v", i, p[i], p2[i])
		}
	}
}

func TestComputeD(t *testing.T) {
	answers := []Answer{
		{QID: "q1", Selections: []string{"W", "F"}},
		{QID: "q2", Selections: []string{"W", "E"}},
		{QID: "q3", Selections: []string{"W", "M"}},
		{QID: "q4", Selections: []string{"W", "R"}},
		{QID: "q5", Selections: []string{"W"}},
	}
	// raw: W=5, F=1, E=1, M=1, R=1, total=9
	// d[i] = 5 * (raw[i]/total - 0.2)
	d := ComputeD(answers)
	wWood := 5 * (5.0/9 - 0.2)
	wOther := 5 * (1.0/9 - 0.2)
	if math.Abs(d[0]-wWood) > 1e-10 {
		t.Errorf("Wood = %v, want %v", d[0], wWood)
	}
	for i := 1; i < 5; i++ {
		if math.Abs(d[i]-wOther) > 1e-10 {
			t.Errorf("d[%d] = %v, want %v", i, d[i], wOther)
		}
	}
}

func TestComputeD_Empty(t *testing.T) {
	d := ComputeD(nil)
	if d != (Deviation{}) {
		t.Errorf("ComputeD(nil) = %v, want zero", d)
	}
	d = ComputeD([]Answer{})
	if d != (Deviation{}) {
		t.Errorf("ComputeD([]) = %v, want zero", d)
	}
}

func TestComputeP(t *testing.T) {
	answers := []Answer{
		{QID: "q1", Selections: []string{"W"}},
		{QID: "q2", Selections: []string{"F"}},
		{QID: "q3", Selections: []string{"E"}},
		{QID: "q4", Selections: []string{"M"}},
		{QID: "q5", Selections: []string{"R"}},
	}
	p := ComputeP(answers)
	want := Proportion{0.2, 0.2, 0.2, 0.2, 0.2}
	if p != want {
		t.Errorf("ComputeP = %v, want %v", p, want)
	}
}

func TestComputeP_Empty(t *testing.T) {
	p := ComputeP(nil)
	want := Proportion{0.2, 0.2, 0.2, 0.2, 0.2}
	if p != want {
		t.Errorf("ComputeP(nil) = %v, want %v", p, want)
	}
}

func TestApplyTranspose(t *testing.T) {
	// S: Wood→Fire, Fire→Earth, Earth→Metal, Metal→Water, Water→Wood
	v := Deviation{1, 0, 0, 0, 0} // pure Wood
	r := ApplyTranspose(&ganzhi.S, v)
	// Sᵀ = S with rows/cols swapped. Since S[j][k]=1 means j→k,
	// column 0 of S has S[4][0]=1 (Water→Wood). So Sᵀ * v =
	// [S[0][0]+S[1][0]+S[2][0]+S[3][0]+S[4][0]]*v = [0,0,0,0,1] when v=[1,0,0,0,0]
	// Wait, ApplyTranspose computes Mᵀv = sum_j M[j][i] * v[j]
	// For M=S, M[j][i] means element j nourishes element i.
	// M[0][1] = 1 (Wood→Fire), M[1][2]=1 (Fire→Earth), etc.
	// M[4][0] = 1 (Water→Wood) — only non-zero in column 0.
	// So r[0] = M[4][0]*v[4] = 0*? = 0... no, v has only v[0]=1.
	// r[0] = M[0][0]*v[0] = 0
	// But column 1: M[0][1]*v[0] = 1*1 = 1 → r[1] = 1
	if r[1] != 1 {
		t.Errorf("ApplyTranspose(S, [1,0,0,0,0])[1] = %v, want 1 (Wood→Fire)", r[1])
	}
}

func TestComputeBond(t *testing.T) {
	dA := Deviation{0.5, 0, -0.2, -0.1, -0.2}
	dB := Deviation{0, 0.4, 0, -0.2, -0.2}
	result := ComputeBond(dA, dB)

	// Verify DeltaA and DeltaB sum to 0 (conservation)
	if math.Abs(result.DeltaA.Sum()) > 1e-12 {
		t.Errorf("DeltaA sum = %v, want 0", result.DeltaA.Sum())
	}
	if math.Abs(result.DeltaB.Sum()) > 1e-12 {
		t.Errorf("DeltaB sum = %v, want 0", result.DeltaB.Sum())
	}

	// dEff = d + delta
	dEffA := dA.Add(result.DeltaA)
	if dEffA != result.DEffA {
		t.Errorf("DEffA mismatch: %v != %v", dEffA, result.DEffA)
	}
}

func TestComputeBond_Symmetric(t *testing.T) {
	// When dA == dB, DeltaA should equal DeltaB
	dA := Deviation{0.3, -0.05, 0.4, -0.1, -0.55}
	result := ComputeBond(dA, dA)
	if result.DeltaA != result.DeltaB {
		t.Errorf("symmetric profiles: DeltaA=%v, DeltaB=%v should be equal", result.DeltaA, result.DeltaB)
	}
}

func TestComputeConcord_Shun(t *testing.T) {
	// Wood nourishes Fire: a person with high Fire (dB Fire↑) receives 顺 from Wood-dominant person (dA Wood↑)
	dA := Deviation{0.5, 0, -0.2, 0, -0.3} // Wood elevated
	dB := Deviation{0, 0.5, -0.2, 0, -0.3} // Fire elevated (nourished by Wood)
	c := ComputeConcord(dA, dB)
	if c != ConcordShun {
		t.Errorf("Concord = %v, want %v", c, ConcordShun)
	}
}

func TestComputeConcord_Ni(t *testing.T) {
	// Metal controls Wood: a person with high Wood (dB Wood↑) suffers 逆 from Metal-dominant person (dA Metal↑)
	dA := Deviation{0, 0, 0, 0.6, -0.6} // Metal elevated
	dB := Deviation{0.6, 0, -0.3, 0, -0.3} // Wood elevated (controlled by Metal)
	c := ComputeConcord(dA, dB)
	if c != ConcordNi {
		t.Errorf("Concord = %v, want %v", c, ConcordNi)
	}
}

func TestComputeConcord_Ping(t *testing.T) {
	// Zero deviation: all Relu values are 0, so sh=0 and ke=0 → Ping
	dA := Deviation{}
	c := ComputeConcord(dA, dA)
	if c != ConcordPing {
		t.Errorf("self-concord = %v, want %v", c, ConcordPing)
	}
}

func TestClassifyIdentity_Pure(t *testing.T) {
	// Pure Wood prototype: {0.8, -0.2, -0.2, -0.2, -0.2}
	d := Deviation{0.8, -0.2, -0.2, -0.2, -0.2}
	id := ClassifyIdentity(d, ganzhi.BuiltinPrototypes)
	if id.ID != "W" {
		t.Errorf("ID = %q, want %q", id.ID, "W")
	}
	if id.Category != "W" {
		t.Errorf("Category = %q, want %q", id.Category, "W")
	}
}

func TestClassifyIdentity_Approximate(t *testing.T) {
	// Slightly off the W prototype should still classify as W
	d := Deviation{0.7, -0.1, -0.2, -0.2, -0.2}
	id := ClassifyIdentity(d, ganzhi.BuiltinPrototypes)
	if id.ID != "W" {
		t.Errorf("ID = %q, want %q", id.ID, "W")
	}
}

func TestClassifyIdentity_Degenerate(t *testing.T) {
	// All zeros → fallback to max element (all equal → index 0 = "W")
	d := Deviation{0, 0, 0, 0, 0}
	id := ClassifyIdentity(d, ganzhi.BuiltinPrototypes)
	if id.ID != "W" {
		t.Errorf("degenerate ID = %q, want %q", id.ID, "W")
	}
}

func TestDeriveCategory(t *testing.T) {
	tests := []struct{ id, want string }{
		{"W", "W"},
		{"FW", "F"},
		{"EM", "E"},
		{"MR", "M"},
		{"RW", "R"},
		{"", ""},
	}
	for _, tc := range tests {
		got := DeriveCategory(tc.id)
		if got != tc.want {
			t.Errorf("DeriveCategory(%q) = %q, want %q", tc.id, got, tc.want)
		}
	}
}

func TestComputeFlow(t *testing.T) {
	// Flow depends only on the current time, not on d
	d := Deviation{}
	result := ComputeFlow(d, time.Date(2024, 2, 10, 12, 0, 0, 0, time.UTC))
	if result.MonthID == "" {
		t.Error("MonthID is empty")
	}
	if result.MonthEN == "" {
		t.Error("MonthEN is empty")
	}
	// Generates/Restrains should be element indices 0-4
	if result.Generates < 0 || result.Generates > 4 {
		t.Errorf("Generates = %d, want 0-4", result.Generates)
	}
	if result.Restrains < 0 || result.Restrains > 4 {
		t.Errorf("Restrains = %d, want 0-4", result.Restrains)
	}
}

func TestComputeFlowYearly(t *testing.T) {
	d := Deviation{}
	results := ComputeFlowYearly(d)
	if len(results) != 12 {
		t.Errorf("got %d months, want 12", len(results))
	}
	seen := make(map[string]bool, 12)
	for _, r := range results {
		if r.MonthID == "" {
			t.Error("empty MonthID")
		}
		if seen[r.MonthID] {
			t.Errorf("duplicate MonthID: %s", r.MonthID)
		}
		seen[r.MonthID] = true
	}
}
