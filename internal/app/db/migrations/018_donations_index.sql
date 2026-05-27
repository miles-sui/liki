-- Migration 018: Add composite index on donations(user_id, created_at)
-- Enables efficient MIN(created_at) lookup per user for supporter_since.
CREATE INDEX idx_donations_user_created ON donations(user_id, created_at);
