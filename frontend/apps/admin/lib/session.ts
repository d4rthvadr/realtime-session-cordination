export function formatClock(
  totalSeconds: number | string | null | undefined,
  fallback = "--:--",
): string {
  const parsed =
    typeof totalSeconds === "number" ? totalSeconds : Number(totalSeconds);

  if (!Number.isFinite(parsed)) {
    return fallback;
  }

  const sign = parsed < 0 ? "-" : "";
  const abs = Math.floor(Math.abs(parsed));
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
