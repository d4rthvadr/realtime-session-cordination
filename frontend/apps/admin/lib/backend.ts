export const ADMIN_BACKEND_BASE_URL =
  process.env.NEXT_PUBLIC_BACKEND_BASE_URL || "http://localhost:8080";

export function buildAdminApiUrl(path: string): string {
  return `${ADMIN_BACKEND_BASE_URL.replace(/\/$/, "")}${path}`;
}

export function buildAdminWsUrl(path: string): string {
  const base = ADMIN_BACKEND_BASE_URL.replace(/\/$/, "");
  const wsBase = base.startsWith("https://")
    ? base.replace("https://", "wss://")
    : base.replace("http://", "ws://");

  return `${wsBase}${path}`;
}

export function getViewerUrl(sessionId: string): string {
  const base = process.env.NEXT_PUBLIC_USER_APP_URL || "http://localhost:3001";
  return `${base.replace(/\/$/, "")}/sessions/${sessionId}`;
}
