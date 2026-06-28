package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"liki/internal/engine/ganzhi"
)

const btOK = `{"time":"1984-02-15T08:00:00+08:00","longitude":116.4}`
const btOK2 = `{"time":"1986-08-20T12:00:00+08:00","longitude":121.5}`

// ── helpers ──

func hasKey(raw json.RawMessage, key string) bool {
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return false
	}
	_, ok := m[key]
	return ok
}

func getStr(raw json.RawMessage, path ...string) string {
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return ""
	}
	cur := any(m)
	for _, k := range path {
		cm, ok := cur.(map[string]any)
		if !ok {
			return ""
		}
		cur = cm[k]
	}
	s, _ := cur.(string)
	return s
}

// ── gender validation ──

func TestValidateGender(t *testing.T) {
	tests := []struct {
		name    string
		gender  string
		wantErr bool
	}{
		{"male valid", "male", false},
		{"female valid", "female", false},
		{"空字符串", "", true},
		{"无效值 x", "x", true},
		{"无效值 other", "other", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateGender(ganzhi.Gender(tt.gender))
			if (err != nil) != tt.wantErr {
				t.Errorf("validateGender(%q) err=%v, wantErr=%v", tt.gender, err, tt.wantErr)
			}
		})
	}
}

// ── TimePoint.Timeset ──

func TestTimePoint_Timeset(t *testing.T) {
	tests := []struct {
		name    string
		time    string
		wantErr bool
	}{
		{"有效时间", "1984-02-15T08:00:00+08:00", false},
		{"无效格式", "1984-02-15", true},
		{"空字符串", "", true},
		{"无时区", "1984-02-15T08:00:00", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := TimePoint{Time: tt.time, Longitude: 116.4}
			_, err := tp.Timeset()
			if (err != nil) != tt.wantErr {
				t.Errorf("Timeset(%q) err=%v, wantErr=%v", tt.time, err, tt.wantErr)
			}
		})
	}
}

// ── handler error paths ──

func TestHandler_InvalidJSON(t *testing.T) {
	r := NewRPCRegistry()
	handlers := []string{
		"bazi.chart", "bazi.bond", "bazi.liunian", "bazi.liuyue",
		"bazi.liuri", "bazi.liushi", "bazi.xiaoyun", "bazi.xiaoxian",
		"ziwei.chart", "ziwei.daxian", "ziwei.liunian", "ziwei.liuyue",
		"ziwei.liuri", "ziwei.bond",
		"qimen.pan",
		"qiming.wuge", "qiming.compose", "qiming.detail", "qiming.evaluate",
		"bazhai.chart", "bazhai.minggua",
		"xuankong.sanyuan", "xuankong.chart",
		"liuyao.chart",
		"huangli.date", "huangli.month", "huangli.bond.date", "huangli.bond.month",
		"city",
	}
	badJSON := json.RawMessage(`{bad`)
	for _, name := range handlers {
		t.Run(name, func(t *testing.T) {
			_, err := r.Execute(context.Background(), name, badJSON)
			if err == nil {
				t.Error("expected error for invalid JSON")
			}
		})
	}
}

func TestHandler_MissingGender(t *testing.T) {
	r := NewRPCRegistry()
	handlers := []string{
		"bazi.chart", "ziwei.chart", "bazhai.chart",
		"bazi.xiaoyun",
	}
	noGender := json.RawMessage(fmt.Sprintf(`{"birth":%s}`, btOK))
	for _, name := range handlers {
		t.Run(name, func(t *testing.T) {
			_, err := r.Execute(context.Background(), name, noGender)
			if err == nil {
				t.Errorf("%s: expected error for missing gender", name)
			}
		})
	}
}

func TestHandler_BadGender(t *testing.T) {
	r := NewRPCRegistry()
	handlers := []struct {
		name   string
		params string
	}{
		{"bazi.chart", fmt.Sprintf(`{"birth":%s,"gender":"other"}`, btOK)},
		{"bazi.bond", fmt.Sprintf(`{"a":{"birth":%s,"gender":"x"},"b":{"birth":%s,"gender":"female"}}`, btOK, btOK2)},
		{"bazi.liushi", fmt.Sprintf(`{"year":2026,"month":6,"day":15,"hour":12,"birth":%s,"gender":"bad"}`, btOK)},
		{"bazi.xiaoxian", `{"gender":"bad"}`},
		{"ziwei.chart", fmt.Sprintf(`{"birth":%s,"gender":"bad"}`, btOK)},
		{"bazhai.chart", fmt.Sprintf(`{"birth":%s,"gender":"bad"}`, btOK)},
		{"bazhai.minggua", `{"gender":"x","birth_year":1984}`},
	}
	for _, tt := range handlers {
		t.Run(tt.name, func(t *testing.T) {
			_, err := r.Execute(context.Background(), tt.name, json.RawMessage(tt.params))
			if err == nil {
				t.Error("expected error for invalid gender")
			}
		})
	}
}

