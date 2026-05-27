-- Migration 020: Add payment_id to donations for idempotency.
-- Both the confirm endpoint and webhook use payment_id as a dedup key,
-- so the same Dodo payment never creates duplicate donation records.

ALTER TABLE donations ADD COLUMN payment_id TEXT NOT NULL DEFAULT '';
CREATE UNIQUE INDEX idx_donations_payment_id ON donations(payment_id);
