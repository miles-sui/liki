package agent

import (
	"context"
	"encoding/json"
	"testing"
)

const btAgent = `"birth":{"time":"1984-02-15T08:00:00+08:00","longitude":116.4}`

// ============================================================
// BUG HUNT: agent tool handlers have NO input validation.
// Unlike HTTP handlers (which use decodeAndValidate), the
// agent tool handlers just json.Unmarshal and call engines.
// Invalid inputs may produce garbage results or crash.
// ============================================================

// --- compute_chart: invalid gender ---

func TestBugAgent_ComputeChart_InvalidGender(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{` + btAgent + `,"gender":"other"}`)
	result, err := r.Execute(context.Background(), "compute_chart", args)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	// BUG CONFIRMED: invalid gender silently accepted.
	var env struct {
		Product string `json:"_product"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if env.Product == "chart" {
		t.Log("BUG CONFIRMED: compute_chart accepts invalid gender='other'")
	}
}

func TestBugAgent_ComputeChart_EmptyGender(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{` + btAgent + `,"gender":""}`)
	result, err := r.Execute(context.Background(), "compute_chart", args)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	var env struct {
		Product string `json:"_product"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if env.Product == "chart" {
		t.Log("BUG CONFIRMED: compute_chart accepts empty gender=''")
	}
}

func TestBugAgent_ComputeChart_MissingBirth(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{"gender":"male"}`)
	_, err := r.Execute(context.Background(), "compute_chart", args)
	if err == nil {
		t.Log("BUG CONFIRMED: compute_chart accepts missing birth")
	} else {
		t.Logf("OK: compute_chart rejects missing birth: %v", err)
	}
}

// --- compute_chart: birth with empty time string ---

func TestBugAgent_ComputeChart_EmptyTime(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{"birth":{"time":"","longitude":116.4},"gender":"male"}`)
	_, err := r.Execute(context.Background(), "compute_chart", args)
	if err == nil {
		t.Log("BUG CONFIRMED: compute_chart accepts empty birth.time")
	} else {
		t.Logf("OK: rejects empty time: %v", err)
	}
}

// --- compute_bond: missing b ---

func TestBugAgent_ComputeBond_MissingB(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{"a":{` + btAgent + `,"gender":"male"}}`)
	_, err := r.Execute(context.Background(), "compute_bond", args)
	if err == nil {
		t.Log("BUG CONFIRMED: compute_bond accepts missing 'b' (zero-value birth)")
	} else {
		t.Logf("OK: rejects missing b: %v", err)
	}
}

// --- compute_bond: missing a ---

func TestBugAgent_ComputeBond_MissingA(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{"b":{` + btAgent + `,"gender":"female"}}`)
	_, err := r.Execute(context.Background(), "compute_bond", args)
	if err == nil {
		t.Log("BUG CONFIRMED: compute_bond accepts missing 'a'")
	} else {
		t.Logf("OK: rejects missing a: %v", err)
	}
}

// --- compute_liunian: missing year ---

func TestBugAgent_ComputeLiunian_MissingYear(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{` + btAgent + `}`)
	_, err := r.Execute(context.Background(), "compute_liunian", args)
	if err == nil {
		t.Error("compute_liunian should reject missing year (year=0)")
	}
}

// --- compute_liunian: negative year ---

func TestBugAgent_ComputeLiunian_NegativeYear(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{"year":-1,` + btAgent + `}`)
	result, err := r.Execute(context.Background(), "compute_liunian", args)
	if err != nil {
		t.Logf("compute_liunian year=-1 error: %v", err)
		return
	}
	var env struct {
		Product string `json:"_product"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if env.Product == "liunian" {
		t.Log("BUG CONFIRMED: compute_liunian accepts negative year=-1")
	}
}

// --- compute_liuyue: missing month ---

func TestBugAgent_ComputeLiuyue_MissingMonth(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{"year":2025,` + btAgent + `}`)
	_, err := r.Execute(context.Background(), "compute_liuyue", args)
	if err == nil {
		t.Error("compute_liuyue should reject missing month (month=0)")
	}
}

