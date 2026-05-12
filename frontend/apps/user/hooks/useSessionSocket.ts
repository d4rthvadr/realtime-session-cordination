"use client";

import { useEffect } from "react";
import { buildUserApiUrl, buildUserWsUrl } from "@/lib/backend";
import { useSessionStore, type SessionSnapshot } from "@/store/sessionStore";

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
  const setConnectionState = useSessionStore(
    (state) => state.setConnectionState,
  );

  useEffect(() => {
    let cancelled = false;
    let socket: WebSocket | null = null;
    let reconnectTimer: ReturnType<typeof setTimeout> | null = null;

    const connect = async () => {
      setConnectionState("connecting");

      try {
        const response = await fetch(
          buildUserApiUrl(`/api/v1/sessions/${sessionId}`),
        );
        if (!response.ok) {
          throw new Error(`failed to load session ${sessionId}`);
        }

        const payload = (await response.json()) as BackendSessionResponse;
        if (!cancelled) {
          setSnapshot(normalizeSnapshot(payload.session));
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
      if (socket) {
        socket.close();
      }
    };
  }, [sessionId, setConnectionState, setSnapshot]);
}
