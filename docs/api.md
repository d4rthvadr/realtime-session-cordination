# API Structure & Endpoints

## Runtime Contract Update (2026-05)

This section is the active contract for runtime countdown behavior and response shapes.
Older examples below remain useful historical context, but runtime control endpoints now return a unified envelope.

### Unified Runtime Envelope

Runtime endpoints now return:

```json
{
  "type": "SESSION_PAUSED",
  "session": {
    "id": "sess_abc123",
    "title": "Engineering Talks Q&A",
    "speakerName": "Alice Johnson",
    "durationSeconds": 600,
    "status": "PAUSED",
    "remainingSeconds": 312,
    "createdAt": "2026-05-31T10:00:00Z"
  },
  "programItem": {
    "id": "pi_abc123",
    "sessionId": "sess_abc123",
    "title": "Panel Discussion",
    "type": "panel",
    "status": "paused",
    "runtimeDurationSeconds": 1200,
    "actualStart": "2026-05-31T10:20:00Z",
    "pausedAt": "2026-05-31T10:28:08Z",
    "totalPausedDurationSeconds": 75,
    "adjustmentSeconds": 60,
    "endedRemainingSeconds": null,
    "actualEnd": null,
    "pauseCount": 1,
    "endedReason": "",
    "remainingSeconds": 717,
    "hostName": "Alice Johnson",
    "scheduledStart": "2026-05-31T10:20:00Z",
    "scheduledEnd": "2026-05-31T10:40:00Z",
    "expectedDurationMinutes": 20,
    "position": 4,
    "location": "Main Hall",
    "metadata": { "track": "engineering" },
    "createdAt": "2026-05-31T09:00:00Z",
    "updatedAt": "2026-05-31T10:28:08Z"
  },
  "nextProgramItem": {
    "id": "pi_def456",
    "sessionId": "sess_abc123",
    "title": "Q&A",
    "type": "q&a",
    "status": "scheduled",
    "runtimeDurationSeconds": 600,
    "totalPausedDurationSeconds": 0,
    "adjustmentSeconds": 0,
    "pauseCount": 0,
    "remainingSeconds": 600,
    "scheduledStart": "2026-05-31T10:40:00Z",
    "scheduledEnd": "2026-05-31T10:50:00Z",
    "expectedDurationMinutes": 10,
    "position": 5,
    "createdAt": "2026-05-31T09:00:00Z",
    "updatedAt": "2026-05-31T09:00:00Z"
  },
  "deltaSeconds": 60
}
```

Notes:

- type is included on runtime mutations and websocket snapshots.
- deltaSeconds is only present on adjust-time operations.
- programItem is the active runtime authority when present.
- nextProgramItem may be null when no upcoming item exists.

### ProgramItem Runtime Status Values

- scheduled
- in_progress
- paused
- ended
- canceled

### ProgramItem Runtime Fields

- runtimeDurationSeconds: Base runtime budget for this item.
- adjustmentSeconds: Net manual adjustment applied to runtime budget.
- actualStart: Runtime start timestamp.
- pausedAt: Current pause start timestamp, present only while paused.
- totalPausedDurationSeconds: Cumulative completed paused duration.
- endedRemainingSeconds: Frozen remaining time at end.
- actualEnd: Runtime end timestamp.
- pauseCount: Number of completed pause intervals.
- endedReason: End reason label, currently manual.
- remainingSeconds: Server-computed runtime countdown.

See detailed formulas in docs/programitem-time-calculation.md.

### Runtime Endpoint Summary

- GET /api/v1/sessions/:id returns unified runtime envelope.
- GET /api/v1/sessions/:id/current-program-item returns current and next ProgramItem snapshots.
- POST /api/v1/sessions/:id/start returns unified runtime envelope.
- POST /api/v1/sessions/:id/pause returns unified runtime envelope.
- POST /api/v1/sessions/:id/resume returns unified runtime envelope.
- POST /api/v1/sessions/:id/end returns unified runtime envelope.
- POST /api/v1/sessions/:id/adjust-time returns unified runtime envelope with deltaSeconds.
- POST /api/v1/program-items/:itemId/start returns unified runtime envelope.
- POST /api/v1/program-items/:itemId/pause returns unified runtime envelope.
- POST /api/v1/program-items/:itemId/resume returns unified runtime envelope.
- POST /api/v1/program-items/:itemId/adjust-time returns unified runtime envelope with deltaSeconds.
- POST /api/v1/program-items/:itemId/end returns unified runtime envelope.

