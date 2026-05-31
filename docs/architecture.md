# Architecture & Design

## Runtime Authority Update (2026-05)

Runtime countdown authority is now split by intent:

- Session remains the container lifecycle and control-token authority.
- Active ProgramItem is the runtime countdown authority when present.
- Runtime actions return a unified envelope with session, current programItem, and nextProgramItem.

This removes countdown drift between session controls and agenda controls.

Detailed countdown math by ProgramItem status is documented in docs/programitem-time-calculation.md.

## Session Entity Cleanup Status

Session cleanup follow-up has been applied for runtime mutation paths.

Current state:

- Session countdown fields remain transitional compatibility mirrors.
- Adjust-time mutation coupling no longer writes to both Session and active ProgramItem.
- Session adjust-time now prioritizes active ProgramItem runtime when present, with Session-only fallback when no active runtime item exists.

## System Overview

The Realtime Session Coordination Platform is a three-tier system:

1. **Frontend (User App)** — Public countdown viewer with real-time synchronization
2. **Frontend (Admin App)** — Host control panel for session management
3. **Backend** — Authoritative session manager with REST API and WebSocket broadcast

```
┌─────────────────────────────────────────────────────────────────┐
│                    PUBLIC INTERNET (VIEWERS)                    │
└─────────────────────────────────────────────────────────────────┘
                               ↓
                    ┌──────────┴──────────┐
                    │                     │
            ┌───────▼────────┐    ┌──────▼──────────┐
            │  User App      │    │  Browser N...   │
            │  (localhost:   │    │  (Viewer)       │
            │   3001)        │    └─────────────────┘
            │  React + Next  │
            │  Viewer Viewer └──────┐
            └───────┬────────┘      │
                    │               │ GET /api/v1/sessions/:id
                    └───────────────┼────────────────────────┐
                                    │                        │
                    ┌───────────────┼────────────────────────┼─┐
                    │               │                        │ │
            ┌───────▼────────┐      │     ┌─────────▼──────────▼─┐
            │  Admin App     │      │     │                       │
            │  (localhost:   │      │     │  Backend Service      │
            │   3002)        │      │     │  (localhost: 8080)    │
            │  React + Next  │      │     │                       │
            │  Host Panel    │      │     │  ┌─────────────────┐  │
            │                │      │     │  │ Session Manager │  │
            │ POST create    │      │     │  │ (Authoritative  │  │
            │ POST start/    │──────┼────►│  │  Timer)         │  │
            │  pause/resume/ │      │     │  └─────────────────┘  │
            │  end/adjust    │      │     │                       │
            └────────────────┘      │     │  ┌─────────────────┐  │
                                    │     │  │ WebSocket Hub   │  │
                    ┌───────────────┼─────┼──┤ (Broadcast)     │  │
                    │ WebSocket     │     │  └─────────────────┘  │
                    │ connect       │     │                       │
                    └───────────────┤     │  REST Handlers + WS   │
                                    └────►│  Upgrade              │
                                          └─────────────────────┘
```

---

## Frontend Architecture

### User App (`frontend/apps/user`)

**Purpose:** Public-facing countdown viewer that displays real-time session state.

**Key Components:**

- `CountdownBoard.tsx` — Main countdown display with urgency states
- `useSessionSocket.ts` — WebSocket hook managing connection and state sync
- `sessionStore.ts` — Zustand store for session snapshot and connection state
- `time.ts` — Utility functions for duration formatting and timer state calculation
- `backend.ts` — API client builders

**Data Flow:**

1. Page mounts → loads session via `GET /api/v1/sessions/:id`
2. Store initializes from unified runtime envelope (`session`, `programItem`, `nextProgramItem`)
3. WebSocket connects to `ws://localhost:8080/ws/sessions/:id`
4. Receives unified runtime snapshot on connect (same shape as REST)
5. On runtime updates, store applies session + program item snapshot atomically
6. UI renders countdown from active ProgramItem `remainingSeconds` when present, else `00:00`
   - Green (safe): > 25% of duration
   - Yellow (warning): 10-25% of duration
   - Red (critical): < 10% of duration
   - Red + blinking (overtime): ≤ 0 seconds

