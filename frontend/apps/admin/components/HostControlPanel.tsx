"use client";

import { useEffect, useMemo, useState } from "react";
import { buildAdminApiUrl, getViewerUrl } from "@/lib/backend";
import { formatClock } from "@/lib/session";

interface SessionSnapshot {
  id: string;
  title: string;
  speakerName: string;
  durationSeconds: number;
  status: "CREATED" | "LIVE" | "PAUSED" | "ENDED";
  remainingSeconds: number;
  createdAt?: string;
}

interface HostControlPanelProps {
  sessionId: string;
}

type ActionState = "idle" | "loading" | "error";

export default function HostControlPanel({ sessionId }: HostControlPanelProps) {
  const [session, setSession] = useState<SessionSnapshot | null>(null);
  const [actionState, setActionState] = useState<ActionState>("idle");
  const [message, setMessage] = useState<string | null>(null);
  const [controlToken, setControlToken] = useState<string | null>(null);

  useEffect(() => {
    setControlToken(window.sessionStorage.getItem(`controlToken:${sessionId}`));
  }, [sessionId]);

  useEffect(() => {
    let cancelled = false;

    const loadSession = async () => {
      try {
        const response = await fetch(
          buildAdminApiUrl(`/api/v1/sessions/${sessionId}`),
        );
        if (!response.ok) {
          throw new Error("session not found");
        }

        const payload = (await response.json()) as { session: SessionSnapshot };
        if (!cancelled) {
          setSession(payload.session);
        }
      } catch {
        if (!cancelled) {
          setMessage("Could not load session state from the backend.");
        }
      }
    };

    void loadSession();

    return () => {
      cancelled = true;
    };
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
  }, [session]);

  const viewerLink = useMemo(() => getViewerUrl(sessionId), [sessionId]);

  const authorizedFetch = async (path: string, init: RequestInit = {}) => {
    if (!controlToken) {
      throw new Error("Missing control token for this session");
    }

    const response = await fetch(buildAdminApiUrl(path), {
      ...init,
      headers: {
        "Content-Type": "application/json",
        "X-Control-Token": controlToken,
        ...(init.headers || {}),
      },
    });

    if (!response.ok) {
      throw new Error(await response.text());
    }

    return response;
  };

  const runAction = async (path: string, body?: unknown) => {
    setActionState("loading");
    setMessage(null);

    try {
      const response = await authorizedFetch(path, {
        method: "POST",
        body: body ? JSON.stringify(body) : undefined,
      });
      const payload = (await response.json()) as { session: SessionSnapshot };
      setSession(payload.session);
      setActionState("idle");
    } catch (error) {
      setActionState("error");
      setMessage(error instanceof Error ? error.message : "Action failed");
    }
  };

  if (!session) {
    return (
      <div className="rounded-2xl border border-amber-300 bg-amber-50 p-4 text-amber-800">
        {message || "Loading session from the backend..."}
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

      {message ? <p className="text-sm text-rose-700">{message}</p> : null}

      <div className="grid gap-2 sm:grid-cols-2">
        <button
          disabled={!canStart || actionState === "loading"}
          onClick={() => void runAction(`/api/v1/sessions/${sessionId}/start`)}
          className="rounded-md bg-emerald-600 px-4 py-2 font-medium text-white disabled:cursor-not-allowed disabled:bg-slate-300"
        >
          Start
        </button>
        <button
          disabled={!canPause || actionState === "loading"}
          onClick={() => void runAction(`/api/v1/sessions/${sessionId}/pause`)}
          className="rounded-md bg-amber-500 px-4 py-2 font-medium text-white disabled:cursor-not-allowed disabled:bg-slate-300"
        >
          Pause
        </button>
        <button
          disabled={!canResume || actionState === "loading"}
          onClick={() => void runAction(`/api/v1/sessions/${sessionId}/resume`)}
          className="rounded-md bg-sky-600 px-4 py-2 font-medium text-white disabled:cursor-not-allowed disabled:bg-slate-300"
        >
          Resume
        </button>
        <button
          disabled={!canEnd || actionState === "loading"}
          onClick={() => void runAction(`/api/v1/sessions/${sessionId}/end`)}
          className="rounded-md bg-rose-600 px-4 py-2 font-medium text-white disabled:cursor-not-allowed disabled:bg-slate-300"
        >
          End
        </button>
      </div>

      <div className="grid gap-2 sm:grid-cols-2">
        <button
          disabled={actionState === "loading"}
          onClick={() =>
            void runAction(`/api/v1/sessions/${sessionId}/adjust-time`, {
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
            void runAction(`/api/v1/sessions/${sessionId}/adjust-time`, {
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
