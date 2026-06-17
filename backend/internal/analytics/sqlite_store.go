package analytics

import (
	"context"
	"database/sql"
	"errors"
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

// SqliteStore persists analytics events, outbox records, processor checkpoints, and projections.
type SqliteStore struct {
	db *sql.DB
}

type dbExec interface {
	Exec(query string, args ...any) (sql.Result, error)
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
	statements := []string{
		`CREATE TABLE IF NOT EXISTS analytics_events (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL,
			program_item_id TEXT,
			event_key TEXT NOT NULL,
			occurred_at TEXT NOT NULL,
			ingested_at TEXT NOT NULL,
			source TEXT NOT NULL CHECK(source IN ('server', 'client')),
			payload_json TEXT NOT NULL,
			created_at TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_analytics_events_session_occurred ON analytics_events(session_id, occurred_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_analytics_events_event_key ON analytics_events(event_key)`,
		`CREATE TABLE IF NOT EXISTS analytics_outbox (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			event_id TEXT NOT NULL UNIQUE,
			state TEXT NOT NULL CHECK(state IN ('pending', 'processing', 'processed', 'dead_letter')),
			lease_owner TEXT,
			leased_until TEXT,
			next_retry_at TEXT,
			attempt INTEGER NOT NULL DEFAULT 0,
			last_error TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			FOREIGN KEY (event_id) REFERENCES analytics_events(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_analytics_outbox_state_retry ON analytics_outbox(state, next_retry_at, leased_until)`,
		`CREATE INDEX IF NOT EXISTS idx_analytics_outbox_state_lease ON analytics_outbox(state, leased_until)`,
		`CREATE INDEX IF NOT EXISTS idx_analytics_outbox_created ON analytics_outbox(created_at)`,
		`CREATE TABLE IF NOT EXISTS analytics_checkpoints (
			worker_name TEXT PRIMARY KEY,
			last_event_id TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			FOREIGN KEY (last_event_id) REFERENCES analytics_events(id)
		)`,
		`CREATE TABLE IF NOT EXISTS analytics_session_projections (
			session_id TEXT PRIMARY KEY,
			session_status TEXT NOT NULL DEFAULT '',
			session_duration_seconds INTEGER NOT NULL DEFAULT 0,
			program_item_count INTEGER NOT NULL DEFAULT 0,
			scheduled_count INTEGER NOT NULL DEFAULT 0,
			in_progress_count INTEGER NOT NULL DEFAULT 0,
			paused_count INTEGER NOT NULL DEFAULT 0,
			ended_count INTEGER NOT NULL DEFAULT 0,
			canceled_count INTEGER NOT NULL DEFAULT 0,
			planned_seconds INTEGER NOT NULL DEFAULT 0,
			effective_budget_seconds INTEGER NOT NULL DEFAULT 0,
			total_adjustment_seconds INTEGER NOT NULL DEFAULT 0,
			total_pause_seconds INTEGER NOT NULL DEFAULT 0,
			total_pause_count INTEGER NOT NULL DEFAULT 0,
			ended_on_time_count INTEGER NOT NULL DEFAULT 0,
			overrun_item_count INTEGER NOT NULL DEFAULT 0,
			total_overrun_seconds INTEGER NOT NULL DEFAULT 0,
			total_underrun_seconds INTEGER NOT NULL DEFAULT 0,
			ended_on_time_ratio REAL NOT NULL DEFAULT 0,
			computed_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS analytics_platform_projection (
			id INTEGER PRIMARY KEY CHECK(id = 1),
			total_sessions INTEGER NOT NULL DEFAULT 0,
			created_sessions INTEGER NOT NULL DEFAULT 0,
			live_sessions INTEGER NOT NULL DEFAULT 0,
			paused_sessions INTEGER NOT NULL DEFAULT 0,
			ended_sessions INTEGER NOT NULL DEFAULT 0,
			total_program_items INTEGER NOT NULL DEFAULT 0,
			ended_program_items INTEGER NOT NULL DEFAULT 0,
			on_time_ended_program_items INTEGER NOT NULL DEFAULT 0,
			overrun_program_items INTEGER NOT NULL DEFAULT 0,
			total_session_duration_secs INTEGER NOT NULL DEFAULT 0,
			total_planned_seconds INTEGER NOT NULL DEFAULT 0,
			effective_budget_seconds INTEGER NOT NULL DEFAULT 0,
			total_adjustment_seconds INTEGER NOT NULL DEFAULT 0,
			total_pause_seconds INTEGER NOT NULL DEFAULT 0,
			total_pause_count INTEGER NOT NULL DEFAULT 0,
			total_overrun_seconds INTEGER NOT NULL DEFAULT 0,
			total_underrun_seconds INTEGER NOT NULL DEFAULT 0,
			session_completion_ratio REAL NOT NULL DEFAULT 0,
			program_item_on_time_ratio REAL NOT NULL DEFAULT 0,
			computed_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
	}

	for _, stmt := range statements {
		if _, err := s.db.Exec(stmt); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}
	if err := s.ensureOutboxRetryColumn(); err != nil {
		return err
	}
	return nil
}

func (s *SqliteStore) ensureOutboxRetryColumn() error {
	var columnCount int
	if err := s.db.QueryRow(`
		SELECT COUNT(1)
		FROM pragma_table_info('analytics_outbox')
		WHERE name = ?
	`, "next_retry_at").Scan(&columnCount); err != nil {
		return fmt.Errorf("failed to inspect analytics_outbox columns: %w", err)
	}
	if columnCount > 0 {
		return nil
	}
	if _, err := s.db.Exec(`ALTER TABLE analytics_outbox ADD COLUMN next_retry_at TEXT`); err != nil {
		return fmt.Errorf("failed to add analytics_outbox.next_retry_at column: %w", err)
	}
	return nil
}

func (s *SqliteStore) AppendEvent(record EventRecord) error {
	return s.appendEvent(s.db, record)
}

func (s *SqliteStore) AppendEventAndEnqueue(record EventRecord, now time.Time) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin analytics ingestion transaction: %w", err)
	}

	if err = s.appendEvent(tx, record); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err = s.enqueue(tx, record.ID, now); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit analytics ingestion transaction: %w", err)
	}

	return nil
}

func (s *SqliteStore) appendEvent(exec dbExec, record EventRecord) error {
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

	_, err := exec.Exec(`
		INSERT INTO analytics_events (
			id, session_id, program_item_id, event_key,
			occurred_at, ingested_at, source, payload_json, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
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
	return s.enqueue(s.db, eventID, now)
}

func (s *SqliteStore) enqueue(exec dbExec, eventID string, now time.Time) error {
	if eventID == "" {
		return fmt.Errorf("event id is required")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	timestamp := now.UTC().Format(time.RFC3339Nano)

	_, err := exec.Exec(`
		INSERT INTO analytics_outbox (
			event_id, state, lease_owner, leased_until, next_retry_at, attempt, last_error, created_at, updated_at
		) VALUES (?, ?, NULL, NULL, NULL, 0, NULL, ?, ?)
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
		INSERT INTO analytics_checkpoints (worker_name, last_event_id, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(worker_name) DO UPDATE SET
			last_event_id = excluded.last_event_id,
			updated_at = excluded.updated_at
	`, checkpoint.WorkerName, checkpoint.LastEventID, updatedAt.Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("failed to save processor checkpoint: %w", err)
	}

	return nil
}

func (s *SqliteStore) ClaimPendingForProcessing(workerName string, leaseUntil time.Time, limit int, now time.Time) ([]OutboxRecord, error) {
	if workerName == "" {
		return nil, fmt.Errorf("worker name is required")
	}
	if limit <= 0 {
		limit = 50
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	if leaseUntil.IsZero() {
		leaseUntil = now.Add(15 * time.Second)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin outbox claim transaction: %w", err)
	}

	rows, err := tx.Query(`
		SELECT id, event_id, state, lease_owner, leased_until, next_retry_at, attempt, last_error, created_at, updated_at
		FROM analytics_outbox
		WHERE state = ?
		  AND (next_retry_at IS NULL OR next_retry_at <= ?)
		  AND (leased_until IS NULL OR leased_until < ?)
		ORDER BY COALESCE(next_retry_at, created_at) ASC, created_at ASC
		LIMIT ?
	`, OutboxStatePending, now.UTC().Format(time.RFC3339Nano), now.UTC().Format(time.RFC3339Nano), limit)
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("failed to query pending outbox rows: %w", err)
	}
	defer rows.Close()

	type candidate struct {
		id      int64
		eventID string
	}
	candidates := make([]candidate, 0, limit)
	for rows.Next() {
		var rec OutboxRecord
		var leaseOwnerRaw, leasedUntilRaw, nextRetryAtRaw, lastErrorRaw sql.NullString
		var createdAtRaw, updatedAtRaw string
		if err = rows.Scan(&rec.ID, &rec.EventID, &rec.State, &leaseOwnerRaw, &leasedUntilRaw, &nextRetryAtRaw, &rec.Attempt, &lastErrorRaw, &createdAtRaw, &updatedAtRaw); err != nil {
			_ = tx.Rollback()
			return nil, fmt.Errorf("failed to scan pending outbox row: %w", err)
		}
		if leaseOwnerRaw.Valid {
			rec.LeaseOwner = leaseOwnerRaw.String
		}
		if lastErrorRaw.Valid {
			rec.LastError = lastErrorRaw.String
		}
		if nextRetryAtRaw.Valid {
			ts, parseErr := time.Parse(time.RFC3339Nano, nextRetryAtRaw.String)
			if parseErr != nil {
				_ = tx.Rollback()
				return nil, fmt.Errorf("failed to parse next_retry_at for pending outbox row: %w", parseErr)
			}
			rec.NextRetryAt = &ts
		}
		createdAt, parseErr := time.Parse(time.RFC3339Nano, createdAtRaw)
		if parseErr != nil {
			_ = tx.Rollback()
			return nil, fmt.Errorf("failed to parse created_at for pending outbox row: %w", parseErr)
		}
		updatedAt, parseErr := time.Parse(time.RFC3339Nano, updatedAtRaw)
		if parseErr != nil {
			_ = tx.Rollback()
			return nil, fmt.Errorf("failed to parse updated_at for pending outbox row: %w", parseErr)
		}
		rec.CreatedAt = createdAt
		rec.UpdatedAt = updatedAt
		candidates = append(candidates, candidate{id: rec.ID, eventID: rec.EventID})
	}
	if err = rows.Err(); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("failed to iterate pending outbox rows: %w", err)
	}

	claimed := make([]OutboxRecord, 0, len(candidates))
	for _, c := range candidates {
		res, updateErr := tx.Exec(`
			UPDATE analytics_outbox
			SET state = ?, lease_owner = ?, leased_until = ?, attempt = attempt + 1, updated_at = ?
			WHERE id = ?
			  AND state = ?
			  AND (leased_until IS NULL OR leased_until < ?)
		`, OutboxStateProcessing, workerName, leaseUntil.UTC().Format(time.RFC3339Nano), now.UTC().Format(time.RFC3339Nano), c.id, OutboxStatePending, now.UTC().Format(time.RFC3339Nano))
		if updateErr != nil {
			_ = tx.Rollback()
			return nil, fmt.Errorf("failed to claim outbox row %d: %w", c.id, updateErr)
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			continue
		}

		var rec OutboxRecord
		var leaseOwnerRaw, leasedUntilRaw, nextRetryAtRaw, lastErrorRaw sql.NullString
		var createdAtRaw, updatedAtRaw string
		if err = tx.QueryRow(`
			SELECT id, event_id, state, lease_owner, leased_until, next_retry_at, attempt, last_error, created_at, updated_at
			FROM analytics_outbox WHERE id = ?
		`, c.id).Scan(&rec.ID, &rec.EventID, &rec.State, &leaseOwnerRaw, &leasedUntilRaw, &nextRetryAtRaw, &rec.Attempt, &lastErrorRaw, &createdAtRaw, &updatedAtRaw); err != nil {
			_ = tx.Rollback()
			return nil, fmt.Errorf("failed to read claimed outbox row %d: %w", c.id, err)
		}
		if leaseOwnerRaw.Valid {
			rec.LeaseOwner = leaseOwnerRaw.String
		}
		if lastErrorRaw.Valid {
			rec.LastError = lastErrorRaw.String
		}
		if nextRetryAtRaw.Valid {
			ts, parseErr := time.Parse(time.RFC3339Nano, nextRetryAtRaw.String)
			if parseErr != nil {
				_ = tx.Rollback()
				return nil, fmt.Errorf("failed to parse next_retry_at for outbox row %d: %w", c.id, parseErr)
			}
			rec.NextRetryAt = &ts
		}
		createdAt, parseErr := time.Parse(time.RFC3339Nano, createdAtRaw)
		if parseErr != nil {
			_ = tx.Rollback()
			return nil, fmt.Errorf("failed to parse created_at for claimed outbox row %d: %w", c.id, parseErr)
		}
		updatedAt, parseErr := time.Parse(time.RFC3339Nano, updatedAtRaw)
		if parseErr != nil {
			_ = tx.Rollback()
			return nil, fmt.Errorf("failed to parse updated_at for claimed outbox row %d: %w", c.id, parseErr)
		}
		rec.CreatedAt = createdAt
		rec.UpdatedAt = updatedAt
		if leasedUntilRaw.Valid {
			ts, parseErr := time.Parse(time.RFC3339Nano, leasedUntilRaw.String)
			if parseErr != nil {
				_ = tx.Rollback()
				return nil, fmt.Errorf("failed to parse leased_until for outbox row %d: %w", c.id, parseErr)
			}
			rec.LeasedUntil = &ts
		}
		claimed = append(claimed, rec)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit outbox claim transaction: %w", err)
	}

	return claimed, nil
}

func (s *SqliteStore) GetEvent(eventID string) (EventRecord, error) {
	if eventID == "" {
		return EventRecord{}, fmt.Errorf("event id is required")
	}

	var rec EventRecord
	var occurredAtRaw, ingestedAtRaw, sourceRaw string
	var programItemIDRaw sql.NullString
	if err := s.db.QueryRow(`
		SELECT id, session_id, program_item_id, event_key, occurred_at, ingested_at, source, payload_json
		FROM analytics_events WHERE id = ?
	`, eventID).Scan(&rec.ID, &rec.SessionID, &programItemIDRaw, &rec.EventKey, &occurredAtRaw, &ingestedAtRaw, &sourceRaw, &rec.PayloadJSON); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return EventRecord{}, fmt.Errorf("analytics event not found: %w", err)
		}
		return EventRecord{}, fmt.Errorf("failed to load analytics event: %w", err)
	}

	occurredAt, err := time.Parse(time.RFC3339Nano, occurredAtRaw)
	if err != nil {
		return EventRecord{}, fmt.Errorf("failed to parse occurred_at: %w", err)
	}
	ingestedAt, err := time.Parse(time.RFC3339Nano, ingestedAtRaw)
	if err != nil {
		return EventRecord{}, fmt.Errorf("failed to parse ingested_at: %w", err)
	}

	rec.OccurredAt = occurredAt
	rec.IngestedAt = ingestedAt
	rec.Source = EventSource(sourceRaw)
	if programItemIDRaw.Valid {
		pid := programItemIDRaw.String
		rec.ProgramItemID = &pid
	}

	return rec, nil
}

func (s *SqliteStore) MarkProcessed(outboxID int64, now time.Time) error {
	if outboxID <= 0 {
		return fmt.Errorf("outbox id is required")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}

	_, err := s.db.Exec(`
		UPDATE analytics_outbox
		SET state = ?, lease_owner = NULL, leased_until = NULL, next_retry_at = NULL, last_error = NULL, updated_at = ?
		WHERE id = ?
	`, OutboxStateProcessed, now.UTC().Format(time.RFC3339Nano), outboxID)
	if err != nil {
		return fmt.Errorf("failed to mark outbox row processed: %w", err)
	}

	return nil
}

func (s *SqliteStore) MarkFailed(outboxID int64, lastError string, deadLetter bool, nextRetryAt *time.Time, now time.Time) error {
	if outboxID <= 0 {
		return fmt.Errorf("outbox id is required")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}

	nextState := OutboxStatePending
	var retryAtValue any = nil
	if deadLetter {
		nextState = OutboxStateDeadLetter
	} else if nextRetryAt != nil {
		retryAtValue = nextRetryAt.UTC().Format(time.RFC3339Nano)
	}

	_, err := s.db.Exec(`
		UPDATE analytics_outbox
		SET state = ?, lease_owner = NULL, leased_until = NULL, next_retry_at = ?, last_error = ?, updated_at = ?
		WHERE id = ?
	`, nextState, retryAtValue, lastError, now.UTC().Format(time.RFC3339Nano), outboxID)
	if err != nil {
		return fmt.Errorf("failed to mark outbox row failed: %w", err)
	}

	return nil
}

func (s *SqliteStore) LoadCheckpoint(workerName string) (ProcessorCheckpoint, bool, error) {
	if workerName == "" {
		return ProcessorCheckpoint{}, false, fmt.Errorf("worker name is required")
	}

	var cp ProcessorCheckpoint
	var updatedAtRaw string
	err := s.db.QueryRow(`
		SELECT worker_name, last_event_id, updated_at
		FROM analytics_checkpoints WHERE worker_name = ?
	`, workerName).Scan(&cp.WorkerName, &cp.LastEventID, &updatedAtRaw)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ProcessorCheckpoint{}, false, nil
		}
		return ProcessorCheckpoint{}, false, fmt.Errorf("failed to load checkpoint: %w", err)
	}

	updatedAt, parseErr := time.Parse(time.RFC3339Nano, updatedAtRaw)
	if parseErr != nil {
		return ProcessorCheckpoint{}, false, fmt.Errorf("failed to parse checkpoint timestamp: %w", parseErr)
	}
	cp.UpdatedAt = updatedAt

	return cp, true, nil
}

