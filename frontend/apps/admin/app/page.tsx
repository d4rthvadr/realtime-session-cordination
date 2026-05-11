import SessionCreateForm from "@/components/SessionCreateForm";

export default function AdminHomePage() {
  return (
    <main className="min-h-screen bg-gradient-to-b from-slate-100 to-white px-6 py-10">
      <section className="mx-auto max-w-4xl space-y-8">
        <header className="space-y-2">
          <p className="text-xs uppercase tracking-[0.2em] text-slate-500">
            Realtime Session Coordination
          </p>
          <h1 className="text-3xl font-bold tracking-tight text-slate-900 sm:text-4xl">
            Admin Console
          </h1>
          <p className="text-slate-600">
            Create and control session timing before wiring to backend
            endpoints.
          </p>
        </header>

        <SessionCreateForm />
      </section>
    </main>
  );
}
