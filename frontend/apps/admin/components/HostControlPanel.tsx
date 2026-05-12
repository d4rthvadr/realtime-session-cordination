"use client";

import { useEffect, useMemo, useState } from "react";
import { getViewerUrl } from "@/lib/backend";
import { formatClock } from "@/lib/session";
import { useSessionActions } from "@/hooks/useSessionActions";
import { useSessionSnapshot } from "@/hooks/useSessionSnapshot";

interface HostControlPanelProps {
  sessionId: string;
}

type ActionState = "idle" | "loading" | "error";

export default function HostControlPanel({ sessionId }: HostControlPanelProps) {
  const {
    session,
    setSession,
    error: loadError,
  } = useSessionSnapshot(sessionId);
  const [controlToken, setControlToken] = useState<string | null>(null);
  const { actionState, actionError, runAction, clearActionError } =
    useSessionActions(controlToken);

  useEffect(() => {
    setControlToken(window.sessionStorage.getItem(`controlToken:${sessionId}`));
  }, [sessionId]);

  useEffect(() => {
    if (!session || session.status === "ENDED") {
      return;
    }

    const timer = setInterval(() => {
      setSession((current) => {
        if (!current || current.status !== "LIVE") {
          return current;
        }
        return {
          ...current,
          remainingSeconds: current.remainingSeconds - 1,
        };
      });
    }, 1000);

    return () => clearInterval(timer);
  }, [session, setSession]);

  const viewerLink = useMemo(() => getViewerUrl(sessionId), [sessionId]);

  const onRunAction = async (path: string, body?: unknown) => {
    clearActionError();
    const updated = await runAction(path, body);
    if (updated) {
      setSession(updated);
    }
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
          Session: {session.title} · Speaker: {session.speakerName}
        </p>
      </header>

      <div className="rounded-xl border border-slate-200 bg-slate-50 p-4">
        <p className="text-sm text-slate-500">Status</p>
        <p className="text-2xl font-semibold text-slate-900">
          {session.status}
        </p>
        <p className="mt-2 text-sm text-slate-500">Remaining</p>
        <p className="text-3xl font-bold text-slate-900">
          {formatClock(session.remainingSeconds)}
        </p>
      </div>

      {actionError ? (
        <p className="text-sm text-rose-700">{actionError}</p>
      ) : null}

      <div className="grid gap-2 sm:grid-cols-2">
        <button
          disabled={!canStart || actionState === "loading"}
          onClick={() =>
            void onRunAction(`/api/v1/sessions/${sessionId}/start`)
          }
          className="rounded-md bg-emerald-600 px-4 py-2 font-medium text-white disabled:cursor-not-allowed disabled:bg-slate-300"
        >
          Start
        </button>
        <button
          disabled={!canPause || actionState === "loading"}
          onClick={() =>
            void onRunAction(`/api/v1/sessions/${sessionId}/pause`)
          }
          className="rounded-md bg-amber-500 px-4 py-2 font-medium text-white disabled:cursor-not-allowed disabled:bg-slate-300"
        >
          Pause
        </button>
        <button
          disabled={!canResume || actionState === "loading"}
          onClick={() =>
            void onRunAction(`/api/v1/sessions/${sessionId}/resume`)
          }
          className="rounded-md bg-sky-600 px-4 py-2 font-medium text-white disabled:cursor-not-allowed disabled:bg-slate-300"
        >
          Resume
        </button>
        <button
          disabled={!canEnd || actionState === "loading"}
          onClick={() => void onRunAction(`/api/v1/sessions/${sessionId}/end`)}
          className="rounded-md bg-rose-600 px-4 py-2 font-medium text-white disabled:cursor-not-allowed disabled:bg-slate-300"
        >
          End
        </button>
      </div>

      <div className="grid gap-2 sm:grid-cols-2">
        <button
          disabled={actionState === "loading"}
          onClick={() =>
            void onRunAction(`/api/v1/sessions/${sessionId}/adjust-time`, {
              deltaSeconds: 60,
            })
          }
          className="rounded-md border border-slate-300 px-4 py-2 font-medium text-slate-900 disabled:cursor-not-allowed disabled:bg-slate-100"
        >
          +60 seconds
        </button>
        <button
          disabled={actionState === "loading"}
          onClick={() =>
            void onRunAction(`/api/v1/sessions/${sessionId}/adjust-time`, {
              deltaSeconds: -60,
            })
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
