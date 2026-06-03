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

### Session Runtime Field Deprecation (Applied)

Session runtime columns were removed from persistence. Session now remains lifecycle/container metadata (`id`, `title`, `speakerName`, `durationSeconds`, `status`, `createdAt`).

Removed session persistence fields:

- startedAt
- pausedAt
- totalPausedDurationSeconds
- adjustmentSeconds
- endedRemainingSeconds

See detailed formulas in docs/programitem-time-calculation.md.

### Runtime Endpoint Summary

- GET /api/v1/sessions/:id returns unified runtime envelope.
- GET /api/v1/sessions/:id/current-program-item returns current and next ProgramItem snapshots.
- GET /api/v1/sessions/:id/logs returns paginated session log entries (auth required).
- GET /api/v1/sessions/:id/analytics returns per-session analytics summary (auth required).
- GET /api/v1/analytics/overview returns cross-session analytics overview (auth required).
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

Session log appends are also broadcast on the same session channel as lightweight payloads:

```json
{
  "event": "session_log_appended",
  "sessionLog": {
    "id": "log_abc123",
    "sessionId": "sess_abc123",
    "programItemId": "pi_abc123",
    "eventType": "PROGRAM_ITEM_PAUSED",
    "message": "Paused program item \"Panel\"",
    "metadata": { "deltaSeconds": 60 },
    "occurredAt": "2026-06-02T10:20:00Z",
    "requestId": "req_123",
    "createdAt": "2026-06-02T10:20:00Z"
  }
}
```

These payloads intentionally do not include runtime `session/programItem` fields.

### Analytics Summary Endpoints (Phase 2)

Analytics raw event pipeline contract (Phase 2A):

- `analytics_events` fields:
  - `id` (string, primary key)
  - `session_id` (string, required)
  - `program_item_id` (string, optional)
  - `event_key` (string, required)
  - `occurred_at` (RFC3339 timestamp, required)
  - `ingested_at` (RFC3339 timestamp, required)
  - `source` (enum: `server`, `client`, required)
  - `payload_json` (JSON string payload, required)
- `analytics_outbox` states:
  - `pending`
  - `processing`
  - `processed`
  - `dead_letter`
- `analytics_checkpoints` fields:
  - `worker_name` (string, primary key)
  - `last_event_id` (string, required)
  - `updated_at` (RFC3339 timestamp, required)

Per-session analytics:

`GET /api/v1/sessions/:id/analytics`

Response shape:

```json
{
  "analytics": {
    "sessionId": "sess_abc123",
    "sessionStatus": "ENDED",
    "sessionDurationSeconds": 3600,
    "programItemCount": 8,
    "scheduledCount": 0,
    "inProgressCount": 0,
    "pausedCount": 0,
    "endedCount": 7,
    "canceledCount": 1,
    "plannedSeconds": 3600,
    "effectiveBudgetSeconds": 3720,
    "totalAdjustmentSeconds": 120,
    "totalPauseSeconds": 180,
    "totalPauseCount": 3,
    "endedOnTimeCount": 5,
    "overrunItemCount": 2,
    "totalOverrunSeconds": 65,
    "totalUnderrunSeconds": 140,
    "endedOnTimeRatio": 0.7142857143,
    "computedAt": "2026-06-02T11:00:00Z"
  }
}
```

Platform overview analytics:

`GET /api/v1/analytics/overview`

Response shape:

```json
{
  "overview": {
    "totalSessions": 24,
    "createdSessions": 3,
    "liveSessions": 1,
    "pausedSessions": 0,
    "endedSessions": 20,
    "totalProgramItems": 164,
    "endedProgramItems": 141,
    "onTimeEndedProgramItems": 109,
    "overrunProgramItems": 32,
    "totalSessionDurationSeconds": 54000,
    "totalPlannedSeconds": 53520,
    "effectiveBudgetSeconds": 54840,
    "totalAdjustmentSeconds": 1320,
    "totalPauseSeconds": 2180,
    "totalPauseCount": 79,
    "totalOverrunSeconds": 1820,
    "totalUnderrunSeconds": 3015,
    "sessionCompletionRatio": 0.8333333333,
    "programItemOnTimeRatio": 0.7730496454,
    "computedAt": "2026-06-02T11:00:00Z"
  }
}
```

### Session Log Taxonomy (Phase 1A)

Session logs are append-only timeline records generated from host mutations and server-driven cascades.

Canonical event types:

