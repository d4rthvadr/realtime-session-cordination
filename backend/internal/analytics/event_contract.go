package analytics

import "time"

// EventSource describes where an analytics event originated.
type EventSource string

const (
	EventSourceServer EventSource = "server"
	EventSourceClient EventSource = "client"
)

// EventRecord is the canonical raw analytics event payload persisted for processing.
type EventRecord struct {
	ID            string
	SessionID     string
	ProgramItemID *string
	EventKey      string
	OccurredAt    time.Time
	IngestedAt    time.Time
	Source        EventSource
	PayloadJSON   []byte
}

// OutboxRecord tracks an event pending asynchronous processor consumption.
type OutboxRecord struct {
	ID          int64
	EventID     string
	State       string
	LeaseOwner  string
	LeasedUntil *time.Time
	Attempt     int
	LastError   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ProcessorCheckpoint stores processor progress for idempotent aggregation.
type ProcessorCheckpoint struct {
	WorkerName  string
	LastEventID string
	UpdatedAt   time.Time
}

// EventStore persists raw analytics events.
type EventStore interface {
	AppendEvent(record EventRecord) error
}

// IngestionStore persists raw analytics events and outbox records atomically.
type IngestionStore interface {
	AppendEventAndEnqueue(record EventRecord, now time.Time) error
}

// OutboxStore persists outbox records consumed by the analytics processor.
type OutboxStore interface {
	Enqueue(eventID string, now time.Time) error
}

// CheckpointStore persists processor checkpoints.
type CheckpointStore interface {
	SaveCheckpoint(checkpoint ProcessorCheckpoint) error
}
