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
	result := bazi.ComputeChart(ts.Solar, ganzhi.Gender(p.Gender))
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
	chartA := bazi.ComputeChart(tsA.Solar, ganzhi.Gender(p.A.Gender))
	tsB, err := p.B.Birth.Timeset()
	if err != nil {
		return nil, fmt.Errorf("compute_bond: %w", err)
	}
	chartB := bazi.ComputeChart(tsB.Solar, ganzhi.Gender(p.B.Gender))
	result := bazi.ComputeBond(chartA.ChartBase, chartB.ChartBase)
	return wrapResult("bond", result)
}

func computeLiunianHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Year  int             `json:"year"`
		Birth TimePoint      `json:"birth"`
		Dayun json.RawMessage `json:"dayun"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_liunian: %w", err)
	}
	cdu := parseDaYunZhu(p.Dayun)
	ts, err := p.Birth.Timeset()
	if err != nil {
		return nil, fmt.Errorf("compute_liunian: %w", err)
	}
	result, err := bazi.ComputeLiuNian(ts.Solar, p.Year, cdu)
	if err != nil {
		return nil, fmt.Errorf("compute_liunian: %w", err)
	}
	return wrapResult("liunian", result)
}

func computeLiuyueHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Year  int        `json:"year"`
		Month int        `json:"month"`
		Birth TimePoint `json:"birth"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_liuyue: %w", err)
	}
	ts, err := p.Birth.Timeset()
	if err != nil {
		return nil, fmt.Errorf("compute_liuyue: %w", err)
	}
	result, err := bazi.ComputeLiuYue(ts.Solar, p.Year, p.Month)
	if err != nil {
		return nil, fmt.Errorf("compute_liuyue: %w", err)
	}
	return wrapResult("liuyue", result)
}

func computeLiuriHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Date    string          `json:"date"`
		Birth   TimePoint      `json:"birth"`
		Dayun   json.RawMessage `json:"dayun"`
		LiuNian json.RawMessage `json:"liunian"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_liuri: %w", err)
	}
	dp := parseZhu(p.Dayun)
	lp := parseZhu(p.LiuNian)
	ts, err := p.Birth.Timeset()
	if err != nil {
		return nil, fmt.Errorf("compute_liuri: %w", err)
	}
	result, err := bazi.ComputeLiuRi(ts.Solar, p.Date, dp, lp)
	if err != nil {
		return nil, fmt.Errorf("compute_liuri: %w", err)
	}
	return wrapResult("liuri", result)
}

func computeLiushiHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Date  string     `json:"date"`
		Hour  int        `json:"hour"`
		Birth TimePoint `json:"birth"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_liushi: %w", err)
	}
	ts, err := p.Birth.Timeset()
	if err != nil {
		return nil, fmt.Errorf("compute_liushi: %w", err)
	}
	result, err := bazi.ComputeLiuShi(ts.Solar, p.Date, p.Hour)
	if err != nil {
		return nil, fmt.Errorf("compute_liushi: %w", err)
	}
	return wrapResult("liushi", result)
}

func computeXiaoYunHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Birth  TimePoint `json:"birth"`
		Gender string     `json:"gender"`
		Count  int        `json:"count"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_xiaoyun: %w", err)
	}
	ts, err := p.Birth.Timeset()
	if err != nil {
		return nil, fmt.Errorf("compute_xiaoyun: %w", err)
	}
	result := bazi.ComputeXiaoYun(ts.Solar, ganzhi.Gender(p.Gender), p.Count)
	return wrapResult("xiaoyun", result)
}

func computeXiaoXianHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Gender string `json:"gender"`
		Count  int    `json:"count"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_xiaoxian: %w", err)
	}
	result := bazi.ComputeXiaoXian(ganzhi.Gender(p.Gender), p.Count)
	return wrapResult("xiaoxian", result)
}
