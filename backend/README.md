# Backend (Standalone Service)

This folder contains the standalone Go backend service for realtime session coordination.

## Planned Stack

- Go
- Gin
- Gorilla WebSocket

## Status

Scaffold initialized. API and WebSocket implementation will be added in the next tasks.

## Logging

Backend logging uses structured logs via Go `slog`.

### Environment Variables

- `LOG_LEVEL`: `debug` | `info` | `warn` | `error` (default: `info`)
- `LOG_FORMAT`: `json` | `text` (default: `json`)

### Correlation (`request_id`)

- Every HTTP request is assigned a request ID.
- If the client sends `X-Request-ID`, that value is reused.
- If absent, the server generates a value like `req_<hex>`.
- The request ID is returned in the response header `X-Request-ID`.
- Access logs include `request_id`.
- Related WebSocket error logs emitted during request handling include `request_id` when available.
