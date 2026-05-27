-- 008_subscription_cleanup.sql
-- 1. Remove the redundant has_passport column from user_subscriptions.
--    passport_expires_at > now() already encodes "has passport" — the boolean
--    column adds no independent information and can drift from the date.
-- 2. Add subscription_events audit table so passport lifecycle is traceable.
-- 3. Add plan CHECK constraint — plan must be a known value.
-- 4. Tighten assessments CHECK constraints (identity_id and profile_json
--    must be non-empty for self-assessments).
-- 5. Drop idx_users_email_verified — idx_users_email_unique already covers
--    the same query paths (verified email implies email != '').

-- Audit log for subscription changes (purchase, code redemption).
CREATE TABLE IF NOT EXISTS subscription_events (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id    INTEGER NOT NULL REFERENCES users(id),
    event_type TEXT    NOT NULL CHECK (event_type IN ('purchase', 'redeem')),
    plan       TEXT    NOT NULL DEFAULT '',
    expires_at TEXT,
    created_at TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE INDEX IF NOT EXISTS idx_subscription_events_user
    ON subscription_events(user_id, created_at DESC);

-- Rebuild user_subscriptions: drop has_passport, add plan CHECK.
CREATE TABLE IF NOT EXISTS user_subscriptions_new (
    user_id             INTEGER PRIMARY KEY REFERENCES users(id),
    passport_expires_at TEXT,
    plan                TEXT    NOT NULL DEFAULT ''
                        CHECK (plan IN ('', 'monthly', 'yearly', 'code')),
    bond_count          INTEGER NOT NULL DEFAULT 0
);

INSERT INTO user_subscriptions_new (user_id, passport_expires_at, plan, bond_count)
    SELECT user_id, passport_expires_at, plan, bond_count
    FROM user_subscriptions;

DROP TABLE user_subscriptions;
ALTER TABLE user_subscriptions_new RENAME TO user_subscriptions;

-- Rebuild assessments: add self-assessment quality CHECKs.
CREATE TABLE IF NOT EXISTS assessments_new (
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
    CHECK (assessment_type != 'self' OR (review_link_id IS NULL AND reviewer_name = '')),
    CHECK (assessment_type != 'self' OR identity_id != ''),
    CHECK (assessment_type != 'self' OR profile_json != '{}')
);

INSERT INTO assessments_new SELECT * FROM assessments;

DROP TABLE assessments;
ALTER TABLE assessments_new RENAME TO assessments;

-- Recreate assessments indices.
CREATE INDEX IF NOT EXISTS idx_assessments_user ON assessments(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_assessments_review_link ON assessments(review_link_id);
CREATE INDEX IF NOT EXISTS idx_assessments_user_self ON assessments(user_id, id DESC)
    WHERE assessment_type = 'self';
CREATE INDEX IF NOT EXISTS idx_assessments_anonymous_token ON assessments(legacy_user_token)
    WHERE legacy_user_token != '';

-- Immutability trigger: core assessment data cannot be modified.
CREATE TRIGGER IF NOT EXISTS trg_assessments_no_update BEFORE UPDATE ON assessments
WHEN OLD.identity_id != NEW.identity_id
   OR OLD.answers_json != NEW.answers_json
   OR OLD.profile_json != NEW.profile_json
   OR OLD.assessment_type != NEW.assessment_type
   OR (OLD.review_link_id IS NOT NULL AND OLD.review_link_id != NEW.review_link_id)
BEGIN
    SELECT RAISE(ABORT, 'assessment core data is immutable');
END;

-- idx_users_email_unique (WHERE email != '') already serves the password-reset
-- lookup path, so this narrower partial index is redundant.
DROP INDEX IF EXISTS idx_users_email_verified;
