-- Create sessions table for persisting session data
CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    speaker_name TEXT NOT NULL,
    duration_seconds INTEGER NOT NULL,
    status TEXT NOT NULL,
    control_token TEXT NOT NULL UNIQUE,
    created_at TEXT NOT NULL
);

-- Index on control_token for faster token validation
CREATE INDEX IF NOT EXISTS idx_control_token ON sessions(control_token);

-- Index on status for quick lookup of active sessions
CREATE INDEX IF NOT EXISTS idx_status ON sessions(status);
