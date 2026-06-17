package analytics

import (
	"testing"
	"time"

	"realtime-session-coordination/backend/internal/programitem"
	"realtime-session-coordination/backend/internal/session"
)

type projectionStoreSpy struct {
	sessionProjection  SessionProjection
	platformProjection PlatformProjection
	sessionUpserts     int
	platformUpserts    int
}

func (s *projectionStoreSpy) UpsertSessionProjection(p SessionProjection) error {
	s.sessionProjection = p
	s.sessionUpserts++
	return nil
}

func (s *projectionStoreSpy) GetSessionProjection(sessionID string) (SessionProjection, bool, error) {
	return SessionProjection{}, false, nil
}

func (s *projectionStoreSpy) UpsertPlatformProjection(p PlatformProjection) error {
	s.platformProjection = p
	s.platformUpserts++
	return nil
}

func (s *projectionStoreSpy) GetPlatformProjection() (PlatformProjection, bool, error) {
	return PlatformProjection{}, false, nil
}

func TestProjectionBuilderRebuildSessionProjectionUpsertsMappedProjection(t *testing.T) {
	store := &projectionStoreSpy{}
	builder := NewProjectionBuilder(store, NewManager())

	now := time.Date(2026, 6, 15, 9, 0, 0, 0, time.UTC)
	endedOnTime := 12

	err := builder.RebuildSessionProjection(
		"sess_1",
		session.Snapshot{ID: "sess_1", Status: session.StatusEnded, DurationSeconds: 3600},
		[]programitem.Snapshot{{Status: programitem.StatusEnded, RuntimeDurationSeconds: 900, EndedRemainingSeconds: &endedOnTime}},
		now,
	)
	if err != nil {
		t.Fatalf("rebuild session projection: %v", err)
	}

	if store.sessionUpserts != 1 {
		t.Fatalf("expected one session upsert, got %d", store.sessionUpserts)
	}
	if store.sessionProjection.SessionID != "sess_1" {
		t.Fatalf("expected session id sess_1, got %s", store.sessionProjection.SessionID)
	}
	if store.sessionProjection.EndedCount != 1 {
		t.Fatalf("expected ended count 1, got %d", store.sessionProjection.EndedCount)
	}
	if store.sessionProjection.EndedOnTimeCount != 1 {
		t.Fatalf("expected ended on time count 1, got %d", store.sessionProjection.EndedOnTimeCount)
	}
	if store.sessionProjection.UpdatedAt.IsZero() {
		t.Fatalf("expected updated_at to be set")
	}
}

func TestProjectionBuilderRebuildSessionProjectionRejectsMismatchedIDs(t *testing.T) {
	store := &projectionStoreSpy{}
	builder := NewProjectionBuilder(store, NewManager())

	err := builder.RebuildSessionProjection(
		"sess_1",
		session.Snapshot{ID: "sess_2", Status: session.StatusEnded, DurationSeconds: 1200},
		nil,
		time.Now().UTC(),
	)
	if err == nil {
		t.Fatalf("expected mismatch error")
	}
	if store.sessionUpserts != 0 {
		t.Fatalf("expected no upsert on mismatched ids")
	}
}

func TestProjectionBuilderRebuildPlatformProjectionUpsertsMappedProjection(t *testing.T) {
	store := &projectionStoreSpy{}
	builder := NewProjectionBuilder(store, NewManager())

	now := time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC)
	endedOverrun := -15

	err := builder.RebuildPlatformProjection(
		[]session.Snapshot{{ID: "sess_1", Status: session.StatusEnded, DurationSeconds: 1800}},
		map[string][]programitem.Snapshot{
			"sess_1": {
				{Status: programitem.StatusEnded, RuntimeDurationSeconds: 600, EndedRemainingSeconds: &endedOverrun},
			},
		},
		now,
	)
	if err != nil {
		t.Fatalf("rebuild platform projection: %v", err)
	}

	if store.platformUpserts != 1 {
		t.Fatalf("expected one platform upsert, got %d", store.platformUpserts)
	}
	if store.platformProjection.TotalSessions != 1 {
		t.Fatalf("expected total sessions 1, got %d", store.platformProjection.TotalSessions)
	}
	if store.platformProjection.OverrunProgramItems != 1 {
		t.Fatalf("expected overrun program items 1, got %d", store.platformProjection.OverrunProgramItems)
	}
	if store.platformProjection.TotalOverrunSeconds != 15 {
		t.Fatalf("expected total overrun seconds 15, got %d", store.platformProjection.TotalOverrunSeconds)
	}
	if store.platformProjection.UpdatedAt.IsZero() {
		t.Fatalf("expected updated_at to be set")
	}
}
