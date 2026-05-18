import HostControlPanel from "../../../components/HostControlPanel";
import Link from "next/link";

interface SessionPageProps {
  params: {
    sessionId: string;
  };
}

export default function SessionAdminPage({ params }: SessionPageProps) {
  return (
    <main className="min-h-screen bg-gradient-to-b from-slate-100 to-white px-6 py-10">
      <section className="mx-auto max-w-4xl space-y-6">
        <header>
          <p className="text-xs uppercase tracking-[0.2em] text-slate-500">
            Host Session Controls
          </p>
          <h1 className="text-3xl font-bold text-slate-900">Session Admin</h1>
          <Link
            href="/sessions"
            className="mt-2 inline-flex text-sm font-medium text-sky-700 underline"
          >
            Back to Sessions List
          </Link>
        </header>
        <HostControlPanel sessionId={params.sessionId} />
      </section>
    </main>
  );
}
