package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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
	r.register("get_city_coords", handleGetCityCoords, "根据城市名查询经纬度")

	return r
}

func (r *ChatToolRegistry) register(name string, h func(context.Context, json.RawMessage) (json.RawMessage, error), desc string) {
	r.handlers[name] = h
	fn, err := json.Marshal(struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}{Name: name, Description: desc})
	if err != nil {
		panic("marshal tool def: " + err.Error())
	}
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

// --- shared types ---

// LunarDate holds a lunar calendar date.
type LunarDate struct {
	Year   int  `json:"year"`
	Month  int  `json:"month"`
	Day    int  `json:"day"`
	Hour   int  `json:"hour"`
	Minute int  `json:"minute"`
	Leap   bool `json:"leap"`
}

// TimePoint is a point in time with optional longitude for solar time correction.
type TimePoint struct {
	Time      string       `json:"time"`
	Lunar     *LunarDate `json:"lunar"`
	Longitude float64      `json:"longitude"`
}

// Timeset converts TimePoint to a tianwen.Timeset for engine computation.
func (b TimePoint) Timeset() (tianwen.Timeset, error) {
	if b.Lunar != nil {
		ly, lm, ld := tianwen.LunarToSolar(b.Lunar.Year, b.Lunar.Month, b.Lunar.Day, b.Lunar.Leap)
		lon, tz := NormGeo(b.Longitude, 0)
		return tianwen.ComputeTime(ly, lm, ld, b.Lunar.Hour, b.Lunar.Minute, lon, tz), nil
	}

	t, err := time.Parse(time.RFC3339, b.Time)
	if err != nil {
		return tianwen.Timeset{}, fmt.Errorf("invalid time: %w", err)
	}
	_, offset := t.Zone()
	tz := float64(offset) / 3600
	lon, _ := NormGeo(b.Longitude, 0) // only normalize lon; tz from timestamp
	return tianwen.ComputeTime(t.Year(), int(t.Month()), t.Day(), t.Hour(), t.Minute(), lon, tz), nil
}

type Person struct {
	Birth  TimePoint `json:"birth"`
	Gender string     `json:"gender"`
}

// --- helpers ---

func wrapResult(product string, data any) (json.RawMessage, error) {
	return json.Marshal(struct {
		Product string `json:"_product"`
		Data    any    `json:"data"`
	}{Product: product, Data: data})
}

// NormGeo normalizes longitude and timezone defaults.
// Defaults: longitude 120 (Beijing), timezone UTC+8.
func NormGeo(lon, tz float64) (float64, float64) {
	if lon == 0 {
		lon = 120
	}
	if tz == 0 {
		tz = 8
	}
	return lon, tz
}

func parseDaYunZhu(raw json.RawMessage) *bazi.DaYunZhu {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var di struct {
		CurrentZhuIndex int `json:"CurrentZhuIndex"`
		Zhus            []struct {
			Gan int `json:"Gan"`
			Zhi int `json:"Zhi"`
		} `json:"Zhus"`
	}
	if err := json.Unmarshal(raw, &di); err != nil {
		return nil
	}
	if di.CurrentZhuIndex < 0 || di.CurrentZhuIndex >= len(di.Zhus) {
		return nil
	}
	dp := di.Zhus[di.CurrentZhuIndex]
	return &bazi.DaYunZhu{Gan: ganzhi.Gan(dp.Gan), Zhi: ganzhi.Zhi(dp.Zhi)}
}

func parseZhu(raw json.RawMessage) *ganzhi.Zhu {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var z struct {
		Gan int `json:"Gan"`
		Zhi int `json:"Zhi"`
	}
	if err := json.Unmarshal(raw, &z); err != nil {
		return nil
	}
	return &ganzhi.Zhu{Gan: ganzhi.Gan(z.Gan), Zhi: ganzhi.Zhi(z.Zhi)}
}
