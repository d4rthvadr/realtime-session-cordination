package session

import (
	"sort"
	"sync"
)

// MemoryStore is an in-memory implementation of the Store interface.
// It uses a map protected by RWMutex, matching the current behavior.
type MemoryStore struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

// NewMemoryStore creates a new in-memory session store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		sessions: make(map[string]*Session),
	}
}

// Create persists a new session in memory.
func (ms *MemoryStore) Create(session *Session) (*Session, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.sessions[session.ID] = session
	return session, nil
}

// Get retrieves a session by ID from memory.
func (ms *MemoryStore) Get(id string) (*Session, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	s, ok := ms.sessions[id]
	if !ok {
		return nil, ErrNotFound
	}
	return s, nil
}

// List retrieves all sessions from memory ordered by newest first.
func (ms *MemoryStore) List() ([]*Session, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	result := make([]*Session, 0, len(ms.sessions))
	for _, s := range ms.sessions {
		result = append(result, s)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})

	return result, nil
}

// Update persists changes to a session in memory.
func (ms *MemoryStore) Update(session *Session) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if _, ok := ms.sessions[session.ID]; !ok {
		return ErrNotFound
	}
	ms.sessions[session.ID] = session
	return nil
}

// SessionExists checks if a session exists in memory.
func (ms *MemoryStore) SessionExists(id string) bool {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	_, ok := ms.sessions[id]
	return ok
}

// ValidateControlToken verifies the control token for a session.
func (ms *MemoryStore) ValidateControlToken(id, token string) error {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	s, ok := ms.sessions[id]
	if !ok {
		return ErrNotFound
	}
	if token == "" || s.ControlToken != token {
		return ErrUnauthorized
	}
	return nil
}
