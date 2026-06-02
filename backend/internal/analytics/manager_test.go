package analytics

import (
	"testing"
	"time"

	"realtime-session-coordination/backend/internal/programitem"
	"realtime-session-coordination/backend/internal/session"
)

func TestBuildSessionSummaryAggregatesRuntimeFields(t *testing.T) {
	now := time.Date(2026, 6, 2, 10, 0, 0, 0, time.UTC)
	mgr := NewManager()

	endedOnTime := 15
	endedOverrun := -20

	summary := mgr.BuildSessionSummary(
		session.Snapshot{ID: "sess_1", Status: session.StatusEnded, DurationSeconds: 3600},
		[]programitem.Snapshot{
			{
				ID:                         "pi_1",
				Status:                     programitem.StatusEnded,
				RuntimeDurationSeconds:     1200,
				AdjustmentSeconds:          60,
				TotalPausedDurationSeconds: 30,
				PauseCount:                 1,
				EndedRemainingSeconds:      &endedOnTime,
			},
			{
				ID:                         "pi_2",
				Status:                     programitem.StatusEnded,
				RuntimeDurationSeconds:     900,
				AdjustmentSeconds:          -30,
				TotalPausedDurationSeconds: 20,
				PauseCount:                 2,
				EndedRemainingSeconds:      &endedOverrun,
			},
			{ID: "pi_3", Status: programitem.StatusCanceled, RuntimeDurationSeconds: 300},
			{ID: "pi_4", Status: programitem.StatusScheduled, RuntimeDurationSeconds: 300},
		},
		now,
	)

	if summary.ProgramItemCount != 4 {
		t.Fatalf("expected ProgramItemCount=4, got %d", summary.ProgramItemCount)
	}
	if summary.EndedCount != 2 {
		t.Fatalf("expected EndedCount=2, got %d", summary.EndedCount)
	}
	if summary.CanceledCount != 1 {
		t.Fatalf("expected CanceledCount=1, got %d", summary.CanceledCount)
	}
	if summary.ScheduledCount != 1 {
		t.Fatalf("expected ScheduledCount=1, got %d", summary.ScheduledCount)
	}
	if summary.TotalPauseSeconds != 50 {
		t.Fatalf("expected TotalPauseSeconds=50, got %d", summary.TotalPauseSeconds)
	}
	if summary.TotalPauseCount != 3 {
		t.Fatalf("expected TotalPauseCount=3, got %d", summary.TotalPauseCount)
	}
	if summary.PlannedSeconds != 2700 {
		t.Fatalf("expected PlannedSeconds=2700, got %d", summary.PlannedSeconds)
	}
	if summary.EffectiveBudgetSeconds != 2730 {
		t.Fatalf("expected EffectiveBudgetSeconds=2730, got %d", summary.EffectiveBudgetSeconds)
	}
	if summary.EndedOnTimeCount != 1 {
		t.Fatalf("expected EndedOnTimeCount=1, got %d", summary.EndedOnTimeCount)
	}
	if summary.OverrunItemCount != 1 {
		t.Fatalf("expected OverrunItemCount=1, got %d", summary.OverrunItemCount)
	}
	if summary.TotalOverrunSeconds != 20 {
		t.Fatalf("expected TotalOverrunSeconds=20, got %d", summary.TotalOverrunSeconds)
	}
	if summary.TotalUnderrunSeconds != 15 {
		t.Fatalf("expected TotalUnderrunSeconds=15, got %d", summary.TotalUnderrunSeconds)
	}
	if summary.EndedOnTimeRatio != 0.5 {
		t.Fatalf("expected EndedOnTimeRatio=0.5, got %f", summary.EndedOnTimeRatio)
	}
	if !summary.ComputedAt.Equal(now) {
		t.Fatalf("expected ComputedAt to equal now")
	}
}
