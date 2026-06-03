package analytics

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	OutboxStatePending    = "pending"
	OutboxStateProcessing = "processing"
	OutboxStateProcessed  = "processed"
	OutboxStateDeadLetter = "dead_letter"
)

// SqliteStore persists analytics events, outbox records, and processor checkpoints.
type SqliteStore struct {
	db *sql.DB
}

func NewSqliteStore(dbPath string) (*SqliteStore, error) {
	dsn := fmt.Sprintf("file:%s?_busy_timeout=5000&_journal_mode=WAL&_foreign_keys=on", dbPath)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite db: %w", err)
	}
	db.SetMaxOpenConns(1)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping sqlite db: %w", err)
	}

	store := &SqliteStore{db: db}
	if err = store.runMigrations(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run analytics migrations: %w", err)
	}

	return store, nil
}

func (s *SqliteStore) runMigrations() error {
	query := `
	CREATE TABLE IF NOT EXISTS analytics_events (
		id TEXT PRIMARY KEY,
		session_id TEXT NOT NULL,
		program_item_id TEXT,
		event_key TEXT NOT NULL,
		occurred_at TEXT NOT NULL,
		ingested_at TEXT NOT NULL,
		source TEXT NOT NULL CHECK(source IN ('server', 'client')),
		payload_json TEXT NOT NULL,
		created_at TEXT NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_analytics_events_session_occurred
	ON analytics_events(session_id, occurred_at DESC);

	CREATE INDEX IF NOT EXISTS idx_analytics_events_event_key
	ON analytics_events(event_key);

	CREATE TABLE IF NOT EXISTS analytics_outbox (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		event_id TEXT NOT NULL UNIQUE,
		state TEXT NOT NULL CHECK(state IN ('pending', 'processing', 'processed', 'dead_letter')),
		lease_owner TEXT,
		leased_until TEXT,
		attempt INTEGER NOT NULL DEFAULT 0,
		last_error TEXT,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		FOREIGN KEY (event_id) REFERENCES analytics_events(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_analytics_outbox_state_lease
	ON analytics_outbox(state, leased_until);

	CREATE INDEX IF NOT EXISTS idx_analytics_outbox_created
	ON analytics_outbox(created_at);

	CREATE TABLE IF NOT EXISTS analytics_checkpoints (
		worker_name TEXT PRIMARY KEY,
		last_event_id TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		FOREIGN KEY (last_event_id) REFERENCES analytics_events(id)
	);
	`

	_, err := s.db.Exec(query)
	return err
}

func (s *SqliteStore) AppendEvent(record EventRecord) error {
	if record.ID == "" {
		return fmt.Errorf("event id is required")
	}
	if record.SessionID == "" {
		return fmt.Errorf("session id is required")
	}
	if record.EventKey == "" {
		return fmt.Errorf("event key is required")
	}
	if record.Source != EventSourceServer && record.Source != EventSourceClient {
		return fmt.Errorf("event source must be 'server' or 'client'")
	}
	if len(record.PayloadJSON) == 0 {
		return fmt.Errorf("payload json is required")
	}

	occurredAt := record.OccurredAt.UTC()
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}
	ingestedAt := record.IngestedAt.UTC()
	if ingestedAt.IsZero() {
		ingestedAt = time.Now().UTC()
	}

	_, err := s.db.Exec(`
		INSERT INTO analytics_events (
			id,
			session_id,
			program_item_id,
			event_key,
			occurred_at,
			ingested_at,
			source,
			payload_json,
			created_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		record.ID,
		record.SessionID,
		nullableProgramItemID(record.ProgramItemID),
		record.EventKey,
		occurredAt.Format(time.RFC3339Nano),
		ingestedAt.Format(time.RFC3339Nano),
		string(record.Source),
		string(record.PayloadJSON),
		time.Now().UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("failed to append analytics event: %w", err)
	}

	return nil
}

func (s *SqliteStore) Enqueue(eventID string, now time.Time) error {
	if eventID == "" {
		return fmt.Errorf("event id is required")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	timestamp := now.UTC().Format(time.RFC3339Nano)

	_, err := s.db.Exec(`
		INSERT INTO analytics_outbox (
			event_id,
			state,
			lease_owner,
			leased_until,
			attempt,
			last_error,
			created_at,
			updated_at
		)
		VALUES (?, ?, NULL, NULL, 0, NULL, ?, ?)
	`, eventID, OutboxStatePending, timestamp, timestamp)
	if err != nil {
		return fmt.Errorf("failed to enqueue analytics event: %w", err)
	}

	return nil
}

func (s *SqliteStore) SaveCheckpoint(checkpoint ProcessorCheckpoint) error {
	if checkpoint.WorkerName == "" {
		return fmt.Errorf("worker name is required")
	}
	if checkpoint.LastEventID == "" {
		return fmt.Errorf("last event id is required")
	}

	updatedAt := checkpoint.UpdatedAt.UTC()
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}

	_, err := s.db.Exec(`
		INSERT INTO analytics_checkpoints (
			worker_name,
			last_event_id,
			updated_at
		)
		VALUES (?, ?, ?)
		ON CONFLICT(worker_name)
		DO UPDATE SET
			last_event_id = excluded.last_event_id,
			updated_at = excluded.updated_at
	`, checkpoint.WorkerName, checkpoint.LastEventID, updatedAt.Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("failed to save processor checkpoint: %w", err)
	}

	return nil
}

func nullableProgramItemID(id *string) any {
	if id == nil || *id == "" {
		return nil
	}
	return *id
}
