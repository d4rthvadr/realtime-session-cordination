import type { ConnectionState } from "@/store/sessionStore";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";

interface SessionLoadingStateProps {
  sessionId: string;
  connectionState: ConnectionState;
}

export default function SessionLoadingState({
  sessionId,
  connectionState,
}: SessionLoadingStateProps) {
  const isDisconnected = connectionState === "disconnected";

  return (
    <section className="mx-auto flex min-h-screen max-w-5xl flex-col justify-center px-4 py-8 sm:px-6 lg:px-8">
      <div className="mb-4 flex items-center justify-center gap-2">
        <Badge className="border-slate-700 bg-slate-800/80 text-slate-300">
          SESSION VIEWER
        </Badge>
        <Badge
          variant={isDisconnected ? "warning" : "success"}
          className={
            isDisconnected
              ? "border-amber-500/40 bg-amber-500/15 text-amber-300"
              : "border-emerald-500/40 bg-emerald-500/15 text-emerald-300"
          }
        >
          {connectionState}
        </Badge>
      </div>

      <Card className="border-slate-800 bg-slate-900/70 shadow-[0_0_0_1px_rgba(255,255,255,0.02),0_20px_60px_rgba(0,0,0,0.4)] backdrop-blur">
        <CardContent className="p-6 sm:p-10">
          <div className="mx-auto h-8 w-48 animate-pulse rounded-full bg-slate-800" />
          <div className="mx-auto mt-4 h-5 w-32 animate-pulse rounded-full bg-slate-800" />
          <div className="mx-auto mt-10 h-24 w-64 animate-pulse rounded-2xl bg-slate-800" />

          <p className="mt-8 text-center text-sm text-slate-400">
            {isDisconnected
              ? "Waiting for live session data. Attempting to reconnect."
              : "Loading live session."}
          </p>

          <div className="mt-8 flex flex-wrap items-center justify-center gap-3 text-sm">
            <Badge
              variant="outline"
              className="border-slate-700 text-slate-400"
            >
              Session: {sessionId}
            </Badge>
            <Badge
              variant="outline"
              className="border-slate-700 text-slate-400"
            >
              Connection: {connectionState}
            </Badge>
          </div>
        </CardContent>
      </Card>
    </section>
  );
}
