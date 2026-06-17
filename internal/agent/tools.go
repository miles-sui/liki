package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"math"

	"liki/internal/engine/bazi"
	"liki/internal/engine/ganzhi"
	"liki/internal/engine/tianwen"
	"liki/internal/llm"
)

type ChatToolRegistry struct {
	handlers map[string]func(context.Context, json.RawMessage) (json.RawMessage, error)
	defs     []json.RawMessage
}

func NewChatToolRegistry() *ChatToolRegistry {
	r := &ChatToolRegistry{
		handlers: map[string]func(context.Context, json.RawMessage) (json.RawMessage, error){},
	}
	r.register("compute_chart", computeChartHandler, "排八字命盘，需出生年月日时经纬度性别")
	r.register("compute_bond", computeBondHandler, "合盘配对，需两人出生信息")
	r.register("compute_liunian", computeLiunianHandler, "推算流年运势，需命盘和流年")
	return r
}

func (r *ChatToolRegistry) register(name string, h func(context.Context, json.RawMessage) (json.RawMessage, error), desc string) {
	r.handlers[name] = h
	fn, _ := json.Marshal(struct { //nolint:errcheck
		Name        string `json:"name"`
		Description string `json:"description"`
	}{Name: name, Description: desc})
	r.defs = append(r.defs, fn)
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

// --- handler implementations ---

type birthParams struct {
	Year, Month, Day, Hour, Minute int
	Longitude, Timezone            float64
	Gender                         string
}

func computeChartHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p birthParams
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_chart: %w", err)
	}
	birthLong, birthTz := normGeo(p.Longitude, p.Timezone)
	bt := tianwen.ComputeBirthTime(p.Year, p.Month, p.Day, p.Hour, p.Minute, birthLong, birthTz)
	result := bazi.ComputeChart(bt.Solar, ganzhi.Gender(p.Gender))
	w := struct {
		Product string `json:"_product"`
		Data    any    `json:"data"`
	}{Product: "chart", Data: result}
	return json.Marshal(w)
}

type bondParams struct {
	A birthParams `json:"a"`
	B birthParams `json:"b"`
}

func computeBondHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p bondParams
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_bond: %w", err)
	}
	lonA, tzA := normGeo(p.A.Longitude, p.A.Timezone)
	btA := tianwen.ComputeBirthTime(p.A.Year, p.A.Month, p.A.Day, p.A.Hour, p.A.Minute, lonA, tzA)
	chartA := bazi.ComputeChart(btA.Solar, ganzhi.Gender(p.A.Gender))

	lonB, tzB := normGeo(p.B.Longitude, p.B.Timezone)
	btB := tianwen.ComputeBirthTime(p.B.Year, p.B.Month, p.B.Day, p.B.Hour, p.B.Minute, lonB, tzB)
	chartB := bazi.ComputeChart(btB.Solar, ganzhi.Gender(p.B.Gender))

	result := bazi.ComputeBond(chartA.ChartBase, chartB.ChartBase)
	w := struct {
		Product string `json:"_product"`
		Data    any    `json:"data"`
	}{Product: "bond", Data: result}
	return json.Marshal(w)
}

type liunianParams struct {
	ChartID string          `json:"chart_id"`
	Chart   json.RawMessage `json:"chart"`
	Year    int             `json:"year"`
	Dayun   json.RawMessage `json:"dayun"`
}

func computeLiunianHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p liunianParams
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_liunian: %w", err)
	}
	// Parse chart to get bazi info
	var ci struct {
		Year  struct{ Gan, Zhi int } `json:"Year"`
		Month struct{ Gan, Zhi int } `json:"Month"`
		Day   struct{ Gan, Zhi int } `json:"Day"`
		Hour  struct{ Gan, Zhi int } `json:"Hour"`
	}
	if err := json.Unmarshal(p.Chart, &ci); err != nil {
		return nil, fmt.Errorf("compute_liunian: parse chart: %w", err)
	}
	bz := ganzhi.Bazi{
		Nian: ganzhi.Zhu{Gan: ganzhi.Gan(ci.Year.Gan), Zhi: ganzhi.Zhi(ci.Year.Zhi)},
		Yue:  ganzhi.Zhu{Gan: ganzhi.Gan(ci.Month.Gan), Zhi: ganzhi.Zhi(ci.Month.Zhi)},
		Ri:   ganzhi.Zhu{Gan: ganzhi.Gan(ci.Day.Gan), Zhi: ganzhi.Zhi(ci.Day.Zhi)},
		Shi:  ganzhi.Zhu{Gan: ganzhi.Gan(ci.Hour.Gan), Zhi: ganzhi.Zhi(ci.Hour.Zhi)},
	}
	dayMaster := ganzhi.Gan(ci.Day.Gan)

	// Parse current dayun
	var cdu *bazi.DaYunPillar
	if len(p.Dayun) > 0 && string(p.Dayun) != "null" {
		var di struct {
			CurrentPillarIndex int `json:"CurrentPillarIndex"`
			Pillars            []struct {
				Gan int `json:"Gan"`
				Zhi int `json:"Zhi"`
			} `json:"Pillars"`
		}
		if json.Unmarshal(p.Dayun, &di) == nil && di.CurrentPillarIndex >= 0 && di.CurrentPillarIndex < len(di.Pillars) {
			dp := di.Pillars[di.CurrentPillarIndex]
			cdu = &bazi.DaYunPillar{Gan: ganzhi.Gan(dp.Gan), Zhi: ganzhi.Zhi(dp.Zhi)}
		}
	}

	result := bazi.ComputeLiuNian(p.Year, dayMaster, bz, cdu)
	w := struct {
		Product string `json:"_product"`
		Data    any    `json:"data"`
	}{Product: "liunian", Data: result}
	return json.Marshal(w)
}

// --- helpers ---

func normGeo(lon, tz float64) (float64, float64) {
	if lon == 0 { lon = 120 }
	if tz == 0 { tz = 8 }
	if math.Abs(tz) > 12 { tz = tz / 15 } // normalize degrees → hours
	return lon, tz
}
