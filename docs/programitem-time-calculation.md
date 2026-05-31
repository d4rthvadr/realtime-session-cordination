# ProgramItem Time Calculation Guide

## Purpose

This document defines how countdown values are computed from persisted ProgramItem runtime fields.
It is the source of truth for runtime math and analytics derivation.

## Persisted Runtime Fields

- runtime_duration_seconds: Base runtime budget for the item.
- adjustment_seconds: Net time change from manual + or - adjustments.
- actual_start: Timestamp when runtime actually started.
- paused_at: Timestamp when current pause began.
- total_paused_duration_seconds: Sum of completed pause intervals.
- ended_remaining_seconds: Frozen remaining value captured when ending.
- actual_end: Timestamp when runtime ended.
- pause_count: Number of completed pause intervals.
- ended_reason: End reason label.

## Effective Budget

Effective budget is:

effective_budget_seconds = runtime_duration_seconds + adjustment_seconds

## Remaining Time By Status

### scheduled

remaining_seconds = effective_budget_seconds

### in_progress

elapsed_active_seconds = (now - actual_start) - total_paused_duration_seconds
remaining_seconds = effective_budget_seconds - elapsed_active_seconds

### paused

elapsed_active_seconds = (paused_at - actual_start) - total_paused_duration_seconds
remaining_seconds = effective_budget_seconds - elapsed_active_seconds

### ended

If ended_remaining_seconds exists:

remaining_seconds = ended_remaining_seconds

Fallback (legacy safety path):

elapsed_active_seconds = (actual_end - actual_start) - total_paused_duration_seconds
remaining_seconds = effective_budget_seconds - elapsed_active_seconds

### canceled

No runtime countdown is authoritative after cancellation.
UI should treat canceled as non-active.

## Transition Math

### start

- status becomes in_progress
- actual_start is set to now
- paused_at is cleared
- ended_remaining_seconds is cleared
- pause_count resets to 0 for fresh runtime
- total_paused_duration_seconds resets to 0 for fresh runtime

### pause

- status becomes paused
- paused_at is set to now

### resume

- paused_for_seconds = now - paused_at
- total_paused_duration_seconds += paused_for_seconds
- pause_count += 1
- paused_at is cleared
- status becomes in_progress

### adjust time (+60 or -60)

- adjustment_seconds += delta_seconds
- no other runtime field needs mutation

### end

- remaining is computed using current status branch
- ended_remaining_seconds is set to computed remaining
- actual_end is set to now
- paused_at is cleared
- status becomes ended

## Analytics Derivations

From persisted fields you can compute:

- planned_seconds: runtime_duration_seconds
- effective_budget_seconds: runtime_duration_seconds + adjustment_seconds
- total_pause_seconds: total_paused_duration_seconds
- end_remaining_seconds: ended_remaining_seconds
- overrun_seconds: max(0, -ended_remaining_seconds)
- underrun_seconds: max(0, ended_remaining_seconds)
- pause_frequency: pause_count per item/session

## Notes

- Keep raw remaining values for analytics, even if UI clamps display at zero.
- All timestamps are UTC.
- Runtime calculations should be done server-side for consistency.
