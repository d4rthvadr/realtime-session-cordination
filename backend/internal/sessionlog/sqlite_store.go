package sessionlog

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type SqliteStore struct {
	db *sql.DB
}

func NewSqliteStore(dbPath string) (*SqliteStore, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite db: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping sqlite db: %w", err)
	}

	store := &SqliteStore{db: db}
	if err = store.runMigrations(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return store, nil
}

func (s *SqliteStore) runMigrations() error {
	query := `
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
	`

	_, err := s.db.Exec(query)
	return err
}

func (s *SqliteStore) Append(entry *Entry) (*Entry, error) {
	metadataJSON, err := metadataToJSON(entry.Metadata)
	if err != nil {
		return nil, err
	}

	_, err = s.db.Exec(`
		INSERT INTO session_logs (
			id,
			session_id,
			program_item_id,
			event_type,
			message,
			metadata,
			occurred_at,
			request_id,
			created_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		entry.ID,
		entry.SessionID,
		stringPtrToSQL(entry.ProgramItemID),
		entry.EventType.String(),
		entry.Message,
		metadataJSON,
		entry.OccurredAt.Format(time.RFC3339Nano),
		nullableString(entry.RequestID),
		entry.CreatedAt.Format(time.RFC3339Nano),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to append session log: %w", err)
	}

	return cloneEntry(entry), nil
}

func (s *SqliteStore) ListBySession(sessionID string, options ListOptions) ([]*Entry, error) {
	rows, err := s.db.Query(`
		SELECT
			id,
			session_id,
			program_item_id,
			event_type,
			message,
			metadata,
			occurred_at,
			request_id,
			created_at
		FROM session_logs
		WHERE session_id = ?
		ORDER BY occurred_at DESC, created_at DESC, id DESC
		LIMIT ? OFFSET ?
	`, sessionID, options.Limit, options.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list session logs: %w", err)
	}
	defer rows.Close()

	entries := make([]*Entry, 0)
	for rows.Next() {
		entry, scanErr := scanEntry(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		entries = append(entries, entry)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed while iterating session logs: %w", err)
	}

	return entries, nil
}

func scanEntry(scanner interface{ Scan(dest ...any) error }) (*Entry, error) {
	var (
		id            string
		sessionID     string
		programItemID sql.NullString
		eventType     string
		message       string
		metadata      sql.NullString
		occurredAtRaw string
		requestID     sql.NullString
		createdAtRaw  string
	)

	if err := scanner.Scan(
		&id,
		&sessionID,
		&programItemID,
		&eventType,
		&message,
		&metadata,
		&occurredAtRaw,
		&requestID,
		&createdAtRaw,
	); err != nil {
		return nil, fmt.Errorf("failed to scan session log row: %w", err)
	}

	occurredAt, err := parseTimeRFC3339Nano(occurredAtRaw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse occurred_at: %w", err)
	}

	createdAt, err := parseTimeRFC3339Nano(createdAtRaw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}

	entry := &Entry{
		ID:         id,
		SessionID:  sessionID,
		EventType:  EventType(eventType),
		Message:    message,
		OccurredAt: occurredAt,
		CreatedAt:  createdAt,
	}

	if programItemID.Valid {
		copied := programItemID.String
		entry.ProgramItemID = &copied
	}

	if requestID.Valid {
		entry.RequestID = requestID.String
	}

	if metadata.Valid && metadata.String != "" {
		parsed, parseErr := metadataFromJSON(metadata.String)
		if parseErr != nil {
			return nil, parseErr
		}
		entry.Metadata = parsed
	}

	return entry, nil
}

func parseTimeRFC3339Nano(value string) (time.Time, error) {
	ts, err := time.Parse(time.RFC3339Nano, value)
	if err == nil {
		return ts, nil
	}

	legacy, legacyErr := time.Parse(time.RFC3339, value)
	if legacyErr != nil {
		return time.Time{}, err
	}

	return legacy, nil
}

func metadataToJSON(metadata map[string]any) (sql.NullString, error) {
	if len(metadata) == 0 {
		return sql.NullString{}, nil
	}

	bytes, err := json.Marshal(metadata)
	if err != nil {
		return sql.NullString{}, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	return sql.NullString{String: string(bytes), Valid: true}, nil
}

func metadataFromJSON(value string) (map[string]any, error) {
	if value == "" {
		return nil, nil
	}

	var metadata map[string]any
	if err := json.Unmarshal([]byte(value), &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return metadata, nil
}

func stringPtrToSQL(value *string) sql.NullString {
	if value == nil {
		return sql.NullString{}
	}
	if *value == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: *value, Valid: true}
}

func nullableString(value string) sql.NullString {
	if value == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: value, Valid: true}
}
