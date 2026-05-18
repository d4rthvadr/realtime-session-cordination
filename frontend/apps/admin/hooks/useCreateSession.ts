"use client";

import { useCallback, useState } from "react";
import { buildAdminApiUrl } from "@/lib/backend";

interface CreateSessionInput {
  title: string;
  speakerName: string;
  durationSeconds: number;
}

interface CreateSessionResponse {
  session: { id: string };
  controlToken: string;
}

export function useCreateSession() {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const createSession = useCallback(async (input: CreateSessionInput) => {
    setIsSubmitting(true);
    setError(null);

    try {
      const response = await fetch(buildAdminApiUrl("/api/v1/sessions"), {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(input),
      });

      if (!response.ok) {
        throw new Error("Failed to create session");
      }

      const payload = (await response.json()) as CreateSessionResponse;
      return payload;
    } catch {
      setError("Could not create the session in the backend.");
      return null;
    } finally {
      setIsSubmitting(false);
    }
  }, []);

  return {
    createSession,
    isSubmitting,
    error,
    clearError: () => setError(null),
  };
}
