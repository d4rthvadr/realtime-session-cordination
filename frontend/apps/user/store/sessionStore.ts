import { create } from "zustand";

export type SessionStatus = "CREATED" | "LIVE" | "PAUSED" | "ENDED";
export type ConnectionState =
  | "mocked"
  | "connecting"
  | "connected-mock"
  | "connected"
  | "disconnected";

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
  tickFromClient: () => void;
  setConnectionState: (state: ConnectionState) => void;
  setSessionNotFound: (notFound: boolean) => void;
  resetSession: () => void;
}

const INITIAL_DURATION_SECONDS = 25 * 60;

const DEFAULT_SESSION_VALUES: SessionSnapshot = {
  title: "Demo Session",
  speakerName: "Sample Speaker",
  durationSeconds: INITIAL_DURATION_SECONDS,
  serverRemainingSeconds: INITIAL_DURATION_SECONDS,
  status: "CREATED",
  serverNowMs: Date.now(),
};

export const useSessionStore = create<SessionStore>((set) => ({
  ...DEFAULT_SESSION_VALUES,
  connectionState: "mocked",
  hasReceivedSnapshot: false,
  sessionNotFound: false,
  setSnapshot: (snapshot) =>
    set({
      ...snapshot,
      serverNowMs: Date.now(),
      hasReceivedSnapshot: true,
      sessionNotFound: false,
    }),
  tickFromClient: () =>
    set((state) => {
      if (state.status !== "LIVE") {
        return state;
      }

      const elapsedSeconds = (Date.now() - state.serverNowMs!) / 1000;
      return {
        serverRemainingSeconds: state.serverRemainingSeconds - elapsedSeconds,
        serverNowMs: Date.now(),
      };
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
