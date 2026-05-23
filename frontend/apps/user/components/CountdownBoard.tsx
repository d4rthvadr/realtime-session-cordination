"use client";

import { useEffect, useMemo, useState } from "react";
import SessionLoadingState from "@/components/SessionLoadingState";
import SessionNotFoundState from "@/components/SessionNotFoundState";
import { useSessionSocket } from "@/hooks/useSessionSocket";
import { formatDuration, getTimerState } from "@/lib/time";
import { useSessionStore } from "@/store/sessionStore";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";

interface CountdownBoardProps {
  sessionId: string;
}

const STATUS_BADGE_CLASS: Record<string, string> = {
  LIVE: "bg-emerald-100 text-emerald-700 border-emerald-200",
  PAUSED: "bg-amber-100 text-amber-700 border-amber-200",
  ENDED: "bg-slate-100 text-slate-600 border-slate-200",
  CREATED: "bg-blue-100 text-blue-700 border-blue-200",
};

const TIMER_CLASS_BY_LEVEL = {
  safe: "text-emerald-700",
  warning: "text-amber-600",
  critical: "text-red-600",
  overtime: "text-red-700",
} as const;

export default function CountdownBoard({ sessionId }: CountdownBoardProps) {
  const title = useSessionStore((state) => state.title);
  const speakerName = useSessionStore((state) => state.speakerName);
  const durationSeconds = useSessionStore((state) => state.durationSeconds);
  const serverRemainingSeconds = useSessionStore(
    (state) => state.serverRemainingSeconds,
  );
  const status = useSessionStore((state) => state.status);
  const serverNowMs = useSessionStore((state) => state.serverNowMs);
  const connectionState = useSessionStore((state) => state.connectionState);
  const hasReceivedSnapshot = useSessionStore(
    (state) => state.hasReceivedSnapshot,
  );
  const sessionNotFound = useSessionStore((state) => state.sessionNotFound);
  const [nowMs, setNowMs] = useState(() => Date.now());

  useSessionSocket(sessionId);

  useEffect(() => {
    setNowMs(Date.now());
    if (status !== "LIVE") {
      return;
    }

    const timer = setInterval(() => {
      setNowMs(Date.now());
    }, 1000);

    return () => clearInterval(timer);
  }, [status, serverNowMs]);

  const remainingSeconds = useMemo(() => {
    const baseRemaining =
      typeof serverRemainingSeconds === "number"
        ? serverRemainingSeconds
        : Number(serverRemainingSeconds);

    if (!Number.isFinite(baseRemaining)) {
      return 0;
    }

    if (status !== "LIVE") {
      return baseRemaining;
    }

    const baseNow = typeof serverNowMs === "number" ? serverNowMs : nowMs;
    const elapsedSeconds = Math.max(0, Math.floor((nowMs - baseNow) / 1000));
    return baseRemaining - elapsedSeconds;
  }, [serverNowMs, serverRemainingSeconds, status, nowMs]);

  const timerState = useMemo(
    () => getTimerState(remainingSeconds, durationSeconds),
    [remainingSeconds, durationSeconds],
  );

  const progressValue = useMemo(() => {
    if (!Number.isFinite(durationSeconds) || durationSeconds <= 0) {
      return 0;
    }

    const cappedRemaining = Math.max(0, remainingSeconds);
    const elapsed = durationSeconds - cappedRemaining;
    return Math.min(100, Math.max(0, (elapsed / durationSeconds) * 100));
  }, [durationSeconds, remainingSeconds]);

  const isLoadingInitialSession = !hasReceivedSnapshot;

  if (sessionNotFound) {
    return <SessionNotFoundState sessionId={sessionId} />;
  }

  if (isLoadingInitialSession) {
    return (
      <SessionLoadingState
        sessionId={sessionId}
        connectionState={connectionState}
      />
    );
  }

  return (
    <section className="mx-auto flex min-h-screen max-w-5xl flex-col justify-center px-4 py-8 sm:px-6 lg:px-8">
      <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
        <Badge className="bg-slate-100 text-slate-700 border-slate-200">
          SESSION VIEWER
        </Badge>
        <div className="flex items-center gap-2">
          <Badge
            className={STATUS_BADGE_CLASS[status] ?? STATUS_BADGE_CLASS.CREATED}
          >
            {status}
          </Badge>
          <Badge
            variant={connectionState === "connected" ? "success" : "warning"}
          >
            {connectionState}
          </Badge>
        </div>
      </div>

      <Card className="border-slate-200">
        <CardContent className="p-6 sm:p-10">
          <div className="text-center">
            <h1 className="text-2xl font-semibold text-slate-900 sm:text-4xl">
              {title}
            </h1>
            <p className="mt-2 text-sm text-slate-500 sm:text-base">
              Speaker:{" "}
              <span className="font-medium text-slate-700">{speakerName}</span>
            </p>
          </div>

          <div className="mt-10 text-center">
            <p className="text-xs uppercase tracking-[0.2em] text-slate-400">
              Remaining Time
            </p>
            <p
              className={`mt-2 text-6xl font-bold tracking-tight sm:text-8xl ${TIMER_CLASS_BY_LEVEL[timerState.level]}`}
            >
              {remainingSeconds < 0 ? "-" : ""}
              {formatDuration(remainingSeconds)}
              <span className="ml-2 text-3xl font-semibold text-slate-500 sm:text-5xl">
                / {formatDuration(durationSeconds)}
              </span>
            </p>
          </div>

          <div className="mt-8">
            <Progress value={progressValue} className="h-2 bg-slate-100" />
          </div>

          <div className="mt-6 flex flex-wrap items-center justify-center gap-2">
            <Badge
              variant="outline"
              className="border-slate-200 text-slate-600"
            >
              Urgency: {timerState.label}
            </Badge>
            <Badge
              variant="outline"
              className="border-slate-200 text-slate-600"
            >
              Session: {sessionId}
            </Badge>
          </div>
        </CardContent>
      </Card>
    </section>
  );
}
