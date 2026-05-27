-- 007_user_tokens.sql
-- Extract email verification and password reset tokens from the users table
-- into a dedicated user_tokens table.
--
-- Rationale:
--   1. The two token+expiry column pairs on users are structurally identical.
--      A third token type (magic link, OAuth state, etc.) would need yet another
--      ALTER TABLE ADD COLUMN pair — user_tokens absorbs that with a CHECK value.
--   2. Tokens have independent lifecycles (create, verify, expire, delete) that
--      do not belong on the core user identity row.
--   3. Looking up a token no longer requires scanning all 15 user columns — the
--      user_tokens table is narrow and indexed.

-- The token table. One row per (user, token type). Old rows are replaced on re-issue.
CREATE TABLE IF NOT EXISTS user_tokens (
    user_id    INTEGER NOT NULL REFERENCES users(id),
    token_type TEXT    NOT NULL CHECK (token_type IN ('email_verify', 'password_reset')),
    token      TEXT    NOT NULL,
    expires_at TEXT    NOT NULL,
    created_at TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    PRIMARY KEY (user_id, token_type)
);

-- Lookup by token value (the common query pattern: "who owns this token?").
CREATE INDEX IF NOT EXISTS idx_user_tokens_token ON user_tokens(token);

-- Migrate existing email verification tokens.
INSERT OR IGNORE INTO user_tokens (user_id, token_type, token, expires_at)
    SELECT id, 'email_verify', email_verification_token, email_verification_expires_at
    FROM users
    WHERE email_verification_token IS NOT NULL;

-- Migrate existing password reset tokens.
INSERT OR IGNORE INTO user_tokens (user_id, token_type, token, expires_at)
    SELECT id, 'password_reset', password_reset_token, password_reset_expires_at
    FROM users
    WHERE password_reset_token IS NOT NULL;

-- Old columns on users are left in place (SQLite ALTER TABLE DROP COLUMN
-- requires a table rebuild). They are no longer read or written by the
-- application. A future major-version migration can remove them safely via
-- the CREATE ... INSERT ... DROP ... RENAME pattern if desired.
