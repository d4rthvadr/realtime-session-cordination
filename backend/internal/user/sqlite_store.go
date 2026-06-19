package user

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
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
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		name TEXT,
		email TEXT,
		type TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'user',
		email_verified_at TEXT,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		deleted_at TEXT,
		is_visible INTEGER NOT NULL DEFAULT 1,
		avatar_url TEXT,
		bio TEXT,
		is_active INTEGER NOT NULL DEFAULT 1
	);

	CREATE INDEX IF NOT EXISTS idx_users_type ON users(type);
	CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);
	CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);
	`

	if _, err := s.db.Exec(query); err != nil {
		return err
	}

	if err := s.ensureRoleColumn(); err != nil {
		return err
	}

	if err := s.ensureEmailColumns(); err != nil {
		return err
	}

	if err := s.normalizeLegacyRoleValues(); err != nil {
		return err
	}

	if err := s.normalizeLegacyUserTypeValues(); err != nil {
		return err
	}

	_, err := s.db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_unique ON users(email);
	`)
	return err
}

func (s *SqliteStore) ensureRoleColumn() error {
	var columnCount int
	if err := s.db.QueryRow(`
		SELECT COUNT(1)
		FROM pragma_table_info('users')
		WHERE name = ?
	`, "role").Scan(&columnCount); err != nil {
		return fmt.Errorf("failed to inspect users columns: %w", err)
	}
	if columnCount > 0 {
		return nil
	}
	if _, err := s.db.Exec(`ALTER TABLE users ADD COLUMN role TEXT NOT NULL DEFAULT 'user'`); err != nil {
		return fmt.Errorf("failed to add users.role column: %w", err)
	}
	return nil
}

func (s *SqliteStore) normalizeLegacyRoleValues() error {
	if _, err := s.db.Exec(`UPDATE users SET role = 'user' WHERE role = 'normal' OR role = ''`); err != nil {
		return fmt.Errorf("failed to normalize users.role values: %w", err)
	}
	return nil
}

func (s *SqliteStore) normalizeLegacyUserTypeValues() error {
	if _, err := s.db.Exec(`UPDATE users SET type = 'user' WHERE type = 'normal' OR type = ''`); err != nil {
		return fmt.Errorf("failed to normalize users.type values: %w", err)
	}
	return nil
}

func (s *SqliteStore) ensureEmailColumns() error {
	var emailColumnCount int
	if err := s.db.QueryRow(`
		SELECT COUNT(1)
		FROM pragma_table_info('users')
		WHERE name = ?
	`, "email").Scan(&emailColumnCount); err != nil {
		return fmt.Errorf("failed to inspect users columns: %w", err)
	}
	if emailColumnCount == 0 {
		if _, err := s.db.Exec(`ALTER TABLE users ADD COLUMN email TEXT`); err != nil {
			return fmt.Errorf("failed to add users.email column: %w", err)
		}
	}

	var verifiedColumnCount int
	if err := s.db.QueryRow(`
		SELECT COUNT(1)
		FROM pragma_table_info('users')
		WHERE name = ?
	`, "email_verified_at").Scan(&verifiedColumnCount); err != nil {
		return fmt.Errorf("failed to inspect users columns: %w", err)
	}
	if verifiedColumnCount == 0 {
		if _, err := s.db.Exec(`ALTER TABLE users ADD COLUMN email_verified_at TEXT`); err != nil {
			return fmt.Errorf("failed to add users.email_verified_at column: %w", err)
		}
	}

	return nil
}

