package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"liki/internal/agent/city"
	"liki/internal/engine/ganzhi"
	"liki/internal/engine/tianwen"
	"liki/internal/llm"
)

var (
	toolsOnce sync.Once
	toolsMap  map[string]toolEntry
)

type toolEntry struct {
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

func initTools() {
	var f struct {
		Tools []struct {
			Name        string          `json:"name"`
			Description string          `json:"description"`
			Parameters  json.RawMessage `json:"parameters"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(ToolsJSON, &f); err != nil {
		panic("initTools: " + err.Error())
	}
	toolsMap = make(map[string]toolEntry, len(f.Tools))
	for _, t := range f.Tools {
		toolsMap[t.Name] = toolEntry{Description: t.Description, Parameters: t.Parameters}
	}
}

// ValidateTools parses tools.json eagerly so configuration errors fail at startup
// instead of at first request time.
func ValidateTools() error {
	var err error
	toolsOnce.Do(func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("tools: %v", r)
			}
		}()
		initTools()
	})
	return err
}

func toolParams(name string) (string, json.RawMessage, error) {
	toolsOnce.Do(initTools)
	e, ok := toolsMap[name]
	if !ok {
		return "", nil, fmt.Errorf("tool not found in tools.json: %s", name)
	}
	return e.Description, e.Parameters, nil
}

func computeTimeHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var tp TimePoint
	if err := json.Unmarshal(raw, &tp); err != nil {
		return nil, fmt.Errorf("compute_time: %w", err)
	}
	ts, err := tp.Timeset()
	if err != nil {
		return nil, fmt.Errorf("compute_time: %w", err)
	}
	return json.Marshal(ts)
}

const (
	ToolQueryCity          = "query_city"
	ToolComputeTime        = "compute_time"
	ToolComputeChart       = "compute_chart"
	ToolComputeZiwei       = "compute_ziwei"
	ToolComputeNamingWuge  = "compute_naming_wuge"
	ToolComputeNamingCompose = "compute_naming_compose"
	ToolComputeNamingDetail  = "compute_naming_detail"
	ToolComputeNamingEvaluate = "compute_naming_evaluate"
)

type ChatToolRegistry struct {
	handlers map[string]func(context.Context, json.RawMessage) (json.RawMessage, error)
	defs     []json.RawMessage
}

// NewNamingToolRegistry creates a tool registry for the naming chat flow with 8 tools.
func NewNamingToolRegistry() *ChatToolRegistry {
	r := &ChatToolRegistry{
		handlers: map[string]func(context.Context, json.RawMessage) (json.RawMessage, error){},
	}

	// --- collection ---
	r.registerTool(ToolQueryCity, city.SearchCity)
	r.registerTool(ToolComputeTime, computeTimeHandler)

	// --- naming ---
	r.registerTool(ToolComputeChart, computeChartHandler)
	r.registerTool(ToolComputeZiwei, computeZiweiHandler)
	r.registerTool(ToolComputeNamingWuge, computeNamingWuGeHandler)
	r.registerTool(ToolComputeNamingCompose, computeNamingComposeHandler)
	r.registerTool(ToolComputeNamingDetail, computeNamingDetailHandler)
	r.registerTool(ToolComputeNamingEvaluate, computeNamingEvaluateHandler)

	return r
}

func (r *ChatToolRegistry) registerTool(name string, h func(context.Context, json.RawMessage) (json.RawMessage, error)) {
	desc, params, err := toolParams(name)
	if err != nil {
		panic("registerTool: " + err.Error())
	}
	r.handlers[name] = h
	r.defs = append(r.defs, toolDef(name, desc, params))
}

func toolDef(name, desc string, params json.RawMessage) json.RawMessage {
	fn := map[string]any{
		"name":        name,
		"description": desc,
	}
	if params != nil {
		fn["parameters"] = params
	}
	b, err := json.Marshal(fn)
	if err != nil {
		panic("marshal tool def: " + err.Error())
	}
	return b
}

func (r *ChatToolRegistry) Execute(ctx context.Context, name string, args json.RawMessage) (json.RawMessage, error) {
	h, ok := r.handlers[name]
	if !ok {
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
	return h(ctx, args)
}

func (r *ChatToolRegistry) Schemas() []llm.ToolDef {
	out := make([]llm.ToolDef, len(r.defs))
	for i, fn := range r.defs {
		out[i] = llm.ToolDef{Type: "function", Function: fn}
	}
	return out
}

// --- shared types ---

// TimePoint is a gregorian time with longitude for solar time correction.
type TimePoint struct {
	Time      string  `json:"time"`
	Longitude float64 `json:"longitude"`
}

// Timeset converts TimePoint to a tianwen.Timeset for engine computation.
func (b TimePoint) Timeset() (tianwen.Timeset, error) {
	t, err := time.Parse(time.RFC3339, b.Time)
	if err != nil {
		return tianwen.Timeset{}, fmt.Errorf("invalid time: %w", err)
	}
	_, offset := t.Zone()
	tz := float64(offset) / 3600
	return tianwen.ComputeTimeset(tianwen.GregorianTime(t.In(time.FixedZone("", int(tz*3600)))), b.Longitude), nil
}

type Person struct {
	Birth  TimePoint     `json:"birth"`
	Gender ganzhi.Gender `json:"gender"`
}

// resolvePerson unmarshals raw JSON as a Person, validates gender, and computes
// the solar time. Used by handlers that accept bare Person input (compute_chart,
// compute_ziwei, etc.).
func resolvePerson(raw json.RawMessage, name string) (tianwen.SolarTime, ganzhi.Gender, error) {
	var p Person
	if err := json.Unmarshal(raw, &p); err != nil {
		return tianwen.SolarTime{}, "", fmt.Errorf("%s: %w", name, err)
	}
	if err := validateGender(p.Gender); err != nil {
		return tianwen.SolarTime{}, "", fmt.Errorf("%s: %w", name, err)
	}
	ts, err := p.Birth.Timeset()
	if err != nil {
		return tianwen.SolarTime{}, "", fmt.Errorf("%s: %w", name, err)
	}
	return ts.Solar, p.Gender, nil
}

// --- helpers ---

func wrapResult[T any](product string, data T) (json.RawMessage, error) {
	return json.Marshal(struct {
		Product string `json:"_product"`
		Data    T      `json:"data"`
	}{Product: product, Data: data})
}

// validateGender checks that the gender is male or female (not empty or invalid).
func validateGender(g ganzhi.Gender) error {
	if _, err := ganzhi.ParseGender(string(g)); err != nil {
		return fmt.Errorf("gender must be male or female, got %q", g)
	}
	return nil
}

