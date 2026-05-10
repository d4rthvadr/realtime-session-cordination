# Realtime Session Coordination Platform — MVP PRD

## Overview

A lightweight realtime platform that helps speakers, moderators, and audiences stay synchronized around allocated presentation time through a shared public countdown experience.

The system provides:
- A host-controlled session timer
- Public synchronized countdown pages
- Realtime updates across all connected viewers
- Simple session lifecycle management

The long-term vision is to evolve into a realtime presentation coordination platform capable of synchronizing captions, slides, translated text streams, and interactive audience experiences.

---

# Problem Statement

Speakers and lecturers frequently exceed their allocated speaking time due to:
- lack of visible pacing
- poor time awareness during presentations
- absence of synchronized timing tools
- manual moderation inefficiencies

Current approaches are fragmented:
- personal phone timers
- moderator interventions
- projected clocks
- verbal interruptions

These methods create:
- audience frustration
- schedule overruns
- operational inefficiency
- poor session coordination

The platform introduces a shared source of temporal truth visible to all participants.

---

# Goals

## Primary Goal

Enable organizers and speakers to stay aligned with allocated presentation time through synchronized realtime countdowns.

## Secondary Goals

- Improve speaker pacing awareness
- Increase audience visibility into session timing
- Reduce moderator interruptions
- Provide a foundation for future realtime presentation features

---

# Non-Goals (MVP)

The following are intentionally excluded from MVP scope:

- Authentication systems
- Payments
- Multi-host collaboration
- Chat or audience reactions
- Slide synchronization
- Live captions
- Recording/playback
- Session analytics
- Calendar integrations
- Mobile apps
- Offline support

---

# Target Users

## Organizers / Hosts

Users who:
- create sessions
- control timers
- manage presentation timing

## Speakers

Users presenting content during a session.

## Audience / Listeners

Users viewing the public countdown page.

---

# User Stories

## Organizer

- As an organizer, I want to create a timed session so I can share a public countdown link.
- As an organizer, I want to start, pause, resume, and end a session.
- As an organizer, I want all viewers to stay synchronized in realtime.

## Speaker

- As a speaker, I want to know exactly how much time remains.
- As a speaker, I want visual urgency indicators as time runs out.

## Audience

- As an audience member, I want visibility into session timing and pacing.
- As an audience member, I want countdown updates without refreshing the page.

---

# MVP Scope

## Included Features

### Session Creation

Host can:
- create a session
- define title
- define speaker name
- define duration

### Public Share Link

Each session generates:
- unique public URL
- viewer-accessible realtime page

### Host Controls

Host can:
- start timer
- pause timer
- resume timer
- end timer
- extend or reduce remaining time

### Realtime Synchronization

All connected viewers receive synchronized updates through WebSockets.

### Countdown Experience

Viewer page displays:
- session title
- speaker name
- remaining time
- session state
- overtime indication

### Visual States

Timer colors:
- Green → safe
- Yellow → warning
- Red → critical
- Overtime → blinking/red emphasis

---

# Functional Requirements

## Session Lifecycle

States:
- CREATED
- LIVE
- PAUSED
- ENDED

### Allowed Transitions

CREATED → LIVE  
LIVE → PAUSED  
PAUSED → LIVE  
LIVE → ENDED  
PAUSED → ENDED

---

# Realtime Synchronization Requirements

## Server Authoritative Time

The backend owns:
- session start time
- pause durations
- remaining time calculations

Clients must not independently own timer truth.

## Synchronization Strategy

The server sends:
- start timestamps
- duration
- pause state
- correction events

Clients compute remaining time locally and periodically reconcile with server state.

## Reconnection Behavior

On reconnect:
- client requests latest session state
- timer rehydrates automatically

---

# Technical Requirements

## Frontend

### Recommended Stack
- Next.js
- React
- Tailwind CSS
- Zustand

### Responsibilities
- render countdown
- websocket connection
- fullscreen timer UI
- session state rendering

---

## Backend

