package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"liki/internal/engine/ganzhi"
	"liki/internal/engine/tianwen"
	doc "liki"
	"liki/internal/llm"
)

// openapiParams extracts the JSON Schema parameters for the given tool name.
func openapiParams(tool string) json.RawMessage {
	var api struct {
		Paths map[string]map[string]struct {
			XAgentTool string `json:"x-agent-tool"`
			Parameters []struct {
				Name     string          `json:"name"`
				Required bool            `json:"required"`
				Schema   json.RawMessage `json:"schema"`
			} `json:"parameters"`
			RequestBody struct {
				Content map[string]struct {
					Schema json.RawMessage `json:"schema"`
				} `json:"content"`
			} `json:"requestBody"`
		} `json:"paths"`
		XAgentTools map[string]struct {
			Parameters json.RawMessage `json:"parameters"`
		} `json:"x-agent-tools"`
	}
	if err := json.Unmarshal(doc.OpenAPIJSON, &api); err != nil {
		return nil
	}
	if t, ok := api.XAgentTools[tool]; ok {
		return t.Parameters
	}
	for _, methods := range api.Paths {
		for _, op := range methods {
			if op.XAgentTool != tool {
				continue
			}
			if s, ok := op.RequestBody.Content["application/json"]; ok {
				return s.Schema
			}
			if len(op.Parameters) > 0 {
				props := map[string]any{}
				required := []string{}
				for _, p := range op.Parameters {
					var ps map[string]any
					if err := json.Unmarshal(p.Schema, &ps); err != nil {
						panic("openapiParams: invalid param schema for " + p.Name + ": " + err.Error())
					}
					props[p.Name] = ps
					if p.Required {
						required = append(required, p.Name)
					}
				}
				schema := map[string]any{
					"type":       "object",
					"properties": props,
					"required":   required,
				}
				b, err := json.Marshal(schema)
				if err != nil {
					panic("openapiParams: marshal schema: " + err.Error())
				}
				return json.RawMessage(b)
			}
			return nil
		}
	}
	return nil
}

type ChatToolRegistry struct {
	handlers map[string]func(context.Context, json.RawMessage) (json.RawMessage, error)
	defs     []json.RawMessage
}

func NewChatToolRegistry() *ChatToolRegistry {
	r := &ChatToolRegistry{
		handlers: map[string]func(context.Context, json.RawMessage) (json.RawMessage, error){},
	}

	// --- bazi ---
	r.register("compute_chart", computeChartHandler, "排八字命盘")
	r.register("compute_bond", computeBondHandler, "八字合盘配对")
	r.register("compute_liunian", computeLiunianHandler, "八字流年运势")
	r.register("compute_liuyue", computeLiuyueHandler, "八字流月运势")
	r.register("compute_liuri", computeLiuriHandler, "八字流日运势")
	r.register("compute_liushi", computeLiushiHandler, "八字流时运势")
	r.register("compute_xiaoyun", computeXiaoYunHandler, "八字小运")
	r.register("compute_xiaoxian", computeXiaoXianHandler, "八字小限")

	// --- ziwei ---
	r.register("compute_ziwei", computeZiweiHandler, "紫微斗数命盘")
	r.register("compute_ziwei_daxian", computeZiweiDaXianHandler, "紫微斗数大限")
	r.register("compute_ziwei_liunian", computeZiweiLiuNianHandler, "紫微斗数流年")
	r.register("compute_ziwei_liuyue", computeZiweiLiuYueHandler, "紫微斗数流月")
	r.register("compute_ziwei_liuri", computeZiweiLiuRiHandler, "紫微斗数流日")
	r.register("compute_ziwei_bond", computeZiweiBondHandler, "紫微斗数合盘")

	// --- qiming ---
	r.register("compute_naming_wuge", computeNamingWuGeHandler, "起名五格计算")
	r.register("compute_naming_compose", computeNamingComposeHandler, "起名候选名字组合")
	r.register("compute_naming_detail", computeNamingDetailHandler, "起名候选名字详析")
	r.register("compute_naming_evaluate", computeNamingEvaluateHandler, "起名单名评分")

	// --- qimen ---
	r.register("compute_qimen", computeQimenHandler, "奇门遁甲排盘")

	// --- bazhai ---
	r.register("compute_bazhai", computeBazhaiHandler, "八宅风水命盘")
	r.register("compute_minggua", computeMingGuaHandler, "命卦计算")

	// --- xuankong ---
	r.register("compute_xuankong", computeXuankongHandler, "玄空飞星排盘")
	r.register("compute_sanyuan_yun", computeSanYuanYunHandler, "三元九运查询")

	// --- liuyao ---
	r.register("compute_liuyao", computeLiuyaoHandler, "六爻起卦排盘")

	// --- huangli ---
	r.register("query_huangli_date", queryHuangliDateHandler, "黄历按日查询宜忌")
	r.register("query_huangli_month", queryHuangliMonthHandler, "黄历按月查询宜忌")
	r.register("query_huangli_bond_date", queryHuangliBondDateHandler, "八字合参按日择日")
	r.register("query_huangli_bond_month", queryHuangliBondMonthHandler, "八字合参按月择日")

	// --- infra ---
	r.register("query_city", queryCity, "根据城市名查询经纬度")

	return r
}

func (r *ChatToolRegistry) register(name string, h func(context.Context, json.RawMessage) (json.RawMessage, error), desc string) {
	r.handlers[name] = h
	fn := map[string]any{
		"name":        name,
		"description": desc,
	}
	if params := openapiParams(name); params != nil {
		fn["parameters"] = params
	}
	b, err := json.Marshal(fn)
	if err != nil {
		panic("marshal tool def: " + err.Error())
	}
	r.defs = append(r.defs, b)
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

// --- helpers ---

func wrapResult(product string, data any) (json.RawMessage, error) {
	return json.Marshal(struct {
		Product string `json:"_product"`
		Data    any    `json:"data"`
	}{Product: product, Data: data})
}

