package programitem

import (
	"sort"
	"sync"
	"time"
)

type SessionExistsFunc func(sessionID string) bool

// MemoryStore is an in-memory ProgramItem store implementation.
type MemoryStore struct {
	mu      sync.RWMutex
	locksMu sync.Mutex
	// sessionLocks keeps one mutex per session so writes for different sessions stay independent.
	sessionLocks    map[string]*sync.Mutex
	items           map[string]*ProgramItem
	sessionExistsFn SessionExistsFunc
}

func NewMemoryStore(sessionExistsFn SessionExistsFunc) *MemoryStore {
	return &MemoryStore{
		items:           make(map[string]*ProgramItem),
		sessionLocks:    make(map[string]*sync.Mutex),
		sessionExistsFn: sessionExistsFn,
	}
}

func (ms *MemoryStore) Create(item *ProgramItem) (*ProgramItem, error) {
	// Lock the session first so overlap and position checks cannot race with another writer.
	unlock := ms.lockSession(item.SessionID)
	defer unlock()

	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ms.hasOverlapLocked(item.SessionID, item.ScheduledStart, item.ScheduledEnd, "") {
		return nil, ErrOverlap
	}
	if ms.positionExistsLocked(item.SessionID, item.Position, "") {
		return nil, ErrDuplicatePosition
	}

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
	// Update uses the same session-level lock as create so validation stays atomic.
	unlock := ms.lockSession(item.SessionID)
	defer unlock()

	ms.mu.Lock()
	defer ms.mu.Unlock()

	if _, ok := ms.items[item.ID]; !ok {
		return ErrNotFound
	}

	if item.Status == StatusScheduled && ms.hasOverlapLocked(item.SessionID, item.ScheduledStart, item.ScheduledEnd, item.ID) {
		return ErrOverlap
	}
	if ms.positionExistsLocked(item.SessionID, item.Position, item.ID) {
		return ErrDuplicatePosition
	}

	ms.items[item.ID] = cloneItem(item)
	return nil
}

func (ms *MemoryStore) TransitionToInProgress(id string, at time.Time) (*ProgramItem, error) {
	sessionID, err := ms.sessionIDForItem(id)
	if err != nil {
		return nil, err
	}

	unlock := ms.lockSession(sessionID)
	defer unlock()

	ms.mu.Lock()
	defer ms.mu.Unlock()

	item, ok := ms.items[id]
	if !ok {
		return nil, ErrNotFound
	}
	if item.Status != StatusScheduled {
		return nil, ErrInvalidStatusTransition
	}

	for _, candidate := range ms.items {
		if candidate.SessionID != item.SessionID || candidate.ID == item.ID {
			continue
		}
		if candidate.Status == StatusInProgress || candidate.Status == StatusPaused {
			return nil, ErrInProgressExists
		}
	}

	now := at.UTC()
	item.Status = StatusInProgress
	item.ActualStart = &now
	item.PausedAt = nil
	item.TotalPausedDurationSeconds = 0
	item.EndedRemainingSeconds = nil
	item.ActualEnd = nil
	item.PauseCount = 0
	item.EndedReason = ""
	item.UpdatedAt = now

	return cloneItem(item), nil
}

func (ms *MemoryStore) TransitionToPaused(id string, at time.Time) (*ProgramItem, error) {
	sessionID, err := ms.sessionIDForItem(id)
	if err != nil {
		return nil, err
	}

	unlock := ms.lockSession(sessionID)
	defer unlock()

	ms.mu.Lock()
	defer ms.mu.Unlock()

	item, ok := ms.items[id]
	if !ok {
		return nil, ErrNotFound
	}
	if item.Status != StatusInProgress {
		return nil, ErrInvalidStatusTransition
	}

	now := at.UTC()
	item.Status = StatusPaused
	item.PausedAt = &now
	item.UpdatedAt = now

	return cloneItem(item), nil
}

func (ms *MemoryStore) TransitionToResumed(id string, at time.Time) (*ProgramItem, error) {
	sessionID, err := ms.sessionIDForItem(id)
	if err != nil {
		return nil, err
	}

	unlock := ms.lockSession(sessionID)
	defer unlock()

	ms.mu.Lock()
	defer ms.mu.Unlock()

	item, ok := ms.items[id]
	if !ok {
		return nil, ErrNotFound
	}
	if item.Status != StatusPaused || item.PausedAt == nil {
		return nil, ErrInvalidStatusTransition
	}

	now := at.UTC()
	pausedFor := int(now.Sub(*item.PausedAt).Seconds())
	if pausedFor < 0 {
		pausedFor = 0
	}
	item.TotalPausedDurationSeconds += pausedFor
	item.PauseCount++
	item.Status = StatusInProgress
	item.PausedAt = nil
	item.UpdatedAt = now

	return cloneItem(item), nil
}

