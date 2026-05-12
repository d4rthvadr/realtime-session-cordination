"use client";

import { useEffect, useState } from "react";
import { buildAdminApiUrl } from "@/lib/backend";
import type { SessionSnapshot } from "@/hooks/types";

export function useSessionSnapshot(sessionId: string) {
  const [session, setSession] = useState<SessionSnapshot | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    const loadSession = async () => {
      try {
        setIsLoading(true);
        const response = await fetch(
          buildAdminApiUrl(`/api/v1/sessions/${sessionId}`),
        );
        if (!response.ok) {
          throw new Error("session not found");
        }

        const payload = (await response.json()) as { session: SessionSnapshot };
        if (!cancelled) {
          setSession(payload.session);
          setError(null);
        }
      } catch {
        if (!cancelled) {
          setSession(null);
          setError("Could not load session state from the backend.");
        }
      } finally {
        if (!cancelled) {
          setIsLoading(false);
        }
      }
    };

    void loadSession();

    return () => {
      cancelled = true;
    };
  }, [sessionId]);

  return {
    session,
    setSession,
    isLoading,
    error,
  };
}
