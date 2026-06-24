package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"liki/internal/llm"
)

// reportSections maps product to required section titles.
var reportSections = map[Product][]string{
	ProductChart: {
		"一、格局总论", "二、用神详解", "三、四柱十神分析",
		"四、大运提示", "五、流年分析",
	},
	ProductBond: {
		"一、双方八字概览", "二、天干互动分析", "三、地支配合分析",
		"四、十神互动分析", "五、五行与用神互补", "六、神煞互动",
		"七、大运同步与结构", "八、综合建议",
	},
	ProductNaming: {
		"一、命理基础与用神", "二、候选名字速览",
		"三、候选名字逐一分析", "四、横向对比与推荐",
	},
	ProductZiwei: {
		"一、命盘总览", "二、命宫详解", "三、十二宫逐一解读",
		"四、四化飞布", "五、格局", "六、大限", "七、流年",
	},
	ProductBazhai: {
		"一、命卦定位", "二、四吉四凶方位",
		"三、年飞星分析", "四、四柱八卦",
	},
	ProductXuankong: {
		"一、三元九运与坐向", "二、九宫飞星盘", "三、格局判断",
		"四、双星加会", "五、收山出煞", "六、综合建议",
	},
}

// conditionalSections lists sections that are only required when specific data is present.
var conditionalSections = map[string]string{
	"五、流年分析": "has_liunian",
}

// CheckToolRegistry implements ToolRegistry for report verification tools.
type CheckToolRegistry struct {
	handlers map[string]func(context.Context, json.RawMessage) (json.RawMessage, error)
	defs     []json.RawMessage
}

// NewCheckToolRegistry creates a new CheckToolRegistry with verification tools.
func NewCheckToolRegistry() *CheckToolRegistry {
	r := &CheckToolRegistry{
		handlers: map[string]func(context.Context, json.RawMessage) (json.RawMessage, error){},
	}
	r.register("verify_terminology", verifyTerminology, "校验报告中使用的中文命理术语是否合法。传入terms数组，返回不在术语表中的未知术语列表。")
	r.register("verify_chart_data", verifyChartData, "校验报告中的关键数据引用是否与原始数据一致。传入数据路径和报告中声称的值，返回是否匹配及实际值。")
	r.register("verify_structure", verifyStructure, "校验报告结构是否完整。传入报告中的章节标题列表和产品名，返回缺失的必选章节。")
	return r
}

func (r *CheckToolRegistry) register(name string, h func(context.Context, json.RawMessage) (json.RawMessage, error), desc string) {
	r.handlers[name] = h
	fn := map[string]any{
		"name":        name,
		"description": desc,
		"parameters": map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}
	b, err := json.Marshal(fn)
	if err != nil {
		panic("marshal tool def: " + err.Error())
	}
	r.defs = append(r.defs, b)
}

// Execute runs the named check tool with the given arguments.
func (r *CheckToolRegistry) Execute(ctx context.Context, name string, args json.RawMessage) (json.RawMessage, error) {
	h, ok := r.handlers[name]
	if !ok {
		return nil, fmt.Errorf("unknown check tool: %s", name)
	}
	return h(ctx, args)
}

// Schemas returns the check tool definitions for the LLM.
func (r *CheckToolRegistry) Schemas() []llm.ToolDef {
	out := make([]llm.ToolDef, len(r.defs))
	for i, d := range r.defs {
		out[i] = llm.ToolDef{Type: "function", Function: d}
	}
	return out
}

// verifyTerminology checks each term against the known terminology set.
func verifyTerminology(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	var input struct {
		Terms []string `json:"terms"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return json.RawMessage(`{"unknown":["invalid input"]}`), nil
	}

	var unknown []string
	for _, t := range input.Terms {
		if !knownTerms[t] {
			unknown = append(unknown, t)
		}
	}
	if unknown == nil {
		unknown = []string{}
	}
	out, _ := json.Marshal(map[string]any{"unknown": unknown})
	return out, nil
}

// verifyChartData checks that a claimed value matches the actual value in chart data.
func verifyChartData(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	var input struct {
		Chart    json.RawMessage `json:"chart"`
		Path     string          `json:"path"`
		Expected string          `json:"expected"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return json.RawMessage(`{"match":false,"actual":"","error":"invalid input"}`), nil
	}

	actual := gjsonGet(input.Chart, input.Path)
	match := actual == input.Expected
	out, _ := json.Marshal(map[string]any{"match": match, "actual": actual})
	return out, nil
}

