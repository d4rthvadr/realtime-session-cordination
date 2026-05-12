"use client";

import { useEffect, useMemo, useState, useTransition } from "react";
import { getViewerUrl } from "@/lib/backend";
import { formatClock } from "@/lib/session";
import {
  getSessionSnapshot,
  startSession,
  pauseSession,
  resumeSession,
  endSession,
  adjustSessionTime,
  SessionSnapshot,
} from "@/lib/actions";

interface HostControlPanelProps {
  sessionId: string;
}

export default function HostControlPanel({ sessionId }: HostControlPanelProps) {
  const [session, setSession] = useState<SessionSnapshot | null>(null);
  const [controlToken, setControlToken] = useState<string | null>(null);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);
  const [isPending, startTransition] = useTransition();

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

  if (!session) {
    return (
      <div className="rounded-2xl border border-amber-300 bg-amber-50 p-4 text-amber-800">
        {loadError || "Loading session from the backend..."}
      </div>
    );
  }

  const canStart = session.status === "CREATED";
  const canPause = session.status === "LIVE";
  const canResume = session.status === "PAUSED";
  const canEnd = session.status === "LIVE" || session.status === "PAUSED";

  return (
    <section className="space-y-5 rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
      <header>
        <h2 className="text-xl font-semibold text-slate-900">Host Controls</h2>
        <p className="text-sm text-slate-600">
          {session.title} • {session.speakerName}
        </p>
      </header>

      <div className="rounded-xl border border-slate-200 bg-slate-50 p-4">
        <p className="text-sm text-slate-500">Status</p>
        <p className="text-2xl font-semibold text-slate-900">
          {session.status}
        </p>
        <p className="mt-2 text-sm text-slate-500">Time Remaining</p>
        <p className="text-3xl font-bold text-slate-900">
          {formatClock(session.remainingSeconds, "--:--")}
        </p>
      </div>

      {actionError ? (
        <p className="text-sm text-rose-700">{actionError}</p>
      ) : null}

      <div className="grid gap-2 sm:grid-cols-2">
        <button
          disabled={!canStart || isPending}
          onClick={() =>
            handleAction((token) => startSession(sessionId, token))
          }
          className="rounded-md bg-emerald-600 px-4 py-2 font-medium text-white disabled:cursor-not-allowed disabled:bg-slate-300"
        >
          Start
        </button>
        <button
          disabled={!canPause || isPending}
          onClick={() =>
            handleAction((token) => pauseSession(sessionId, token))
          }
          className="rounded-md bg-amber-500 px-4 py-2 font-medium text-white disabled:cursor-not-allowed disabled:bg-slate-300"
        >
          Pause
        </button>
        <button
          disabled={!canResume || isPending}
          onClick={() =>
            handleAction((token) => resumeSession(sessionId, token))
          }
          className="rounded-md bg-sky-600 px-4 py-2 font-medium text-white disabled:cursor-not-allowed disabled:bg-slate-300"
        >
          Resume
        </button>
        <button
          disabled={!canEnd || isPending}
          onClick={() => handleAction((token) => endSession(sessionId, token))}
          className="rounded-md bg-rose-600 px-4 py-2 font-medium text-white disabled:cursor-not-allowed disabled:bg-slate-300"
        >
          End
        </button>
      </div>

      <div className="grid gap-2 sm:grid-cols-2">
        <button
          disabled={isPending}
          onClick={() =>
            handleAction((token) =>
              adjustSessionTime(sessionId, { deltaSeconds: 60 }, token),
            )
          }
          className="rounded-md border border-slate-300 px-4 py-2 font-medium text-slate-900 disabled:cursor-not-allowed disabled:bg-slate-100"
        >
          +60 seconds
        </button>
        <button
          disabled={isPending}
          onClick={() =>
            handleAction((token) =>
              adjustSessionTime(sessionId, { deltaSeconds: -60 }, token),
            )
          }
          className="rounded-md border border-slate-300 px-4 py-2 font-medium text-slate-900 disabled:cursor-not-allowed disabled:bg-slate-100"
        >
          -60 seconds
        </button>
      </div>

      <div>
        <p className="text-xs uppercase tracking-[0.15em] text-slate-500">
          Share Link
        </p>
        <a
          className="break-all text-sm font-medium text-sky-700 underline"
          href={viewerLink}
          target="_blank"
          rel="noreferrer"
        >
          {viewerLink}
        </a>
      </div>
    </section>
  );
}