func (s *SqliteStore) Create(user *User) (*User, error) {
	if user.Role == "" {
		user.Role = RoleUser
	}

	normalizedEmail := normalizeEmailForSQL(user.Email)
	user.Email = normalizedStringPtr(normalizedEmail)

	_, err := s.db.Exec(`
		INSERT INTO users (
			id, name, email, type, role, email_verified_at, created_at, updated_at,
			deleted_at, is_visible, avatar_url, bio, is_active
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		user.ID,
		user.Name,
		user.Email,
		user.Type,
		user.Role,
		timeToString(user.EmailVerifiedAt),
		user.CreatedAt.Format(time.RFC3339),
		user.UpdatedAt.Format(time.RFC3339),
		timeToString(user.DeletedAt),
		boolToInt(user.IsVisible),
		user.AvatarURL,
		user.Bio,
		boolToInt(user.IsActive),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *SqliteStore) GetByID(id string) (*User, error) {
	row := s.db.QueryRow(`
		SELECT id, name, email, type, role, email_verified_at, created_at, updated_at, deleted_at,
		       is_visible, avatar_url, bio, is_active
		FROM users WHERE id = ?
	`, id)

	var user User
	var name sql.NullString
	var email sql.NullString
	var emailVerifiedAt sql.NullString
	var deletedAt sql.NullString
	var avatarURL sql.NullString
	var bio sql.NullString
	var createdAtStr string
	var updatedAtStr string
	var isVisible int
	var isActive int

	err := row.Scan(
		&user.ID,
		&name,
		&email,
		&user.Type,
		&user.Role,
		&emailVerifiedAt,
		&createdAtStr,
		&updatedAtStr,
		&deletedAt,
		&isVisible,
		&avatarURL,
		&bio,
		&isActive,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	user.Name = nullStringPtr(name)
	user.Email = nullStringPtr(email)
	if user.Role == "" {
		user.Role = RoleUser
	}
	if user.Role == "normal" {
		user.Role = RoleUser
	}
	user.EmailVerifiedAt = nullStringToTime(emailVerifiedAt)
	user.DeletedAt = nullStringToTime(deletedAt)
	user.AvatarURL = nullStringPtr(avatarURL)
	user.Bio = nullStringPtr(bio)
	user.IsVisible = isVisible == 1
	user.IsActive = isActive == 1

	createdAt, err := time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}
	updatedAt, err := time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse updated_at: %w", err)
	}
	user.CreatedAt = createdAt
	user.UpdatedAt = updatedAt

	return &user, nil
}

func (s *SqliteStore) GetByEmail(email string) (*User, error) {
	normalizedEmail := normalizeEmailForSQL(&email)
	if normalizedEmail == nil {
		return nil, ErrNotFound
	}

	row := s.db.QueryRow(`
		SELECT id, name, email, type, role, email_verified_at, created_at, updated_at, deleted_at,
		       is_visible, avatar_url, bio, is_active
		FROM users WHERE email = ?
	`, *normalizedEmail)

	var user User
	var name sql.NullString
	var userEmail sql.NullString
	var emailVerifiedAt sql.NullString
	var deletedAt sql.NullString
	var avatarURL sql.NullString
	var bio sql.NullString
	var createdAtStr string
	var updatedAtStr string
	var isVisible int
	var isActive int

	err := row.Scan(
		&user.ID,
		&name,
		&userEmail,
		&user.Type,
		&user.Role,
		&emailVerifiedAt,
		&createdAtStr,
		&updatedAtStr,
		&deletedAt,
		&isVisible,
		&avatarURL,
		&bio,
		&isActive,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	user.Name = nullStringPtr(name)
	user.Email = nullStringPtr(userEmail)
	if user.Role == "" {
		user.Role = RoleUser
	}
	if user.Role == "normal" {
		user.Role = RoleUser
	}
	user.EmailVerifiedAt = nullStringToTime(emailVerifiedAt)
	user.DeletedAt = nullStringToTime(deletedAt)
	user.AvatarURL = nullStringPtr(avatarURL)
	user.Bio = nullStringPtr(bio)
	user.IsVisible = isVisible == 1
	user.IsActive = isActive == 1

	createdAt, err := time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}
	updatedAt, err := time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse updated_at: %w", err)
	}
	user.CreatedAt = createdAt
	user.UpdatedAt = updatedAt

	return &user, nil
}

func (s *SqliteStore) Close() error {
	return s.db.Close()
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func timeToString(t *time.Time) *string {
	if t == nil || t.IsZero() {
		return nil
	}
	str := t.Format(time.RFC3339)
	return &str
}

func normalizedStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	v := *value
	return &v
}

func normalizeEmailForSQL(email *string) *string {
	if email == nil {
		return nil
	}

	normalized := strings.ToLower(strings.TrimSpace(*email))
	if normalized == "" {
		return nil
	}

	return &normalized
}

func nullStringPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	v := ns.String
	return &v
}

func nullStringToTime(ns sql.NullString) *time.Time {
	if !ns.Valid || ns.String == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, ns.String)
	if err != nil {
		return nil
	}
	return &t
}