// verifyStructure checks that all required sections are present in the report.
func verifyStructure(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	var input struct {
		Sections  []string `json:"sections"`
		Product   string   `json:"product"`
		HasLiuNian *bool   `json:"has_liunian"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return json.RawMessage(`{"missing":["invalid input"]}`), nil
	}

	required, ok := reportSections[Product(input.Product)]
	if !ok {
		return json.RawMessage(`{"missing":[]}`), nil
	}

	var missing []string
	for _, sec := range required {
		if slices.Contains(input.Sections, sec) {
			continue
		}
		// Check if this is a conditional section that's not needed
		if cond, isCond := conditionalSections[sec]; isCond {
			if cond == "has_liunian" && input.HasLiuNian != nil && !*input.HasLiuNian {
				continue
			}
		}
		missing = append(missing, sec)
	}
	if missing == nil {
		missing = []string{}
	}
	out, _ := json.Marshal(map[string]any{"missing": missing})
	return out, nil
}

// gjsonGet navigates a JSON document using dot-separated path and returns the value as a string.
func gjsonGet(data json.RawMessage, path string) string {
	parts := strings.Split(path, ".")
	var current any
	if err := json.Unmarshal(data, &current); err != nil {
		return ""
	}

	for _, part := range parts {
		m, ok := current.(map[string]any)
		if !ok {
			return ""
		}
		v, ok := m[part]
		if !ok {
			return ""
		}
		current = v
	}

	switch v := current.(type) {
	case string:
		return v
	case float64:
		return fmt.Sprintf("%v", v)
	case bool:
		return fmt.Sprintf("%v", v)
	default:
		b, _ := json.Marshal(v)
		return string(b)
	}
}

// knownTerms is the set of all valid Chinese terminology from docs/terminology.md.
var knownTerms = buildKnownTerms()

func buildKnownTerms() map[string]bool {
	terms := []string{
		// 基础
		"八字", "阴阳", "阳", "阴",
		// 干支
		"天干", "地支", "干支", "柱",
		"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸",
		"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥",
		// 五行
		"五行", "木", "火", "土", "金", "水", "生", "克",
		// 四柱
		"年柱", "月柱", "日柱", "时柱",
		// 日主
		"日元", "日主",
		// 十神
		"十神", "比肩", "劫财", "食神", "伤官", "偏财", "正财", "七杀", "正官", "偏印", "正印",
		"食神制杀", "财破印", "官印相生", "伤官见官", "财生官",
		// 藏干
		"藏干", "本气", "中气", "余气",
		// 人元司令分野
		"人元司令分野", "人元",
		// 长生十二宫
		"长生十二宫", "长生", "沐浴", "冠带", "临官", "帝旺", "衰", "病", "死", "墓", "绝", "胎", "养",
		// 纳音
		"纳音",
		// 用神
		"用神", "喜神", "忌神", "调候", "扶抑", "格局", "强弱", "旺衰",
		// 身强身弱
		"身强", "身弱", "中和",
		// 大运
		"大运", "大运柱", "起运岁数", "顺逆",
		// 流年
		"流年", "流月", "流日", "流时",
		// 小运小限
		"小运", "小限",
		// 伏吟反吟
		"伏吟", "反吟",
		// 合会冲刑害
		"合会", "合", "天干合", "地支合", "三合", "三会", "六合", "冲", "刑", "害", "拱夹",
		"天干五合", "地支六合", "六冲", "六害", "三刑",
		// 神煞
		"神煞", "天乙贵人", "桃花", "驿马", "空亡", "魁罡", "日德", "日贵", "禄",
		"文昌", "羊刃", "华盖", "天德", "月德", "劫煞", "孤辰", "寡宿",
		// 胎元命宫身宫
		"胎元", "命宫", "身宫",
		// 合盘
		"合盘", "天干关系", "地支关系", "柱柱关系", "十神互看", "纳音关系", "神煞共现",
		// 起名
		"五格", "天格", "人格", "地格", "外格", "总格", "三才", "三奇",
		// 紫微
		"紫微斗数", "命宫", "兄弟", "夫妻", "子女", "财帛", "疾厄", "迁移", "交友", "官禄", "田宅", "福德", "父母",
		"四化", "禄", "权", "科", "忌", "庙", "旺", "利", "平", "陷", "局数", "身宫",
		// 风水
		"风水", "命卦", "东四命", "西四命", "生气", "天医", "延年", "伏位", "祸害", "五鬼", "六煞", "绝命",
		"飞星", "三元", "三元九运", "二十四山",
		// 六爻
		"六爻", "六亲", "六兽", "伏神", "旬空",
		// 奇门
		"奇门", "式盘", "天盘", "人盘", "神盘", "营气", "局数", "暗干",
		// 黄历
		"黄历", "节气", "黄道", "值日星宿",
		// 常见组合
		"得令", "失令", "得地", "失地", "得势", "失势", "通关", "病药",
		"盖头", "截脚", "自合", "命盘", "排盘",
		"正官格", "七杀格", "正印格", "偏印格", "正财格", "偏财格", "食神格", "伤官格", "建禄格", "从强格",
		"夫妻宫", "子女宫", "父母宫", "事业宫",
	}

	m := make(map[string]bool, len(terms))
	for _, t := range terms {
		m[t] = true
	}
	return m
}