func (ms *MemoryStore) AdjustRuntime(id string, deltaSeconds int, at time.Time) (*ProgramItem, error) {
	sessionID, err := ms.sessionIDForItem(id)
	if err != nil {
		return nil, err
	}

	unlock := ms.lockSession(sessionID)
	defer unlock()

	ms.mu.Lock()
	defer ms.mu.Unlock()

	item, ok := ms.items[id]
	if !ok {
		return nil, ErrNotFound
	}
	if item.Status == StatusEnded || item.Status == StatusCanceled {
		return nil, ErrInvalidStatusTransition
	}

	item.AdjustmentSeconds += deltaSeconds
	item.UpdatedAt = at.UTC()

	return cloneItem(item), nil
}

func (ms *MemoryStore) TransitionToEnded(id string, at time.Time) (*ProgramItem, error) {
	sessionID, err := ms.sessionIDForItem(id)
	if err != nil {
		return nil, err
	}

	unlock := ms.lockSession(sessionID)
	defer unlock()

	ms.mu.Lock()
	defer ms.mu.Unlock()

	item, ok := ms.items[id]
	if !ok {
		return nil, ErrNotFound
	}
	if item.Status != StatusInProgress && item.Status != StatusPaused {
		return nil, ErrInvalidStatusTransition
	}

	now := at.UTC()
	item.Status = StatusEnded
	if item.ActualStart == nil {
		item.ActualStart = &now
	}
	item.PausedAt = nil
	item.EndedReason = "manual"
	item.ActualEnd = &now
	item.UpdatedAt = now

	return cloneItem(item), nil
}

func (ms *MemoryStore) Reorder(sessionID string, positions map[string]int) error {
	// Reorder is session-scoped; only one reorder/write should run for the same session at a time.
	unlock := ms.lockSession(sessionID)
	defer unlock()

	ms.mu.Lock()
	defer ms.mu.Unlock()

	targetPositions := make(map[int]struct{}, len(positions))
	for _, position := range positions {
		if _, ok := targetPositions[position]; ok {
			return ErrDuplicatePosition
		}
		targetPositions[position] = struct{}{}
	}

	for _, item := range ms.items {
		if item.SessionID != sessionID {
			continue
		}
		if _, moving := positions[item.ID]; moving {
			continue
		}
		if _, clash := targetPositions[item.Position]; clash {
			return ErrDuplicatePosition
		}
	}

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
	unlock := ms.lockSession(sessionID)
	defer unlock()

	ms.mu.RLock()
	defer ms.mu.RUnlock()

	return ms.hasOverlapLocked(sessionID, start, end, excludeID), nil
}

func (ms *MemoryStore) PositionExists(sessionID string, position int, excludeID string) (bool, error) {
	unlock := ms.lockSession(sessionID)
	defer unlock()

	ms.mu.RLock()
	defer ms.mu.RUnlock()

	return ms.positionExistsLocked(sessionID, position, excludeID), nil
}

func (ms *MemoryStore) HasInProgressItem(sessionID string, excludeID string) (bool, error) {
	unlock := ms.lockSession(sessionID)
	defer unlock()

	ms.mu.RLock()
	defer ms.mu.RUnlock()

	for _, item := range ms.items {
		if item.SessionID != sessionID || item.ID == excludeID {
			continue
		}
		if item.Status == StatusInProgress || item.Status == StatusPaused {
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

func (ms *MemoryStore) sessionIDForItem(id string) (string, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	item, ok := ms.items[id]
	if !ok {
		return "", ErrNotFound
	}

	return item.SessionID, nil
}

func (ms *MemoryStore) lockSession(sessionID string) func() {
	// Lazily create a mutex per session so unrelated sessions do not block each other.
	ms.locksMu.Lock()
	lock, ok := ms.sessionLocks[sessionID]
	if !ok {
		lock = &sync.Mutex{}
		ms.sessionLocks[sessionID] = lock
	}
	ms.locksMu.Unlock()

	lock.Lock()
	return lock.Unlock
}

func (ms *MemoryStore) hasOverlapLocked(sessionID string, start, end time.Time, excludeID string) bool {
	// Mirror the SQL overlap predicate in memory for identical validation behavior.
	for _, item := range ms.items {
		if item.SessionID != sessionID || item.ID == excludeID || item.Status != StatusScheduled {
			continue
		}
		if item.ScheduledStart.Before(end) && item.ScheduledEnd.After(start) {
			return true
		}
	}

	return false
}

func (ms *MemoryStore) positionExistsLocked(sessionID string, position int, excludeID string) bool {
	// Position uniqueness is per session, not global.
	for _, item := range ms.items {
		if item.SessionID == sessionID && item.ID != excludeID && item.Position == position {
			return true
		}
	}

	return false
}

func cloneItem(item *ProgramItem) *ProgramItem {
	if item == nil {
		return nil
	}
	cloned := *item
	cloned.Metadata = cloneMetadata(item.Metadata)
	return &cloned
}
