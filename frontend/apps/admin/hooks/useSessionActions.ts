"use client";

import { useCallback, useState } from "react";
import { buildAdminApiUrl } from "@/lib/backend";
import type { SessionSnapshot } from "@/hooks/types";

export type ActionState = "idle" | "loading" | "error";

export function useSessionActions(controlToken: string | null) {
  const [actionState, setActionState] = useState<ActionState>("idle");
  const [actionError, setActionError] = useState<string | null>(null);

  const runAction = useCallback(
    async (path: string, body?: unknown) => {
      setActionState("loading");
      setActionError(null);

      try {
        if (!controlToken) {
          throw new Error("Missing control token for this session");
        }

        const response = await fetch(buildAdminApiUrl(path), {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            "X-Control-Token": controlToken,
          },
          body: body ? JSON.stringify(body) : undefined,
        });

        if (!response.ok) {
          throw new Error(await response.text());
        }

        const payload = (await response.json()) as { session: SessionSnapshot };
        setActionState("idle");
        return payload.session;
      } catch (error) {
        setActionState("error");
        setActionError(
          error instanceof Error ? error.message : "Action failed",
        );
        return null;
      }
    },
    [controlToken],
  );

  return {
    actionState,
    actionError,
    runAction,
    clearActionError: () => setActionError(null),
  };
}
