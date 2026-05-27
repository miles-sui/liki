package sqlite

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/25types/25types/internal/app/application/profile"
	"github.com/25types/25types/internal/app/domain"
	persona "github.com/25types/25types/internal/25types"
)

func newTestProfileRepo(t *testing.T) *ProfileRepo {
	t.Helper()
	userRepo := newTestUserRepo(t)
	assRepo := newTestAssessRepo(t)
	return NewProfileRepo(userRepo, assRepo)
}

// createTestUserWithProfile creates a user with a self-assessment so LoadProfile succeeds.
func createTestUserWithProfile(t *testing.T, repo *ProfileRepo, name string) int64 {
	t.Helper()
	uid := createTestUser(t, repo.UserRepo, name)

	d := persona.Deviation{0.2, -0.1, 0.3, -0.2, -0.2}
	p := persona.Proportion{0.3, 0.15, 0.3, 0.1, 0.15}
	identity := persona.Identity{ID: "WF", Label: "WF"}
	prof := domain.NewProfile(d, p, identity)
	answers := []persona.Answer{{QID: "q1", Selections: []string{"W", "F"}}}

	_, err := repo.AssessmentRepo.CreateSelf(context.Background(), uid, prof, answers)
	if err != nil {
		t.Fatalf("CreateSelf for user %d: %v", uid, err)
	}
	return uid
}

func makeTestBond() *domain.Bond {
	d := persona.Deviation{0.3, 0.0, 0.1, -0.1, -0.3}
	return &domain.Bond{
		Self:   d,
		Other:  d,
		DeltaA: persona.Deviation{0.05, -0.1, 0.0, 0.05, 0.0},
		DeltaB: persona.Deviation{-0.05, 0.05, 0.0, -0.05, 0.05},
	}
}

// =============================================================================
// InsertBondEvent
// =============================================================================

