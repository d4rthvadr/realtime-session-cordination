"use client";

import { useEffect, useMemo, useState, useTransition } from "react";
import {
  getSessionSnapshot,
  startSession,
  pauseSession,
  resumeSession,
  endSession,
  adjustSessionTime,
  SessionSnapshot,
} from "@/lib/actions";
import { formatClock } from "@/lib/session";
import { getViewerUrl } from "@/lib/backend";
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
  const [session, setSession] = useState<SessionSnapshot | null>(null);
  const [controlToken, setControlToken] = useState<string | null>(null);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);
  const [isPending, startTransition] = useTransition();

  // Mock data for demonstration - replace with real data
  const mockLogs: LogEntry[] = [
    {
      timestamp: "14:02:11",
      message: "System: Recording started automatically.",
      type: "success",
    },
    {
      timestamp: "14:05:45",
      message: "User: Sarah Chen shared screen.",
      type: "info",
    },
    {
      timestamp: "14:12:02",
      message: "Poll: '2024 Priorities' was published.",
      type: "info",
    },
    {
      timestamp: "14:25:30",
      message: "Admin: Microphones muted globally by moderator.",
      type: "warning",
    },
  ];

  // Load initial session and control token
  useEffect(() => {
    const token = window.sessionStorage.getItem(`controlToken:${sessionId}`);
    setControlToken(token);

    startTransition(async () => {
      const result = await getSessionSnapshot(sessionId);
      if (result.error) {
        setLoadError(result.error);
      } else if (result.session) {
        setSession(result.session);
      }
    });
  }, [sessionId]);

  // Auto-decrement time when session is running
  useEffect(() => {
    if (!session || session.status !== "LIVE") {
      return;
    }

    const timer = setInterval(() => {
      setSession((current) => {
        if (!current || current.status !== "LIVE") {
          return current;
        }
        return {
          ...current,
          remainingSeconds: Math.max(0, current.remainingSeconds - 1),
        };
      });
    }, 1000);

    return () => clearInterval(timer);
  }, [session, session?.status]);

  const viewerLink = useMemo(() => getViewerUrl(sessionId), [sessionId]);

  const handleAction = async (
    action: (
      token: string,
    ) => Promise<{ session: SessionSnapshot | null; error: string | null }>,
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
      } else if (result.session) {
        setSession(result.session);
      }
    });
  };

  const canStart = session?.status === "CREATED";
  const canPause = session?.status === "LIVE";
  const canResume = session?.status === "PAUSED";
  const canEnd = session?.status === "LIVE" || session?.status === "PAUSED";

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

  if (!session) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="text-center">
          <h2 className="text-2xl font-bold">Loading Session...</h2>
          <p className="text-muted-foreground mt-2">Fetching session data</p>
        </div>
      </div>
    );
  }

  const currentTime = formatClock(session.remainingSeconds, "00:00");
  const totalTime = formatClock(session.durationSeconds, "00:00");
  const progress =
    ((session.durationSeconds - session.remainingSeconds) /
      session.durationSeconds) *
    100;

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="border-b bg-card/50 backdrop-blur-sm sticky top-0 z-10">
        <div className="container mx-auto px-6 py-4">
          <div className="flex justify-between items-center">
            <div>
              <div className="flex items-center gap-3">
                <Link href="/sessions">
                  <Button variant="ghost" size="icon" className="rounded-full">
                    <ArrowLeft className="w-5 h-5" />
                  </Button>
                </Link>
                <div>
                  <h1 className="text-2xl font-semibold tracking-tight">
                    {session.title}
                  </h1>
                  <p className="text-sm text-muted-foreground">
                    {session.speakerName} • Session Control Dashboard
                  </p>
                </div>
              </div>
            </div>
            <div className="flex items-center gap-4">
              <div className="text-right">
                <p className="text-sm text-muted-foreground">Status</p>
                <p className="text-lg font-semibold">{session.status}</p>
              </div>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content - Bento Grid */}
      <main className="container mx-auto px-6 py-8">
        <div className="grid grid-cols-12 gap-4">
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

          {/* Session Controls */}
          <Card className="col-span-6">
            <CardHeader>
              <CardTitle className="text-lg">Session Controls</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              {actionError && (
                <div className="text-sm text-destructive bg-destructive/10 p-2 rounded">
                  {actionError}
                </div>
              )}

              <div className="grid grid-cols-2 gap-2">
                <Button
                  disabled={!canStart || isPending}
                  onClick={() =>
                    handleAction((token) => startSession(sessionId, token))
                  }
                  className="bg-emerald-600 hover:bg-emerald-700 rounded-full"
                >
                  <Play className="w-4 h-4 mr-2" />
                  Start
                </Button>
                <Button
                  disabled={!canPause || isPending}
                  onClick={() =>
                    handleAction((token) => pauseSession(sessionId, token))
                  }
                  className="bg-amber-500 hover:bg-amber-600 rounded-full"
                >
                  <PauseIcon className="w-4 h-4 mr-2" />
                  Pause
                </Button>
                <Button
                  disabled={!canResume || isPending}
                  onClick={() =>
                    handleAction((token) => resumeSession(sessionId, token))
                  }
                  className="bg-sky-600 hover:bg-sky-700 rounded-full"
                >
                  <RotateCcw className="w-4 h-4 mr-2" />
                  Resume
                </Button>
                <Button
                  disabled={!canEnd || isPending}
                  onClick={() =>
                    handleAction((token) => endSession(sessionId, token))
                  }
                  variant="destructive"
                  className="rounded-full"
                >
                  <Square className="w-4 h-4 mr-2" />
                  End
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
                  className="text-xs text-primary hover:underline flex items-center gap-1 break-all"
                >
                  {viewerLink}
                  <ExternalLink className="w-3 h-3 flex-shrink-0" />
                </a>
              </div>
            </CardContent>
          </Card>

          <AgendaProgress
            currentItem={3}
            totalItems={5}
            currentTitle="Q4 Resource Allocation"
            timeRemaining="12m"
            progress={60}
          />

          <SessionLog
            entries={mockLogs}
            onExport={() => console.log("Export logs")}
          />
        </div>
      </main>
    </div>
  );
}