- SESSION_CREATED
- SESSION_STARTED
- SESSION_PAUSED
- SESSION_RESUMED
- SESSION_ENDED
- SESSION_TIME_ADJUSTED
- PROGRAM_ITEM_CREATED
- PROGRAM_ITEM_UPDATED
- PROGRAM_ITEMS_REORDERED
- PROGRAM_ITEM_CANCELED
- PROGRAM_ITEM_STARTED
- PROGRAM_ITEM_PAUSED
- PROGRAM_ITEM_RESUMED
- PROGRAM_ITEM_ENDED
- PROGRAM_ITEM_TIME_ADJUSTED
- CASCADE_PROGRAM_ITEM_PAUSED_BY_SESSION
- CASCADE_PROGRAM_ITEM_RESUMED_BY_SESSION
- CASCADE_PROGRAM_ITEM_ENDED_BY_SESSION
- CASCADE_PROGRAM_ITEM_TIME_ADJUSTED_BY_SESSION

Human-readable trail message templates:

- SESSION_CREATED: Created session "{sessionTitle}"
- SESSION_STARTED: Started session "{sessionTitle}"
- SESSION_PAUSED: Paused session "{sessionTitle}"
- SESSION_RESUMED: Resumed session "{sessionTitle}"
- SESSION_ENDED: Ended session "{sessionTitle}"
- SESSION_TIME_ADJUSTED: Adjusted session time by {+/-deltaSeconds}s
- PROGRAM_ITEM_CREATED: Added program item "{programItemTitle}"
- PROGRAM_ITEM_UPDATED: Updated program item "{programItemTitle}"
- PROGRAM_ITEMS_REORDERED: Reordered {count} program items
- PROGRAM_ITEM_CANCELED: Canceled program item "{programItemTitle}"
- PROGRAM_ITEM_STARTED: Started program item "{programItemTitle}"
- PROGRAM_ITEM_PAUSED: Paused program item "{programItemTitle}"
- PROGRAM_ITEM_RESUMED: Resumed program item "{programItemTitle}"
- PROGRAM_ITEM_ENDED: Ended program item "{programItemTitle}"
- PROGRAM_ITEM_TIME_ADJUSTED: Adjusted program item "{programItemTitle}" by {+/-deltaSeconds}s
- CASCADE_PROGRAM_ITEM_PAUSED_BY_SESSION: Auto-paused program item "{programItemTitle}" because session was paused
- CASCADE_PROGRAM_ITEM_RESUMED_BY_SESSION: Auto-resumed program item "{programItemTitle}" because session was resumed
- CASCADE_PROGRAM_ITEM_ENDED_BY_SESSION: Auto-ended program item "{programItemTitle}" because session was ended
- CASCADE_PROGRAM_ITEM_TIME_ADJUSTED_BY_SESSION: Auto-adjusted program item "{programItemTitle}" by {+/-deltaSeconds}s from session adjustment

This taxonomy is the canonical source for timeline semantics and should be reused by API serializers and admin UI mapping.

### Session Log Persistence (Phase 1B)

Session logs are stored as append-only rows with the following fields:

- id
- session_id
- program_item_id (nullable)
- event_type
- message
- metadata (JSON, nullable)
- occurred_at (RFC3339 timestamp)
- request_id (nullable)
- created_at (RFC3339 timestamp)

Ordering rule for timeline reads:

- ORDER BY occurred_at DESC, created_at DESC, id DESC

Pagination defaults used by the log manager:

- default limit: 50
- max limit: 200

Phase 1B only adds persistence/wiring. Log emission endpoints are implemented in later phases.

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
- `createdAt`: ISO 8601 timestamp of creation
- `controlToken`: Secret token for host operations (store in sessionStorage)

---

### Get Session

```
GET /api/v1/sessions/:id
```

Retrieve current runtime envelope. No authentication required (read-only).

**Response (200):**

```json
{
  "type": "SESSION_SNAPSHOT",
  "session": {
    "id": "sess_abc123",
    "title": "Engineering Talks Q&A",
    "speakerName": "Alice Johnson",
    "durationSeconds": 600,
    "status": "LIVE",
    "createdAt": "2025-12-15T10:30:00Z"
  },
  "programItem": null,
  "nextProgramItem": null
}
```

**Response (404):**

```json
{ "error": "session not found" }
```

---

### List Session Logs

```
GET /api/v1/sessions/:id/logs
Authorization: Bearer <token>
```

Returns append-only session log entries ordered by newest first.

Optional query params:

- `limit` (int, default 50, max 200)
- `offset` (int, default 0)
- `eventType` (exact canonical event type, e.g. `SESSION_PAUSED`)
- `entityType` (`session` | `program_item` | `cascade`)

**Response (200):**

