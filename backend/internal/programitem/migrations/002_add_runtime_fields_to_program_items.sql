-- SQLite only supports adding one column per ALTER TABLE statement.
-- This migration is intentionally split so existing SQLite databases can be upgraded safely.
-- When the project migrates to PostgreSQL, equivalent schema updates can be consolidated
-- into a single ALTER TABLE ... ADD COLUMN ... statement for cleaner DDL.
ALTER TABLE program_items ADD COLUMN runtime_duration_seconds INTEGER NOT NULL DEFAULT 0;
ALTER TABLE program_items ADD COLUMN actual_start TEXT;
ALTER TABLE program_items ADD COLUMN paused_at TEXT;
ALTER TABLE program_items ADD COLUMN total_paused_duration_seconds INTEGER NOT NULL DEFAULT 0;
ALTER TABLE program_items ADD COLUMN adjustment_seconds INTEGER NOT NULL DEFAULT 0;
ALTER TABLE program_items ADD COLUMN ended_remaining_seconds INTEGER;
ALTER TABLE program_items ADD COLUMN actual_end TEXT;
ALTER TABLE program_items ADD COLUMN pause_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE program_items ADD COLUMN ended_reason TEXT;
