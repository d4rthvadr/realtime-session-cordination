package otp

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

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
	if err := db.PingContext(ctx); err != nil {
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
	CREATE TABLE IF NOT EXISTS otp_challenges (
		id TEXT PRIMARY KEY,
		email TEXT NOT NULL,
		intent TEXT NOT NULL,
		code_hash TEXT NOT NULL,
		attempt_count INTEGER NOT NULL DEFAULT 0,
		max_attempts INTEGER NOT NULL DEFAULT 3,
		resend_count INTEGER NOT NULL DEFAULT 0,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		expires_at TEXT NOT NULL,
		used_at TEXT,
		verified_at TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_otp_challenges_email_intent_created
	ON otp_challenges(email, intent, created_at DESC);

	CREATE INDEX IF NOT EXISTS idx_otp_challenges_expires_at
	ON otp_challenges(expires_at);
	`

	_, err := s.db.Exec(query)
	return err
}

func (s *SqliteStore) Create(challenge *Challenge) (*Challenge, error) {
	if challenge == nil {
		return nil, ErrNotFound
	}

	normalizedIntent, err := NormalizeIntent(challenge.Intent)
	if err != nil {
		return nil, err
	}

	normalizedEmail := NormalizeEmail(challenge.Email)

	challengeToInsert := cloneChallenge(challenge)
	challengeToInsert.Intent = normalizedIntent
	challengeToInsert.Email = normalizedEmail

	_, err = s.db.Exec(`
		INSERT INTO otp_challenges (
			id, email, intent, code_hash, attempt_count, max_attempts,
			resend_count, created_at, updated_at, expires_at, used_at, verified_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		challengeToInsert.ID,
		challengeToInsert.Email,
		challengeToInsert.Intent,
		challengeToInsert.CodeHash,
		challengeToInsert.AttemptCount,
		challengeToInsert.MaxAttempts,
		challengeToInsert.ResendCount,
		challengeToInsert.CreatedAt.Format(time.RFC3339),
		challengeToInsert.UpdatedAt.Format(time.RFC3339),
		challengeToInsert.ExpiresAt.Format(time.RFC3339),
		timeToString(challengeToInsert.UsedAt),
		timeToString(challengeToInsert.VerifiedAt),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create otp challenge: %w", err)
	}

	return challengeToInsert, nil
}

func (s *SqliteStore) GetByID(id string) (*Challenge, error) {
	row := s.db.QueryRow(`
		SELECT id, email, intent, code_hash, attempt_count, max_attempts,
		       resend_count, created_at, updated_at, expires_at, used_at, verified_at
		FROM otp_challenges
		WHERE id = ?
	`, id)

	challenge, err := scanChallenge(row)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get otp challenge: %w", err)
	}

	return challenge, nil
}

func (s *SqliteStore) GetLatestByEmailIntent(email, intent string) (*Challenge, error) {
	normalizedIntent, err := NormalizeIntent(intent)
	if err != nil {
		return nil, err
	}
	normalizedEmail := NormalizeEmail(email)

	row := s.db.QueryRow(`
		SELECT id, email, intent, code_hash, attempt_count, max_attempts,
		       resend_count, created_at, updated_at, expires_at, used_at, verified_at
		FROM otp_challenges
		WHERE email = ? AND intent = ?
		ORDER BY created_at DESC, id DESC
		LIMIT 1
	`, normalizedEmail, normalizedIntent)

	challenge, err := scanChallenge(row)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest otp challenge: %w", err)
	}

	return challenge, nil
}

func (s *SqliteStore) IncrementAttempts(id string, updatedAt time.Time) (*Challenge, error) {
	result, err := s.db.Exec(`
		UPDATE otp_challenges
		SET attempt_count = attempt_count + 1, updated_at = ?
		WHERE id = ?
	`, updatedAt.UTC().Format(time.RFC3339), id)
	if err != nil {
		return nil, fmt.Errorf("failed to increment attempts: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to read update result: %w", err)
	}
	if affected == 0 {
		return nil, ErrNotFound
	}

	return s.GetByID(id)
}

func (s *SqliteStore) IncrementResendCount(id string, updatedAt time.Time) (*Challenge, error) {
	result, err := s.db.Exec(`
		UPDATE otp_challenges
		SET resend_count = resend_count + 1, updated_at = ?
		WHERE id = ?
	`, updatedAt.UTC().Format(time.RFC3339), id)
	if err != nil {
		return nil, fmt.Errorf("failed to increment resend count: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to read update result: %w", err)
	}
	if affected == 0 {
		return nil, ErrNotFound
	}

	return s.GetByID(id)
}

func (s *SqliteStore) MarkVerifiedAndUsed(id string, verifiedAt time.Time) (*Challenge, error) {
	verifiedAtUTC := verifiedAt.UTC().Format(time.RFC3339)

	result, err := s.db.Exec(`
		UPDATE otp_challenges
		SET verified_at = ?, used_at = ?, updated_at = ?
		WHERE id = ?
	`, verifiedAtUTC, verifiedAtUTC, verifiedAtUTC, id)
	if err != nil {
		return nil, fmt.Errorf("failed to mark challenge as used: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to read update result: %w", err)
	}
	if affected == 0 {
		return nil, ErrNotFound
	}

	return s.GetByID(id)
}

func (s *SqliteStore) DeleteExpired(now time.Time) (int, error) {
	result, err := s.db.Exec(`DELETE FROM otp_challenges WHERE expires_at < ?`, now.UTC().Format(time.RFC3339))
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired challenges: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to read delete result: %w", err)
	}

	return int(affected), nil
}

func (s *SqliteStore) Close() error {
	return s.db.Close()
}

func scanChallenge(scanner interface{ Scan(dest ...any) error }) (*Challenge, error) {
	var (
		challenge     Challenge
		createdAtRaw  string
		updatedAtRaw  string
		expiresAtRaw  string
		usedAtRaw     sql.NullString
		verifiedAtRaw sql.NullString
	)

	err := scanner.Scan(
		&challenge.ID,
		&challenge.Email,
		&challenge.Intent,
		&challenge.CodeHash,
		&challenge.AttemptCount,
		&challenge.MaxAttempts,
		&challenge.ResendCount,
		&createdAtRaw,
		&updatedAtRaw,
		&expiresAtRaw,
		&usedAtRaw,
		&verifiedAtRaw,
	)
	if err != nil {
		return nil, err
	}

	createdAt, err := time.Parse(time.RFC3339, createdAtRaw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}
	updatedAt, err := time.Parse(time.RFC3339, updatedAtRaw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse updated_at: %w", err)
	}
	expiresAt, err := time.Parse(time.RFC3339, expiresAtRaw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse expires_at: %w", err)
	}

	challenge.Email = NormalizeEmail(challenge.Email)
	challenge.CreatedAt = createdAt
	challenge.UpdatedAt = updatedAt
	challenge.ExpiresAt = expiresAt
	challenge.UsedAt = nullStringToTime(usedAtRaw)
	challenge.VerifiedAt = nullStringToTime(verifiedAtRaw)

	return &challenge, nil
}

func timeToString(value *time.Time) *string {
	if value == nil || value.IsZero() {
		return nil
	}

	v := value.UTC().Format(time.RFC3339)
	return &v
}

func nullStringToTime(value sql.NullString) *time.Time {
	if !value.Valid || value.String == "" {
		return nil
	}

	t, err := time.Parse(time.RFC3339, value.String)
	if err != nil {
		return nil
	}

	return &t
}
