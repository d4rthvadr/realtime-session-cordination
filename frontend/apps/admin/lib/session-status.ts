export type SessionStatus = "CREATED" | "LIVE" | "PAUSED" | "ENDED";

export function getSessionStatusBadgeClasses(status: string): string {
  switch (status as SessionStatus) {
    case "LIVE":
      return "bg-emerald-100 text-emerald-700 border-emerald-200";
    case "PAUSED":
      return "bg-amber-100 text-amber-700 border-amber-200";
    case "ENDED":
      return "bg-slate-100 text-slate-600 border-slate-200";
    default:
      return "bg-blue-100 text-blue-700 border-blue-200";
  }
}
