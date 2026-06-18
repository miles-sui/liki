package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"liki/internal/engine/qiming"
)

func computeNamingWuGeHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Surname  string   `json:"surname"`
		YongShen string   `json:"yong_shen"`
		XiShen   []string `json:"xi_shen"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_naming_wuge: %w", err)
	}
	result, err := qiming.PrepareWuGe(p.Surname, p.YongShen, p.XiShen)
	if err != nil {
		return nil, fmt.Errorf("compute_naming_wuge: %w", err)
	}
	return wrapResult("naming_wuge", result)
}

func computeNamingComposeHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Surname   string               `json:"surname"`
		Combos    []qiming.StrokeCombo `json:"combos"`
		YongChars map[int][]qiming.CharLite `json:"yong_chars"`
		XiChars   map[int][]qiming.CharLite `json:"xi_chars"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_naming_compose: %w", err)
	}
	result := qiming.ComposeNames(p.Surname, p.Combos, p.YongChars, p.XiChars)
	return wrapResult("naming_compose", result)
}

func computeNamingDetailHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Surname string   `json:"surname"`
		Names   []string `json:"names"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_naming_detail: %w", err)
	}
	result, err := qiming.DetailNames(p.Surname, p.Names)
	if err != nil {
		return nil, fmt.Errorf("compute_naming_detail: %w", err)
	}
	return wrapResult("naming_detail", result)
}

func computeNamingEvaluateHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Surname   string `json:"surname"`
		GivenName string `json:"given_name"`
		YongShen  string `json:"yong_shen"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_naming_evaluate: %w", err)
	}
	result, err := qiming.EvaluateName(p.Surname, p.GivenName, p.YongShen)
	if err != nil {
		return nil, fmt.Errorf("compute_naming_evaluate: %w", err)
	}
	return wrapResult("naming_evaluate", result)
}