func (s *SqliteStore) GetFreshness(workerName string, now time.Time) (ProcessorFreshness, error) {
	if workerName == "" {
		return ProcessorFreshness{}, fmt.Errorf("worker name is required")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}

	freshness := ProcessorFreshness{WorkerName: workerName}

	cp, ok, err := s.LoadCheckpoint(workerName)
	if err != nil {
		return ProcessorFreshness{}, err
	}
	if ok {
		freshness.LastEventID = cp.LastEventID
		ts := cp.UpdatedAt.UTC()
		freshness.LastProcessedAt = &ts
	}

	var oldestPendingRaw sql.NullString
	if err = s.db.QueryRow(`
		SELECT COUNT(*), MIN(e.occurred_at)
		FROM analytics_outbox o
		JOIN analytics_events e ON e.id = o.event_id
		WHERE o.state = ?
	`, OutboxStatePending).Scan(&freshness.PendingCount, &oldestPendingRaw); err != nil {
		return ProcessorFreshness{}, fmt.Errorf("failed to compute freshness pending stats: %w", err)
	}

	if oldestPendingRaw.Valid {
		ts, parseErr := time.Parse(time.RFC3339Nano, oldestPendingRaw.String)
		if parseErr != nil {
			return ProcessorFreshness{}, fmt.Errorf("failed to parse oldest pending timestamp: %w", parseErr)
		}
		freshness.OldestPendingAt = &ts
	}

	return freshness, nil
}