```json
{
  "logs": [
    {
      "id": "log_abc123",
      "sessionId": "sess_abc123",
      "programItemId": "pi_abc123",
      "eventType": "PROGRAM_ITEM_PAUSED",
      "message": "Paused program item \"Panel\"",
      "metadata": {},
      "occurredAt": "2026-06-02T10:20:00Z",
      "requestId": "req_123",
      "createdAt": "2026-06-02T10:20:00Z"
    }
  ],
  "count": 1
}
```

---

## ProgramItem Endpoints (Phase 1 Contract)

ProgramItems are managed under a session timeline and must not overlap.

All mutation endpoints below require:

- `Authorization: Bearer <token>`
- `X-Control-Token: <token>` (or `?controlToken=<token>`)

Read endpoints require bearer authorization only.

### ProgramItem Shape

Base scheduling fields:

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
- `paused`
- `ended`
- `canceled`

Runtime fields for active contract (`runtimeDurationSeconds`, `remainingSeconds`, `actualStart`, `pausedAt`, `totalPausedDurationSeconds`, `adjustmentSeconds`, `endedRemainingSeconds`, `actualEnd`, `pauseCount`, `endedReason`) are documented in Runtime Contract Update above.

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
This endpoint is retained as a compatibility read endpoint; primary viewer sync should consume `GET /api/v1/sessions/:id` unified runtime envelope and websocket runtime updates.

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

Transition session from CREATED → LIVE.

**Response (200):**

```json
{
  "type": "SESSION_STARTED",
  "session": {
    "id": "sess_abc123",
    "status": "LIVE"
  },
  "programItem": null,
  "nextProgramItem": null
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

Transition session from LIVE → PAUSED.

**Response (200):**

```json
{
  "type": "SESSION_PAUSED",
  "session": {
    "id": "sess_abc123",
    "status": "PAUSED"
  },
  "programItem": {
    "id": "pi_abc123",
    "status": "paused",
    "remainingSeconds": 300
  },
  "nextProgramItem": null
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
  "type": "SESSION_RESUMED",
  "session": {
    "id": "sess_abc123",
    "status": "LIVE"
  },
  "programItem": {
    "id": "pi_abc123",
    "status": "in_progress",
    "remainingSeconds": 300
  },
  "nextProgramItem": null
}
```

---

### Adjust Time

```
POST /api/v1/sessions/:id/adjust-time
X-Control-Token: <token>
Content-Type: application/json
```

Add or subtract seconds from runtime. Useful for extending or shortening live delivery.

Runtime behavior:

- If an active ProgramItem runtime exists (`in_progress` or `paused`), delta is applied to that ProgramItem.
- If no active ProgramItem runtime exists, delta updates session `durationSeconds` lifecycle metadata.

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
  "type": "TIME_ADJUSTED",
  "session": {
    "id": "sess_abc123",
    "status": "LIVE"
  },
  "programItem": {
    "id": "pi_abc123",
    "status": "in_progress",
    "remainingSeconds": 360
  },
  "nextProgramItem": null,
  "deltaSeconds": 60
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
  "type": "SESSION_ENDED",
  "session": {
    "id": "sess_abc123",
    "status": "ENDED"
  },
  "programItem": null,
  "nextProgramItem": null
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

When the client connects, the server immediately sends the unified runtime envelope:

```json
{
  "type": "SESSION_SNAPSHOT",
  "session": {
    "id": "sess_abc123",
    "title": "Engineering Talks Q&A",
    "speakerName": "Alice Johnson",
    "durationSeconds": 600,
    "status": "LIVE",
    "createdAt": "2025-12-15T10:30:00Z"
  },
  "programItem": {
    "id": "pi_abc123",
    "status": "in_progress",
    "remainingSeconds": 420
  },
  "nextProgramItem": null
}
```

### Session Update (on state change)

When runtime state changes (session or program item actions), connected clients receive an updated runtime envelope.

```json
{
  "type": "SESSION_PAUSED",
  "session": {
    "id": "sess_abc123",
    "status": "PAUSED",
    "remainingSeconds": 300
  },
  "programItem": {
    "id": "pi_abc123",
    "status": "paused",
    "remainingSeconds": 300
  },
  "nextProgramItem": null
}
```

### ProgramItem Update (on timeline change)

When ProgramItems are created, updated, canceled, or reordered, connected clients may receive event messages such as:

- `PROGRAM_ITEM_CREATED`
- `PROGRAM_ITEM_UPDATED`
- `PROGRAM_ITEM_CANCELED`
- `PROGRAM_ITEMS_REORDERED`

For runtime timer and now/next state, user viewers should rely on unified runtime envelope updates.

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
