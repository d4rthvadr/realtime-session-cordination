package sessionlog

// Store defines append-only persistence operations for session logs.
type Store interface {
	Append(entry *Entry) (*Entry, error)
	ListBySession(sessionID string, options ListOptions) ([]*Entry, error)
}
