package agent

import (
	"context"
	"encoding/json"
	"testing"
)

func TestCheckToolRegistry_VerifyTerminology(t *testing.T) {
	r := NewCheckToolRegistry()

	tests := []struct {
		name string
		args json.RawMessage
		want string // expected unknown field content
	}{
		{
			name: "all valid terms",
			args: json.RawMessage(`{"terms":["正官","七杀","偏印","用神","格局","身强"]}`),
			want: "",
		},
		{
			name: "unknown term",
			args: json.RawMessage(`{"terms":["正官","not_a_real_term_xyz"]}`),
			want: "not_a_real_term_xyz",
		},
		{
			name: "empty terms",
			args: json.RawMessage(`{"terms":[]}`),
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := r.Execute(context.Background(), "verify_terminology", tt.args)
			if err != nil {
				t.Fatalf("verify_terminology: %v", err)
			}
			var out struct {
				Unknown []string `json:"unknown"`
			}
			if err := json.Unmarshal(result, &out); err != nil {
				t.Fatalf("decode result: %v", err)
			}
			if tt.want == "" && len(out.Unknown) > 0 {
				t.Errorf("unexpected unknown terms: %v", out.Unknown)
			}
			if tt.want != "" {
				found := false
				for _, u := range out.Unknown {
					if u == tt.want {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected %q in unknown list, got %v", tt.want, out.Unknown)
				}
			}
		})
	}
}

func TestCheckToolRegistry_VerifyChartData(t *testing.T) {
	r := NewCheckToolRegistry()

	chart := json.RawMessage(`{"fu_yi":{"qiangruo":"身强","yong":"火"},"nian":{"gan":"甲","zhi":"子"}}`)

	tests := []struct {
		name     string
		chart    json.RawMessage
		args     json.RawMessage
		wantOK   bool
		wantVal  string
	}{
		{
			name:    "matching value",
			chart:   chart,
			args:    json.RawMessage(`{"path":"fu_yi.qiangruo","expected":"身强"}`),
			wantOK:  true,
			wantVal: "身强",
		},
		{
			name:    "mismatching value",
			chart:   chart,
			args:    json.RawMessage(`{"path":"fu_yi.qiangruo","expected":"身弱"}`),
			wantOK:  false,
			wantVal: "身强",
		},
		{
			name:    "missing path",
			chart:   chart,
			args:    json.RawMessage(`{"path":"fu_yi.nonexistent","expected":"xxx"}`),
			wantOK:  false,
			wantVal: "",
		},
		{
			name:    "nested value",
			chart:   chart,
			args:    json.RawMessage(`{"path":"fu_yi.yong","expected":"火"}`),
			wantOK:  true,
			wantVal: "火",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Inject chart data into the args
			var argsMap map[string]any
			if err := json.Unmarshal(tt.args, &argsMap); err != nil {
				t.Fatal(err)
			}
			argsMap["chart"] = tt.chart
			argsBytes, err := json.Marshal(argsMap)
			if err != nil {
				t.Fatal(err)
			}

			result, err := r.Execute(context.Background(), "verify_chart_data", argsBytes)
			if err != nil {
				t.Fatalf("verify_chart_data: %v", err)
			}
			var out struct {
				Match  bool   `json:"match"`
				Actual string `json:"actual"`
			}
			if err := json.Unmarshal(result, &out); err != nil {
				t.Fatalf("decode result: %v", err)
			}
			if out.Match != tt.wantOK {
				t.Errorf("match = %v, want %v", out.Match, tt.wantOK)
			}
			if out.Actual != tt.wantVal {
				t.Errorf("actual = %q, want %q", out.Actual, tt.wantVal)
			}
		})
	}
}

func TestCheckToolRegistry_VerifyStructure(t *testing.T) {
	r := NewCheckToolRegistry()

	tests := []struct {
		name     string
		product  string
		sections []string
		hasData  map[string]bool
		wantMiss []string
	}{
		{
			name:     "chart complete",
			product:  "chart",
			sections: []string{"一、格局总论", "二、用神详解", "三、四柱十神分析", "四、大运提示", "五、流年分析"},
			wantMiss: nil,
		},
		{
			name:     "chart missing liunian",
			product:  "chart",
			sections: []string{"一、格局总论", "二、用神详解", "三、四柱十神分析", "四、大运提示"},
			wantMiss: []string{"五、流年分析"},
		},
		{
			name:     "chart missing liunian but has_liunian=false",
			product:  "chart",
			sections: []string{"一、格局总论", "二、用神详解", "三、四柱十神分析", "四、大运提示"},
			hasData:  map[string]bool{"has_liunian": false},
			wantMiss: nil,
		},
		{
			name:     "bond complete",
			product:  "bond",
			sections: []string{"一、双方八字概览", "二、天干互动分析", "三、地支配合分析", "四、十神互动分析", "五、五行与用神互补", "六、神煞互动", "七、大运同步与结构", "八、综合建议"},
			wantMiss: nil,
		},
		{
			name:     "unknown product",
			product:  "unknown",
			sections: []string{"一、格局总论"},
			wantMiss: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			argsMap := map[string]any{
				"sections": tt.sections,
				"product":  tt.product,
			}
			for k, v := range tt.hasData {
				argsMap[k] = v
			}
			argsBytes, err := json.Marshal(argsMap)
			if err != nil {
				t.Fatal(err)
			}

			result, err := r.Execute(context.Background(), "verify_structure", argsBytes)
			if err != nil {
				t.Fatalf("verify_structure: %v", err)
			}
			var out struct {
				Missing []string `json:"missing"`
			}
			if err := json.Unmarshal(result, &out); err != nil {
				t.Fatalf("decode result: %v", err)
			}
			if len(out.Missing) != len(tt.wantMiss) {
				t.Errorf("missing = %v, want %v", out.Missing, tt.wantMiss)
				return
			}
			for i, m := range out.Missing {
				if m != tt.wantMiss[i] {
					t.Errorf("missing[%d] = %q, want %q", i, m, tt.wantMiss[i])
				}
			}
		})
	}
}

