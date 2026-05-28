package programitem

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// SqliteStore persists ProgramItems using SQLite.
type SqliteStore struct {
	db *sql.DB
}

func NewSqliteStore(dbPath string) (*SqliteStore, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping sqlite db: %w", err)
	}

	store := &SqliteStore{db: db}
	if err := store.runMigrations(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return store, nil
}

func (s *SqliteStore) runMigrations() error {
	query := `
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
	`

	_, err := s.db.Exec(query)
	return err
}

func (s *SqliteStore) Create(item *ProgramItem) (*ProgramItem, error) {
	metadataJSON, err := metadataToJSON(item.Metadata)
	if err != nil {
		return nil, err
	}

	_, err = s.db.Exec(`
		INSERT INTO program_items (
			id, session_id, title, type, status, host_name,
			scheduled_start, scheduled_end, expected_duration_minutes,
			position, location, metadata, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		item.ID,
		item.SessionID,
		item.Title,
		item.Type,
		item.Status,
		item.HostName,
		item.ScheduledStart.Format(time.RFC3339),
		item.ScheduledEnd.Format(time.RFC3339),
		item.ExpectedDurationMinutes,
		item.Position,
		item.Location,
		metadataJSON,
		item.CreatedAt.Format(time.RFC3339),
		item.UpdatedAt.Format(time.RFC3339),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create program item: %w", err)
	}

	return cloneItem(item), nil
}

func (s *SqliteStore) Get(id string) (*ProgramItem, error) {
	row := s.db.QueryRow(`
		SELECT id, session_id, title, type, status, host_name,
		       scheduled_start, scheduled_end, expected_duration_minutes,
		       position, location, metadata, created_at, updated_at
		FROM program_items
		WHERE id = ?
	`, id)

	item, err := scanProgramItem(row)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get program item: %w", err)
	}

	return item, nil
}

func (s *SqliteStore) ListBySession(sessionID string) ([]*ProgramItem, error) {
	rows, err := s.db.Query(`
		SELECT id, session_id, title, type, status, host_name,
		       scheduled_start, scheduled_end, expected_duration_minutes,
		       position, location, metadata, created_at, updated_at
		FROM program_items
		WHERE session_id = ?
		ORDER BY position ASC, created_at ASC
	`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to list program items: %w", err)
	}
	defer rows.Close()

	items := make([]*ProgramItem, 0)
	for rows.Next() {
		item, scanErr := scanProgramItem(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan program item: %w", scanErr)
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed while iterating program items: %w", err)
	}

	return items, nil
}

func (s *SqliteStore) Update(item *ProgramItem) error {
	metadataJSON, err := metadataToJSON(item.Metadata)
	if err != nil {
		return err
	}

	result, err := s.db.Exec(`
		UPDATE program_items
		SET title = ?,
		    type = ?,
		    status = ?,
		    host_name = ?,
		    scheduled_start = ?,
		    scheduled_end = ?,
		    expected_duration_minutes = ?,
		    position = ?,
		    location = ?,
		    metadata = ?,
		    updated_at = ?
		WHERE id = ?
	`,
		item.Title,
		item.Type,
		item.Status,
		item.HostName,
		item.ScheduledStart.Format(time.RFC3339),
		item.ScheduledEnd.Format(time.RFC3339),
		item.ExpectedDurationMinutes,
		item.Position,
		item.Location,
		metadataJSON,
		item.UpdatedAt.Format(time.RFC3339),
		item.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update program item: %w", err)
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

func (s *SqliteStore) Reorder(sessionID string, positions map[string]int) error {
	if len(positions) == 0 {
		return nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin reorder transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	now := time.Now().UTC().Format(time.RFC3339)
	for id, position := range positions {
		res, execErr := tx.Exec(`
			UPDATE program_items
			SET position = ?, updated_at = ?
			WHERE id = ? AND session_id = ?
		`, position, now, id, sessionID)
		if execErr != nil {
			return fmt.Errorf("failed to reorder program items: %w", execErr)
		}
		affected, affectedErr := res.RowsAffected()
		if affectedErr != nil {
			return fmt.Errorf("failed to get rows affected: %w", affectedErr)
		}
		if affected == 0 {
			return ErrNotFound
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit reorder transaction: %w", err)
	}

	return nil
}

func (s *SqliteStore) HasOverlap(sessionID string, start, end time.Time, excludeID string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM program_items
			WHERE session_id = ?
			  AND status = ?
			  AND scheduled_start < ?
			  AND scheduled_end > ?
	`

	args := []any{sessionID, StatusScheduled, end.Format(time.RFC3339), start.Format(time.RFC3339)}
	if excludeID != "" {
		query += " AND id != ?"
		args = append(args, excludeID)
	}
	query += ")"

	var exists bool
	err := s.db.QueryRow(query, args...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check overlap: %w", err)
	}

	return exists, nil
}

func (s *SqliteStore) PositionExists(sessionID string, position int, excludeID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM program_items WHERE session_id = ? AND position = ?`
	args := []any{sessionID, position}
	if excludeID != "" {
		query += " AND id != ?"
		args = append(args, excludeID)
	}
	query += ")"

	var exists bool
	err := s.db.QueryRow(query, args...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check duplicate position: %w", err)
	}

	return exists, nil
}

func (s *SqliteStore) SessionExists(sessionID string) bool {
	var exists bool
	err := s.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM sessions WHERE id = ?)`, sessionID).Scan(&exists)
	return err == nil && exists
}

