package session

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

const (
	StatusCreated = "CREATED"
	StatusLive    = "LIVE"
	StatusPaused  = "PAUSED"
	StatusEnded   = "ENDED"
)

var (
	ErrNotFound          = errors.New("session not found")
	ErrInvalidTransition = errors.New("invalid state transition")
	ErrUnauthorized      = errors.New("invalid control token")
)

type Session struct {
	ID                         string
	Title                      string
	SpeakerName                string
	DurationSeconds            int
	Status                     string
	StartedAt                  *time.Time
	PausedAt                   *time.Time
	TotalPausedDurationSeconds int
	AdjustmentSeconds          int
	EndedRemainingSeconds      *int
	ControlToken               string
	CreatedAt                  time.Time
}

type Snapshot struct {
	ID                         string     `json:"id"`
	Title                      string     `json:"title"`
	SpeakerName                string     `json:"speakerName"`
	DurationSeconds            int        `json:"durationSeconds"`
	Status                     string     `json:"status"`
	StartedAt                  *time.Time `json:"startedAt,omitempty"`
	PausedAt                   *time.Time `json:"pausedAt,omitempty"`
	TotalPausedDurationSeconds int        `json:"totalPausedDurationSeconds"`
	CreatedAt                  time.Time  `json:"createdAt"`
	RemainingSeconds           int        `json:"remainingSeconds"`
}

type Event struct {
	Type         string   `json:"type"`
	Session      Snapshot `json:"session"`
	StartedAt    int64    `json:"startedAt,omitempty"`
	DeltaSeconds int      `json:"deltaSeconds,omitempty"`
}

type CreateInput struct {
	Title           string `json:"title"`
	SpeakerName     string `json:"speakerName"`
	DurationSeconds int    `json:"durationSeconds"`
}

type Manager struct {
	store Store
}

// NewManager creates a new Manager with the given Store implementation.
func NewManager(store Store) *Manager {
	return &Manager{store: store}
}

// NewManagerWithMemory creates a Manager with an in-memory store (for backward compatibility).
func NewManagerWithMemory() *Manager {
	return NewManager(NewMemoryStore())
}

func (m *Manager) Create(input CreateInput) (Snapshot, string, error) {
	if input.Title == "" || input.SpeakerName == "" || input.DurationSeconds <= 0 {
		return Snapshot{}, "", fmt.Errorf("invalid create session payload")
	}

	now := time.Now().UTC()
	id := newID()
	token, err := newControlToken()
	if err != nil {
		return Snapshot{}, "", err
	}

	s := &Session{
		ID:              id,
		Title:           input.Title,
		SpeakerName:     input.SpeakerName,
		DurationSeconds: input.DurationSeconds,
		Status:          StatusCreated,
		ControlToken:    token,
		CreatedAt:       now,
	}

	_, err = m.store.Create(s)
	if err != nil {
		return Snapshot{}, "", err
	}

	return buildSnapshot(s, now), token, nil
}

func (m *Manager) GetSnapshot(id string) (Snapshot, error) {
	s, err := m.store.Get(id)
	if err != nil {
		return Snapshot{}, err
	}
	return buildSnapshot(s, time.Now().UTC()), nil
}

func (m *Manager) ListSnapshots() ([]Snapshot, error) {
	sessions, err := m.store.List()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	snapshots := make([]Snapshot, 0, len(sessions))
	for _, s := range sessions {
		snapshots = append(snapshots, buildSnapshot(s, now))
	}

	return snapshots, nil
}

func (m *Manager) ValidateControlToken(id, token string) error {
	return m.store.ValidateControlToken(id, token)
}

func (m *Manager) Start(id string) (Event, error) {
	now := time.Now().UTC()

	s, err := m.store.Get(id)
	if err != nil {
		return Event{}, err
	}
	if s.Status != StatusCreated {
		return Event{}, ErrInvalidTransition
	}

	s.Status = StatusLive
	s.StartedAt = &now
	s.PausedAt = nil

	if err = m.store.Update(s); err != nil {
		return Event{}, err
	}

	snapshot := buildSnapshot(s, now)
	return Event{Type: "SESSION_STARTED", Session: snapshot, StartedAt: now.Unix()}, nil
}