### WebSocket Snapshot

The websocket connect snapshot and runtime updates now use the same unified runtime envelope shape.

## Base URL

**Development:** `http://localhost:8080`

## Authentication

Host operations require a **control token** obtained when creating a session.

## Request ID Correlation

All HTTP endpoints support request correlation via `X-Request-ID`.

- Optional request header: `X-Request-ID: <id>`
- If omitted, backend generates a request ID.
- Response includes `X-Request-ID` for every request.
- Backend structured logs include `request_id` to correlate API and related websocket-side errors.

**Header Authorization:**

```
X-Control-Token: <token>
```

**Query Parameter Authorization:**

```
?controlToken=<token>
```

---

## REST Endpoints

### Health Check

```
GET /healthz
```

Returns service health status.

**Response (200):**

```json
{ "status": "ok" }
```

---

### Create Session

```
POST /api/v1/sessions
Content-Type: application/json
```

Create a new timed session.

**Request Body:**

```json
{
  "title": "Engineering Talks Q&A",
  "speakerName": "Alice Johnson",
  "durationSeconds": 600
}
```

**Response (201):**

```json
{
  "session": {
    "id": "sess_abc123",
    "title": "Engineering Talks Q&A",
    "speakerName": "Alice Johnson",
    "durationSeconds": 600,
    "status": "CREATED",
    "remainingSeconds": 600,
    "createdAt": "2025-12-15T10:30:00Z"
  },
  "controlToken": "token_xyz789",
  "viewerPath": "/sessions/sess_abc123"
}
```

**Fields:**

- `id`: Unique session identifier
- `title`: Session title
- `speakerName`: Speaker or presenter name
- `durationSeconds`: Total session duration in seconds
- `status`: Session state (CREATED | LIVE | PAUSED | ENDED)
- `remainingSeconds`: Time remaining in seconds (server-authoritative)
- `createdAt`: ISO 8601 timestamp of creation
- `controlToken`: Secret token for host operations (store in sessionStorage)

---

### Get Session

```
GET /api/v1/sessions/:id
```

Retrieve current session state. No authentication required (read-only).

**Response (200):**

```json
{
  "session": {
    "id": "sess_abc123",
    "title": "Engineering Talks Q&A",
    "speakerName": "Alice Johnson",
    "durationSeconds": 600,
    "status": "LIVE",
    "remainingSeconds": 450,
    "createdAt": "2025-12-15T10:30:00Z"
  }
}
```

**Response (404):**

```json
{ "error": "session not found" }
```

---

## ProgramItem Endpoints (Phase 1 Contract)

ProgramItems are managed under a session timeline and must not overlap.

All mutation endpoints below require:

- `Authorization: Bearer <token>`
- `X-Control-Token: <token>` (or `?controlToken=<token>`)

Read endpoints require bearer authorization only.

### ProgramItem Shape

```json
{
  "id": "pi_abc123",
  "sessionId": "sess_abc123",
  "title": "Panel Discussion",
  "type": "panel",
  "status": "scheduled",
  "hostName": "Alice Johnson",
  "scheduledStart": "2026-05-28T10:20:00Z",
  "scheduledEnd": "2026-05-28T10:40:00Z",
  "expectedDurationMinutes": 20,
  "position": 4,
  "location": "Main Hall",
  "metadata": { "track": "engineering" },
  "createdAt": "2026-05-28T09:00:00Z",
  "updatedAt": "2026-05-28T09:00:00Z"
}
```

`status` values:

- `scheduled`
- `in_progress`
- `ended`
- `canceled`

### List ProgramItems

```
GET /api/v1/sessions/:id/program-items
Authorization: Bearer <token>
```

Returns ordered timeline items for a session, including canceled items.

### Get Current ProgramItem (Public Viewer)

```
GET /api/v1/sessions/:id/current-program-item
```

Returns current and next ProgramItem context for the supplied timestamp on the server.

Selection behavior:

- current selection prefers an `in_progress` item when one exists
- otherwise current is derived from scheduled window `scheduledStart <= now < scheduledEnd`
- next selection returns the first non-canceled upcoming scheduled item
- returns `null` fields when no matching item exists