**State Shape:**

```typescript
interface RuntimeEnvelope {
  session: {
    id: string;
    title: string;
    speakerName: string;
    durationSeconds: number;
    status: "CREATED" | "LIVE" | "PAUSED" | "ENDED";
    remainingSeconds: number; // transitional compatibility mirror
  };
  programItem: ProgramItemSnapshot | null;
  nextProgramItem: ProgramItemSnapshot | null;
}

interface SessionState {
  runtime: RuntimeEnvelope | null;
  connectionState: "disconnected" | "connecting" | "connected";
  hasReceivedSnapshot: boolean;
  sessionNotFound: boolean;
}
```

**Synchronization Strategy:**

- **ProgramItem runtime authoritative:** Active ProgramItem owns countdown when present
- **Envelope-driven UI:** Client renders one payload shape across REST and WebSocket
- **Reconnect handling:** On WebSocket reconnect, loads fresh session state and re-establishes connection

---

### Admin App (`frontend/apps/admin`)

**Purpose:** Host/organizer control panel for session creation and lifecycle management.

**Key Components:**

- `SessionCreateForm.tsx` — Form to create new sessions
- `HostControlPanel.tsx` — Display session state and provide control buttons
- `adminSessionStore.ts` — Zustand store for admin-facing sessions
- `session.ts` — Utility functions for formatting and URL generation
- `backend.ts` — API client builders

**Data Flow:**

1. Host fills form → clicks "Create Session"
2. `POST /api/v1/sessions` → receives `{ session, controlToken, viewerPath }`
3. Stores `controlToken` in `sessionStorage` as `controlToken:sessionId`
4. Navigates to session detail page
5. Loads session via `GET /api/v1/sessions/:id` (no auth required for read)
6. Host clicks action button (start/pause/resume/end/adjust)
7. `POST /api/v1/sessions/:id/{action}` with `X-Control-Token` header
8. Updates local session state on success
9. Displays public viewer link (can be shared with audience)

**State Shape:**

```typescript
interface Session {
  id: string;
  title: string;
  speakerName: string;
  durationSeconds: number;
  status: "CREATED" | "LIVE" | "PAUSED" | "ENDED";
  remainingSeconds: number;
}

interface AdminSessionState {
  sessions: { [id: string]: Session };
  createLoading: boolean;
  createError: string | null;
  activeSessionId: string | null;
}
```

**Authorization Flow:**

- Control token generated on session creation (backend secret)
- Stored in browser `sessionStorage` with key `controlToken:sessionId`
- Sent in `X-Control-Token` header on host action requests
- If token lost/refreshed, admin loses ability to control session (by design)

---

## Backend Architecture

### Session Manager (`internal/session/session.go`)

**Responsibility:** Domain model and state machine for sessions.

**Core Operations:**

- `Create(input)` → creates new session, generates control token, returns snapshot
- `GetSnapshot(id)` → returns current session state
- `Start(id)` → CREATED → LIVE transition
- `Pause(id)` → LIVE → PAUSED transition
- `Resume(id)` → PAUSED → LIVE transition
- `End(id)` → any → ENDED transition
- `AdjustTime(id, deltaSeconds)` → modifies remaining time

**State Machine:**

```
         CREATE
           │
           ▼
        ┌──────┐
        │CREATED│◄─────────┐
        └───┬──┘           │
            │              │
            │ START        │
            ▼              │
        ┌──────┐           │
        │ LIVE │           │
        └──┬─┬─┘           │
        PAUSE PAUSE        │
           ▼   ▼           │
        ┌──────┐           │
        │PAUSED├───RESUME──┤
        └───┬──┘           │
            │              │
            └──────END─────┴─┐
                             │
                          ┌──▼──┐
                          │ENDED│
                          └─────┘
```

