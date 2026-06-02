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

func TestBuildPlatformOverviewAggregatesSessions(t *testing.T) {
	now := time.Date(2026, 6, 2, 10, 0, 0, 0, time.UTC)
	mgr := NewManager()

	endedOnTime := 10
	endedOverrun := -30

	overview := mgr.BuildPlatformOverview(
		[]session.Snapshot{
			{ID: "sess_1", Status: session.StatusEnded, DurationSeconds: 1800},
			{ID: "sess_2", Status: session.StatusLive, DurationSeconds: 2400},
			{ID: "sess_3", Status: session.StatusCreated, DurationSeconds: 1200},
		},
		map[string][]programitem.Snapshot{
			"sess_1": {
				{
					Status:                     programitem.StatusEnded,
					RuntimeDurationSeconds:     600,
					AdjustmentSeconds:          60,
					TotalPausedDurationSeconds: 20,
					PauseCount:                 1,
					EndedRemainingSeconds:      &endedOnTime,
				},
				{
					Status:                     programitem.StatusEnded,
					RuntimeDurationSeconds:     600,
					AdjustmentSeconds:          -10,
					TotalPausedDurationSeconds: 10,
					PauseCount:                 1,
					EndedRemainingSeconds:      &endedOverrun,
				},
			},
			"sess_2": {
				{Status: programitem.StatusInProgress, RuntimeDurationSeconds: 300},
			},
		},
		now,
	)

	if overview.TotalSessions != 3 {
		t.Fatalf("expected TotalSessions=3, got %d", overview.TotalSessions)
	}
	if overview.EndedSessions != 1 {
		t.Fatalf("expected EndedSessions=1, got %d", overview.EndedSessions)
	}
	if overview.LiveSessions != 1 {
		t.Fatalf("expected LiveSessions=1, got %d", overview.LiveSessions)
	}
	if overview.CreatedSessions != 1 {
		t.Fatalf("expected CreatedSessions=1, got %d", overview.CreatedSessions)
	}
	if overview.TotalProgramItems != 3 {
		t.Fatalf("expected TotalProgramItems=3, got %d", overview.TotalProgramItems)
	}
	if overview.EndedProgramItems != 2 {
		t.Fatalf("expected EndedProgramItems=2, got %d", overview.EndedProgramItems)
	}
	if overview.OnTimeEndedProgramItems != 1 {
		t.Fatalf("expected OnTimeEndedProgramItems=1, got %d", overview.OnTimeEndedProgramItems)
	}
	if overview.OverrunProgramItems != 1 {
		t.Fatalf("expected OverrunProgramItems=1, got %d", overview.OverrunProgramItems)
	}
	if overview.TotalSessionDurationSecs != 5400 {
		t.Fatalf("expected TotalSessionDurationSecs=5400, got %d", overview.TotalSessionDurationSecs)
	}
	if overview.TotalPlannedSeconds != 1500 {
		t.Fatalf("expected TotalPlannedSeconds=1500, got %d", overview.TotalPlannedSeconds)
	}
	if overview.EffectiveBudgetSeconds != 1550 {
		t.Fatalf("expected EffectiveBudgetSeconds=1550, got %d", overview.EffectiveBudgetSeconds)
	}
	if overview.TotalAdjustmentSeconds != 50 {
		t.Fatalf("expected TotalAdjustmentSeconds=50, got %d", overview.TotalAdjustmentSeconds)
	}
	if overview.TotalPauseSeconds != 30 {
		t.Fatalf("expected TotalPauseSeconds=30, got %d", overview.TotalPauseSeconds)
	}
	if overview.TotalPauseCount != 2 {
		t.Fatalf("expected TotalPauseCount=2, got %d", overview.TotalPauseCount)
	}
	if overview.TotalOverrunSeconds != 30 {
		t.Fatalf("expected TotalOverrunSeconds=30, got %d", overview.TotalOverrunSeconds)
	}
	if overview.TotalUnderrunSeconds != 10 {
		t.Fatalf("expected TotalUnderrunSeconds=10, got %d", overview.TotalUnderrunSeconds)
	}
	if overview.SessionCompletionRatio != (1.0 / 3.0) {
		t.Fatalf("expected SessionCompletionRatio=1/3, got %f", overview.SessionCompletionRatio)
	}
	if overview.ProgramItemOnTimeRatio != 0.5 {
		t.Fatalf("expected ProgramItemOnTimeRatio=0.5, got %f", overview.ProgramItemOnTimeRatio)
	}
	if !overview.ComputedAt.Equal(now) {
		t.Fatalf("expected ComputedAt to equal now")
	}
}
