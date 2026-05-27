package db

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpen_CreatesAndMigrates(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")

	database, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer database.Close()

	// Verify schema_migrations table records the migration.
	var count int
	if err := database.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count); err != nil {
		t.Fatalf("query schema_migrations: %v", err)
	}
	if count == 0 {
		t.Error("expected at least one migration applied")
	}

	// Verify all tables exist.
	tables := []string{"users", "assessments", "review_links",
		"user_tokens", "frontend_errors", "match_links", "bond_events"}
	for _, table := range tables {
		var n int
		if err := database.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&n); err != nil {
			t.Errorf("%s: %v", table, err)
		}
		if n == 0 {
			t.Errorf("table %s not created", table)
		}
	}

	// Verify triggers exist.
	triggers := []string{"trg_assessments_no_update"}
	for _, trig := range triggers {
		var n int
		if err := database.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='trigger' AND name=?", trig).Scan(&n); err != nil {
			t.Errorf("%s: %v", trig, err)
		}
		if n == 0 {
			t.Errorf("trigger %s not created", trig)
		}
	}

	// Verify indexes exist.
	indexes := []string{"idx_users_deactivated", "idx_assessments_user",
		"idx_review_links_subject",
		"idx_users_email_unique", "idx_user_tokens_token",
		"idx_frontend_errors_created_at", "idx_bond_events_initiator", "idx_bond_events_link_id"}
	for _, idx := range indexes {
		var n int
		if err := database.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name=?", idx).Scan(&n); err != nil {
			t.Errorf("%s: %v", idx, err)
		}
		if n == 0 {
			t.Errorf("index %s not created", idx)
		}
	}
}

func TestOpen_Idempotent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")

	db1, err := Open(path)
	if err != nil {
		t.Fatalf("first Open: %v", err)
	}
	db1.Close()

	db2, err := Open(path)
	if err != nil {
		t.Fatalf("second Open: %v", err)
	}
	defer db2.Close()

	// Should still have exactly the same number of migrations.
	var count int
	if err := db2.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count); err != nil {
		t.Fatalf("query: %v", err)
	}
	if count == 0 {
		t.Error("expected migrations tracked")
	}
}

func TestOpen_WALMode(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")

	db, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	var mode string
	if err := db.QueryRow("PRAGMA journal_mode").Scan(&mode); err != nil {
		t.Fatalf("pragma: %v", err)
	}
	if mode != "wal" {
		t.Errorf("journal_mode = %s, want wal", mode)
	}
}

func TestOpen_ForeignKeysEnabled(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")

	db, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	var fk int
	if err := db.QueryRow("PRAGMA foreign_keys").Scan(&fk); err != nil {
		t.Fatalf("pragma: %v", err)
	}
	if fk != 1 {
		t.Errorf("foreign_keys = %d, want 1", fk)
	}
}

func TestAssessments_ImmutabilityTrigger(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")

	db, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	// Insert a valid assessment.
	if _, err := db.Exec(`INSERT INTO assessments (identity_id, answers_json, profile_json) VALUES (?, ?, ?)`,
		"W", `[]`, `{"d":[],"p":[]}`); err != nil {
		t.Fatalf("insert: %v", err)
	}

	// Attempt to modify immutable field should fail.
	_, err = db.Exec(`UPDATE assessments SET identity_id = 'F' WHERE id = 1`)
	if err == nil {
		t.Error("expected trigger to prevent identity_id modification")
	}
}

func TestMatchLinks_UniqueToken(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")

	db, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	// Insert a user first (FK).
	db.Exec(`INSERT INTO users (name, password_hash) VALUES ('u', 'h')`)

	// Insert a match link with a token.
	if _, err := db.Exec(`INSERT INTO match_links (user_id, token) VALUES (1, 'tok1')`); err != nil {
		t.Fatalf("insert match link: %v", err)
	}

	// Same token should fail.
	_, err = db.Exec(`INSERT INTO match_links (user_id, token) VALUES (1, 'tok1')`)
	if err == nil {
		t.Error("expected UNIQUE constraint on match_links.token")
	}
}

func TestBondEvents_Insert(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")

	db, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	// Insert two users (FKs).
	db.Exec(`INSERT INTO users (name, password_hash) VALUES ('a', 'h')`)
	db.Exec(`INSERT INTO users (name, password_hash) VALUES ('b', 'h')`)

	// Insert a bond event (instant compare, no link).
	if _, err := db.Exec(`INSERT INTO bond_events (initiator_user_id, other_user_id) VALUES (1, 2)`); err != nil {
		t.Fatalf("insert bond event: %v", err)
	}

	// Insert a match link and a bond event linked to it.
	db.Exec(`INSERT INTO match_links (user_id, token) VALUES (1, 'link-tok')`)
	if _, err := db.Exec(`INSERT INTO bond_events (link_id, initiator_user_id, other_user_id) VALUES (1, 1, 2)`); err != nil {
		t.Fatalf("insert bond event with link: %v", err)
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM bond_events").Scan(&count)
	if count != 2 {
		t.Errorf("bond_events count = %d, want 2", count)
	}
}

func TestOpen_NonexistentDir(t *testing.T) {
	path := filepath.Join(os.TempDir(), "nonexistent-25types-test", "subdir", "test.db")

	db, err := Open(path)
	if err != nil {
		t.Fatalf("Open should create dirs: %v", err)
	}
	db.Close()
	os.RemoveAll(filepath.Dir(path))
}
