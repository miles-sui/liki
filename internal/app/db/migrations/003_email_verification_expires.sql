-- 003_email_verification_expires.sql
-- Add expiry column for email verification tokens.
-- Password reset tokens already have password_reset_expires_at;
-- this brings email verification to parity.

ALTER TABLE users ADD COLUMN email_verification_expires_at TEXT;
