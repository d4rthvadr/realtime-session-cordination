package analytics

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"
)

type testEventStore struct {
	events []EventRecord
}

func (s *testEventStore) AppendEvent(record EventRecord) error {
	s.events = append(s.events, record)
	return nil
}

type testOutboxStore struct {
	enqueued []string
}

func (s *testOutboxStore) Enqueue(eventID string, now time.Time) error {
	s.enqueued = append(s.enqueued, eventID)
	return nil
}

type testCheckpointStore struct{}

func (s *testCheckpointStore) SaveCheckpoint(checkpoint ProcessorCheckpoint) error {
	return nil
}

type testIngestionStore struct {
	events   []EventRecord
	enqueued []string
}

func (s *testIngestionStore) AppendEventAndEnqueue(record EventRecord, now time.Time) error {
	s.events = append(s.events, record)
	s.enqueued = append(s.enqueued, record.ID)
	return nil
}

func TestEmitterSessionEvent(t *testing.T) {
	ingestionStore := &testIngestionStore{}
	emitter := NewEmitter(ingestionStore)

	payload := map[string]any{
		"type":  "SESSION_STARTED",
		"title": "Demo",
	}

	err := emitter.EmitSessionEvent("sess_1", "SESSION_STARTED", payload)
	if err != nil {
		t.Fatalf("emit session event: %v", err)
	}

	if len(ingestionStore.events) != 1 {
		t.Fatalf("expected 1 event appended, got %d", len(ingestionStore.events))
	}

	if ingestionStore.events[0].SessionID != "sess_1" {
		t.Fatalf("expected session id sess_1, got %s", ingestionStore.events[0].SessionID)
	}

	if ingestionStore.events[0].EventKey != "SESSION_STARTED" {
		t.Fatalf("expected event key SESSION_STARTED, got %s", ingestionStore.events[0].EventKey)
	}

	if len(ingestionStore.enqueued) != 1 {
		t.Fatalf("expected 1 event enqueued, got %d", len(ingestionStore.enqueued))
	}

	if ingestionStore.enqueued[0] != ingestionStore.events[0].ID {
		t.Fatalf("enqueued event id does not match appended event")
	}

	var payloadParsed map[string]any
	if err = json.Unmarshal(ingestionStore.events[0].PayloadJSON, &payloadParsed); err != nil {
		t.Fatalf("parse payload: %v", err)
	}

	if payloadParsed["type"] != "SESSION_STARTED" {
		t.Fatalf("expected payload type SESSION_STARTED, got %v", payloadParsed["type"])
	}
}

func TestEmitterProgramItemEvent(t *testing.T) {
	ingestionStore := &testIngestionStore{}
	emitter := NewEmitter(ingestionStore)

	payload := map[string]any{
		"type": "PROGRAM_ITEM_STARTED",
		"itemId": "pi_abc",
	}

	err := emitter.EmitProgramItemEvent("sess_1", "pi_abc", "PROGRAM_ITEM_STARTED", payload)
	if err != nil {
		t.Fatalf("emit program item event: %v", err)
	}

	if len(ingestionStore.events) != 1 {
		t.Fatalf("expected 1 event appended, got %d", len(ingestionStore.events))
	}

	if ingestionStore.events[0].ProgramItemID == nil || *ingestionStore.events[0].ProgramItemID != "pi_abc" {
		t.Fatalf("expected program item id pi_abc")
	}

	if len(ingestionStore.enqueued) != 1 {
		t.Fatalf("expected 1 event enqueued, got %d", len(ingestionStore.enqueued))
	}
}

func TestEmitterWithSqliteStore(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "emitter-test.db")
	store, err := NewSqliteStore(dbPath)
	if err != nil {
		t.Fatalf("new sqlite store: %v", err)
	}

	emitter := NewEmitter(store)

	payload := map[string]any{
		"type":     "SESSION_STARTED",
		"title":    "Integration Test",
		"duration": 3600,
	}

	err = emitter.EmitSessionEvent("sess_integration", "SESSION_STARTED", payload)
	if err != nil {
		t.Fatalf("emit session event: %v", err)
	}

	var count int
	if err = store.db.QueryRow(`SELECT COUNT(*) FROM analytics_events WHERE session_id = ?`, "sess_integration").Scan(&count); err != nil {
		t.Fatalf("count events: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 event in db, got %d", count)
	}

	var outboxCount int
	if err = store.db.QueryRow(`SELECT COUNT(*) FROM analytics_outbox WHERE state = ?`, OutboxStatePending).Scan(&outboxCount); err != nil {
		t.Fatalf("count outbox: %v", err)
	}
	if outboxCount != 1 {
		t.Fatalf("expected 1 pending outbox record, got %d", outboxCount)
	}
}