func (s *SqliteStore) Close() error {
	return s.db.Close()
}

type scanner interface {
	Scan(dest ...any) error
}

func scanProgramItem(s scanner) (*ProgramItem, error) {
	var item ProgramItem
	var metadataRaw sql.NullString
	var scheduledStartRaw string
	var scheduledEndRaw string
	var createdAtRaw string
	var updatedAtRaw string

	err := s.Scan(
		&item.ID,
		&item.SessionID,
		&item.Title,
		&item.Type,
		&item.Status,
		&item.HostName,
		&scheduledStartRaw,
		&scheduledEndRaw,
		&item.ExpectedDurationMinutes,
		&item.Position,
		&item.Location,
		&metadataRaw,
		&createdAtRaw,
		&updatedAtRaw,
	)
	if err != nil {
		return nil, err
	}

	if item.ScheduledStart, err = time.Parse(time.RFC3339, scheduledStartRaw); err != nil {
		return nil, fmt.Errorf("failed to parse scheduled_start: %w", err)
	}
	if item.ScheduledEnd, err = time.Parse(time.RFC3339, scheduledEndRaw); err != nil {
		return nil, fmt.Errorf("failed to parse scheduled_end: %w", err)
	}
	if item.CreatedAt, err = time.Parse(time.RFC3339, createdAtRaw); err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}
	if item.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtRaw); err != nil {
		return nil, fmt.Errorf("failed to parse updated_at: %w", err)
	}

	if metadataRaw.Valid && strings.TrimSpace(metadataRaw.String) != "" {
		if err = json.Unmarshal([]byte(metadataRaw.String), &item.Metadata); err != nil {
			return nil, fmt.Errorf("failed to parse metadata: %w", err)
		}
	}

	return &item, nil
}

func metadataToJSON(metadata map[string]any) (string, error) {
	if len(metadata) == 0 {
		return "", nil
	}

	// Sort map keys for deterministic snapshot diffs and predictable tests.
	keys := make([]string, 0, len(metadata))
	for key := range metadata {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	normalized := make(map[string]any, len(metadata))
	for _, key := range keys {
		normalized[key] = metadata[key]
	}

	b, err := json.Marshal(normalized)
	if err != nil {
		return "", fmt.Errorf("failed to marshal metadata: %w", err)
	}

	return string(b), nil
}
