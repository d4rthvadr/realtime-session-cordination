package programitem

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// SqliteStore persists ProgramItems using SQLite.
type SqliteStore struct {
	db *sql.DB
	// writeMu serializes mutating operations so reorder/update checks stay stable.
	writeMu sync.Mutex
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
	CREATE UNIQUE INDEX IF NOT EXISTS idx_program_items_single_in_progress
	ON program_items(session_id)
	WHERE status = 'in_progress';
	`

	_, err := s.db.Exec(query)
	return err
}

func (s *SqliteStore) Create(item *ProgramItem) (*ProgramItem, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	// Create runs inside one transaction so overlap and position checks see a stable view.
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin create transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	metadataJSON, err := metadataToJSON(item.Metadata)
	if err != nil {
		return nil, err
	}

	overlap, err := txHasOverlap(tx, item.SessionID, item.ScheduledStart, item.ScheduledEnd, "")
	if err != nil {
		return nil, err
	}
	if overlap {
		return nil, ErrOverlap
	}

	dupPos, err := txPositionExists(tx, item.SessionID, item.Position, "")
	if err != nil {
		return nil, err
	}
	if dupPos {
		return nil, ErrDuplicatePosition
	}

	_, err = tx.Exec(`
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

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit create transaction: %w", err)
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
	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	// Update validates the edited time range and position before committing the row change.
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin update transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	metadataJSON, err := metadataToJSON(item.Metadata)
	if err != nil {
		return err
	}

	if item.Status == StatusScheduled {
		overlap, overlapErr := txHasOverlap(tx, item.SessionID, item.ScheduledStart, item.ScheduledEnd, item.ID)
		if overlapErr != nil {
			return overlapErr
		}
		if overlap {
			return ErrOverlap
		}
	}

	dupPos, posErr := txPositionExists(tx, item.SessionID, item.Position, item.ID)
	if posErr != nil {
		return posErr
	}
	if dupPos {
		return ErrDuplicatePosition
	}

	result, err := tx.Exec(`
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

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit update transaction: %w", err)
	}

	return nil
}

func (s *SqliteStore) TransitionToInProgress(id string, at time.Time) (*ProgramItem, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transition transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	current, err := txGetProgramItem(tx, id)
	if err != nil {
		return nil, err
	}
	if current.Status != StatusScheduled {
		return nil, ErrInvalidStatusTransition
	}

	var hasInProgress bool
	err = tx.QueryRow(`
		SELECT EXISTS(
			SELECT 1
			FROM program_items
			WHERE session_id = ?
			  AND status = ?
			  AND id != ?
		)
	`, current.SessionID, StatusInProgress, current.ID).Scan(&hasInProgress)
	if err != nil {
		return nil, fmt.Errorf("failed to check in-progress item: %w", err)
	}
	if hasInProgress {
		return nil, ErrInProgressExists
	}

	now := at.UTC().Format(time.RFC3339)
	result, err := tx.Exec(`
		UPDATE program_items
		SET status = ?, updated_at = ?
		WHERE id = ? AND status = ?
	`, StatusInProgress, now, id, StatusScheduled)
	if err != nil {
		return nil, fmt.Errorf("failed to transition program item to in_progress: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}
	if affected == 0 {
		return nil, ErrInvalidStatusTransition
	}

	updated, err := txGetProgramItem(tx, id)
	if err != nil {
		return nil, err
	}
	atUTC := at.UTC()
	updated.ActualStart = &atUTC
	updated.ActualEnd = nil

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transition transaction: %w", err)
	}

	return updated, nil
}

func (s *SqliteStore) TransitionToEnded(id string, at time.Time) (*ProgramItem, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transition transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	current, err := txGetProgramItem(tx, id)
	if err != nil {
		return nil, err
	}
	if current.Status != StatusInProgress {
		return nil, ErrInvalidStatusTransition
	}

	now := at.UTC().Format(time.RFC3339)
	result, err := tx.Exec(`
		UPDATE program_items
		SET status = ?, updated_at = ?
		WHERE id = ? AND status = ?
	`, StatusEnded, now, id, StatusInProgress)
	if err != nil {
		return nil, fmt.Errorf("failed to transition program item to ended: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}
	if affected == 0 {
		return nil, ErrInvalidStatusTransition
	}

	updated, err := txGetProgramItem(tx, id)
	if err != nil {
		return nil, err
	}
	atUTC := at.UTC()
	if updated.ActualStart == nil {
		updated.ActualStart = &atUTC
	}
	updated.ActualEnd = &atUTC

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transition transaction: %w", err)
	}

	return updated, nil
}

func (s *SqliteStore) Reorder(sessionID string, positions map[string]int) error {
	if len(positions) == 0 {
		return nil
	}
	// Reorder is serialized because it temporarily moves rows out of the unique position range.
	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin reorder transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	ids := make([]string, 0, len(positions))
	targetPositions := make([]int, 0, len(positions))
	seen := make(map[int]struct{}, len(positions))
	for id, position := range positions {
		ids = append(ids, id)
		targetPositions = append(targetPositions, position)
		if _, ok := seen[position]; ok {
			return ErrDuplicatePosition
		}
		seen[position] = struct{}{}
	}

	existingRows, err := tx.Query(`
		SELECT id FROM program_items WHERE session_id = ?
	`, sessionID)
	if err != nil {
		return fmt.Errorf("failed to validate reorder ids: %w", err)
	}
	defer existingRows.Close()

	existing := make(map[string]struct{})
	for existingRows.Next() {
		var id string
		if scanErr := existingRows.Scan(&id); scanErr != nil {
			return fmt.Errorf("failed to scan reorder ids: %w", scanErr)
		}
		existing[id] = struct{}{}
	}
	if err = existingRows.Err(); err != nil {
		return fmt.Errorf("failed to iterate reorder ids: %w", err)
	}

	for _, id := range ids {
		if _, ok := existing[id]; !ok {
			return ErrNotFound
		}
	}

	occupiedRows, err := tx.Query(`
		SELECT id, position
		FROM program_items
		WHERE session_id = ?
	`, sessionID)
	if err != nil {
		return fmt.Errorf("failed to validate target positions: %w", err)
	}
	defer occupiedRows.Close()

	for occupiedRows.Next() {
		var id string
		var position int
		if scanErr := occupiedRows.Scan(&id, &position); scanErr != nil {
			return fmt.Errorf("failed to scan target positions: %w", scanErr)
		}
		if _, isMoving := positions[id]; isMoving {
			continue
		}
		if _, clashes := seen[position]; clashes {
			return ErrDuplicatePosition
		}
	}
	if err = occupiedRows.Err(); err != nil {
		return fmt.Errorf("failed while checking target positions: %w", err)
	}

	// First move every affected row out of the way so final position writes cannot clash.
	now := time.Now().UTC().Format(time.RFC3339)
	for idx, id := range ids {
		tempPosition := -1000000 - idx
		_, execErr := tx.Exec(`
			UPDATE program_items
			SET position = ?, updated_at = ?
			WHERE id = ? AND session_id = ?
		`, tempPosition, now, id, sessionID)
		if execErr != nil {
			return fmt.Errorf("failed to set temporary reorder position: %w", execErr)
		}
	}

	// Then write the requested final positions once the temporary gap is in place.
	for _, id := range ids {
		position := positions[id]
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
	return txHasOverlap(s.db, sessionID, start, end, excludeID)
}

func (s *SqliteStore) PositionExists(sessionID string, position int, excludeID string) (bool, error) {
	return txPositionExists(s.db, sessionID, position, excludeID)
}

func (s *SqliteStore) HasInProgressItem(sessionID string, excludeID string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM program_items
			WHERE session_id = ?
			  AND status = ?
	`
	args := []any{sessionID, StatusInProgress}
	if excludeID != "" {
		query += " AND id != ?"
		args = append(args, excludeID)
	}
	query += ")"

	var exists bool
	err := s.db.QueryRow(query, args...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check in-progress item: %w", err)
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

type queryRowScanner interface {
	QueryRow(query string, args ...any) *sql.Row
}

func txGetProgramItem(q queryRowScanner, id string) (*ProgramItem, error) {
	row := q.QueryRow(`
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

type queryRower interface {
	QueryRow(query string, args ...any) *sql.Row
}

func txHasOverlap(q queryRower, sessionID string, start, end time.Time, excludeID string) (bool, error) {
	// The overlap predicate is the same one used in the PRD and Phase 1 API contract.
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
	err := q.QueryRow(query, args...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check overlap: %w", err)
	}

	return exists, nil
}

func txPositionExists(q queryRower, sessionID string, position int, excludeID string) (bool, error) {
	// Position is enforced per session so the timeline remains a strict ordered sequence.
	query := `SELECT EXISTS(SELECT 1 FROM program_items WHERE session_id = ? AND position = ?`
	args := []any{sessionID, position}
	if excludeID != "" {
		query += " AND id != ?"
		args = append(args, excludeID)
	}
	query += ")"

	var exists bool
	err := q.QueryRow(query, args...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check duplicate position: %w", err)
	}

	return exists, nil
}
