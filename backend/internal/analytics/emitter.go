package analytics

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// Emitter emits analytics events to the persistence layer with atomic event+outbox creation.
type Emitter struct {
	ingestionStore IngestionStore
}

func NewEmitter(ingestionStore IngestionStore) *Emitter {
	return &Emitter{
		ingestionStore: ingestionStore,
	}
}

// EmitSessionEvent emits a server-side session event and enqueues it.
func (e *Emitter) EmitSessionEvent(sessionID string, eventKey string, payload any) error {
	if sessionID == "" {
		return fmt.Errorf("session id is required")
	}
	if eventKey == "" {
		return fmt.Errorf("event key is required")
	}
	if e == nil || e.ingestionStore == nil {
		return fmt.Errorf("analytics ingestion store is not configured")
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal event payload: %w", err)
	}

	eventID := newEventID()
	now := time.Now().UTC()

	record := EventRecord{
		ID:          eventID,
		SessionID:   sessionID,
		EventKey:    eventKey,
		OccurredAt:  now,
		IngestedAt:  now,
		Source:      EventSourceServer,
		PayloadJSON: payloadJSON,
	}

	if err = e.ingestionStore.AppendEventAndEnqueue(record, now); err != nil {
		return fmt.Errorf("failed to atomically append+enqueue event: %w", err)
	}

	return nil
}

// EmitProgramItemEvent emits a server-side program-item event and enqueues it.
func (e *Emitter) EmitProgramItemEvent(sessionID string, programItemID string, eventKey string, payload any) error {
	if sessionID == "" {
		return fmt.Errorf("session id is required")
	}
	if programItemID == "" {
		return fmt.Errorf("program item id is required")
	}
	if eventKey == "" {
		return fmt.Errorf("event key is required")
	}
	if e == nil || e.ingestionStore == nil {
		return fmt.Errorf("analytics ingestion store is not configured")
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal event payload: %w", err)
	}

	eventID := newEventID()
	now := time.Now().UTC()
	pidCopy := programItemID

	record := EventRecord{
		ID:            eventID,
		SessionID:     sessionID,
		ProgramItemID: &pidCopy,
		EventKey:      eventKey,
		OccurredAt:    now,
		IngestedAt:    now,
		Source:        EventSourceServer,
		PayloadJSON:   payloadJSON,
	}

	if err = e.ingestionStore.AppendEventAndEnqueue(record, now); err != nil {
		return fmt.Errorf("failed to atomically append+enqueue event: %w", err)
	}

	return nil
}

func newEventID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp-based ID if random fails
		return "evt_" + fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return "evt_" + hex.EncodeToString(b)
}
