package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"liki/internal/engine/ziwei"
)

func computeZiweiHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p Person
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_ziwei: %w", err)
	}
	ts, err := p.Birth.Timeset()
	if err != nil {
		return nil, fmt.Errorf("compute_ziwei: %w", err)
	}
	result := ziwei.ComputeChart(ts.Solar, p.Gender)
	return wrapResult("ziwei", result)
}

func computeZiweiDaXianHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Chart json.RawMessage `json:"chart"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_ziwei_daxian: %w", err)
	}
	var chart ziwei.Chart
	if err := json.Unmarshal(p.Chart, &chart); err != nil {
		return nil, fmt.Errorf("compute_ziwei_daxian: parse chart: %w", err)
	}
	result := ziwei.ComputeDaXian(chart)
	return wrapResult("ziwei_daxian", result)
}

func computeZiweiLiuNianHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		LiuYear int             `json:"liu_year"`
		Chart   json.RawMessage `json:"chart"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_ziwei_liunian: %w", err)
	}
	var chart ziwei.Chart
	if err := json.Unmarshal(p.Chart, &chart); err != nil {
		return nil, fmt.Errorf("compute_ziwei_liunian: parse chart: %w", err)
	}
	result := ziwei.ComputeLiuNian(p.LiuYear, chart)
	return wrapResult("ziwei_liunian", result)
}

func computeZiweiLiuYueHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		LiuYear    int             `json:"liu_year"`
		LunarMonth int             `json:"lunar_month"`
		Chart      json.RawMessage `json:"chart"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_ziwei_liuyue: %w", err)
	}
	var chart ziwei.Chart
	if err := json.Unmarshal(p.Chart, &chart); err != nil {
		return nil, fmt.Errorf("compute_ziwei_liuyue: parse chart: %w", err)
	}
	result := ziwei.ComputeLiuYue(p.LiuYear, p.LunarMonth, chart)
	return wrapResult("ziwei_liuyue", result)
}

func computeZiweiLiuRiHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		LiuYear    int             `json:"liu_year"`
		LunarMonth int             `json:"lunar_month"`
		LunarDay   int             `json:"lunar_day"`
		Chart      json.RawMessage `json:"chart"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_ziwei_liuri: %w", err)
	}
	var chart ziwei.Chart
	if err := json.Unmarshal(p.Chart, &chart); err != nil {
		return nil, fmt.Errorf("compute_ziwei_liuri: parse chart: %w", err)
	}
	result := ziwei.ComputeLiuRi(p.LiuYear, p.LunarMonth, p.LunarDay, chart)
	return wrapResult("ziwei_liuri", result)
}

func computeZiweiBondHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		A json.RawMessage `json:"a"`
		B json.RawMessage `json:"b"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_ziwei_bond: %w", err)
	}
	var chartA, chartB ziwei.Chart
	if err := json.Unmarshal(p.A, &chartA); err != nil {
		return nil, fmt.Errorf("compute_ziwei_bond: parse chart a: %w", err)
	}
	if err := json.Unmarshal(p.B, &chartB); err != nil {
		return nil, fmt.Errorf("compute_ziwei_bond: parse chart b: %w", err)
	}
	result := ziwei.ComputeBond(chartA, chartB)
	return wrapResult("ziwei_bond", result)
}