**Authoritative Timing:**

The server owns runtime calculations and returns current values in each runtime envelope.

- `programItem.remainingSeconds` is authoritative when an active item exists.
- `session.remainingSeconds` is a transitional compatibility mirror.
- Without an active ProgramItem, viewer countdown is `00:00`.

Detailed formula and status-based math are documented in docs/programitem-time-calculation.md.

---

### WebSocket Hub (`internal/ws/hub.go`)

**Responsibility:** Connection registry and message broadcast.

**Operations:**

- `Register(sessionID, connection)` — add connection to session channel
- `Unregister(sessionID, connection)` — remove connection
- `Broadcast(sessionID, message)` — send to all connected clients
- `SendSnapshot(sessionID, connection)` — send current state to single client

**Message Shape:**

- WebSocket connect and update payloads use the unified runtime envelope.
- Runtime action events include `type` and envelope fields (`session`, `programItem`, `nextProgramItem`).

**Broadcast Flow:**

1. Host calls control action (e.g., start, pause)
2. Backend handler updates session state
3. Handler broadcasts update to all WebSocket clients of that session
4. Clients receive update and refresh their UI

---

### API Handlers (`internal/api/handlers.go`)

**Responsibility:** HTTP request handling, authorization, routing.

**Endpoints:**

- `GET /healthz` — health check
- `POST /api/v1/sessions` — create session
- `GET /api/v1/sessions/:id` — get session state
- `POST /api/v1/sessions/:id/{action}` — control actions (start/pause/resume/end/adjust-time)
- `GET /api/v1/sessions/:id/program-items` — list timeline items
- `POST /api/v1/sessions/:id/program-items` — create timeline item
- `PATCH /api/v1/program-items/:itemId` — update timeline item
- `POST /api/v1/program-items/:itemId/cancel` — cancel timeline item
- `POST /api/v1/sessions/:id/program-items/reorder` — bulk reorder timeline items
- `GET /ws/sessions/:id` — WebSocket upgrade

**Authorization:**

Control endpoints require valid `X-Control-Token`:

1. Extract token from header or query parameter
2. Fetch session from manager
3. Compare token against session's control token
4. Return 401 if mismatch

Public session read endpoints (`GET /api/v1/sessions/:id`, `GET /ws/sessions/:id`) require no auth.

The user viewer also uses `GET /api/v1/sessions/:id/current-program-item` as a public read endpoint.

ProgramItem mutations require both bearer authorization and session control token.

---

## ProgramItem Scheduling Model (Runtime Contract)

ProgramItems are session-scoped timeline blocks with explicit position ordering.

ProgramItem statuses:

- `scheduled`
- `in_progress`
- `paused`
- `ended`
- `canceled`

Core rules:

1. `scheduledStart` must be before `scheduledEnd`.
2. Overlap is not allowed among scheduled items in the same session.
3. `position` is unique per session.
4. Reorder uses a bulk transaction to keep position integrity.

Overlap predicate for create/update:

```
existing.scheduled_start < candidate.scheduled_end
AND existing.scheduled_end > candidate.scheduled_start
```

Update operations exclude the current item id from overlap checks.

### Cancellation Behavior

ProgramItem deletion is represented as cancellation:

1. Status changes from `scheduled` to `canceled`.
2. Original time slot is preserved in API responses.
3. Subsequent items are not auto-shifted.
4. Viewer context may auto-advance to next non-canceled item.

This preserves timeline history for future metrics and audit trails.

### Current ProgramItem (Viewer Context)

Viewer context is server-derived and exposed as:

```json
{
  "programItem": { "...": "current item or null" },
  "nextProgramItem": { "...": "next item or null" }
}
```

Selection policy:

1. Prefer explicit `in_progress` item as current.
2. If none is in progress, derive current from scheduled window (`start <= now < end`) among `scheduled` items.
3. Derive next as first upcoming non-canceled scheduled item.

