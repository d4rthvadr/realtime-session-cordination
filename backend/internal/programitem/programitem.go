package programitem

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

const (
	// Program items are either scheduled normally or kept as a canceled timeline slot.
	StatusScheduled = "scheduled"
	StatusCanceled  = "canceled"
)

var (
	ErrNotFound          = errors.New("program item not found")
	ErrSessionNotFound   = errors.New("session not found")
	ErrInvalidRange      = errors.New("scheduled_start must be before scheduled_end")
	ErrOverlap           = errors.New("program item overlaps with existing item")
	ErrDuplicatePosition = errors.New("position already exists in session")
	ErrInvalidStatus     = errors.New("invalid program item status")
)

type ProgramItem struct {
	ID                      string
	SessionID               string
	Title                   string
	Type                    string
	Status                  string
	HostName                string
	ScheduledStart          time.Time
	ScheduledEnd            time.Time
	ExpectedDurationMinutes int
	Position                int
	Location                string
	Metadata                map[string]any
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

type Snapshot struct {
	ID                      string         `json:"id"`
	SessionID               string         `json:"sessionId"`
	Title                   string         `json:"title"`
	Type                    string         `json:"type"`
	Status                  string         `json:"status"`
	HostName                string         `json:"hostName,omitempty"`
	ScheduledStart          time.Time      `json:"scheduledStart"`
	ScheduledEnd            time.Time      `json:"scheduledEnd"`
	ExpectedDurationMinutes int            `json:"expectedDurationMinutes"`
	Position                int            `json:"position"`
	Location                string         `json:"location,omitempty"`
	Metadata                map[string]any `json:"metadata,omitempty"`
	CreatedAt               time.Time      `json:"createdAt"`
	UpdatedAt               time.Time      `json:"updatedAt"`
}

type CreateInput struct {
	SessionID               string         `json:"sessionId"`
	Title                   string         `json:"title"`
	Type                    string         `json:"type"`
	HostName                string         `json:"hostName"`
	ScheduledStart          time.Time      `json:"scheduledStart"`
	ScheduledEnd            time.Time      `json:"scheduledEnd"`
	ExpectedDurationMinutes int            `json:"expectedDurationMinutes"`
	Position                int            `json:"position"`
	Location                string         `json:"location"`
	Metadata                map[string]any `json:"metadata"`
}

type UpdateInput struct {
	Title                   *string         `json:"title"`
	Type                    *string         `json:"type"`
	Status                  *string         `json:"status"`
	HostName                *string         `json:"hostName"`
	ScheduledStart          *time.Time      `json:"scheduledStart"`
	ScheduledEnd            *time.Time      `json:"scheduledEnd"`
	ExpectedDurationMinutes *int            `json:"expectedDurationMinutes"`
	Position                *int            `json:"position"`
	Location                *string         `json:"location"`
	Metadata                *map[string]any `json:"metadata"`
}

type ReorderItem struct {
	ID       string `json:"id"`
	Position int    `json:"position"`
}

type Manager struct {
	store Store
}

func NewManager(store Store) *Manager {
	return &Manager{store: store}
}

func (m *Manager) Create(input CreateInput) (Snapshot, error) {
	// Manager-level validation catches obvious payload issues before the store is touched.
	if input.SessionID == "" || input.Title == "" || input.Type == "" || input.Position <= 0 {
		return Snapshot{}, fmt.Errorf("invalid create program item payload")
	}
	if !m.store.SessionExists(input.SessionID) {
		return Snapshot{}, ErrSessionNotFound
	}
	if !input.ScheduledStart.Before(input.ScheduledEnd) {
		return Snapshot{}, ErrInvalidRange
	}

	hasOverlap, err := m.store.HasOverlap(input.SessionID, input.ScheduledStart, input.ScheduledEnd, "")
	if err != nil {
		return Snapshot{}, err
	}
	if hasOverlap {
		return Snapshot{}, ErrOverlap
	}

	posExists, err := m.store.PositionExists(input.SessionID, input.Position, "")
	if err != nil {
		return Snapshot{}, err
	}
	if posExists {
		return Snapshot{}, ErrDuplicatePosition
	}

	now := time.Now().UTC()
	item := &ProgramItem{
		ID:                      newID(),
		SessionID:               input.SessionID,
		Title:                   input.Title,
		Type:                    input.Type,
		Status:                  StatusScheduled,
		HostName:                input.HostName,
		ScheduledStart:          input.ScheduledStart.UTC(),
		ScheduledEnd:            input.ScheduledEnd.UTC(),
		ExpectedDurationMinutes: input.ExpectedDurationMinutes,
		Position:                input.Position,
		Location:                input.Location,
		Metadata:                cloneMetadata(input.Metadata),
		CreatedAt:               now,
		UpdatedAt:               now,
	}

	if item.ExpectedDurationMinutes <= 0 {
		item.ExpectedDurationMinutes = int(item.ScheduledEnd.Sub(item.ScheduledStart).Minutes())
	}

	created, err := m.store.Create(item)
	if err != nil {
		return Snapshot{}, err
	}

	return buildSnapshot(created), nil
}

func (m *Manager) GetSnapshot(id string) (Snapshot, error) {
	item, err := m.store.Get(id)
	if err != nil {
		return Snapshot{}, err
	}
	return buildSnapshot(item), nil
}

func (m *Manager) ListSnapshots(sessionID string) ([]Snapshot, error) {
	items, err := m.store.ListBySession(sessionID)
	if err != nil {
		return nil, err
	}

	snaps := make([]Snapshot, 0, len(items))
	for _, item := range items {
		snaps = append(snaps, buildSnapshot(item))
	}

	return snaps, nil
}

func (m *Manager) Update(id string, input UpdateInput) (Snapshot, error) {
	// Update applies only the fields the client supplied, then revalidates the final item state.
	item, err := m.store.Get(id)
	if err != nil {
		return Snapshot{}, err
	}

	if input.Title != nil {
		item.Title = *input.Title
	}
	if input.Type != nil {
		item.Type = *input.Type
	}
	if input.HostName != nil {
		item.HostName = *input.HostName
	}
	if input.Location != nil {
		item.Location = *input.Location
	}
	if input.Metadata != nil {
		item.Metadata = cloneMetadata(*input.Metadata)
	}
	if input.ExpectedDurationMinutes != nil {
		item.ExpectedDurationMinutes = *input.ExpectedDurationMinutes
	}
	if input.Position != nil {
		if *input.Position <= 0 {
			return Snapshot{}, fmt.Errorf("invalid update program item payload")
		}
		item.Position = *input.Position
	}
	if input.ScheduledStart != nil {
		item.ScheduledStart = input.ScheduledStart.UTC()
	}
	if input.ScheduledEnd != nil {
		item.ScheduledEnd = input.ScheduledEnd.UTC()
	}
	if input.Status != nil {
		if *input.Status != StatusScheduled && *input.Status != StatusCanceled {
			return Snapshot{}, ErrInvalidStatus
		}
		item.Status = *input.Status
	}

	if !item.ScheduledStart.Before(item.ScheduledEnd) {
		return Snapshot{}, ErrInvalidRange
	}

	if item.Status == StatusScheduled {
		hasOverlap, overlapErr := m.store.HasOverlap(item.SessionID, item.ScheduledStart, item.ScheduledEnd, item.ID)
		if overlapErr != nil {
			return Snapshot{}, overlapErr
		}
		if hasOverlap {
			return Snapshot{}, ErrOverlap
		}
	}

	posExists, posErr := m.store.PositionExists(item.SessionID, item.Position, item.ID)
	if posErr != nil {
		return Snapshot{}, posErr
	}
	if posExists {
		return Snapshot{}, ErrDuplicatePosition
	}

	if item.ExpectedDurationMinutes <= 0 {
		item.ExpectedDurationMinutes = int(item.ScheduledEnd.Sub(item.ScheduledStart).Minutes())
	}
	item.UpdatedAt = time.Now().UTC()

	if err = m.store.Update(item); err != nil {
		return Snapshot{}, err
	}

	return buildSnapshot(item), nil
}

func (m *Manager) Cancel(id string) (Snapshot, error) {
	item, err := m.store.Get(id)
	if err != nil {
		return Snapshot{}, err
	}

	item.Status = StatusCanceled
	item.UpdatedAt = time.Now().UTC()
	if err = m.store.Update(item); err != nil {
		return Snapshot{}, err
	}

	return buildSnapshot(item), nil
}

func (m *Manager) Reorder(sessionID string, items []ReorderItem) ([]Snapshot, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("invalid reorder program items payload")
	}

	positions := make(map[string]int, len(items))
	seen := make(map[int]struct{}, len(items))
	for _, item := range items {
		if item.ID == "" || item.Position <= 0 {
			return nil, fmt.Errorf("invalid reorder program items payload")
		}
		if _, ok := seen[item.Position]; ok {
			return nil, ErrDuplicatePosition
		}
		seen[item.Position] = struct{}{}
		positions[item.ID] = item.Position
	}

	if err := m.store.Reorder(sessionID, positions); err != nil {
		return nil, err
	}

	return m.ListSnapshots(sessionID)
}

func buildSnapshot(item *ProgramItem) Snapshot {
	return Snapshot{
		ID:                      item.ID,
		SessionID:               item.SessionID,
		Title:                   item.Title,
		Type:                    item.Type,
		Status:                  item.Status,
		HostName:                item.HostName,
		ScheduledStart:          item.ScheduledStart,
		ScheduledEnd:            item.ScheduledEnd,
		ExpectedDurationMinutes: item.ExpectedDurationMinutes,
		Position:                item.Position,
		Location:                item.Location,
		Metadata:                cloneMetadata(item.Metadata),
		CreatedAt:               item.CreatedAt,
		UpdatedAt:               item.UpdatedAt,
	}
}

func cloneMetadata(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func newID() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return fmt.Sprintf("pi_%s", hex.EncodeToString(b))
}
