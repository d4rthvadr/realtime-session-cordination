"use server";

import { cookies } from "next/headers";

const ADMIN_BACKEND_URL =
  process.env.NEXT_PUBLIC_ADMIN_BACKEND_URL || "http://localhost:8080";

const ADMIN_AUTH_COOKIE_NAME = "admin_auth_token";

function getAdminAuthToken(): string | null {
  return cookies().get(ADMIN_AUTH_COOKIE_NAME)?.value ?? null;
}

function getProtectedRequestHeaders(): HeadersInit | null {
  const token = getAdminAuthToken();
  if (!token) {
    return null;
  }

  return {
    "Content-Type": "application/json",
    Authorization: `Bearer ${token}`,
  };
}

function unauthorizedResult<T>(fallbackValue: T) {
  return {
    ...fallbackValue,
    error: "Unauthorized. Please sign in again.",
  };
}

export interface SessionSnapshot {
  id: string;
  title: string;
  speakerName: string;
  durationSeconds: number;
  status: string;
  createdAt: string;
}

export type ProgramItemStatus =
  | "scheduled"
  | "in_progress"
  | "paused"
  | "ended"
  | "canceled";

export interface CreateSessionInput {
  name: string;
  duration: number;
}

export interface AdjustTimeInput {
  deltaSeconds: number;
}

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

export interface ProgramItemCreateInput {
  title: string;
  type: string;
  hostName?: string;
  scheduledStart: string;
  scheduledEnd: string;
  expectedDurationMinutes?: number;
  position: number;
  location?: string;
  metadata?: Record<string, unknown>;
}

export interface ProgramItemUpdateInput {
  title?: string;
  type?: string;
  status?: "scheduled" | "in_progress" | "paused" | "ended" | "canceled";
  hostName?: string;
  scheduledStart?: string;
  scheduledEnd?: string;
  expectedDurationMinutes?: number;
  position?: number;
  location?: string;
  metadata?: Record<string, unknown>;
}

export interface RuntimeSnapshot {
  type?: string;
  session: SessionSnapshot;
  programItem: ProgramItemSnapshot | null;
  nextProgramItem: ProgramItemSnapshot | null;
  deltaSeconds?: number;
}

export interface SessionLogSnapshot {
  id: string;
  sessionId: string;
  programItemId?: string;
  eventType: string;
  message: string;
  metadata?: Record<string, unknown>;
  occurredAt: string;
  requestId?: string;
  createdAt: string;
}

export interface SessionLogListInput {
  limit?: number;
  offset?: number;
  eventType?: string;
  entityType?: "session" | "program_item" | "cascade";
}

export interface SessionAnalyticsSummary {
  sessionId: string;
  sessionStatus: string;
  sessionDurationSeconds: number;
  programItemCount: number;
  scheduledCount: number;
  inProgressCount: number;
  pausedCount: number;
  endedCount: number;
  canceledCount: number;
  plannedSeconds: number;
  effectiveBudgetSeconds: number;
  totalAdjustmentSeconds: number;
  totalPauseSeconds: number;
  totalPauseCount: number;
  endedOnTimeCount: number;
  overrunItemCount: number;
  totalOverrunSeconds: number;
  totalUnderrunSeconds: number;
  endedOnTimeRatio: number;
  computedAt: string;
}

export interface AnalyticsOverview {
  totalSessions: number;
  createdSessions: number;
  liveSessions: number;
  pausedSessions: number;
  endedSessions: number;
  totalProgramItems: number;
  endedProgramItems: number;
  onTimeEndedProgramItems: number;
  overrunProgramItems: number;
  totalSessionDurationSeconds: number;
  totalPlannedSeconds: number;
  effectiveBudgetSeconds: number;
  totalAdjustmentSeconds: number;
  totalPauseSeconds: number;
  totalPauseCount: number;
  totalOverrunSeconds: number;
  totalUnderrunSeconds: number;
  sessionCompletionRatio: number;
  programItemOnTimeRatio: number;
  computedAt: string;
}

