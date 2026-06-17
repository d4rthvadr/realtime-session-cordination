import type { AnalyticsDataSource, AnalyticsFreshness } from "@/lib/actions";

export type AnalyticsHealth =
  | "healthy"
  | "lagging"
  | "stale"
  | "unavailable"
  | "error";

const STALE_THRESHOLD_SECONDS = 120;

export function deriveAnalyticsHealth(
  freshness: AnalyticsFreshness | null,
  source: AnalyticsDataSource | null,
): AnalyticsHealth {
  if (source === "error") {
    return "error";
  }
  if (!freshness) {
    return "unavailable";
  }
  if (freshness.pendingCount > 0) {
    return "lagging";
  }
  if (freshness.lastProcessedAt) {
    const lastProcessed = Date.parse(freshness.lastProcessedAt);
    if (Number.isFinite(lastProcessed)) {
      const ageSeconds = Math.max(0, (Date.now() - lastProcessed) / 1000);
      if (ageSeconds > STALE_THRESHOLD_SECONDS) {
        return "stale";
      }
    }
  }
  return "healthy";
}

export function analyticsHealthBadgeClasses(health: AnalyticsHealth): string {
  switch (health) {
    case "healthy":
      return "bg-emerald-50 text-emerald-700 border-emerald-200";
    case "lagging":
      return "bg-amber-50 text-amber-700 border-amber-200";
    case "stale":
      return "bg-orange-50 text-orange-700 border-orange-200";
    case "error":
      return "bg-red-50 text-red-700 border-red-200";
    default:
      return "bg-slate-100 text-slate-700 border-slate-200";
  }
}

export function analyticsHealthLabel(
  health: AnalyticsHealth,
  variant: "dashboard" | "session" = "session",
): string {
  if (variant === "dashboard") {
    switch (health) {
      case "healthy":
        return "ANALYTICS HEALTHY";
      case "lagging":
        return "ANALYTICS LAGGING";
      case "stale":
        return "ANALYTICS STALE";
      case "error":
        return "ANALYTICS ERROR";
      default:
        return "ANALYTICS UNAVAILABLE";
    }
  }

  switch (health) {
    case "healthy":
      return "HEALTHY";
    case "lagging":
      return "LAGGING";
    case "stale":
      return "STALE";
    case "error":
      return "ERROR";
    default:
      return "UNAVAILABLE";
  }
}

export function analyticsSourceLabel(
  source: AnalyticsDataSource | null,
  hasAnalyticsData: boolean,
): string {
  if (source === "error") {
    return "source: fetch_error";
  }
  if (source === "unavailable") {
    return "source: unavailable";
  }
  if (!hasAnalyticsData) {
    return "source: fallback";
  }
  return "source: projection_or_fallback";
}
