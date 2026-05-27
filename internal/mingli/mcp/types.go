package mcp

import (
	"github.com/25types/25types/internal/ganzhi"
	"github.com/25types/25types/internal/mingli/bazi"
	"github.com/25types/25types/internal/mingli/fengshui"
	"github.com/25types/25types/internal/mingli/huangli"
	"github.com/25types/25types/internal/mingli/qiming"
)

// ---- Shared birth profile ----

// BirthProfile holds the birth parameters shared across bazi tools.
type BirthProfile struct {
	Year      int     `json:"year" jsonschema:"出生年"`
	Month     int     `json:"month" jsonschema:"出生月 (1-12)"`
	Day       int     `json:"day" jsonschema:"出生日 (1-31)"`
	Hour      int     `json:"hour" jsonschema:"出生时 (0-23)"`
	Minute    int     `json:"minute,omitempty" jsonschema:"出生分 (0-59), 默认 0"`
	Longitude float64 `json:"longitude,omitempty" jsonschema:"经度, 默认 120.0 (北京时间)"`
	Timezone  float64 `json:"timezone,omitempty" jsonschema:"时区小时偏移, 默认 8 (UTC+8)"`
	Gender    string  `json:"gender" jsonschema:"性别: male 或 female"`
}

// ---- Tool input structs ----

// BaziChartInput is the input for bazi_chart.
type BaziChartInput struct {
	Birth BirthProfile `json:"birth" jsonschema:"出生信息"`
}

// BaziMatchInput is the input for bazi_bond.
type BaziMatchInput struct {
	A BirthProfile `json:"a" jsonschema:"第一个人的出生信息"`
	B BirthProfile `json:"b" jsonschema:"第二个人的出生信息"`
}

// BaziLiunianInput is the input for bazi_liunian.
type BaziLiunianInput struct {
	Bazi         ganzhi.Bazi     `json:"bazi" jsonschema:"四柱(八字)"`
	Year         int              `json:"year" jsonschema:"流年"`
	CurrentDayun *ganzhi.Pillar   `json:"current_dayun,omitempty" jsonschema:"当前大运柱"`
}

// QimingGenerateInput is the input for qiming_generate.
type QimingGenerateInput struct {
	Surname  string   `json:"surname" jsonschema:"姓氏"`
	YongShen string   `json:"yong_shen" jsonschema:"用神五行 (金/木/水/火/土)"`
	XiShen   []string `json:"xi_shen" jsonschema:"喜神五行"`
	Zodiac   int      `json:"zodiac" jsonschema:"生肖年支 (1-12)"`
	Limit    int      `json:"limit" jsonschema:"返回数量 (1-50), 默认 20"`
}

// QimingEvaluateInput is the input for qiming_evaluate.
type QimingEvaluateInput struct {
	Surname   string `json:"surname" jsonschema:"姓氏"`
	GivenName string `json:"given_name" jsonschema:"名字 (1-2 汉字)"`
	YongShen  string `json:"yong_shen" jsonschema:"用神五行"`
	Zodiac    int    `json:"zodiac" jsonschema:"生肖年支 (1-12)"`
}

// FengshuiMingguaInput is the input for fengshui_minggua.
type FengshuiMingguaInput struct {
	Year   int    `json:"year" jsonschema:"出生年"`
	Gender string `json:"gender" jsonschema:"性别: male 或 female"`
}

// HuangliQueryInput is the input for huangli_query.
type HuangliQueryInput struct {
	Date      string `json:"date,omitempty" jsonschema:"日期 YYYY-MM-DD"`
	Month     string `json:"month,omitempty" jsonschema:"月份 YYYY-MM"`
	EventType string `json:"event_type,omitempty" jsonschema:"事件类型 (wedding/engage/open/sign/move 等)"`
}

// HuangliBondInput is the input for huangli_bond.
type HuangliBondInput struct {
	Birth     BirthProfile `json:"birth_info" jsonschema:"出生信息"`
	Month     string       `json:"month,omitempty" jsonschema:"月份 YYYY-MM"`
	Date      string       `json:"date,omitempty" jsonschema:"日期 YYYY-MM-DD"`
	EventType string       `json:"event_type,omitempty" jsonschema:"事件类型"`
}

// HuangliQueryOutput wraps query results (single day or full month).
type HuangliQueryOutput struct {
	Days      []huangli.DayEntry `json:"days"`
	YearMonth string             `json:"year_month"`
}

// HuangliBondOutput wraps bond results (single day or full month).
type HuangliBondOutput struct {
	Days      []huangli.BondDayEntry `json:"days"`
	YearMonth string                  `json:"year_month"`
}

// ---- Tool output structs ----

// ChartOutput is the enriched bazi chart output.
type ChartOutput = bazi.ChartOutput

// BondOutput is the bazi bond output.
type BondOutput = bazi.BondOutput

// QimingGenerateOutput is the qiming generation output.
type QimingGenerateOutput struct {
	Surname        string                 `json:"surname"`
	SurnameElement string                 `json:"surname_element"`
	YongShen       string                 `json:"yong_shen"`
	XiShen         []string               `json:"xi_shen"`
	ZodiacHint     qiming.ZodiacHint      `json:"zodiac_hint"`
	Candidates     []qiming.NameCandidate `json:"candidates"`
}

// FengshuiMingguaOutput holds the minggua result with all trigrams.
type FengshuiMingguaOutput struct {
	Gua         fengshui.Trigram    `json:"gua"`
	GuaNumber   int                 `json:"gua_number"`
	Group       string              `json:"group"`
	AllTrigrams [9]fengshui.Trigram `json:"all_trigrams"`
}

// FengshuiHeCanInput is the input for fengshui_hecan.
type FengshuiHeCanInput struct {
	BirthYear int                `json:"birth_year" jsonschema:"出生年"`
	Gender    string             `json:"gender" jsonschema:"性别: male 或 female"`
	Bazi      ganzhi.Bazi        `json:"bazi" jsonschema:"四柱(八字)"`
	YongShen  bazi.YongShenResult `json:"yong_shen" jsonschema:"八字用神（扶抑+调候）"`
	Year      int                `json:"year" jsonschema:"目标年份"`
}



