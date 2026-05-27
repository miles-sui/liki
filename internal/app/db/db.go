package db

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Open opens the SQLite database at path, applies PRAGMA configuration,
// and runs pending migrations. The returned *sql.DB is ready to use.
func Open(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("mkdir: %w", err)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	// SQLite single-writer — one connection is optimal.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	// PRAGMAs per DATA-MODEL.md §5.
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA journal_size_limit=67108864",
		"PRAGMA mmap_size=33554432",
		"PRAGMA auto_vacuum=INCREMENTAL",
		"PRAGMA busy_timeout=5000",
		"PRAGMA foreign_keys=ON",
		"PRAGMA cache_size=-8000",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			db.Close()
			return nil, fmt.Errorf("pragma %s: %w", p, err)
		}
	}

	if err := runMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return db, nil
}

// runMigrations reads embedded .sql files in order and applies any not yet run.
func runMigrations(db *sql.DB) error {
	// Ensure tracking table exists (not tracked as a migration itself).
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version   INTEGER PRIMARY KEY,
		applied_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
	)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	applied := appliedVersions(db)

	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	type mig struct {
		version int
		name    string
	}
	var migs []mig
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		v, err := strconv.Atoi(strings.SplitN(e.Name(), "_", 2)[0])
		if err != nil {
			return fmt.Errorf("parse migration version from %s: %w", e.Name(), err)
		}
		migs = append(migs, mig{version: v, name: e.Name()})
	}
	sort.Slice(migs, func(i, j int) bool { return migs[i].version < migs[j].version })

	for _, m := range migs {
		if applied[m.version] {
			continue
		}
		log.Printf("[db] applying migration %s", m.name)
		sqlBytes, err := migrationsFS.ReadFile("migrations/" + m.name)
		if err != nil {
			return fmt.Errorf("read %s: %w", m.name, err)
		}
		if err := execMigrationTx(db, string(sqlBytes), m.version); err != nil {
			return fmt.Errorf("apply %s: %w", m.name, err)
		}
	}
	return nil
}

// execMigrationTx runs all statements in a migration file inside a single transaction,
// including the schema_migrations version record. If any statement fails, the entire
// migration rolls back — no partial state.
//
// Migrations that rebuild tables with incoming FK constraints must declare
//
//	-- @no_fk
//
// at the top of the file. For those, foreign key checks are temporarily disabled
// at the connection level for the duration of the transaction. All other
// migrations run with FK checks enabled as a safety net.
func execMigrationTx(db *sql.DB, sqlText string, version int) error {
	noFK := strings.Contains(sqlText, "@no_fk")
	if noFK {
		db.Exec("PRAGMA foreign_keys=OFF")
		defer db.Exec("PRAGMA foreign_keys=ON")
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	clean := stripSQLComments(sqlText)
	stmts := splitSQL(clean)
	for _, s := range stmts {
		if _, err := tx.Exec(s); err != nil {
			preview := s
			if len(preview) > 120 {
				preview = preview[:120]
			}
			return fmt.Errorf("%w\n  SQL: %s", err, preview)
		}
	}

	if _, err := tx.Exec(`INSERT INTO schema_migrations (version) VALUES (?)`, version); err != nil {
		return fmt.Errorf("record version: %w", err)
	}

	return tx.Commit()
}

// stripSQLComments removes -- line comments from SQL text.
func stripSQLComments(s string) string {
	lines := strings.Split(s, "\n")
	var out []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "--") {
			continue
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

// splitSQL splits SQL text on top-level semicolons (outside BEGIN/END blocks and strings).
func splitSQL(s string) []string {
	var out []string
	start := 0
	blockDepth := 0 // for BEGIN...END
	inString := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if inString {
			if ch == '\'' {
				inString = false
			}
			continue
		}
		if ch == '\'' {
			inString = true
			continue
		}
		// Track BEGIN/END blocks (case-insensitive word boundary check).
		if isWordAt(s, i, "BEGIN") && blockBoundary(s, i, "BEGIN") {
			blockDepth++
		} else if isWordAt(s, i, "END") && blockBoundary(s, i, "END") {
			if blockDepth > 0 {
				blockDepth--
			}
		}
		if ch == ';' && blockDepth == 0 {
			stmt := strings.TrimSpace(s[start:i])
			if stmt != "" {
				out = append(out, stmt)
			}
			start = i + 1
		}
	}
	last := strings.TrimSpace(s[start:])
	if last != "" {
		out = append(out, last)
	}
	return out
}

func isWordAt(s string, i int, word string) bool {
	return i+len(word) <= len(s) && strings.EqualFold(s[i:i+len(word)], word)
}

func blockBoundary(s string, i int, word string) bool {
	// Must be at a word boundary: preceded by non-alphanumeric (or start), followed by non-alphanumeric (or end).
	before := i == 0 || !isAlpha(s[i-1])
	after := i+len(word) >= len(s) || !isAlpha(s[i+len(word)])
	return before && after
}

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

func appliedVersions(db *sql.DB) map[int]bool {
	rows, err := db.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return map[int]bool{}
	}
	defer rows.Close()
	m := map[int]bool{}
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err == nil {
			m[v] = true
		}
	}
	return m
}
