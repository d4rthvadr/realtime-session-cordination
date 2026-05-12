import EmptyState from "@/components/EmptyState";

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

      <div className="mt-6">
        <EmptyState
          title="Session Not Found"
          description={`We could not find a live session for this link. Verify the session ID (${sessionId}) and try again.`}
          isDark
        />
      </div>
    </section>
  );
}