func TestHandler_MissingRequiredFields(t *testing.T) {
	r := NewRPCRegistry()
	tests := []struct {
		name   string
		params string
	}{
		{"bazi.bond", fmt.Sprintf(`{"a":{"birth":%s,"gender":"male"}}`, btOK)},
		{"bazi.liunian", fmt.Sprintf(`{"birth":%s,"gender":"male"}`, btOK)},
		{"bazi.liuyue", fmt.Sprintf(`{"year":2026,"birth":%s,"gender":"male"}`, btOK)},
		{"bazi.xiaoxian", `{}`},
		{"ziwei.daxian", `{}`},
		{"ziwei.bond", fmt.Sprintf(`{"a":{"birth":%s,"gender":"male"}}`, btOK)},
		{"qiming.wuge", `{"surname":"王"}`},
		{"qiming.evaluate", `{"surname":"王"}`},
		{"xuankong.chart", fmt.Sprintf(`{"birth":%s}`, btOK)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := r.Execute(context.Background(), tt.name, json.RawMessage(tt.params))
			if err == nil {
				t.Errorf("%s: expected error for missing required fields", tt.name)
			}
		})
	}
}

func TestHandler_RangeValidation(t *testing.T) {
	r := NewRPCRegistry()
	tests := []struct {
		name   string
		params string
	}{
		{"bazi.liunian", fmt.Sprintf(`{"year":0,"birth":%s,"gender":"male"}`, btOK)},
		{"bazi.liushi", fmt.Sprintf(`{"year":2026,"month":6,"day":15,"hour":-1,"birth":%s,"gender":"male"}`, btOK)},
		{"bazi.liushi", fmt.Sprintf(`{"year":2026,"month":6,"day":15,"hour":24,"birth":%s,"gender":"male"}`, btOK)},
		{"qimen.pan", fmt.Sprintf(`{"birth":%s,"kind":"invalid"}`, btOK)},
		{"xuankong.chart", fmt.Sprintf(`{"birth":%s,"sit_mountain":-1,"face_mountain":0}`, btOK)},
		{"xuankong.chart", fmt.Sprintf(`{"birth":%s,"sit_mountain":0,"face_mountain":24}`, btOK)},
		{"xuankong.sanyuan", `{"year":0}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := r.Execute(context.Background(), tt.name, json.RawMessage(tt.params))
			if err == nil {
				t.Errorf("%s: expected error for invalid range", tt.name)
			}
		})
	}
}

func TestHandler_InvalidWuxing(t *testing.T) {
	r := NewRPCRegistry()
	tests := []struct {
		name   string
		params string
	}{
		{"qiming.wuge", `{"surname":"王","yong_shen":"木星"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := r.Execute(context.Background(), tt.name, json.RawMessage(tt.params))
			if err == nil {
				t.Errorf("%s: expected error for invalid wuxing", tt.name)
			}
		})
	}
}

func TestHandler_QimenKindDefault(t *testing.T) {
	r := NewRPCRegistry()
	params := json.RawMessage(fmt.Sprintf(`{"birth":%s}`, btOK))
	result, err := r.Execute(context.Background(), "qimen.pan", params)
	if err != nil {
		t.Fatalf("qimen.pan (default kind): %v", err)
	}
	if !hasKey(result, "_product") || !hasKey(result, "data") {
		t.Error("expected envelope with _product and data")
	}
}

func TestHandler_LiuyaoYongShenDefault(t *testing.T) {
	r := NewRPCRegistry()
	params := json.RawMessage(fmt.Sprintf(`{"birth":%s}`, btOK))
	result, err := r.Execute(context.Background(), "liuyao.chart", params)
	if err != nil {
		t.Fatalf("liuyao.chart (default yong_shen): %v", err)
	}
	if getStr(result, "_product") != "liuyao" {
		t.Errorf("_product = %q, want liuyao", getStr(result, "_product"))
	}
}

// ── handler valid paths (envelope + data) ──

func TestHandler_ComputeChart_Valid(t *testing.T) {
	r := NewRPCRegistry()
	params := json.RawMessage(fmt.Sprintf(`{"birth":%s,"gender":"male"}`, btOK))
	result, err := r.Execute(context.Background(), "bazi.chart", params)
	if err != nil {
		t.Fatalf("bazi.chart: %v", err)
	}
	if getStr(result, "_product") != "chart" {
		t.Errorf("_product = %q, want chart", getStr(result, "_product"))
	}
	if !hasKey(result, "data") {
		t.Error("missing data")
	}
}

