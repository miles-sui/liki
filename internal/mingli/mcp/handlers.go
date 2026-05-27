package mcp

import (
	"context"
	"fmt"

	"github.com/25types/25types/internal/ganzhi"
	"github.com/25types/25types/internal/mingli/bazi"
	"github.com/25types/25types/internal/mingli/fengshui"
	"github.com/25types/25types/internal/mingli/huangli"
	"github.com/25types/25types/internal/mingli/qiming"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func computeChart(bp BirthProfile) bazi.ChartResult {
	return bazi.ComputeChartFromBirth(bp.Year, bp.Month, bp.Day, bp.Hour, bp.Minute, bp.Longitude, bp.Timezone, bazi.Gender(bp.Gender))
}

// ---- Tool handlers ----

// HandleBaziChart computes a full BaZi chart.
func HandleBaziChart(ctx context.Context, req *mcp.CallToolRequest, input BaziChartInput) (*mcp.CallToolResult, ChartOutput, error) {
	chart := computeChart(input.Birth)
	return nil, bazi.BuildChartOutput(chart, input.Birth.Year, input.Birth.Month, input.Birth.Hour), nil
}

// HandleBaziBond computes bazi bond (合盘) cross-chart analysis.
func HandleBaziBond(ctx context.Context, req *mcp.CallToolRequest, input BaziMatchInput) (*mcp.CallToolResult, BondOutput, error) {
	ca := computeChart(input.A)
	cb := computeChart(input.B)

	bond := bazi.ComputeBond(ca, cb, input.A.Year, input.A.Month, input.A.Hour, input.B.Year, input.B.Month, input.B.Hour)

	return nil, BondOutput{
		ChartA: bazi.BuildChartOutput(ca, input.A.Year, input.A.Month, input.A.Hour),
		ChartB: bazi.BuildChartOutput(cb, input.B.Year, input.B.Month, input.B.Hour),
		Bond:   bond,
	}, nil
}

// HandleBaziLiunian computes yearly fortune.
func HandleBaziLiunian(ctx context.Context, req *mcp.CallToolRequest, input BaziLiunianInput) (*mcp.CallToolResult, *bazi.LiunianResult, error) {
	var cd *bazi.DayunPillar
	if input.CurrentDayun != nil {
		cd = &bazi.DayunPillar{Stem: input.CurrentDayun.Stem, Branch: input.CurrentDayun.Branch}
	}

	result := bazi.ComputeLiunian(input.Year, input.Bazi.Day.Stem, input.Bazi, cd)
	return nil, result, nil
}

// HandleQimingGenerate generates name candidates.
func HandleQimingGenerate(ctx context.Context, req *mcp.CallToolRequest, input QimingGenerateInput) (*mcp.CallToolResult, QimingGenerateOutput, error) {
	if input.Limit <= 0 || input.Limit > 50 {
		input.Limit = 20
	}

	surnameElem := qiming.LookupSurnameElement(input.Surname)

	var zodiac qiming.ZodiacHint
	if input.Zodiac >= 1 && input.Zodiac <= 12 {
		zodiac = qiming.ZodiacFromYearBranch(ganzhi.Branch(input.Zodiac))
	}

	analysis := qiming.NamingAnalysis{
		Surname:    input.Surname,
		YongShen:   input.YongShen,
		XiShen:     input.XiShen,
		ZodiacHint: zodiac,
	}

	candidates := qiming.GenerateCandidates(input.Surname, analysis, input.Limit)
	if candidates == nil {
		candidates = []qiming.NameCandidate{}
	}

	return nil, QimingGenerateOutput{
		Surname:        input.Surname,
		SurnameElement: surnameElem.String(),
		YongShen:       input.YongShen,
		XiShen:         input.XiShen,
		ZodiacHint:     zodiac,
		Candidates:     candidates,
	}, nil
}

// HandleQimingEvaluate evaluates a name.
func HandleQimingEvaluate(ctx context.Context, req *mcp.CallToolRequest, input QimingEvaluateInput) (*mcp.CallToolResult, qiming.NameEvaluation, error) {
	result := qiming.EvaluateName(input.Surname, input.GivenName, input.YongShen, ganzhi.Branch(input.Zodiac))
	return nil, result, nil
}

// HandleFengshuiMinggua computes the fate trigram.
func HandleFengshuiMinggua(ctx context.Context, req *mcp.CallToolRequest, input FengshuiMingguaInput) (*mcp.CallToolResult, FengshuiMingguaOutput, error) {
	mg := fengshui.ComputeMingGua(ganzhi.Gender(input.Gender), input.Year)

	return nil, FengshuiMingguaOutput{
		Gua:         mg.Gua,
		GuaNumber:   mg.GuaNumber,
		Group:       mg.Group,
		AllTrigrams: fengshui.AllTrigrams(),
	}, nil
}

// HandleHuangliQuery looks up huangli (黄历) info.
func HandleHuangliQuery(ctx context.Context, req *mcp.CallToolRequest, input HuangliQueryInput) (*mcp.CallToolResult, HuangliQueryOutput, error) {
	if input.Date != "" {
		entry, err := huangli.QueryDate(input.Date, input.EventType)
		if err != nil {
			return nil, HuangliQueryOutput{}, err
		}
		return nil, HuangliQueryOutput{Days: []huangli.DayEntry{entry}, YearMonth: input.Date[:7]}, nil
	}
	if input.Month != "" {
		entries, err := huangli.QueryMonth(input.Month, input.EventType)
		if err != nil {
			return nil, HuangliQueryOutput{}, err
		}
		return nil, HuangliQueryOutput{Days: entries, YearMonth: input.Month}, nil
	}
	return nil, HuangliQueryOutput{}, fmt.Errorf("date or month is required")
}
// HandleFengshuiHeCan assembles combined Feng Shui reference data.
func HandleFengshuiHeCan(ctx context.Context, req *mcp.CallToolRequest, input FengshuiHeCanInput) (*mcp.CallToolResult, fengshui.HeCanResult, error) {
	result := fengshui.ComputeHeCan(
		input.BirthYear,
		ganzhi.Gender(input.Gender),
		input.Bazi,
		input.YongShen,
		input.Year,
	)
	return nil, result, nil
}



// HandleHuangliBond cross-references birth info against huangli days.
func HandleHuangliBond(ctx context.Context, req *mcp.CallToolRequest, input HuangliBondInput) (*mcp.CallToolResult, HuangliBondOutput, error) {
	if input.Date == "" && input.Month == "" {
		return nil, HuangliBondOutput{}, fmt.Errorf("month or date is required")
	}

	chart := computeChart(input.Birth)

	if input.Date != "" {
		entry, err := huangli.CrossDate(chart.DayMaster, chart.Day.Branch, input.Date, input.EventType)
		if err != nil {
			return nil, HuangliBondOutput{}, err
		}
		return nil, HuangliBondOutput{Days: []huangli.BondDayEntry{entry}, YearMonth: input.Date[:7]}, nil
	}

	entries, err := huangli.CrossMonth(chart.DayMaster, chart.Day.Branch, input.Month, input.EventType)
	if err != nil {
		return nil, HuangliBondOutput{}, err
	}
	return nil, HuangliBondOutput{Days: entries, YearMonth: input.Month}, nil
}
