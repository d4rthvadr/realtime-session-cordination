package sessionlog

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// Entry is an append-only timeline record scoped to a session.
type Entry struct {
	ID            string
	SessionID     string
	ProgramItemID *string
	EventType     EventType
	Message       string
	Metadata      map[string]any
	OccurredAt    time.Time
	RequestID     string
	CreatedAt     time.Time
}

// Snapshot is the API-facing shape for a session log entry.
type Snapshot struct {
	ID            string         `json:"id"`
	SessionID     string         `json:"sessionId"`
	ProgramItemID *string        `json:"programItemId,omitempty"`
	EventType     string         `json:"eventType"`
	Message       string         `json:"message"`
	Metadata      map[string]any `json:"metadata,omitempty"`
	OccurredAt    time.Time      `json:"occurredAt"`
	RequestID     string         `json:"requestId,omitempty"`
	CreatedAt     time.Time      `json:"createdAt"`
}

// AppendInput defines data required to append a new entry.
type AppendInput struct {
	SessionID     string
	ProgramItemID *string
	EventType     EventType
	MessageInput  MessageInput
	Message       string
	Metadata      map[string]any
	OccurredAt    time.Time
	RequestID     string
}

// ListOptions controls session-scoped timeline retrieval.
type ListOptions struct {
	Limit      int
	Offset     int
	EventType  EventType
	EntityType string
}

func (e *Entry) Snapshot() Snapshot {
	return Snapshot{
		ID:            e.ID,
		SessionID:     e.SessionID,
		ProgramItemID: e.ProgramItemID,
		EventType:     e.EventType.String(),
		Message:       e.Message,
		Metadata:      cloneMetadata(e.Metadata),
		OccurredAt:    e.OccurredAt,
		RequestID:     e.RequestID,
		CreatedAt:     e.CreatedAt,
	}
}

func newID() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return fmt.Sprintf("log_%s", hex.EncodeToString(b))
}

func cloneMetadata(metadata map[string]any) map[string]any {
	if metadata == nil {
		return nil
	}
	clone := make(map[string]any, len(metadata))
	for k, v := range metadata {
		clone[k] = v
	}
	return clone
}
