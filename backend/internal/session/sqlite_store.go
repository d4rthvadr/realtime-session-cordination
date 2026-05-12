package session

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// SqliteStore implements Store interface using SQLite for persistence
type SqliteStore struct {
	db *sql.DB
}

// NewSqliteStore creates a new SQLite store and runs migrations
func NewSqliteStore(dbPath string) (*SqliteStore, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite db: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping sqlite db: %w", err)
	}

	store := &SqliteStore{db: db}

	// Run migrations
	if err := store.runMigrations(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return store, nil
}

// runMigrations executes the schema migration
func (s *SqliteStore) runMigrations() error {
	// Create sessions table with indices
	query := `
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		speaker_name TEXT NOT NULL,
		duration_seconds INTEGER NOT NULL,
		status TEXT NOT NULL,
		started_at TEXT,
		paused_at TEXT,
		total_paused_duration_seconds INTEGER DEFAULT 0,
		adjustment_seconds INTEGER DEFAULT 0,
		ended_remaining_seconds INTEGER,
		control_token TEXT NOT NULL UNIQUE,
		created_at TEXT NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_control_token ON sessions(control_token);
	CREATE INDEX IF NOT EXISTS idx_status ON sessions(status);
	`

	_, err := s.db.Exec(query)
	return err
}

// Create persists a new session to the database
func (s *SqliteStore) Create(session *Session) (*Session, error) {
	_, err := s.db.Exec(`
		INSERT INTO sessions (
			id, title, speaker_name, duration_seconds, status,
			started_at, paused_at, total_paused_duration_seconds,
			adjustment_seconds, ended_remaining_seconds,
			control_token, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		session.ID,
		session.Title,
		session.SpeakerName,
		session.DurationSeconds,
		session.Status,
		timeToString(session.StartedAt),
		timeToString(session.PausedAt),
		session.TotalPausedDurationSeconds,
		session.AdjustmentSeconds,
		session.EndedRemainingSeconds,
		session.ControlToken,
		session.CreatedAt.Format(time.RFC3339),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// Get retrieves a session by ID
func (s *SqliteStore) Get(id string) (*Session, error) {
	row := s.db.QueryRow(`
		SELECT id, title, speaker_name, duration_seconds, status,
		       started_at, paused_at, total_paused_duration_seconds,
		       adjustment_seconds, ended_remaining_seconds,
		       control_token, created_at
		FROM sessions WHERE id = ?
	`, id)

	var session Session
	var startedAtStr, pausedAtStr, createdAtStr string

	err := row.Scan(
		&session.ID,
		&session.Title,
		&session.SpeakerName,
		&session.DurationSeconds,
		&session.Status,
		&startedAtStr,
		&pausedAtStr,
		&session.TotalPausedDurationSeconds,
		&session.AdjustmentSeconds,
		&session.EndedRemainingSeconds,
		&session.ControlToken,
		&createdAtStr,
	)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Parse timestamps
	session.StartedAt = stringToTime(&startedAtStr)
	session.PausedAt = stringToTime(&pausedAtStr)

	createdAt, err := time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}
	session.CreatedAt = createdAt

	return &session, nil
}

// Update persists changes to an existing session
func (s *SqliteStore) Update(session *Session) error {
	result, err := s.db.Exec(`
		UPDATE sessions SET
			title = ?,
			speaker_name = ?,
			duration_seconds = ?,
			status = ?,
			started_at = ?,
			paused_at = ?,
			total_paused_duration_seconds = ?,
			adjustment_seconds = ?,
			ended_remaining_seconds = ?
		WHERE id = ?
	`,
		session.Title,
		session.SpeakerName,
		session.DurationSeconds,
		session.Status,
		timeToString(session.StartedAt),
		timeToString(session.PausedAt),
		session.TotalPausedDurationSeconds,
		session.AdjustmentSeconds,
		session.EndedRemainingSeconds,
		session.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// SessionExists checks if a session exists by ID
func (s *SqliteStore) SessionExists(id string) bool {
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM sessions WHERE id = ?)", id).Scan(&exists)
	return err == nil && exists
}

// ValidateControlToken verifies that a control token matches the session's token
func (s *SqliteStore) ValidateControlToken(id, token string) error {
	var storedToken string
	err := s.db.QueryRow("SELECT control_token FROM sessions WHERE id = ?", id).Scan(&storedToken)

	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("failed to validate token: %w", err)
	}

	if storedToken != token {
		return ErrUnauthorized
	}

	return nil
}

// Close closes the database connection
func (s *SqliteStore) Close() error {
	return s.db.Close()
}

// Helper functions for timestamp conversion

// timeToString converts a time.Time to RFC3339 string, or empty string if zero
func timeToString(t *time.Time) *string {
	if t == nil || t.IsZero() {
		return nil
	}
	str := t.Format(time.RFC3339)
	return &str
}

// stringToTime converts a string back to time.Time pointer, handling nil/empty strings
func stringToTime(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, *s)
	if err != nil {
		return nil
	}
	return &t
}
