import BentoSessionView from "@/components/BentoSessionView";
import { cookies } from "next/headers";

interface SessionPageProps {
  params: {
    sessionId: string;
  };
}

export default function DashboardSessionPage({ params }: SessionPageProps) {
  const wsAccessToken = cookies().get("admin_auth_token")?.value ?? null;
  return (
    <BentoSessionView
      sessionId={params.sessionId}
      wsAccessToken={wsAccessToken}
    />
  );
}
