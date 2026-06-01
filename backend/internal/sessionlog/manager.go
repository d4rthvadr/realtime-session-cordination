package sessionlog

import (
	"fmt"
	"strings"
	"time"
)

const (
	defaultListLimit = 50
	maxListLimit     = 200
)

type Manager struct {
	store Store
}

func NewManager(store Store) *Manager {
	return &Manager{store: store}
}

func (m *Manager) Append(input AppendInput) (Snapshot, error) {
	if strings.TrimSpace(input.SessionID) == "" {
		return Snapshot{}, fmt.Errorf("session_id is required")
	}
	if strings.TrimSpace(input.EventType.String()) == "" {
		return Snapshot{}, fmt.Errorf("event_type is required")
	}

	now := time.Now().UTC()
	occurredAt := input.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = now
	}

	message := strings.TrimSpace(input.Message)
	if message == "" {
		message = RenderMessage(input.EventType, input.MessageInput)
	}

	entry := &Entry{
		ID:            newID(),
		SessionID:     input.SessionID,
		ProgramItemID: input.ProgramItemID,
		EventType:     input.EventType,
		Message:       message,
		Metadata:      cloneMetadata(input.Metadata),
		OccurredAt:    occurredAt.UTC(),
		RequestID:     strings.TrimSpace(input.RequestID),
		CreatedAt:     now,
	}

	created, err := m.store.Append(entry)
	if err != nil {
		return Snapshot{}, err
	}

	return created.Snapshot(), nil
}

func (m *Manager) ListBySession(sessionID string, options ListOptions) ([]Snapshot, error) {
	if strings.TrimSpace(sessionID) == "" {
		return nil, fmt.Errorf("session_id is required")
	}

	normalized := normalizeListOptions(options)
	entries, err := m.store.ListBySession(sessionID, normalized)
	if err != nil {
		return nil, err
	}

	snaps := make([]Snapshot, 0, len(entries))
	for _, entry := range entries {
		snaps = append(snaps, entry.Snapshot())
	}

	return snaps, nil
}

func normalizeListOptions(options ListOptions) ListOptions {
	limit := options.Limit
	if limit <= 0 {
		limit = defaultListLimit
	}
	if limit > maxListLimit {
		limit = maxListLimit
	}

	offset := options.Offset
	if offset < 0 {
		offset = 0
	}

	return ListOptions{Limit: limit, Offset: offset}
}
