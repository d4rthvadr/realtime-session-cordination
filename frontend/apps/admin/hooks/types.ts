export type SessionStatus = "CREATED" | "LIVE" | "PAUSED" | "ENDED";

export interface SessionSnapshot {
  id: string;
  title: string;
  speakerName: string;
  durationSeconds: number;
  status: SessionStatus;
  remainingSeconds: number;
  createdAt?: string;
}
