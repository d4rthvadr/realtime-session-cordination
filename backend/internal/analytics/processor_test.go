package analytics

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"
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
