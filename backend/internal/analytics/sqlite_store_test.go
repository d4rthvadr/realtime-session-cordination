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

func TestSqliteStoreDeadLetterListGetAndRetry(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "analytics-test-dlq.db")

	store, err := NewSqliteStore(dbPath)
	if err != nil {
		t.Fatalf("new sqlite store: %v", err)
	}

	now := time.Date(2026, 6, 3, 15, 0, 0, 0, time.UTC)
	eventID := "evt_dlq_1"

	err = store.AppendEventAndEnqueue(EventRecord{
		ID:          eventID,
		SessionID:   "sess_dlq",
		EventKey:    "PROGRAM_ITEM_ENDED",
		OccurredAt:  now,
		IngestedAt:  now,
		Source:      EventSourceServer,
		PayloadJSON: []byte(`{"type":"PROGRAM_ITEM_ENDED"}`),
	}, now)
	if err != nil {
		t.Fatalf("append+enqueue: %v", err)
	}

	claimed, err := store.ClaimPendingForProcessing("analytics_processor", now.Add(15*time.Second), 10, now)
	if err != nil {
		t.Fatalf("claim pending: %v", err)
	}
	if len(claimed) != 1 {
		t.Fatalf("expected one claimed row, got %d", len(claimed))
	}
	outboxID := claimed[0].ID

	err = store.MarkFailed(outboxID, "projection rebuild failed", true, nil, now.Add(2*time.Second))
	if err != nil {
		t.Fatalf("mark failed dead-letter: %v", err)
	}

	rows, err := store.ListDeadLetters(20, 0)
	if err != nil {
		t.Fatalf("list dead letters: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected one dead-letter row, got %d", len(rows))
	}
	if rows[0].OutboxID != outboxID {
		t.Fatalf("expected dead-letter outbox id %d, got %d", outboxID, rows[0].OutboxID)
	}
	if rows[0].LastError == "" {
		t.Fatalf("expected dead-letter error to be set")
	}

	detail, found, err := store.GetDeadLetter(outboxID)
	if err != nil {
		t.Fatalf("get dead letter: %v", err)
	}
	if !found {
		t.Fatalf("expected dead-letter row to be found")
	}
	if detail.EventID != eventID {
		t.Fatalf("expected eventID %s, got %s", eventID, detail.EventID)
	}

	err = store.RetryDeadLetter(outboxID, now.Add(3*time.Second))
	if err != nil {
		t.Fatalf("retry dead letter: %v", err)
	}

	rows, err = store.ListDeadLetters(20, 0)
	if err != nil {
		t.Fatalf("list dead letters after retry: %v", err)
	}
	if len(rows) != 0 {
		t.Fatalf("expected dead-letter queue to be empty after retry, got %d", len(rows))
	}

	claimedAgain, err := store.ClaimPendingForProcessing("analytics_processor", now.Add(30*time.Second), 10, now.Add(4*time.Second))
	if err != nil {
		t.Fatalf("claim pending after retry reset: %v", err)
	}
	if len(claimedAgain) != 1 {
		t.Fatalf("expected retried row to be claimable, got %d", len(claimedAgain))
	}
	if claimedAgain[0].Attempt != 1 {
		t.Fatalf("expected attempt to restart at 1 after retry, got %d", claimedAgain[0].Attempt)
	}
}

