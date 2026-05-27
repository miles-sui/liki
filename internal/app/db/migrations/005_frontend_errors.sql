CREATE TABLE IF NOT EXISTS frontend_errors (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    message    TEXT NOT NULL DEFAULT '',
    filename   TEXT NOT NULL DEFAULT '',
    lineno     INTEGER NOT NULL DEFAULT 0,
    colno      INTEGER NOT NULL DEFAULT 0,
    stack      TEXT NOT NULL DEFAULT '',
    url         TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE INDEX IF NOT EXISTS idx_frontend_errors_created_at ON frontend_errors(created_at);
