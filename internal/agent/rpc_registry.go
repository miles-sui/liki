package agent

import (
	"context"
	"encoding/json"
	"sync"

	"liki/internal/agent/city"
)

// RPC method schemas are defined as inline Go strings rather than in tools.json.
// tools.json serves the 8 naming-chat tools sent to the LLM as function definitions.
// The 29 RPC methods here are external API only (not LLM tools) — their schemas
// drive the OpenRPC document and parameter validation. Keeping them inline avoids
// a second JSON file that would need to stay in sync with handler signatures.

// Common JSON Schema fragments (inline, no $ref — self-contained for AI agents).
const (
	schemaTimePoint = `{"type":"object","properties":{"time":{"type":"string","format":"date-time","description":"出生时间（RFC3339），如 1984-02-04T06:00:00+08:00"},"longitude":{"type":"number","description":"出生地经度，用于真太阳时校正。北京≈116.4"}},"required":["time","longitude"]}`
	schemaGender    = `{"type":"string","enum":["male","female"],"description":"性别"}`
)

var personParamsSchema = json.RawMessage(
	`{"type":"object","properties":{"birth":` + schemaTimePoint + `,"gender":` + schemaGender + `},"required":["birth","gender"]}`,
)

func mustSchema(s string) json.RawMessage {
	if !json.Valid([]byte(s)) {
		panic("invalid JSON schema")
	}
	return json.RawMessage(s)
}

// envelopeSchema wraps a data schema in the standard {"_product":"...","data":<schema>} envelope.
func envelopeSchema(dataSchema string) json.RawMessage {
	return mustSchema(`{"type":"object","properties":{"_product":{"type":"string"},"data":` + dataSchema + `},"required":["_product","data"]}`)
}

// RPCMethod describes a single JSON-RPC method.
type RPCMethod struct {
	Name        string
	Description string
	Params      json.RawMessage // JSON Schema for params
	Result      json.RawMessage // JSON Schema for result (optional)
	Handler     func(context.Context, json.RawMessage) (json.RawMessage, error)
}

// RPCRegistry holds all registered JSON-RPC methods.
type RPCRegistry struct {
	methods map[string]*RPCMethod
	names   []string // registration order preserved for deterministic output

	openrpcOnce sync.Once
	openrpcDoc  json.RawMessage
}