func TestSqliteStoreCleanupRetentionMethods(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "analytics-test-cleanup.db")

	store, err := NewSqliteStore(dbPath)
	if err != nil {
		t.Fatalf("new sqlite store: %v", err)
	}

	now := time.Date(2026, 6, 4, 10, 0, 0, 0, time.UTC)
	old := now.Add(-72 * time.Hour)

	oldProcessedEventID := "evt_old_processed"
	if err = store.AppendEventAndEnqueue(EventRecord{
		ID:          oldProcessedEventID,
		SessionID:   "sess_cleanup",
		EventKey:    "SESSION_STARTED",
		OccurredAt:  old,
		IngestedAt:  old,
		Source:      EventSourceServer,
		PayloadJSON: []byte(`{"type":"SESSION_STARTED"}`),
	}, old); err != nil {
		t.Fatalf("append old processed event: %v", err)
	}

	var oldProcessedOutboxID int64
	if err = store.db.QueryRow(`SELECT id FROM analytics_outbox WHERE event_id = ?`, oldProcessedEventID).Scan(&oldProcessedOutboxID); err != nil {
		t.Fatalf("read old processed outbox id: %v", err)
	}
	if _, err = store.db.Exec(`UPDATE analytics_outbox SET state = ?, updated_at = ? WHERE id = ?`, OutboxStateProcessed, old.Format(time.RFC3339Nano), oldProcessedOutboxID); err != nil {
		t.Fatalf("seed old processed row: %v", err)
	}

	oldDeadEventID := "evt_old_dead"
	if err = store.AppendEventAndEnqueue(EventRecord{
		ID:          oldDeadEventID,
		SessionID:   "sess_cleanup",
		EventKey:    "SESSION_PAUSED",
		OccurredAt:  old,
		IngestedAt:  old,
		Source:      EventSourceServer,
		PayloadJSON: []byte(`{"type":"SESSION_PAUSED"}`),
	}, old); err != nil {
		t.Fatalf("append old dead-letter event: %v", err)
	}

	var oldDeadOutboxID int64
	if err = store.db.QueryRow(`SELECT id FROM analytics_outbox WHERE event_id = ?`, oldDeadEventID).Scan(&oldDeadOutboxID); err != nil {
		t.Fatalf("read old dead outbox id: %v", err)
	}
	if _, err = store.db.Exec(`UPDATE analytics_outbox SET state = ?, updated_at = ? WHERE id = ?`, OutboxStateDeadLetter, old.Format(time.RFC3339Nano), oldDeadOutboxID); err != nil {
		t.Fatalf("seed old dead-letter row: %v", err)
	}

	// This event has no outbox row and should be removable by event cleanup.
	orphanEventID := "evt_orphan_old"
	if err = store.AppendEvent(EventRecord{
		ID:          orphanEventID,
		SessionID:   "sess_cleanup",
		EventKey:    "SESSION_RESUMED",
		OccurredAt:  old,
		IngestedAt:  old,
		Source:      EventSourceServer,
		PayloadJSON: []byte(`{"type":"SESSION_RESUMED"}`),
	}); err != nil {
		t.Fatalf("append orphan event: %v", err)
	}

	processedDeleted, err := store.CleanupProcessedOutbox(now.Add(-24 * time.Hour))
	if err != nil {
		t.Fatalf("cleanup processed outbox: %v", err)
	}
	if processedDeleted != 1 {
		t.Fatalf("expected 1 processed row deleted, got %d", processedDeleted)
	}

	deadDeleted, err := store.CleanupDeadLetters(now.Add(-24 * time.Hour))
	if err != nil {
		t.Fatalf("cleanup dead letters: %v", err)
	}
	if deadDeleted != 1 {
		t.Fatalf("expected 1 dead-letter row deleted, got %d", deadDeleted)
	}

	eventsDeleted, err := store.CleanupEvents(now.Add(-24 * time.Hour))
	if err != nil {
		t.Fatalf("cleanup events: %v", err)
	}
	if eventsDeleted < 1 {
		t.Fatalf("expected at least 1 event deleted, got %d", eventsDeleted)
	}

	var orphanCount int
	if err = store.db.QueryRow(`SELECT COUNT(*) FROM analytics_events WHERE id = ?`, orphanEventID).Scan(&orphanCount); err != nil {
		t.Fatalf("read orphan event count: %v", err)
	}
	if orphanCount != 0 {
		t.Fatalf("expected orphan old event to be deleted")
	}
}

func TestSqliteStoreGetFreshnessIncludesOperationalMetadata(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "analytics-test-freshness-metadata.db")

	store, err := NewSqliteStore(dbPath)
	if err != nil {
		t.Fatalf("new sqlite store: %v", err)
	}

	now := time.Date(2026, 6, 17, 11, 0, 0, 0, time.UTC)
	retryEventID := "evt_retry_due"
	deadEventID := "evt_dead"

	err = store.AppendEventAndEnqueue(EventRecord{
		ID:          retryEventID,
		SessionID:   "sess_ops",
		EventKey:    "SESSION_PAUSED",
		OccurredAt:  now.Add(-2 * time.Minute),
		IngestedAt:  now.Add(-2 * time.Minute),
		Source:      EventSourceServer,
		PayloadJSON: []byte(`{"type":"SESSION_PAUSED"}`),
	}, now.Add(-2*time.Minute))
	if err != nil {
		t.Fatalf("append retry event: %v", err)
	}

	var retryOutboxID int64
	if err = store.db.QueryRow(`SELECT id FROM analytics_outbox WHERE event_id = ?`, retryEventID).Scan(&retryOutboxID); err != nil {
		t.Fatalf("read retry outbox id: %v", err)
	}
	retryAt := now.Add(-30 * time.Second)
	if err = store.MarkFailed(retryOutboxID, "retry me", false, &retryAt, now.Add(-1*time.Minute)); err != nil {
		t.Fatalf("mark retry event failed: %v", err)
	}

	err = store.AppendEventAndEnqueue(EventRecord{
		ID:          deadEventID,
		SessionID:   "sess_ops",
		EventKey:    "SESSION_ENDED",
		OccurredAt:  now.Add(-3 * time.Minute),
		IngestedAt:  now.Add(-3 * time.Minute),
		Source:      EventSourceServer,
		PayloadJSON: []byte(`{"type":"SESSION_ENDED"}`),
	}, now.Add(-3*time.Minute))
	if err != nil {
		t.Fatalf("append dead event: %v", err)
	}

	var deadOutboxID int64
	if err = store.db.QueryRow(`SELECT id FROM analytics_outbox WHERE event_id = ?`, deadEventID).Scan(&deadOutboxID); err != nil {
		t.Fatalf("read dead outbox id: %v", err)
	}
	if err = store.MarkFailed(deadOutboxID, "dead letter", true, nil, now.Add(-2*time.Minute)); err != nil {
		t.Fatalf("mark dead event failed: %v", err)
	}

	freshness, err := store.GetFreshness("analytics_processor", now)
	if err != nil {
		t.Fatalf("get freshness: %v", err)
	}

	if freshness.RetryDueCount != 1 {
		t.Fatalf("expected retry due count 1, got %d", freshness.RetryDueCount)
	}
	if freshness.DeadLetterCount != 1 {
		t.Fatalf("expected dead-letter count 1, got %d", freshness.DeadLetterCount)
	}
	if freshness.RetryLagSeconds <= 0 {
		t.Fatalf("expected retry lag to be positive, got %d", freshness.RetryLagSeconds)
	}
}
