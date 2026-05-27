package profile

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/25types/25types/internal/app/application/testutil"
	"github.com/25types/25types/internal/app/domain"
	persona "github.com/25types/25types/internal/25types"
)

// =============================================================================
// Stubs
// =============================================================================

type stubBondStore struct {
	events   []domain.BondEvent
	nextID   int64
	inserted []domain.Bond
}

func (s *stubBondStore) InsertBondEvent(ctx context.Context, params InsertBondParams) error {
	b, _ := json.Marshal(params.Bond)
	s.nextID++
	e := domain.BondEvent{
		ID:              s.nextID,
		LinkID:          params.LinkID,
		InitiatorUserID: params.InitiatorID,
		OtherUserID:     &params.OtherID,
		BondJSON:        string(b),
	}
	s.events = append(s.events, e)
	if params.Bond != nil {
		s.inserted = append(s.inserted, *params.Bond)
	}
	return nil
}

func (s *stubBondStore) ListBondEvents(ctx context.Context, userID int64) ([]domain.BondEvent, error) {
	var result []domain.BondEvent
	for _, e := range s.events {
		if e.InitiatorUserID == userID || (e.OtherUserID != nil && *e.OtherUserID == userID) {
			result = append(result, e)
		}
	}
	return result, nil
}

func makeTestProfile() domain.PersonalityProfile {
	d := persona.Deviation{0.3, -0.1, 0.2, -0.2, -0.2}
	p := persona.Proportion{0.35, 0.1, 0.3, 0.1, 0.15}
	return domain.NewProfile(d, p, persona.Identity{ID: "WF", Label: "WF"})
}

func ref[T any](v T) *T { return &v }

// =============================================================================
// GetBonds
// =============================================================================

func TestGetBonds_PerspectiveInitiator(t *testing.T) {
	store := &stubBondStore{}
	ctx := context.Background()

	bond := domain.Bond{
		Self:   persona.Deviation{0.3, 0.0, 0.1, -0.1, -0.3},
		Other:  persona.Deviation{0.1, 0.2, 0.0, -0.1, -0.2},
		DeltaA: persona.Deviation{0.05, -0.1, 0.0, 0.05, 0.0},
		DeltaB: persona.Deviation{-0.05, 0.05, 0.0, -0.05, 0.05},
	}
	store.InsertBondEvent(ctx, InsertBondParams{LinkID: nil, InitiatorID: 1, OtherID: 2, AssessmentID: nil, Bond: &bond})

	items, err := GetBonds(ctx, store, 1)
	if err != nil {
		t.Fatalf("GetBonds: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Bond.Self != bond.Self {
		t.Errorf("initiator should see self=original Self")
	}
	if items[0].Bond.Other != bond.Other {
		t.Errorf("initiator should see other=original Other")
	}
}

func TestGetBonds_PerspectiveSwap(t *testing.T) {
	store := &stubBondStore{}
	ctx := context.Background()

	bond := domain.Bond{
		Self:   persona.Deviation{0.3, 0.0, 0.1, -0.1, -0.3},
		Other:  persona.Deviation{0.1, 0.2, 0.0, -0.1, -0.2},
		DeltaA: persona.Deviation{0.05, -0.1, 0.0, 0.05, 0.0},
		DeltaB: persona.Deviation{-0.05, 0.05, 0.0, -0.05, 0.05},
	}
	store.InsertBondEvent(ctx, InsertBondParams{LinkID: nil, InitiatorID: 1, OtherID: 2, AssessmentID: nil, Bond: &bond})

	items, err := GetBonds(ctx, store, 2)
	if err != nil {
		t.Fatalf("GetBonds: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Bond.Self != bond.Other {
		t.Errorf("other user should see self=original Other after swap")
	}
	if items[0].Bond.Other != bond.Self {
		t.Errorf("other user should see other=original Self after swap")
	}
	if items[0].OtherUserID != 1 {
		t.Errorf("OtherUserID should be 1 (initiator), got %d", items[0].OtherUserID)
	}
}

func TestGetBonds_EmptyList(t *testing.T) {
	store := &stubBondStore{}
	items, err := GetBonds(context.Background(), store, 1)
	if err != nil {
		t.Fatalf("GetBonds: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}

func TestGetBonds_SourceMatchLink(t *testing.T) {
	store := &stubBondStore{}
	ctx := context.Background()

	bond := domain.Bond{Self: persona.Deviation{0.1}, Other: persona.Deviation{0.2}}
	var linkID int64 = 42
	store.InsertBondEvent(ctx, InsertBondParams{LinkID: &linkID, InitiatorID: 1, OtherID: 2, AssessmentID: nil, Bond: &bond})

	items, _ := GetBonds(ctx, store, 1)
	if items[0].Source != "match_link" {
		t.Errorf("source should be match_link when linkID present, got %q", items[0].Source)
	}
}

func TestGetBonds_SourceInstant(t *testing.T) {
	store := &stubBondStore{}
	ctx := context.Background()

	bond := domain.Bond{Self: persona.Deviation{0.1}, Other: persona.Deviation{0.2}}
	store.InsertBondEvent(ctx, InsertBondParams{LinkID: nil, InitiatorID: 1, OtherID: 2, AssessmentID: nil, Bond: &bond})

	items, _ := GetBonds(ctx, store, 1)
	if items[0].Source != "instant" {
		t.Errorf("source should be instant when linkID is nil, got %q", items[0].Source)
	}
}

// =============================================================================
// ComputeAndStoreBond
// =============================================================================

func TestComputeAndStoreBond_Normal(t *testing.T) {
	loader := &testutil.StubProfileLoader{Profiles: map[int64]*domain.PersonalityProfile{
		1: ref(makeTestProfile()),
		2: ref(makeTestProfile()),
	}}
	store := &stubBondStore{}

	result, err := ComputeAndStoreBond(context.Background(), store, loader, 1, 2)
	if err != nil {
		t.Fatalf("ComputeAndStoreBond: %v", err)
	}
	if result.Bond == nil {
		t.Fatal("bond should not be nil")
	}
	if len(store.inserted) != 1 {
		t.Fatalf("expected 1 InsertBondEvent call, got %d", len(store.inserted))
	}
}

func TestComputeAndStoreBond_NoProfile(t *testing.T) {
	loader := &testutil.StubProfileLoader{Profiles: map[int64]*domain.PersonalityProfile{
		1: ref(makeTestProfile()),
	}}
	store := &stubBondStore{}

	_, err := ComputeAndStoreBond(context.Background(), store, loader, 1, 2)
	if err == nil {
		t.Fatal("expected error when one user has no profile")
	}
}

func TestComputeAndStoreBond_MultipleCalls(t *testing.T) {
	loader := &testutil.StubProfileLoader{Profiles: map[int64]*domain.PersonalityProfile{
		1: ref(makeTestProfile()),
		2: ref(makeTestProfile()),
	}}
	store := &stubBondStore{}

	ComputeAndStoreBond(context.Background(), store, loader, 1, 2)
	ComputeAndStoreBond(context.Background(), store, loader, 1, 2)

	if len(store.inserted) != 2 {
		t.Errorf("expected 2 InsertBondEvent calls, got %d", len(store.inserted))
	}
}
