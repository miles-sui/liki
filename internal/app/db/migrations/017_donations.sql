-- Migration 017: Add donation support
-- Replaces the removed commerce tables with a minimal donation tracking table.
-- Supporter status is derived from the existence of a donation record.

CREATE TABLE donations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id),
    amount INTEGER NOT NULL CHECK (amount > 0),
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);
CREATE INDEX idx_donations_user_id ON donations(user_id);
