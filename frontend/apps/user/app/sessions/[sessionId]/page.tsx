import CountdownBoard from "@/components/CountdownBoard";

interface SessionPageParams {
  params: {
    sessionId: string;
  };
}

export default function SessionPage({ params }: SessionPageParams) {
  return (
    <main className="min-h-screen bg-background">
      <CountdownBoard sessionId={params.sessionId} />
    </main>
  );
}
