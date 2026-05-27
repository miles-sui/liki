-- 001_init.sql — Schema for Fivefold Types
-- Applied by migration runner on first run.

PRAGMA foreign_keys = ON;

-- ============================================================
-- users
-- ============================================================
CREATE TABLE IF NOT EXISTS users (
    id                        INTEGER PRIMARY KEY AUTOINCREMENT,
    name                      TEXT NOT NULL UNIQUE,
    email                     TEXT NOT NULL DEFAULT '',
    password_hash             TEXT NOT NULL,
    token_version             INTEGER NOT NULL DEFAULT 1,
    is_public                 INTEGER NOT NULL DEFAULT 0 CHECK (is_public IN (0, 1)),
    email_verified_at         TEXT,
    pending_email             TEXT NOT NULL DEFAULT '',
    email_verification_token  TEXT,
    password_reset_token      TEXT,
    password_reset_expires_at TEXT,
    deactivated_at            TEXT,
    created_at                TEXT DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at                TEXT DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);
CREATE INDEX IF NOT EXISTS idx_users_deactivated ON users(deactivated_at)
    WHERE deactivated_at IS NOT NULL;

-- ============================================================
-- assessments
-- ============================================================
CREATE TABLE IF NOT EXISTS assessments (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id           INTEGER REFERENCES users(id),
    assessment_type   TEXT NOT NULL DEFAULT 'self' CHECK (assessment_type IN ('self', 'peer')),
    identity_id       TEXT NOT NULL DEFAULT '',
    answers_json      TEXT NOT NULL DEFAULT '[]',
    profile_json      TEXT NOT NULL DEFAULT '{}',
    created_at        TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    review_link_id    INTEGER REFERENCES review_links(id),
    reviewer_name     TEXT NOT NULL DEFAULT '',
    legacy_user_token TEXT NOT NULL DEFAULT '',
    CHECK (assessment_type != 'peer' OR review_link_id IS NOT NULL),
    CHECK (assessment_type != 'self' OR (review_link_id IS NULL AND reviewer_name = ''))
);

-- Immutability trigger: core data cannot change once created (only user_id/legacy_user_token for claims).
CREATE TRIGGER IF NOT EXISTS trg_assessments_no_update BEFORE UPDATE ON assessments
WHEN OLD.identity_id != NEW.identity_id
   OR OLD.answers_json != NEW.answers_json
   OR OLD.profile_json != NEW.profile_json
   OR OLD.assessment_type != NEW.assessment_type
   OR (OLD.review_link_id IS NOT NULL AND OLD.review_link_id != NEW.review_link_id)
BEGIN
    SELECT RAISE(ABORT, 'assessment core data is immutable');
END;

CREATE INDEX IF NOT EXISTS idx_assessments_user ON assessments(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_assessments_review_link ON assessments(review_link_id);
CREATE INDEX IF NOT EXISTS idx_assessments_type ON assessments(assessment_type);
CREATE INDEX IF NOT EXISTS idx_assessments_anonymous_token ON assessments(legacy_user_token)
    WHERE legacy_user_token != '';

-- ============================================================
-- review_links
-- ============================================================
CREATE TABLE IF NOT EXISTS review_links (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    subject_user_id INTEGER NOT NULL REFERENCES users(id),
    token           TEXT UNIQUE NOT NULL,
    expires_at      TEXT,
    created_at      TEXT DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    deleted_at      TEXT
);
CREATE INDEX IF NOT EXISTS idx_review_links_subject ON review_links(subject_user_id);

-- ============================================================
-- match_requests
-- ============================================================
CREATE TABLE IF NOT EXISTS match_requests (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    from_user_id INTEGER NOT NULL REFERENCES users(id),
    to_user_id   INTEGER NOT NULL REFERENCES users(id),
    status       TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'declined')),
    created_at   TEXT DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at   TEXT DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);
CREATE INDEX IF NOT EXISTS idx_match_requests_to ON match_requests(to_user_id, status);
CREATE INDEX IF NOT EXISTS idx_match_requests_from ON match_requests(from_user_id);

-- Only one active request per (from, to) pair.
CREATE UNIQUE INDEX IF NOT EXISTS idx_match_active ON match_requests(from_user_id, to_user_id)
    WHERE status IN ('pending', 'accepted');

-- Prevent cross-direction duplicate active requests (A→B and B→A simultaneously).
CREATE TRIGGER IF NOT EXISTS trg_match_no_cross BEFORE INSERT ON match_requests
WHEN EXISTS (
    SELECT 1 FROM match_requests
    WHERE from_user_id = NEW.to_user_id
      AND to_user_id = NEW.from_user_id
      AND status IN ('pending', 'accepted')
)
BEGIN
    SELECT RAISE(ABORT, 'A match request already exists between these users');
END;

-- ============================================================
-- schema_migrations (migration tracking)
-- ============================================================
CREATE TABLE IF NOT EXISTS schema_migrations (
    version   INTEGER PRIMARY KEY,
    applied_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);
