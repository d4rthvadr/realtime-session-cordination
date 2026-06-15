"use client";

import { useEffect, useMemo, useState, useTransition } from "react";
import Link from "next/link";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import SessionCreateModal from "@/components/SessionCreateModal";
import {
  getAnalyticsOverview,
  getSessionsList,
  SessionSnapshot,
  AnalyticsOverview,
  AnalyticsFreshness,
  AnalyticsDataSource,
} from "@/lib/actions";
import { formatClock } from "@/lib/session";
import { cn } from "@/lib/utils";
import {
  ChartConfig,
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from "@/components/ui/chart";
import {
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  Pie,
  PieChart,
  PolarAngleAxis,
  RadialBar,
  RadialBarChart,
  XAxis,
} from "recharts";
import { Calendar, CircleHelp, Clock, ExternalLink, Plus } from "lucide-react";

function clampPercent(value: number): number {
  if (!Number.isFinite(value)) {
    return 0;
  }

  return Math.max(0, Math.min(100, Math.round(value)));
}

function formatMinutes(seconds: number): string {
  return `${Math.round(seconds / 60)} min`;
}

function formatSignedSeconds(seconds: number): string {
  const sign = seconds > 0 ? "+" : seconds < 0 ? "-" : "";
  return `${sign}${formatClock(Math.abs(seconds), "00:00")}`;
}

function InfoHint({ text }: { text: string }) {
  return (
    <span
      className="inline-flex cursor-help text-slate-400 transition-colors hover:text-slate-600"
      title={text}
      aria-label={text}
    >
      <CircleHelp className="h-3.5 w-3.5" />
    </span>
  );
}

type AnalyticsHealth =
  | "healthy"
  | "lagging"
  | "stale"
  | "unavailable"
  | "error";

function deriveAnalyticsHealth(
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
      if (ageSeconds > 120) {
        return "stale";
      }
    }
  }

  return "healthy";
}

