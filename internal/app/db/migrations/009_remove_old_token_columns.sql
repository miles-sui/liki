-- 009_remove_old_token_columns.sql
-- @no_fk — this migration rebuilds the users table, which is referenced by
-- FK constraints from other tables. The rebuild preserves all PKs so referential
-- integrity is maintained after the swap.
-- Drop the 4 dead token columns from users (replaced by user_tokens in 007)
-- and the index that references a dead column.
-- Also adds created_at to code_redemptions (missing since 002).

-- 1. Drop the dead index first.
DROP INDEX IF EXISTS idx_users_email_ver_token;

-- 2. Rebuild users without the 4 dead token columns.
CREATE TABLE users_new (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    name              TEXT NOT NULL UNIQUE,
    email             TEXT NOT NULL DEFAULT '',
    password_hash     TEXT NOT NULL,
    token_version     INTEGER NOT NULL DEFAULT 1,
    is_public         INTEGER NOT NULL DEFAULT 0 CHECK (is_public IN (0, 1)),
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
    is_public, email_verified_at, pending_email, deactivated_at,
    created_at, updated_at
FROM users;

DROP TABLE users;
ALTER TABLE users_new RENAME TO users;

-- 3. Recreate indexes.
CREATE INDEX IF NOT EXISTS idx_users_deactivated ON users(deactivated_at)
    WHERE deactivated_at IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_unique
    ON users(email) WHERE email != '';

-- 4. Add created_at to code_redemptions (missing since 002).
-- Existing rows get NULL; new rows get the default timestamp.
ALTER TABLE code_redemptions ADD COLUMN created_at TEXT
    DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'));
