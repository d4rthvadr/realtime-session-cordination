import EmptyState from "@/components/EmptyState";

interface SessionNotFoundStateProps {
  sessionId: string;
}

export default function SessionNotFoundState({
  sessionId,
}: SessionNotFoundStateProps) {
  return (
    <section className="mx-auto flex min-h-screen max-w-4xl flex-col justify-center px-4 py-8 sm:px-6 lg:px-8">
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
