-- 026_bazi_match_events_index.sql
-- Add index on bazi_match_events.link_id for the GROUP BY link_id subquery
-- in MatchLinkRepo.ListByUser. Without this, every match link list causes
-- a full scan of bazi_match_events.

CREATE INDEX IF NOT EXISTS idx_bazi_match_events_link_id ON bazi_match_events(link_id);
