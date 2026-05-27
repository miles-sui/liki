-- 012_default_public.sql
-- Per plan: profiles are public by default. Change is_public default from 0 to 1.
-- @no_fk — rebuilds users table; PKs preserved so FK integrity is maintained.

CREATE TABLE users_new (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    name              TEXT NOT NULL UNIQUE,
    email             TEXT NOT NULL DEFAULT '',
    password_hash     TEXT NOT NULL,
    token_version     INTEGER NOT NULL DEFAULT 1,
    is_public         INTEGER NOT NULL DEFAULT 1 CHECK (is_public IN (0, 1)),
    email_verified_at TEXT,
    pending_email     TEXT NOT NULL DEFAULT '',
    deactivated_at    TEXT,
    created_at        TEXT DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at        TEXT DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

INSERT INTO users_new (id, name, email, password_hash, token_version,
    is_public, email_verified_at, pending_email, deactivated_at,
    created_at, updated_at)
SELECT id, name, email, password_hash, token_version,
    1, email_verified_at, pending_email, deactivated_at,
    created_at, updated_at
FROM users
WHERE deactivated_at IS NULL;

INSERT INTO users_new (id, name, email, password_hash, token_version,
    is_public, email_verified_at, pending_email, deactivated_at,
    created_at, updated_at)
SELECT id, name, email, password_hash, token_version,
    is_public, email_verified_at, pending_email, deactivated_at,
    created_at, updated_at
FROM users
WHERE deactivated_at IS NOT NULL;

DROP TABLE users;
ALTER TABLE users_new RENAME TO users;

CREATE INDEX IF NOT EXISTS idx_users_deactivated ON users(deactivated_at)
    WHERE deactivated_at IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_unique
    ON users(email) WHERE email != '';
