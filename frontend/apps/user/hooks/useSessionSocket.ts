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

interface BackendRuntimeEnvelope {
  type?: string;
  session: BackendSessionSnapshot;
  programItem?: ProgramItemSnapshot | null;
  nextProgramItem?: ProgramItemSnapshot | null;
}

interface BackendProgramItemMessage {
  type?: string;
  session?: BackendSessionResponse["session"];
  programItem?: ProgramItemSnapshot | null;
  nextProgramItem?: ProgramItemSnapshot | null;
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
  const setRuntimeSnapshot = useSessionStore(
    (state) => state.setRuntimeSnapshot,
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

    resetSession();

    const applyRuntime = (runtime: BackendRuntimeEnvelope) => {
      setRuntimeSnapshot({
        session: normalizeSnapshot(runtime.session),
        programItem: runtime.programItem ?? null,
        nextProgramItem: runtime.nextProgramItem ?? null,
      });
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

        const payload = (await response.json()) as BackendRuntimeEnvelope;
        if (!cancelled) {
          setSessionNotFound(false);
          applyRuntime(payload);
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
            const message = JSON.parse(
              String(event.data),
            ) as BackendProgramItemMessage;

            if (message.session) {
              applyRuntime({
                type: message.type,
                session: message.session,
                programItem: message.programItem,
                nextProgramItem: message.nextProgramItem,
              });
              return;
            }

            // Backward-compatible fallback: legacy session-only messages.
            if (message.type && !message.session) {
              setSnapshot({ serverNowMs: Date.now() });
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
  }, [
    resetSession,
    sessionId,
    setConnectionState,
    setSessionNotFound,
    setRuntimeSnapshot,
    setSnapshot,
  ]);
}
