import { create } from "zustand";

export type SessionStatus = "CREATED" | "LIVE" | "PAUSED" | "ENDED";

export interface SessionDraft {
  title: string;
  speakerName: string;
  durationSeconds: number;
}

export interface SessionRecord extends SessionDraft {
  id: string;
  status: SessionStatus;
  remainingSeconds: number;
  createdAtIso: string;
}

interface AdminSessionState {
  sessions: Record<string, SessionRecord>;
  createSession: (draft: SessionDraft) => SessionRecord;
  getSession: (id: string) => SessionRecord | undefined;
  updateStatus: (id: string, nextStatus: SessionStatus) => void;
  adjustRemainingTime: (id: string, deltaSeconds: number) => void;
  tickSession: (id: string) => void;
}

function slugify(value: string): string {
  return value
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/(^-|-$)/g, "")
    .slice(0, 40);
}

function makeSessionId(title: string): string {
  const prefix = slugify(title) || "session";
  return `${prefix}-${Math.random().toString(36).slice(2, 8)}`;
}

export const useAdminSessionStore = create<AdminSessionState>((set, get) => ({
  sessions: {},
  createSession: (draft) => {
    const id = makeSessionId(draft.title);
    const record: SessionRecord = {
      id,
      title: draft.title,
      speakerName: draft.speakerName,
      durationSeconds: draft.durationSeconds,
      remainingSeconds: draft.durationSeconds,
      status: "CREATED",
      createdAtIso: new Date().toISOString(),
    };

    set((state) => ({
      sessions: {
        ...state.sessions,
        [record.id]: record,
      },
    }));

    return record;
  },
  getSession: (id) => get().sessions[id],
  updateStatus: (id, nextStatus) =>
    set((state) => {
      const current = state.sessions[id];
      if (!current) {
        return state;
      }

      const allowed =
        (current.status === "CREATED" && nextStatus === "LIVE") ||
        (current.status === "LIVE" &&
          (nextStatus === "PAUSED" || nextStatus === "ENDED")) ||
        (current.status === "PAUSED" &&
          (nextStatus === "LIVE" || nextStatus === "ENDED"));

      if (!allowed) {
        return state;
      }

      return {
        sessions: {
          ...state.sessions,
          [id]: {
            ...current,
            status: nextStatus,
          },
        },
      };
    }),
  adjustRemainingTime: (id, deltaSeconds) =>
    set((state) => {
      const current = state.sessions[id];
      if (!current) {
        return state;
      }

      return {
        sessions: {
          ...state.sessions,
          [id]: {
            ...current,
            remainingSeconds: current.remainingSeconds + deltaSeconds,
          },
        },
      };
    }),
  tickSession: (id) =>
    set((state) => {
      const current = state.sessions[id];
      if (!current || current.status !== "LIVE") {
        return state;
      }

      return {
        sessions: {
          ...state.sessions,
          [id]: {
            ...current,
            remainingSeconds: current.remainingSeconds - 1,
          },
        },
      };
    }),
}));
