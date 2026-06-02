"use client";

import { useEffect, useMemo, useState, useTransition } from "react";
import {
  getSessionSnapshot,
  getProgramItems,
  getSessionLogs,
  getSessionAnalytics,
  createProgramItem,
  cancelProgramItem,
  startProgramItem,
  pauseProgramItem,
  resumeProgramItem,
  adjustProgramItemTime,
  endProgramItem,
  reorderProgramItems,
  startSession,
  pauseSession,
  resumeSession,
  endSession,
  adjustSessionTime,
  RuntimeSnapshot,
  ProgramItemSnapshot,
  ProgramItemCreateInput,
  SessionLogSnapshot,
  SessionAnalyticsSummary,
} from "@/lib/actions";
import { formatClock } from "@/lib/session";
import { buildAdminWsUrl, getViewerUrl } from "@/lib/backend";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  ArrowLeft,
  MicOff,
  Share2,
  Radio,
  MoreHorizontal,
  Play,
  Pause as PauseIcon,
  Square,
  RotateCcw,
  Plus,
  Minus,
  ExternalLink,
} from "lucide-react";

// Import widgets
import TimerWidget from "@/components/widgets/TimerWidget";
import AttendeeStats from "@/components/widgets/AttendeeStats";
import StatusCard, {
  SignalIcon,
  CPUIcon,
} from "@/components/widgets/StatusCard";
import QuickActions from "@/components/widgets/QuickActions";
import SessionLog, { LogEntry } from "@/components/widgets/SessionLog";
import AgendaProgress from "@/components/widgets/AgendaProgress";

interface BentoSessionViewProps {
  sessionId: string;
}