// --- compute_liushi: hour out of range ---

func TestBugAgent_ComputeLiushi_HourOutOfRange(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{"date":"2025-06-15","hour":25,"birth":{"time":"1984-02-15T08:00:00+08:00","longitude":116.4}}`)
	result, err := r.Execute(context.Background(), "compute_liushi", args)
	if err != nil {
		t.Logf("compute_liushi hour=25 error: %v", err)
		return
	}
	var env struct {
		Product string `json:"_product"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if env.Product == "liushi" {
		t.Log("BUG CONFIRMED: compute_liushi accepts hour=25 (out of range)")
	}
}

// --- compute_liuyao: invalid yong_shen ---

func TestBugAgent_ComputeLiuyao_InvalidYongShen(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{` + btAgent + `,"yong_shen":"invalid"}`)
	result, err := r.Execute(context.Background(), "compute_liuyao", args)
	if err != nil {
		t.Logf("compute_liuyao invalid yong_shen error: %v", err)
		return
	}
	// BUG: invalid yong_shen maps to zero-value YongShen(0), which is NOT in yongShenMap
	// So it will pass YongShen(0) to liuyao.ComputeChart — is this valid?
	var env struct {
		Product string          `json:"_product"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if env.Product == "liuyao" {
		var chart struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(env.Data, &chart); err != nil {
			t.Fatal(err)
		}
		t.Logf("BUG CONFIRMED: compute_liuyao accepts invalid yong_shen='invalid', name=%q", chart.Name)
	}
}

// --- compute_xiaoyun: missing count ---

func TestBugAgent_ComputeXiaoyun_MissingCount(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{` + btAgent + `,"gender":"male"}`)
	result, err := r.Execute(context.Background(), "compute_xiaoyun", args)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	var env struct {
		Product string          `json:"_product"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	// count=0 — engine may return empty slice or compute default
	var pillars []json.RawMessage
	if err := json.Unmarshal(env.Data, &pillars); err != nil {
		t.Fatal(err)
	}
	t.Logf("compute_xiaoyun missing count: results=%d", len(pillars))
}

// --- compute_xiaoxian: missing count ---

func TestBugAgent_ComputeXiaoxian_MissingCount(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{"gender":"female"}`)
	result, err := r.Execute(context.Background(), "compute_xiaoxian", args)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	var env struct {
		Product string          `json:"_product"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	var entries []json.RawMessage
	if err := json.Unmarshal(env.Data, &entries); err != nil {
		t.Fatal(err)
	}
	t.Logf("compute_xiaoxian missing count: results=%d", len(entries))
}

// --- compute_ziwei_bond: missing b ---