function runtimeResult(runtime: RuntimeSnapshot | null, error: string | null) {
  return {
    runtime,
    session: runtime?.session ?? null,
    error,
  };
}

function normalizeRuntimePayload(data: any): RuntimeSnapshot {
  const programItem = (data?.programItem ?? null) as ProgramItemSnapshot | null;
  const nextProgramItem = (data?.nextProgramItem ??
    null) as ProgramItemSnapshot | null;

  return {
    type: data?.type,
    session: data?.session as SessionSnapshot,
    programItem,
    nextProgramItem,
    deltaSeconds:
      typeof data?.deltaSeconds === "number" ? data.deltaSeconds : undefined,
  };
}

export interface ProgramItemReorderInput {
  items: Array<{ id: string; position: number }>;
}

// GET /api/v1/sessions - List all sessions
export async function getSessionsList() {
  try {
    const headers = getProtectedRequestHeaders();
    if (!headers) {
      return unauthorizedResult({ sessions: [] as SessionSnapshot[] });
    }

    const response = await fetch(`${ADMIN_BACKEND_URL}/api/v1/sessions`, {
      method: "GET",
      headers,
      cache: "no-store",
    });

    if (response.status === 401) {
      return unauthorizedResult({ sessions: [] as SessionSnapshot[] });
    }

    if (!response.ok) {
      throw new Error(`Failed to fetch sessions: ${response.statusText}`);
    }

    const data = await response.json();
    return { sessions: data.sessions || [], error: null };
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "Failed to fetch sessions";
    return { sessions: [], error: message };
  }
}

// GET /api/v1/sessions/:id/analytics - Get per-session analytics summary
export async function getSessionAnalytics(sessionId: string) {
  try {
    const headers = getProtectedRequestHeaders();
    if (!headers) {
      return unauthorizedResult({
        analytics: null as SessionAnalyticsSummary | null,
      });
    }

    const response = await fetch(
      `${ADMIN_BACKEND_URL}/api/v1/sessions/${sessionId}/analytics`,
      {
        method: "GET",
        headers,
        cache: "no-store",
      },
    );

    if (response.status === 401) {
      return unauthorizedResult({
        analytics: null as SessionAnalyticsSummary | null,
      });
    }

    if (!response.ok) {
      throw new Error(
        `Failed to fetch session analytics: ${response.statusText}`,
      );
    }

    const data = await response.json();
    return {
      analytics: data.analytics as SessionAnalyticsSummary,
      error: null,
    };
  } catch (error) {
    const message =
      error instanceof Error
        ? error.message
        : "Failed to fetch session analytics";
    return {
      analytics: null as SessionAnalyticsSummary | null,
      error: message,
    };
  }
}

// GET /api/v1/analytics/overview - Get cross-session analytics overview
export async function getAnalyticsOverview() {
  try {
    const headers = getProtectedRequestHeaders();
    if (!headers) {
      return unauthorizedResult({
        overview: null as AnalyticsOverview | null,
      });
    }

    const response = await fetch(
      `${ADMIN_BACKEND_URL}/api/v1/analytics/overview`,
      {
        method: "GET",
        headers,
        cache: "no-store",
      },
    );

    if (response.status === 401) {
      return unauthorizedResult({
        overview: null as AnalyticsOverview | null,
      });
    }

    if (!response.ok) {
      throw new Error(
        `Failed to fetch analytics overview: ${response.statusText}`,
      );
    }

    const data = await response.json();
    return {
      overview: data.overview as AnalyticsOverview,
      error: null,
    };
  } catch (error) {
    const message =
      error instanceof Error
        ? error.message
        : "Failed to fetch analytics overview";
    return { overview: null as AnalyticsOverview | null, error: message };
  }
}

