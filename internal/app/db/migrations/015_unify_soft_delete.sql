-- Migration 015: Unify soft-delete strategy.
-- match_links used is_deleted INTEGER (0/1) while review_links used deleted_at TEXT.
-- Adds deleted_at column and populates it from is_deleted rows.
-- is_deleted is kept but no longer read by Go code; can be dropped in a future
-- migration when the table is recreated for another reason.

ALTER TABLE match_links ADD COLUMN deleted_at TEXT;

UPDATE match_links SET deleted_at = created_at WHERE is_deleted = 1;

-- Fix the created_at DEFAULT while we're here (same issue as bond_events).
-- SQLite doesn't support ALTER COLUMN SET DEFAULT, so we leave it for now.
-- Go code always provides created_at in INSERT, so the DEFAULT is never used.
