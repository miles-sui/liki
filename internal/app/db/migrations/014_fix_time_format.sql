-- Migration 014: Fix created_at time format and add missing index.
-- match_links and bond_events used datetime('now') which produces
-- "2026-05-16 14:16:07" but Go code expects RFC3339 "2026-05-16T14:16:07Z".
-- This migration converts existing data to RFC3339 format.
-- Also adds index on bond_events.other_user_id for the OR query in ListBondEvents.

UPDATE match_links SET created_at = strftime('%Y-%m-%dT%H:%M:%SZ', created_at)
    WHERE created_at NOT LIKE '%T%';

UPDATE bond_events SET created_at = strftime('%Y-%m-%dT%H:%M:%SZ', created_at)
    WHERE created_at NOT LIKE '%T%';

CREATE INDEX idx_bond_events_other ON bond_events(other_user_id);
