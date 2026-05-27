-- Migration 019: Fix donations.created_at time format.
-- 017 used datetime('now') which produces "2026-05-16 14:16:07"
-- but Go code expects RFC3339 "2026-05-16T14:16:07Z".
-- Same pattern as migration 014 for match_links/bond_events.

UPDATE donations SET created_at = strftime('%Y-%m-%dT%H:%M:%SZ', created_at)
    WHERE created_at NOT LIKE '%T%';
