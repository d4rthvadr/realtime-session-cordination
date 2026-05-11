export interface TimerState {
  level: "safe" | "warning" | "critical" | "overtime";
  label: "Safe" | "Warning" | "Critical" | "Overtime";
  colorClass: string;
}

export function formatDuration(totalSeconds: number): string {
  const absSeconds = Math.max(0, Math.floor(Math.abs(totalSeconds)));
  const mins = Math.floor(absSeconds / 60)
    .toString()
    .padStart(2, "0");
  const secs = (absSeconds % 60).toString().padStart(2, "0");
  return `${mins}:${secs}`;
}

export function getTimerState(
  remainingSeconds: number,
  durationSeconds: number,
): TimerState {
  if (remainingSeconds < 0) {
    return {
      level: "overtime",
      label: "Overtime",
      colorClass: "text-critical",
    };
  }

  const ratio = remainingSeconds / Math.max(durationSeconds, 1);
  if (ratio > 0.4) {
    return {
      level: "safe",
      label: "Safe",
      colorClass: "text-safe",
    };
  }

  if (ratio > 0.15) {
    return {
      level: "warning",
      label: "Warning",
      colorClass: "text-warning",
    };
  }

  return {
    level: "critical",
    label: "Critical",
    colorClass: "text-critical",
  };
}
