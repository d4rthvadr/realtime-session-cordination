import BentoSessionView from "@/components/BentoSessionView";

interface SessionPageProps {
  params: {
    sessionId: string;
  };
}

export default function DashboardSessionPage({ params }: SessionPageProps) {
  return <BentoSessionView sessionId={params.sessionId} />;
}
