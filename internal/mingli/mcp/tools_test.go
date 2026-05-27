package mcp

import (
	"context"
	"testing"

	"github.com/25types/25types/internal/ganzhi"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func newTestServer() *mcp.Server {
	return mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.0.0"}, nil)
}

func TestRegisterTools(t *testing.T) {
	s := newTestServer()
	RegisterTools(s)
	// Registration without panicking is success.
}

func TestHandleBaziChart(t *testing.T) {
	input := BaziChartInput{
		Birth: BirthProfile{Year: 1984, Month: 3, Day: 15, Hour: 8, Minute: 0, Longitude: 120, Timezone: 120, Gender: "male"},
	}
	_, out, err := HandleBaziChart(context.Background(), nil, input)
	if err != nil {
		t.Fatalf("HandleBaziChart: %v", err)
	}
	if out.DayMaster == "" {
		t.Error("day_master is empty")
	}
	if out.YearPillar.Stem == 0 || out.MonthPillar.Stem == 0 || out.DayPillar.Stem == 0 || out.HourPillar.Stem == 0 {
		t.Error("missing pillar data")
	}
	if out.Dayun == nil || len(out.Dayun.Pillars) == 0 {
		t.Error("dayun missing")
	}
	if out.YongShen.FuYi.Strength == "" || out.YongShen.FuYi.Yong == "" {
		t.Error("yong_shen.fuyi missing")
	}
}

func TestHandleBaziBond(t *testing.T) {
	bp := BirthProfile{Year: 1990, Month: 5, Day: 20, Hour: 12, Gender: "male"}
	input := BaziMatchInput{A: bp, B: bp}
	_, out, err := HandleBaziBond(context.Background(), nil, input)
	if err != nil {
		t.Fatalf("HandleBaziBond: %v", err)
	}
	if len(out.Bond.PillarCross.Pairs) != 16 {
		t.Errorf("pillar_cross pairs = %d, want 16", len(out.Bond.PillarCross.Pairs))
	}
	if out.ChartA.DayMaster == "" {
		t.Error("chart_a day_master is empty")
	}
	if out.ChartB.DayMaster == "" {
		t.Error("chart_b day_master is empty")
	}
}

func TestHandleBaziLiunian(t *testing.T) {
	input := BaziLiunianInput{
		Bazi: ganzhi.Bazi{
			Year:  ganzhi.Pillar{Stem: 8, Branch: 10},  // 辛酉
			Month: ganzhi.Pillar{Stem: 3, Branch: 9},   // 丙申
			Day:   ganzhi.Pillar{Stem: 2, Branch: 12},  // 乙亥
			Hour:  ganzhi.Pillar{Stem: 4, Branch: 2},   // 丁丑
		},
		Year: 2026,
	}
	_, out, err := HandleBaziLiunian(context.Background(), nil, input)
	if err != nil {
		t.Fatalf("HandleBaziLiunian: %v", err)
	}
	if out == nil {
		t.Fatal("liunian result is nil")
	}
	if out.Year != 2026 {
		t.Errorf("year = %d, want 2026", out.Year)
	}
	if out.TenGod == "" {
		t.Error("liunian ten_god is empty")
	}
}

func TestHandleQimingGenerate(t *testing.T) {
	input := QimingGenerateInput{
		Surname: "张", YongShen: "金", XiShen: []string{"土"}, Zodiac: 4, Limit: 5,
	}
	_, out, err := HandleQimingGenerate(context.Background(), nil, input)
	if err != nil {
		t.Fatalf("HandleQimingGenerate: %v", err)
	}
	if out.Surname != "张" {
		t.Errorf("surname = %s, want 张", out.Surname)
	}
	if len(out.Candidates) == 0 {
		t.Error("no candidates generated")
	}
	if len(out.Candidates) > 5 {
		t.Errorf("got %d candidates, want <= 5", len(out.Candidates))
	}
}

func TestHandleQimingEvaluate(t *testing.T) {
	input := QimingEvaluateInput{
		Surname: "张", GivenName: "伟", YongShen: "金", Zodiac: 4,
	}
	_, out, err := HandleQimingEvaluate(context.Background(), nil, input)
	if err != nil {
		t.Fatalf("HandleQimingEvaluate: %v", err)
	}
	if out.Surname != "张" {
		t.Errorf("surname = %s, want 张", out.Surname)
	}
	if out.GivenName == "" {
		t.Error("given_name is empty")
	}
	if out.WuGe.TianGe.Stroke == 0 {
		t.Error("wuge tian_ge stroke is zero")
	}
}

func TestHandleFengshuiMinggua(t *testing.T) {
	input := FengshuiMingguaInput{Year: 1984, Gender: "male"}
	_, out, err := HandleFengshuiMinggua(context.Background(), nil, input)
	if err != nil {
		t.Fatalf("HandleFengshuiMinggua: %v", err)
	}
	if out.GuaNumber < 1 || out.GuaNumber > 9 {
		t.Errorf("gua_number = %d, want 1-9", out.GuaNumber)
	}
	if out.Group != "东四命" && out.Group != "西四命" {
		t.Errorf("group = %s, want 东四命 or 西四命", out.Group)
	}
	if len(out.AllTrigrams) != 9 {
		t.Errorf("all_trigrams len = %d, want 9", len(out.AllTrigrams))
	}
}

func TestHandleHuangliQuery(t *testing.T) {
	input := HuangliQueryInput{Date: "2026-05-26"}
	_, out, err := HandleHuangliQuery(context.Background(), nil, input)
	if err != nil {
		t.Fatalf("HandleHuangliQuery: %v", err)
	}
	if len(out.Days) != 1 {
		t.Fatalf("expected 1 day, got %d", len(out.Days))
	}
	d := out.Days[0]
	if d.DayPillar.Stem == 0 || d.DayPillar.Branch == 0 {
		t.Error("day pillar missing")
	}
	if d.DayPillar.NaYin == "" {
		t.Error("nayin missing")
	}
	if d.JianChu == "" {
		t.Error("jian_chu missing")
	}
	if out.YearMonth != "2026-05" {
		t.Errorf("year_month = %s, want 2026-05", out.YearMonth)
	}
}

func TestHandleHuangliQuery_Month(t *testing.T) {
	input := HuangliQueryInput{Month: "2026-05"}
	_, out, err := HandleHuangliQuery(context.Background(), nil, input)
	if err != nil {
		t.Fatalf("HandleHuangliQuery month: %v", err)
	}
	if len(out.Days) == 0 {
		t.Error("no days returned for month query")
	}
	if out.YearMonth != "2026-05" {
		t.Errorf("year_month = %s, want 2026-05", out.YearMonth)
	}
}

func TestHandleHuangliQuery_BadDate(t *testing.T) {
	_, _, err := HandleHuangliQuery(context.Background(), nil, HuangliQueryInput{Date: "not-a-date"})
	if err == nil {
		t.Error("expected error for bad date")
	}
}

func TestHandleHuangliQuery_NoInput(t *testing.T) {
	_, _, err := HandleHuangliQuery(context.Background(), nil, HuangliQueryInput{})
	if err == nil {
		t.Error("expected error when neither date nor month provided")
	}
}

func TestHandleHuangliBond(t *testing.T) {
	birth := BirthProfile{Year: 1984, Month: 3, Day: 15, Hour: 8, Gender: "male"}
	input := HuangliBondInput{Birth: birth, Date: "2026-05-26"}
	_, out, err := HandleHuangliBond(context.Background(), nil, input)
	if err != nil {
		t.Fatalf("HandleHuangliBond: %v", err)
	}
	if len(out.Days) != 1 {
		t.Fatalf("expected 1 day, got %d", len(out.Days))
	}
	d := out.Days[0]
	if d.DayPillar.Stem == 0 || d.DayPillar.Branch == 0 {
		t.Error("day pillar missing")
	}
	if d.GanRelation == "" {
		t.Error("gan_relation missing")
	}
	if d.ZhiRelation == "" {
		t.Error("zhi_relation missing")
	}
	if d.TaiSuiRelation == "" {
		t.Error("tai_sui_relation missing")
	}
}

func TestHandleHuangliBond_NoInput(t *testing.T) {
	_, _, err := HandleHuangliBond(context.Background(), nil, HuangliBondInput{})
	if err == nil {
		t.Error("expected error when neither date nor month provided")
	}
}

func TestToolSchemaGeneration(t *testing.T) {
	// Verify registration with full structured types succeeds without panicking.
	// This confirms jsonschema tags produce valid schemas.
	s := newTestServer()
	RegisterTools(s)
	// Success = no panic during schema inference.
}

func TestHandleBaziChart_DefaultLongitude(t *testing.T) {
	// Verify defaults are applied when longitude/timezone are zero.
	input := BaziChartInput{
		Birth: BirthProfile{Year: 2000, Month: 6, Day: 1, Hour: 12, Gender: "female"},
	}
	_, out, err := HandleBaziChart(context.Background(), nil, input)
	if err != nil {
		t.Fatalf("HandleBaziChart with defaults: %v", err)
	}
	if out.SolarTimeMinutes == 0 && out.HourPillar.Stem == 0 {
		t.Error("solar time not computed — defaults may not have been applied")
	}
}

func TestHandleQimingGenerate_BadZodiac(t *testing.T) {
	// Out-of-range zodiac should not crash.
	input := QimingGenerateInput{
		Surname: "王", YongShen: "木", Zodiac: 99, Limit: 3,
	}
	_, out, err := HandleQimingGenerate(context.Background(), nil, input)
	if err != nil {
		t.Fatalf("HandleQimingGenerate with bad zodiac: %v", err)
	}
	if len(out.Candidates) == 0 {
		t.Error("expected candidates even with out-of-range zodiac")
	}
}