// GET /api/v1/sessions/:id - Get single session
export async function getSessionSnapshot(sessionId: string) {
  try {
    const response = await fetch(
      `${ADMIN_BACKEND_URL}/api/v1/sessions/${sessionId}`,
      {
        method: "GET",
        headers: { "Content-Type": "application/json" },
        cache: "no-store",
      },
    );

    if (!response.ok) {
      throw new Error(`Failed to fetch session: ${response.statusText}`);
    }

    const data = await response.json();
    return runtimeResult(normalizeRuntimePayload(data), null);
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "Failed to fetch session";
    return runtimeResult(null, message);
  }
}

// POST /api/v1/sessions - Create new session
export async function createSession(input: CreateSessionInput) {
  try {
    const headers = getProtectedRequestHeaders();
    if (!headers) {
      return unauthorizedResult({
        session: null as (SessionSnapshot & { controlToken: string }) | null,
      });
    }

    // Map to backend expected field names
    const payload = {
      title: input.name,
      speakerName: "Admin",
      durationSeconds: input.duration,
    };

    const response = await fetch(`${ADMIN_BACKEND_URL}/api/v1/sessions`, {
      method: "POST",
      headers,
      body: JSON.stringify(payload),
    });

    if (response.status === 401) {
      return unauthorizedResult({
        session: null as (SessionSnapshot & { controlToken: string }) | null,
      });
    }

    if (!response.ok) {
      throw new Error(`Failed to create session: ${response.statusText}`);
    }

    const data = await response.json();
    // Backend returns { session: SessionSnapshot, controlToken: string }
    return {
      session: {
        ...data.session,
        controlToken: data.controlToken,
      } as SessionSnapshot & { controlToken: string },
      error: null,
    };
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "Failed to create session";
    return { session: null, error: message };
  }
}

// POST /api/v1/sessions/:id/start - Start session
export async function startSession(sessionId: string, controlToken: string) {
  try {
    const headers = getProtectedRequestHeaders();
    if (!headers) {
      return unauthorizedResult({
        runtime: null as RuntimeSnapshot | null,
        session: null as SessionSnapshot | null,
      });
    }

    const response = await fetch(
      `${ADMIN_BACKEND_URL}/api/v1/sessions/${sessionId}/start`,
      {
        method: "POST",
        headers: {
          ...headers,
          "X-Control-Token": controlToken,
        },
      },
    );

    if (response.status === 401) {
      return unauthorizedResult({
        runtime: null as RuntimeSnapshot | null,
        session: null as SessionSnapshot | null,
      });
    }

    if (!response.ok) {
      throw new Error(`Failed to start session: ${response.statusText}`);
    }

    const data = await response.json();
    return runtimeResult(normalizeRuntimePayload(data), null);
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "Failed to start session";
    return runtimeResult(null, message);
  }
}

// POST /api/v1/sessions/:id/pause - Pause session
export async function pauseSession(sessionId: string, controlToken: string) {
  try {
    const headers = getProtectedRequestHeaders();
    if (!headers) {
      return unauthorizedResult({
        runtime: null as RuntimeSnapshot | null,
        session: null as SessionSnapshot | null,
      });
    }

    const response = await fetch(
      `${ADMIN_BACKEND_URL}/api/v1/sessions/${sessionId}/pause`,
      {
        method: "POST",
        headers: {
          ...headers,
          "X-Control-Token": controlToken,
        },
      },
    );

    if (response.status === 401) {
      return unauthorizedResult({
        runtime: null as RuntimeSnapshot | null,
        session: null as SessionSnapshot | null,
      });
    }

    if (!response.ok) {
      throw new Error(`Failed to pause session: ${response.statusText}`);
    }

    const data = await response.json();
    return runtimeResult(normalizeRuntimePayload(data), null);
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "Failed to pause session";
    return runtimeResult(null, message);
  }
}

