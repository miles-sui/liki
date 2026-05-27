package flow

import (
	"context"
	"sync"
	"time"

	"github.com/25types/25types/internal/25types"
	"github.com/25types/25types/internal/app/domain"
	"github.com/25types/25types/internal/tianwen"
)

// ProfileLoader is an alias for the shared domain interface.
type ProfileLoader = domain.ProfileLoader

// Cache GetSolarTerms call since it recomputes expensive trig.
var (
	solarCacheOnce sync.Once
	cachedEntries  []persona.SolarTermEntry
	cachedYear     int
)

// GetFlow computes the current solar month's flow.
func GetFlow(ctx context.Context, loader ProfileLoader, userID int64) (*persona.FlowResult, error) {
	prof, err := loader.LoadProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := persona.ComputeFlow(prof.D, time.Now())
	return &result, nil
}

// GetFlowYearly computes the flow for all 12 solar months.
func GetFlowYearly(ctx context.Context, loader ProfileLoader, userID int64) ([]persona.FlowResult, *persona.FlowResult, error) {
	prof, err := loader.LoadProfile(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	months := persona.ComputeFlowYearly(prof.D)
	current := persona.ComputeFlow(prof.D, time.Now())
	return months, &current, nil
}

// GetSolarTerms returns the solar term calendar for the current year.
func GetSolarTerms() ([]persona.SolarTermEntry, string) {
	now := time.Now()
	year := now.Year()
	solarCacheOnce.Do(func() {
		cachedEntries = tianwen.PrecomputeSolarTerms(year)
		cachedYear = year
	})
	if cachedYear != year {
		cachedEntries = tianwen.PrecomputeSolarTerms(year)
		cachedYear = year
	}
	currentID := tianwen.GetCurrentSolarMonth(time.Now())
	return cachedEntries, currentID
}