func TestCheckToolRegistry_VerifyTerminology_InvalidJSON(t *testing.T) {
	r := NewCheckToolRegistry()
	result, err := r.Execute(context.Background(), "verify_terminology", json.RawMessage(`not json`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var out struct {
		Unknown []string `json:"unknown"`
	}
	if err := json.Unmarshal(result, &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out.Unknown) == 0 || out.Unknown[0] != "invalid input" {
		t.Errorf("expected ['invalid input'], got %v", out.Unknown)
	}
}

func TestCheckToolRegistry_VerifyChartData_InvalidJSON(t *testing.T) {
	r := NewCheckToolRegistry()
	result, err := r.Execute(context.Background(), "verify_chart_data", json.RawMessage(`bad`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var out struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(result, &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Error != "invalid input" {
		t.Errorf("expected error 'invalid input', got %q", out.Error)
	}
}

func TestCheckToolRegistry_VerifyChartData_NumberAndBool(t *testing.T) {
	r := NewCheckToolRegistry()

	chart := json.RawMessage(`{"score":95,"active":true}`)

	// number value
	argsMap := map[string]any{
		"chart":    chart,
		"path":     "score",
		"expected": "95",
	}
	args, err := json.Marshal(argsMap)
	if err != nil {
		t.Fatal(err)
	}
	result, err := r.Execute(context.Background(), "verify_chart_data", args)
	if err != nil {
		t.Fatalf("number: %v", err)
	}
	var out struct {
		Match  bool   `json:"match"`
		Actual string `json:"actual"`
	}
	if err := json.Unmarshal(result, &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !out.Match || out.Actual != "95" {
		t.Errorf("number: match=%v actual=%q, want true/95", out.Match, out.Actual)
	}

	// bool value
	argsMap2 := map[string]any{
		"chart":    chart,
		"path":     "active",
		"expected": "true",
	}
	args2, err := json.Marshal(argsMap2)
	if err != nil {
		t.Fatal(err)
	}
	result2, err := r.Execute(context.Background(), "verify_chart_data", args2)
	if err != nil {
		t.Fatalf("bool: %v", err)
	}
	if err := json.Unmarshal(result2, &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !out.Match || out.Actual != "true" {
		t.Errorf("bool: match=%v actual=%q, want true/true", out.Match, out.Actual)
	}
}

func TestCheckToolRegistry_VerifyStructure_InvalidJSON(t *testing.T) {
	r := NewCheckToolRegistry()
	result, err := r.Execute(context.Background(), "verify_structure", json.RawMessage(`bad`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var out struct {
		Missing []string `json:"missing"`
	}
	if err := json.Unmarshal(result, &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out.Missing) == 0 || out.Missing[0] != "invalid input" {
		t.Errorf("expected ['invalid input'], got %v", out.Missing)
	}
}

func TestCheckToolRegistry_VerifyStructure_HasLiuNianTrue(t *testing.T) {
	// has_liunian=true with section present: no missing sections reported.
	r := NewCheckToolRegistry()
	argsMap := map[string]any{
		"sections":    []string{"一、格局总论", "二、用神详解", "三、四柱十神分析", "四、大运提示", "五、流年分析"},
		"product":     "chart",
		"has_liunian": true,
	}
	args, err := json.Marshal(argsMap)
	if err != nil {
		t.Fatal(err)
	}
	result, err := r.Execute(context.Background(), "verify_structure", args)
	if err != nil {
		t.Fatalf("verify_structure: %v", err)
	}
	var out struct {
		Missing []string `json:"missing"`
	}
	if err := json.Unmarshal(result, &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out.Missing) != 0 {
		t.Errorf("missing = %v, want empty", out.Missing)
	}
}

func TestCheckToolRegistry_VerifyStructure_MultipleMissing(t *testing.T) {
	r := NewCheckToolRegistry()
	argsMap := map[string]any{
		"sections": []string{"一、格局总论"},
		"product":  "chart",
	}
	args, err := json.Marshal(argsMap)
	if err != nil {
		t.Fatal(err)
	}
	result, err := r.Execute(context.Background(), "verify_structure", args)
	if err != nil {
		t.Fatalf("verify_structure: %v", err)
	}
	var out struct {
		Missing []string `json:"missing"`
	}
	if err := json.Unmarshal(result, &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out.Missing) < 3 {
		t.Errorf("expected >=3 missing sections, got %v", out.Missing)
	}
}

func TestCheckToolRegistry_UnknownTool(t *testing.T) {
	r := NewCheckToolRegistry()
	_, err := r.Execute(context.Background(), "nonexistent", json.RawMessage(`{}`))
	if err == nil {
		t.Error("expected error for unknown tool")
	}
}

func TestCheckToolRegistry_Schemas(t *testing.T) {
	r := NewCheckToolRegistry()
	defs := r.Schemas()
	if len(defs) != 3 {
		t.Errorf("schema count = %d, want 3", len(defs))
	}
	names := map[string]bool{}
	for _, d := range defs {
		var fn struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(d.Function, &fn); err != nil {
			t.Errorf("invalid def: %v", err)
			continue
		}
		names[fn.Name] = true
	}
	for _, n := range []string{"verify_terminology", "verify_chart_data", "verify_structure"} {
		if !names[n] {
			t.Errorf("missing tool schema: %s", n)
		}
	}
}
