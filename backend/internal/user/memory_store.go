package user

import "sync"

type MemoryStore struct {
	mu    sync.RWMutex
	users map[string]*User
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		users: make(map[string]*User),
	}
}

func (ms *MemoryStore) Create(user *User) (*User, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if user.Role == "" {
		user.Role = RoleUser
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
