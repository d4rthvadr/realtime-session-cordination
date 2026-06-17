package sessionlog

import "testing"

func TestListBySessionRejectsInvalidEventType(t *testing.T) {
	m := NewManager(NewMemoryStore())

	_, err := m.ListBySession("sess_1", ListOptions{EventType: EventType("NOT_A_REAL_EVENT")})
	if err == nil {
		t.Fatalf("expected error for invalid event type")
	}
}

func TestListBySessionRejectsInvalidEntityType(t *testing.T) {
	m := NewManager(NewMemoryStore())

	_, err := m.ListBySession("sess_1", ListOptions{EntityType: "unknown"})
	if err == nil {
		t.Fatalf("expected error for invalid entity type")
	}
}

func TestListBySessionEntityFilter(t *testing.T) {
	store := NewMemoryStore()
	m := NewManager(store)

	if _, err := m.Append(AppendInput{SessionID: "sess_1", EventType: SessionStarted}); err != nil {
		t.Fatalf("append session event: %v", err)
	}

	if _, err := m.Append(AppendInput{SessionID: "sess_1", EventType: ProgramItemStarted}); err != nil {
		t.Fatalf("append program item event: %v", err)
	}

	if _, err := m.Append(AppendInput{SessionID: "sess_1", EventType: CascadeProgramItemPausedBySession}); err != nil {
		t.Fatalf("append cascade event: %v", err)
	}

	logs, err := m.ListBySession("sess_1", ListOptions{EntityType: EntityProgramItem})
	if err != nil {
		t.Fatalf("list by entity: %v", err)
	}

	if len(logs) != 1 {
		t.Fatalf("expected 1 program_item log, got %d", len(logs))
	}

	if logs[0].EventType != ProgramItemStarted.String() {
		t.Fatalf("unexpected event type: %s", logs[0].EventType)
	}
}