Response body:

```json
{
  "programItem": {
    "id": "pi_abc123",
    "sessionId": "sess_abc123",
    "title": "Panel Discussion",
    "type": "panel",
    "status": "scheduled",
    "hostName": "Alice Johnson",
    "scheduledStart": "2026-05-28T10:20:00Z",
    "scheduledEnd": "2026-05-28T10:40:00Z",
    "expectedDurationMinutes": 20,
    "position": 4,
    "location": "Main Hall",
    "metadata": { "track": "engineering" },
    "createdAt": "2026-05-28T09:00:00Z",
    "updatedAt": "2026-05-28T09:00:00Z"
  },
  "nextProgramItem": {
    "id": "pi_def456",
    "sessionId": "sess_abc123",
    "title": "Q&A",
    "type": "q&a",
    "status": "scheduled",
    "hostName": "Alice Johnson",
    "scheduledStart": "2026-05-28T10:40:00Z",
    "scheduledEnd": "2026-05-28T10:50:00Z",
    "expectedDurationMinutes": 10,
    "position": 5,
    "location": "Main Hall",
    "metadata": { "track": "engineering" },
    "createdAt": "2026-05-28T09:00:00Z",
    "updatedAt": "2026-05-28T09:00:00Z"
  }
}
```

When no active item exists:

```json
{
  "programItem": null,
  "nextProgramItem": null
}
```

### Create ProgramItem

```
POST /api/v1/sessions/:id/program-items
Authorization: Bearer <token>
X-Control-Token: <token>
Content-Type: application/json
```

Request body:

```json
{
  "title": "Panel Discussion",
  "type": "panel",
  "hostName": "Alice Johnson",
  "scheduledStart": "2026-05-28T10:20:00Z",
  "scheduledEnd": "2026-05-28T10:40:00Z",
  "expectedDurationMinutes": 20,
  "position": 4,
  "location": "Main Hall",
  "metadata": { "track": "engineering" }
}
```

Validation:

- `scheduledStart < scheduledEnd`
- no overlap with existing scheduled ProgramItems in same session
- `position` must be unique within session

### Update ProgramItem

```
PATCH /api/v1/program-items/:itemId
Authorization: Bearer <token>
X-Control-Token: <token>
Content-Type: application/json
```

Allows title/type/host/time/duration/location/metadata/position updates.
If `scheduledStart` or `scheduledEnd` changes, overlap validation runs again.

### Cancel ProgramItem

```
POST /api/v1/program-items/:itemId/cancel
Authorization: Bearer <token>
X-Control-Token: <token>
```

Sets status to `canceled` and preserves the original timeline slot for audit and future metrics.

### Reorder ProgramItems (Bulk)

```
POST /api/v1/sessions/:id/program-items/reorder
Authorization: Bearer <token>
X-Control-Token: <token>
Content-Type: application/json
```

Request body:

```json
{
  "items": [
    { "id": "pi_1", "position": 1 },
    { "id": "pi_2", "position": 2 },
    { "id": "pi_3", "position": 3 }
  ]
}
```

Performs position updates transactionally to avoid transient uniqueness conflicts.

---

## Host Control Endpoints

All control endpoints require `X-Control-Token` authorization header or `?controlToken` query parameter.

### Start Session

```
POST /api/v1/sessions/:id/start
X-Control-Token: <token>
```

Transition session from CREATED → LIVE. Starts authoritative server timer.

**Response (200):**

```json
{
  "id": "sess_abc123",
  "status": "LIVE",
  "remainingSeconds": 600,
  "updatedAt": "2025-12-15T10:31:00Z"
}
```

**Possible Errors:**

- 401: Invalid or missing control token
- 400: Invalid state transition (e.g., already LIVE)

---

### Pause Session

```
POST /api/v1/sessions/:id/pause
X-Control-Token: <token>
```

Transition session from LIVE → PAUSED. Freezes countdown.

**Response (200):**

```json
{
  "id": "sess_abc123",
  "status": "PAUSED",
  "remainingSeconds": 300,
  "updatedAt": "2025-12-15T10:32:00Z"
}
```

---

### Resume Session

```
POST /api/v1/sessions/:id/resume
X-Control-Token: <token>
```

