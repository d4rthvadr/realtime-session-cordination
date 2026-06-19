package user

import (
	"strings"
	"sync"
)

type MemoryStore struct {
	mu         sync.RWMutex
	users      map[string]*User
	emailIndex map[string]string
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		users:      make(map[string]*User),
		emailIndex: make(map[string]string),
	}
}

func (ms *MemoryStore) Create(user *User) (*User, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if user.Role == "" {
		user.Role = RoleUser
	}

	if normalizedEmail, ok := normalizeEmail(user.Email); ok {
		ms.emailIndex[normalizedEmail] = user.ID
	}

	ms.users[user.ID] = user
	return user, nil
}

func (ms *MemoryStore) GetByID(id string) (*User, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	u, ok := ms.users[id]
	if !ok {
		return nil, ErrNotFound
	}
	return u, nil
}

func (ms *MemoryStore) GetByEmail(email string) (*User, error) {
	normalizedEmail, ok := normalizeEmail(&email)
	if !ok {
		return nil, ErrNotFound
	}

	ms.mu.RLock()
	defer ms.mu.RUnlock()

	userID, ok := ms.emailIndex[normalizedEmail]
	if !ok {
		return nil, ErrNotFound
	}

	u, ok := ms.users[userID]
	if !ok {
		return nil, ErrNotFound
	}

	return u, nil
}

func normalizeEmail(email *string) (string, bool) {
	if email == nil {
		return "", false
	}

	normalized := strings.ToLower(strings.TrimSpace(*email))
	if normalized == "" {
		return "", false
	}

	return normalized, true
}
