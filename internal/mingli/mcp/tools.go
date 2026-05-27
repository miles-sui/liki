package mcp

import (
	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterTools registers all mingli MCP tools on the provided server.
func RegisterTools(s *mcp.Server) {
	// BaZi tools
	mcp.AddTool(s, &mcp.Tool{
		Name:         "bazi_chart",
		Description:  "Compute a full BaZi (八字) chart from birth information. Returns the four pillars, ten gods, na yin, hidden stems, life stages, big fortune (大运), day master strength, yong/xi/ji shen, and related metadata.",
		Title:        "八字排盘",
		OutputSchema: &jsonschema.Schema{Type: "object"},
	}, HandleBaziChart)

	mcp.AddTool(s, &mcp.Tool{
		Name:         "bazi_bond",
		Description:  "Compute BaZi bond (合盘/对敲) cross-chart analysis. Returns full pillar cross (16 pairs), ten-god mutual perspective, nayin element relations, shensha mutual occurrence, and structural comparisons (taiyuan/minggong/dayun/xun). No subjective scoring — structured factual data only.",
		Title:        "八字合盘",
		OutputSchema: &jsonschema.Schema{Type: "object"},
	}, HandleBaziBond)

	mcp.AddTool(s, &mcp.Tool{
		Name:         "bazi_liunian",
		Description:  "Compute yearly fortune (流年运势) for a given year. Returns the year's ten god relationship, generating/restraining effects, bazi interactions, dayun interactions, shensha (神煞), and fuyin/fanyin analysis.",
		Title:        "流年运势",
		OutputSchema: &jsonschema.Schema{Type: "object"},
	}, HandleBaziLiunian)

	// Naming tools
	mcp.AddTool(s, &mcp.Tool{
		Name:         "qiming_generate",
		Description:  "Generate Chinese name (起名) candidates based on surname, yong shen, xi shen, and zodiac. Each candidate includes wu ge (五格), san cai (三才), phonetic analysis, and highlights. Returns up to 20 candidates by default.",
		Title:        "起名",
		OutputSchema: &jsonschema.Schema{Type: "object"},
	}, HandleQimingGenerate)

	mcp.AddTool(s, &mcp.Tool{
		Name:         "qiming_evaluate",
		Description:  "Evaluate a specific Chinese name (测名) against wuxing requirements and zodiac compatibility. Returns wu ge (五格) scores, san cai (三才) configuration, phonetic mark, wuxing match status, and zodiac notes.",
		Title:        "测名",
		OutputSchema: &jsonschema.Schema{Type: "object"},
	}, HandleQimingEvaluate)

	// Fengshui tools
	mcp.AddTool(s, &mcp.Tool{
		Name:         "fengshui_minggua",
		Description:  "Compute the fate trigram (命卦) using the Eight Mansions (八宅) method from birth year and gender. Returns the personal trigram, gua number, East/West group classification, and all 8 trigrams for reference.",
		Title:        "命卦计算",
		OutputSchema: &jsonschema.Schema{Type: "object"},
	}, HandleFengshuiMinggua)

	mcp.AddTool(s, &mcp.Tool{
		Name:         "fengshui_hecan",
		Description:  "Combined Feng Shui reference (风水合参). Takes birth year, gender, four pillars, BaZi yong-shen (from bazi_chart), and target year. Returns fate trigram (命卦, computed internally), Eight Mansions four-auspicious-four-inauspicious directions (八宅四吉四凶), annual purple-white flying stars (年紫白飞星), pillar bagua (四柱纳甲卦), and yong-shen pass-through. No scoring — each system speaks for itself.",
		Title:        "风水合参",
		OutputSchema: &jsonschema.Schema{Type: "object"},
	}, HandleFengshuiHeCan)

	// Huangli tools
	mcp.AddTool(s, &mcp.Tool{
		Name:         "huangli_query",
		Description:  "Query huangli (黄历) information for a single date or full month. Returns day pillar (stem/branch/na yin), jianchu (建除) god with suitability marks/warnings, huangdao star, auspicious directions, Peng Zu taboos, and day mansion. Pass 'event_type' to get jianchu suitability judged for a specific event (wedding/open/sign/move etc.).",
		Title:        "黄历查询",
		OutputSchema: &jsonschema.Schema{Type: "object"},
	}, HandleHuangliQuery)

	mcp.AddTool(s, &mcp.Tool{
		Name:         "huangli_bond",
		Description:  "Cross-reference (对敲) birth info against huangli days. Returns everything from huangli_query plus personal annotations: gan relation (day stem vs day master), zhi relation (day branch vs birth day pillar), and tai sui relation (day branch vs year branch). Use for personalized date selection (择日).",
		Title:        "黄历对敲",
		OutputSchema: &jsonschema.Schema{Type: "object"},
	}, HandleHuangliBond)


}