func (s *SqliteStore) ListDeadLetters(limit int, offset int) ([]DeadLetterRecord, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := s.db.Query(`
		SELECT o.id, o.event_id, e.session_id, e.program_item_id, e.event_key,
			e.occurred_at, e.ingested_at, o.attempt, o.last_error, o.updated_at, e.payload_json
		FROM analytics_outbox o
		JOIN analytics_events e ON e.id = o.event_id
		WHERE o.state = ?
		ORDER BY o.updated_at DESC, o.id DESC
		LIMIT ? OFFSET ?
	`, OutboxStateDeadLetter, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list dead-letter rows: %w", err)
	}
	defer rows.Close()

	results := make([]DeadLetterRecord, 0, limit)
	for rows.Next() {
		var rec DeadLetterRecord
		var programItemIDRaw, lastErrorRaw sql.NullString
		var occurredAtRaw, ingestedAtRaw, failedAtRaw string
		if err = rows.Scan(
			&rec.OutboxID,
			&rec.EventID,
			&rec.SessionID,
			&programItemIDRaw,
			&rec.EventKey,
			&occurredAtRaw,
			&ingestedAtRaw,
			&rec.Attempt,
			&lastErrorRaw,
			&failedAtRaw,
			&rec.PayloadJSON,
		); err != nil {
			return nil, fmt.Errorf("failed to scan dead-letter row: %w", err)
		}
		if programItemIDRaw.Valid {
			pid := programItemIDRaw.String
			rec.ProgramItemID = &pid
		}
		if lastErrorRaw.Valid {
			rec.LastError = lastErrorRaw.String
		}
		rec.OccurredAt, err = time.Parse(time.RFC3339Nano, occurredAtRaw)
		if err != nil {
			return nil, fmt.Errorf("failed to parse dead-letter occurred_at: %w", err)
		}
		rec.IngestedAt, err = time.Parse(time.RFC3339Nano, ingestedAtRaw)
		if err != nil {
			return nil, fmt.Errorf("failed to parse dead-letter ingested_at: %w", err)
		}
		rec.FailedAt, err = time.Parse(time.RFC3339Nano, failedAtRaw)
		if err != nil {
			return nil, fmt.Errorf("failed to parse dead-letter failed_at: %w", err)
		}
		results = append(results, rec)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate dead-letter rows: %w", err)
	}

	return results, nil
}

