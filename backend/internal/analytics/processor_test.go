package analytics

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"realtime-session-coordination/backend/internal/programitem"
	"realtime-session-coordination/backend/internal/session"
)

func TestProcessorProcessBatchMarksProcessedAndCheckpoint(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "analytics-processor.db")
	store, err := NewSqliteStore(dbPath)
	if err != nil {
		t.Fatalf("new sqlite store: %v", err)
	}

	now := time.Now().UTC()
	event := EventRecord{
		ID:          "evt_proc_1",
		SessionID:   "sess_1",
		EventKey:    "SESSION_STARTED",
		OccurredAt:  now,
		IngestedAt:  now,
		Source:      EventSourceServer,
		PayloadJSON: []byte(`{"type":"SESSION_STARTED"}`),
	}
	if err = store.AppendEventAndEnqueue(event, now); err != nil {
		t.Fatalf("append+enqueue: %v", err)
	}

	processor := NewProcessor(store, nil, ProcessorConfig{
		WorkerName:    "analytics_processor",
		PollInterval:  5 * time.Second,
		LeaseDuration: 5 * time.Second,
		BatchSize:     10,
		MaxAttempts:   3,
		RetryPolicy:   DefaultRetryPolicy(),
	})

	if err = processor.processBatch(context.Background()); err != nil {
		t.Fatalf("process batch: %v", err)
	}

	var state string
	if err = store.db.QueryRow(`SELECT state FROM analytics_outbox WHERE event_id = ?`, event.ID).Scan(&state); err != nil {
		t.Fatalf("read outbox state: %v", err)
	}
	if state != OutboxStateProcessed {
		t.Fatalf("expected outbox state %q, got %q", OutboxStateProcessed, state)
	}

	cp, ok, err := store.LoadCheckpoint("analytics_processor")
	if err != nil {
		t.Fatalf("load checkpoint: %v", err)
	}
	if !ok {
		t.Fatalf("expected checkpoint to exist")
	}
	if cp.LastEventID != event.ID {
		t.Fatalf("expected checkpoint last_event_id=%s, got %s", event.ID, cp.LastEventID)
	}
}

type failingProcessorStore struct {
	claimedRows      []OutboxRecord
	markFailedCall   int
	deadLetter       bool
	nextRetryAt      *time.Time
}

func (s *failingProcessorStore) ClaimPendingForProcessing(workerName string, leaseUntil time.Time, limit int, now time.Time) ([]OutboxRecord, error) {
	return s.claimedRows, nil
}

func (s *failingProcessorStore) GetEvent(eventID string) (EventRecord, error) {
	return EventRecord{}, errors.New("event not found")
}

func (s *failingProcessorStore) MarkProcessed(outboxID int64, now time.Time) error {
	return nil
}

func (s *failingProcessorStore) MarkFailed(outboxID int64, lastError string, deadLetter bool, nextRetryAt *time.Time, now time.Time) error {
	s.markFailedCall++
	s.deadLetter = deadLetter
	s.nextRetryAt = nextRetryAt
	return nil
}

func (s *failingProcessorStore) SaveCheckpoint(checkpoint ProcessorCheckpoint) error {
	return nil
}

func (s *failingProcessorStore) LoadCheckpoint(workerName string) (ProcessorCheckpoint, bool, error) {
	return ProcessorCheckpoint{}, false, nil
}

func (s *failingProcessorStore) GetFreshness(workerName string, now time.Time) (ProcessorFreshness, error) {
	return ProcessorFreshness{}, nil
}

func TestProcessorMarksDeadLetterWhenAttemptsExceeded(t *testing.T) {
	store := &failingProcessorStore{
		claimedRows: []OutboxRecord{{
			ID:      1,
			EventID: "evt_dead",
			Attempt: 4,
		}},
	}

	processor := NewProcessor(store, nil, ProcessorConfig{MaxAttempts: 3, RetryPolicy: DefaultRetryPolicy()})
	if err := processor.processBatch(context.Background()); err != nil {
		t.Fatalf("process batch: %v", err)
	}

	if store.markFailedCall != 1 {
		t.Fatalf("expected mark failed to be called once, got %d", store.markFailedCall)
	}
	if !store.deadLetter {
		t.Fatalf("expected dead letter flag to be true when attempts are exceeded")
	}
	if store.nextRetryAt != nil {
		t.Fatalf("expected no retry timestamp once dead lettered")
	}
}

