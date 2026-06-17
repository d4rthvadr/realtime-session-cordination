package analytics

import (
	"fmt"
	"time"

	"realtime-session-coordination/backend/internal/programitem"
	"realtime-session-coordination/backend/internal/session"
)

// ProjectionBuilder rebuilds materialized analytics projections from domain snapshots.
type ProjectionBuilder struct {
	store   ProjectionStore
	manager *Manager
}

func NewProjectionBuilder(store ProjectionStore, manager *Manager) *ProjectionBuilder {
	if manager == nil {
		manager = NewManager()
	}
	return &ProjectionBuilder{store: store, manager: manager}
}

func (b *ProjectionBuilder) RebuildSessionProjection(sessionID string, sessionSnap session.Snapshot, items []programitem.Snapshot, now time.Time) error {
	if b.store == nil {
		return fmt.Errorf("projection store is required")
	}
	if sessionID == "" {
		return fmt.Errorf("session id is required")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	if sessionSnap.ID == "" {
		sessionSnap.ID = sessionID
	}
	if sessionSnap.ID != sessionID {
		return fmt.Errorf("session snapshot id does not match requested session id")
	}

	summary := b.manager.BuildSessionSummary(sessionSnap, items, now)
	projection := SessionProjection{
		SessionID:              summary.SessionID,
		SessionStatus:          summary.SessionStatus,
		SessionDurationSeconds: summary.SessionDurationSeconds,
		ProgramItemCount:       summary.ProgramItemCount,
		ScheduledCount:         summary.ScheduledCount,
		InProgressCount:        summary.InProgressCount,
		PausedCount:            summary.PausedCount,
		EndedCount:             summary.EndedCount,
		CanceledCount:          summary.CanceledCount,
		PlannedSeconds:         summary.PlannedSeconds,
		EffectiveBudgetSeconds: summary.EffectiveBudgetSeconds,
		TotalAdjustmentSeconds: summary.TotalAdjustmentSeconds,
		TotalPauseSeconds:      summary.TotalPauseSeconds,
		TotalPauseCount:        summary.TotalPauseCount,
		EndedOnTimeCount:       summary.EndedOnTimeCount,
		OverrunItemCount:       summary.OverrunItemCount,
		TotalOverrunSeconds:    summary.TotalOverrunSeconds,
		TotalUnderrunSeconds:   summary.TotalUnderrunSeconds,
		EndedOnTimeRatio:       summary.EndedOnTimeRatio,
		ComputedAt:             summary.ComputedAt,
		UpdatedAt:              now.UTC(),
	}

	return b.store.UpsertSessionProjection(projection)
}

func (b *ProjectionBuilder) RebuildPlatformProjection(sessions []session.Snapshot, itemsBySession map[string][]programitem.Snapshot, now time.Time) error {
	if b.store == nil {
		return fmt.Errorf("projection store is required")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}

	overview := b.manager.BuildPlatformOverview(sessions, itemsBySession, now)
	projection := PlatformProjection{
		TotalSessions:            overview.TotalSessions,
		CreatedSessions:          overview.CreatedSessions,
		LiveSessions:             overview.LiveSessions,
		PausedSessions:           overview.PausedSessions,
		EndedSessions:            overview.EndedSessions,
		TotalProgramItems:        overview.TotalProgramItems,
		EndedProgramItems:        overview.EndedProgramItems,
		OnTimeEndedProgramItems:  overview.OnTimeEndedProgramItems,
		OverrunProgramItems:      overview.OverrunProgramItems,
		TotalSessionDurationSecs: overview.TotalSessionDurationSecs,
		TotalPlannedSeconds:      overview.TotalPlannedSeconds,
		EffectiveBudgetSeconds:   overview.EffectiveBudgetSeconds,
		TotalAdjustmentSeconds:   overview.TotalAdjustmentSeconds,
		TotalPauseSeconds:        overview.TotalPauseSeconds,
		TotalPauseCount:          overview.TotalPauseCount,
		TotalOverrunSeconds:      overview.TotalOverrunSeconds,
		TotalUnderrunSeconds:     overview.TotalUnderrunSeconds,
		SessionCompletionRatio:   overview.SessionCompletionRatio,
		ProgramItemOnTimeRatio:   overview.ProgramItemOnTimeRatio,
		ComputedAt:               overview.ComputedAt,
		UpdatedAt:                now.UTC(),
	}

	return b.store.UpsertPlatformProjection(projection)
}