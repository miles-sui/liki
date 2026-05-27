-- 006_index_cleanup.sql
-- Fix index strategy: add missing indexes for token/email lookup paths,
-- drop a low-selectivity index that wastes write I/O,
-- and optimize the FindLatestProfile query with a partial index.

-- Password reset lookup by verified email (SetPasswordResetToken).
-- Without this, every password reset request scans the full users table.
CREATE INDEX IF NOT EXISTS idx_users_email_verified
    ON users(email) WHERE email_verified_at IS NOT NULL;

-- Email verification token lookup (VerifyEmailByToken).
CREATE INDEX IF NOT EXISTS idx_users_email_ver_token
    ON users(email_verification_token)
    WHERE email_verification_token IS NOT NULL;

-- Prevent duplicate verified emails. The application does not enforce this,
-- but two accounts sharing the same verified email breaks password reset
-- (SetPasswordResetToken matches the first row arbitrarily).
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_unique
    ON users(email) WHERE email != '';

-- Low-selectivity: only two values ('self'/'peer'), SQLite ignores it.
-- Every INSERT on assessments pays the maintenance cost for no query benefit.
DROP INDEX IF EXISTS idx_assessments_type;

-- Partial index for FindLatestProfile: WHERE user_id=? AND assessment_type='self'
-- ORDER BY id DESC LIMIT 1. Smaller and more targeted than the general
-- idx_assessments_user which has to cover both assessment types.
CREATE INDEX IF NOT EXISTS idx_assessments_user_self
    ON assessments(user_id, id DESC) WHERE assessment_type = 'self';
