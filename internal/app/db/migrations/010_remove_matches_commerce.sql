-- Migration 010: Remove match requests and commerce/payments
-- Part of UX redesign: Profile-as-Product simplification.
-- Drops match_requests (replaced by match_links + bond_events),
-- drops commerce tables (payments removed).

-- Drop match-related objects
DROP TABLE IF EXISTS match_requests;
DROP TRIGGER IF EXISTS trg_match_no_cross;

-- Drop commerce/payment tables
DROP TABLE IF EXISTS subscription_events;
DROP TABLE IF EXISTS code_redemptions;
DROP TABLE IF EXISTS redemption_codes;
DROP TABLE IF EXISTS user_subscriptions;

-- New: Match links (paid feature, free for now)
CREATE TABLE match_links (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id),
    token TEXT NOT NULL UNIQUE,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    is_deleted INTEGER NOT NULL DEFAULT 0
);

-- New: Bond events (instant compare + match link, unified storage)
CREATE TABLE bond_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    link_id INTEGER REFERENCES match_links(id),  -- NULL for instant compare
    initiator_user_id INTEGER NOT NULL REFERENCES users(id),
    other_user_id INTEGER NOT NULL REFERENCES users(id),
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX idx_bond_events_initiator ON bond_events(initiator_user_id);
CREATE INDEX idx_bond_events_link_id ON bond_events(link_id);
