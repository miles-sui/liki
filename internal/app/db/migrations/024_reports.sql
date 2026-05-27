-- 024_reports.sql -- LLM interpretation reports with SSE streaming, persistence, and sharing.
CREATE TABLE IF NOT EXISTS reports (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id),
    scene TEXT NOT NULL DEFAULT '',
    sub_scene TEXT NOT NULL DEFAULT '',
    question TEXT NOT NULL DEFAULT '',
    engine_data TEXT NOT NULL DEFAULT '',
    content TEXT NOT NULL DEFAULT '',
    locale TEXT NOT NULL DEFAULT 'zh-CN',
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    deleted_at TEXT
);

CREATE INDEX IF NOT EXISTS idx_reports_user_scene ON reports(user_id, scene, created_at DESC);

CREATE TABLE IF NOT EXISTS report_shares (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    report_id INTEGER NOT NULL REFERENCES reports(id),
    token TEXT NOT NULL UNIQUE,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    expires_at TEXT
);

CREATE INDEX IF NOT EXISTS idx_report_shares_token ON report_shares(token);