func TestBugAgent_ComputeZiweiBond_MissingB(t *testing.T) {
	r := NewChatToolRegistry()
	// Need a valid chart for A
	chartArgs := json.RawMessage(`{` + btAgent + `,"gender":"male"}`)
	chartResult, err := r.Execute(context.Background(), "compute_ziwei", chartArgs)
	if err != nil {
		t.Fatalf("setup ziwei chart: %v", err)
	}
	var chartEnv struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(chartResult, &chartEnv); err != nil {
		t.Fatal(err)
	}

	args := json.RawMessage(`{"a":` + string(chartEnv.Data) + `}`)
	result, err := r.Execute(context.Background(), "compute_ziwei_bond", args)
	if err != nil {
		t.Logf("compute_ziwei_bond missing b error: %v", err)
		return
	}
	var env struct {
		Product string `json:"_product"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if env.Product == "ziwei_bond" {
		t.Log("BUG CONFIRMED: compute_ziwei_bond accepts missing 'b'")
	}
}

// --- compute_ziwei_bond: empty charts ---

func TestBugAgent_ComputeZiweiBond_EmptyCharts(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{"a":{},"b":{}}`)
	result, err := r.Execute(context.Background(), "compute_ziwei_bond", args)
	if err != nil {
		t.Logf("compute_ziwei_bond empty charts error: %v", err)
		return
	}
	var env struct {
		Product string `json:"_product"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if env.Product == "ziwei_bond" {
		t.Log("BUG CONFIRMED: compute_ziwei_bond accepts empty charts {}")
	}
}

// --- compute_ziwei_daxian: empty chart ---

func TestBugAgent_ComputeZiweiDaxian_EmptyChart(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{"chart":{}}`)
	result, err := r.Execute(context.Background(), "compute_ziwei_daxian", args)
	if err != nil {
		t.Logf("compute_ziwei_daxian empty chart error: %v", err)
		return
	}
	var env struct {
		Product string          `json:"_product"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	var steps []json.RawMessage
	if err := json.Unmarshal(env.Data, &steps); err != nil {
		t.Fatal(err)
	}
	t.Logf("compute_ziwei_daxian empty chart: steps=%d", len(steps))
}

// --- compute_ziwei: invalid gender ---

func TestBugAgent_ComputeZiwei_InvalidGender(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{` + btAgent + `,"gender":"unknown"}`)
	result, err := r.Execute(context.Background(), "compute_ziwei", args)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	var env struct {
		Product string `json:"_product"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if env.Product == "ziwei" {
		t.Log("BUG CONFIRMED: compute_ziwei accepts invalid gender='unknown'")
	}
}

// --- compute_qimen: invalid kind ---

func TestBugAgent_ComputeQimen_InvalidKind(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{"birth":{"time":"1984-02-15T08:00:00+08:00","longitude":116.4},"kind":"xun"}`)
	result, err := r.Execute(context.Background(), "compute_qimen", args)
	if err != nil {
		t.Logf("compute_qimen kind=xun error: %v", err)
		return
	}
	var env struct {
		Product string `json:"_product"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if env.Product == "qimen" {
		t.Log("BUG CONFIRMED: compute_qimen accepts invalid kind='xun'")
	}
}

// --- compute_minggua: invalid gender ---

func TestBugAgent_ComputeMingGua_InvalidGender(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{"gender":"unknown","birth_year":1990}`)
	result, err := r.Execute(context.Background(), "compute_minggua", args)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	var env struct {
		Product string `json:"_product"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if env.Product == "minggua" {
		t.Log("BUG CONFIRMED: compute_minggua accepts invalid gender='unknown'")
	}
}

// --- compute_minggua: missing birth_year ---

func TestBugAgent_ComputeMingGua_MissingBirthYear(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{"gender":"male"}`)
	result, err := r.Execute(context.Background(), "compute_minggua", args)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	var env struct {
		Product string          `json:"_product"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	var mg struct {
		GuaNumber int `json:"gua_number"`
	}
	if err := json.Unmarshal(env.Data, &mg); err != nil {
		t.Fatal(err)
	}
	t.Logf("compute_minggua birth_year=0: gua_number=%d", mg.GuaNumber)
}

// --- compute_xuankong: missing mountains ---

func TestBugAgent_ComputeXuankong_MissingMountains(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{` + btAgent + `}`)
	result, err := r.Execute(context.Background(), "compute_xuankong", args)
	if err != nil {
		t.Logf("compute_xuankong missing mountains error: %v", err)
		return
	}
	var env struct {
		Product string          `json:"_product"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if env.Product == "xuankong" {
		var chart struct {
			Palaces []json.RawMessage `json:"palaces"`
		}
		if err := json.Unmarshal(env.Data, &chart); err != nil {
			t.Fatal(err)
		}
		t.Logf("BUG CONFIRMED: compute_xuankong accepts sit=0,face=0, palaces=%d", len(chart.Palaces))
	}
}

// --- compute_sanyuan_yun: missing year ---

func TestBugAgent_ComputeSanYuanYun_MissingYear(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf("CRASH BUG: compute_sanyuan_yun panics on year=0: %v", r)
			return
		}
	}()
	r := NewChatToolRegistry()
	args := json.RawMessage(`{}`)
	result, err := r.Execute(context.Background(), "compute_sanyuan_yun", args)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	var env struct {
		Product string          `json:"_product"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	var yun struct {
		YunNumber int `json:"YunNumber"`
		StartYear int `json:"StartYear"`
	}
	if err := json.Unmarshal(env.Data, &yun); err != nil {
		t.Fatal(err)
	}
	t.Logf("compute_sanyuan_yun year=0: YunNumber=%d, StartYear=%d", yun.YunNumber, yun.StartYear)
}

// --- compute_naming_wuge: empty surname ---

func TestBugAgent_ComputeNamingWuge_EmptySurname(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{"surname":"","yong_shen":"木","xi_shen":["水"]}`)
	_, err := r.Execute(context.Background(), "compute_naming_wuge", args)
	if err == nil {
		t.Log("BUG CONFIRMED: compute_naming_wuge accepts empty surname")
	} else {
		t.Logf("OK: rejects empty surname: %v", err)
	}
}

// --- compute_naming_wuge: invalid yong_shen ---

func TestBugAgent_ComputeNamingWuge_InvalidYongShen(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{"surname":"张","yong_shen":"x","xi_shen":["水"]}`)
	_, err := r.Execute(context.Background(), "compute_naming_wuge", args)
	if err == nil {
		t.Log("BUG CONFIRMED: compute_naming_wuge accepts invalid yong_shen='x'")
	} else {
		t.Logf("OK: rejects invalid yong_shen: %v", err)
	}
}

