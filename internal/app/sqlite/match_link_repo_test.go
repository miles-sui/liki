package sqlite

import (
	"context"
	"errors"
	"testing"

	"github.com/25types/25types/internal/app/domain"
)

func newTestMatchLinkRepo(t *testing.T) *MatchLinkRepo {
	t.Helper()
	return NewMatchLinkRepo(openTestDB(t))
}

// =============================================================================
// Create
// =============================================================================

func TestMatchLinkRepo_Create(t *testing.T) {
	repo := newTestMatchLinkRepo(t)
	userID := createTestUser(t, NewUserRepo(openTestDB(t)), "ml-creator")

	id, err := repo.Create(context.Background(), userID, "token-abc", "assessment")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected id > 0, got %d", id)
	}
}

func TestMatchLinkRepo_Create_TokenCollision(t *testing.T) {
	repo := newTestMatchLinkRepo(t)
	userID := createTestUser(t, NewUserRepo(openTestDB(t)), "ml-collide")

	_, err := repo.Create(context.Background(), userID, "same-token", "assessment")
	if err != nil {
		t.Fatalf("first Create: %v", err)
	}
	_, err = repo.Create(context.Background(), userID, "same-token", "assessment")
	if err == nil {
		t.Error("expected unique constraint error on duplicate token")
	}
}

// =============================================================================
// FindByToken
// =============================================================================

func TestMatchLinkRepo_FindByToken(t *testing.T) {
	repo := newTestMatchLinkRepo(t)
	userID := createTestUser(t, NewUserRepo(openTestDB(t)), "ml-finder")

	_, err := repo.Create(context.Background(), userID, "find-me", "assessment")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	ml, err := repo.FindByToken(context.Background(), "find-me")
	if err != nil {
		t.Fatalf("FindByToken: %v", err)
	}
	if ml.ID <= 0 {
		t.Errorf("expected ID > 0, got %d", ml.ID)
	}
	if ml.UserID != userID {
		t.Errorf("expected UserID %d, got %d", userID, ml.UserID)
	}
	if ml.Token != "find-me" {
		t.Errorf("expected token 'find-me', got %q", ml.Token)
	}
}

func TestMatchLinkRepo_FindByToken_NotFound(t *testing.T) {
	repo := newTestMatchLinkRepo(t)

	_, err := repo.FindByToken(context.Background(), "no-such-token")
	if err == nil {
		t.Error("expected error for unknown token")
	}
	if !errors.Is(err, domain.ErrMatchLinkNotFound) {
		t.Errorf("expected ErrMatchLinkNotFound, got %v", err)
	}
}

func TestMatchLinkRepo_FindByToken_SoftDeleted(t *testing.T) {
	repo := newTestMatchLinkRepo(t)
	userID := createTestUser(t, NewUserRepo(openTestDB(t)), "ml-del")

	id, err := repo.Create(context.Background(), userID, "delete-me", "assessment")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	ok, err := repo.SoftDelete(context.Background(), id, userID)
	if err != nil || !ok {
		t.Fatalf("SoftDelete: %v / ok=%v", err, ok)
	}

	_, err = repo.FindByToken(context.Background(), "delete-me")
	if !errors.Is(err, domain.ErrMatchLinkNotFound) {
		t.Errorf("expected ErrMatchLinkNotFound for soft-deleted link, got %v", err)
	}
}

// =============================================================================
// ListByUser
// =============================================================================

func TestMatchLinkRepo_ListByUser(t *testing.T) {
	repo := newTestMatchLinkRepo(t)
	userID := createTestUser(t, NewUserRepo(openTestDB(t)), "ml-lister")

	for i := 0; i < 3; i++ {
		if _, err := repo.Create(context.Background(), userID, "list-token-"+string(rune('a'+i)), "assessment"); err != nil {
			t.Fatalf("Create %d: %v", i, err)
		}
	}

	items, err := repo.ListByUser(context.Background(), userID, "assessment")
	if err != nil {
		t.Fatalf("ListByUser: %v", err)
	}
	if len(items) != 3 {
		t.Errorf("expected 3 items, got %d", len(items))
	}
	// Should be DESC by id
	if items[0].ID < items[1].ID {
		t.Error("expected items in descending id order")
	}
}

func TestMatchLinkRepo_ListByUser_Empty(t *testing.T) {
	repo := newTestMatchLinkRepo(t)
	userID := createTestUser(t, NewUserRepo(openTestDB(t)), "ml-empty")

	items, err := repo.ListByUser(context.Background(), userID, "assessment")
	if err != nil {
		t.Fatalf("ListByUser: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}

func TestMatchLinkRepo_ListByUser_ExcludesDeleted(t *testing.T) {
	repo := newTestMatchLinkRepo(t)
	userID := createTestUser(t, NewUserRepo(openTestDB(t)), "ml-excl")

	if _, err := repo.Create(context.Background(), userID, "keep-me", "assessment"); err != nil {
		t.Fatalf("Create keep: %v", err)
	}
	delID, err := repo.Create(context.Background(), userID, "del-me", "assessment")
	if err != nil {
		t.Fatalf("Create del: %v", err)
	}
	if _, err := repo.SoftDelete(context.Background(), delID, userID); err != nil {
		t.Fatalf("SoftDelete: %v", err)
	}

	items, err := repo.ListByUser(context.Background(), userID, "assessment")
	if err != nil {
		t.Fatalf("ListByUser: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 non-deleted item, got %d", len(items))
	}
	if items[0].Token != "keep-me" {
		t.Errorf("expected 'keep-me', got %q", items[0].Token)
	}
}

// =============================================================================
// SoftDelete
// =============================================================================

func TestMatchLinkRepo_SoftDelete(t *testing.T) {
	repo := newTestMatchLinkRepo(t)
	userID := createTestUser(t, NewUserRepo(openTestDB(t)), "ml-softdel")

	id, err := repo.Create(context.Background(), userID, "soft-delete-me", "assessment")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	ok, err := repo.SoftDelete(context.Background(), id, userID)
	if err != nil {
		t.Fatalf("SoftDelete: %v", err)
	}
	if !ok {
		t.Error("expected SoftDelete to return true")
	}
}

func TestMatchLinkRepo_SoftDelete_WrongUser(t *testing.T) {
	repo := newTestMatchLinkRepo(t)
	ownerID := createTestUser(t, NewUserRepo(openTestDB(t)), "ml-owner")
	otherID := createTestUser(t, NewUserRepo(openTestDB(t)), "ml-other")

	id, err := repo.Create(context.Background(), ownerID, "not-yours", "assessment")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	ok, err := repo.SoftDelete(context.Background(), id, otherID)
	if err != nil {
		t.Fatalf("SoftDelete: %v", err)
	}
	if ok {
		t.Error("expected SoftDelete to return false for non-owner")
	}
}

func TestMatchLinkRepo_SoftDelete_AlreadyDeleted(t *testing.T) {
	repo := newTestMatchLinkRepo(t)
	userID := createTestUser(t, NewUserRepo(openTestDB(t)), "ml-twice")

	id, err := repo.Create(context.Background(), userID, "twice", "assessment")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := repo.SoftDelete(context.Background(), id, userID); err != nil {
		t.Fatalf("first SoftDelete: %v", err)
	}

	ok, err := repo.SoftDelete(context.Background(), id, userID)
	if err != nil {
		t.Fatalf("second SoftDelete: %v", err)
	}
	if ok {
		t.Error("expected SoftDelete to return false for already deleted")
	}
}
