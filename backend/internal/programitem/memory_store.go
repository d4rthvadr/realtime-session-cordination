package programitem

import (
	"sort"
	"sync"
	"time"
)

type SessionExistsFunc func(sessionID string) bool

// MemoryStore is an in-memory ProgramItem store implementation.
type MemoryStore struct {
	mu              sync.RWMutex
	items           map[string]*ProgramItem
	sessionExistsFn SessionExistsFunc
}

func NewMemoryStore(sessionExistsFn SessionExistsFunc) *MemoryStore {
	return &MemoryStore{
		items:           make(map[string]*ProgramItem),
		sessionExistsFn: sessionExistsFn,
	}
}

func (ms *MemoryStore) Create(item *ProgramItem) (*ProgramItem, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	cloned := cloneItem(item)
	ms.items[item.ID] = cloned
	return cloneItem(cloned), nil
}

func (ms *MemoryStore) Get(id string) (*ProgramItem, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	item, ok := ms.items[id]
	if !ok {
		return nil, ErrNotFound
	}

	return cloneItem(item), nil
}

func (ms *MemoryStore) ListBySession(sessionID string) ([]*ProgramItem, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	result := make([]*ProgramItem, 0)
	for _, item := range ms.items {
		if item.SessionID == sessionID {
			result = append(result, cloneItem(item))
		}
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Position == result[j].Position {
			return result[i].CreatedAt.Before(result[j].CreatedAt)
		}
		return result[i].Position < result[j].Position
	})

	return result, nil
}

func (ms *MemoryStore) Update(item *ProgramItem) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if _, ok := ms.items[item.ID]; !ok {
		return ErrNotFound
	}

	ms.items[item.ID] = cloneItem(item)
	return nil
}

func (ms *MemoryStore) Reorder(sessionID string, positions map[string]int) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	now := time.Now().UTC()
	for id, position := range positions {
		item, ok := ms.items[id]
		if !ok || item.SessionID != sessionID {
			return ErrNotFound
		}
		item.Position = position
		item.UpdatedAt = now
	}

	return nil
}

func (ms *MemoryStore) HasOverlap(sessionID string, start, end time.Time, excludeID string) (bool, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	for _, item := range ms.items {
		if item.SessionID != sessionID || item.ID == excludeID || item.Status != StatusScheduled {
			continue
		}
		if item.ScheduledStart.Before(end) && item.ScheduledEnd.After(start) {
			return true, nil
		}
	}

	return false, nil
}

func (ms *MemoryStore) PositionExists(sessionID string, position int, excludeID string) (bool, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	for _, item := range ms.items {
		if item.SessionID == sessionID && item.ID != excludeID && item.Position == position {
			return true, nil
		}
	}

	return false, nil
}

func (ms *MemoryStore) SessionExists(sessionID string) bool {
	if ms.sessionExistsFn == nil {
		return false
	}
	return ms.sessionExistsFn(sessionID)
}

func cloneItem(item *ProgramItem) *ProgramItem {
	if item == nil {
		return nil
	}
	cloned := *item
	cloned.Metadata = cloneMetadata(item.Metadata)
	return &cloned
}
