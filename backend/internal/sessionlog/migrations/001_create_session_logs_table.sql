CREATE TABLE IF NOT EXISTS session_logs (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    program_item_id TEXT,
    event_type TEXT NOT NULL,
    message TEXT NOT NULL,
    metadata TEXT,
    occurred_at TEXT NOT NULL,
    request_id TEXT,
    created_at TEXT NOT NULL,
    FOREIGN KEY (session_id) REFERENCES sessions(id)
);

CREATE INDEX IF NOT EXISTS idx_session_logs_session_id
ON session_logs(session_id);

CREATE INDEX IF NOT EXISTS idx_session_logs_session_order
ON session_logs(session_id, occurred_at DESC, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_session_logs_event_type
ON session_logs(event_type);
