package session

// Store defines the persistence interface for sessions.
// Implementations can be in-memory, SQLite, Postgres, etc.
type Store interface {
	// Create persists a new session and returns the created session.
	Create(session *Session) (*Session, error)

	// Get retrieves a session by ID.
	Get(id string) (*Session, error)

	// GetForUser retrieves a session by ID when visible to the given user.
	// Admin users can access any session. Non-admin users can only access sessions they created.
	GetForUser(id, userID string, isAdmin bool) (*Session, error)

	// List retrieves all sessions.
	List() ([]*Session, error)

	// ListForUser retrieves sessions visible to the given user.
	// Admin users can see all sessions. Non-admin users only see sessions they created.
	ListForUser(userID string, isAdmin bool) ([]*Session, error)

	// Update persists changes to an existing session.
	Update(session *Session) error

	// SessionExists checks if a session with the given ID exists.
	SessionExists(id string) bool

	// ValidateControlToken verifies that the token matches the session's control token.
	// Returns ErrNotFound if session doesn't exist, ErrUnauthorized if token is invalid.
	ValidateControlToken(id, token string) error
}
