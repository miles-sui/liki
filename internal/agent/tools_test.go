package agent

import (
	"context"
	"encoding/json"
	"testing"
)

const bt = `"birth":{"time":"1984-02-15T08:00:00+08:00","longitude":116.4}`

func TestNamingRegistry_Execute_ComputeChart(t *testing.T) {
	r := NewNamingToolRegistry()

	args := json.RawMessage(`{` + bt + `,"gender":"male"}`)
	result, err := r.Execute(context.Background(), "compute_chart", args)
	if err != nil {
		t.Fatalf("compute_chart: %v", err)
	}

	var env struct {
		Product string          `json:"_product"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatalf("decode result: %v", err)
	}
	if env.Product != "chart" {
		t.Errorf("_product = %q, want chart", env.Product)
	}

	var chart struct {
		DaYun struct {
			Zhu []json.RawMessage `json:"zhu"`
		} `json:"da_yun"`
		FuYi struct {
			Strength string `json:"qiangruo"`
		} `json:"fu_yi"`
		Nian struct {
			Gan string `json:"Gan"`
			Zhi string `json:"Zhi"`
		} `json:"nian"`
		Yue struct {
			Gan string `json:"Gan"`
			Zhi string `json:"Zhi"`
		} `json:"yue"`
		Ri struct {
			Gan string `json:"Gan"`
			Zhi string `json:"Zhi"`
		} `json:"ri"`
		Shi struct {
			Gan string `json:"Gan"`
			Zhi string `json:"Zhi"`
		} `json:"shi"`
	}
	if err := json.Unmarshal(env.Data, &chart); err != nil {
		t.Fatal(err)
	}
	if len(chart.DaYun.Zhu) == 0 {
		t.Error("DaYun.Zhus is empty")
	}
	if chart.FuYi.Strength == "" {
		t.Error("FuYi.QiangRuo is empty")
	}
	if chart.Nian.Gan == "" || chart.Nian.Zhi == "" {
		t.Error("Nian.Gan/Zhi is empty")
	}
}

func TestNamingRegistry_Execute_ComputeZiwei(t *testing.T) {
	r := NewNamingToolRegistry()

	args := json.RawMessage(`{` + bt + `,"gender":"male"}`)
	result, err := r.Execute(context.Background(), "compute_ziwei", args)
	if err != nil {
		t.Fatalf("compute_ziwei: %v", err)
	}
	var env struct {
		Product string          `json:"_product"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if env.Product != "ziwei" {
		t.Errorf("_product = %q, want ziwei", env.Product)
	}

	var chart struct {
		Palaces []struct {
			Name  string `json:"name"`
			Gan   string `json:"gan"`
			Zhi   string `json:"zhi"`
			Stars []struct {
				Star int    `json:"star"`
				Name string `json:"name"`
			} `json:"stars"`
		} `json:"palaces"`
		MingGong int `json:"ming_gong"`
		ShenGong int `json:"shen_gong"`
		JuShu    int `json:"ju_shu"`
	}
	if err := json.Unmarshal(env.Data, &chart); err != nil {
		t.Fatal(err)
	}
	if len(chart.Palaces) != 12 {
		t.Errorf("palaces = %d, want 12", len(chart.Palaces))
	}
	if chart.JuShu < 2 || chart.JuShu > 6 {
		t.Errorf("ju_shu = %d, want [2,6]", chart.JuShu)
	}
}

func TestNamingRegistry_Execute_ComputeNamingWuGe(t *testing.T) {
	r := NewNamingToolRegistry()
	args := json.RawMessage(`{"surname":"王","yong_shen":"金","xi_shen":["土","金"]}`)
	result, err := r.Execute(context.Background(), "compute_naming_wuge", args)
	if err != nil {
		t.Fatalf("compute_naming_wuge: %v", err)
	}
	var env struct {
		Product string          `json:"_product"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if env.Product != "naming_wuge" {
		t.Errorf("_product = %q, want naming_wuge", env.Product)
	}

	var wuge struct {
		Surname string                   `json:"surname"`
		Combos  []map[string]interface{} `json:"combos"`
	}
	if err := json.Unmarshal(env.Data, &wuge); err != nil {
		t.Fatal(err)
	}
	if wuge.Surname != "王" {
		t.Errorf("surname = %q, want 王", wuge.Surname)
	}
	if len(wuge.Combos) == 0 {
		t.Error("combos is empty")
	}
}

func TestNamingRegistry_Execute_ComputeNamingDetail(t *testing.T) {
	r := NewNamingToolRegistry()
	args := json.RawMessage(`{"surname":"王","names":["王伟杰","王浩然"]}`)
	result, err := r.Execute(context.Background(), "compute_naming_detail", args)
	if err != nil {
		t.Fatalf("compute_naming_detail: %v", err)
	}
	var env struct {
		Data []struct {
			Name string `json:"name"`
		} `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if len(env.Data) == 0 {
		t.Skip("naming detail: no dictionary entries for test names")
	}
}

func TestNamingRegistry_Execute_ComputeNamingEvaluate(t *testing.T) {
	r := NewNamingToolRegistry()
	args := json.RawMessage(`{"surname":"王","given_name":"小明","yong_shen":"金"}`)
	result, err := r.Execute(context.Background(), "compute_naming_evaluate", args)
	if err != nil {
		t.Fatalf("compute_naming_evaluate: %v", err)
	}
	var env struct {
		Data struct {
			Surname   string `json:"surname"`
			GivenName string `json:"given_name"`
		} `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if env.Data.Surname != "王" {
		t.Errorf("surname = %q, want 王", env.Data.Surname)
	}
}

func TestNamingRegistry_Execute_ComputeNamingCompose(t *testing.T) {
	r := NewNamingToolRegistry()
	args := json.RawMessage(`{
		"surname":"王",
		"combos":[{"stroke1":4,"stroke2":5}],
		"yong_chars":{"4":[{"char":"杰","tone":2},{"char":"凯","tone":3}],"5":[{"char":"可","tone":3},{"char":"永","tone":3}]},
		"xi_chars":{"4":[{"char":"杰","tone":2}]}
	}`)
	result, err := r.Execute(context.Background(), "compute_naming_compose", args)
	if err != nil {
		t.Fatalf("compute_naming_compose: %v", err)
	}
	var env struct {
		Data []string `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if len(env.Data) == 0 {
		t.Error("compose returned no names")
	}
	for _, name := range env.Data {
		if len(name) == 0 {
			t.Error("composed name is empty string")
		}
	}
}

func TestNamingRegistry_Execute_ComputeTime(t *testing.T) {
	r := NewNamingToolRegistry()
	args := json.RawMessage(`{"time":"1984-02-15T08:00:00+08:00","longitude":116.4}`)
	result, err := r.Execute(context.Background(), "compute_time", args)
	if err != nil {
		t.Fatalf("compute_time: %v", err)
	}
	var ts struct {
		Gregorian string `json:"Gregorian"`
		Solar     string `json:"Solar"`
		Lunar     struct {
			Year    int  `json:"Year"`
			Month   int  `json:"Month"`
			Day     int  `json:"Day"`
			Leap    bool `json:"Leap"`
			Shichen string `json:"Shichen"`
		} `json:"Lunar"`
	}
	if err := json.Unmarshal(result, &ts); err != nil {
		t.Fatal(err)
	}
	if ts.Solar == "" {
		t.Error("Solar is empty")
	}
	if ts.Lunar.Year == 0 {
		t.Error("Lunar.Year is zero")
	}
}

func TestNamingRegistry_Execute_UnknownTool(t *testing.T) {
	r := NewNamingToolRegistry()
	_, err := r.Execute(context.Background(), "nonexistent", json.RawMessage(`{}`))
	if err == nil {
		t.Error("expected error for unknown tool")
	}
}

func TestNamingRegistry_Execute_InvalidJSON(t *testing.T) {
	r := NewNamingToolRegistry()
	_, err := r.Execute(context.Background(), "compute_chart", json.RawMessage(`{bad`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestNamingRegistry_Schemas(t *testing.T) {
	r := NewNamingToolRegistry()
	defs := r.Schemas()
	if len(defs) != 8 {
		t.Errorf("schema count = %d, want 8", len(defs))
	}

	found := map[string]bool{}
	for _, d := range defs {
		var fn struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		if err := json.Unmarshal(d.Function, &fn); err != nil {
			t.Errorf("invalid def: %v", err)
			continue
		}
		if fn.Name == "" {
			t.Error("tool missing name")
		}
		if fn.Description == "" {
			t.Errorf("tool %q missing description", fn.Name)
		}
		found[fn.Name] = true
	}

	want := []string{
		"query_city", "compute_time", "compute_chart", "compute_ziwei",
		"compute_naming_wuge", "compute_naming_compose",
		"compute_naming_detail", "compute_naming_evaluate",
	}
	for _, name := range want {
		if !found[name] {
			t.Errorf("missing tool: %s", name)
		}
	}
}

func TestToolParams_NamingTools(t *testing.T) {
	names := []string{
		"compute_chart", "compute_ziwei",
		"compute_naming_wuge", "compute_naming_compose",
		"compute_naming_detail", "compute_naming_evaluate",
	}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			desc, params, err := toolParams(name)
			if err != nil {
				t.Fatalf("toolParams(%q): %v", name, err)
			}
			if desc == "" {
				t.Error("empty description")
			}
			var schema map[string]any
			if err := json.Unmarshal(params, &schema); err != nil {
				t.Fatal("invalid parameters JSON:", err)
			}
			if tpe, ok := schema["type"]; !ok || tpe != "object" {
				t.Errorf("schema type = %v, want object", tpe)
			}
		})
	}
}

func TestToolParams_UnknownTool(t *testing.T) {
	_, _, err := toolParams("nonexistent_tool")
	if err == nil {
		t.Error("unknown tool should return error")
	}
}

func TestValidateTools(t *testing.T) {
	if err := ValidateTools(); err != nil {
		t.Fatalf("ValidateTools: %v", err)
	}
	// Second call should be idempotent (sync.Once)
	if err := ValidateTools(); err != nil {
		t.Fatalf("ValidateTools (2nd): %v", err)
	}
}
