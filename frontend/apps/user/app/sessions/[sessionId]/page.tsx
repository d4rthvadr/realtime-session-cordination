import CountdownBoard from "@/components/CountdownBoard";

interface SessionPageParams {
  params: {
    sessionId: string;
  };
}

export default function SessionPage({ params }: SessionPageParams) {
  return (
    <main className="min-h-screen bg-slate">
      <CountdownBoard sessionId={params.sessionId} />
    </main>
  );
}
