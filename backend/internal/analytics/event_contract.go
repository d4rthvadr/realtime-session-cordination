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

// ProcessorFreshness exposes progress and lag metadata for analytics reads.
type ProcessorFreshness struct {
	WorkerName        string
	LastEventID       string
	LastProcessedAt   *time.Time
	PendingCount      int
	OldestPendingAt   *time.Time
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

// ProcessorStore supports outbox claiming and state transitions.
type ProcessorStore interface {
	ClaimPendingForProcessing(workerName string, leaseUntil time.Time, limit int, now time.Time) ([]OutboxRecord, error)
	GetEvent(eventID string) (EventRecord, error)
	MarkProcessed(outboxID int64, now time.Time) error
	MarkFailed(outboxID int64, lastError string, deadLetter bool, now time.Time) error
	SaveCheckpoint(checkpoint ProcessorCheckpoint) error
	LoadCheckpoint(workerName string) (ProcessorCheckpoint, bool, error)
	GetFreshness(workerName string, now time.Time) (ProcessorFreshness, error)
}

// SessionProjection is the materialized per-session analytics summary produced by the processor.
type SessionProjection struct {
	SessionID              string
	SessionStatus          string
	SessionDurationSeconds int
	ProgramItemCount       int
	ScheduledCount         int
	InProgressCount        int
	PausedCount            int
	EndedCount             int
	CanceledCount          int
	PlannedSeconds         int
	EffectiveBudgetSeconds int
	TotalAdjustmentSeconds int
	TotalPauseSeconds      int
	TotalPauseCount        int
	EndedOnTimeCount       int
	OverrunItemCount       int
	TotalOverrunSeconds    int
	TotalUnderrunSeconds   int
	EndedOnTimeRatio       float64
	ComputedAt             time.Time
	UpdatedAt              time.Time
}

// PlatformProjection is the materialized platform-wide analytics summary produced by the processor.
type PlatformProjection struct {
	TotalSessions            int
	CreatedSessions          int
	LiveSessions             int
	PausedSessions           int
	EndedSessions            int
	TotalProgramItems        int
	EndedProgramItems        int
	OnTimeEndedProgramItems  int
	OverrunProgramItems      int
	TotalSessionDurationSecs int
	TotalPlannedSeconds      int
	EffectiveBudgetSeconds   int
	TotalAdjustmentSeconds   int
	TotalPauseSeconds        int
	TotalPauseCount          int
	TotalOverrunSeconds      int
	TotalUnderrunSeconds     int
	SessionCompletionRatio   float64
	ProgramItemOnTimeRatio   float64
	ComputedAt               time.Time
	UpdatedAt                time.Time
}

// ProjectionStore persists and retrieves materialized analytics projections.
type ProjectionStore interface {
	UpsertSessionProjection(p SessionProjection) error
	GetSessionProjection(sessionID string) (SessionProjection, bool, error)
	UpsertPlatformProjection(p PlatformProjection) error
	GetPlatformProjection() (PlatformProjection, bool, error)
}
