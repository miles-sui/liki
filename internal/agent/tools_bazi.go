package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"liki/internal/engine/bazi"
	"liki/internal/engine/ganzhi"
)

func computeChartHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p Person
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_chart: %w", err)
	}
	ts, err := p.Birth.Timeset()
	if err != nil {
		return nil, fmt.Errorf("compute_chart: %w", err)
	}
	result := bazi.ComputeChart(ts.Solar, p.Gender)
	return wrapResult("chart", result)
}

func computeBondHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		A Person `json:"a"`
		B Person `json:"b"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_bond: %w", err)
	}
	tsA, err := p.A.Birth.Timeset()
	if err != nil {
		return nil, fmt.Errorf("compute_bond: %w", err)
	}
	chartA := bazi.ComputeChart(tsA.Solar, p.A.Gender)
	tsB, err := p.B.Birth.Timeset()
	if err != nil {
		return nil, fmt.Errorf("compute_bond: %w", err)
	}
	chartB := bazi.ComputeChart(tsB.Solar, p.B.Gender)
	result := bazi.ComputeBond(chartA.ChartBase, chartB.ChartBase)
	return wrapResult("bond", result)
}

func computeLiunianHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Year   int             `json:"year"`
		Birth  TimePoint       `json:"birth"`
		Gender ganzhi.Gender   `json:"gender"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_liunian: %w", err)
	}
	ts, err := p.Birth.Timeset()
	if err != nil {
		return nil, fmt.Errorf("compute_liunian: %w", err)
	}
	chart := bazi.ComputeChart(ts.Solar, p.Gender)
	result, err := bazi.ComputeLiuNian(chart.ChartBase, p.Year)
	if err != nil {
		return nil, fmt.Errorf("compute_liunian: %w", err)
	}
	return wrapResult("liunian", result)
}

func computeLiuyueHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Year   int             `json:"year"`
		Month  int             `json:"month"`
		Birth  TimePoint       `json:"birth"`
		Gender ganzhi.Gender   `json:"gender"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_liuyue: %w", err)
	}
	ts, err := p.Birth.Timeset()
	if err != nil {
		return nil, fmt.Errorf("compute_liuyue: %w", err)
	}
	chart := bazi.ComputeChart(ts.Solar, p.Gender)
	result, err := bazi.ComputeLiuYue(chart.ChartBase, p.Year, p.Month)
	if err != nil {
		return nil, fmt.Errorf("compute_liuyue: %w", err)
	}
	return wrapResult("liuyue", result)
}

func computeLiuriHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Year   int             `json:"year"`
		Month  int             `json:"month"`
		Day    int             `json:"day"`
		Birth  TimePoint       `json:"birth"`
		Gender ganzhi.Gender   `json:"gender"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_liuri: %w", err)
	}
	ts, err := p.Birth.Timeset()
	if err != nil {
		return nil, fmt.Errorf("compute_liuri: %w", err)
	}
	chart := bazi.ComputeChart(ts.Solar, p.Gender)
	result, err := bazi.ComputeLiuRi(chart.ChartBase, p.Year, p.Month, p.Day)
	if err != nil {
		return nil, fmt.Errorf("compute_liuri: %w", err)
	}
	return wrapResult("liuri", result)
}

func computeLiushiHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Year   int             `json:"year"`
		Month  int             `json:"month"`
		Day    int             `json:"day"`
		Hour   int             `json:"hour"`
		Birth  TimePoint       `json:"birth"`
		Gender ganzhi.Gender   `json:"gender"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_liushi: %w", err)
	}
	ts, err := p.Birth.Timeset()
	if err != nil {
		return nil, fmt.Errorf("compute_liushi: %w", err)
	}
	chart := bazi.ComputeChart(ts.Solar, p.Gender)
	result, err := bazi.ComputeLiuShi(chart.ChartBase, p.Year, p.Month, p.Day, p.Hour)
	if err != nil {
		return nil, fmt.Errorf("compute_liushi: %w", err)
	}
	return wrapResult("liushi", result)
}

func computeXiaoYunHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Birth  TimePoint     `json:"birth"`
		Gender ganzhi.Gender `json:"gender"`
		Count  int            `json:"count"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_xiaoyun: %w", err)
	}
	ts, err := p.Birth.Timeset()
	if err != nil {
		return nil, fmt.Errorf("compute_xiaoyun: %w", err)
	}
	result := bazi.ComputeXiaoYun(ts.Solar, p.Gender, p.Count)
	return wrapResult("xiaoyun", result)
}

func computeXiaoXianHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Gender ganzhi.Gender `json:"gender"`
		Count  int            `json:"count"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_xiaoxian: %w", err)
	}
	result := bazi.ComputeXiaoXian(p.Gender, p.Count)
	return wrapResult("xiaoxian", result)
}