func TestProcessorSchedulesRetryWhenAttemptsRemain(t *testing.T) {
	store := &failingProcessorStore{
		claimedRows: []OutboxRecord{{
			ID:      2,
			EventID: "evt_retry",
			Attempt: 1,
		}},
	}

	processor := NewProcessor(store, nil, ProcessorConfig{MaxAttempts: 3, RetryPolicy: RetryPolicy{BaseDelay: 2 * time.Second, MaxDelay: 10 * time.Second}})
	if err := processor.processBatch(context.Background()); err != nil {
		t.Fatalf("process batch: %v", err)
	}

	if store.markFailedCall != 1 {
		t.Fatalf("expected mark failed to be called once, got %d", store.markFailedCall)
	}
	if store.deadLetter {
		t.Fatalf("expected retryable failure, not dead letter")
	}
	if store.nextRetryAt == nil {
		t.Fatalf("expected retry timestamp to be scheduled")
	}
	if !store.nextRetryAt.After(time.Now().UTC()) {
		t.Fatalf("expected retry timestamp to be in the future, got %s", store.nextRetryAt.Format(time.RFC3339Nano))
	}
}

type cleanupProcessorStore struct {
	processedCalls int
	deadCalls      int
	eventCalls     int
}

func (s *cleanupProcessorStore) ClaimPendingForProcessing(workerName string, leaseUntil time.Time, limit int, now time.Time) ([]OutboxRecord, error) {
	return nil, nil
}

func (s *cleanupProcessorStore) GetEvent(eventID string) (EventRecord, error) {
	return EventRecord{}, nil
}

func (s *cleanupProcessorStore) MarkProcessed(outboxID int64, now time.Time) error {
	return nil
}

func (s *cleanupProcessorStore) MarkFailed(outboxID int64, lastError string, deadLetter bool, nextRetryAt *time.Time, now time.Time) error {
	return nil
}

func (s *cleanupProcessorStore) SaveCheckpoint(checkpoint ProcessorCheckpoint) error {
	return nil
}

func (s *cleanupProcessorStore) LoadCheckpoint(workerName string) (ProcessorCheckpoint, bool, error) {
	return ProcessorCheckpoint{}, false, nil
}

func (s *cleanupProcessorStore) GetFreshness(workerName string, now time.Time) (ProcessorFreshness, error) {
	return ProcessorFreshness{}, nil
}

func (s *cleanupProcessorStore) CleanupProcessedOutbox(olderThan time.Time) (int64, error) {
	s.processedCalls++
	return 1, nil
}

func (s *cleanupProcessorStore) CleanupDeadLetters(olderThan time.Time) (int64, error) {
	s.deadCalls++
	return 1, nil
}

func (s *cleanupProcessorStore) CleanupEvents(olderThan time.Time) (int64, error) {
	s.eventCalls++
	return 1, nil
}

