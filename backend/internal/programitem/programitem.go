package programitem

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"time"
)

const (
	// Program item lifecycle states.
	StatusScheduled  = "scheduled"
	StatusInProgress = "in_progress"
	StatusPaused     = "paused"
	StatusEnded      = "ended"
	StatusCanceled   = "canceled"

	// These event types are emitted by the manager when program items change, and used by clients to trigger UI updates.
	EventCreated   = "PROGRAM_ITEM_CREATED"
	EventUpdated   = "PROGRAM_ITEM_UPDATED"
	EventCanceled  = "PROGRAM_ITEM_CANCELED"
	EventStarted   = "PROGRAM_ITEM_STARTED"
	EventEnded     = "PROGRAM_ITEM_ENDED"
	EventReordered = "PROGRAM_ITEMS_REORDERED"
)

var (
	ErrNotFound                = errors.New("program item not found")
	ErrSessionNotFound         = errors.New("session not found")
	ErrInvalidRange            = errors.New("scheduled_start must be before scheduled_end")
	ErrOverlap                 = errors.New("program item overlaps with existing item")
	ErrDuplicatePosition       = errors.New("position already exists in session")
	ErrInvalidType             = errors.New("invalid program item type")
	ErrInvalidStatus           = errors.New("invalid program item status")
	ErrInvalidStatusTransition = errors.New("invalid program item status transition")
	ErrInProgressExists        = errors.New("another program item is already in progress")
)

var allowedTypes = map[string]struct{}{
	"announcement": {},
	"break":        {},
	"keynote":      {},
	"lecture":      {},
	"panel":        {},
	"q&a":          {},
}

type ProgramItem struct {
	ID                         string
	SessionID                  string
	Title                      string
	Type                       string
	Status                     string
	RuntimeDurationSeconds     int
	ActualStart                *time.Time
	PausedAt                   *time.Time
	TotalPausedDurationSeconds int
	AdjustmentSeconds          int
	EndedRemainingSeconds      *int
	ActualEnd                  *time.Time
	PauseCount                 int
	EndedReason                string
	HostName                   string
	ScheduledStart             time.Time
	ScheduledEnd               time.Time
	ExpectedDurationMinutes    int
	Position                   int
	Location                   string
	Metadata                   map[string]any
	CreatedAt                  time.Time
	UpdatedAt                  time.Time
}

