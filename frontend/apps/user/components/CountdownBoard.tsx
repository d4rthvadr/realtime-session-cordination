"use client";

import { useEffect, useMemo } from "react";
import { useSessionSocket } from "@/hooks/useSessionSocket";
import { formatDuration, getTimerState } from "@/lib/time";
import { useSessionStore } from "@/store/sessionStore";

interface CountdownBoardProps {
  sessionId: string;
}

export default function CountdownBoard({ sessionId }: CountdownBoardProps) {
  const title = useSessionStore((state) => state.title);
  const speakerName = useSessionStore((state) => state.speakerName);
  const durationSeconds = useSessionStore((state) => state.durationSeconds);
  const serverRemainingSeconds = useSessionStore(
    (state) => state.serverRemainingSeconds,
  );
  const status = useSessionStore((state) => state.status);
  const connectionState = useSessionStore((state) => state.connectionState);
  const tickFromClient = useSessionStore((state) => state.tickFromClient);

  useSessionSocket(sessionId);

  useEffect(() => {
    const interval = setInterval(() => {
      tickFromClient();
    }, 1000);

    return () => clearInterval(interval);
  }, [tickFromClient]);

  const timerState = useMemo(
    () => getTimerState(serverRemainingSeconds, durationSeconds),
    [serverRemainingSeconds, durationSeconds],
  );

  return (
    <section className="mx-auto flex min-h-screen max-w-4xl flex-col justify-center px-6 py-10 text-center text-slate-100">
      <p className="text-xs uppercase tracking-[0.25em] text-slate-400">
        Session Viewer
      </p>
      <h1 className="mt-3 text-3xl font-semibold sm:text-5xl">{title}</h1>
      <p className="mt-2 text-lg text-slate-300 sm:text-xl">
        Speaker: {speakerName}
      </p>

      <div className="mt-10 rounded-2xl border border-slate-800 bg-slate-900/70 p-8 shadow-2xl backdrop-blur">
        <p className="text-sm uppercase tracking-[0.2em] text-slate-400">
          Remaining Time
        </p>
        <p
          className={`mt-2 text-7xl font-bold tracking-tight sm:text-8xl ${timerState.colorClass}`}
        >
          {serverRemainingSeconds < 0 ? "-" : ""}
          {formatDuration(serverRemainingSeconds)}
        </p>
        <p className="mt-4 text-sm text-slate-300">
          State: <span className="font-semibold text-slate-100">{status}</span>{" "}
          · Urgency:{" "}
          <span className="font-semibold text-slate-100">
            {timerState.label}
          </span>
        </p>
      </div>

      <div className="mt-8 flex flex-wrap items-center justify-center gap-3 text-sm">
        <span className="rounded-full border border-slate-700 px-4 py-1 text-slate-300">
          Session: {sessionId}
        </span>
        <span className="rounded-full border border-slate-700 px-4 py-1 text-slate-300">
          Connection: {connectionState}
        </span>
      </div>

      <p className="mt-8 text-sm text-slate-400">
        This view is currently wired to mock realtime events and will be
        connected to backend WebSocket events in integration.
      </p>
    </section>
  );
}
