package analytics

import (
	"time"

	"realtime-session-coordination/backend/internal/programitem"
	"realtime-session-coordination/backend/internal/session"
)

// SessionSummary is a compact analytics view for a single session.
type SessionSummary struct {
	SessionID              string    `json:"sessionId"`
	SessionStatus          string    `json:"sessionStatus"`
	SessionDurationSeconds int       `json:"sessionDurationSeconds"`
	ProgramItemCount       int       `json:"programItemCount"`
	ScheduledCount         int       `json:"scheduledCount"`
	InProgressCount        int       `json:"inProgressCount"`
	PausedCount            int       `json:"pausedCount"`
	EndedCount             int       `json:"endedCount"`
	CanceledCount          int       `json:"canceledCount"`
	PlannedSeconds         int       `json:"plannedSeconds"`
	EffectiveBudgetSeconds int       `json:"effectiveBudgetSeconds"`
	TotalAdjustmentSeconds int       `json:"totalAdjustmentSeconds"`
	TotalPauseSeconds      int       `json:"totalPauseSeconds"`
	TotalPauseCount        int       `json:"totalPauseCount"`
	EndedOnTimeCount       int       `json:"endedOnTimeCount"`
	OverrunItemCount       int       `json:"overrunItemCount"`
	TotalOverrunSeconds    int       `json:"totalOverrunSeconds"`
	TotalUnderrunSeconds   int       `json:"totalUnderrunSeconds"`
	EndedOnTimeRatio       float64   `json:"endedOnTimeRatio"`
	ComputedAt             time.Time `json:"computedAt"`
}

// Manager computes analytics snapshots from existing domain snapshots.
type Manager struct{}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) BuildSessionSummary(sessionSnap session.Snapshot, items []programitem.Snapshot, now time.Time) SessionSummary {
	summary := SessionSummary{
		SessionID:              sessionSnap.ID,
		SessionStatus:          sessionSnap.Status,
		SessionDurationSeconds: sessionSnap.DurationSeconds,
		ProgramItemCount:       len(items),
		ComputedAt:             now.UTC(),
	}

	for _, item := range items {
		summary.PlannedSeconds += plannedSeconds(item)
		summary.EffectiveBudgetSeconds += item.RuntimeDurationSeconds + item.AdjustmentSeconds
		summary.TotalAdjustmentSeconds += item.AdjustmentSeconds
		summary.TotalPauseSeconds += item.TotalPausedDurationSeconds
		summary.TotalPauseCount += item.PauseCount

		switch item.Status {
		case programitem.StatusScheduled:
			summary.ScheduledCount++
		case programitem.StatusInProgress:
			summary.InProgressCount++
		case programitem.StatusPaused:
			summary.PausedCount++
		case programitem.StatusEnded:
			summary.EndedCount++
			if item.EndedRemainingSeconds != nil {
				remaining := *item.EndedRemainingSeconds
				if remaining >= 0 {
					summary.EndedOnTimeCount++
					summary.TotalUnderrunSeconds += remaining
				} else {
					summary.OverrunItemCount++
					summary.TotalOverrunSeconds += -remaining
				}
			}
		case programitem.StatusCanceled:
			summary.CanceledCount++
		}
	}

	if summary.EndedCount > 0 {
		summary.EndedOnTimeRatio = float64(summary.EndedOnTimeCount) / float64(summary.EndedCount)
	}

	return summary
}

func plannedSeconds(item programitem.Snapshot) int {
	if item.RuntimeDurationSeconds > 0 {
		return item.RuntimeDurationSeconds
	}

	if item.ExpectedDurationMinutes > 0 {
		return item.ExpectedDurationMinutes * 60
	}

	seconds := int(item.ScheduledEnd.Sub(item.ScheduledStart).Seconds())
	if seconds > 0 {
		return seconds
	}

	return 0
}