func (s *SqliteStore) GetDeadLetter(outboxID int64) (DeadLetterRecord, bool, error) {
	if outboxID <= 0 {
		return DeadLetterRecord{}, false, fmt.Errorf("outbox id is required")
	}

	var rec DeadLetterRecord
	var programItemIDRaw, lastErrorRaw sql.NullString
	var occurredAtRaw, ingestedAtRaw, failedAtRaw string
	err := s.db.QueryRow(`
		SELECT o.id, o.event_id, e.session_id, e.program_item_id, e.event_key,
			e.occurred_at, e.ingested_at, o.attempt, o.last_error, o.updated_at, e.payload_json
		FROM analytics_outbox o
		JOIN analytics_events e ON e.id = o.event_id
		WHERE o.id = ? AND o.state = ?
	`, outboxID, OutboxStateDeadLetter).Scan(
		&rec.OutboxID,
		&rec.EventID,
		&rec.SessionID,
		&programItemIDRaw,
		&rec.EventKey,
		&occurredAtRaw,
		&ingestedAtRaw,
		&rec.Attempt,
		&lastErrorRaw,
		&failedAtRaw,
		&rec.PayloadJSON,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return DeadLetterRecord{}, false, nil
		}
		return DeadLetterRecord{}, false, fmt.Errorf("failed to load dead-letter row: %w", err)
	}

	if programItemIDRaw.Valid {
		pid := programItemIDRaw.String
		rec.ProgramItemID = &pid
	}
	if lastErrorRaw.Valid {
		rec.LastError = lastErrorRaw.String
	}
	rec.OccurredAt, err = time.Parse(time.RFC3339Nano, occurredAtRaw)
	if err != nil {
		return DeadLetterRecord{}, false, fmt.Errorf("failed to parse dead-letter occurred_at: %w", err)
	}
	rec.IngestedAt, err = time.Parse(time.RFC3339Nano, ingestedAtRaw)
	if err != nil {
		return DeadLetterRecord{}, false, fmt.Errorf("failed to parse dead-letter ingested_at: %w", err)
	}
	rec.FailedAt, err = time.Parse(time.RFC3339Nano, failedAtRaw)
	if err != nil {
		return DeadLetterRecord{}, false, fmt.Errorf("failed to parse dead-letter failed_at: %w", err)
	}

	return rec, true, nil
}

