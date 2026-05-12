import { create } from "zustand";

export type SessionStatus = "CREATED" | "LIVE" | "PAUSED" | "ENDED";
export type ConnectionState = "connecting" | "connected" | "disconnected";

export interface SessionSnapshot {
  title: string;
  speakerName: string;
  durationSeconds: number;
  serverRemainingSeconds: number;
  status: SessionStatus;
  serverNowMs?: number;
}

interface SessionStore extends SessionSnapshot {
  connectionState: ConnectionState;
  hasReceivedSnapshot: boolean;
  sessionNotFound: boolean;
  setSnapshot: (snapshot: Partial<SessionSnapshot>) => void;
  setConnectionState: (state: ConnectionState) => void;
  setSessionNotFound: (notFound: boolean) => void;
  resetSession: () => void;
}

const DEFAULT_SESSION_VALUES: SessionSnapshot = {
  title: "",
  speakerName: "",
  durationSeconds: 0,
  serverRemainingSeconds: 0,
  status: "CREATED",
  serverNowMs: Date.now(),
};

export const useSessionStore = create<SessionStore>((set) => ({
  ...DEFAULT_SESSION_VALUES,
  connectionState: "connecting",
  hasReceivedSnapshot: false,
  sessionNotFound: false,
  setSnapshot: (snapshot) =>
    set({
      ...snapshot,
      serverNowMs: Date.now(),
      hasReceivedSnapshot: true,
      sessionNotFound: false,
    }),
  setConnectionState: (connectionState) => set({ connectionState }),
  setSessionNotFound: (sessionNotFound) => set({ sessionNotFound }),
  resetSession: () =>
    set({
      ...DEFAULT_SESSION_VALUES,
      connectionState: "connecting",
      hasReceivedSnapshot: false,
      sessionNotFound: false,
    }),
}));