// NewRPCRegistry creates a registry with all 29 external compute methods.
func NewRPCRegistry() *RPCRegistry {
	r := &RPCRegistry{methods: make(map[string]*RPCMethod, 29)}

	// ── bazi (8) ──────────────────────────────────────────────
	r.mustRegister(RPCMethod{
		Name: "bazi.chart", Description: "排八字命盘。返回四柱、十神、藏干、纳音、神煞、用神（扶抑+调候）、格局、身强身弱、大运。",
		Params: personParamsSchema, Handler: computeChartHandler,
		Result: envelopeSchema(`{"type":"object","properties":{"nian":{"type":"object"},"yue":{"type":"object"},"ri":{"type":"object"},"shi":{"type":"object"},"da_yun":{"type":"object"},"fu_yi":{"type":"object"},"tiao_hou":{"type":"object"},"wuxing_count":{"type":"object"},"solar_time":{"type":"string"}},"required":["nian","yue","ri","shi","da_yun","fu_yi","wuxing_count"]}`),
	})
	r.mustRegister(RPCMethod{
		Name: "bazi.bond", Description: "八字合盘。返回双方日主、天干关系（合/生/克）、地支关系（六合/三合/六冲）、纳音配合、五行互补。",
		Params: mustSchema(`{"type":"object","properties":{"a":{"type":"object","properties":{"birth":` + schemaTimePoint + `,"gender":` + schemaGender + `},"required":["birth","gender"]},"b":{"type":"object","properties":{"birth":` + schemaTimePoint + `,"gender":` + schemaGender + `},"required":["birth","gender"]}},"required":["a","b"]}`),
		Handler: computeBondHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"zhu_cross":{"type":"object"},"shi_shen_cross":{"type":"object"},"structure":{"type":"object"}},"required":["zhu_cross","shi_shen_cross","structure"]}`),
	})
	r.mustRegister(RPCMethod{
		Name: "bazi.liunian", Description: "流年运势。返回流年干支与命局的十神、神煞、伏吟反吟。",
		Params: mustSchema(`{"type":"object","properties":{"year":{"type":"integer","description":"目标年份"},"birth":` + schemaTimePoint + `,"gender":` + schemaGender + `},"required":["year","birth","gender"]}`),
		Handler: computeLiunianHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"year":{"type":"integer"},"year_name":{"type":"string"},"shi_shen":{"type":"string"}},"required":["year","year_name","shi_shen"]}`),
	})
	r.mustRegister(RPCMethod{
		Name: "bazi.liuyue", Description: "流月运势。返回流月干支与命局的十神、神煞。",
		Params: mustSchema(`{"type":"object","properties":{"year":{"type":"integer","description":"目标年份"},"month":{"type":"integer","minimum":1,"maximum":12,"description":"目标月份"},"birth":` + schemaTimePoint + `,"gender":` + schemaGender + `},"required":["year","month","birth","gender"]}`),
		Handler: computeLiuyueHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"year":{"type":"integer"},"month":{"type":"integer"},"month_name":{"type":"string"},"shi_shen":{"type":"string"}},"required":["year","month","month_name","shi_shen"]}`),
	})
	r.mustRegister(RPCMethod{
		Name: "bazi.liuri", Description: "流日运势。返回流日干支、十神、纳音。",
		Params: mustSchema(`{"type":"object","properties":{"year":{"type":"integer","description":"目标年份"},"month":{"type":"integer","minimum":1,"maximum":12,"description":"目标月份"},"day":{"type":"integer","minimum":1,"maximum":31,"description":"目标日期"},"birth":` + schemaTimePoint + `,"gender":` + schemaGender + `},"required":["year","month","day","birth","gender"]}`),
		Handler: computeLiuriHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"date":{"type":"string"},"day_name":{"type":"string"},"shi_shen":{"type":"string"}},"required":["date","day_name","shi_shen"]}`),
	})
	r.mustRegister(RPCMethod{
		Name: "bazi.liushi", Description: "流时运势。返回流时干支、十神。hour 为时辰（0-23）。",
		Params: mustSchema(`{"type":"object","properties":{"year":{"type":"integer","description":"目标年份"},"month":{"type":"integer","minimum":1,"maximum":12,"description":"目标月份"},"day":{"type":"integer","minimum":1,"maximum":31,"description":"目标日期"},"hour":{"type":"integer","minimum":0,"maximum":23,"description":"时辰"},"birth":` + schemaTimePoint + `,"gender":` + schemaGender + `},"required":["year","month","day","hour","birth","gender"]}`),
		Handler: computeLiushiHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"time":{"type":"string"},"hour_name":{"type":"string"},"shi_shen":{"type":"string"}},"required":["time","hour_name","shi_shen"]}`),
	})
	r.mustRegister(RPCMethod{
		Name: "bazi.xiaoyun", Description: "小运。返回小运流年列表。count 默认 5。",
		Params: mustSchema(`{"type":"object","properties":{"birth":` + schemaTimePoint + `,"gender":` + schemaGender + `,"count":{"type":"integer","description":"返回年数，默认 5"}},"required":["birth","gender"]}`),
		Handler: computeXiaoYunHandler,
		Result:  envelopeSchema(`{"type":"array","items":{"type":"object","properties":{"age":{"type":"integer"},"gan":{"type":"string"},"zhi":{"type":"string"},"name":{"type":"string"}},"required":["age","gan","zhi","name"]}}`),
	})
	r.mustRegister(RPCMethod{
		Name: "bazi.xiaoxian", Description: "小限。返回小限列表。count 默认 16。",
		Params: mustSchema(`{"type":"object","properties":{"gender":` + schemaGender + `,"count":{"type":"integer","description":"返回年数，默认 16"}},"required":["gender"]}`),
		Handler: computeXiaoXianHandler,
		Result:  envelopeSchema(`{"type":"array","items":{"type":"object","properties":{"age":{"type":"integer"},"branch":{"type":"string"}},"required":["age","branch"]}}`),
	})

	// ── ziwei (6) ─────────────────────────────────────────────
	r.mustRegister(RPCMethod{
		Name: "ziwei.chart", Description: "紫微斗数排盘。返回十二宫星曜分布、亮度、四化。",
		Params: personParamsSchema, Handler: computeZiweiHandler,
		Result: envelopeSchema(`{"type":"object","properties":{"palaces":{"type":"array"},"ming_gong":{"type":"integer"},"si_hua":{"type":"object"},"shen_gong":{"type":"integer"},"ju_shu":{"type":"integer"}},"required":["palaces","ming_gong","si_hua"]}`),
	})
	r.mustRegister(RPCMethod{
		Name: "ziwei.daxian", Description: "紫微斗数大限。返回十年大限各宫吉凶。chart 为 compute_ziwei 返回的完整 chart 对象。",
		Params: mustSchema(`{"type":"object","properties":{"chart":{"type":"object","description":"ziwei.chart 返回的完整 chart 对象"}},"required":["chart"]}`),
		Handler: computeZiweiDaXianHandler,
		Result:  envelopeSchema(`{"type":"array","items":{"type":"object","properties":{"start_age":{"type":"integer"},"end_age":{"type":"integer"},"palace":{"type":"integer"},"name":{"type":"string"}},"required":["start_age","end_age","palace","name"]}}`),
	})
	r.mustRegister(RPCMethod{
		Name: "ziwei.liunian", Description: "紫微流年。返回流年命盘及各宫变化。",
		Params: mustSchema(`{"type":"object","properties":{"liu_year":{"type":"integer","description":"流年年份"},"chart":{"type":"object","description":"ziwei.chart 返回的完整 chart 对象"}},"required":["liu_year","chart"]}`),
		Handler: computeZiweiLiuNianHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"ming_gong":{"type":"integer"},"ming_gong_name":{"type":"string"},"si_hua":{"type":"object"}},"required":["ming_gong","si_hua"]}`),
	})
	r.mustRegister(RPCMethod{
		Name: "ziwei.liuyue", Description: "紫微流月。返回流月命盘及各宫变化。",
		Params: mustSchema(`{"type":"object","properties":{"liu_year":{"type":"integer","description":"流年年份"},"lunar_month":{"type":"integer","minimum":1,"maximum":12,"description":"农历月份"},"chart":{"type":"object","description":"ziwei.chart 返回的完整 chart 对象"}},"required":["liu_year","lunar_month","chart"]}`),
		Handler: computeZiweiLiuYueHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"ming_gong":{"type":"integer"},"ming_gong_name":{"type":"string"},"si_hua":{"type":"object"}},"required":["ming_gong","si_hua"]}`),
	})
	r.mustRegister(RPCMethod{
		Name: "ziwei.liuri", Description: "紫微流日。返回流日命盘及各宫变化。",
		Params: mustSchema(`{"type":"object","properties":{"liu_year":{"type":"integer","description":"流年年份"},"lunar_month":{"type":"integer","minimum":1,"maximum":12,"description":"农历月份"},"lunar_day":{"type":"integer","minimum":1,"maximum":30,"description":"农历日期"},"chart":{"type":"object","description":"ziwei.chart 返回的完整 chart 对象"}},"required":["liu_year","lunar_month","lunar_day","chart"]}`),
		Handler: computeZiweiLiuRiHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"ming_gong":{"type":"integer"},"ming_gong_name":{"type":"string"},"si_hua":{"type":"object"}},"required":["ming_gong","si_hua"]}`),
	})
	r.mustRegister(RPCMethod{
		Name: "ziwei.bond", Description: "紫微合盘。返回双方命盘交互分析。",
		Params: mustSchema(`{"type":"object","properties":{"a":{"type":"object","description":"甲方紫微盘（ziwei.chart 返回的完整对象）"},"b":{"type":"object","description":"乙方紫微盘（ziwei.chart 返回的完整对象）"}},"required":["a","b"]}`),
		Handler: computeZiweiBondHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"a_into_b":{"type":"integer"},"b_into_a":{"type":"integer"},"star_cross":{"type":"array"}},"required":["a_into_b","b_into_a","star_cross"]}`),
	})

	// ── qimen (1) ─────────────────────────────────────────────
	r.mustRegister(RPCMethod{
		Name: "qimen.pan", Description: "奇门排盘。返回天盘、人盘、神盘、九星八门格局。kind 默认 shi（时家奇门），可选 ri/yue/nian。",
		Params: mustSchema(`{"type":"object","properties":{"birth":` + schemaTimePoint + `,"kind":{"type":"string","enum":["shi","ri","yue","nian"],"description":"奇门类型，默认 shi"}},"required":["birth"]}`),
		Handler: computeQimenHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"pan":{"type":"object"},"patterns":{"type":"array"}},"required":["pan","patterns"]}`),
	})

	// ── qiming (4) ────────────────────────────────────────────
	r.mustRegister(RPCMethod{
		Name: "qiming.wuge", Description: "起名五格。返回三才五格组合 + 可用字库。yong_shen 取值 木/火/土/金/水。",
		Params: mustSchema(`{"type":"object","properties":{"surname":{"type":"string","minLength":1,"maxLength":2,"description":"姓氏"},"yong_shen":{"type":"string","enum":["木","火","土","金","水"],"description":"用神五行"},"xi_shen":{"type":"array","items":{"type":"string","enum":["木","火","土","金","水"]},"description":"喜神五行（可选）"}},"required":["surname","yong_shen"]}`),
		Handler: computeNamingWuGeHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"surname":{"type":"string"},"combos":{"type":"array"},"yong_chars":{"type":"object"}},"required":["surname","combos","yong_chars"]}`),
	})
	r.mustRegister(RPCMethod{
		Name: "qiming.compose", Description: "起名组名。从上一步五格组合中选取 combos 生成候选名字。",
		Params: mustSchema(`{"type":"object","properties":{"surname":{"type":"string","minLength":1,"maxLength":2,"description":"姓氏"},"combos":{"type":"array","description":"五格组合（从 wuge 返回的 combos 中选取）"},"yong_chars":{"type":"object","description":"用神字库（从 wuge 返回的 yong_chars）"},"xi_chars":{"type":"object","description":"喜神字库（从 wuge 返回的 xi_chars，可选）"}},"required":["surname","combos","yong_chars"]}`),
		Handler: computeNamingComposeHandler,
		Result:  envelopeSchema(`{"type":"array","items":{"type":"string"}}`),
	})
	r.mustRegister(RPCMethod{
		Name: "qiming.detail", Description: "起名详析。对候选名字逐一详析，返回五格数理、三才配置、五行、音韵。",
		Params: mustSchema(`{"type":"object","properties":{"surname":{"type":"string","minLength":1,"maxLength":2,"description":"姓氏"},"names":{"type":"array","items":{"type":"string"},"description":"候选名字列表"}},"required":["surname","names"]}`),
		Handler: computeNamingDetailHandler,
		Result: envelopeSchema(`{"type":"array","items":{"type":"object","properties":{"name":{"type":"string"},"characters":{"type":"array"},"wu_ge":{"type":"object"},"san_cai":{"type":"object"},"phonetic":{"type":"object","properties":{"tones":{"type":"string"}},"required":["tones"]}},"required":["name","wu_ge","san_cai","phonetic"]}}`),
	})
	r.mustRegister(RPCMethod{
		Name: "qiming.evaluate", Description: "起名评估。用户自选名字时评估单名。",
		Params: mustSchema(`{"type":"object","properties":{"surname":{"type":"string","minLength":1,"maxLength":2,"description":"姓氏"},"given_name":{"type":"string","description":"名字"},"yong_shen":{"type":"string","enum":["木","火","土","金","水"],"description":"用神五行"}},"required":["surname","given_name","yong_shen"]}`),
		Handler: computeNamingEvaluateHandler,
		Result: envelopeSchema(`{"type":"object","properties":{"surname":{"type":"string"},"given_name":{"type":"string"},"wuxing_match":{"type":"boolean"},"phonetic":{"type":"object","properties":{"tones":{"type":"string"}},"required":["tones"]}},"required":["surname","given_name","wuxing_match","phonetic"]}`),
	})

	// ── bazhai (2) ────────────────────────────────────────────
	r.mustRegister(RPCMethod{
		Name: "bazhai.chart", Description: "八宅风水。综合命卦与飞星分析。",
		Params: personParamsSchema, Handler: computeBazhaiHandler,
		Result: envelopeSchema(`{"type":"object","properties":{"ming_gua":{"type":"object"},"ba_zhai_dirs":{"type":"object"},"pillar_bagua":{"type":"array"}},"required":["ming_gua","ba_zhai_dirs","pillar_bagua"]}`),
	})
	r.mustRegister(RPCMethod{
		Name: "bazhai.minggua", Description: "命卦查询。返回东四命/西四命 + 命卦 + 四吉四凶方。",
		Params: mustSchema(`{"type":"object","properties":{"gender":` + schemaGender + `,"birth_year":{"type":"integer","description":"出生年份"}},"required":["gender","birth_year"]}`),
		Handler: computeMingGuaHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"gua":{"type":"object","properties":{"index":{"type":"integer"},"name":{"type":"string"},"wuxing":{"type":"string"},"yin_yang":{"type":"string"}},"required":["index","name","wuxing","yin_yang"]},"gua_number":{"type":"integer"},"group":{"type":"string"}},"required":["gua","gua_number","group"]}`),
	})

	// ── xuankong (2) ──────────────────────────────────────────
	r.mustRegister(RPCMethod{
		Name: "xuankong.sanyuan", Description: "三元九运查询。返回当前三元九运的时间表。",
		Params: mustSchema(`{"type":"object","properties":{"year":{"type":"integer","description":"年份"}},"required":["year"]}`),
		Handler: computeSanYuanYunHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"year":{"type":"integer"},"yuan":{"type":"string"},"yun_number":{"type":"integer"}},"required":["year","yuan","yun_number"]}`),
	})
	r.mustRegister(RPCMethod{
		Name: "xuankong.chart", Description: "玄空飞星。返回山向飞星盘。sit_mountain/face_mountain 为坐向（0-23）。",
		Params: mustSchema(`{"type":"object","properties":{"birth":` + schemaTimePoint + `,"sit_mountain":{"type":"integer","minimum":0,"maximum":23,"description":"山向"},"face_mountain":{"type":"integer","minimum":0,"maximum":23,"description":"朝向"}},"required":["birth","sit_mountain","face_mountain"]}`),
		Handler: computeXuankongHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"yun":{"type":"object"},"palaces":{"type":"array"},"wang_shan":{"type":"boolean"}},"required":["yun","palaces","wang_shan"]}`),
	})

	// ── liuyao (2) ────────────────────────────────────────────
	r.mustRegister(RPCMethod{
		Name: "liuyao.qigua", Description: "六爻起卦。摇卦（三枚铜钱起六次），返回原始爻值和动爻位置。",
		Params: mustSchema(`{"type":"object","properties":{},"required":[]}`),
		Handler: computeLiuyaoQiguaHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"yaos":{"type":"array","items":{"type":"integer"}, "description":"六爻值 6-9"},"dong_yao":{"type":"array","items":{"type":"integer"},"description":"动爻位置 1-6"}},"required":["yaos","dong_yao"]}`),
	})
	r.mustRegister(RPCMethod{
		Name: "liuyao.chart", Description: "六爻装卦。传入起卦结果和问事时辰，装卦并分析：纳甲、六亲、六兽、用神、旺衰、应期。",
		Params: mustSchema(`{"type":"object","properties":{"birth":` + schemaTimePoint + `,"yong_shen":{"type":"string","description":"用神六亲（如 妻财/官鬼/父母/兄弟/子孙/世爻），可选，默认世爻"},"yaos":{"type":"array","items":{"type":"integer"},"minItems":6,"maxItems":6,"description":"六爻值（6-9），必填，先调 liuyao.qigua 获取"}},"required":["birth","yaos"]}`),
		Handler: computeLiuyaoHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"name":{"type":"string"},"ben_gua":{"type":"integer"},"lines":{"type":"array"},"yong_shen":{"type":"object"},"wang_shuai":{"type":"array"},"ying_qi":{"type":"object"}},"required":["name","ben_gua","lines","yong_shen"]}`),
	})

	// ── huangli (4) ───────────────────────────────────────────
	r.mustRegister(RPCMethod{
		Name: "huangli.date", Description: "按日查宜忌。查询指定日期的黄历宜忌。",
		Params: mustSchema(`{"type":"object","properties":{"date":{"type":"string","description":"日期 YYYY-MM-DD"},"event":{"type":"string","description":"事项（如 嫁娶/开业/搬家）"}},"required":["date","event"]}`),
		Handler: queryHuangliDateHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"date":{"type":"string"},"day_pillar":{"type":"object"},"suitable":{"type":"boolean"}},"required":["date","day_pillar","suitable"]}`),
	})
	r.mustRegister(RPCMethod{
		Name: "huangli.month", Description: "按月查宜忌。返回当月每日宜忌汇总。",
		Params: mustSchema(`{"type":"object","properties":{"month":{"type":"string","description":"月份 YYYY-MM"},"event":{"type":"string","description":"事项"}},"required":["month","event"]}`),
		Handler: queryHuangliMonthHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"month":{"type":"string"},"days":{"type":"array"}},"required":["month","days"]}`),
	})
	r.mustRegister(RPCMethod{
		Name: "huangli.bond.date", Description: "八字合参择日。基于命主八字筛选单日宜忌。",
		Params: mustSchema(`{"type":"object","properties":{"birth":` + schemaTimePoint + `,"event_type":{"type":"string","description":"事项"},"date":{"type":"string","description":"日期 YYYY-MM-DD"}},"required":["birth","event_type","date"]}`),
		Handler: queryHuangliBondDateHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"date":{"type":"string"},"gan_relation":{"type":"string"},"zhi_relation":{"type":"string"}},"required":["date","gan_relation","zhi_relation"]}`),
	})
	r.mustRegister(RPCMethod{
		Name: "huangli.bond.month", Description: "八字合参择月。基于命主八字筛选当月吉日。",
		Params: mustSchema(`{"type":"object","properties":{"birth":` + schemaTimePoint + `,"event_type":{"type":"string","description":"事项"},"month":{"type":"string","description":"月份 YYYY-MM"}},"required":["birth","event_type","month"]}`),
		Handler: queryHuangliBondMonthHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"month":{"type":"string"},"days":{"type":"array"}},"required":["month","days"]}`),
	})

	// ── infra (2) ─────────────────────────────────────────────
	r.mustRegister(RPCMethod{
		Name: "time.now", Description: "获取服务端当前时间。返回 UTC、本地、北京时间，用于 AI agent 获取准确时间避免幻觉。",
		Params: mustSchema(`{"type":"object","properties":{},"required":[]}`),
		Handler: nowTimeHandler,
		Result:  mustSchema(`{"type":"object","properties":{"utc":{"type":"string"},"local":{"type":"string"},"cst":{"type":"string"}},"required":["utc","cst"]}`),
	})
	r.mustRegister(RPCMethod{
		Name: "city", Description: "根据城市名查询经纬度。基于 Nominatim 服务。",
		Params: mustSchema(`{"type":"object","properties":{"city":{"type":"string","description":"城市名称"}},"required":["city"]}`),
		Handler: city.SearchCity,
		Result:  mustSchema(`{"type":"object","properties":{"name":{"type":"string"},"longitude":{"type":"number"},"latitude":{"type":"number"},"country":{"type":"string"}},"required":["name","longitude","latitude","country"]}`),
	})

	return r
}

