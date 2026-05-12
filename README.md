# Realtime Session Coordination Platform

A lightweight realtime platform that helps speakers, moderators, and audiences stay synchronized around allocated presentation time through a shared public countdown experience.

## Intent

This project provides a **host-controlled session timer** with a **public synchronized countdown page** and **realtime updates** across all connected viewers. It solves the problem of speakers exceeding their allocated time by providing:

- Visible countdown for all participants (speakers, moderators, audience)
- Real-time synchronization across multiple viewers
- Host controls to pause, resume, adjust, and end sessions
- Visual urgency indicators (safe → warning → critical → overtime)

The long-term vision is to evolve into a comprehensive realtime presentation coordination platform capable of synchronizing captions, slides, translated text streams, and interactive audience experiences.

## Quick Start

See [Setup Guide](docs/setup.md) for detailed installation and run instructions.

### TL;DR

```bash
make install          # Install dependencies
make backend          # Start backend on :8080
make frontend-user    # Start user app on :3001
make frontend-admin   # Start admin app on :3002
```

## Project Structure

```
.
├── frontend/              # Monorepo root (npm workspaces)
│   ├── apps/
│   │   ├── user/         # Public countdown viewer app (port 3001)
│   │   └── admin/        # Host control panel app (port 3002)
│   └── package.json
├── backend/               # Standalone Go service (Gin + WebSocket)
│   ├── cmd/api/
│   ├── internal/
│   │   ├── api/          # HTTP handlers and REST endpoints
│   │   ├── session/      # Domain model and state machine
│   │   └── ws/           # WebSocket hub and connection management
│   └── go.mod
├── docs/                  # Documentation
│   ├── setup.md          # Installation and development setup
│   ├── api.md            # API structure and endpoints
│   └── architecture.md   # System architecture and design
└── Makefile              # Development convenience targets
```

## Technology Stack

**Frontend:**

- Next.js 14 (React 18, TypeScript)
- Tailwind CSS for styling
- Zustand for state management
- WebSocket for real-time updates

**Backend:**

- Go 1.25 with Gin framework
- Gorilla WebSocket for real-time connectivity
- In-memory session manager with authoritative timing

## Features (MVP)

- ✅ Session creation with title, speaker name, and duration
- ✅ Host controls: start, pause, resume, end, adjust time
- ✅ Real-time countdown across all connected viewers
- ✅ Visual urgency indicators (green/yellow/red/overtime)
- ✅ Unique share links for public viewers
- ✅ Session persistence and state machine validation
- ✅ Control token authorization for host operations

## Documentation

- [API Structure](docs/api.md) — REST endpoints and WebSocket protocol
- [Architecture](docs/architecture.md) — System design and data flow
- [Setup Guide](docs/setup.md) — Installation and development workflow

## Status

✅ **MVP Complete**

- Both frontend apps fully implemented in TypeScript
- Backend fully implemented with REST + WebSocket
- End-to-end integration tested and validated
- Ready for development and testing