function analyticsHealthBadgeClasses(health: AnalyticsHealth): string {
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

function analyticsHealthLabel(health: AnalyticsHealth): string {
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

function analyticsSourceLabel(
  source: AnalyticsDataSource | null,
  hasOverview: boolean,
): string {
  if (source === "error") {
    return "source: fetch_error";
  }
  if (source === "unavailable") {
    return "source: unavailable";
  }
  if (!hasOverview) {
    return "source: fallback";
  }
  return "source: projection_or_fallback";
}

function GlobalTimeHealthSection({
  totalSessions,
  activeSessions,
  todaySessions,
  avgDuration,
  analyticsHealth,
  analyticsSource,
  freshness,
  onRefresh,
}: {
  totalSessions: number;
  activeSessions: number;
  todaySessions: number;
  avgDuration: number;
  analyticsHealth: AnalyticsHealth;
  analyticsSource: string;
  freshness: AnalyticsFreshness | null;
  onRefresh: () => void;
}) {
  return (
    <section className="grid grid-cols-1 md:grid-cols-4 lg:grid-cols-5 gap-4">
      <Card className="md:col-span-3 lg:col-span-4 border-slate-200">
        <CardHeader className="pb-3">
          <div className="flex justify-between items-start">
            <div>
              <CardTitle className="text-2xl font-bold text-slate-900">
                Global Time Health
              </CardTitle>
              <p className="text-sm text-slate-600 mt-1">
                Real-time enterprise synchronization status
              </p>
            </div>
            <div className="flex flex-col items-end gap-2">
              <Badge className="bg-emerald-50 text-emerald-700 border-emerald-200 hover:bg-emerald-50">
                <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse mr-2"></span>
                NETWORK OPTIMIZED
              </Badge>
              <Badge
                className={cn(
                  "text-[10px] font-semibold border uppercase tracking-wider",
                  analyticsHealthBadgeClasses(analyticsHealth),
                )}
              >
                {analyticsHealthLabel(analyticsHealth)}
              </Badge>
              <p className="text-[11px] text-slate-500 text-right">
                {analyticsSource}
                {freshness ? ` • pending ${freshness.pendingCount}` : ""}
              </p>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-6">
            <div className="space-y-1">
              <span className="text-xs font-semibold text-slate-500 uppercase tracking-wider">
                Total Sessions
              </span>
              <div className="flex items-baseline gap-1">
                <span className="text-3xl md:text-4xl font-bold text-slate-900">
                  {totalSessions}
                </span>
                <span className="text-sm text-slate-500">total</span>
              </div>
            </div>
            <div className="space-y-1">
              <span className="text-xs font-semibold text-slate-500 uppercase tracking-wider">
                Active Sessions
              </span>
              <div className="flex items-baseline gap-1">
                <span className="text-3xl md:text-4xl font-bold text-blue-600">
                  {activeSessions}
                </span>
                <span className="text-sm text-slate-500">live/paused</span>
              </div>
            </div>
            <div className="space-y-1">
              <span className="text-xs font-semibold text-slate-500 uppercase tracking-wider">
                Today&apos;s Sessions
              </span>
              <div className="flex items-baseline gap-1">
                <span className="text-3xl md:text-4xl font-bold text-slate-900">
                  {todaySessions}
                </span>
                <span className="text-sm text-slate-500">today</span>
              </div>
            </div>
            <div className="space-y-1">
              <span className="text-xs font-semibold text-slate-500 uppercase tracking-wider">
                Avg Duration
              </span>
              <div className="flex items-baseline gap-1">
                <span className="text-3xl md:text-4xl font-bold text-slate-900">
                  {avgDuration}
                </span>
                <span className="text-sm text-slate-500">min</span>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card className="bg-slate-900 text-white border-slate-800 hover:scale-[1.02] transition-transform duration-300 cursor-pointer">
        <CardContent className="p-6 flex flex-col justify-between h-full">
          <div>
            <Plus className="w-8 h-8 mb-4 text-slate-400" />
            <h3 className="text-xl font-semibold mb-2">Initialize Session</h3>
            <p className="text-sm text-slate-400">
              Deploy a new atomic-locked session across the cluster.
            </p>
          </div>
          <div className="mt-6">
            <SessionCreateModal
              onSuccess={onRefresh}
              trigger={
                <Button className="w-full bg-white text-slate-900 hover:bg-slate-100 rounded-full">
                  Rapid Launch
                  <ExternalLink className="w-4 h-4 ml-2" />
                </Button>
              }
            />
          </div>
        </CardContent>
      </Card>
    </section>
  );
}

function AnalyticsSection({
  sessionCompletionPercent,
  endedSessions,
  totalSessions,
  onTimePercent,
  onTimeEndedProgramItems,
  endedProgramItems,
  completionGaugeConfig,
  onTimeGaugeConfig,
  sessionStatusChartConfig,
  sessionStatusData,
  outcomesChartConfig,
  programOutcomeData,
  totalProgramItems,
  overrunProgramItems,
  budgetChartConfig,
  budgetData,
  totalPlannedSeconds,
  effectiveBudgetSeconds,
  totalSessionDurationSeconds,
  impactsChartConfig,
  timingImpactData,
  totalPauseSeconds,
  totalPauseCount,
  totalAdjustmentSeconds,
  totalOverrunSeconds,
  totalUnderrunSeconds,
}: {
  sessionCompletionPercent: number;
  endedSessions: number;
  totalSessions: number;
  onTimePercent: number;
  onTimeEndedProgramItems: number;
  endedProgramItems: number;
  completionGaugeConfig: ChartConfig;
  onTimeGaugeConfig: ChartConfig;
  sessionStatusChartConfig: ChartConfig;
  sessionStatusData: Array<{ name: string; value: number; fill: string }>;
  outcomesChartConfig: ChartConfig;
  programOutcomeData: Array<{ name: string; value: number; fill: string }>;
  totalProgramItems: number;
  overrunProgramItems: number;
  budgetChartConfig: ChartConfig;
  budgetData: Array<{ name: string; value: number }>;
  totalPlannedSeconds: number;
  effectiveBudgetSeconds: number;
  totalSessionDurationSeconds: number;
  impactsChartConfig: ChartConfig;
  timingImpactData: Array<{ name: string; value: number }>;
  totalPauseSeconds: number;
  totalPauseCount: number;
  totalAdjustmentSeconds: number;
  totalOverrunSeconds: number;
  totalUnderrunSeconds: number;
}) {
  return (
    <>
      <Card className="lg:col-span-3 border-slate-200">
        <CardHeader className="pb-2">
          <CardTitle className="text-base font-semibold flex items-center gap-2">
            Session Completion
            <InfoHint text="Ended sessions divided by total sessions in analytics overview." />
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-1">
          <ChartContainer
            config={completionGaugeConfig}
            className="h-40 w-full"
          >
            <RadialBarChart
              data={[{ name: "completion", value: sessionCompletionPercent }]}
              startAngle={180}
              endAngle={0}
              innerRadius="68%"
              outerRadius="100%"
            >
              <PolarAngleAxis type="number" domain={[0, 100]} tick={false} />
              <ChartTooltip content={<ChartTooltipContent hideLabel />} />
              <RadialBar dataKey="value" cornerRadius={10} background />
            </RadialBarChart>
          </ChartContainer>
          <div className="text-center">
            <p className="text-3xl font-bold text-slate-900">
              {sessionCompletionPercent}%
            </p>
            <p className="text-xs text-slate-500">
              {endedSessions} of {totalSessions} sessions ended
            </p>
          </div>
        </CardContent>
      </Card>

      <Card className="lg:col-span-3 border-slate-200">
        <CardHeader className="pb-2">
          <CardTitle className="text-base font-semibold flex items-center gap-2">
            On-Time Delivery
            <InfoHint text="On-time ended program items divided by ended program items." />
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-1">
          <ChartContainer config={onTimeGaugeConfig} className="h-40 w-full">
            <RadialBarChart
              data={[{ name: "onTime", value: onTimePercent }]}
              startAngle={180}
              endAngle={0}
              innerRadius="68%"
              outerRadius="100%"
            >
              <PolarAngleAxis type="number" domain={[0, 100]} tick={false} />
              <ChartTooltip content={<ChartTooltipContent hideLabel />} />
              <RadialBar dataKey="value" cornerRadius={10} background />
            </RadialBarChart>
          </ChartContainer>
          <div className="text-center">
            <p className="text-3xl font-bold text-slate-900">
              {onTimePercent}%
            </p>
            <p className="text-xs text-slate-500">
              {onTimeEndedProgramItems} on-time out of {endedProgramItems} ended
            </p>
          </div>
        </CardContent>
      </Card>

      <Card className="lg:col-span-6 border-slate-200">
        <CardHeader className="pb-2">
          <CardTitle className="text-base font-semibold flex items-center gap-2">
            Session Status Mix
            <InfoHint text="Created, live, paused, and ended sessions from overview." />
          </CardTitle>
        </CardHeader>
        <CardContent>
          <ChartContainer
            config={sessionStatusChartConfig}
            className="h-52 w-full"
          >
            <BarChart data={sessionStatusData} margin={{ left: 10, right: 10 }}>
              <CartesianGrid vertical={false} />
              <XAxis dataKey="name" tickLine={false} axisLine={false} />
              <ChartTooltip content={<ChartTooltipContent hideLabel />} />
              <Bar dataKey="value" radius={8}>
                {sessionStatusData.map((entry) => (
                  <Cell key={entry.name} fill={entry.fill} />
                ))}
              </Bar>
            </BarChart>
          </ChartContainer>
        </CardContent>
      </Card>

      <Card className="lg:col-span-6 border-slate-200">
        <CardHeader className="pb-2">
          <CardTitle className="text-base font-semibold flex items-center gap-2">
            Program Item Outcomes
            <InfoHint text="Breakdown of all program items: on-time, overrun, other ended, and remaining." />
          </CardTitle>
        </CardHeader>
        <CardContent className="grid grid-cols-1 gap-4 items-center md:grid-cols-3">
          <div className="md:col-span-2">
            <ChartContainer
              config={outcomesChartConfig}
              className="h-56 w-full"
            >
              <PieChart>
                <ChartTooltip content={<ChartTooltipContent hideLabel />} />
                <Pie
                  data={programOutcomeData}
                  dataKey="value"
                  nameKey="name"
                  innerRadius={55}
                  outerRadius={85}
                  strokeWidth={4}
                >
                  {programOutcomeData.map((entry) => (
                    <Cell key={entry.name} fill={entry.fill} />
                  ))}
                </Pie>
              </PieChart>
            </ChartContainer>
          </div>
          <div className="space-y-2 text-sm">
            <div className="flex justify-between text-slate-700">
              <span>Total Program Items</span>
              <span className="font-semibold">{totalProgramItems}</span>
            </div>
            <div className="flex justify-between text-slate-700">
              <span>Ended Items</span>
              <span className="font-semibold">{endedProgramItems}</span>
            </div>
            <div className="flex justify-between text-slate-700">
              <span>On-Time Ended</span>
              <span className="font-semibold">{onTimeEndedProgramItems}</span>
            </div>
            <div className="flex justify-between text-slate-700">
              <span>Overrun Items</span>
              <span className="font-semibold">{overrunProgramItems}</span>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card className="lg:col-span-6 border-slate-200">
        <CardHeader className="pb-2">
          <CardTitle className="text-base font-semibold flex items-center gap-2">
            Budget vs Actual Time
            <InfoHint text="Planned, effective budget, and total session duration across all sessions." />
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <ChartContainer config={budgetChartConfig} className="h-56 w-full">
            <BarChart data={budgetData} margin={{ left: 10, right: 10 }}>
              <CartesianGrid vertical={false} />
              <XAxis dataKey="name" tickLine={false} axisLine={false} />
              <ChartTooltip content={<ChartTooltipContent hideLabel />} />
              <Bar dataKey="value" radius={8} fill="var(--color-budget)" />
            </BarChart>
          </ChartContainer>
          <div className="grid grid-cols-1 gap-3 text-sm sm:grid-cols-3">
            <div className="rounded-lg border border-slate-200 p-3">
              <p className="text-slate-500">Planned</p>
              <p className="font-semibold text-slate-900">
                {formatMinutes(totalPlannedSeconds)}
              </p>
            </div>
            <div className="rounded-lg border border-slate-200 p-3">
              <p className="text-slate-500">Effective Budget</p>
              <p className="font-semibold text-slate-900">
                {formatMinutes(effectiveBudgetSeconds)}
              </p>
            </div>
            <div className="rounded-lg border border-slate-200 p-3">
              <p className="text-slate-500">Actual Session Time</p>
              <p className="font-semibold text-slate-900">
                {formatMinutes(totalSessionDurationSeconds)}
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card className="lg:col-span-6 border-slate-200">
        <CardHeader className="pb-2">
          <CardTitle className="text-base font-semibold flex items-center gap-2">
            Timing Impact Drivers
            <InfoHint text="Pause, adjustment, overrun, and underrun totals; hover chart bars for values." />
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <ChartContainer config={impactsChartConfig} className="h-56 w-full">
            <BarChart data={timingImpactData} margin={{ left: 10, right: 10 }}>
              <CartesianGrid vertical={false} />
              <XAxis dataKey="name" tickLine={false} axisLine={false} />
              <ChartTooltip content={<ChartTooltipContent hideLabel />} />
              <Bar dataKey="value" radius={8} fill="var(--color-impact)" />
            </BarChart>
          </ChartContainer>
          <div className="grid grid-cols-2 gap-3 text-sm">
            <div className="rounded-lg border border-slate-200 p-3">
              <p className="text-slate-500">Total Pause Duration</p>
              <p className="font-semibold text-slate-900">
                {formatClock(totalPauseSeconds, "00:00")}
              </p>
            </div>
            <div className="rounded-lg border border-slate-200 p-3">
              <p className="text-slate-500">Pause Events</p>
              <p className="font-semibold text-slate-900">{totalPauseCount}</p>
            </div>
            <div className="rounded-lg border border-slate-200 p-3">
              <p className="text-slate-500">Net Adjustment</p>
              <p className="font-semibold text-slate-900">
                {formatSignedSeconds(totalAdjustmentSeconds)}
              </p>
            </div>
            <div className="rounded-lg border border-slate-200 p-3">
              <p className="text-slate-500">Overrun / Underrun</p>
              <p className="font-semibold text-slate-900">
                {formatClock(totalOverrunSeconds, "00:00")} /{" "}
                {formatClock(totalUnderrunSeconds, "00:00")}
              </p>
            </div>
          </div>
        </CardContent>
      </Card>
    </>
  );
}

function SessionCoordinationLogSection({
  sessions,
  isPending,
  error,
  onRefresh,
  getStatusColor,
}: {
  sessions: SessionSnapshot[];
  isPending: boolean;
  error: string | null;
  onRefresh: () => void;
  getStatusColor: (status: string) => string;
}) {
  return (
    <Card className="lg:col-span-6 border-slate-200">
      <CardHeader className="border-b border-slate-200 pb-3">
        <div className="flex items-center justify-between gap-3">
          <div>
            <CardTitle className="text-base font-semibold">
              Session Coordination Log
            </CardTitle>
            <p className="text-xs text-slate-500 mt-1">
              Showing latest 5 sessions for quick monitoring
            </p>
          </div>
          <SessionCreateModal onSuccess={onRefresh} />
        </div>
      </CardHeader>
      <CardContent className="p-0">
        {isPending && sessions.length === 0 ? (
          <div className="text-center py-10 text-slate-500">
            <Clock className="w-8 h-8 mx-auto mb-3 animate-spin text-slate-300" />
            <p className="text-sm">Loading sessions...</p>
          </div>
        ) : error ? (
          <div className="text-center py-10 text-red-600">
            <p className="text-sm">Error loading sessions: {error}</p>
          </div>
        ) : sessions.length === 0 ? (
          <div className="text-center py-10 text-slate-500">
            <Calendar className="w-8 h-8 mx-auto mb-3 text-slate-300" />
            <p className="text-sm">No sessions created yet</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse">
              <thead className="bg-slate-50 border-b border-slate-200">
                <tr>
                  <th className="p-3 text-[11px] font-semibold text-slate-600 uppercase tracking-wider">
                    Session
                  </th>
                  <th className="p-3 text-[11px] font-semibold text-slate-600 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="p-3 text-[11px] font-semibold text-slate-600 uppercase tracking-wider">
                    Duration
                  </th>
                  <th className="p-3 text-[11px] font-semibold text-slate-600 uppercase tracking-wider">
                    Action
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {sessions.slice(0, 5).map((session) => (
                  <tr
                    key={session.id}
                    className="hover:bg-slate-50 transition-colors"
                  >
                    <td className="p-3">
                      <div className="font-medium text-slate-900 text-sm leading-tight">
                        {session.title}
                      </div>
                      <div className="text-[11px] text-slate-500 mt-1">
                        {session.speakerName} • #{session.id.slice(0, 8)}
                      </div>
                    </td>
                    <td className="p-3">
                      <Badge
                        className={cn(
                          "text-[10px] font-semibold border",
                          getStatusColor(session.status),
                        )}
                      >
                        {session.status}
                      </Badge>
                    </td>
                    <td className="p-3 text-sm text-slate-700 font-mono whitespace-nowrap">
                      {formatClock(session.durationSeconds, "00:00")}
                    </td>
                    <td className="p-3">
                      <Link href={`/dashboard/sessions/${session.id}`}>
                        <Button
                          variant="outline"
                          size="sm"
                          className="h-8 rounded-full px-3"
                        >
                          View
                          <ExternalLink className="w-3 h-3 ml-1.5" />
                        </Button>
                      </Link>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

function DashboardPageComponent() {
  const [sessions, setSessions] = useState<SessionSnapshot[]>([]);
  const [overview, setOverview] = useState<AnalyticsOverview | null>(null);
  const [overviewFreshness, setOverviewFreshness] =
    useState<AnalyticsFreshness | null>(null);
  const [overviewSource, setOverviewSource] =
    useState<AnalyticsDataSource | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [isPending, startTransition] = useTransition();

  const loadSessions = () => {
    startTransition(async () => {
      const [sessionsResult, overviewResult] = await Promise.all([
        getSessionsList(),
        getAnalyticsOverview(),
      ]);

      if (sessionsResult.error) {
        setError(sessionsResult.error);
      } else {
        setSessions(sessionsResult.sessions);
      }

      if (overviewResult.error) {
        setOverview(null);
        setOverviewFreshness(overviewResult.freshness);
        setOverviewSource(overviewResult.source);
      } else {
        setOverview(overviewResult.overview);
        setOverviewFreshness(overviewResult.freshness);
        setOverviewSource(overviewResult.source);
      }
    });
  };

  useEffect(() => {
    loadSessions();
  }, []);

  const fallbackTotalSessions = sessions.length;
  const fallbackCreatedSessions = sessions.filter(
    (s) => s.status === "CREATED",
  ).length;
  const fallbackLiveSessions = sessions.filter(
    (s) => s.status === "LIVE",
  ).length;
  const fallbackPausedSessions = sessions.filter(
    (s) => s.status === "PAUSED",
  ).length;
  const fallbackEndedSessions = sessions.filter(
    (s) => s.status === "ENDED",
  ).length;
  const fallbackTodaySessions = sessions.filter((s) => {
    const createdAt = new Date(s.createdAt);
    const today = new Date();
    return createdAt.toDateString() === today.toDateString();
  }).length;
  const fallbackTotalSessionDuration = sessions.reduce(
    (acc, s) => acc + s.durationSeconds,
    0,
  );

  const totalSessions = overview?.totalSessions ?? fallbackTotalSessions;
  const createdSessions = overview?.createdSessions ?? fallbackCreatedSessions;
  const liveSessions = overview?.liveSessions ?? fallbackLiveSessions;
  const pausedSessions = overview?.pausedSessions ?? fallbackPausedSessions;
  const endedSessions = overview?.endedSessions ?? fallbackEndedSessions;

  const activeSessions = liveSessions + pausedSessions;
  const avgDuration =
    totalSessions > 0
      ? Math.round(
          (overview?.totalSessionDurationSeconds ??
            fallbackTotalSessionDuration) /
            totalSessions /
            60,
        )
      : 0;

  const totalProgramItems = overview?.totalProgramItems ?? 0;
  const endedProgramItems = overview?.endedProgramItems ?? 0;
  const onTimeEndedProgramItems = overview?.onTimeEndedProgramItems ?? 0;
  const overrunProgramItems = overview?.overrunProgramItems ?? 0;
  const remainingProgramItems = Math.max(
    totalProgramItems - endedProgramItems,
    0,
  );
  const otherEndedProgramItems = Math.max(
    endedProgramItems - onTimeEndedProgramItems - overrunProgramItems,
    0,
  );

  const totalSessionDurationSeconds =
    overview?.totalSessionDurationSeconds ?? fallbackTotalSessionDuration;
  const totalPlannedSeconds =
    overview?.totalPlannedSeconds ?? totalSessionDurationSeconds;
  const effectiveBudgetSeconds =
    overview?.effectiveBudgetSeconds ?? totalSessionDurationSeconds;
  const totalAdjustmentSeconds = overview?.totalAdjustmentSeconds ?? 0;
  const totalPauseSeconds = overview?.totalPauseSeconds ?? 0;
  const totalPauseCount = overview?.totalPauseCount ?? 0;
  const totalOverrunSeconds = overview?.totalOverrunSeconds ?? 0;
  const totalUnderrunSeconds = overview?.totalUnderrunSeconds ?? 0;

  const sessionCompletionPercent = clampPercent(
    (overview?.sessionCompletionRatio ??
      (totalSessions > 0 ? endedSessions / totalSessions : 0)) * 100,
  );
  const onTimePercent = clampPercent(
    (overview?.programItemOnTimeRatio ??
      (endedProgramItems > 0
        ? onTimeEndedProgramItems / endedProgramItems
        : 0)) * 100,
  );

  const sessionStatusData = useMemo(
    () => [
      { name: "Created", value: createdSessions, fill: "var(--color-created)" },
      { name: "Live", value: liveSessions, fill: "var(--color-live)" },
      { name: "Paused", value: pausedSessions, fill: "var(--color-paused)" },
      { name: "Ended", value: endedSessions, fill: "var(--color-ended)" },
    ],
    [createdSessions, liveSessions, pausedSessions, endedSessions],
  );

  const programOutcomeData = useMemo(
    () => [
      {
        name: "On Time",
        value: onTimeEndedProgramItems,
        fill: "var(--color-onTime)",
      },
      {
        name: "Overrun",
        value: overrunProgramItems,
        fill: "var(--color-overrun)",
      },
      {
        name: "Other Ended",
        value: otherEndedProgramItems,
        fill: "var(--color-otherEnded)",
      },
      {
        name: "Remaining",
        value: remainingProgramItems,
        fill: "var(--color-remaining)",
      },
    ],
    [
      onTimeEndedProgramItems,
      overrunProgramItems,
      otherEndedProgramItems,
      remainingProgramItems,
    ],
  );

  const budgetData = useMemo(
    () => [
      { name: "Planned", value: Math.round(totalPlannedSeconds / 60) },
      { name: "Effective", value: Math.round(effectiveBudgetSeconds / 60) },
      { name: "Actual", value: Math.round(totalSessionDurationSeconds / 60) },
    ],
    [totalPlannedSeconds, effectiveBudgetSeconds, totalSessionDurationSeconds],
  );

  const timingImpactData = useMemo(
    () => [
      { name: "Adjustment", value: Math.round(totalAdjustmentSeconds / 60) },
      { name: "Pause", value: Math.round(totalPauseSeconds / 60) },
      { name: "Overrun", value: Math.round(totalOverrunSeconds / 60) },
      { name: "Underrun", value: Math.round(totalUnderrunSeconds / 60) },
    ],
    [
      totalAdjustmentSeconds,
      totalPauseSeconds,
      totalOverrunSeconds,
      totalUnderrunSeconds,
    ],
  );

  const sessionStatusChartConfig = {
    value: { label: "Sessions" },
    created: { label: "Created", color: "hsl(var(--chart-3))" },
    live: { label: "Live", color: "hsl(var(--chart-2))" },
    paused: { label: "Paused", color: "hsl(var(--chart-4))" },
    ended: { label: "Ended", color: "hsl(var(--chart-1))" },
  } satisfies ChartConfig;

  const outcomesChartConfig = {
    value: { label: "Program Items" },
    onTime: { label: "On Time", color: "hsl(var(--chart-2))" },
    overrun: { label: "Overrun", color: "hsl(var(--chart-1))" },
    otherEnded: { label: "Other Ended", color: "hsl(var(--chart-4))" },
    remaining: { label: "Remaining", color: "hsl(var(--muted-foreground))" },
  } satisfies ChartConfig;

  const budgetChartConfig = {
    value: { label: "Minutes" },
    budget: { label: "Minutes", color: "hsl(var(--chart-1))" },
  } satisfies ChartConfig;

  const impactsChartConfig = {
    value: { label: "Minutes" },
    impact: { label: "Minutes", color: "hsl(var(--chart-5))" },
  } satisfies ChartConfig;

  const completionGaugeConfig = {
    completion: { label: "Completed", color: "hsl(var(--chart-2))" },
  } satisfies ChartConfig;

  const onTimeGaugeConfig = {
    onTime: { label: "On Time", color: "hsl(var(--chart-1))" },
  } satisfies ChartConfig;

  const getStatusColor = (status: string) => {
    switch (status) {
      case "LIVE":
        return "bg-emerald-100 text-emerald-700 border-emerald-200";
      case "PAUSED":
        return "bg-amber-100 text-amber-700 border-amber-200";
      case "ENDED":
        return "bg-slate-100 text-slate-600 border-slate-200";
      default:
        return "bg-blue-100 text-blue-700 border-blue-200";
    }
  };

  const analyticsHealth = deriveAnalyticsHealth(
    overviewFreshness,
    overviewSource,
  );
  const overviewAnalyticsSource = analyticsSourceLabel(
    overviewSource,
    Boolean(overview),
  );

  return (
    <div className="max-w-[1600px] mx-auto p-4 sm:p-6 lg:p-8 space-y-6">
      <GlobalTimeHealthSection
        totalSessions={totalSessions}
        activeSessions={activeSessions}
        todaySessions={fallbackTodaySessions}
        avgDuration={avgDuration}
        analyticsHealth={analyticsHealth}
        analyticsSource={overviewAnalyticsSource}
        freshness={overviewFreshness}
        onRefresh={loadSessions}
      />

      <section className="grid grid-cols-1 lg:grid-cols-12 gap-4">
        <AnalyticsSection
          sessionCompletionPercent={sessionCompletionPercent}
          endedSessions={endedSessions}
          totalSessions={totalSessions}
          onTimePercent={onTimePercent}
          onTimeEndedProgramItems={onTimeEndedProgramItems}
          endedProgramItems={endedProgramItems}
          completionGaugeConfig={completionGaugeConfig}
          onTimeGaugeConfig={onTimeGaugeConfig}
          sessionStatusChartConfig={sessionStatusChartConfig}
          sessionStatusData={sessionStatusData}
          outcomesChartConfig={outcomesChartConfig}
          programOutcomeData={programOutcomeData}
          totalProgramItems={totalProgramItems}
          overrunProgramItems={overrunProgramItems}
          budgetChartConfig={budgetChartConfig}
          budgetData={budgetData}
          totalPlannedSeconds={totalPlannedSeconds}
          effectiveBudgetSeconds={effectiveBudgetSeconds}
          totalSessionDurationSeconds={totalSessionDurationSeconds}
          impactsChartConfig={impactsChartConfig}
          timingImpactData={timingImpactData}
          totalPauseSeconds={totalPauseSeconds}
          totalPauseCount={totalPauseCount}
          totalAdjustmentSeconds={totalAdjustmentSeconds}
          totalOverrunSeconds={totalOverrunSeconds}
          totalUnderrunSeconds={totalUnderrunSeconds}
        />

        <SessionCoordinationLogSection
          sessions={sessions}
          isPending={isPending}
          error={error}
          onRefresh={loadSessions}
          getStatusColor={getStatusColor}
        />
      </section>
    </div>
  );
}

export default function DashboardPage() {
  return <DashboardPageComponent />;
}
