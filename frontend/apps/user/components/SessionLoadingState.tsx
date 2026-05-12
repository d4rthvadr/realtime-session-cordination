import type { ConnectionState } from "@/store/sessionStore";

interface SessionLoadingStateProps {
  sessionId: string;
  connectionState: ConnectionState;
}

export default function SessionLoadingState({
  sessionId,
  connectionState,
}: SessionLoadingStateProps) {
  return (
    <section className="mx-auto flex min-h-screen max-w-4xl flex-col justify-center px-6 py-10 text-center text-slate-100">
      <p className="text-xs uppercase tracking-[0.25em] text-slate-400">
        Session Viewer
      </p>
      <div className="mt-6 rounded-2xl border border-slate-800 bg-slate-900/70 p-8 shadow-2xl backdrop-blur">
        <div className="mx-auto h-10 w-56 animate-pulse rounded bg-slate-800" />
        <div className="mx-auto mt-4 h-6 w-40 animate-pulse rounded bg-slate-800" />
        <div className="mx-auto mt-8 h-20 w-64 animate-pulse rounded bg-slate-800" />
        <p className="mt-6 text-sm text-slate-300">
          {connectionState === "disconnected"
            ? "Waiting for live session data... reconnecting."
            : "Loading live session..."}
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
    </section>
  );
}
