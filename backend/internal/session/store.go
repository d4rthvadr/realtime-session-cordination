package session

// Store defines the persistence interface for sessions.
// Implementations can be in-memory, SQLite, Postgres, etc.
type Store interface {
	// Create persists a new session and returns the created session.
	Create(session *Session) (*Session, error)

	// Get retrieves a session by ID.
	Get(id string) (*Session, error)

	// List retrieves all sessions.
	List() ([]*Session, error)

	// Update persists changes to an existing session.
	Update(session *Session) error

	// SessionExists checks if a session with the given ID exists.
	SessionExists(id string) bool

	// ValidateControlToken verifies that the token matches the session's control token.
	// Returns ErrNotFound if session doesn't exist, ErrUnauthorized if token is invalid.
	ValidateControlToken(id, token string) error
}
