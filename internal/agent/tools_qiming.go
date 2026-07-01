package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"liki/internal/engine/ganzhi"
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
	if _, err := ganzhi.ParseWuxing(p.YongShen); err != nil {
		return nil, fmt.Errorf("compute_naming_wuge: yong_shen must be one of 木/火/土/金/水, got %q", p.YongShen)
	}
	result, err := qiming.PrepareWuGe(p.Surname, p.YongShen, p.XiShen)
	if err != nil {
		return nil, fmt.Errorf("compute_naming_wuge: %w", err)
	}
	return wrapResult("naming_wuge", result)
}

func computeNamingSancaiHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Surname string `json:"surname"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_naming_sancai: %w", err)
	}
	combos, err := qiming.EnumerateSancai(p.Surname)
	if err != nil {
		return nil, fmt.Errorf("compute_naming_sancai: %w", err)
	}
	return wrapResult("naming_sancai", map[string]any{"combos": combos})
}

func computeNamingCharsHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Wuxing    string `json:"wuxing"`
		StrokeMin int    `json:"stroke_min"`
		StrokeMax int    `json:"stroke_max"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_naming_chars: %w", err)
	}
	chars, err := qiming.GetChars(p.Wuxing, p.StrokeMin, p.StrokeMax)
	if err != nil {
		return nil, fmt.Errorf("compute_naming_chars: %w", err)
	}
	return wrapResult("naming_chars", map[string]any{"chars": chars})
}

func computeNamingComposeHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Surname   string                 `json:"surname"`
		Combos    []qiming.StrokeCombo   `json:"combos"`
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
	result, err := qiming.EvaluateNames(p.Surname, p.Names, "", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("compute_naming_detail: %w", err)
	}
	return wrapResult("naming_detail", result)
}

func computeNamingEvaluateHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Surname   string   `json:"surname"`
		Names     []string `json:"names"`
		GivenName string   `json:"given_name"`
		YongShen  string   `json:"yong_shen"`
		XiShen    []string `json:"xi_shen"`
		JiShen    []string `json:"ji_shen"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_naming_evaluate: %w", err)
	}

	// Backward compat: given_name → names[0]
	if len(p.Names) == 0 && p.GivenName != "" {
		p.Names = []string{p.GivenName}
	}

	var results []qiming.Evaluation
	if len(p.Names) == 1 {
		// Single name path: use existing EvaluateName for backward compat
		ev, err := qiming.EvaluateName(p.Surname, p.Names[0], p.YongShen)
		if err != nil {
			return nil, fmt.Errorf("compute_naming_evaluate: %w", err)
		}
		// Add wuxing detail if yong/xi/ji provided
		if p.YongShen != "" && (len(p.XiShen) > 0 || len(p.JiShen) > 0) {
			evals, _ := qiming.EvaluateNames(p.Surname, p.Names, p.YongShen, p.XiShen, p.JiShen)
			if len(evals) > 0 {
				ev.Wuxing = evals[0].Wuxing
			}
		}
		results = []qiming.Evaluation{ev}
	} else {
		var err error
		results, err = qiming.EvaluateNames(p.Surname, p.Names, p.YongShen, p.XiShen, p.JiShen)
		if err != nil {
			return nil, fmt.Errorf("compute_naming_evaluate: %w", err)
		}
	}
	return wrapResult("naming_evaluate", results)
}