func (r *RPCRegistry) mustRegister(m RPCMethod) {
	if _, exists := r.methods[m.Name]; exists {
		panic("duplicate RPC method: " + m.Name)
	}
	r.methods[m.Name] = &m
	r.names = append(r.names, m.Name)
}

// Execute runs the handler for the given method name with raw JSON params.
func (r *RPCRegistry) Execute(ctx context.Context, method string, params json.RawMessage) (json.RawMessage, error) {
	m, ok := r.methods[method]
	if !ok {
		return nil, &RPCError{Code: -32601, Message: "Method not found: " + method}
	}
	result, err := m.Handler(ctx, params)
	if err != nil {
		return nil, &RPCError{Code: -32000, Message: err.Error()}
	}
	return result, nil
}

// RPCError implements error and serializes as a JSON-RPC error object.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *RPCError) Error() string { return e.Message }

// ── OpenRPC document generation ──────────────────────────────

type openRPCDoc struct {
	OpenRPC string        `json:"openrpc"`
	Info    openRPCInfo   `json:"info"`
	Methods []openRPCMeth `json:"methods"`
}

type openRPCInfo struct {
	Title   string `json:"title"`
	Version string `json:"version"`
}

type openRPCMeth struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Params      json.RawMessage `json:"params"`
	Result      json.RawMessage `json:"result,omitempty"`
}