func TestProcessorRunScheduledCleanupHonorsCadence(t *testing.T) {
	store := &cleanupProcessorStore{}

	now := time.Date(2026, 6, 17, 10, 0, 0, 0, time.UTC)
	processor := NewProcessor(store, nil, ProcessorConfig{
		CleanupInterval:          10 * time.Minute,
		ProcessedOutboxRetention: 24 * time.Hour,
		DeadLetterRetention:      7 * 24 * time.Hour,
		EventRetention:           14 * 24 * time.Hour,
	})
	processor.nowFn = func() time.Time { return now }

	if err := processor.runScheduledCleanup(); err != nil {
		t.Fatalf("first cleanup run failed: %v", err)
	}
	if store.processedCalls != 1 || store.deadCalls != 1 || store.eventCalls != 1 {
		t.Fatalf("expected cleanup methods to run once, got processed=%d dead=%d events=%d", store.processedCalls, store.deadCalls, store.eventCalls)
	}

	// second run within interval should not invoke cleanup methods again
	now = now.Add(5 * time.Minute)
	if err := processor.runScheduledCleanup(); err != nil {
		t.Fatalf("second cleanup run failed: %v", err)
	}
	if store.processedCalls != 1 || store.deadCalls != 1 || store.eventCalls != 1 {
		t.Fatalf("expected cleanup cadence gate to skip second run")
	}

	// third run beyond interval should execute again
	now = now.Add(6 * time.Minute)
	if err := processor.runScheduledCleanup(); err != nil {
		t.Fatalf("third cleanup run failed: %v", err)
	}
	if store.processedCalls != 2 || store.deadCalls != 2 || store.eventCalls != 2 {
		t.Fatalf("expected cleanup methods to run twice total, got processed=%d dead=%d events=%d", store.processedCalls, store.deadCalls, store.eventCalls)
	}
}

type projectionOptimizedProcessorStore struct {
	rows              []OutboxRecord
	events            map[string]EventRecord
	markProcessedCall int
	checkpointCall    int
}

func (s *projectionOptimizedProcessorStore) ClaimPendingForProcessing(workerName string, leaseUntil time.Time, limit int, now time.Time) ([]OutboxRecord, error) {
	return s.rows, nil
}

func (s *projectionOptimizedProcessorStore) GetEvent(eventID string) (EventRecord, error) {
	event, ok := s.events[eventID]
	if !ok {
		return EventRecord{}, errors.New("event not found")
	}
	return event, nil
}

func (s *projectionOptimizedProcessorStore) MarkProcessed(outboxID int64, now time.Time) error {
	s.markProcessedCall++
	return nil
}

func (s *projectionOptimizedProcessorStore) MarkFailed(outboxID int64, lastError string, deadLetter bool, nextRetryAt *time.Time, now time.Time) error {
	return nil
}

func (s *projectionOptimizedProcessorStore) SaveCheckpoint(checkpoint ProcessorCheckpoint) error {
	s.checkpointCall++
	return nil
}

func (s *projectionOptimizedProcessorStore) LoadCheckpoint(workerName string) (ProcessorCheckpoint, bool, error) {
	return ProcessorCheckpoint{}, false, nil
}

func (s *projectionOptimizedProcessorStore) GetFreshness(workerName string, now time.Time) (ProcessorFreshness, error) {
	return ProcessorFreshness{}, nil
}

type countingProjectionRebuilder struct {
	sessionRebuildCalls  int
	platformRebuildCalls int
}

func (c *countingProjectionRebuilder) RebuildSessionProjection(sessionID string, sessionSnap session.Snapshot, items []programitem.Snapshot, now time.Time) error {
	c.sessionRebuildCalls++
	return nil
}

func (c *countingProjectionRebuilder) RebuildPlatformProjection(sessions []session.Snapshot, itemsBySession map[string][]programitem.Snapshot, now time.Time) error {
	c.platformRebuildCalls++
	return nil
}