func (s *SqliteStore) RetryDeadLetter(outboxID int64, now time.Time) error {
	if outboxID <= 0 {
		return fmt.Errorf("outbox id is required")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}

	res, err := s.db.Exec(`
		UPDATE analytics_outbox
		SET state = ?, lease_owner = NULL, leased_until = NULL, next_retry_at = NULL, attempt = 0, last_error = NULL, updated_at = ?
		WHERE id = ? AND state = ?
	`, OutboxStatePending, now.UTC().Format(time.RFC3339Nano), outboxID, OutboxStateDeadLetter)
	if err != nil {
		return fmt.Errorf("failed to retry dead-letter row: %w", err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("dead-letter row not found")
	}

	return nil
}

func (s *SqliteStore) UpsertSessionProjection(p SessionProjection) error {
	if p.SessionID == "" {
		return fmt.Errorf("session id is required")
	}
	now := time.Now().UTC()
	if p.ComputedAt.IsZero() {
		p.ComputedAt = now
	}
	if p.UpdatedAt.IsZero() {
		p.UpdatedAt = now
	}
	_, err := s.db.Exec(`
		INSERT INTO analytics_session_projections (
			session_id, session_status, session_duration_seconds,
			program_item_count, scheduled_count, in_progress_count, paused_count,
			ended_count, canceled_count, planned_seconds, effective_budget_seconds,
			total_adjustment_seconds, total_pause_seconds, total_pause_count,
			ended_on_time_count, overrun_item_count, total_overrun_seconds,
			total_underrun_seconds, ended_on_time_ratio, computed_at, updated_at
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(session_id) DO UPDATE SET
			session_status = excluded.session_status,
			session_duration_seconds = excluded.session_duration_seconds,
			program_item_count = excluded.program_item_count,
			scheduled_count = excluded.scheduled_count,
			in_progress_count = excluded.in_progress_count,
			paused_count = excluded.paused_count,
			ended_count = excluded.ended_count,
			canceled_count = excluded.canceled_count,
			planned_seconds = excluded.planned_seconds,
			effective_budget_seconds = excluded.effective_budget_seconds,
			total_adjustment_seconds = excluded.total_adjustment_seconds,
			total_pause_seconds = excluded.total_pause_seconds,
			total_pause_count = excluded.total_pause_count,
			ended_on_time_count = excluded.ended_on_time_count,
			overrun_item_count = excluded.overrun_item_count,
			total_overrun_seconds = excluded.total_overrun_seconds,
			total_underrun_seconds = excluded.total_underrun_seconds,
			ended_on_time_ratio = excluded.ended_on_time_ratio,
			computed_at = excluded.computed_at,
			updated_at = excluded.updated_at
	`,
		p.SessionID, p.SessionStatus, p.SessionDurationSeconds,
		p.ProgramItemCount, p.ScheduledCount, p.InProgressCount, p.PausedCount,
		p.EndedCount, p.CanceledCount, p.PlannedSeconds, p.EffectiveBudgetSeconds,
		p.TotalAdjustmentSeconds, p.TotalPauseSeconds, p.TotalPauseCount,
		p.EndedOnTimeCount, p.OverrunItemCount, p.TotalOverrunSeconds,
		p.TotalUnderrunSeconds, p.EndedOnTimeRatio,
		p.ComputedAt.UTC().Format(time.RFC3339Nano),
		p.UpdatedAt.UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("failed to upsert session projection: %w", err)
	}
	return nil
}

func (s *SqliteStore) GetSessionProjection(sessionID string) (SessionProjection, bool, error) {
	if sessionID == "" {
		return SessionProjection{}, false, fmt.Errorf("session id is required")
	}
	var p SessionProjection
	var computedAtRaw, updatedAtRaw string
	err := s.db.QueryRow(`
		SELECT session_id, session_status, session_duration_seconds,
			program_item_count, scheduled_count, in_progress_count, paused_count,
			ended_count, canceled_count, planned_seconds, effective_budget_seconds,
			total_adjustment_seconds, total_pause_seconds, total_pause_count,
			ended_on_time_count, overrun_item_count, total_overrun_seconds,
			total_underrun_seconds, ended_on_time_ratio, computed_at, updated_at
		FROM analytics_session_projections WHERE session_id = ?
	`, sessionID).Scan(
		&p.SessionID, &p.SessionStatus, &p.SessionDurationSeconds,
		&p.ProgramItemCount, &p.ScheduledCount, &p.InProgressCount, &p.PausedCount,
		&p.EndedCount, &p.CanceledCount, &p.PlannedSeconds, &p.EffectiveBudgetSeconds,
		&p.TotalAdjustmentSeconds, &p.TotalPauseSeconds, &p.TotalPauseCount,
		&p.EndedOnTimeCount, &p.OverrunItemCount, &p.TotalOverrunSeconds,
		&p.TotalUnderrunSeconds, &p.EndedOnTimeRatio, &computedAtRaw, &updatedAtRaw,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return SessionProjection{}, false, nil
		}
		return SessionProjection{}, false, fmt.Errorf("failed to get session projection: %w", err)
	}
	computedAt, err := time.Parse(time.RFC3339Nano, computedAtRaw)
	if err != nil {
		return SessionProjection{}, false, fmt.Errorf("failed to parse computed_at: %w", err)
	}
	updatedAt, err := time.Parse(time.RFC3339Nano, updatedAtRaw)
	if err != nil {
		return SessionProjection{}, false, fmt.Errorf("failed to parse updated_at: %w", err)
	}
	p.ComputedAt = computedAt
	p.UpdatedAt = updatedAt
	return p, true, nil
}