// POST /api/v1/sessions/:id/resume - Resume session
export async function resumeSession(sessionId: string, controlToken: string) {
  try {
    const headers = getProtectedRequestHeaders();
    if (!headers) {
      return unauthorizedResult({
        runtime: null as RuntimeSnapshot | null,
        session: null as SessionSnapshot | null,
      });
    }

    const response = await fetch(
      `${ADMIN_BACKEND_URL}/api/v1/sessions/${sessionId}/resume`,
      {
        method: "POST",
        headers: {
          ...headers,
          "X-Control-Token": controlToken,
        },
      },
    );

    if (response.status === 401) {
      return unauthorizedResult({
        runtime: null as RuntimeSnapshot | null,
        session: null as SessionSnapshot | null,
      });
    }

    if (!response.ok) {
      throw new Error(`Failed to resume session: ${response.statusText}`);
    }

    const data = await response.json();
    return runtimeResult(normalizeRuntimePayload(data), null);
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "Failed to resume session";
    return runtimeResult(null, message);
  }
}

// POST /api/v1/sessions/:id/end - End session
export async function endSession(sessionId: string, controlToken: string) {
  try {
    const headers = getProtectedRequestHeaders();
    if (!headers) {
      return unauthorizedResult({
        runtime: null as RuntimeSnapshot | null,
        session: null as SessionSnapshot | null,
      });
    }

    const response = await fetch(
      `${ADMIN_BACKEND_URL}/api/v1/sessions/${sessionId}/end`,
      {
        method: "POST",
        headers: {
          ...headers,
          "X-Control-Token": controlToken,
        },
      },
    );

    if (response.status === 401) {
      return unauthorizedResult({
        runtime: null as RuntimeSnapshot | null,
        session: null as SessionSnapshot | null,
      });
    }

    if (!response.ok) {
      throw new Error(`Failed to end session: ${response.statusText}`);
    }

    const data = await response.json();
    return runtimeResult(normalizeRuntimePayload(data), null);
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "Failed to end session";
    return runtimeResult(null, message);
  }
}

// POST /api/v1/sessions/:id/adjust-time - Adjust session time
export async function adjustSessionTime(
  sessionId: string,
  input: AdjustTimeInput,
  controlToken: string,
) {
  try {
    const headers = getProtectedRequestHeaders();
    if (!headers) {
      return unauthorizedResult({
        runtime: null as RuntimeSnapshot | null,
        session: null as SessionSnapshot | null,
      });
    }

    const response = await fetch(
      `${ADMIN_BACKEND_URL}/api/v1/sessions/${sessionId}/adjust-time`,
      {
        method: "POST",
        headers: {
          ...headers,
          "X-Control-Token": controlToken,
        },
        body: JSON.stringify(input),
      },
    );

    if (response.status === 401) {
      return unauthorizedResult({
        runtime: null as RuntimeSnapshot | null,
        session: null as SessionSnapshot | null,
      });
    }

    if (!response.ok) {
      throw new Error(`Failed to adjust time: ${response.statusText}`);
    }

    const data = await response.json();
    return runtimeResult(normalizeRuntimePayload(data), null);
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "Failed to adjust time";
    return runtimeResult(null, message);
  }
}

// GET /api/v1/sessions/:id/program-items - List program items
export async function getProgramItems(sessionId: string) {
  try {
    const headers = getProtectedRequestHeaders();
    if (!headers) {
      return unauthorizedResult({ programItems: [] as ProgramItemSnapshot[] });
    }

    const response = await fetch(
      `${ADMIN_BACKEND_URL}/api/v1/sessions/${sessionId}/program-items`,
      {
        method: "GET",
        headers,
        cache: "no-store",
      },
    );

    if (response.status === 401) {
      return unauthorizedResult({ programItems: [] as ProgramItemSnapshot[] });
    }

    if (!response.ok) {
      throw new Error(`Failed to fetch program items: ${response.statusText}`);
    }

    const data = await response.json();
    return {
      programItems: (data.programItems || []) as ProgramItemSnapshot[],
      error: null,
    };
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "Failed to fetch program items";
    return { programItems: [] as ProgramItemSnapshot[], error: message };
  }
}