// OpenRPCDocument returns the full OpenRPC 1.4.1 document as raw JSON.
func (r *RPCRegistry) OpenRPCDocument() json.RawMessage {
	r.openrpcOnce.Do(func() {
		methods := make([]openRPCMeth, 0, len(r.names)+1)

		methods = append(methods, openRPCMeth{
			Name:        "rpc.discover",
			Description: "返回此 OpenRPC 1.4.1 document，包含所有可用 method 及参数定义。",
			Params:      json.RawMessage(`{"type":"object","properties":{},"description":"无需参数"}`),
			Result:      json.RawMessage(`{"type":"object","description":"OpenRPC 1.4.1 完整文档"}`),
		})

		for _, name := range r.names {
			m := r.methods[name]
			methods = append(methods, openRPCMeth{
				Name:        m.Name,
				Description: m.Description,
				Params:      m.Params,
				Result:      m.Result,
			})
		}

		doc := openRPCDoc{
			OpenRPC: "1.4.1",
			Info: openRPCInfo{
				Title:   "Liki (灵机) JSON-RPC API",
				Version: "1.0.0",
			},
			Methods: methods,
		}

		b, err := json.Marshal(doc)
		if err != nil {
			panic("marshal OpenRPC document: " + err.Error())
		}
		r.openrpcDoc = json.RawMessage(b)
	})
	return r.openrpcDoc
}
