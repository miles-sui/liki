-- 004_code_redemptions_unique.sql
-- Prevent a user from redeeming the same code more than once.
-- Without this index, the ErrCodeAlreadyRedeemed path in commerce/service.go
-- is unreachable (no UNIQUE violation to detect).

CREATE UNIQUE INDEX IF NOT EXISTS idx_code_redemptions_unique
    ON code_redemptions(code_id, user_id);
