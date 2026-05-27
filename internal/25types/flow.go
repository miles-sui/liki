package persona

import (
	"time"

	"github.com/25types/25types/internal/ganzhi"
	"github.com/25types/25types/internal/tianwen"
)

type FlowResult struct {
	MonthID   string `json:"month_id"`
	MonthEN   string `json:"month_en"`
	Generates int    `json:"generates"`
	Restrains int    `json:"restrains"`
}

func shengKeForMonth(monthID string) (generates, restrains int) {
	sMonth := Deviation(ganzhi.SolarMonthTable[monthID])
	gen := ApplyTranspose(&ganzhi.S, sMonth)
	res := ApplyTranspose(&ganzhi.C, sMonth)
	for i := 0; i < 5; i++ {
		if gen[i] > 0 {
			generates = i
		}
		if res[i] > 0 {
			restrains = i
		}
	}
	return
}

func ComputeFlow(_ Deviation, t time.Time) FlowResult {
	monthID := GetCurrentSolarMonth(t)
	g, r := shengKeForMonth(monthID)
	return FlowResult{MonthID: monthID, MonthEN: ganzhi.MonthNamesEN[monthID], Generates: g, Restrains: r}
}

func ComputeFlowYearly(_ Deviation) []FlowResult {
	months := make([]FlowResult, 0, 12)
	for _, id := range ganzhi.SolarMonthOrder {
		g, r := shengKeForMonth(id)
		months = append(months, FlowResult{MonthID: id, MonthEN: ganzhi.MonthNamesEN[id], Generates: g, Restrains: r})
	}
	return months
}

func GetCurrentSolarMonth(t time.Time) string {
	return tianwen.GetCurrentSolarMonth(t)
}

type SolarTermEntry = tianwen.SolarTermEntry
