package minglilink

import (
	"context"
	"encoding/json"
	"log"

	"github.com/25types/25types/internal/app/application/matchlink"
	"github.com/25types/25types/internal/app/domain"
	"github.com/25types/25types/internal/mingli/bazi"
)

// GetMingliMatchLink returns BaZi match link info by token (public).
func GetMingliMatchLink(ctx context.Context, repo matchlink.MatchLinkRepository, users BirthInfoLookup, token string) (*MingliMatchLinkInfo, error) {
	ml, err := repo.FindByToken(ctx, token)
	if err != nil {
		return nil, domain.ErrMatchLinkNotFound
	}

	info := &MingliMatchLinkInfo{Token: token, Valid: true}

	creator, err := users.FindByID(ctx, ml.UserID)
	if err != nil || creator.BirthInfo == nil {
		return info, nil
	}

	info.CreatorName = creator.Name
	info.ChartA = chartSummary(creator.BirthInfo)
	return info, nil
}

func SubmitMingliMatch(
	ctx context.Context,
	linkRepo matchlink.MatchLinkRepository,
	users BirthInfoLookup,
	input SubmitMingliMatchInput,
) (*SubmitMingliMatchOutput, error) {
	ml, err := linkRepo.FindByToken(ctx, input.Token)
	if err != nil {
		return nil, domain.ErrMatchLinkNotFound
	}

	creator, err := users.FindByID(ctx, ml.UserID)
	if err != nil || creator.BirthInfo == nil {
		return nil, errNoCreatorBirthInfo
	}
	chartA := computeChart(creator.BirthInfo)

	var (
		chartB        bazi.ChartResult
		otherID       *int64
		otherName     string
		bYear, bMonth, bHour int
	)

	if input.UseExisting && input.UserID != nil {
		u, err := users.FindByID(ctx, *input.UserID)
		if err != nil || u.BirthInfo == nil {
			return nil, errNoBirthInfo
		}
		chartB = computeChart(u.BirthInfo)
		otherID = input.UserID
		otherName = u.Name
		bYear, bMonth, bHour = u.BirthInfo.Year, u.BirthInfo.Month, u.BirthInfo.Hour
	} else if input.BirthInfo != nil {
		bi := input.BirthInfo
		chartB = computeChart(bi)
		otherName = input.OtherName
		bYear, bMonth, bHour = bi.Year, bi.Month, bi.Hour
	} else {
		return nil, domain.ErrAnswersRequired
	}

	bond := bazi.ComputeBond(chartA, chartB, creator.BirthInfo.Year, creator.BirthInfo.Month, creator.BirthInfo.Hour, bYear, bMonth, bHour)

	chartASummary := chartSummaryFromEngine(chartA)
	chartAJSON, _ := json.Marshal(chartASummary)
	chartBJSON, _ := json.Marshal(chartSummaryFromEngine(chartB))
	matchJSON, _ := json.Marshal(bond)

	if linkRepo != nil {
		if err := linkRepo.InsertMingliMatchEvent(ctx, matchlink.InsertMingliMatchEventParams{
			LinkID:          &ml.ID,
			InitiatorUserID: ml.UserID,
			OtherUserID:     otherID,
			OtherName:       otherName,
			ChartAJSON:      string(chartAJSON),
			ChartBJSON:      string(chartBJSON),
			MatchJSON:       string(matchJSON),
		}); err != nil {
			log.Printf("[minglilink] failed to store match event: %v", err)
		}
	}

	return &SubmitMingliMatchOutput{
		ChartA:        chartASummary,
		ChartB:        chartSummaryFromEngine(chartB),
		Bond:          bond,
		CreatorUserID: ml.UserID,
	}, nil
}

// -- helpers --

type errMsg string

func (e errMsg) Error() string { return string(e) }

var (
	errNoCreatorBirthInfo = errMsg("link creator has no birth info saved")
	errNoBirthInfo        = errMsg("no birth info saved — save your chart first")
)

func computeChart(bi *domain.BirthInfo) bazi.ChartResult {
	lon := bi.Longitude
	if lon == 0 {
		lon = 120.0
	}
	tz := bi.Timezone
	if tz == 0 {
		tz = 8.0
	}
		ast := bazi.ComputeSolarTime(bi.Year, bi.Month, bi.Day, bi.Hour, bi.Minute, lon, tz, bi.IsDST)
		bz := bazi.ComputeBazi(ast, bi.Year, bi.Month, bi.Day, bi.Hour, bi.Minute, tz, bi.IsDST)
		return bazi.ComputeChart(bz, bi.Year, bi.Month, bi.Day, bazi.Gender(bi.Gender))
}

func chartSummary(bi *domain.BirthInfo) map[string]interface{} {
	return chartSummaryFromEngine(computeChart(bi))
}

func chartSummaryFromEngine(c bazi.ChartResult) map[string]interface{} {
	ec := make(map[int]int, 5)
	for e, c := range c.ElementCount {
		ec[int(e)] = c
	}
	return map[string]interface{}{
		"year_pillar":   c.Year,
		"month_pillar":  c.Month,
		"day_pillar":    c.Day,
		"hour_pillar":   c.Hour,
		"day_master":    int(c.DayMaster),
		"element_count": ec,
	}
}
