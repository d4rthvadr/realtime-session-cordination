CREATE TABLE IF NOT EXISTS program_items (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    title TEXT NOT NULL,
    type TEXT NOT NULL,
    status TEXT NOT NULL,
    host_name TEXT,
    scheduled_start TEXT NOT NULL,
    scheduled_end TEXT NOT NULL,
    expected_duration_minutes INTEGER NOT NULL,
    position INTEGER NOT NULL,
    location TEXT,
    metadata TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (session_id) REFERENCES sessions(id),
    UNIQUE(session_id, position)
);

CREATE INDEX IF NOT EXISTS idx_program_items_session_id ON program_items(session_id);
CREATE INDEX IF NOT EXISTS idx_program_items_schedule ON program_items(session_id, scheduled_start, scheduled_end);
