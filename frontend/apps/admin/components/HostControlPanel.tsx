"use client";

import { useEffect } from "react";
import { formatClock, getPublicViewerUrl } from "@/lib/session";
import { useAdminSessionStore } from "@/store/adminSessionStore";

interface HostControlPanelProps {
  sessionId: string;
}

export default function HostControlPanel({ sessionId }: HostControlPanelProps) {
  const session = useAdminSessionStore((state) => state.sessions[sessionId]);
  const updateStatus = useAdminSessionStore((state) => state.updateStatus);
  const adjustRemainingTime = useAdminSessionStore(
    (state) => state.adjustRemainingTime,
  );
  const tickSession = useAdminSessionStore((state) => state.tickSession);

  useEffect(() => {
    const timer = setInterval(() => {
      tickSession(sessionId);
    }, 1000);
    return () => clearInterval(timer);
  }, [sessionId, tickSession]);

  if (!session) {
    return (
      <div className="rounded-2xl border border-amber-300 bg-amber-50 p-4 text-amber-800">
        Session not found in local admin store. Create a session first.
      </div>
    );
  }

  const canStart = session.status === "CREATED";
  const canPause = session.status === "LIVE";
  const canResume = session.status === "PAUSED";
  const canEnd = session.status === "LIVE" || session.status === "PAUSED";
  const viewerLink = getPublicViewerUrl(sessionId);

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

      <div className="grid gap-2 sm:grid-cols-2">
        <button
          disabled={!canStart}
          onClick={() => updateStatus(sessionId, "LIVE")}
          className="rounded-md bg-emerald-600 px-4 py-2 font-medium text-white disabled:cursor-not-allowed disabled:bg-slate-300"
        >
          Start
        </button>
        <button
          disabled={!canPause}
          onClick={() => updateStatus(sessionId, "PAUSED")}
          className="rounded-md bg-amber-500 px-4 py-2 font-medium text-white disabled:cursor-not-allowed disabled:bg-slate-300"
        >
          Pause
        </button>
        <button
          disabled={!canResume}
          onClick={() => updateStatus(sessionId, "LIVE")}
          className="rounded-md bg-sky-600 px-4 py-2 font-medium text-white disabled:cursor-not-allowed disabled:bg-slate-300"
        >
          Resume
        </button>
        <button
          disabled={!canEnd}
          onClick={() => updateStatus(sessionId, "ENDED")}
          className="rounded-md bg-rose-600 px-4 py-2 font-medium text-white disabled:cursor-not-allowed disabled:bg-slate-300"
        >
          End
        </button>
      </div>

      <div className="grid gap-2 sm:grid-cols-2">
        <button
          onClick={() => adjustRemainingTime(sessionId, 60)}
          className="rounded-md border border-slate-300 px-4 py-2 font-medium text-slate-900"
        >
          +60 seconds
        </button>
        <button
          onClick={() => adjustRemainingTime(sessionId, -60)}
          className="rounded-md border border-slate-300 px-4 py-2 font-medium text-slate-900"
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
