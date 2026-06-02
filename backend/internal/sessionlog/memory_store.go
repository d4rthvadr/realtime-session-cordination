package sessionlog

import (
	"sort"
	"sync"
)

type MemoryStore struct {
	mu      sync.RWMutex
	entries map[string][]*Entry
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		entries: make(map[string][]*Entry),
	}
}

func (ms *MemoryStore) Append(entry *Entry) (*Entry, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	copyEntry := cloneEntry(entry)
	ms.entries[entry.SessionID] = append(ms.entries[entry.SessionID], copyEntry)

	return cloneEntry(copyEntry), nil
}

func (ms *MemoryStore) ListBySession(sessionID string, options ListOptions) ([]*Entry, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	all := ms.entries[sessionID]
	if len(all) == 0 {
		return []*Entry{}, nil
	}

	ordered := make([]*Entry, 0, len(all))
	for _, entry := range all {
		if options.EventType != "" && entry.EventType != options.EventType {
			continue
		}
		if options.EntityType != "" && EventEntity(entry.EventType) != options.EntityType {
			continue
		}
		ordered = append(ordered, cloneEntry(entry))
	}

	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].OccurredAt.Equal(ordered[j].OccurredAt) {
			if ordered[i].CreatedAt.Equal(ordered[j].CreatedAt) {
				return ordered[i].ID > ordered[j].ID
			}
			return ordered[i].CreatedAt.After(ordered[j].CreatedAt)
		}
		return ordered[i].OccurredAt.After(ordered[j].OccurredAt)
	})

	start := options.Offset
	if start > len(ordered) {
		start = len(ordered)
	}

	end := start + options.Limit
	if end > len(ordered) {
		end = len(ordered)
	}

	return ordered[start:end], nil
}

func cloneEntry(entry *Entry) *Entry {
	if entry == nil {
		return nil
	}

	var programItemID *string
	if entry.ProgramItemID != nil {
		copied := *entry.ProgramItemID
		programItemID = &copied
	}

	return &Entry{
		ID:            entry.ID,
		SessionID:     entry.SessionID,
		ProgramItemID: programItemID,
		EventType:     entry.EventType,
		Message:       entry.Message,
		Metadata:      cloneMetadata(entry.Metadata),
		OccurredAt:    entry.OccurredAt,
		RequestID:     entry.RequestID,
		CreatedAt:     entry.CreatedAt,
	}
}
