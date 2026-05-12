"use client";

import Link from "next/link";
import { formatClock } from "@/lib/session";
import { useSessionsList } from "@/hooks/useSessionsList";

export default function SessionsListPage() {
  const { sessions, isLoading, error } = useSessionsList();

  return (
    <main className="min-h-screen bg-gradient-to-b from-slate-100 to-white px-6 py-10">
      <section className="mx-auto max-w-5xl space-y-6">
        <header className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <p className="text-xs uppercase tracking-[0.2em] text-slate-500">
              Realtime Session Coordination
            </p>
            <h1 className="text-3xl font-bold tracking-tight text-slate-900 sm:text-4xl">
              Sessions
            </h1>
            <p className="text-slate-600">
              Browse all sessions and open host controls.
            </p>
          </div>
          <Link
            href="/"
            className="inline-flex rounded-md border border-slate-300 bg-white px-4 py-2 text-sm font-medium text-slate-800 transition hover:border-slate-400 hover:bg-slate-50"
          >
            Back To Admin Home
          </Link>
        </header>

        {isLoading ? (
          <div className="rounded-2xl border border-slate-200 bg-white p-6 text-slate-600 shadow-sm">
            Loading sessions...
          </div>
        ) : null}

        {error ? (
          <div className="rounded-2xl border border-rose-300 bg-rose-50 p-6 text-rose-700 shadow-sm">
            {error}
          </div>
        ) : null}

        {!isLoading && !error && sessions.length === 0 ? (
          <div className="rounded-2xl border border-slate-200 bg-white p-6 text-slate-600 shadow-sm">
            No sessions found.
          </div>
        ) : null}

        {!isLoading && !error && sessions.length > 0 ? (
          <div className="grid gap-4">
            {sessions.map((session) => (
              <article
                key={session.id}
                className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm"
              >
                <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                  <div className="space-y-1">
                    <h2 className="text-lg font-semibold text-slate-900">
                      {session.title}
                    </h2>
                    <p className="text-sm text-slate-600">
                      Speaker: {session.speakerName}
                    </p>
                    <p className="text-sm text-slate-500">
                      Session ID: {session.id}
                    </p>
                    <p className="text-sm text-slate-500">
                      Created:{" "}
                      {session.createdAt
                        ? new Date(session.createdAt).toLocaleString()
                        : "-"}
                    </p>
                  </div>

                  <div className="text-left sm:text-right">
                    <p className="text-xs uppercase tracking-[0.12em] text-slate-500">
                      Status
                    </p>
                    <p className="text-base font-semibold text-slate-900">
                      {session.status}
                    </p>
                    <p className="mt-2 text-xs uppercase tracking-[0.12em] text-slate-500">
                      Remaining
                    </p>
                    <p className="text-xl font-bold text-slate-900">
                      {formatClock(session.remainingSeconds)}
                    </p>
                  </div>
                </div>

                <div className="mt-4">
                  <Link
                    href={`/sessions/${session.id}`}
                    className="inline-flex rounded-md bg-slate-900 px-4 py-2 text-sm font-medium text-white transition hover:bg-slate-700"
                  >
                    Open Host Controls
                  </Link>
                </div>
              </article>
            ))}
          </div>
        ) : null}
      </section>
    </main>
  );
}