Transition session from PAUSED → LIVE. Resumes countdown from where it was paused.

**Response (200):**

```json
{
  "id": "sess_abc123",
  "status": "LIVE",
  "remainingSeconds": 300,
  "updatedAt": "2025-12-15T10:33:00Z"
}
```

---

### Adjust Time

```
POST /api/v1/sessions/:id/adjust-time
X-Control-Token: <token>
Content-Type: application/json
```

Add or subtract seconds from remaining time. Useful for extending or shortening sessions.

**Request Body:**

```json
{
  "deltaSeconds": 60
}
```

Use negative values to reduce time:

```json
{
  "deltaSeconds": -30
}
```

**Response (200):**

```json
{
  "id": "sess_abc123",
  "status": "LIVE",
  "remainingSeconds": 360,
  "updatedAt": "2025-12-15T10:34:00Z"
}
```

---

### End Session

```
POST /api/v1/sessions/:id/end
X-Control-Token: <token>
```

Transition session to ENDED. No further state changes allowed.

**Response (200):**

```json
{
  "id": "sess_abc123",
  "status": "ENDED",
  "remainingSeconds": 0,
  "updatedAt": "2025-12-15T10:35:00Z"
}
```

---

## WebSocket Endpoint

### Connect to Session

```
GET /ws/sessions/:id
Upgrade: websocket
```

Establishes a persistent WebSocket connection for receiving real-time session updates.

**No authentication required** (broadcast only, no mutations).

---

## WebSocket Messages

### Session Snapshot (on connect)

When the client connects, the server immediately sends the current session state:

```json
{
  "type": "SESSION_SNAPSHOT",
  "data": {
    "id": "sess_abc123",
    "title": "Engineering Talks Q&A",
    "speakerName": "Alice Johnson",
    "durationSeconds": 600,
    "status": "LIVE",
    "remainingSeconds": 420,
    "createdAt": "2025-12-15T10:30:00Z"
  }
}
```

### Session Update (on state change)

When the host changes session state (start, pause, resume, end, adjust), all connected WebSocket clients receive an update:

```json
{
  "type": "SESSION_UPDATE",
  "data": {
    "id": "sess_abc123",
    "status": "PAUSED",
    "remainingSeconds": 300,
    "updatedAt": "2025-12-15T10:32:00Z"
  }
}
```

### ProgramItem Update (on timeline change)

When ProgramItems are created, updated, canceled, or reordered, all connected clients receive one of:

- `PROGRAM_ITEM_CREATED`
- `PROGRAM_ITEM_UPDATED`
- `PROGRAM_ITEM_CANCELED`
- `PROGRAM_ITEMS_REORDERED`

User viewers should treat these as refresh triggers for
`GET /api/v1/sessions/:id/current-program-item` to stay aligned with server-side item selection.

---

## Error Responses

### 400 Bad Request

Invalid input or state transition:

```json
{ "error": "invalid request body" }
```

### 401 Unauthorized

Missing or invalid control token:

```json
{ "error": "unauthorized" }
```

### 404 Not Found

Session does not exist:

```json
{ "error": "session not found" }
```

### 500 Internal Server Error

Unexpected server error:

```json
{ "error": "internal server error" }
```

### 409 Conflict

Conflict on overlap, invalid state mutation, or duplicate position:

```json
{ "error": "program item overlaps with existing item" }
```

---

## Environment Variables

**Frontend:**

- `NEXT_PUBLIC_BACKEND_BASE_URL`: Backend API base URL (default: `http://localhost:8080`)
- `NEXT_PUBLIC_USER_APP_URL`: Viewer app origin used by admin UI to generate share links (default: `http://localhost:3001`)

**Backend:**

- `PORT`: Server port (default: `8080`)
- `GIN_MODE`: Gin logging mode (`debug` or `release`)
- `DB_DRIVER`: Session store driver (`sqlite` or `memory`, default: `sqlite`)
- `SQLITE_DB_PATH`: SQLite database file path (default: `./sessions.db`)
- `CORS_ALLOW_ORIGIN`: Allowed CORS origin (default: `*`)

---

## Rate Limiting

No rate limiting is implemented in the MVP. This should be added for production deployments.

---

## CORS

Cross-Origin Resource Sharing (CORS) is enabled for all origins in development (`CheckOrigin: true`). This should be restricted in production.