// GET /api/v1/sessions/:id/logs - List session logs
export async function getSessionLogs(
  sessionId: string,
  input: SessionLogListInput = {},
) {
  try {
    const headers = getProtectedRequestHeaders();
    if (!headers) {
      return unauthorizedResult({ logs: [] as SessionLogSnapshot[] });
    }

    const params = new URLSearchParams();
    if (typeof input.limit === "number") {
      params.set("limit", String(input.limit));
    }
    if (typeof input.offset === "number") {
      params.set("offset", String(input.offset));
    }
    if (input.eventType) {
      params.set("eventType", input.eventType);
    }
    if (input.entityType) {
      params.set("entityType", input.entityType);
    }

    const query = params.toString();
    const response = await fetch(
      `${ADMIN_BACKEND_URL}/api/v1/sessions/${sessionId}/logs${query ? `?${query}` : ""}`,
      {
        method: "GET",
        headers,
        cache: "no-store",
      },
    );

    if (response.status === 401) {
      return unauthorizedResult({ logs: [] as SessionLogSnapshot[] });
    }

    if (!response.ok) {
      throw new Error(`Failed to fetch session logs: ${response.statusText}`);
    }

    const data = await response.json();
    return {
      logs: (data.logs || []) as SessionLogSnapshot[],
      error: null,
    };
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "Failed to fetch session logs";
    return { logs: [] as SessionLogSnapshot[], error: message };
  }
}

// POST /api/v1/sessions/:id/program-items - Create program item
export async function createProgramItem(
  sessionId: string,
  input: ProgramItemCreateInput,
  controlToken: string,
) {
  try {
    const headers = getProtectedRequestHeaders();
    if (!headers) {
      return unauthorizedResult({
        programItem: null as ProgramItemSnapshot | null,
      });
    }

    const response = await fetch(
      `${ADMIN_BACKEND_URL}/api/v1/sessions/${sessionId}/program-items`,
      {
        method: "POST",
        headers: {
          ...headers,
          "X-Control-Token": controlToken,
        },
        body: JSON.stringify(input),
      },
    );

    if (response.status === 401) {
      return unauthorizedResult({
        programItem: null as ProgramItemSnapshot | null,
      });
    }

    if (!response.ok) {
      throw new Error(`Failed to create program item: ${response.statusText}`);
    }

    const data = await response.json();
    return {
      programItem: data.programItem as ProgramItemSnapshot,
      error: null,
    };
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "Failed to create program item";
    return { programItem: null as ProgramItemSnapshot | null, error: message };
  }
}

// PATCH /api/v1/program-items/:itemId - Update program item
export async function updateProgramItem(
  itemId: string,
  input: ProgramItemUpdateInput,
  controlToken: string,
) {
  try {
    const headers = getProtectedRequestHeaders();
    if (!headers) {
      return unauthorizedResult({
        programItem: null as ProgramItemSnapshot | null,
      });
    }

    const response = await fetch(
      `${ADMIN_BACKEND_URL}/api/v1/program-items/${itemId}`,
      {
        method: "PATCH",
        headers: {
          ...headers,
          "X-Control-Token": controlToken,
        },
        body: JSON.stringify(input),
      },
    );

    if (response.status === 401) {
      return unauthorizedResult({
        programItem: null as ProgramItemSnapshot | null,
      });
    }

    if (!response.ok) {
      throw new Error(`Failed to update program item: ${response.statusText}`);
    }

    const data = await response.json();
    return {
      programItem: data.programItem as ProgramItemSnapshot,
      error: null,
    };
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "Failed to update program item";
    return { programItem: null as ProgramItemSnapshot | null, error: message };
  }
}