func TestHandler_ComputeBond_Valid(t *testing.T) {
	r := NewRPCRegistry()
	params := json.RawMessage(fmt.Sprintf(
		`{"a":{"birth":%s,"gender":"male"},"b":{"birth":%s,"gender":"female"}}`,
		btOK, btOK2))
	result, err := r.Execute(context.Background(), "bazi.bond", params)
	if err != nil {
		t.Fatalf("bazi.bond: %v", err)
	}
	if getStr(result, "_product") != "bond" {
		t.Errorf("_product = %q, want bond", getStr(result, "_product"))
	}
}

func TestHandler_ComputeLiunian_Valid(t *testing.T) {
	r := NewRPCRegistry()
	params := json.RawMessage(fmt.Sprintf(`{"year":2026,"birth":%s,"gender":"male"}`, btOK))
	result, err := r.Execute(context.Background(), "bazi.liunian", params)
	if err != nil {
		t.Fatalf("bazi.liunian: %v", err)
	}
	if getStr(result, "_product") != "liunian" {
		t.Errorf("_product = %q, want liunian", getStr(result, "_product"))
	}
}

func TestHandler_ComputeXiaoXian_Valid(t *testing.T) {
	r := NewRPCRegistry()
	result, err := r.Execute(context.Background(), "bazi.xiaoxian", json.RawMessage(`{"gender":"male"}`))
	if err != nil {
		t.Fatalf("bazi.xiaoxian: %v", err)
	}
	if getStr(result, "_product") != "xiaoxian" {
		t.Errorf("_product = %q, want xiaoxian", getStr(result, "_product"))
	}
}

func TestHandler_ComputeZiwei_Valid(t *testing.T) {
	r := NewRPCRegistry()
	params := json.RawMessage(fmt.Sprintf(`{"birth":%s,"gender":"male"}`, btOK))
	result, err := r.Execute(context.Background(), "ziwei.chart", params)
	if err != nil {
		t.Fatalf("ziwei.chart: %v", err)
	}
	if getStr(result, "_product") != "ziwei" {
		t.Errorf("_product = %q, want ziwei", getStr(result, "_product"))
	}
}

func TestHandler_ComputeMingGua_Valid(t *testing.T) {
	r := NewRPCRegistry()
	result, err := r.Execute(context.Background(), "bazhai.minggua", json.RawMessage(`{"gender":"male","birth_year":1984}`))
	if err != nil {
		t.Fatalf("bazhai.minggua: %v", err)
	}
	if getStr(result, "_product") != "minggua" {
		t.Errorf("_product = %q, want minggua", getStr(result, "_product"))
	}
}

func TestHandler_ComputeSanYuanYun_Valid(t *testing.T) {
	r := NewRPCRegistry()
	result, err := r.Execute(context.Background(), "xuankong.sanyuan", json.RawMessage(`{"year":2026}`))
	if err != nil {
		t.Fatalf("xuankong.sanyuan: %v", err)
	}
	if getStr(result, "_product") != "sanyuan_yun" {
		t.Errorf("_product = %q, want sanyuan_yun", getStr(result, "_product"))
	}
}

func TestHandler_QueryHuangliDate_Valid(t *testing.T) {
	r := NewRPCRegistry()
	result, err := r.Execute(context.Background(), "huangli.date", json.RawMessage(`{"date":"2026-06-26","event":"嫁娶"}`))
	if err != nil {
		t.Fatalf("huangli.date: %v", err)
	}
	if getStr(result, "_product") != "huangli_date" {
		t.Errorf("_product = %q, want huangli_date", getStr(result, "_product"))
	}
}

func TestHandler_QueryHuangliBondDate_Valid(t *testing.T) {
	r := NewRPCRegistry()
	params := json.RawMessage(fmt.Sprintf(`{"birth":%s,"event_type":"嫁娶","date":"2026-06-26"}`, btOK))
	result, err := r.Execute(context.Background(), "huangli.bond.date", params)
	if err != nil {
		t.Fatalf("huangli.bond.date: %v", err)
	}
	if getStr(result, "_product") != "huangli_bond_date" {
		t.Errorf("_product = %q, want huangli_bond_date", getStr(result, "_product"))
	}
}

func TestHandler_HuangliDate_MissingEvent(t *testing.T) {
	r := NewRPCRegistry()
	_, err := r.Execute(context.Background(), "huangli.date", json.RawMessage(`{"date":"2026-06-26"}`))
	if err == nil {
		t.Error("expected error for missing event")
	}
}

