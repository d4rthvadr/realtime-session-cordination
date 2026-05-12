# API Structure & Endpoints

## Base URL

**Development:** `http://localhost:8080`

## Authentication

Host operations require a **control token** obtained when creating a session.

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