type Snapshot struct {
	ID                         string         `json:"id"`
	SessionID                  string         `json:"sessionId"`
	Title                      string         `json:"title"`
	Type                       string         `json:"type"`
	Status                     string         `json:"status"`
	RuntimeDurationSeconds     int            `json:"runtimeDurationSeconds"`
	ActualStart                *time.Time     `json:"actualStart,omitempty"`
	PausedAt                   *time.Time     `json:"pausedAt,omitempty"`
	TotalPausedDurationSeconds int            `json:"totalPausedDurationSeconds"`
	AdjustmentSeconds          int            `json:"adjustmentSeconds"`
	EndedRemainingSeconds      *int           `json:"endedRemainingSeconds,omitempty"`
	ActualEnd                  *time.Time     `json:"actualEnd,omitempty"`
	PauseCount                 int            `json:"pauseCount"`
	EndedReason                string         `json:"endedReason,omitempty"`
	HostName                   string         `json:"hostName,omitempty"`
	ScheduledStart             time.Time      `json:"scheduledStart"`
	ScheduledEnd               time.Time      `json:"scheduledEnd"`
	ExpectedDurationMinutes    int            `json:"expectedDurationMinutes"`
	Position                   int            `json:"position"`
	Location                   string         `json:"location,omitempty"`
	Metadata                   map[string]any `json:"metadata,omitempty"`
	CreatedAt                  time.Time      `json:"createdAt"`
	UpdatedAt                  time.Time      `json:"updatedAt"`
	RemainingSeconds           int            `json:"remainingSeconds"`
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

// Event is the websocket payload emitted when a ProgramItem changes.
type Event struct {
	Type            string     `json:"type"`
	SessionID       string     `json:"sessionId,omitempty"`
	ProgramItem     *Snapshot  `json:"programItem,omitempty"`
	NextProgramItem *Snapshot  `json:"nextProgramItem,omitempty"`
	ProgramItems    []Snapshot `json:"programItems,omitempty"`
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
	if !IsAllowedType(input.Type) {
		return Snapshot{}, ErrInvalidType
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
		RuntimeDurationSeconds:  input.ExpectedDurationMinutes * 60,
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
	if item.RuntimeDurationSeconds <= 0 {
		item.RuntimeDurationSeconds = int(item.ScheduledEnd.Sub(item.ScheduledStart).Seconds())
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

// CurrentAndNextSnapshots returns current and next ProgramItem context for a session.
func (m *Manager) CurrentAndNextSnapshots(sessionID string, at time.Time) (*Snapshot, *Snapshot, error) {
	items, err := m.store.ListBySession(sessionID)
	if err != nil {
		return nil, nil, err
	}

	now := at.UTC()

	sorted := make([]*ProgramItem, 0, len(items))
	for _, item := range items {
		sorted = append(sorted, item)
	}

	if len(sorted) == 0 {
		return nil, nil, nil
	}

	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Position == sorted[j].Position {
			return sorted[i].ScheduledStart.Before(sorted[j].ScheduledStart)
		}
		return sorted[i].Position < sorted[j].Position
	})

	var current *ProgramItem
	for _, item := range sorted {
		if item.Status == StatusInProgress || item.Status == StatusPaused {
			current = item
			break
		}
	}

	if current == nil {
		for _, item := range sorted {
			if item.Status != StatusScheduled {
				continue
			}
			if (item.ScheduledStart.Equal(now) || item.ScheduledStart.Before(now)) && item.ScheduledEnd.After(now) {
				current = item
				break
			}
		}
	}

	var next *ProgramItem
	if current != nil {
		for _, item := range sorted {
			if item.Status != StatusScheduled {
				continue
			}
			if item.Position > current.Position {
				next = item
				break
			}
		}
	} else {
		for _, item := range sorted {
			if item.Status != StatusScheduled {
				continue
			}
			if item.ScheduledStart.After(now) || item.ScheduledStart.Equal(now) {
				next = item
				break
			}
		}
	}

	var currentSnap *Snapshot
	if current != nil {
		snap := buildSnapshot(current)
		currentSnap = &snap
	}

	var nextSnap *Snapshot
	if next != nil {
		snap := buildSnapshot(next)
		nextSnap = &snap
	}

	return currentSnap, nextSnap, nil
}

// CurrentSnapshot returns the currently active ProgramItem context for backwards compatibility.
func (m *Manager) CurrentSnapshot(sessionID string, at time.Time) (*Snapshot, error) {
	current, _, err := m.CurrentAndNextSnapshots(sessionID, at)
	if err != nil {
		return nil, err
	}

	return current, nil
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
		if !IsAllowedType(*input.Type) {
			return Snapshot{}, ErrInvalidType
		}
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

func (m *Manager) Start(id string) (Snapshot, error) {
	now := time.Now().UTC()
	item, err := m.store.TransitionToInProgress(id, now)
	if err != nil {
		return Snapshot{}, err
	}

	return buildSnapshot(item), nil
}

func (m *Manager) Pause(id string) (Snapshot, error) {
	now := time.Now().UTC()
	item, err := m.store.TransitionToPaused(id, now)
	if err != nil {
		return Snapshot{}, err
	}

	return buildSnapshot(item), nil
}

func (m *Manager) Resume(id string) (Snapshot, error) {
	now := time.Now().UTC()
	item, err := m.store.TransitionToResumed(id, now)
	if err != nil {
		return Snapshot{}, err
	}

	return buildSnapshot(item), nil
}

func (m *Manager) AdjustTime(id string, deltaSeconds int) (Snapshot, error) {
	now := time.Now().UTC()
	item, err := m.store.AdjustRuntime(id, deltaSeconds, now)
	if err != nil {
		return Snapshot{}, err
	}

	return buildSnapshot(item), nil
}

func (m *Manager) End(id string) (Snapshot, error) {
	now := time.Now().UTC()
	current, err := m.store.Get(id)
	if err != nil {
		return Snapshot{}, err
	}
	if current.Status != StatusInProgress && current.Status != StatusPaused {
		return Snapshot{}, ErrInvalidStatusTransition
	}
	remaining := computeRemainingSeconds(current, now)
	item, err := m.store.TransitionToEnded(id, now)
	if err != nil {
		return Snapshot{}, err
	}
	item.EndedRemainingSeconds = &remaining

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
	now := time.Now().UTC()
	return Snapshot{
		ID:                         item.ID,
		SessionID:                  item.SessionID,
		Title:                      item.Title,
		Type:                       item.Type,
		Status:                     item.Status,
		RuntimeDurationSeconds:     item.RuntimeDurationSeconds,
		ActualStart:                item.ActualStart,
		PausedAt:                   item.PausedAt,
		TotalPausedDurationSeconds: item.TotalPausedDurationSeconds,
		AdjustmentSeconds:          item.AdjustmentSeconds,
		EndedRemainingSeconds:      item.EndedRemainingSeconds,
		ActualEnd:                  item.ActualEnd,
		PauseCount:                 item.PauseCount,
		EndedReason:                item.EndedReason,
		HostName:                   item.HostName,
		ScheduledStart:             item.ScheduledStart,
		ScheduledEnd:               item.ScheduledEnd,
		ExpectedDurationMinutes:    item.ExpectedDurationMinutes,
		Position:                   item.Position,
		Location:                   item.Location,
		Metadata:                   cloneMetadata(item.Metadata),
		CreatedAt:                  item.CreatedAt,
		UpdatedAt:                  item.UpdatedAt,
		RemainingSeconds:           computeRemainingSeconds(item, now),
	}
}

func computeRemainingSeconds(item *ProgramItem, at time.Time) int {
	base := item.RuntimeDurationSeconds + item.AdjustmentSeconds

	switch item.Status {
	case StatusScheduled:
		return base
	case StatusInProgress:
		if item.ActualStart == nil {
			return base
		}
		elapsed := int(at.Sub(*item.ActualStart).Seconds()) - item.TotalPausedDurationSeconds
		return base - elapsed
	case StatusPaused:
		if item.ActualStart == nil || item.PausedAt == nil {
			return base
		}
		elapsed := int(item.PausedAt.Sub(*item.ActualStart).Seconds()) - item.TotalPausedDurationSeconds
		return base - elapsed
	case StatusEnded:
		if item.EndedRemainingSeconds != nil {
			return *item.EndedRemainingSeconds
		}
		if item.ActualEnd == nil || item.ActualStart == nil {
			return base
		}
		elapsed := int(item.ActualEnd.Sub(*item.ActualStart).Seconds()) - item.TotalPausedDurationSeconds
		return base - elapsed
	default:
		return base
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

func IsAllowedType(value string) bool {
	_, ok := allowedTypes[value]
	return ok
}

func newID() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return fmt.Sprintf("pi_%s", hex.EncodeToString(b))
}
