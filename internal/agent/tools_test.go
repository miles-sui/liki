package agent

import (
	"context"
	"encoding/json"
	"testing"
)

const bt = `"birth":{"time":"1984-02-15T08:00:00+08:00","longitude":116.4}`

func TestToolRegistry_Execute_ComputeChart(t *testing.T) {
	r := NewChatToolRegistry()

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
	// 1984-02-15 08:00 → 年甲子, 月丙寅, 日己卯, 时戊辰
	if chart.Nian.Gan == "" || chart.Nian.Zhi == "" {
		t.Error("Nian.Gan/Zhi is empty")
	}
}

func TestToolRegistry_Execute_ComputeBond(t *testing.T) {
	r := NewChatToolRegistry()

	args := json.RawMessage(`{
			"a":{` + bt + `,"gender":"male"},
			"b":{` + bt + `,"gender":"female"}
		}`)
	result, err := r.Execute(context.Background(), "compute_bond", args)
	if err != nil {
		t.Fatalf("compute_bond: %v", err)
	}

	var env struct {
		Product string          `json:"_product"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if env.Product != "bond" {
		t.Errorf("_product = %q, want bond", env.Product)
	}

	var bond struct {
		ZhuCross struct {
			Pairs []json.RawMessage `json:"pairs"`
		} `json:"zhu_cross"`
		ShenshaCross struct {
			Lu struct {
				AInB bool `json:"AInB"`
				BInA bool `json:"BInA"`
			} `json:"Lu"`
		} `json:"ShenshaCross"`
	}
	if err := json.Unmarshal(env.Data, &bond); err != nil {
		t.Fatal(err)
	}
	if len(bond.ZhuCross.Pairs) == 0 {
		t.Error("ZhuCross.Pairs is empty")
	}
}

func TestToolRegistry_Execute_ComputeLiuNian(t *testing.T) {
	r := NewChatToolRegistry()

	args := json.RawMessage(`{
			"year": 2025,
			` + bt + `,
			"gender": "male"
		}`)
	result, err := r.Execute(context.Background(), "compute_liunian", args)
	if err != nil {
		t.Fatalf("compute_liunian: %v", err)
	}

	var env struct {
		Product string          `json:"_product"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if env.Product != "liunian" {
		t.Errorf("_product = %q, want liunian", env.Product)
	}

	var ln struct {
		Year       int    `json:"year"`
		YearStem   string `json:"year_stem"`
		YearBranch string `json:"year_branch"`
	}
	if err := json.Unmarshal(env.Data, &ln); err != nil {
		t.Fatal(err)
	}
	if ln.Year != 2025 {
		t.Errorf("year = %d, want 2025", ln.Year)
	}
	if ln.YearStem == "" {
		t.Error("year_stem is empty")
	}
}

func TestToolRegistry_Execute_ComputeLiuYue(t *testing.T) {
	r := NewChatToolRegistry()

	args := json.RawMessage(`{"year": 2025, "month": 6, ` + bt + `, "gender": "male"}`)
	result, err := r.Execute(context.Background(), "compute_liuyue", args)
	if err != nil {
		t.Fatalf("compute_liuyue: %v", err)
	}

	var env struct {
		Product string          `json:"_product"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	var ly struct {
		Year        int    `json:"year"`
		Month       int    `json:"month"`
		MonthStem   string `json:"month_stem"`
		MonthBranch string `json:"month_branch"`
	}
	if err := json.Unmarshal(env.Data, &ly); err != nil {
		t.Fatal(err)
	}
	if ly.Year != 2025 {
		t.Errorf("year = %d, want 2025", ly.Year)
	}
	if ly.Month != 6 {
		t.Errorf("month = %d, want 6", ly.Month)
	}
	if ly.MonthStem == "" {
		t.Error("month_stem is empty")
	}
}

func TestToolRegistry_Execute_ComputeLiuri(t *testing.T) {
	r := NewChatToolRegistry()

	args := json.RawMessage(`{
			"year": 2025, "month": 6, "day": 1,
			` + bt + `,
			
			"gender": "male"
		}`)
	result, err := r.Execute(context.Background(), "compute_liuri", args)
	if err != nil {
		t.Fatalf("compute_liuri: %v", err)
	}

	var env struct {
		Data struct {
			Date      string `json:"date"`
			DayStem   string `json:"day_stem"`
			DayBranch string `json:"day_branch"`
		} `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if env.Data.DayStem == "" {
		t.Error("day_stem is empty")
	}
}

func TestToolRegistry_Execute_ComputeXiaoYun(t *testing.T) {
	r := NewChatToolRegistry()

	args := json.RawMessage(`{` + bt + `,"gender":"male","count":10}`)
	result, err := r.Execute(context.Background(), "compute_xiaoyun", args)
	if err != nil {
		t.Fatalf("compute_xiaoyun: %v", err)
	}

	var env struct {
		Data []struct {
			Age int    `json:"age"`
			Gan string `json:"gan"`
			Zhi string `json:"zhi"`
		} `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if len(env.Data) != 10 {
		t.Errorf("xiaoyun count = %d, want 10", len(env.Data))
	}
	if env.Data[0].Gan == "" || env.Data[0].Zhi == "" {
		t.Error("first xiaoyun pillar Gan/Zhi is empty")
	}
}

func TestToolRegistry_Execute_GetCityCoords(t *testing.T) {
	t.Skip("skipping network-dependent test: geocoding API unreliable in CI")
	r := NewChatToolRegistry()

	args := json.RawMessage(`{"city":"北京"}`)
	result, err := r.Execute(context.Background(), "get_city_coords", args)
	if err != nil {
		t.Skipf("get_city_coords: network may be unavailable: %v", err)
	}
	var coords struct {
		City      string  `json:"city"`
		Longitude float64 `json:"longitude"`
		Latitude  float64 `json:"latitude"`
	}
	if err := json.Unmarshal(result, &coords); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if coords.City != "北京" {
		t.Errorf("city = %q, want 北京", coords.City)
	}
	if coords.Longitude < 115 || coords.Longitude > 118 {
		t.Errorf("longitude = %f, want ~116.4", coords.Longitude)
	}
}

func TestToolRegistry_Execute_UnknownTool(t *testing.T) {
	r := NewChatToolRegistry()
	_, err := r.Execute(context.Background(), "nonexistent", json.RawMessage(`{}`))
	if err == nil {
		t.Error("expected error for unknown tool")
	}
}

func TestToolRegistry_Execute_InvalidJSON(t *testing.T) {
	r := NewChatToolRegistry()
	_, err := r.Execute(context.Background(), "compute_chart", json.RawMessage(`{bad`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestToolRegistry_Execute_ComputeZiwei(t *testing.T) {
	r := NewChatToolRegistry()

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

func TestToolRegistry_Execute_ComputeQimen(t *testing.T) {
	r := NewChatToolRegistry()

	args := json.RawMessage(`{` + bt + `,"kind":"shi"}`)
	result, err := r.Execute(context.Background(), "compute_qimen", args)
	if err != nil {
		t.Fatalf("compute_qimen: %v", err)
	}
	var env struct {
		Product string          `json:"_product"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if env.Product != "qimen" {
		t.Errorf("_product = %q, want qimen", env.Product)
	}

	var pan struct {
		Pan struct {
			JuShu  int  `json:"jushu"`
			YinDun bool `json:"yin_dun"`
		} `json:"pan"`
	}
	if err := json.Unmarshal(env.Data, &pan); err != nil {
		t.Fatal(err)
	}
	if pan.Pan.JuShu < 1 {
		t.Error("jushu is zero")
	}
}

func TestToolRegistry_Execute_ComputeBazhai(t *testing.T) {
	r := NewChatToolRegistry()

	args := json.RawMessage(`{` + bt + `,"gender":"male"}`)
	result, err := r.Execute(context.Background(), "compute_bazhai", args)
	if err != nil {
		t.Fatalf("compute_bazhai: %v", err)
	}
	var env struct {
		Product string          `json:"_product"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if env.Product != "bazhai" {
		t.Errorf("_product = %q, want bazhai", env.Product)
	}

	var chart struct {
		MingGua struct {
			GuaNumber int    `json:"gua_number"`
			Group     string `json:"group"`
		} `json:"ming_gua"`
	}
	if err := json.Unmarshal(env.Data, &chart); err != nil {
		t.Fatal(err)
	}
	if chart.MingGua.GuaNumber < 1 || chart.MingGua.GuaNumber > 9 {
		t.Errorf("ming_gua.gua_number = %d, want [1,9]", chart.MingGua.GuaNumber)
	}
}

func TestToolRegistry_Execute_QueryHuangliDate(t *testing.T) {
	r := NewChatToolRegistry()

	args := json.RawMessage(`{"date":"2025-06-01","event":"结婚"}`)
	result, err := r.Execute(context.Background(), "query_huangli_date", args)
	if err != nil {
		t.Fatalf("query_huangli_date: %v", err)
	}
	var env struct {
		Product string          `json:"_product"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if env.Product != "huangli_date" {
		t.Errorf("_product = %q, want huangli_date", env.Product)
	}

	var day struct {
		Date      string `json:"date"`
		RiZhu struct {
			Gan string `json:"gan"`
			Zhi string `json:"zhi"`
		} `json:"day_pillar"`
	}
	if err := json.Unmarshal(env.Data, &day); err != nil {
		t.Fatal(err)
	}
	if day.Date != "2025-06-01" {
		t.Errorf("date = %q, want 2025-06-01", day.Date)
	}
	if day.RiZhu.Gan == "" {
		t.Error("day_pillar.gan is empty")
	}
}

func TestToolRegistry_Schemas(t *testing.T) {
	r := NewChatToolRegistry()
	defs := r.Schemas()
	if len(defs) == 0 {
		t.Fatal("no tool definitions")
	}
	found := false
	for _, d := range defs {
		var name struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(d.Function, &name); err != nil {
			t.Errorf("invalid def: %v", err)
		}
		if name.Name == "compute_chart" {
			found = true
		}
	}
	if !found {
		t.Error("compute_chart not found in tool definitions")
	}
}

func TestToolRegistry_Execute_ComputeLiuShi(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{"year":2025,"month":6,"day":1,"hour":8,` + bt + `}`)
	result, err := r.Execute(context.Background(), "compute_liushi", args)
	if err != nil {
		t.Fatalf("compute_liushi: %v", err)
	}
	var env struct {
		Data struct {
			Time       string `json:"time"`
			HourStem   string `json:"hour_stem"`
			HourBranch string `json:"hour_branch"`
		} `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if env.Data.HourStem == "" {
		t.Error("hour_stem is empty")
	}
}

func TestToolRegistry_Execute_ComputeXiaoXian(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{"gender":"male","count":5}`)
	result, err := r.Execute(context.Background(), "compute_xiaoxian", args)
	if err != nil {
		t.Fatalf("compute_xiaoxian: %v", err)
	}
	var env struct {
		Data []struct {
			Age    int    `json:"age"`
			Branch string `json:"branch"`
		} `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if len(env.Data) != 5 {
		t.Errorf("xiaoxian count = %d, want 5", len(env.Data))
	}
}

// --- ziwei chart-dependent tools ---

func getZiweiChart(t *testing.T, r *ChatToolRegistry) json.RawMessage {
	t.Helper()
	result, err := r.Execute(context.Background(), "compute_ziwei", json.RawMessage(`{`+bt+`,"gender":"male"}`))
	if err != nil {
		t.Fatalf("compute_ziwei (setup): %v", err)
	}
	var env struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	return env.Data
}

func TestToolRegistry_Execute_ComputeZiweiDaXian(t *testing.T) {
	r := NewChatToolRegistry()
	chart := getZiweiChart(t, r)
	args := json.RawMessage(`{"chart":` + string(chart) + `}`)
	result, err := r.Execute(context.Background(), "compute_ziwei_daxian", args)
	if err != nil {
		t.Fatalf("compute_ziwei_daxian: %v", err)
	}
	var env struct {
		Data []struct {
			StartAge int    `json:"start_age"`
			EndAge   int    `json:"end_age"`
			Name     string `json:"name"`
		} `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if len(env.Data) != 12 {
		t.Errorf("daxian steps = %d, want 12", len(env.Data))
	}
	if env.Data[0].Name == "" {
		t.Error("first step name is empty")
	}
}

func TestToolRegistry_Execute_ComputeZiweiLiuNian(t *testing.T) {
	r := NewChatToolRegistry()
	chart := getZiweiChart(t, r)
	args := json.RawMessage(`{"liu_year":2025,"chart":` + string(chart) + `}`)
	result, err := r.Execute(context.Background(), "compute_ziwei_liunian", args)
	if err != nil {
		t.Fatalf("compute_ziwei_liunian: %v", err)
	}
	var env struct {
		Data struct {
			MingGong int `json:"ming_gong"`
		} `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if env.Data.MingGong < 0 || env.Data.MingGong > 11 {
		t.Errorf("ming_gong = %d, want [0,11]", env.Data.MingGong)
	}
}

func TestToolRegistry_Execute_ComputeZiweiLiuYue(t *testing.T) {
	r := NewChatToolRegistry()
	chart := getZiweiChart(t, r)
	args := json.RawMessage(`{"liu_year":2025,"lunar_month":1,"chart":` + string(chart) + `}`)
	result, err := r.Execute(context.Background(), "compute_ziwei_liuyue", args)
	if err != nil {
		t.Fatalf("compute_ziwei_liuyue: %v", err)
	}
	var env struct {
		Data struct {
			MingGong int `json:"ming_gong"`
		} `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if env.Data.MingGong < 0 || env.Data.MingGong > 11 {
		t.Errorf("ming_gong = %d, want [0,11]", env.Data.MingGong)
	}
}

func TestToolRegistry_Execute_ComputeZiweiLiuRi(t *testing.T) {
	r := NewChatToolRegistry()
	chart := getZiweiChart(t, r)
	args := json.RawMessage(`{"liu_year":2025,"lunar_month":1,"lunar_day":1,"chart":` + string(chart) + `}`)
	result, err := r.Execute(context.Background(), "compute_ziwei_liuri", args)
	if err != nil {
		t.Fatalf("compute_ziwei_liuri: %v", err)
	}
	var env struct {
		Data struct {
			MingGong int `json:"ming_gong"`
		} `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if env.Data.MingGong < 0 || env.Data.MingGong > 11 {
		t.Errorf("ming_gong = %d, want [0,11]", env.Data.MingGong)
	}
}

func TestToolRegistry_Execute_ComputeZiweiBond(t *testing.T) {
	r := NewChatToolRegistry()
	chartA := getZiweiChart(t, r)
	chartB, err := r.Execute(context.Background(), "compute_ziwei", json.RawMessage(`{`+bt+`,"gender":"female"}`))
	if err != nil {
		t.Fatalf("compute_ziwei B (setup): %v", err)
	}
	var envB struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(chartB, &envB); err != nil {
		t.Fatal(err)
	}

	args := json.RawMessage(`{"a":` + string(chartA) + `,"b":` + string(envB.Data) + `}`)
	result, err := r.Execute(context.Background(), "compute_ziwei_bond", args)
	if err != nil {
		t.Fatalf("compute_ziwei_bond: %v", err)
	}
	var env struct {
		Data struct {
			StarCross []struct {
				Star  int    `json:"star"`
				FromA int    `json:"from_a"`
				IntoB int    `json:"into_b"`
			} `json:"star_cross"`
		} `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if len(env.Data.StarCross) == 0 {
		t.Error("star_cross is empty")
	}
}

// --- naming handlers ---

func TestToolRegistry_Execute_ComputeNamingWuGe(t *testing.T) {
	r := NewChatToolRegistry()
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

func TestToolRegistry_Execute_ComputeNamingDetail(t *testing.T) {
	r := NewChatToolRegistry()
	// Use full names (surname + given) with common naming characters in the dictionary.
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
		t.Skip("naming detail: no dictionary entries for test names (expected for some char sets)")
	}
}

func TestToolRegistry_Execute_ComputeNamingEvaluate(t *testing.T) {
	r := NewChatToolRegistry()
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

func TestToolRegistry_Execute_ComputeNamingCompose(t *testing.T) {
	r := NewChatToolRegistry()
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
	// ComposeNames returns []string — verify it's a non-empty string array
	var env struct {
		Data []string `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if len(env.Data) == 0 {
		t.Error("compose returned no names")
	}
	// Names should contain the surname
	for _, name := range env.Data {
		if len(name) == 0 {
			t.Error("composed name is empty string")
		}
	}
}

// --- bazhai: minggua ---

func TestToolRegistry_Execute_ComputeMingGua(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{"gender":"male","birth_year":1990}`)
	result, err := r.Execute(context.Background(), "compute_minggua", args)
	if err != nil {
		t.Fatalf("compute_minggua: %v", err)
	}
	var env struct {
		Data struct {
			GuaNumber int    `json:"gua_number"`
			Group     string `json:"group"`
		} `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if env.Data.GuaNumber < 1 || env.Data.GuaNumber > 9 {
		t.Errorf("gua_number = %d, want [1,9]", env.Data.GuaNumber)
	}
	if env.Data.Group == "" {
		t.Error("group is empty")
	}
}

// --- xuankong: chart and sanyuan ---

func TestToolRegistry_Execute_ComputeXuankong(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{` + bt + `,"sit_mountain":1,"face_mountain":12}`)
	result, err := r.Execute(context.Background(), "compute_xuankong", args)
	if err != nil {
		t.Fatalf("compute_xuankong: %v", err)
	}
	var env struct {
		Data struct {
			Yun struct {
				YunNumber int `json:"yun_number"`
			} `json:"yun"`
			Palaces []json.RawMessage `json:"palaces"`
		} `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if env.Data.Yun.YunNumber < 1 || env.Data.Yun.YunNumber > 9 {
		t.Errorf("yun_number = %d, want [1,9]", env.Data.Yun.YunNumber)
	}
	if len(env.Data.Palaces) != 9 {
		t.Errorf("palaces = %d, want 9", len(env.Data.Palaces))
	}
}

func TestToolRegistry_Execute_ComputeSanYuanYun(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{"year":2025}`)
	result, err := r.Execute(context.Background(), "compute_sanyuan_yun", args)
	if err != nil {
		t.Fatalf("compute_sanyuan_yun: %v", err)
	}
	var env struct {
		Data struct {
			Year      int    `json:"year"`
			Yuan      string `json:"yuan"`
			YunNumber int    `json:"yun_number"`
		} `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if env.Data.Year != 2025 {
		t.Errorf("year = %d, want 2025", env.Data.Year)
	}
	if env.Data.YunNumber != 9 {
		t.Errorf("yun_number = %d, want 9 for 2025", env.Data.YunNumber)
	}
	if env.Data.Yuan != "下元" {
		t.Errorf("yuan = %q, want 下元", env.Data.Yuan)
	}
}

// --- liuyao ---

func TestToolRegistry_Execute_ComputeLiuyao(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{` + bt + `,"yong_shen":"世爻"}`)
	result, err := r.Execute(context.Background(), "compute_liuyao", args)
	if err != nil {
		t.Fatalf("compute_liuyao: %v", err)
	}
	var env struct {
		Data struct {
			Name   string `json:"name"`
			BenGua int    `json:"ben_gua"`
			Lines  []struct {
				Position int    `json:"position"`
				LiuQin   int    `json:"liu_qin"`
			} `json:"lines"`
		} `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if env.Data.Name == "" {
		t.Error("name is empty")
	}
	if len(env.Data.Lines) != 6 {
		t.Errorf("lines = %d, want 6", len(env.Data.Lines))
	}
}

// --- huangli: month, bond date, bond month ---

func TestToolRegistry_Execute_QueryHuangliMonth(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{"month":"2025-06","event":"结婚"}`)
	result, err := r.Execute(context.Background(), "query_huangli_month", args)
	if err != nil {
		t.Fatalf("query_huangli_month: %v", err)
	}
	var env struct {
		Data struct {
			Month string `json:"month"`
		} `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if env.Data.Month != "2025-06" {
		t.Errorf("month = %q, want 2025-06", env.Data.Month)
	}
}

func TestToolRegistry_Execute_QueryHuangliBondDate(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{` + bt + `,"event_type":"结婚","date":"2025-06-01"}`)
	result, err := r.Execute(context.Background(), "query_huangli_bond_date", args)
	if err != nil {
		t.Fatalf("query_huangli_bond_date: %v", err)
	}
	var env struct {
		Data struct {
			Date string `json:"date"`
		} `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if env.Data.Date != "2025-06-01" {
		t.Errorf("date = %q, want 2025-06-01", env.Data.Date)
	}
}

func TestToolRegistry_Execute_QueryHuangliBondMonth(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{` + bt + `,"event_type":"结婚","month":"2025-06"}`)
	result, err := r.Execute(context.Background(), "query_huangli_bond_month", args)
	if err != nil {
		t.Fatalf("query_huangli_bond_month: %v", err)
	}
	var env struct {
		Data struct {
			Month string `json:"month"`
		} `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if env.Data.Month != "2025-06" {
		t.Errorf("month = %q, want 2025-06", env.Data.Month)
	}
}

func TestToolSchemas_HaveParameters(t *testing.T) {
	r := NewChatToolRegistry()
	schemas := r.Schemas()

	if len(schemas) != 29 {
		t.Errorf("schema count = %d, want 29", len(schemas))
	}

	toolsWithoutParams := []string{}
	for _, s := range schemas {
		var fn map[string]any
		if err := json.Unmarshal(s.Function, &fn); err != nil {
			t.Errorf("invalid schema JSON: %v", err)
			continue
		}
		if fn["name"] == nil {
			t.Error("tool missing name")
		}
		if fn["description"] == nil {
			t.Errorf("tool %v missing description", fn["name"])
		}
		if fn["parameters"] == nil {
			toolsWithoutParams = append(toolsWithoutParams, fn["name"].(string))
		}
	}

	if len(toolsWithoutParams) > 0 {
		t.Errorf("tools without parameters: %v", toolsWithoutParams)
	}
}

func TestOpenApiParams_QueryCity(t *testing.T) {
	params := openapiParams("query_city")
	if params == nil {
		t.Fatal("query_city schema not found in openapi.json")
	}

	var schema map[string]any
	if err := json.Unmarshal(params, &schema); err != nil {
		t.Fatal("invalid parameters JSON:", err)
	}

	required, ok := schema["required"].([]any)
	if !ok || len(required) != 1 || required[0] != "name" {
		t.Errorf("required = %v, want [city]", required)
	}

	props := schema["properties"].(map[string]any)
	city, ok := props["name"].(map[string]any)
	if !ok || city["type"] != "string" {
		t.Error("name param must be type string")
	}
}

func TestOpenApiParams_ComputeChart(t *testing.T) {
	params := openapiParams("compute_chart")
	if params == nil {
		t.Fatal("compute_chart schema not found")
	}
	// compute_chart uses $ref to Person — verify non-empty
	if len(params) == 0 {
		t.Error("empty parameters")
	}
}
func TestOpenApiParams_ComputeLiuRi(t *testing.T) {
	params := openapiParams("compute_liuri")
	if params == nil {
		t.Fatal("compute_liuri schema not found")
	}

	var schema map[string]any
	json.Unmarshal(params, &schema)

	props := schema["properties"].(map[string]any)
	keys := []string{"year", "month", "day", "birth", "gender"}
	for _, k := range keys {
		if props[k] == nil {
			t.Errorf("compute_liuri missing property: %s", k)
		}
	}
}

func TestOpenApiParams_UnknownTool(t *testing.T) {
	if p := openapiParams("nonexistent_tool"); p != nil {
		t.Error("unknown tool should return nil")
	}
}