func (m *Manager) Pause(id string) (Event, error) {
	now := time.Now().UTC()

	s, err := m.store.Get(id)
	if err != nil {
		return Event{}, err
	}
	if s.Status != StatusLive {
		return Event{}, ErrInvalidTransition
	}

	s.Status = StatusPaused
	s.PausedAt = &now

	if err = m.store.Update(s); err != nil {
		return Event{}, err
	}

	snapshot := buildSnapshot(s, now)
	return Event{Type: "SESSION_PAUSED", Session: snapshot}, nil
}

func (m *Manager) Resume(id string) (Event, error) {
	now := time.Now().UTC()

	s, err := m.store.Get(id)
	if err != nil {
		return Event{}, err
	}
	if s.Status != StatusPaused || s.PausedAt == nil {
		return Event{}, ErrInvalidTransition
	}

	pausedFor := now.Sub(*s.PausedAt)
	s.TotalPausedDurationSeconds += int(pausedFor.Seconds())
	s.PausedAt = nil
	s.Status = StatusLive

	if err = m.store.Update(s); err != nil {
		return Event{}, err
	}

	snapshot := buildSnapshot(s, now)
	return Event{Type: "SESSION_RESUMED", Session: snapshot}, nil
}

func (m *Manager) End(id string) (Event, error) {
	now := time.Now().UTC()

	s, err := m.store.Get(id)
	if err != nil {
		return Event{}, err
	}
	if s.Status != StatusLive && s.Status != StatusPaused {
		return Event{}, ErrInvalidTransition
	}

	remaining := computeRemainingSeconds(s, now)
	s.EndedRemainingSeconds = &remaining
	s.Status = StatusEnded
	s.PausedAt = nil

	if err = m.store.Update(s); err != nil {
		return Event{}, err
	}

	snapshot := buildSnapshot(s, now)
	return Event{Type: "SESSION_ENDED", Session: snapshot}, nil
}

func (m *Manager) AdjustTime(id string, deltaSeconds int) (Event, error) {
	now := time.Now().UTC()

	s, err := m.store.Get(id)
	if err != nil {
		return Event{}, err
	}
	if s.Status == StatusEnded {
		return Event{}, ErrInvalidTransition
	}

	s.AdjustmentSeconds += deltaSeconds

	if err = m.store.Update(s); err != nil {
		return Event{}, err
	}

	snapshot := buildSnapshot(s, now)
	return Event{Type: "TIME_ADJUSTED", Session: snapshot, DeltaSeconds: deltaSeconds}, nil
}

func (m *Manager) SessionExists(id string) bool {
	return m.store.SessionExists(id)
}

func buildSnapshot(s *Session, now time.Time) Snapshot {
	return Snapshot{
		ID:                         s.ID,
		Title:                      s.Title,
		SpeakerName:                s.SpeakerName,
		DurationSeconds:            s.DurationSeconds,
		Status:                     s.Status,
		StartedAt:                  s.StartedAt,
		PausedAt:                   s.PausedAt,
		TotalPausedDurationSeconds: s.TotalPausedDurationSeconds,
		CreatedAt:                  s.CreatedAt,
		RemainingSeconds:           computeRemainingSeconds(s, now),
	}
}

func computeRemainingSeconds(s *Session, now time.Time) int {
	base := s.DurationSeconds + s.AdjustmentSeconds

	switch s.Status {
	case StatusCreated:
		return base
	case StatusLive:
		if s.StartedAt == nil {
			return base
		}
		elapsed := int(now.Sub(*s.StartedAt).Seconds()) - s.TotalPausedDurationSeconds
		return base - elapsed
	case StatusPaused:
		if s.StartedAt == nil || s.PausedAt == nil {
			return base
		}
		elapsed := int(s.PausedAt.Sub(*s.StartedAt).Seconds()) - s.TotalPausedDurationSeconds
		return base - elapsed
	case StatusEnded:
		if s.EndedRemainingSeconds != nil {
			return *s.EndedRemainingSeconds
		}
		return base
	default:
		return base
	}
}

func newID() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return fmt.Sprintf("sess_%s", hex.EncodeToString(b))
}

func newControlToken() (string, error) {
	b := make([]byte, 24)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
