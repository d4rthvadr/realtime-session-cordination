import CountdownBoard from "@/components/CountdownBoard";

interface SessionPageParams {
  params: {
    sessionId: string;
  };
}

export default function SessionPage({ params }: SessionPageParams) {
  return (
    <main className="min-h-screen bg-[#0a0b0f] bg-[radial-gradient(circle_at_20%_0%,rgba(99,102,241,0.14),transparent_45%),radial-gradient(circle_at_80%_100%,rgba(34,197,94,0.1),transparent_40%)]">
      <CountdownBoard sessionId={params.sessionId} />
    </main>
  );
}
