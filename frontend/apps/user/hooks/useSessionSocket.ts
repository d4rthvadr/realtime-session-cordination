"use client";

import { useEffect } from "react";
import { useSessionStore, type SessionSnapshot } from "@/store/sessionStore";

interface MockEvent {
  delayMs: number;
  payload: Partial<SessionSnapshot>;
}

const MOCK_EVENTS: MockEvent[] = [
  {
    delayMs: 1000,
    payload: {
      status: "LIVE",
    },
  },
  {
    delayMs: 1500,
    payload: {
      serverRemainingSeconds: 24 * 60 + 35,
    },
  },
];

export function useSessionSocket(sessionId: string): void {
  const setSnapshot = useSessionStore((state) => state.setSnapshot);
  const setConnectionState = useSessionStore(
    (state) => state.setConnectionState,
  );

  useEffect(() => {
    let cancelled = false;
    const timers: NodeJS.Timeout[] = [];

    setConnectionState("connecting");

    timers.push(
      setTimeout(() => {
        if (cancelled) {
          return;
        }
        setConnectionState("connected-mock");
      }, 200),
    );

    for (const event of MOCK_EVENTS) {
      timers.push(
        setTimeout(() => {
          if (cancelled) {
            return;
          }

          setSnapshot({
            ...event.payload,
            title: `Session ${sessionId}`,
            speakerName: "TBD Speaker",
          });
        }, event.delayMs),
      );
    }

    return () => {
      cancelled = true;
      for (const timer of timers) {
        clearTimeout(timer);
      }
    };
  }, [sessionId, setConnectionState, setSnapshot]);
}
