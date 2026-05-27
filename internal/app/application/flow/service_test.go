package flow

import (
	"context"
	"testing"

	"github.com/25types/25types/internal/app/domain"
	persona "github.com/25types/25types/internal/25types"
)

// stubProfileLoader is a test double implementing domain.ProfileLoader.
type stubProfileLoader struct {
	profiles map[int64]*domain.PersonalityProfile
}

func (l *stubProfileLoader) LoadProfile(ctx context.Context, userID int64) (*domain.PersonalityProfile, error) {
	p, ok := l.profiles[userID]
	if !ok {
		return nil, domain.ErrNoProfile
	}
	return p, nil
}

// full30Answers returns exactly 30 selections across 10 questions (3 picks each).
func full30Answers() []persona.Answer {
	return []persona.Answer{
		{QID: "Q01", Selections: []string{"W", "F", "E"}},
		{QID: "Q02", Selections: []string{"W", "M", "R"}},
		{QID: "Q03", Selections: []string{"F", "E", "M"}},
		{QID: "Q04", Selections: []string{"W", "F", "R"}},
		{QID: "Q05", Selections: []string{"E", "M", "R"}},
		{QID: "Q06", Selections: []string{"W", "F", "E"}},
		{QID: "Q07", Selections: []string{"W", "E", "R"}},
		{QID: "Q08", Selections: []string{"F", "E", "M"}},
		{QID: "Q09", Selections: []string{"W", "E", "R"}},
		{QID: "Q10", Selections: []string{"F", "M", "R"}},
	}
}

func TestGetFlow(t *testing.T) {
	ctx := context.Background()
	prof := domain.ComputeProfileFromAnswers(full30Answers())

	t.Run("OK", func(t *testing.T) {
		loader := &stubProfileLoader{
			profiles: map[int64]*domain.PersonalityProfile{1: &prof},
		}
		result, err := GetFlow(ctx, loader, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.MonthID == "" {
			t.Error("expected non-empty MonthID")
		}
		if result.MonthEN == "" {
			t.Error("expected non-empty MonthEN")
		}
		if result.Generates < 0 || result.Generates > 4 {
			t.Errorf("generates = %d, want 0-4", result.Generates)
		}
		if result.Restrains < 0 || result.Restrains > 4 {
			t.Errorf("restrains = %d, want 0-4", result.Restrains)
		}
	})

	t.Run("NoProfile", func(t *testing.T) {
		loader := &stubProfileLoader{
			profiles: map[int64]*domain.PersonalityProfile{},
		}
		_, err := GetFlow(ctx, loader, 42)
		if err != domain.ErrNoProfile {
			t.Errorf("expected ErrNoProfile, got %v", err)
		}
	})
}

func TestGetFlowYearly(t *testing.T) {
	ctx := context.Background()
	prof := domain.ComputeProfileFromAnswers(full30Answers())

	t.Run("OK", func(t *testing.T) {
		loader := &stubProfileLoader{
			profiles: map[int64]*domain.PersonalityProfile{1: &prof},
		}
		months, current, err := GetFlowYearly(ctx, loader, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(months) != 12 {
			t.Errorf("expected 12 months, got %d", len(months))
		}
		if current == nil {
			t.Error("expected non-nil current month")
		}
		if current.MonthID == "" {
			t.Error("expected non-empty current MonthID")
		}
		seen := make(map[string]bool)
		for _, m := range months {
			if seen[m.MonthID] {
				t.Errorf("duplicate month %s", m.MonthID)
			}
			seen[m.MonthID] = true
		}
	})

	t.Run("NoProfile", func(t *testing.T) {
		loader := &stubProfileLoader{
			profiles: map[int64]*domain.PersonalityProfile{},
		}
		_, _, err := GetFlowYearly(ctx, loader, 42)
		if err != domain.ErrNoProfile {
			t.Errorf("expected ErrNoProfile, got %v", err)
		}
	})
}

func TestGetSolarTerms(t *testing.T) {
	entries, currentID := GetSolarTerms()
	if len(entries) != 12 {
		t.Errorf("expected 12 solar terms, got %d", len(entries))
	}
	if currentID == "" {
		t.Error("expected non-empty current month ID")
	}
	for _, e := range entries {
		if e.MonthID == "" {
			t.Error("entry has empty MonthID")
		}
		if e.NameEN == "" {
			t.Error("entry has empty NameEN")
		}
		if e.Date.IsZero() {
			t.Error("entry has zero Date")
		}
	}
	// Verify caching — second call uses the same cache (same year).
	entries2, _ := GetSolarTerms()
	if len(entries2) != 12 {
		t.Error("cached call returned wrong count")
	}
}