func (s *SqliteStore) UpsertPlatformProjection(p PlatformProjection) error {
	now := time.Now().UTC()
	if p.ComputedAt.IsZero() {
		p.ComputedAt = now
	}
	if p.UpdatedAt.IsZero() {
		p.UpdatedAt = now
	}
	_, err := s.db.Exec(`
		INSERT INTO analytics_platform_projection (
			id, total_sessions, created_sessions, live_sessions, paused_sessions, ended_sessions,
			total_program_items, ended_program_items, on_time_ended_program_items, overrun_program_items,
			total_session_duration_secs, total_planned_seconds, effective_budget_seconds,
			total_adjustment_seconds, total_pause_seconds, total_pause_count,
			total_overrun_seconds, total_underrun_seconds,
			session_completion_ratio, program_item_on_time_ratio, computed_at, updated_at
		) VALUES (1,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(id) DO UPDATE SET
			total_sessions = excluded.total_sessions,
			created_sessions = excluded.created_sessions,
			live_sessions = excluded.live_sessions,
			paused_sessions = excluded.paused_sessions,
			ended_sessions = excluded.ended_sessions,
			total_program_items = excluded.total_program_items,
			ended_program_items = excluded.ended_program_items,
			on_time_ended_program_items = excluded.on_time_ended_program_items,
			overrun_program_items = excluded.overrun_program_items,
			total_session_duration_secs = excluded.total_session_duration_secs,
			total_planned_seconds = excluded.total_planned_seconds,
			effective_budget_seconds = excluded.effective_budget_seconds,
			total_adjustment_seconds = excluded.total_adjustment_seconds,
			total_pause_seconds = excluded.total_pause_seconds,
			total_pause_count = excluded.total_pause_count,
			total_overrun_seconds = excluded.total_overrun_seconds,
			total_underrun_seconds = excluded.total_underrun_seconds,
			session_completion_ratio = excluded.session_completion_ratio,
			program_item_on_time_ratio = excluded.program_item_on_time_ratio,
			computed_at = excluded.computed_at,
			updated_at = excluded.updated_at
	`,
		p.TotalSessions, p.CreatedSessions, p.LiveSessions, p.PausedSessions, p.EndedSessions,
		p.TotalProgramItems, p.EndedProgramItems, p.OnTimeEndedProgramItems, p.OverrunProgramItems,
		p.TotalSessionDurationSecs, p.TotalPlannedSeconds, p.EffectiveBudgetSeconds,
		p.TotalAdjustmentSeconds, p.TotalPauseSeconds, p.TotalPauseCount,
		p.TotalOverrunSeconds, p.TotalUnderrunSeconds,
		p.SessionCompletionRatio, p.ProgramItemOnTimeRatio,
		p.ComputedAt.UTC().Format(time.RFC3339Nano),
		p.UpdatedAt.UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("failed to upsert platform projection: %w", err)
	}
	return nil
}

