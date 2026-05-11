export function formatClock(totalSeconds: number): string {
  const sign = totalSeconds < 0 ? "-" : "";
  const abs = Math.floor(Math.abs(totalSeconds));
  const mins = Math.floor(abs / 60)
    .toString()
    .padStart(2, "0");
  const secs = (abs % 60).toString().padStart(2, "0");
  return `${sign}${mins}:${secs}`;
}

export function getPublicViewerUrl(sessionId: string): string {
  const base = process.env.NEXT_PUBLIC_USER_APP_URL || "http://localhost:3001";
  return `${base.replace(/\/$/, "")}/sessions/${sessionId}`;
}

export function parseDurationToSeconds(rawMinutes: string): number {
  const parsed = Number(rawMinutes);
  if (!Number.isFinite(parsed) || parsed <= 0) {
    return 0;
  }
  return Math.floor(parsed * 60);
}
