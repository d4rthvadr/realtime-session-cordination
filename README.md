# Realtime Session Coordination

Repository layout for MVP:

- frontend: monorepo root
  - apps/user: public countdown app
  - apps/admin: host/admin control app
  - packages/ui: shared UI package (optional)
- backend: standalone Go service (Gin + WebSocket)
- docs: product and architecture docs

## Current Status

- Task 1 complete: repository topology and baseline manifests created.
- Next task: scaffold the user frontend app.

## Planned Flow

1. Build user-facing app first.
2. Build admin app second.
3. Build standalone backend third.
4. Integrate and run realtime smoke checks.
