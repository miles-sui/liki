-- 002_redemption_codes.sql
-- Activation code system for passport/subscription redemption.

-- plan: "monthly" / "yearly" / "code" — acquisition channel.
-- bond_count: number of full Bond computations viewed (first one is free).
CREATE TABLE IF NOT EXISTS user_subscriptions (
    user_id             INTEGER PRIMARY KEY REFERENCES users(id),
    has_passport        INTEGER NOT NULL DEFAULT 0,
    passport_expires_at TEXT,
    plan                TEXT    NOT NULL DEFAULT '',
    bond_count          INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS redemption_codes (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    code        TEXT    NOT NULL UNIQUE,
    duration_d  INTEGER NOT NULL CHECK (duration_d >= 0),  -- 0 = permanent
    max_uses    INTEGER NOT NULL DEFAULT 1,                 -- 0 = unlimited
    created_by  TEXT    NOT NULL DEFAULT '',
    notes       TEXT    NOT NULL DEFAULT '',
    expires_at  TEXT,                                       -- code expiry (NULL = never)
    created_at  TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE TABLE IF NOT EXISTS code_redemptions (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    code_id    INTEGER NOT NULL REFERENCES redemption_codes(id),
    user_id    INTEGER NOT NULL REFERENCES users(id)
);