func TestInsertBondEvent_Normal(t *testing.T) {
	repo := newTestProfileRepo(t)
	ctx := context.Background()

	aID := createTestUserWithProfile(t, repo, "bond-a")
	bID := createTestUserWithProfile(t, repo, "bond-b")

	bond := makeTestBond()
	err := repo.InsertBondEvent(ctx, profile.InsertBondParams{LinkID: nil, InitiatorID: aID, OtherID: bID, AssessmentID: nil, Bond: bond})
	if err != nil {
		t.Fatalf("InsertBondEvent: %v", err)
	}

	events, err := repo.ListBondEvents(ctx, aID)
	if err != nil {
		t.Fatalf("ListBondEvents: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	e := events[0]
	if e.InitiatorUserID != aID {
		t.Errorf("InitiatorUserID = %d, want %d", e.InitiatorUserID, aID)
	}
	if e.OtherUserID == nil || *e.OtherUserID != bID {
		t.Errorf("OtherUserID = %v, want %d", e.OtherUserID, bID)
	}
	if e.LinkID != nil {
		t.Error("LinkID should be nil")
	}
}

func TestInsertBondEvent_BondJSONStructure(t *testing.T) {
	repo := newTestProfileRepo(t)
	ctx := context.Background()

	aID := createTestUserWithProfile(t, repo, "bond-json-a")
	bID := createTestUserWithProfile(t, repo, "bond-json-b")

	bond := makeTestBond()
	err := repo.InsertBondEvent(ctx, profile.InsertBondParams{LinkID: nil, InitiatorID: aID, OtherID: bID, AssessmentID: nil, Bond: bond})
	if err != nil {
		t.Fatalf("InsertBondEvent: %v", err)
	}

	events, _ := repo.ListBondEvents(ctx, aID)
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	var parsed domain.Bond
	if err := json.Unmarshal([]byte(events[0].BondJSON), &parsed); err != nil {
		t.Fatalf("bond_json is not valid JSON: %v", err)
	}
	if parsed.Self != bond.Self {
		t.Error("bond_json self mismatch")
	}
	if parsed.Other != bond.Other {
		t.Error("bond_json other mismatch")
	}
	if parsed.DeltaA != bond.DeltaA {
		t.Error("bond_json delta_a mismatch")
	}
	if parsed.DeltaB != bond.DeltaB {
		t.Error("bond_json delta_b mismatch")
	}
}

func TestInsertBondEvent_Dedup(t *testing.T) {
	repo := newTestProfileRepo(t)
	ctx := context.Background()

	aID := createTestUserWithProfile(t, repo, "bond-dedup-a")
	bID := createTestUserWithProfile(t, repo, "bond-dedup-b")

	// Insert first bond.
	bond1 := makeTestBond()
	if err := repo.InsertBondEvent(ctx, profile.InsertBondParams{LinkID: nil, InitiatorID: aID, OtherID: bID, AssessmentID: nil, Bond: bond1}); err != nil {
		t.Fatalf("first InsertBondEvent: %v", err)
	}

	// Insert second bond (same pair).
	bond2 := makeTestBond()
	bond2.DeltaA = persona.Deviation{0.1, -0.2, 0.0, 0.1, 0.0}
	if err := repo.InsertBondEvent(ctx, profile.InsertBondParams{LinkID: nil, InitiatorID: aID, OtherID: bID, AssessmentID: nil, Bond: bond2}); err != nil {
		t.Fatalf("second InsertBondEvent: %v", err)
	}

	events, _ := repo.ListBondEvents(ctx, aID)
	if len(events) != 1 {
		t.Fatalf("expected 1 event after dedup, got %d", len(events))
	}

	// Verify it's the second bond (latest).
	var parsed domain.Bond
	json.Unmarshal([]byte(events[0].BondJSON), &parsed)
	if parsed.DeltaA != bond2.DeltaA {
		t.Error("kept bond should be the latest one (bond2)")
	}
}

func TestInsertBondEvent_Anonymous(t *testing.T) {
	repo := newTestProfileRepo(t)
	ctx := context.Background()

	aID := createTestUserWithProfile(t, repo, "bond-anon-a")

	bond := makeTestBond()
	// otherID = 0 (anonymous).
	if err := repo.InsertBondEvent(ctx, profile.InsertBondParams{LinkID: nil, InitiatorID: aID, OtherID: 0, AssessmentID: nil, Bond: bond}); err != nil {
		t.Fatalf("InsertBondEvent with otherID=0: %v", err)
	}

	// Insert again with same initiator and otherID=0 — should NOT be deduped.
	bond2 := makeTestBond()
	if err := repo.InsertBondEvent(ctx, profile.InsertBondParams{LinkID: nil, InitiatorID: aID, OtherID: 0, AssessmentID: nil, Bond: bond2}); err != nil {
		t.Fatalf("second InsertBondEvent with otherID=0: %v", err)
	}

	events, _ := repo.ListBondEvents(ctx, aID)
	if len(events) != 2 {
		t.Fatalf("expected 2 events (anonymous not deduped), got %d", len(events))
	}
}

// =============================================================================
// ListBondEvents
// =============================================================================

func TestListBondEvents_BothDirections(t *testing.T) {
	repo := newTestProfileRepo(t)
	ctx := context.Background()

	aID := createTestUserWithProfile(t, repo, "bond-dir-a")
	bID := createTestUserWithProfile(t, repo, "bond-dir-b")

	bond := makeTestBond()
	if err := repo.InsertBondEvent(ctx, profile.InsertBondParams{LinkID: nil, InitiatorID: aID, OtherID: bID, AssessmentID: nil, Bond: bond}); err != nil {
		t.Fatalf("InsertBondEvent: %v", err)
	}

	// Both users should see the bond.
	for _, uid := range []int64{aID, bID} {
		events, err := repo.ListBondEvents(ctx, uid)
		if err != nil {
			t.Fatalf("ListBondEvents for %d: %v", uid, err)
		}
		if len(events) != 1 {
			t.Errorf("user %d: expected 1 event, got %d", uid, len(events))
		}
	}
}

func TestListBondEvents_OrderByCreatedAtDesc(t *testing.T) {
	repo := newTestProfileRepo(t)
	ctx := context.Background()

	aID := createTestUserWithProfile(t, repo, "bond-order-a")

	// Insert bonds with different people.
	bID1 := createTestUserWithProfile(t, repo, "bond-order-b1")
	bID2 := createTestUserWithProfile(t, repo, "bond-order-b2")

	bond := makeTestBond()
	repo.InsertBondEvent(ctx, profile.InsertBondParams{LinkID: nil, InitiatorID: aID, OtherID: bID1, AssessmentID: nil, Bond: bond})
	time.Sleep(10 * time.Millisecond)
	repo.InsertBondEvent(ctx, profile.InsertBondParams{LinkID: nil, InitiatorID: aID, OtherID: bID2, AssessmentID: nil, Bond: bond})

	events, _ := repo.ListBondEvents(ctx, aID)
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].CreatedAt.Before(events[1].CreatedAt) {
		t.Error("events should be ordered by created_at DESC")
	}
}

func TestListBondEvents_RFC3339Format(t *testing.T) {
	repo := newTestProfileRepo(t)
	ctx := context.Background()

	aID := createTestUserWithProfile(t, repo, "bond-rfc-a")
	bID := createTestUserWithProfile(t, repo, "bond-rfc-b")

	bond := makeTestBond()
	if err := repo.InsertBondEvent(ctx, profile.InsertBondParams{LinkID: nil, InitiatorID: aID, OtherID: bID, AssessmentID: nil, Bond: bond}); err != nil {
		t.Fatalf("InsertBondEvent: %v", err)
	}

	events, _ := repo.ListBondEvents(ctx, aID)
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if _, err := time.Parse(time.RFC3339, events[0].CreatedAt.Format(time.RFC3339)); err != nil {
		t.Errorf("created_at should be valid RFC3339: %v", err)
	}
	if events[0].CreatedAt.IsZero() {
		t.Error("created_at should not be zero")
	}
}
