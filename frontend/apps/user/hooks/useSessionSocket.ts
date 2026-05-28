"use client";

import { useEffect } from "react";
import { buildUserApiUrl, buildUserWsUrl } from "@/lib/backend";
import {
  useSessionStore,
  type ProgramItemSnapshot,
  type SessionSnapshot,
} from "@/store/sessionStore";

interface BackendSessionSnapshot {
  id: string;
  title: string;
  speakerName: string;
  durationSeconds: number;
  remainingSeconds: number;
  status: SessionSnapshot["status"];
  createdAt?: string;
}

interface BackendSessionResponse {
  session: BackendSessionSnapshot;
}

interface BackendCurrentProgramItemResponse {
  programItem: ProgramItemSnapshot | null;
}

function normalizeSnapshot(snapshot: BackendSessionSnapshot): SessionSnapshot {
  return {
    title: snapshot.title,
    speakerName: snapshot.speakerName,
    durationSeconds: snapshot.durationSeconds,
    serverRemainingSeconds: snapshot.remainingSeconds,
    status: snapshot.status,
    serverNowMs: Date.now(),
  };
}

export function useSessionSocket(sessionId: string): void {
  const setSnapshot = useSessionStore((state) => state.setSnapshot);
  const setCurrentProgramItem = useSessionStore(
    (state) => state.setCurrentProgramItem,
  );
  const setConnectionState = useSessionStore(
    (state) => state.setConnectionState,
  );
  const setSessionNotFound = useSessionStore(
    (state) => state.setSessionNotFound,
  );
  const resetSession = useSessionStore((state) => state.resetSession);

  useEffect(() => {
    let cancelled = false;
    let socket: WebSocket | null = null;
    let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
    let currentItemRefreshTimer: ReturnType<typeof setInterval> | null = null;

    resetSession();

    const refreshCurrentProgramItem = async () => {
      try {
        const response = await fetch(
          buildUserApiUrl(`/api/v1/sessions/${sessionId}/current-program-item`),
        );

        if (!response.ok) {
          return;
        }

        const payload =
          (await response.json()) as BackendCurrentProgramItemResponse;
        if (!cancelled) {
          setCurrentProgramItem(payload.programItem ?? null);
        }
      } catch {
        // Ignore refresh failures and keep last known value.
      }
    };

    const connect = async () => {
      setConnectionState("connecting");

      try {
        const response = await fetch(
          buildUserApiUrl(`/api/v1/sessions/${sessionId}`),
        );
        if (response.status === 404) {
          if (!cancelled) {
            setSessionNotFound(true);
            setConnectionState("disconnected");
          }
          return;
        }

        if (!response.ok) {
          throw new Error(`failed to load session ${sessionId}`);
        }

        const payload = (await response.json()) as BackendSessionResponse;
        if (!cancelled) {
          setSessionNotFound(false);
          setSnapshot(normalizeSnapshot(payload.session));
          await refreshCurrentProgramItem();
        }

        const wsUrl = buildUserWsUrl(`/ws/sessions/${sessionId}`);
        socket = new WebSocket(wsUrl);

        socket.onopen = () => {
          if (!cancelled) {
            setConnectionState("connected");
          }
        };

        socket.onmessage = (event) => {
          if (cancelled) {
            return;
          }

          try {
            const message = JSON.parse(String(event.data)) as {
              type?: string;
              session?: BackendSessionResponse["session"];
            };

            if (message.session) {
              setSnapshot(normalizeSnapshot(message.session));
              return;
            }

            if (
              message.type === "PROGRAM_ITEM_CREATED" ||
              message.type === "PROGRAM_ITEM_UPDATED" ||
              message.type === "PROGRAM_ITEM_CANCELED" ||
              message.type === "PROGRAM_ITEMS_REORDERED"
            ) {
              void refreshCurrentProgramItem();
              return;
            }
          } catch {
            // Ignore malformed messages during the smoke phase.
          }
        };

        socket.onclose = () => {
          if (cancelled) {
            return;
          }

          setConnectionState("disconnected");
          reconnectTimer = setTimeout(() => {
            if (!cancelled) {
              void connect();
            }
          }, 2000);
        };

        socket.onerror = () => {
          if (!cancelled) {
            setConnectionState("disconnected");
          }
        };

        currentItemRefreshTimer = setInterval(() => {
          void refreshCurrentProgramItem();
        }, 15000);
      } catch {
        if (!cancelled) {
          setConnectionState("disconnected");
        }
      }
    };

    void connect();

    return () => {
      cancelled = true;
      if (reconnectTimer) {
        clearTimeout(reconnectTimer);
      }
      if (currentItemRefreshTimer) {
        clearInterval(currentItemRefreshTimer);
      }
      if (socket) {
        socket.close();
      }
    };
  }, [
    resetSession,
    sessionId,
    setCurrentProgramItem,
    setConnectionState,
    setSessionNotFound,
    setSnapshot,
  ]);
}
