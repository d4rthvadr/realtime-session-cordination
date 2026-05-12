interface SessionNotFoundStateProps {
  sessionId: string;
}

export default function SessionNotFoundState({
  sessionId,
}: SessionNotFoundStateProps) {
  return (
    <section className="mx-auto flex min-h-screen max-w-3xl flex-col justify-center px-6 py-10 text-center text-slate-100">
      <p className="text-xs uppercase tracking-[0.25em] text-slate-400">
        Session Viewer
      </p>

      <div className="mt-6 rounded-2xl border border-amber-700/40 bg-slate-900/70 p-8 shadow-2xl backdrop-blur">
        <h1 className="text-3xl font-semibold text-amber-300 sm:text-4xl">
          Session Not Found
        </h1>
        <p className="mt-4 text-slate-300">
          We could not find a live session for this link.
        </p>
        <p className="mt-2 text-sm text-slate-400">
          Verify the session ID and try again.
        </p>

        <div className="mt-6 inline-flex rounded-full border border-slate-700 px-4 py-1 text-sm text-slate-300">
          Session: {sessionId}
        </div>
      </div>
    </section>
  );
}
