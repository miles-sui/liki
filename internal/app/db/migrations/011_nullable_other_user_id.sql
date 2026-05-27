-- Migration 011: Make other_user_id nullable in bond_events.
-- Match link bonds involve an anonymous recipient (no user account yet),
-- so other_user_id can be NULL. The link_id column identifies the source.
-- We also add assessment_id to link back to the anonymous assessment.

ALTER TABLE bond_events RENAME TO bond_events_old;

CREATE TABLE bond_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    link_id INTEGER REFERENCES match_links(id),
    initiator_user_id INTEGER NOT NULL REFERENCES users(id),
    other_user_id INTEGER REFERENCES users(id),
    assessment_id INTEGER REFERENCES assessments(id),
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

INSERT INTO bond_events (id, link_id, initiator_user_id, other_user_id, created_at)
    SELECT id, link_id, initiator_user_id, other_user_id, created_at FROM bond_events_old;

DROP TABLE bond_events_old;

CREATE INDEX idx_bond_events_initiator ON bond_events(initiator_user_id);
CREATE INDEX idx_bond_events_link_id ON bond_events(link_id);
