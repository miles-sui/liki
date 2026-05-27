-- 028_rename_mingli_tables.sql
-- Rename bazi_match_events to mingli_match_events for consistent mingli naming.

ALTER TABLE bazi_match_events RENAME TO mingli_match_events;
