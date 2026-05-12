"use server";

const ADMIN_BACKEND_URL =
  process.env.NEXT_PUBLIC_ADMIN_BACKEND_URL || "http://localhost:8080";

export interface SessionSnapshot {
  id: string;
  title: string;
  speakerName: string;
  durationSeconds: number;
  status: string;
  createdAt: string;
  remainingSeconds: number;
  startedAt?: string;
  pausedAt?: string;
  totalPausedDurationSeconds: number;
}

export interface CreateSessionInput {
  name: string;
  duration: number;
}

export interface AdjustTimeInput {
  deltaSeconds: number;
}

// GET /api/v1/sessions - List all sessions
export async function getSessionsList() {
  try {
    const response = await fetch(`${ADMIN_BACKEND_URL}/api/v1/sessions`, {
      method: "GET",
      headers: { "Content-Type": "application/json" },
      cache: "no-store",
    });

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
    return {
      session: data.session as SessionSnapshot,
      error: null,
      message: "Session fetched successfully",
    };
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "Failed to fetch session";
    return { session: null, error: message, message };
  }
}

// POST /api/v1/sessions - Create new session
export async function createSession(input: CreateSessionInput) {
  try {
    // Map to backend expected field names
    const payload = {
      title: input.name,
      speakerName: "Admin",
      durationSeconds: input.duration,
    };

    const response = await fetch(`${ADMIN_BACKEND_URL}/api/v1/sessions`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });

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
    const response = await fetch(
      `${ADMIN_BACKEND_URL}/api/v1/sessions/${sessionId}/start`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-Control-Token": controlToken,
        },
      },
    );

    if (!response.ok) {
      throw new Error(`Failed to start session: ${response.statusText}`);
    }

    const data = await response.json();
    return { session: data.session as SessionSnapshot, error: null };
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "Failed to start session";
    return { session: null, error: message };
  }
}

// POST /api/v1/sessions/:id/pause - Pause session
export async function pauseSession(sessionId: string, controlToken: string) {
  try {
    const response = await fetch(
      `${ADMIN_BACKEND_URL}/api/v1/sessions/${sessionId}/pause`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-Control-Token": controlToken,
        },
      },
    );

    if (!response.ok) {
      throw new Error(`Failed to pause session: ${response.statusText}`);
    }

    const data = await response.json();
    return { session: data.session as SessionSnapshot, error: null };
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "Failed to pause session";
    return { session: null, error: message };
  }
}

// POST /api/v1/sessions/:id/resume - Resume session
export async function resumeSession(sessionId: string, controlToken: string) {
  try {
    const response = await fetch(
      `${ADMIN_BACKEND_URL}/api/v1/sessions/${sessionId}/resume`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-Control-Token": controlToken,
        },
      },
    );

    if (!response.ok) {
      throw new Error(`Failed to resume session: ${response.statusText}`);
    }

    const data = await response.json();
    return { session: data.session as SessionSnapshot, error: null };
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "Failed to resume session";
    return { session: null, error: message };
  }
}

// POST /api/v1/sessions/:id/end - End session
export async function endSession(sessionId: string, controlToken: string) {
  try {
    const response = await fetch(
      `${ADMIN_BACKEND_URL}/api/v1/sessions/${sessionId}/end`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-Control-Token": controlToken,
        },
      },
    );

    if (!response.ok) {
      throw new Error(`Failed to end session: ${response.statusText}`);
    }

    const data = await response.json();
    return { session: data.session as SessionSnapshot, error: null };
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "Failed to end session";
    return { session: null, error: message };
  }
}

// POST /api/v1/sessions/:id/adjust-time - Adjust session time
export async function adjustSessionTime(
  sessionId: string,
  input: AdjustTimeInput,
  controlToken: string,
) {
  try {
    const response = await fetch(
      `${ADMIN_BACKEND_URL}/api/v1/sessions/${sessionId}/adjust-time`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-Control-Token": controlToken,
        },
        body: JSON.stringify(input),
      },
    );

    if (!response.ok) {
      throw new Error(`Failed to adjust time: ${response.statusText}`);
    }

    const data = await response.json();
    return { session: data.session as SessionSnapshot, error: null };
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "Failed to adjust time";
    return { session: null, error: message };
  }
}
