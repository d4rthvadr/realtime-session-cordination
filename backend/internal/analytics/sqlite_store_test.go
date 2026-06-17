package analytics

import (
	"path/filepath"
	"testing"
	"time"
)

func TestSqliteStoreAppendEnqueueAndCheckpoint(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "analytics-test.db")

	store, err := NewSqliteStore(dbPath)
	if err != nil {
		t.Fatalf("new sqlite store: %v", err)
	}

	now := time.Date(2026, 6, 3, 12, 0, 0, 0, time.UTC)
	eventID := "evt_1"

	err = store.AppendEvent(EventRecord{
		ID:          eventID,
		SessionID:   "sess_1",
		EventKey:    "SESSION_STARTED",
		OccurredAt:  now,
		IngestedAt:  now,
		Source:      EventSourceServer,
		PayloadJSON: []byte(`{"type":"SESSION_STARTED"}`),
	})
	if err != nil {
		t.Fatalf("append event: %v", err)
	}

	err = store.Enqueue(eventID, now)
	if err != nil {
		t.Fatalf("enqueue event: %v", err)
	}

	err = store.SaveCheckpoint(ProcessorCheckpoint{
		WorkerName:  "analytics_processor",
		LastEventID: eventID,
		UpdatedAt:   now,
	})
	if err != nil {
		t.Fatalf("save checkpoint: %v", err)
	}

	var eventsCount int
	if err = store.db.QueryRow(`SELECT COUNT(*) FROM analytics_events`).Scan(&eventsCount); err != nil {
		t.Fatalf("count analytics_events: %v", err)
	}
	if eventsCount != 1 {
		t.Fatalf("expected 1 analytics event, got %d", eventsCount)
	}

	var outboxState string
	if err = store.db.QueryRow(`SELECT state FROM analytics_outbox WHERE event_id = ?`, eventID).Scan(&outboxState); err != nil {
		t.Fatalf("read outbox state: %v", err)
	}
	if outboxState != OutboxStatePending {
		t.Fatalf("expected outbox state %q, got %q", OutboxStatePending, outboxState)
	}

	var checkpointEventID string
	if err = store.db.QueryRow(`SELECT last_event_id FROM analytics_checkpoints WHERE worker_name = ?`, "analytics_processor").Scan(&checkpointEventID); err != nil {
		t.Fatalf("read checkpoint: %v", err)
	}
	if checkpointEventID != eventID {
		t.Fatalf("expected checkpoint last_event_id=%s, got %s", eventID, checkpointEventID)
	}
}

func TestSqliteStoreClaimHonorsNextRetryAt(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "analytics-test-retry.db")

	store, err := NewSqliteStore(dbPath)
	if err != nil {
		t.Fatalf("new sqlite store: %v", err)
	}

	now := time.Date(2026, 6, 3, 12, 0, 0, 0, time.UTC)
	eventID := "evt_retry_1"

	err = store.AppendEventAndEnqueue(EventRecord{
		ID:          eventID,
		SessionID:   "sess_1",
		EventKey:    "SESSION_STARTED",
		OccurredAt:  now,
		IngestedAt:  now,
		Source:      EventSourceServer,
		PayloadJSON: []byte(`{"type":"SESSION_STARTED"}`),
	}, now)
	if err != nil {
		t.Fatalf("append+enqueue: %v", err)
	}

	var outboxID int64
	if err = store.db.QueryRow(`SELECT id FROM analytics_outbox WHERE event_id = ?`, eventID).Scan(&outboxID); err != nil {
		t.Fatalf("read outbox id: %v", err)
	}

	retryAt := now.Add(10 * time.Minute)
	if err = store.MarkFailed(outboxID, "retry later", false, &retryAt, now); err != nil {
		t.Fatalf("mark failed with retry: %v", err)
	}

	rows, err := store.ClaimPendingForProcessing("analytics_processor", now.Add(15*time.Second), 10, now)
	if err != nil {
		t.Fatalf("claim pending before retry due: %v", err)
	}
	if len(rows) != 0 {
		t.Fatalf("expected no rows to be claimable before retry window, got %d", len(rows))
	}

	rows, err = store.ClaimPendingForProcessing("analytics_processor", now.Add(15*time.Minute), 10, retryAt.Add(time.Second))
	if err != nil {
		t.Fatalf("claim pending after retry due: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected one row to be claimable after retry window, got %d", len(rows))
	}
	if rows[0].Attempt != 1 {
		t.Fatalf("expected claim attempt to increment to 1, got %d", rows[0].Attempt)
	}
}