func TestHandler_HuangliBondDate_EmptyBirth(t *testing.T) {
	r := NewRPCRegistry()
	_, err := r.Execute(context.Background(), "huangli.bond.date", json.RawMessage(`{"birth":{},"event_type":"嫁娶","date":"2026-06-26"}`))
	if err == nil {
		t.Error("expected error for empty birth")
	}
}

func TestHandler_AllHandlersAcceptContext(t *testing.T) {
	r := NewRPCRegistry()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	params := json.RawMessage(`{"gender":"male"}`)
	_, err := r.Execute(ctx, "bazi.xiaoxian", params)
	if err != nil {
		t.Logf("bazi.xiaoxian with canceled ctx: %v", err)
	}
}

// ── valid paths for handlers with low coverage ──

func TestHandler_ComputeXiaoYun_Valid(t *testing.T) {
	r := NewRPCRegistry()
	params := json.RawMessage(fmt.Sprintf(`{"birth":%s,"gender":"male","count":5}`, btOK))
	result, err := r.Execute(context.Background(), "bazi.xiaoyun", params)
	if err != nil {
		t.Fatalf("bazi.xiaoyun: %v", err)
	}
	if getStr(result, "_product") != "xiaoyun" {
		t.Errorf("_product = %q, want xiaoyun", getStr(result, "_product"))
	}
	if !hasKey(result, "data") {
		t.Error("missing data")
	}
}

func TestHandler_ComputeXiaoYun_DefaultCount(t *testing.T) {
	r := NewRPCRegistry()
	params := json.RawMessage(fmt.Sprintf(`{"birth":%s,"gender":"male"}`, btOK))
	result, err := r.Execute(context.Background(), "bazi.xiaoyun", params)
	if err != nil {
		t.Fatalf("bazi.xiaoyun (default count): %v", err)
	}
	if getStr(result, "_product") != "xiaoyun" {
		t.Errorf("_product = %q, want xiaoyun", getStr(result, "_product"))
	}
}

func TestHandler_ComputeLiushi_Valid(t *testing.T) {
	r := NewRPCRegistry()
	params := json.RawMessage(fmt.Sprintf(
		`{"year":2026,"month":6,"day":15,"hour":12,"birth":%s,"gender":"male"}`, btOK))
	result, err := r.Execute(context.Background(), "bazi.liushi", params)
	if err != nil {
		t.Fatalf("bazi.liushi: %v", err)
	}
	if getStr(result, "_product") != "liushi" {
		t.Errorf("_product = %q, want liushi", getStr(result, "_product"))
	}
	if !hasKey(result, "data") {
		t.Error("missing data")
	}
}

func TestHandler_QueryHuangliMonth_Valid(t *testing.T) {
	r := NewRPCRegistry()
	result, err := r.Execute(context.Background(), "huangli.month",
		json.RawMessage(`{"month":"2026-06","event":"嫁娶"}`))
	if err != nil {
		t.Fatalf("huangli.month: %v", err)
	}
	if getStr(result, "_product") != "huangli_month" {
		t.Errorf("_product = %q, want huangli_month", getStr(result, "_product"))
	}
}

func TestHandler_QueryHuangliBondMonth_Valid(t *testing.T) {
	r := NewRPCRegistry()
	params := json.RawMessage(fmt.Sprintf(
		`{"birth":%s,"event_type":"嫁娶","month":"2026-06"}`, btOK))
	result, err := r.Execute(context.Background(), "huangli.bond.month", params)
	if err != nil {
		t.Fatalf("huangli.bond.month: %v", err)
	}
	if getStr(result, "_product") != "huangli_bond_month" {
		t.Errorf("_product = %q, want huangli_bond_month", getStr(result, "_product"))
	}
}

// ── OpenRPC document ──

func TestOpenRPCDocument(t *testing.T) {
	r := NewRPCRegistry()
	doc := r.OpenRPCDocument()

	var raw map[string]any
	if err := json.Unmarshal(doc, &raw); err != nil {
		t.Fatalf("OpenRPCDocument is not valid JSON: %v", err)
	}
	if raw["openrpc"] != "1.4.1" {
		t.Errorf("openrpc version = %v, want 1.4.1", raw["openrpc"])
	}
	methods, ok := raw["methods"].([]any)
	if !ok {
		t.Fatal("missing methods array")
	}
	if len(methods) != 30 {
		t.Errorf("method count = %d, want 30 (29 + rpc.discover)", len(methods))
	}
}

// ── wrapResult ──

func TestWrapResult(t *testing.T) {
	result, err := wrapResult("test", map[string]any{"key": "value"})
	if err != nil {
		t.Fatal(err)
	}
	if getStr(result, "_product") != "test" {
		t.Errorf("_product = %q, want test", getStr(result, "_product"))
	}
	if !hasKey(result, "data") {
		t.Error("missing data field")
	}
}
