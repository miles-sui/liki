-- Migration 013: Add bond_json and other_name to bond_events.
-- bond_json stores {self, other, delta_a, delta_b} as Deviation JSON snapshot
-- for historical bond chart display without recomputation.
-- other_name is an optional display name for anonymous recipients.

ALTER TABLE bond_events RENAME TO bond_events_old;

CREATE TABLE bond_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    link_id INTEGER REFERENCES match_links(id),
    initiator_user_id INTEGER NOT NULL REFERENCES users(id),
    other_user_id INTEGER REFERENCES users(id),
    other_name TEXT NOT NULL DEFAULT '',
    assessment_id INTEGER REFERENCES assessments(id),
    bond_json TEXT NOT NULL DEFAULT '{}',
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

INSERT INTO bond_events (id, link_id, initiator_user_id, other_user_id,
                         assessment_id, created_at)
    SELECT id, link_id, initiator_user_id, other_user_id,
           assessment_id, created_at FROM bond_events_old;

DROP TABLE bond_events_old;

CREATE INDEX idx_bond_events_initiator ON bond_events(initiator_user_id);
CREATE INDEX idx_bond_events_link_id ON bond_events(link_id);
