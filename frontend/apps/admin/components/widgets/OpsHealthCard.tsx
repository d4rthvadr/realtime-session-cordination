import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import type { AnalyticsFreshness, ProcessorMetrics } from "@/lib/actions";
import {
  deriveAnalyticsHealth,
  analyticsHealthBadgeClasses,
  analyticsHealthLabel,
} from "@/lib/analytics-health";
import { cn } from "@/lib/utils";

interface OpsHealthCardProps {
  freshness: AnalyticsFreshness | null;
  metrics: ProcessorMetrics | null;
}

function MetricRow({
  label,
  value,
  highlight,
}: {
  label: string;
  value: string | number;
  highlight?: "warning" | "danger";
}) {
  const valueClass = cn("font-mono text-sm font-medium tabular-nums", {
    "text-amber-700": highlight === "warning",
    "text-red-700": highlight === "danger",
    "text-slate-800": !highlight,
  });

  return (
    <div className="flex items-center justify-between py-1.5 border-b border-slate-100 last:border-0">
      <span className="text-xs text-slate-500">{label}</span>
      <span className={valueClass}>{value}</span>
    </div>
  );
}

function formatTimestamp(ts: string | null | undefined): string {
  if (!ts) return "-";
  const d = new Date(ts);
  if (!Number.isFinite(d.getTime())) return "-";
  return d.toLocaleString();
}

export default function OpsHealthCard({
  freshness,
  metrics,
}: OpsHealthCardProps) {
  const health = deriveAnalyticsHealth(
    freshness,
    freshness ? "projection_or_fallback" : null,
  );
  const badgeClasses = analyticsHealthBadgeClasses(health);
  const badgeLabel = analyticsHealthLabel(health, "dashboard");

  return (
    <Card>
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <CardTitle className="text-base font-semibold text-slate-800">
            Processor Health
          </CardTitle>
          <Badge
            className={cn(
              "text-[10px] font-bold tracking-widest border",
              badgeClasses,
            )}
          >
            {badgeLabel}
          </Badge>
        </div>
        {freshness?.workerName && (
          <p className="text-xs text-slate-400 mt-0.5">
            worker: {freshness.workerName}
          </p>
        )}
      </CardHeader>

      <CardContent className="space-y-5">
        <div>
          <p className="text-[11px] font-semibold uppercase tracking-widest text-slate-400 mb-1">
            Freshness
          </p>
          <MetricRow
            label="Last processed at"
            value={formatTimestamp(freshness?.lastProcessedAt)}
          />
          <MetricRow
            label="Pending events"
            value={freshness?.pendingCount ?? "-"}
            highlight={
              (freshness?.pendingCount ?? 0) > 0 ? "warning" : undefined
            }
          />
          <MetricRow
            label="Oldest pending at"
            value={formatTimestamp(freshness?.oldestPendingAt)}
          />
          <MetricRow
            label="Retry due count"
            value={freshness?.retryDueCount ?? "-"}
            highlight={
              (freshness?.retryDueCount ?? 0) > 0 ? "warning" : undefined
            }
          />
          <MetricRow
            label="Dead-letter count"
            value={freshness?.deadLetterCount ?? "-"}
            highlight={
              (freshness?.deadLetterCount ?? 0) > 0 ? "danger" : undefined
            }
          />
          <MetricRow
            label="Retry lag (s)"
            value={freshness?.retryLagSeconds ?? "-"}
            highlight={
              (freshness?.retryLagSeconds ?? 0) > 30 ? "warning" : undefined
            }
          />
        </div>

        {metrics && (
          <div>
            <p className="text-[11px] font-semibold uppercase tracking-widest text-slate-400 mb-1">
              Runtime metrics
            </p>
            <MetricRow
              label="Processed (total)"
              value={metrics.processedCount}
            />
            <MetricRow
              label="Failed (total)"
              value={metrics.failedCount}
              highlight={metrics.failedCount > 0 ? "warning" : undefined}
            />
            <MetricRow
              label="Dead-lettered (total)"
              value={metrics.deadLetterCount}
              highlight={metrics.deadLetterCount > 0 ? "danger" : undefined}
            />
            <MetricRow
              label="Projection errors"
              value={metrics.projectionErrorCount}
              highlight={
                metrics.projectionErrorCount > 0 ? "warning" : undefined
              }
            />
            <MetricRow
              label="Checkpoint errors"
              value={metrics.checkpointErrorCount}
              highlight={
                metrics.checkpointErrorCount > 0 ? "warning" : undefined
              }
            />
            <MetricRow
              label="Last batch duration"
              value={
                metrics.lastBatchDurationMillis > 0
                  ? `${metrics.lastBatchDurationMillis} ms`
                  : "-"
              }
            />
            <MetricRow
              label="Last batch at"
              value={formatTimestamp(metrics.lastBatchAt)}
            />
          </div>
        )}

        {!metrics && (
          <p className="text-xs text-slate-400 italic">
            Runtime metrics unavailable.
          </p>
        )}
      </CardContent>
    </Card>
  );
}
