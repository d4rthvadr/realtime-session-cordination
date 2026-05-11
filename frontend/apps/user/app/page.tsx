import Link from "next/link";

export default function HomePage() {
  return (
    <main className="min-h-screen bg-slate text-slate-100">
      <section className="mx-auto flex min-h-screen max-w-3xl flex-col items-center justify-center gap-6 px-6 text-center">
        <p className="rounded-full border border-slate-700 px-4 py-1 text-xs uppercase tracking-[0.2em] text-slate-300">
          Realtime Session Coordination
        </p>
        <h1 className="text-4xl font-semibold tracking-tight sm:text-5xl">
          Public Countdown Viewer
        </h1>
        <p className="max-w-xl text-base text-slate-300 sm:text-lg">
          Open a session page to follow the live timer. This app is
          intentionally public and read-only for audience and speaker
          visibility.
        </p>
        <Link
          href="/sessions/demo-session"
          className="rounded-md bg-slate-100 px-6 py-3 font-medium text-slate-900 transition hover:bg-white"
        >
          Open Demo Session
        </Link>
      </section>
    </main>
  );
}