export default function BentoSessionView({ sessionId }: BentoSessionViewProps) {
  const [runtime, setRuntime] = useState<RuntimeSnapshot | null>(null);
  const [programItems, setProgramItems] = useState<ProgramItemSnapshot[]>([]);
  const [sessionLogs, setSessionLogs] = useState<SessionLogSnapshot[]>([]);
  const [analytics, setAnalytics] = useState<SessionAnalyticsSummary | null>(
    null,
  );
  const [controlToken, setControlToken] = useState<string | null>(null);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);
  const [programItemError, setProgramItemError] = useState<string | null>(null);
  const [logError, setLogError] = useState<string | null>(null);
  const [analyticsError, setAnalyticsError] = useState<string | null>(null);
  const [isPending, startTransition] = useTransition();

  // Load initial session and control token
  useEffect(() => {
    const token = window.sessionStorage.getItem(`controlToken:${sessionId}`);
    setControlToken(token);

    startTransition(async () => {
      const [sessionResult, programItemsResult, logsResult, analyticsResult] =
        await Promise.all([
          getSessionSnapshot(sessionId),
          getProgramItems(sessionId),
          getSessionLogs(sessionId, { limit: 100, offset: 0 }),
          getSessionAnalytics(sessionId),
        ]);

      if (sessionResult.error) {
        setLoadError(sessionResult.error);
      } else if (sessionResult.runtime) {
        setRuntime(sessionResult.runtime);
      }

      if (programItemsResult.error) {
        setProgramItemError(programItemsResult.error);
      } else {
        setProgramItems(programItemsResult.programItems);
      }

      if (logsResult.error) {
        setLogError(logsResult.error);
      } else {
        setLogError(null);
        setSessionLogs(logsResult.logs);
      }

      if (analyticsResult.error) {
        setAnalyticsError(analyticsResult.error);
      } else {
        setAnalyticsError(null);
        setAnalytics(analyticsResult.analytics);
      }
    });
  }, [sessionId]);

  useEffect(() => {
    let socket: WebSocket | null = null;
    let closed = false;

    try {
      socket = new WebSocket(buildAdminWsUrl(`/ws/sessions/${sessionId}`));
    } catch {
      return;
    }

    socket.onmessage = (event) => {
      if (closed) {
        return;
      }

      try {
        const payload = JSON.parse(String(event.data)) as {
          event?: string;
          sessionLog?: SessionLogSnapshot;
        };

        if (payload.event !== "session_log_appended" || !payload.sessionLog) {
          return;
        }

        const appendedLog = payload.sessionLog;

        setSessionLogs((current) => {
          const exists = current.some((entry) => entry.id === appendedLog.id);
          if (exists) {
            return current;
          }

          const merged: SessionLogSnapshot[] = [appendedLog, ...current];
          merged.sort((a, b) => {
            const tA = Date.parse(a.occurredAt);
            const tB = Date.parse(b.occurredAt);
            if (tA !== tB) {
              return tB - tA;
            }

            const cA = Date.parse(a.createdAt);
            const cB = Date.parse(b.createdAt);
            if (cA !== cB) {
              return cB - cA;
            }

            return b.id.localeCompare(a.id);
          });

          return merged.slice(0, 200);
        });
      } catch {
        // Ignore non-log websocket payloads.
      }
    };

    return () => {
      closed = true;
      if (socket) {
        socket.close();
      }
    };
  }, [sessionId]);

  // Keep runtime countdown smooth between server updates.
  useEffect(() => {
    if (!runtime || runtime.session.status !== "LIVE") {
      return;
    }

    if (!runtime.programItem || runtime.programItem.status !== "in_progress") {
      return;
    }

    const timer = setInterval(() => {
      setRuntime((current) => {
        if (!current || current.session.status !== "LIVE") {
          return current;
        }

        const currentItem = current.programItem;
        if (!currentItem || currentItem.status !== "in_progress") {
          return current;
        }

        return {
          ...current,
          programItem: {
            ...currentItem,
            remainingSeconds: currentItem.remainingSeconds - 1,
          },
        };
      });
    }, 1000);

    return () => clearInterval(timer);
  }, [runtime]);

  const viewerLink = useMemo(() => getViewerUrl(sessionId), [sessionId]);

  const logEntries = useMemo<LogEntry[]>(() => {
    return sessionLogs.map((entry) => ({
      timestamp: formatLogTime(entry.occurredAt),
      message: entry.message,
      type: logTypeFromEvent(entry.eventType),
    }));
  }, [sessionLogs]);

  const handleExportLogs = () => {
    if (sessionLogs.length === 0) {
      return;
    }

    const header = [
      "id",
      "sessionId",
      "programItemId",
      "eventType",
      "message",
      "occurredAt",
      "createdAt",
      "requestId",
      "metadata",
    ];

    const rows = sessionLogs.map((entry) => [
      entry.id,
      entry.sessionId,
      entry.programItemId ?? "",
      entry.eventType,
      entry.message,
      entry.occurredAt,
      entry.createdAt,
      entry.requestId ?? "",
      JSON.stringify(entry.metadata ?? {}),
    ]);

    const csv = [header, ...rows]
      .map((row) => row.map((field) => csvEscape(field)).join(","))
      .join("\n");

    const blob = new Blob([csv], { type: "text/csv;charset=utf-8" });
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = `${sessionId}-session-logs.csv`;
    anchor.click();
    URL.revokeObjectURL(url);
  };

  const handleAction = async (
    action: (
      token: string,
    ) => Promise<{ runtime: RuntimeSnapshot | null; error: string | null }>,
  ) => {
    if (!controlToken) {
      setActionError("No control token available");
      return;
    }

    setActionError(null);
    startTransition(async () => {
      const result = await action(controlToken);
      if (result.error) {
        setActionError(result.error);
      } else if (result.runtime) {
        setRuntime(result.runtime);

        const listResult = await getProgramItems(sessionId);
        if (!listResult.error) {
          setProgramItems(listResult.programItems);
        }
      }
    });
  };

  const session = runtime?.session ?? null;
  const canStart = session?.status === "CREATED";
  const canPause = session?.status === "LIVE";
  const canResume = session?.status === "PAUSED";
  const canEnd = session?.status === "LIVE" || session?.status === "PAUSED";
  const isProgramItemRuntimeAllowed =
    session?.status === "LIVE" || session?.status === "PAUSED";

  const handleCreateProgramItem = (input: ProgramItemCreateInput) => {
    if (!controlToken) {
      setProgramItemError("No control token available");
      return;
    }

    setProgramItemError(null);
    startTransition(async () => {
      const result = await createProgramItem(sessionId, input, controlToken);
      if (result.error) {
        setProgramItemError(result.error);
        return;
      }
      const programItem = result.programItem;
      if (programItem) {
        setProgramItems((current) =>
          [...current, programItem].sort((a, b) => a.position - b.position),
        );
      }
    });
  };

  const handleCancelProgramItem = (itemId: string) => {
    if (!controlToken) {
      setProgramItemError("No control token available");
      return;
    }

    setProgramItemError(null);
    startTransition(async () => {
      const result = await cancelProgramItem(itemId, controlToken);
      if (result.error) {
        setProgramItemError(result.error);
        return;
      }
      const programItem = result.programItem;
      if (programItem) {
        setProgramItems((current) =>
          current.map((item) => (item.id === itemId ? programItem : item)),
        );
      }
    });
  };

  const handleStartProgramItem = (itemId: string) => {
    if (!controlToken) {
      setProgramItemError("No control token available");
      return;
    }

    setProgramItemError(null);
    startTransition(async () => {
      const result = await startProgramItem(itemId, controlToken);
      if (result.error) {
        setProgramItemError(result.error);
        return;
      }

      if (result.runtime) {
        setRuntime(result.runtime);
      }

      const listResult = await getProgramItems(sessionId);
      if (!listResult.error) {
        setProgramItems(listResult.programItems);
      }
    });
  };

  const handleEndProgramItem = (itemId: string) => {
    if (!controlToken) {
      setProgramItemError("No control token available");
      return;
    }

    setProgramItemError(null);
    startTransition(async () => {
      const result = await endProgramItem(itemId, controlToken);
      if (result.error) {
        setProgramItemError(result.error);
        return;
      }

      if (result.runtime) {
        setRuntime(result.runtime);
      }

      const listResult = await getProgramItems(sessionId);
      if (!listResult.error) {
        setProgramItems(listResult.programItems);
      }
    });
  };

  const handlePauseProgramItem = (itemId: string) => {
    if (!controlToken) {
      setProgramItemError("No control token available");
      return;
    }

    setProgramItemError(null);
    startTransition(async () => {
      const result = await pauseProgramItem(itemId, controlToken);
      if (result.error) {
        setProgramItemError(result.error);
        return;
      }

      if (result.runtime) {
        setRuntime(result.runtime);
      }

      const listResult = await getProgramItems(sessionId);
      if (!listResult.error) {
        setProgramItems(listResult.programItems);
      }
    });
  };

  const handleResumeProgramItem = (itemId: string) => {
    if (!controlToken) {
      setProgramItemError("No control token available");
      return;
    }

    setProgramItemError(null);
    startTransition(async () => {
      const result = await resumeProgramItem(itemId, controlToken);
      if (result.error) {
        setProgramItemError(result.error);
        return;
      }

      if (result.runtime) {
        setRuntime(result.runtime);
      }

      const listResult = await getProgramItems(sessionId);
      if (!listResult.error) {
        setProgramItems(listResult.programItems);
      }
    });
  };

  const handleAdjustProgramItemTime = (
    itemId: string,
    deltaSeconds: number,
  ) => {
    if (!controlToken) {
      setProgramItemError("No control token available");
      return;
    }

    setProgramItemError(null);
    startTransition(async () => {
      const result = await adjustProgramItemTime(
        itemId,
        { deltaSeconds },
        controlToken,
      );
      if (result.error) {
        setProgramItemError(result.error);
        return;
      }

      if (result.runtime) {
        setRuntime(result.runtime);
      }

      const listResult = await getProgramItems(sessionId);
      if (!listResult.error) {
        setProgramItems(listResult.programItems);
      }
    });
  };

  const handleReorderProgramItems = (
    items: Array<{ id: string; position: number }>,
  ) => {
    if (!controlToken) {
      setProgramItemError("No control token available");
      return;
    }

    setProgramItemError(null);
    startTransition(async () => {
      const result = await reorderProgramItems(
        sessionId,
        { items },
        controlToken,
      );
      if (result.error) {
        setProgramItemError(result.error);
        return;
      }
      setProgramItems(result.programItems);
    });
  };

  const quickActions = [
    {
      icon: <MicOff className="w-6 h-6" />,
      label: "Mute All",
      onClick: () => console.log("Mute all"),
      variant: "danger" as const,
    },
    {
      icon: <Share2 className="w-6 h-6" />,
      label: "Share Feed",
      onClick: () => console.log("Share feed"),
      variant: "primary" as const,
    },
    {
      icon: <Radio className="w-6 h-6" />,
      label: "Broadcast",
      onClick: () => console.log("Broadcast"),
      variant: "primary" as const,
    },
    {
      icon: <MoreHorizontal className="w-6 h-6" />,
      label: "More",
      onClick: () => console.log("More options"),
    },
  ];

  if (loadError) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="text-center">
          <h2 className="text-2xl font-bold">Session Not Found</h2>
          <p className="text-muted-foreground mt-2">{loadError}</p>
          <Link href="/sessions">
            <Button className="mt-4" variant="outline">
              <ArrowLeft className="w-4 h-4 mr-2" />
              Back to Sessions
            </Button>
          </Link>
        </div>
      </div>
    );
  }

  if (!runtime || !session) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="text-center">
          <h2 className="text-2xl font-bold">Loading Session...</h2>
          <p className="text-muted-foreground mt-2">Fetching session data</p>
        </div>
      </div>
    );
  }

  const currentCountdownSeconds = runtime.programItem
    ? runtime.programItem.remainingSeconds
    : 0;
  const totalCountdownSeconds = runtime.programItem
    ? runtime.programItem.runtimeDurationSeconds +
      runtime.programItem.adjustmentSeconds
    : 0;
  const currentTime = formatClock(currentCountdownSeconds, "00:00");
  const totalTime = formatClock(totalCountdownSeconds, "00:00");
  const progress =
    totalCountdownSeconds > 0
      ? ((totalCountdownSeconds - currentCountdownSeconds) /
          totalCountdownSeconds) *
        100
      : 0;

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="border-b bg-card/50 backdrop-blur-sm sticky top-0 z-10">
        <div className="container mx-auto px-4 sm:px-6 py-3 sm:py-4">
          <div className="flex justify-between items-center">
            <div>
              <div className="flex items-center gap-3">
                <Link href="/dashboard">
                  <Button variant="ghost" size="icon" className="rounded-full">
                    <ArrowLeft className="w-5 h-5" />
                  </Button>
                </Link>
                <div>
                  <h1 className="text-lg sm:text-xl md:text-2xl font-semibold tracking-tight">
                    {session.title}
                  </h1>
                  <p className="text-xs sm:text-sm text-muted-foreground">
                    {session.speakerName}{" "}
                    <span className="hidden sm:inline">
                      • Session Control Dashboard
                    </span>
                  </p>
                </div>
              </div>
            </div>
            <div className="flex items-center gap-2 sm:gap-4">
              <div className="text-right">
                <p className="text-xs sm:text-sm text-muted-foreground">
                  Status
                </p>
                <p className="text-sm sm:text-lg font-semibold">
                  {session.status}
                </p>
              </div>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content - Bento Grid */}
      <main className="container mx-auto px-4 sm:px-6 py-4 sm:py-8">
        <div className="grid grid-cols-12 gap-3 sm:gap-4">
          {/* Top Row: Timer + Attendee Stats */}
          <TimerWidget
            currentTime={currentTime}
            totalTime={totalTime}
            progress={progress}
            status={session.status as "CREATED" | "LIVE" | "PAUSED" | "ENDED"}
            onPause={() =>
              canPause &&
              handleAction((token) => pauseSession(sessionId, token))
            }
            onRefresh={() => window.location.reload()}
          />
          <AttendeeStats
            totalOnline={124}
            participationRate={88}
            attentionLevel="High"
          />

          <Card className="col-span-12 md:col-span-6">
            <CardHeader className="pb-3">
              <CardTitle className="text-base sm:text-lg">
                Analytics Snapshot
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-3 sm:space-y-4">
              {analyticsError ? (
                <div className="text-xs sm:text-sm text-destructive bg-destructive/10 p-2 sm:p-3 rounded">
                  {analyticsError}
                </div>
              ) : analytics ? (
                <div className="grid grid-cols-2 gap-3 text-sm">
                  <div>
                    <p className="text-muted-foreground">Program Items</p>
                    <p className="text-xl font-semibold">
                      {analytics.programItemCount}
                    </p>
                  </div>
                  <div>
                    <p className="text-muted-foreground">Ended On Time</p>
                    <p className="text-xl font-semibold">
                      {analytics.endedOnTimeCount}
                    </p>
                  </div>
                  <div>
                    <p className="text-muted-foreground">Overrun (s)</p>
                    <p className="text-xl font-semibold">
                      {analytics.totalOverrunSeconds}
                    </p>
                  </div>
                  <div>
                    <p className="text-muted-foreground">Pause Count</p>
                    <p className="text-xl font-semibold">
                      {analytics.totalPauseCount}
                    </p>
                  </div>
                  <div className="col-span-2 pt-2 border-t">
                    <p className="text-muted-foreground">On-Time Ratio</p>
                    <p className="text-xl font-semibold">
                      {(analytics.endedOnTimeRatio * 100).toFixed(1)}%
                    </p>
                  </div>
                </div>
              ) : (
                <p className="text-sm text-muted-foreground">
                  Loading analytics...
                </p>
              )}
            </CardContent>
          </Card>

          {/* Session Controls */}
          <Card className="col-span-12 md:col-span-6">
            <CardHeader className="pb-3">
              <CardTitle className="text-base sm:text-lg">
                Session Controls
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-3 sm:space-y-4">
              {actionError && (
                <div className="text-xs sm:text-sm text-destructive bg-destructive/10 p-2 sm:p-3 rounded">
                  {actionError}
                </div>
              )}
              {logError && (
                <div className="text-xs sm:text-sm text-destructive bg-destructive/10 p-2 sm:p-3 rounded">
                  {logError}
                </div>
              )}

              <div className="grid grid-cols-2 gap-2">
                <Button
                  disabled={!canStart || isPending}
                  onClick={() =>
                    handleAction((token) => startSession(sessionId, token))
                  }
                  className="bg-emerald-600 hover:bg-emerald-700 rounded-full h-10 sm:h-11"
                >
                  <Play className="w-4 h-4 mr-2" />
                  <span className="text-sm sm:text-base">Start</span>
                </Button>
                <Button
                  disabled={!canPause || isPending}
                  onClick={() =>
                    handleAction((token) => pauseSession(sessionId, token))
                  }
                  className="bg-amber-500 hover:bg-amber-600 rounded-full h-10 sm:h-11"
                >
                  <PauseIcon className="w-4 h-4 mr-2" />
                  <span className="text-sm sm:text-base">Pause</span>
                </Button>
                <Button
                  disabled={!canResume || isPending}
                  onClick={() =>
                    handleAction((token) => resumeSession(sessionId, token))
                  }
                  className="bg-sky-600 hover:bg-sky-700 rounded-full h-10 sm:h-11"
                >
                  <RotateCcw className="w-4 h-4 mr-2" />
                  <span className="text-sm sm:text-base">Resume</span>
                </Button>
                <Button
                  disabled={!canEnd || isPending}
                  onClick={() =>
                    handleAction((token) => endSession(sessionId, token))
                  }
                  variant="destructive"
                  className="rounded-full h-10 sm:h-11"
                >
                  <Square className="w-4 h-4 mr-2" />
                  <span className="text-sm sm:text-base">End</span>
                </Button>
              </div>

              <div className="pt-2 border-t">
                <p className="text-xs text-muted-foreground mb-2">
                  Time Adjustment
                </p>
                <div className="grid grid-cols-2 gap-2">
                  <Button
                    disabled={isPending}
                    onClick={() =>
                      handleAction((token) =>
                        adjustSessionTime(
                          sessionId,
                          { deltaSeconds: 60 },
                          token,
                        ),
                      )
                    }
                    variant="outline"
                    size="sm"
                    className="rounded-full"
                  >
                    <Plus className="w-4 h-4 mr-1" />
                    60s
                  </Button>
                  <Button
                    disabled={isPending}
                    onClick={() =>
                      handleAction((token) =>
                        adjustSessionTime(
                          sessionId,
                          { deltaSeconds: -60 },
                          token,
                        ),
                      )
                    }
                    variant="outline"
                    size="sm"
                    className="rounded-full"
                  >
                    <Minus className="w-4 h-4 mr-1" />
                    60s
                  </Button>
                </div>
              </div>

              <div className="pt-2 border-t">
                <p className="text-xs text-muted-foreground mb-1">
                  Viewer Link
                </p>
                <a
                  href={viewerLink}
                  target="_blank"
                  rel="noreferrer"
                  className="text-xs text-primary hover:underline flex items-center gap-1 break-all leading-relaxed"
                >
                  <span className="break-all">{viewerLink}</span>
                  <ExternalLink className="w-3 h-3 flex-shrink-0 inline-block" />
                </a>
              </div>
            </CardContent>
          </Card>

          <AgendaProgress
            items={programItems}
            isPending={isPending}
            error={programItemError}
            onCreateAction={handleCreateProgramItem}
            onCancelAction={handleCancelProgramItem}
            onStartAction={handleStartProgramItem}
            onPauseAction={handlePauseProgramItem}
            onResumeAction={handleResumeProgramItem}
            onEndAction={handleEndProgramItem}
            onAdjustTimeAction={handleAdjustProgramItemTime}
            onReorderAction={handleReorderProgramItems}
            runtimeEnabled={isProgramItemRuntimeAllowed}
          />

          <SessionLog entries={logEntries} onExport={handleExportLogs} />
        </div>
      </main>
    </div>
  );
}

function formatLogTime(value: string): string {
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) {
    return "--:--:--";
  }

  return parsed.toLocaleTimeString([], {
    hour12: false,
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
}

function logTypeFromEvent(eventType: string): LogEntry["type"] {
  if (eventType.includes("ENDED") || eventType.includes("CANCELED")) {
    return "warning";
  }
  if (eventType.includes("FAILED") || eventType.includes("ERROR")) {
    return "error";
  }
  if (
    eventType.includes("STARTED") ||
    eventType.includes("RESUMED") ||
    eventType.includes("CREATED")
  ) {
    return "success";
  }
  return "info";
}

function csvEscape(value: unknown): string {
  const stringValue = String(value ?? "");
  return `"${stringValue.replace(/"/g, '""')}"`;
}
