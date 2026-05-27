-- 016: Add match_links(user_id) partial index and drop redundant is_deleted column.
-- is_deleted was replaced by deleted_at in migration 015.

-- Partial index for listing active match links by user.
CREATE INDEX idx_match_links_user ON match_links(user_id) WHERE deleted_at IS NULL;

-- Drop the now-redundant is_deleted column (SQLite 3.35+).
ALTER TABLE match_links DROP COLUMN is_deleted;