func TestProcessorRebuildProjectionsMemoizesBatchInputs(t *testing.T) {
	store := &projectionOptimizedProcessorStore{
		rows: []OutboxRecord{
			{ID: 1, EventID: "evt_1", Attempt: 1},
			{ID: 2, EventID: "evt_2", Attempt: 1},
		},
		events: map[string]EventRecord{
			"evt_1": {ID: "evt_1", SessionID: "sess_a", EventKey: "SESSION_STARTED"},
			"evt_2": {ID: "evt_2", SessionID: "sess_a", EventKey: "PROGRAM_ITEM_PAUSED"},
		},
	}

	rebuilder := &countingProjectionRebuilder{}
	sessionSnapshotCalls := 0
	listSessionCalls := 0
	programItemCalls := map[string]int{}

	processor := NewProcessor(store, nil, ProcessorConfig{
		MaxAttempts:       3,
		RetryPolicy:       DefaultRetryPolicy(),
		ProjectionBuilder: rebuilder,
		GetSessionSnapshot: func(sessionID string) (session.Snapshot, error) {
			sessionSnapshotCalls++
			return session.Snapshot{ID: sessionID}, nil
		},
		ListProgramItemSnapshots: func(sessionID string) ([]programitem.Snapshot, error) {
			programItemCalls[sessionID]++
			return []programitem.Snapshot{}, nil
		},
		ListSessionSnapshots: func() ([]session.Snapshot, error) {
			listSessionCalls++
			return []session.Snapshot{{ID: "sess_a"}, {ID: "sess_b"}}, nil
		},
	})

	if err := processor.processBatch(context.Background()); err != nil {
		t.Fatalf("process batch: %v", err)
	}

	if store.markProcessedCall != 2 {
		t.Fatalf("expected 2 processed rows, got %d", store.markProcessedCall)
	}
	if store.checkpointCall != 2 {
		t.Fatalf("expected 2 checkpoints, got %d", store.checkpointCall)
	}
	if rebuilder.sessionRebuildCalls != 2 {
		t.Fatalf("expected 2 session rebuilds, got %d", rebuilder.sessionRebuildCalls)
	}
	if rebuilder.platformRebuildCalls != 2 {
		t.Fatalf("expected 2 platform rebuilds, got %d", rebuilder.platformRebuildCalls)
	}
	if listSessionCalls != 1 {
		t.Fatalf("expected session list to be memoized to 1 call, got %d", listSessionCalls)
	}
	if sessionSnapshotCalls != 1 {
		t.Fatalf("expected session snapshot to be memoized to 1 call, got %d", sessionSnapshotCalls)
	}
	if programItemCalls["sess_a"] != 1 {
		t.Fatalf("expected session program items for sess_a to be memoized to 1 call, got %d", programItemCalls["sess_a"])
	}
	if programItemCalls["sess_b"] != 1 {
		t.Fatalf("expected session program items for sess_b to be memoized to 1 call, got %d", programItemCalls["sess_b"])
	}
}

func TestProcessorSkipsPlatformRebuildForNonImpactingEvent(t *testing.T) {
	store := &projectionOptimizedProcessorStore{
		rows: []OutboxRecord{{ID: 1, EventID: "evt_heartbeat", Attempt: 1}},
		events: map[string]EventRecord{
			"evt_heartbeat": {ID: "evt_heartbeat", SessionID: "sess_a", EventKey: "SESSION_HEARTBEAT"},
		},
	}

	rebuilder := &countingProjectionRebuilder{}
	listSessionCalls := 0

	processor := NewProcessor(store, nil, ProcessorConfig{
		MaxAttempts:       3,
		RetryPolicy:       DefaultRetryPolicy(),
		ProjectionBuilder: rebuilder,
		GetSessionSnapshot: func(sessionID string) (session.Snapshot, error) {
			return session.Snapshot{ID: sessionID}, nil
		},
		ListProgramItemSnapshots: func(sessionID string) ([]programitem.Snapshot, error) {
			return []programitem.Snapshot{}, nil
		},
		ListSessionSnapshots: func() ([]session.Snapshot, error) {
			listSessionCalls++
			return []session.Snapshot{{ID: "sess_a"}}, nil
		},
	})

	if err := processor.processBatch(context.Background()); err != nil {
		t.Fatalf("process batch: %v", err)
	}

	if rebuilder.sessionRebuildCalls != 1 {
		t.Fatalf("expected 1 session rebuild, got %d", rebuilder.sessionRebuildCalls)
	}
	if rebuilder.platformRebuildCalls != 0 {
		t.Fatalf("expected 0 platform rebuilds for non-impacting event, got %d", rebuilder.platformRebuildCalls)
	}
	if listSessionCalls != 0 {
		t.Fatalf("expected no session list calls when platform rebuild is skipped, got %d", listSessionCalls)
	}
}
