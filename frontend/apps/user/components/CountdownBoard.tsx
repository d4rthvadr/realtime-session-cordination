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
  LIVE: "border-emerald-500/40 bg-emerald-500/15 text-emerald-300",
  PAUSED: "border-amber-500/40 bg-amber-500/15 text-amber-300",
  ENDED: "border-slate-500/40 bg-slate-500/15 text-slate-300",
  CREATED: "border-blue-500/40 bg-blue-500/15 text-blue-300",
};

const TIMER_CLASS_BY_LEVEL = {
  safe: "text-emerald-400",
  warning: "text-amber-400",
  critical: "text-red-400",
  overtime: "text-red-500",
} as const;

export default function CountdownBoard({ sessionId }: CountdownBoardProps) {
  const title = useSessionStore((state) => state.title);
  const speakerName = useSessionStore((state) => state.speakerName);
  const status = useSessionStore((state) => state.status);
  const serverNowMs = useSessionStore((state) => state.serverNowMs);
  const connectionState = useSessionStore((state) => state.connectionState);
  const hasReceivedSnapshot = useSessionStore(
    (state) => state.hasReceivedSnapshot,
  );
  const sessionNotFound = useSessionStore((state) => state.sessionNotFound);
  const currentProgramItem = useSessionStore(
    (state) => state.currentProgramItem,
  );
  const nextProgramItem = useSessionStore((state) => state.nextProgramItem);
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
    if (!currentProgramItem) {
      return 0;
    }

    const baseRemaining =
      typeof currentProgramItem.remainingSeconds === "number"
        ? currentProgramItem.remainingSeconds
        : Number(currentProgramItem.remainingSeconds);

    if (!Number.isFinite(baseRemaining)) {
      return 0;
    }

    if (status !== "LIVE" || currentProgramItem.status !== "in_progress") {
      return baseRemaining;
    }

    const baseNow = typeof serverNowMs === "number" ? serverNowMs : nowMs;
    const elapsedSeconds = Math.max(0, Math.floor((nowMs - baseNow) / 1000));
    return baseRemaining - elapsedSeconds;
  }, [currentProgramItem, serverNowMs, status, nowMs]);

  const countdownBudgetSeconds = useMemo(() => {
    if (!currentProgramItem) {
      return 0;
    }

    const base =
      Number(currentProgramItem.runtimeDurationSeconds) +
      Number(currentProgramItem.adjustmentSeconds);
    return Number.isFinite(base) ? base : 0;
  }, [currentProgramItem]);

  const timerState = useMemo(
    () => getTimerState(remainingSeconds, countdownBudgetSeconds),
    [remainingSeconds, countdownBudgetSeconds],
  );

  const progressValue = useMemo(() => {
    if (
      !Number.isFinite(countdownBudgetSeconds) ||
      countdownBudgetSeconds <= 0
    ) {
      return 0;
    }

    const cappedRemaining = Math.max(0, remainingSeconds);
    const elapsed = countdownBudgetSeconds - cappedRemaining;
    return Math.min(100, Math.max(0, (elapsed / countdownBudgetSeconds) * 100));
  }, [countdownBudgetSeconds, remainingSeconds]);

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
        <Badge className="border-slate-700 bg-slate-800/80 text-slate-300">
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
            className={
              connectionState === "connected"
                ? "border-emerald-500/40 bg-emerald-500/15 text-emerald-300"
                : "border-amber-500/40 bg-amber-500/15 text-amber-300"
            }
          >
            {connectionState}
          </Badge>
        </div>
      </div>

      <Card className="border-slate-800 bg-slate-900/70 shadow-[0_0_0_1px_rgba(255,255,255,0.02),0_20px_60px_rgba(0,0,0,0.4)] backdrop-blur">
        <CardContent className="p-6 sm:p-10">
          <div className="text-center">
            <h1 className="text-2xl font-semibold text-slate-100 sm:text-4xl">
              {title}
            </h1>
            <p className="mt-2 text-sm text-slate-400 sm:text-base">
              Speaker:{" "}
              <span className="font-medium text-slate-200">{speakerName}</span>
            </p>
          </div>

          <div className="mt-10 text-center">
            <p className="text-xs uppercase tracking-[0.2em] text-slate-500">
              Remaining Time
            </p>
            <p
              className={`mt-2 text-6xl font-bold tracking-tight sm:text-8xl ${TIMER_CLASS_BY_LEVEL[timerState.level]}`}
            >
              {remainingSeconds < 0 ? "-" : ""}
              {formatDuration(remainingSeconds)}
              <span className="ml-2 text-3xl font-semibold text-slate-400 sm:text-5xl">
                / {formatDuration(countdownBudgetSeconds)}
              </span>
            </p>
          </div>

          <div className="mt-8">
            <Progress
              value={progressValue}
              className="h-2 bg-slate-800 [&_[data-state]]:bg-indigo-500"
            />
          </div>

          <div className="mt-6 flex flex-wrap items-center justify-center gap-2">
            <Badge
              variant="outline"
              className="border-slate-700 text-slate-400"
            >
              Urgency: {timerState.label}
            </Badge>
            <Badge
              variant="outline"
              className="border-slate-700 text-slate-400"
            >
              Session: {sessionId}
            </Badge>
          </div>

          <div className="mt-8 rounded-lg border border-slate-800 bg-slate-950/60 p-4">
            <p className="text-xs uppercase tracking-[0.2em] text-slate-500">
              Program Timeline
            </p>

            {currentProgramItem ? (
              <div className="mt-3 rounded-lg border border-emerald-500/20 bg-emerald-500/5 p-4">
                <div className="flex items-center gap-2">
                  <span className="h-2 w-2 rounded-full bg-emerald-400" />
                  <p className="text-xs uppercase tracking-[0.18em] text-emerald-300">
                    Now Live
                  </p>
                </div>
                <p className="mt-2 text-lg font-semibold text-slate-100 sm:text-xl">
                  {currentProgramItem.title}
                </p>
                <div className="mt-3 flex flex-wrap items-center gap-2">
                  <Badge
                    variant="outline"
                    className="border-indigo-500/40 bg-indigo-500/10 text-indigo-300"
                  >
                    {currentProgramItem.type.toUpperCase()}
                  </Badge>
                  <Badge
                    variant="outline"
                    className="border-slate-700 text-slate-300"
                  >
                    {currentProgramItem.status.replace("_", " ").toUpperCase()}
                  </Badge>
                  {currentProgramItem.hostName ? (
                    <Badge
                      variant="outline"
                      className="border-slate-700 text-slate-300"
                    >
                      Host: {currentProgramItem.hostName}
                    </Badge>
                  ) : null}
                  {currentProgramItem.location ? (
                    <Badge
                      variant="outline"
                      className="border-slate-700 text-slate-300"
                    >
                      Location: {currentProgramItem.location}
                    </Badge>
                  ) : null}
                </div>
              </div>
            ) : (
              <p className="mt-3 text-sm text-slate-400">
                No active program item right now.
              </p>
            )}

            {nextProgramItem ? (
              <div className="mt-4 rounded-lg border border-slate-800 bg-slate-900/60 p-4">
                <p className="text-xs uppercase tracking-[0.18em] text-slate-500">
                  Up Next
                </p>
                <p className="mt-2 text-base font-medium text-slate-200 sm:text-lg">
                  {nextProgramItem.title}
                </p>
                <div className="mt-3 flex flex-wrap items-center gap-2">
                  <Badge
                    variant="outline"
                    className="border-slate-700 text-slate-300"
                  >
                    {nextProgramItem.type.toUpperCase()}
                  </Badge>
                  {nextProgramItem.hostName ? (
                    <Badge
                      variant="outline"
                      className="border-slate-700 text-slate-300"
                    >
                      Host: {nextProgramItem.hostName}
                    </Badge>
                  ) : null}
                </div>
              </div>
            ) : null}
          </div>
        </CardContent>
      </Card>
    </section>
  );
}
