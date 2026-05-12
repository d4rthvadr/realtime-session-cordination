export const USER_BACKEND_BASE_URL =
  process.env.NEXT_PUBLIC_BACKEND_BASE_URL || "http://localhost:8080";

export function buildUserApiUrl(path: string): string {
  return `${USER_BACKEND_BASE_URL.replace(/\/$/, "")}${path}`;
}

export function buildUserWsUrl(path: string): string {
  const base = USER_BACKEND_BASE_URL.replace(/^http/i, "ws");
  return `${base.replace(/\/$/, "")}${path}`;
}
