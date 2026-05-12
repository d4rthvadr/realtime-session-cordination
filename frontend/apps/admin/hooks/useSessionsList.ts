"use client";

import { useEffect, useState } from "react";
import { buildAdminApiUrl } from "@/lib/backend";
import type { SessionSnapshot } from "@/hooks/types";

export function useSessionsList() {
  const [sessions, setSessions] = useState<SessionSnapshot[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    const loadSessions = async () => {
      try {
        setIsLoading(true);
        const response = await fetch(buildAdminApiUrl("/api/v1/sessions"));
        if (!response.ok) {
          throw new Error("failed to fetch sessions");
        }

        const payload = (await response.json()) as {
          sessions: SessionSnapshot[];
        };
        if (!cancelled) {
          setSessions(payload.sessions || []);
          setError(null);
        }
      } catch {
        if (!cancelled) {
          setError("Could not load sessions from backend.");
        }
      } finally {
        if (!cancelled) {
          setIsLoading(false);
        }
      }
    };

    void loadSessions();

    return () => {
      cancelled = true;
    };
  }, []);

  return {
    sessions,
    isLoading,
    error,
  };
}