func (s *SqliteStore) GetPlatformProjection() (PlatformProjection, bool, error) {
	var p PlatformProjection
	var computedAtRaw, updatedAtRaw string
	err := s.db.QueryRow(`
		SELECT total_sessions, created_sessions, live_sessions, paused_sessions, ended_sessions,
			total_program_items, ended_program_items, on_time_ended_program_items, overrun_program_items,
			total_session_duration_secs, total_planned_seconds, effective_budget_seconds,
			total_adjustment_seconds, total_pause_seconds, total_pause_count,
			total_overrun_seconds, total_underrun_seconds,
			session_completion_ratio, program_item_on_time_ratio, computed_at, updated_at
		FROM analytics_platform_projection WHERE id = 1
	`).Scan(
		&p.TotalSessions, &p.CreatedSessions, &p.LiveSessions, &p.PausedSessions, &p.EndedSessions,
		&p.TotalProgramItems, &p.EndedProgramItems, &p.OnTimeEndedProgramItems, &p.OverrunProgramItems,
		&p.TotalSessionDurationSecs, &p.TotalPlannedSeconds, &p.EffectiveBudgetSeconds,
		&p.TotalAdjustmentSeconds, &p.TotalPauseSeconds, &p.TotalPauseCount,
		&p.TotalOverrunSeconds, &p.TotalUnderrunSeconds,
		&p.SessionCompletionRatio, &p.ProgramItemOnTimeRatio, &computedAtRaw, &updatedAtRaw,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return PlatformProjection{}, false, nil
		}
		return PlatformProjection{}, false, fmt.Errorf("failed to get platform projection: %w", err)
	}
	computedAt, err := time.Parse(time.RFC3339Nano, computedAtRaw)
	if err != nil {
		return PlatformProjection{}, false, fmt.Errorf("failed to parse computed_at: %w", err)
	}
	updatedAt, err := time.Parse(time.RFC3339Nano, updatedAtRaw)
	if err != nil {
		return PlatformProjection{}, false, fmt.Errorf("failed to parse updated_at: %w", err)
	}
	p.ComputedAt = computedAt
	p.UpdatedAt = updatedAt
	return p, true, nil
}

func nullableProgramItemID(id *string) any {
	if id == nil || *id == "" {
		return nil
	}
	return *id
}
