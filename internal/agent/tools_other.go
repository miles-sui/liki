package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"liki/internal/engine/bazhai"
	"liki/internal/engine/ganzhi"
	"liki/internal/engine/huangli"
	"liki/internal/engine/liuyao"
	"liki/internal/engine/qimen"
	"liki/internal/engine/xuankong"
)

var validQimenKinds = map[string]bool{"shi": true, "ri": true, "yue": true, "nian": true}

func computeQimenHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Birth TimePoint `json:"birth"`
		Kind  string     `json:"kind"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_qimen: %w", err)
	}
	if p.Kind == "" {
		p.Kind = "shi"
	}
	if !validQimenKinds[p.Kind] {
		return nil, fmt.Errorf("compute_qimen: kind must be shi/ri/yue/nian, got %q", p.Kind)
	}
	ts, err := p.Birth.Timeset()
	if err != nil {
		return nil, fmt.Errorf("compute_qimen: %w", err)
	}
	result := qimen.ComputeChart(ts.Solar, p.Kind)
	return wrapResult("qimen", result)
}

func computeBazhaiHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p Person
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_bazhai: %w", err)
	}
	if err := validateGender(p.Gender); err != nil {
		return nil, fmt.Errorf("compute_bazhai: %w", err)
	}
	ts, err := p.Birth.Timeset()
	if err != nil {
		return nil, fmt.Errorf("compute_bazhai: %w", err)
	}
	result := bazhai.ComputeChart(ts.Solar, p.Gender)
	return wrapResult("bazhai", result)
}

func computeMingGuaHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Gender    ganzhi.Gender `json:"gender"`
		BirthYear int            `json:"birth_year"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_minggua: %w", err)
	}
	if err := validateGender(p.Gender); err != nil {
		return nil, fmt.Errorf("compute_minggua: %w", err)
	}
	result := bazhai.ComputeMingGua(p.Gender, p.BirthYear)
	return wrapResult("minggua", result)
}

func computeXuankongHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Birth        TimePoint `json:"birth"`
		SitMountain  *int       `json:"sit_mountain"`
		FaceMountain *int       `json:"face_mountain"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_xuankong: %w", err)
	}
	if p.SitMountain == nil {
		return nil, fmt.Errorf("compute_xuankong: sit_mountain is required")
	}
	if p.FaceMountain == nil {
		return nil, fmt.Errorf("compute_xuankong: face_mountain is required")
	}
	if *p.SitMountain < 0 || *p.SitMountain > 23 {
		return nil, fmt.Errorf("compute_xuankong: sit_mountain must be 0-23, got %d", *p.SitMountain)
	}
	if *p.FaceMountain < 0 || *p.FaceMountain > 23 {
		return nil, fmt.Errorf("compute_xuankong: face_mountain must be 0-23, got %d", *p.FaceMountain)
	}
	ts, err := p.Birth.Timeset()
	if err != nil {
		return nil, fmt.Errorf("compute_xuankong: %w", err)
	}
	result := xuankong.ComputeChart(ts.Solar, *p.SitMountain, *p.FaceMountain)
	return wrapResult("xuankong", result)
}

func computeSanYuanYunHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Year int `json:"year"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_sanyuan_yun: %w", err)
	}
	if p.Year <= 0 {
		return nil, fmt.Errorf("compute_sanyuan_yun: year must be positive, got %d", p.Year)
	}
	result := xuankong.ComputeSanYuanYun(p.Year)
	return wrapResult("sanyuan_yun", result)
}

func computeLiuyaoHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Birth    TimePoint `json:"birth"`
		YongShen string     `json:"yong_shen"`
		Fixed    [6]int     `json:"fixed"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_liuyao: %w", err)
	}
	if p.YongShen == "" {
		p.YongShen = "世爻"
	}
	ys, err := liuyao.ParseYongShen(p.YongShen)
	if err != nil {
		return nil, fmt.Errorf("compute_liuyao: invalid yong_shen %q", p.YongShen)
	}
	ts, err := p.Birth.Timeset()
	if err != nil {
		return nil, fmt.Errorf("compute_liuyao: %w", err)
	}
	result := liuyao.ComputeChart(ts.Solar, ys, p.Fixed)
	return wrapResult("liuyao", result)
}

func queryHuangliDateHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Date  string `json:"date"`
		Event string `json:"event"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("query_huangli_date: %w", err)
	}
	if p.Event == "" {
		return nil, fmt.Errorf("query_huangli_date: event is required")
	}
	result, err := huangli.QueryDate(p.Date, p.Event)
	if err != nil {
		return nil, fmt.Errorf("query_huangli_date: %w", err)
	}
	return wrapResult("huangli_date", result)
}

func queryHuangliMonthHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Month string `json:"month"`
		Event string `json:"event"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("query_huangli_month: %w", err)
	}
	result, err := huangli.QueryMonth(p.Month, p.Event)
	if err != nil {
		return nil, fmt.Errorf("query_huangli_month: %w", err)
	}
	return wrapResult("huangli_month", result)
}

func queryHuangliBondDateHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Birth     TimePoint `json:"birth"`
		EventType string     `json:"event_type"`
		Date      string     `json:"date"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("query_huangli_bond_date: %w", err)
	}
	ts, err := p.Birth.Timeset()
	if err != nil {
		return nil, fmt.Errorf("query_huangli_bond_date: %w", err)
	}
	result, err := huangli.ComputeBondDay(ts.Solar, p.EventType, p.Date)
	if err != nil {
		return nil, fmt.Errorf("query_huangli_bond_date: %w", err)
	}
	return wrapResult("huangli_bond_date", result)
}

func queryHuangliBondMonthHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Birth     TimePoint `json:"birth"`
		EventType string     `json:"event_type"`
		Month     string     `json:"month"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("query_huangli_bond_month: %w", err)
	}
	ts, err := p.Birth.Timeset()
	if err != nil {
		return nil, fmt.Errorf("query_huangli_bond_month: %w", err)
	}
	result, err := huangli.ComputeBondMonth(ts.Solar, p.EventType, p.Month)
	if err != nil {
		return nil, fmt.Errorf("query_huangli_bond_month: %w", err)
	}
	return wrapResult("huangli_bond_month", result)
}
