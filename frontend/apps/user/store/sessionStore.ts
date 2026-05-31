import { create } from "zustand";

export type SessionStatus = "CREATED" | "LIVE" | "PAUSED" | "ENDED";
export type ConnectionState = "connecting" | "connected" | "disconnected";
export type ProgramItemStatus =
  | "scheduled"
  | "in_progress"
  | "paused"
  | "ended"
  | "canceled";

export interface ProgramItemSnapshot {
  id: string;
  sessionId: string;
  title: string;
  type: string;
  status: ProgramItemStatus;
  runtimeDurationSeconds: number;
  remainingSeconds: number;
  actualStart?: string;
  pausedAt?: string;
  totalPausedDurationSeconds: number;
  adjustmentSeconds: number;
  endedRemainingSeconds?: number;
  actualEnd?: string;
  pauseCount: number;
  endedReason?: string;
  hostName?: string;
  scheduledStart: string;
  scheduledEnd: string;
  expectedDurationMinutes: number;
  position: number;
  location?: string;
  metadata?: Record<string, unknown>;
  createdAt: string;
  updatedAt: string;
}

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
  currentProgramItem: ProgramItemSnapshot | null;
  nextProgramItem: ProgramItemSnapshot | null;
  setRuntimeSnapshot: (runtime: {
    session: Partial<SessionSnapshot>;
    programItem: ProgramItemSnapshot | null;
    nextProgramItem: ProgramItemSnapshot | null;
  }) => void;
  setSnapshot: (snapshot: Partial<SessionSnapshot>) => void;
  setCurrentProgramItem: (item: ProgramItemSnapshot | null) => void;
  setNextProgramItem: (item: ProgramItemSnapshot | null) => void;
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
  currentProgramItem: null,
  nextProgramItem: null,
  setRuntimeSnapshot: ({ session, programItem, nextProgramItem }) =>
    set({
      ...session,
      currentProgramItem: programItem,
      nextProgramItem,
      serverNowMs: Date.now(),
      hasReceivedSnapshot: true,
      sessionNotFound: false,
    }),
  setSnapshot: (snapshot) =>
    set({
      ...snapshot,
      serverNowMs: Date.now(),
      hasReceivedSnapshot: true,
      sessionNotFound: false,
    }),
  setCurrentProgramItem: (currentProgramItem) => set({ currentProgramItem }),
  setNextProgramItem: (nextProgramItem) => set({ nextProgramItem }),
  setConnectionState: (connectionState) => set({ connectionState }),
  setSessionNotFound: (sessionNotFound) => set({ sessionNotFound }),
  resetSession: () =>
    set({
      ...DEFAULT_SESSION_VALUES,
      connectionState: "connecting",
      hasReceivedSnapshot: false,
      sessionNotFound: false,
      currentProgramItem: null,
      nextProgramItem: null,
    }),
}));
