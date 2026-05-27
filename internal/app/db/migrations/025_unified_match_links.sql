-- 025_unified_match_links.sql
-- Merge assessment match_links and bazi_match_links into a single table with a type column.

-- Add type column to match_links (default 'assessment' for existing rows).
ALTER TABLE match_links ADD COLUMN type TEXT NOT NULL DEFAULT 'assessment';

-- Migrate bazi_match_links into match_links.
INSERT INTO match_links (user_id, token, created_at, deleted_at, type)
SELECT user_id, token, created_at, deleted_at, 'bazi'
FROM bazi_match_links
WHERE deleted_at IS NULL OR true;

-- Drop the old bazi_match_links table.
DROP TABLE bazi_match_links;

-- Update bond_events.link_id references: bazi_match_events.link_id now points to match_links(id).
-- bazi_match_events stays, but its link_id now references the unified match_links table.
-- No schema change needed for bazi_match_events — link_id is INTEGER REFERENCES match_links(id).