// POST /api/v1/program-items/:itemId/cancel - Cancel program item
export async function cancelProgramItem(itemId: string, controlToken: string) {
  try {
    const headers = getProtectedRequestHeaders();
    if (!headers) {
      return unauthorizedResult({
        programItem: null as ProgramItemSnapshot | null,
      });
    }

    const response = await fetch(
      `${ADMIN_BACKEND_URL}/api/v1/program-items/${itemId}/cancel`,
      {
        method: "POST",
        headers: {
          ...headers,
          "X-Control-Token": controlToken,
        },
      },
    );

    if (response.status === 401) {
      return unauthorizedResult({
        programItem: null as ProgramItemSnapshot | null,
      });
    }

    if (!response.ok) {
      throw new Error(`Failed to cancel program item: ${response.statusText}`);
    }

    const data = await response.json();
    return {
      programItem: data.programItem as ProgramItemSnapshot,
      error: null,
    };
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "Failed to cancel program item";
    return { programItem: null as ProgramItemSnapshot | null, error: message };
  }
}

// POST /api/v1/program-items/:itemId/start - Start program item
export async function startProgramItem(itemId: string, controlToken: string) {
  try {
    const headers = getProtectedRequestHeaders();
    if (!headers) {
      return unauthorizedResult({
        runtime: null as RuntimeSnapshot | null,
        session: null as SessionSnapshot | null,
      });
    }

    const response = await fetch(
      `${ADMIN_BACKEND_URL}/api/v1/program-items/${itemId}/start`,
      {
        method: "POST",
        headers: {
          ...headers,
          "X-Control-Token": controlToken,
        },
      },
    );

    if (response.status === 401) {
      return unauthorizedResult({
        runtime: null as RuntimeSnapshot | null,
        session: null as SessionSnapshot | null,
      });
    }

    if (!response.ok) {
      throw new Error(`Failed to start program item: ${response.statusText}`);
    }

    const data = await response.json();
    return runtimeResult(normalizeRuntimePayload(data), null);
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "Failed to start program item";
    return runtimeResult(null, message);
  }
}

// POST /api/v1/program-items/:itemId/pause - Pause program item
export async function pauseProgramItem(itemId: string, controlToken: string) {
  try {
    const headers = getProtectedRequestHeaders();
    if (!headers) {
      return unauthorizedResult({
        runtime: null as RuntimeSnapshot | null,
        session: null as SessionSnapshot | null,
      });
    }

    const response = await fetch(
      `${ADMIN_BACKEND_URL}/api/v1/program-items/${itemId}/pause`,
      {
        method: "POST",
        headers: {
          ...headers,
          "X-Control-Token": controlToken,
        },
      },
    );

    if (response.status === 401) {
      return unauthorizedResult({
        runtime: null as RuntimeSnapshot | null,
        session: null as SessionSnapshot | null,
      });
    }

    if (!response.ok) {
      throw new Error(`Failed to pause program item: ${response.statusText}`);
    }

    const data = await response.json();
    return runtimeResult(normalizeRuntimePayload(data), null);
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "Failed to pause program item";
    return runtimeResult(null, message);
  }
}

// POST /api/v1/program-items/:itemId/resume - Resume program item
export async function resumeProgramItem(itemId: string, controlToken: string) {
  try {
    const headers = getProtectedRequestHeaders();
    if (!headers) {
      return unauthorizedResult({
        runtime: null as RuntimeSnapshot | null,
        session: null as SessionSnapshot | null,
      });
    }

    const response = await fetch(
      `${ADMIN_BACKEND_URL}/api/v1/program-items/${itemId}/resume`,
      {
        method: "POST",
        headers: {
          ...headers,
          "X-Control-Token": controlToken,
        },
      },
    );

    if (response.status === 401) {
      return unauthorizedResult({
        runtime: null as RuntimeSnapshot | null,
        session: null as SessionSnapshot | null,
      });
    }

    if (!response.ok) {
      throw new Error(`Failed to resume program item: ${response.statusText}`);
    }

    const data = await response.json();
    return runtimeResult(normalizeRuntimePayload(data), null);
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "Failed to resume program item";
    return runtimeResult(null, message);
  }
}

