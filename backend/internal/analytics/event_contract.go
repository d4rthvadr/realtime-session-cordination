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
	NextRetryAt *time.Time
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
	WorkerName      string     `json:"workerName"`
	LastEventID     string     `json:"lastEventId"`
	LastProcessedAt *time.Time `json:"lastProcessedAt,omitempty"`
	PendingCount    int        `json:"pendingCount"`
	OldestPendingAt *time.Time `json:"oldestPendingAt,omitempty"`
	RetryDueCount   int        `json:"retryDueCount"`
	DeadLetterCount int        `json:"deadLetterCount"`
	RetryLagSeconds int        `json:"retryLagSeconds"`
}

// ProcessorMetrics captures processor execution counters and recent timings.
type ProcessorMetrics struct {
	ProcessedCount          int64      `json:"processedCount"`
	FailedCount             int64      `json:"failedCount"`
	DeadLetterCount         int64      `json:"deadLetterCount"`
	ProjectionErrorCount    int64      `json:"projectionErrorCount"`
	CheckpointErrorCount    int64      `json:"checkpointErrorCount"`
	LastBatchDurationMillis int64      `json:"lastBatchDurationMillis"`
	LastBatchAt             *time.Time `json:"lastBatchAt,omitempty"`
}

// ProcessorMetricsStore allows API handlers to read processor runtime metrics.
type ProcessorMetricsStore interface {
	GetProcessorMetrics() ProcessorMetrics
}

// DeadLetterRecord represents a failed outbox row parked in dead-letter state.
type DeadLetterRecord struct {
	OutboxID      int64     `json:"outboxId"`
	EventID       string    `json:"eventId"`
	SessionID     string    `json:"sessionId"`
	ProgramItemID *string   `json:"programItemId,omitempty"`
	EventKey      string    `json:"eventKey"`
	OccurredAt    time.Time `json:"occurredAt"`
	IngestedAt    time.Time `json:"ingestedAt"`
	Attempt       int       `json:"attempt"`
	LastError     string    `json:"lastError"`
	FailedAt      time.Time `json:"failedAt"`
	PayloadJSON   []byte    `json:"payloadJson,omitempty"`
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
	MarkFailed(outboxID int64, lastError string, deadLetter bool, nextRetryAt *time.Time, now time.Time) error
	SaveCheckpoint(checkpoint ProcessorCheckpoint) error
	LoadCheckpoint(workerName string) (ProcessorCheckpoint, bool, error)
	GetFreshness(workerName string, now time.Time) (ProcessorFreshness, error)
}

// DeadLetterStore provides DLQ visibility and replay operations.
type DeadLetterStore interface {
	ListDeadLetters(limit int, offset int) ([]DeadLetterRecord, error)
	GetDeadLetter(outboxID int64) (DeadLetterRecord, bool, error)
	RetryDeadLetter(outboxID int64, now time.Time) error
}

// CleanupStore provides retention cleanup operations for analytics persistence.
type CleanupStore interface {
	CleanupProcessedOutbox(olderThan time.Time) (int64, error)
	CleanupDeadLetters(olderThan time.Time) (int64, error)
	CleanupEvents(olderThan time.Time) (int64, error)
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