This keeps both admin and user interfaces synchronized even when runtime transitions are manually controlled.

Viewer context resolves the active timeline item from backend logic:

1. Include only `scheduled` items.
2. Select item where `scheduledStart <= now < scheduledEnd`.
3. Return `null` when no active item exists.

This logic is exposed by `GET /api/v1/sessions/:id/current-program-item` as a compatibility read endpoint.
Primary viewer synchronization should use `GET /api/v1/sessions/:id` and WebSocket runtime envelopes.

---

## Data Flow Diagrams

### Create Session Flow

```
Admin App                         Backend
  │                                │
  ├─ POST /api/v1/sessions ──────► │
  │                                 │ Create session
  │                                 │ Generate token
  │◄────── 201 + session + token ──│
  │                                │
  ├─ Store token in sessionStorage
  └─ Navigate to /sessions/{id}
```

### Control Action Flow

```
Admin App                   WebSocket    Backend              User App
  │                            │           │                    │
  ├─ POST /api/v1/sessions/:id/start (with token)              │
  │                            │           │                    │
  │                            │   Update session state         │
  │                            │   status = "LIVE"              │
  │◄─────── 200 OK ────────────│           │                    │
  │                            │           │                    │
  │                            ├─ Broadcast SESSION_UPDATE ────►│
  │                            │           │                    │
  │                            │           │   Update UI        │
  │                            │           │   Start countdown  │
  │                            │           │                    │
```

### Viewer Sync Flow

```
User App                              Backend
  │                                     │
  ├─ GET /api/v1/sessions/:id ────────►│
  │◄─ 200 + runtime envelope ──────────│
  │                                     │
  ├─ Upgrade to WebSocket──────────────►│
  │                                      │ Send runtime envelope
  │◄─ Runtime snapshot ──────────────────│
  │                                      │
  ├─ Store runtime envelope              │
  ├─ Start local display tick (1s)       │
  │                                      │
  │◄─ Runtime update envelope ──────────│
  │                                      │
  ├─ Replace runtime envelope (re-sync)  │
  ├─ Derive countdown from active item   │
  │                                      │
```

---

## Deployment Considerations

### Frontend

- **Build Target:** Next.js static export or server
- **Environment Variables:** `NEXT_PUBLIC_BACKEND_BASE_URL` must point to backend
- **CORS:** Requires backend to allow frontend origin

### Backend

- **Build Target:** Go binary
- **Port:** Configurable via environment (default 8080)
- **CORS:** Currently allows all origins (should restrict in production)
- **Session Persistence:** SQLite by default via configurable store abstraction (`DB_DRIVER=sqlite`)

### Network

- WebSocket and REST use same origin (simplifies CORS, firewall rules)
- No external services or databases required for MVP
- Can run entirely locally or on single server

---

## Extension Points

### Future Features

1. **Postgres Adapter** — Add Postgres store implementation using the existing store interface
2. **Authentication** — Add user accounts and session ownership
3. **Analytics** — Track session metrics and viewer engagement
4. **Recording** — Store countdown timeline and events
5. **Slide Sync** — Broadcast slide changes alongside countdown
6. **Captions** — Real-time caption delivery per session
7. **Audience Reactions** — Live polls, reactions, Q&A
8. **Mobile Apps** — Native iOS/Android clients
9. **Rate Limiting** — Protect backend from abuse
10. **Admin Dashboard** — Multi-session management UI

---

## Performance Notes

- **SQLite Session Store:** Durable single-instance persistence with low operational overhead
- **WebSocket Broadcasting:** Fan-out pattern scales linearly with connected clients per session
- **Server-Authoritative Timing:** Snapshot-based updates avoid client clock drift issues
- **Server Snapshot Frequency:** Broadcasting only on state changes, not on every tick

For high-scale deployments (10k+ concurrent sessions), consider:

- Distributed session store (Redis)
- Horizontal WebSocket scaling (message broker)
- Read replicas for public viewer endpoints
