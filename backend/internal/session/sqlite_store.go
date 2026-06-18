package session

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// SqliteStore implements Store interface using SQLite for persistence
type SqliteStore struct {
	db *sql.DB
}

// NewSqliteStore creates a new SQLite store and runs migrations
func NewSqliteStore(dbPath string) (*SqliteStore, error) {
	dsn := fmt.Sprintf("file:%s?_busy_timeout=5000&_journal_mode=WAL&_foreign_keys=on", dbPath)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite db: %w", err)
	}
	db.SetMaxOpenConns(1)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
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
		control_token TEXT NOT NULL UNIQUE,
		created_by TEXT,
		created_at TEXT NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_control_token ON sessions(control_token);
	CREATE INDEX IF NOT EXISTS idx_status ON sessions(status);
	`

	if _, err := s.db.Exec(query); err != nil {
		return err
	}

	if err := s.ensureLegacyRuntimeColumnsDropped(); err != nil {
		return err
	}

	if err := s.ensureCreatedByColumn(); err != nil {
		return err
	}

	_, err := s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_sessions_created_by ON sessions(created_by)`)
	return err
}

func (s *SqliteStore) ensureLegacyRuntimeColumnsDropped() error {
	rows, err := s.db.Query("PRAGMA table_info(sessions)")
	if err != nil {
		return fmt.Errorf("failed to inspect sessions schema: %w", err)
	}
	defer rows.Close()

	legacy := map[string]bool{
		"started_at":                    false,
		"paused_at":                     false,
		"total_paused_duration_seconds": false,
		"adjustment_seconds":            false,
		"ended_remaining_seconds":       false,
	}

	for rows.Next() {
		var cid int
		var name string
		var ctype string
		var notnull int
		var dflt sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return fmt.Errorf("failed to read sessions schema row: %w", err)
		}
		if _, ok := legacy[name]; ok {
			legacy[name] = true
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("failed while reading sessions schema: %w", err)
	}

	hasLegacy := false
	for _, present := range legacy {
		if present {
			hasLegacy = true
			break
		}
	}
	if !hasLegacy {
		return nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin session schema migration: %w", err)
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	stmts := []string{
		`DROP INDEX IF EXISTS idx_control_token`,
		`DROP INDEX IF EXISTS idx_status`,
		`CREATE TABLE sessions_new (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			speaker_name TEXT NOT NULL,
			duration_seconds INTEGER NOT NULL,
			status TEXT NOT NULL,
			control_token TEXT NOT NULL UNIQUE,
			created_by TEXT,
			created_at TEXT NOT NULL
		)`,
		`INSERT INTO sessions_new (id, title, speaker_name, duration_seconds, status, control_token, created_by, created_at)
		 SELECT id, title, speaker_name, duration_seconds, status, control_token, NULL, created_at FROM sessions`,
		`DROP TABLE sessions`,
		`ALTER TABLE sessions_new RENAME TO sessions`,
		`CREATE INDEX IF NOT EXISTS idx_control_token ON sessions(control_token)`,
		`CREATE INDEX IF NOT EXISTS idx_status ON sessions(status)`,
	}

	for _, stmt := range stmts {
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("failed during session schema migration (%s): %w", strings.Split(stmt, "\n")[0], err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit session schema migration: %w", err)
	}
	tx = nil

	return nil
}

func (s *SqliteStore) ensureCreatedByColumn() error {
	var columnCount int
	if err := s.db.QueryRow(`
		SELECT COUNT(1)
		FROM pragma_table_info('sessions')
		WHERE name = ?
	`, "created_by").Scan(&columnCount); err != nil {
		return fmt.Errorf("failed to inspect sessions columns: %w", err)
	}
	if columnCount > 0 {
		return nil
	}
	if _, err := s.db.Exec(`ALTER TABLE sessions ADD COLUMN created_by TEXT`); err != nil {
		return fmt.Errorf("failed to add sessions.created_by column: %w", err)
	}
	return nil
}

// Create persists a new session to the database
func (s *SqliteStore) Create(session *Session) (*Session, error) {
	_, err := s.db.Exec(`
		INSERT INTO sessions (
			id, title, speaker_name, duration_seconds, status,
			control_token, created_by, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`,
		session.ID,
		session.Title,
		session.SpeakerName,
		session.DurationSeconds,
		session.Status,
		session.ControlToken,
		session.CreatedBy,
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
		       control_token, created_by, created_at
		FROM sessions WHERE id = ?
	`, id)

	var session Session
	var createdBy sql.NullString
	var createdAtStr string

	err := row.Scan(
		&session.ID,
		&session.Title,
		&session.SpeakerName,
		&session.DurationSeconds,
		&session.Status,
		&session.ControlToken,
		&createdBy,
		&createdAtStr,
	)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	createdAt, err := time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}
	session.CreatedBy = nullStringPtr(createdBy)
	session.CreatedAt = createdAt

	return &session, nil
}

// GetForUser retrieves a session by ID when visible to the given user.
func (s *SqliteStore) GetForUser(id, userID string, isAdmin bool) (*Session, error) {
	if isAdmin {
		return s.Get(id)
	}

	row := s.db.QueryRow(`
		SELECT id, title, speaker_name, duration_seconds, status,
		       control_token, created_by, created_at
		FROM sessions
		WHERE id = ? AND created_by = ?
	`, id, userID)

	var session Session
	var createdBy sql.NullString
	var createdAtStr string

	err := row.Scan(
		&session.ID,
		&session.Title,
		&session.SpeakerName,
		&session.DurationSeconds,
		&session.Status,
		&session.ControlToken,
		&createdBy,
		&createdAtStr,
	)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	createdAt, err := time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}
	session.CreatedBy = nullStringPtr(createdBy)
	session.CreatedAt = createdAt

	return &session, nil
}

// List retrieves all sessions ordered by creation time descending.
func (s *SqliteStore) List() ([]*Session, error) {
	rows, err := s.db.Query(`
		SELECT id, title, speaker_name, duration_seconds, status,
		       control_token, created_by, created_at
		FROM sessions
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	defer rows.Close()

	return scanSessions(rows)
}

// ListForUser retrieves sessions visible to the given user.
// Admin users can see all sessions. Non-admin users only see sessions they created.
func (s *SqliteStore) ListForUser(userID string, isAdmin bool) ([]*Session, error) {
	if isAdmin {
		return s.List()
	}

	rows, err := s.db.Query(`
		SELECT id, title, speaker_name, duration_seconds, status,
		       control_token, created_by, created_at
		FROM sessions
		WHERE created_by = ?
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	defer rows.Close()

	return scanSessions(rows)
}

func scanSessions(rows *sql.Rows) ([]*Session, error) {
	sessions := make([]*Session, 0)
	for rows.Next() {
		var session Session
		var createdBy sql.NullString
		var createdAtStr string

		err := rows.Scan(
			&session.ID,
			&session.Title,
			&session.SpeakerName,
			&session.DurationSeconds,
			&session.Status,
			&session.ControlToken,
			&createdBy,
			&createdAtStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session row: %w", err)
		}

		createdAt, parseErr := time.Parse(time.RFC3339, createdAtStr)
		if parseErr != nil {
			return nil, fmt.Errorf("failed to parse created_at: %w", parseErr)
		}
		session.CreatedBy = nullStringPtr(createdBy)
		session.CreatedAt = createdAt

		sessions = append(sessions, &session)
	}

	if rowErr := rows.Err(); rowErr != nil {
		return nil, fmt.Errorf("failed while iterating sessions: %w", rowErr)
	}

	return sessions, nil
}

// Update persists changes to an existing session
func (s *SqliteStore) Update(session *Session) error {
	result, err := s.db.Exec(`
		UPDATE sessions SET
			title = ?,
			speaker_name = ?,
			duration_seconds = ?,
			status = ?
		WHERE id = ?
	`,
		session.Title,
		session.SpeakerName,
		session.DurationSeconds,
		session.Status,
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

func nullStringPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	v := ns.String
	return &v
}
