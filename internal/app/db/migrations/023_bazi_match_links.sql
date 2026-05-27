-- 023_bazi_match_links.sql — BaZi match sharing links + match event history.
CREATE TABLE IF NOT EXISTS bazi_match_links (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id),
    token TEXT NOT NULL UNIQUE,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    deleted_at TEXT
);

CREATE TABLE IF NOT EXISTS bazi_match_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    link_id INTEGER REFERENCES bazi_match_links(id),
    initiator_user_id INTEGER NOT NULL REFERENCES users(id),
    other_user_id INTEGER REFERENCES users(id),
    other_name TEXT NOT NULL DEFAULT '',
    chart_a_json TEXT NOT NULL DEFAULT '',
    chart_b_json TEXT NOT NULL DEFAULT '',
    match_json TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);
