package agent

import (
	"context"
	"encoding/json"
	"testing"
)

const btAgent = `"birth":{"time":"1984-02-15T08:00:00+08:00","longitude":116.4}`

// ============================================================
// Regression tests: agent tool input validation.
// Each test sends invalid input and asserts the tool rejects it.
// ============================================================

// --- compute_chart: invalid gender ---

func TestBugAgent_ComputeChart_InvalidGender(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{` + btAgent + `,"gender":"other"}`)
	_, err := r.Execute(context.Background(), "compute_chart", args)
	if err == nil {
		t.Error("BUG: compute_chart accepts invalid gender='other'")
	}
}

func TestBugAgent_ComputeChart_EmptyGender(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{` + btAgent + `,"gender":""}`)
	_, err := r.Execute(context.Background(), "compute_chart", args)
	if err == nil {
		t.Error("BUG: compute_chart accepts empty gender=''")
	}
}

func TestBugAgent_ComputeChart_MissingBirth(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{"gender":"male"}`)
	_, err := r.Execute(context.Background(), "compute_chart", args)
	if err == nil {
		t.Error("BUG: compute_chart accepts missing birth")
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
		t.Error("BUG: compute_chart accepts empty birth.time")
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
		t.Error("BUG: compute_bond accepts missing 'b' (zero-value birth)")
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
		t.Error("BUG: compute_bond accepts missing 'a'")
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
	_, err := r.Execute(context.Background(), "compute_liunian", args)
	if err == nil {
		t.Error("BUG: compute_liunian accepts negative year=-1")
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
	_, err := r.Execute(context.Background(), "compute_liushi", args)
	if err == nil {
		t.Error("BUG: compute_liushi accepts hour=25 (out of range)")
	}
}

// --- compute_liuyao: invalid yong_shen ---

func TestBugAgent_ComputeLiuyao_InvalidYongShen(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{` + btAgent + `,"yong_shen":"invalid"}`)
	_, err := r.Execute(context.Background(), "compute_liuyao", args)
	if err == nil {
		t.Error("BUG: compute_liuyao accepts invalid yong_shen='invalid'")
	}
}


// --- compute_ziwei_bond: missing b ---

func TestBugAgent_ComputeZiweiBond_MissingB(t *testing.T) {
	r := NewChatToolRegistry()
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
	_, err = r.Execute(context.Background(), "compute_ziwei_bond", args)
	if err == nil {
		t.Error("BUG: compute_ziwei_bond accepts missing 'b'")
	}
}

// --- compute_ziwei_bond: empty charts ---

func TestBugAgent_ComputeZiweiBond_EmptyCharts(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{"a":{},"b":{}}`)
	_, err := r.Execute(context.Background(), "compute_ziwei_bond", args)
	if err == nil {
		t.Error("BUG: compute_ziwei_bond accepts empty charts {}")
	}
}


// --- compute_ziwei: invalid gender ---

func TestBugAgent_ComputeZiwei_InvalidGender(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{` + btAgent + `,"gender":"unknown"}`)
	_, err := r.Execute(context.Background(), "compute_ziwei", args)
	if err == nil {
		t.Error("BUG: compute_ziwei accepts invalid gender='unknown'")
	}
}

// --- compute_qimen: invalid kind ---

func TestBugAgent_ComputeQimen_InvalidKind(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{"birth":{"time":"1984-02-15T08:00:00+08:00","longitude":116.4},"kind":"xun"}`)
	_, err := r.Execute(context.Background(), "compute_qimen", args)
	if err == nil {
		t.Error("BUG: compute_qimen accepts invalid kind='xun'")
	}
}

// --- compute_minggua: invalid gender ---

func TestBugAgent_ComputeMingGua_InvalidGender(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{"gender":"unknown","birth_year":1990}`)
	_, err := r.Execute(context.Background(), "compute_minggua", args)
	if err == nil {
		t.Error("BUG: compute_minggua accepts invalid gender='unknown'")
	}
}


// --- compute_sanyuan_yun: invalid year ---

func TestBugAgent_ComputeSanYuanYun_InvalidYear(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{"year":0}`)
	_, err := r.Execute(context.Background(), "compute_sanyuan_yun", args)
	if err == nil {
		t.Error("BUG: compute_sanyuan_yun accepts year=0")
	}
}

// --- compute_xuankong: missing mountains ---

func TestBugAgent_ComputeXuankong_MissingMountains(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{` + btAgent + `}`)
	_, err := r.Execute(context.Background(), "compute_xuankong", args)
	if err == nil {
		t.Error("BUG: compute_xuankong accepts missing sit_mountain/face_mountain")
	}
}


// --- compute_naming_wuge: empty surname ---

func TestBugAgent_ComputeNamingWuge_EmptySurname(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{"surname":"","yong_shen":"木","xi_shen":["水"]}`)
	_, err := r.Execute(context.Background(), "compute_naming_wuge", args)
	if err == nil {
		t.Error("BUG: compute_naming_wuge accepts empty surname")
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
		t.Error("BUG: compute_naming_wuge accepts invalid yong_shen='x'")
	}
}

// --- compute_naming_evaluate: empty given_name ---

func TestBugAgent_ComputeNamingEvaluate_EmptyGivenName(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{"surname":"张","given_name":"","yong_shen":"木"}`)
	_, err := r.Execute(context.Background(), "compute_naming_evaluate", args)
	if err == nil {
		t.Error("BUG: compute_naming_evaluate accepts empty given_name")
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
		t.Error("BUG: query_huangli_date accepts missing event")
	}
}

// --- query_huangli_bond_date: missing birth ---

func TestBugAgent_QueryHuangliBondDate_MissingBirth(t *testing.T) {
	r := NewChatToolRegistry()
	args := json.RawMessage(`{"event_type":"结婚","date":"2025-06-01"}`)
	_, err := r.Execute(context.Background(), "query_huangli_bond_date", args)
	if err == nil {
		t.Error("BUG: query_huangli_bond_date accepts missing birth")
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
	_, err := r.Execute(context.Background(), "compute_bazhai", args)
	if err == nil {
		t.Error("BUG: compute_bazhai accepts missing gender")
	}
}
