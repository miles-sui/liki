package profile

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/25types/25types/internal/25types"
	"github.com/25types/25types/internal/app/domain"
	"github.com/25types/25types/internal/ganzhi"
	"github.com/25types/25types/internal/mingli/bazi"
)

// ProfileOutput is the aggregated profile response for GET /api/profiles/{name}.
type ProfileOutput struct {
	User          ProfileUser       `json:"user"`
	Profile       *ProfileData      `json:"profile,omitempty"`
	FlowMonth     *FlowMonthData    `json:"flow_month,omitempty"`
	Peers         *PeersData        `json:"peers,omitempty"`
	IsOwner       bool              `json:"is_owner"`
	IsPublic      bool              `json:"is_public"`
	HasReviewLink *string           `json:"has_review_link,omitempty"`
	BirthInfo     *domain.BirthInfo `json:"birth_info,omitempty"`
	MingliChart   *MingliChartSummary `json:"mingli_chart,omitempty"`
}

// MingliChartSummary is the profile-level BaZi chart view.
type MingliChartSummary struct {
	DayMaster    int                  `json:"day_master"`
	YearPillar   ganzhi.Pillar           `json:"year_pillar"`
	MonthPillar  ganzhi.Pillar           `json:"month_pillar"`
	DayPillar    ganzhi.Pillar           `json:"day_pillar"`
	HourPillar   ganzhi.Pillar           `json:"hour_pillar"`
	ElementCount map[int]int             `json:"element_count"`
	NaYin        [4]string               `json:"na_yin"`
	HiddenStems  [4]bazi.HiddenStemsOut `json:"hidden_stems"`
	TenGods      [4][2]string         `json:"ten_gods"`
	LifeStages   []LifeStageItem      `json:"life_stages"`
	SolarTime    float64              `json:"solar_time"`
	StemBranch   int                  `json:"stem_branch"`
	Dayun        MingliDayun       `json:"dayun"`
}

// LifeStageItem is one node of the 12 life stages for profile output.
type LifeStageItem struct {
	Name   string        `json:"name"`
	Branch ganzhi.Branch `json:"branch"`
}

// MingliDayun is the dayun (大运) for profile output.
type MingliDayun struct {
	StartAge  int             `json:"start_age"`
	Direction string          `json:"direction"`
	Pillars   []ganzhi.Pillar `json:"pillars"`
}

// ProfileUser is the public user info embedded in a profile response.
type ProfileUser struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	IsPublic bool   `json:"is_public"`
}

// ProfileData holds the personality profile display data.
type ProfileData struct {
	domain.PersonalityProfile
}

// FlowMonthData holds the current solar month flow.
type FlowMonthData struct {
	MonthID   string `json:"month_id"`
	MonthEN   string `json:"month_en"`
	Generates int    `json:"generates"`
	Restrains int    `json:"restrains"`
}

// PeersData holds aggregated peer review data.
type PeersData struct {
	domain.PersonalityProfile
	Count int `json:"count"`
}

// GetProfile aggregates public profile data for a user by name.
func GetProfile(ctx context.Context, repo ProfilePageRepo, name string, viewerID *int64) (*ProfileOutput, error) {
	u, err := repo.FindByName(ctx, name)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	isOwner := viewerID != nil && *viewerID == u.ID

	if !u.IsPublic && !isOwner {
		return nil, domain.ErrUserNotFound
	}

	out := &ProfileOutput{
		User:     ProfileUser{ID: u.ID, Name: u.Name, IsPublic: u.IsPublic},
		IsOwner:  isOwner,
		IsPublic: u.IsPublic,
	}

	if token, ok := repo.FindActiveReviewLink(ctx, u.ID); ok {
		out.HasReviewLink = &token
	}

	// Birth info: only include for owner or public profiles.
	if u.BirthInfo != nil && (isOwner || u.IsPublic) {
		out.BirthInfo = u.BirthInfo
		ast := bazi.ComputeSolarTime(
			u.BirthInfo.Year, u.BirthInfo.Month, u.BirthInfo.Day,
			u.BirthInfo.Hour, u.BirthInfo.Minute,
			u.BirthInfo.Longitude, u.BirthInfo.Timezone,
			u.BirthInfo.IsDST,
		)
		bz := bazi.ComputeBazi(ast, u.BirthInfo.Year, u.BirthInfo.Month, u.BirthInfo.Day, u.BirthInfo.Hour, u.BirthInfo.Minute, u.BirthInfo.Timezone, u.BirthInfo.IsDST)
		chart := bazi.ComputeChart(bz, u.BirthInfo.Year, u.BirthInfo.Month, u.BirthInfo.Day, bazi.Gender(u.BirthInfo.Gender))
		ec := make(map[int]int, 5)
		for e, c := range chart.ElementCount {
			ec[int(e)] = c
		}
		// Convert life stages
		ls := make([]LifeStageItem, 0, 12)
		for _, s := range chart.LifeStages.Slice() {
			ls = append(ls, LifeStageItem{Name: s.Name, Branch: s.Branch})
		}
		out.MingliChart = &MingliChartSummary{
			DayMaster:    int(chart.DayMaster),
			YearPillar:   ganzhi.Pillar{Stem: chart.Year.Stem, Branch: chart.Year.Branch},
			MonthPillar:  ganzhi.Pillar{Stem: chart.Month.Stem, Branch: chart.Month.Branch},
			DayPillar:    ganzhi.Pillar{Stem: chart.Day.Stem, Branch: chart.Day.Branch},
			HourPillar:   ganzhi.Pillar{Stem: chart.Hour.Stem, Branch: chart.Hour.Branch},
			ElementCount: ec,
			NaYin:        chart.NaYinArray(),
			HiddenStems: chart.HiddenStemsArray(),
			TenGods:      chart.TenGodsArray(),
			LifeStages:   ls,
			SolarTime:    chart.SolarTime,
			StemBranch:   ganzhi.SixtyCycleName(chart.Year.Stem, chart.Year.Branch),
			Dayun: MingliDayun{
				StartAge:  chart.Dayun.StartAge,
				Direction: chart.Dayun.Direction,
				Pillars:   chart.Dayun.Pillars,
			},
		}
	}

	if prof, err := repo.LoadProfile(ctx, u.ID); err == nil {
		out.Profile = &ProfileData{
			PersonalityProfile: *prof,
		}

		result := persona.ComputeFlow(prof.D, time.Now())
		out.FlowMonth = &FlowMonthData{
			MonthID: result.MonthID,
			MonthEN: result.MonthEN,
				Generates: result.Generates,
			Restrains: result.Restrains,
		}
	}

	if peerAnswers, peerCount, _ := repo.ListPeerAnswersForUser(ctx, u.ID); peerCount > 0 {
		peerProfile := domain.ComputeProfileFromAnswers(peerAnswers)
		out.Peers = &PeersData{
			PersonalityProfile: peerProfile,
			Count:              peerCount,
		}
	}

	return out, nil
}