// --- compute_naming_evaluate: empty given_name ---

func TestBugAgent_ComputeNamingEvaluate_EmptyGivenName(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{"surname":"张","given_name":"","yong_shen":"木"}`)
	_, err := r.Execute(context.Background(), "compute_naming_evaluate", args)
	if err == nil {
		t.Log("BUG CONFIRMED: compute_naming_evaluate accepts empty given_name")
	} else {
		t.Logf("OK: rejects empty given_name: %v", err)
	}
}

// --- query_huangli_date: missing event ---

func TestBugAgent_QueryHuangliDate_MissingEvent(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{"date":"2025-06-01"}`)
	_, err := r.Execute(context.Background(), "query_huangli_date", args)
	if err == nil {
		t.Log("BUG CONFIRMED: query_huangli_date accepts missing event")
	} else {
		t.Logf("OK: rejects missing event: %v", err)
	}
}

// --- query_huangli_bond_date: missing birth ---

func TestBugAgent_QueryHuangliBondDate_MissingBirth(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{"event_type":"结婚","date":"2025-06-01"}`)
	_, err := r.Execute(context.Background(), "query_huangli_bond_date", args)
	if err == nil {
		t.Log("BUG CONFIRMED: query_huangli_bond_date accepts missing birth")
	} else {
		t.Logf("OK: rejects missing birth: %v", err)
	}
}

// --- unknown tool name ---

func TestBugAgent_UnknownTool(t *testing.T) {
	r := NewChatToolRegistry()
	_, err := r.Execute(context.Background(), "nonexistent_tool", json.RawMessage(`{}`))
	if err == nil {
		t.Error("BUG: unknown tool should return error")
	}
}

// --- compute_bazhai: missing gender ---

func TestBugAgent_ComputeBazhai_MissingGender(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{` + btAgent + `}`)
	result, err := r.Execute(context.Background(), "compute_bazhai", args)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	var env struct {
		Product string          `json:"_product"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if env.Product == "bazhai" {
		var chart struct {
			MingGua struct {
				GuaNumber int `json:"gua_number"`
			} `json:"ming_gua"`
		}
		if err := json.Unmarshal(env.Data, &chart); err != nil {
			t.Fatal(err)
		}
		t.Logf("BUG CONFIRMED: compute_bazhai missing gender, gua_number=%d", chart.MingGua.GuaNumber)
	}
}