// POST /api/v1/program-items/:itemId/adjust-time - Adjust program item time
export async function adjustProgramItemTime(
  itemId: string,
  input: AdjustTimeInput,
  controlToken: string,
) {
  try {
    const headers = getProtectedRequestHeaders();
    if (!headers) {
      return unauthorizedResult({
        runtime: null as RuntimeSnapshot | null,
        session: null as SessionSnapshot | null,
      });
    }

    const response = await fetch(
      `${ADMIN_BACKEND_URL}/api/v1/program-items/${itemId}/adjust-time`,
      {
        method: "POST",
        headers: {
          ...headers,
          "X-Control-Token": controlToken,
        },
        body: JSON.stringify(input),
      },
    );

    if (response.status === 401) {
      return unauthorizedResult({
        runtime: null as RuntimeSnapshot | null,
        session: null as SessionSnapshot | null,
      });
    }

    if (!response.ok) {
      throw new Error(
        `Failed to adjust program item time: ${response.statusText}`,
      );
    }

    const data = await response.json();
    return runtimeResult(normalizeRuntimePayload(data), null);
  } catch (error) {
    const message =
      error instanceof Error
        ? error.message
        : "Failed to adjust program item time";
    return runtimeResult(null, message);
  }
}

// POST /api/v1/program-items/:itemId/end - End program item
export async function endProgramItem(itemId: string, controlToken: string) {
  try {
    const headers = getProtectedRequestHeaders();
    if (!headers) {
      return unauthorizedResult({
        runtime: null as RuntimeSnapshot | null,
        session: null as SessionSnapshot | null,
      });
    }

    const response = await fetch(
      `${ADMIN_BACKEND_URL}/api/v1/program-items/${itemId}/end`,
      {
        method: "POST",
        headers: {
          ...headers,
          "X-Control-Token": controlToken,
        },
      },
    );

    if (response.status === 401) {
      return unauthorizedResult({
        runtime: null as RuntimeSnapshot | null,
        session: null as SessionSnapshot | null,
      });
    }

    if (!response.ok) {
      throw new Error(`Failed to end program item: ${response.statusText}`);
    }

    const data = await response.json();
    return runtimeResult(normalizeRuntimePayload(data), null);
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "Failed to end program item";
    return runtimeResult(null, message);
  }
}

// POST /api/v1/sessions/:id/program-items/reorder - Reorder program items
export async function reorderProgramItems(
  sessionId: string,
  input: ProgramItemReorderInput,
  controlToken: string,
) {
  try {
    const headers = getProtectedRequestHeaders();
    if (!headers) {
      return unauthorizedResult({ programItems: [] as ProgramItemSnapshot[] });
    }

    const response = await fetch(
      `${ADMIN_BACKEND_URL}/api/v1/sessions/${sessionId}/program-items/reorder`,
      {
        method: "POST",
        headers: {
          ...headers,
          "X-Control-Token": controlToken,
        },
        body: JSON.stringify(input),
      },
    );

    if (response.status === 401) {
      return unauthorizedResult({ programItems: [] as ProgramItemSnapshot[] });
    }

    if (!response.ok) {
      throw new Error(
        `Failed to reorder program items: ${response.statusText}`,
      );
    }

    const data = await response.json();
    return {
      programItems: (data.programItems || []) as ProgramItemSnapshot[],
      error: null,
    };
  } catch (error) {
    const message =
      error instanceof Error
        ? error.message
        : "Failed to reorder program items";
    return { programItems: [] as ProgramItemSnapshot[], error: message };
  }
}
