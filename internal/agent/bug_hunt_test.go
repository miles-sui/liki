package agent

import (
	"context"
	"encoding/json"
	"testing"
)

const btAgent = `"birth":{"time":"1984-02-15T08:00:00+08:00","longitude":116.4}`

// ============================================================
// Regression tests: naming tool input validation.
// ============================================================

// --- compute_chart: invalid gender ---

func TestBugAgent_ComputeChart_InvalidGender(t *testing.T) {
	r := NewNamingToolRegistry()
	args := json.RawMessage(`{` + btAgent + `,"gender":"other"}`)
	_, err := r.Execute(context.Background(), "compute_chart", args)
	if err == nil {
		t.Error("BUG: compute_chart accepts invalid gender='other'")
	}
}

func TestBugAgent_ComputeChart_EmptyGender(t *testing.T) {
	r := NewNamingToolRegistry()
	args := json.RawMessage(`{` + btAgent + `,"gender":""}`)
	_, err := r.Execute(context.Background(), "compute_chart", args)
	if err == nil {
		t.Error("BUG: compute_chart accepts empty gender=''")
	}
}

func TestBugAgent_ComputeChart_MissingBirth(t *testing.T) {
	r := NewNamingToolRegistry()
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
	r := NewNamingToolRegistry()
	args := json.RawMessage(`{"birth":{"time":"","longitude":116.4},"gender":"male"}`)
	_, err := r.Execute(context.Background(), "compute_chart", args)
	if err == nil {
		t.Error("BUG: compute_chart accepts empty birth.time")
	} else {
		t.Logf("OK: rejects empty time: %v", err)
	}
}

// --- compute_ziwei: invalid gender ---

func TestBugAgent_ComputeZiwei_InvalidGender(t *testing.T) {
	r := NewNamingToolRegistry()
	args := json.RawMessage(`{` + btAgent + `,"gender":"unknown"}`)
	_, err := r.Execute(context.Background(), "compute_ziwei", args)
	if err == nil {
		t.Error("BUG: compute_ziwei accepts invalid gender='unknown'")
	}
}

// --- compute_naming_wuge: empty surname ---

func TestBugAgent_ComputeNamingWuge_EmptySurname(t *testing.T) {
	r := NewNamingToolRegistry()
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
	r := NewNamingToolRegistry()
	args := json.RawMessage(`{"surname":"张","yong_shen":"x","xi_shen":["水"]}`)
	_, err := r.Execute(context.Background(), "compute_naming_wuge", args)
	if err == nil {
		t.Error("BUG: compute_naming_wuge accepts invalid yong_shen='x'")
	}
}

// --- compute_naming_evaluate: empty given_name ---

func TestBugAgent_ComputeNamingEvaluate_EmptyGivenName(t *testing.T) {
	r := NewNamingToolRegistry()
	args := json.RawMessage(`{"surname":"张","given_name":"","yong_shen":"木"}`)
	_, err := r.Execute(context.Background(), "compute_naming_evaluate", args)
	if err == nil {
		t.Error("BUG: compute_naming_evaluate accepts empty given_name")
	} else {
		t.Logf("OK: rejects empty given_name: %v", err)
	}
}

// --- unknown tool name ---

func TestBugAgent_UnknownTool(t *testing.T) {
	r := NewNamingToolRegistry()
	_, err := r.Execute(context.Background(), "nonexistent_tool", json.RawMessage(`{}`))
	if err == nil {
		t.Error("BUG: unknown tool should return error")
	}
}
