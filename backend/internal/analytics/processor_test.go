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
	claimedRows    []OutboxRecord
	markFailedCall int
	deadLetter     bool
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

func (s *failingProcessorStore) MarkFailed(outboxID int64, lastError string, deadLetter bool, now time.Time) error {
	s.markFailedCall++
	s.deadLetter = deadLetter
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

	processor := NewProcessor(store, nil, ProcessorConfig{MaxAttempts: 3})
	if err := processor.processBatch(context.Background()); err != nil {
		t.Fatalf("process batch: %v", err)
	}

	if store.markFailedCall != 1 {
		t.Fatalf("expected mark failed to be called once, got %d", store.markFailedCall)
	}
	if !store.deadLetter {
		t.Fatalf("expected dead letter flag to be true when attempts are exceeded")
	}
}