### Recommended Stack
- Go
- Gin or Fiber
- Gorilla WebSocket

### Responsibilities
- session management
- websocket broadcasting
- authoritative timing calculations
- session lifecycle orchestration

---

## Database

### MVP Recommendation
PostgreSQL or SQLite.

Persistent storage requirements are minimal for MVP.

---

# Data Model

## Session

```json
{
  "id": "string",
  "title": "string",
  "speakerName": "string",
  "durationSeconds": 1800,
  "status": "CREATED | LIVE | PAUSED | ENDED",
  "startedAt": "timestamp",
  "pausedAt": "timestamp",
  "totalPausedDurationSeconds": 0,
  "createdAt": "timestamp"
}
```

---

# API Design

## Create Session

```http
POST /sessions
```

Request:
```json
{
  "title": "Kubernetes Workshop",
  "speakerName": "John Doe",
  "durationSeconds": 1800
}
```

---

## Get Session

```http
GET /sessions/:id
```

---

## Start Session

```http
POST /sessions/:id/start
```

---

## Pause Session

```http
POST /sessions/:id/pause
```

---

## Resume Session

```http
POST /sessions/:id/resume
```

---

## End Session

```http
POST /sessions/:id/end
```

---

## Adjust Time

```http
POST /sessions/:id/adjust-time
```

Request:
```json
{
  "deltaSeconds": 120
}
```

---

# WebSocket Design

## Connection Endpoint

```http
GET /ws/sessions/:id
```

## Event Types

### SESSION_STARTED

```json
{
  "type": "SESSION_STARTED",
  "startedAt": 1710000000
}
```

### SESSION_PAUSED

```json
{
  "type": "SESSION_PAUSED"
}
```

### SESSION_RESUMED

```json
{
  "type": "SESSION_RESUMED"
}
```

### SESSION_ENDED

```json
{
  "type": "SESSION_ENDED"
}
```

### TIME_ADJUSTED

```json
{
  "type": "TIME_ADJUSTED",
  "deltaSeconds": 120
}
```

---

# UX Requirements

## Countdown Display

Must:
- be large and readable
- support fullscreen mode
- remain visible from distance

## Accessibility

Should:
- maintain high contrast
- support responsive layouts
- avoid relying only on color cues

---

# Performance Requirements

## Realtime Performance

- countdown updates visible within 1 second
- websocket reconnect support
- minimal drift between clients

## Scalability

MVP target:
- 1,000 concurrent viewers per session

---

# Security Considerations

## Public Sessions

Viewer links are public.

## Host Control Protection

Host control endpoints require:
- signed control token
or
- secret session management link

---

# MVP Constraints

To maintain focus:
- single host per session
- single active speaker
- no account system
- no persistent audience identities

---

# Success Metrics

## Product Metrics

- sessions created
- sessions completed
- average session duration
- percentage of sessions finishing on time

## UX Metrics

- synchronization reliability
- reconnect success rate
- average websocket latency

---

# Future Roadmap

## Phase 2

- speaker queue
- multi-segment agendas
- moderator dashboard
- session templates

## Phase 3

- synchronized slides
- translated captions
- live links/resources
- audience reactions
- realtime annotations

## Phase 4

- analytics dashboards
- recordings/playback
- integrations (Zoom, Google Meet)
- mobile applications

---

# Open Questions

- Should sessions auto-expire after inactivity?
- Should organizers be able to transfer control?
- Should countdown audio alerts exist?
- Should overtime handling be configurable?
- Should public links support branding/themes?

---

# Recommended Architecture Direction

The system should follow:
- stateless backend nodes
- websocket-based realtime synchronization
- server-authoritative timing
- modular domain boundaries

The architecture should be designed to support future realtime collaborative features without major rewrites.

---

# Summary

The MVP focuses on solving a narrow but meaningful operational problem:
keeping speakers and audiences synchronized around time.

The system intentionally prioritizes:
- simplicity
- realtime reliability
- synchronization accuracy
- operational clarity

Future expansion should build on the core concept of shared realtime presentation coordination.