// BondEventItem is the API response item for GET /api/profiles/{name}/bonds.
type BondEventItem struct {
	ID          int64        `json:"id"`
	OtherUserID int64        `json:"other_user_id"`
	OtherName   string       `json:"other_name"`
	Bond        *domain.Bond `json:"bond,omitempty"`
	Source      string       `json:"source"`
	CreatedAt   string       `json:"created_at"`
}

// GetBonds returns all bond events for a user, with bond data deserialized and
// normalized so the viewer is always in the "self" position.
func GetBonds(ctx context.Context, store BondStore, userID int64) ([]BondEventItem, error) {
	events, err := store.ListBondEvents(ctx, userID)
	if err != nil {
		return nil, err
	}

	items := make([]BondEventItem, 0, len(events))
	for _, e := range events {
		var otherUserID int64
		if e.OtherUserID != nil {
			otherUserID = *e.OtherUserID
		}
		otherName := e.OtherName
		isViewer := e.InitiatorUserID == userID

		item := BondEventItem{
			ID:          e.ID,
			OtherUserID: otherUserID,
			OtherName:   otherName,
			CreatedAt:   e.CreatedAt.Format(time.RFC3339),
		}
		if e.LinkID != nil {
			item.Source = "match_link"
		} else {
			item.Source = "instant"
		}
		if e.BondJSON != "" {
			var b domain.Bond
			if err := json.Unmarshal([]byte(e.BondJSON), &b); err == nil {
				// Normalize: if viewer is the "other" in the stored bond,
				// swap so viewer is always self.
				if !isViewer && e.OtherUserID != nil && *e.OtherUserID == userID {
					b.Self, b.Other = b.Other, b.Self
					b.DeltaA, b.DeltaB = b.DeltaB, b.DeltaA
				}
				item.Bond = &b
			}
		}
		// Normalize OtherUserID: always point to the other person (not viewer).
		if !isViewer && e.OtherUserID != nil && *e.OtherUserID == userID {
			item.OtherUserID = e.InitiatorUserID
		}
		items = append(items, item)
	}
	return items, nil
}

// ComputeBondOutput is the API response for POST /api/bond.
type ComputeBondOutput struct {
	Self      persona.Deviation `json:"self"`
	Other     persona.Deviation `json:"other"`
	DeltaA    persona.Deviation `json:"delta_a"`
	DeltaB    persona.Deviation `json:"delta_b"`
	Concord  persona.Concord  `json:"concord"`
	OtherUser *BondOtherUser   `json:"other_user,omitempty"`
}

type BondOtherUser struct {
	Name          string `json:"name"`
	IdentityLabel string `json:"identity_label,omitempty"`
	IdentityID    string `json:"identity_id,omitempty"`
}

type ComputeAndStoreBondResult struct {
	Bond  *domain.Bond
	ProfA *domain.PersonalityProfile
	ProfB *domain.PersonalityProfile
}

// ComputeAndStoreBond computes bond between two users and stores the event.
func ComputeAndStoreBond(ctx context.Context, store BondStore, loader domain.ProfileLoader, initiatorID, otherID int64) (*ComputeAndStoreBondResult, error) {
	return ComputeAndStoreBondWithOpts(ctx, store, loader, InsertBondParams{
		InitiatorID: initiatorID,
		OtherID:     otherID,
	})
}

// ComputeAndStoreBondWithOpts computes bond between two users and stores the event,
// accepting optional linkID and assessmentID for match-link context.
func ComputeAndStoreBondWithOpts(ctx context.Context, store BondStore, loader domain.ProfileLoader, params InsertBondParams) (*ComputeAndStoreBondResult, error) {
	profA, err := loader.LoadProfile(ctx, params.InitiatorID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrNoProfile
		}
		return nil, err
	}
	profB, err := loader.LoadProfile(ctx, params.OtherID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrNoProfile
		}
		return nil, err
	}

	bond := domain.NewBond(*profA, *profB)

	params.Bond = &bond
	if err := store.InsertBondEvent(ctx, params); err != nil {
		return nil, err
	}

	return &ComputeAndStoreBondResult{Bond: &bond, ProfA: profA, ProfB: profB}, nil
}
