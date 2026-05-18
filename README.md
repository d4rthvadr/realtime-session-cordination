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

### Local Development (3 commands)

```bash
make install          # Install dependencies
make backend          # Start backend on :8080
make frontend-user    # Start user app on :3001
make frontend-admin   # Start admin app on :3002
```

### Docker (1 command)

```bash
# From root directory (or cd infra/ for canonical location)
docker-compose up --build

# Services: http://localhost:8080 (backend), :3001 (user), :3002 (admin)
```

## Project Structure

```
.
├── frontend/              # Monorepo root (npm workspaces)
│   ├── apps/
│   │   ├── user/         # Public countdown viewer app + landing page (port 3001)
│   │   │   ├── app/      # Next.js 14 app directory
│   │   │   ├── components/
│   │   │   │   └── landing/  # Landing page sections (Navigation, Hero, Features, etc.)
│   │   │   ├── hooks/    # WebSocket and state hooks
│   │   │   ├── lib/      # Backend API client and utilities
│   │   │   └── store/    # Zustand stores
│   │   └── admin/        # Host control panel app (port 3002)
│   └── package.json
├── backend/               # Standalone Go service (Gin + WebSocket)
│   ├── cmd/api/
│   ├── internal/
│   │   ├── api/          # HTTP handlers and REST endpoints
│   │   ├── session/      # Domain model and state machine
│   │   └── ws/           # WebSocket hub and connection management
│   └── go.mod
├── infra/                 # Infrastructure & deployment
│   ├── docker/
│   │   ├── backend/      # Backend Docker image definition
│   │   └── frontend/     # Frontend Docker image definitions
│   └── docker-compose.yml    # Multi-service orchestration
├── docs/                  # Documentation
│   ├── setup.md          # Installation and development setup
│   ├── api.md            # API structure and endpoints
│   ├── architecture.md   # System architecture and design
│   └── cicd.md           # CI/CD pipelines and deployment
├── .github/
│   └── workflows/        # GitHub Actions workflows
│       ├── ci.yml        # Lint, build, test, security scan
│       └── build-deploy.yml  # Docker build and deployment
├── docker-compose.yml    # Root-level convenience symlink to infra/
└── Makefile              # Development convenience targets
```

## Technology Stack

**Frontend:**

- Next.js 14 (React 18, TypeScript)
- Tailwind CSS with custom Stripe-inspired design system
- Zustand for state management
- WebSocket for real-time updates
- Geist & Inter fonts for modern typography
- Material Symbols for consistent iconography

**Backend:**

- Go 1.25 with Gin framework
- Gorilla WebSocket for real-time connectivity
- In-memory session manager with authoritative timing

## Design System

The application features a **Stripe-inspired design language** with:

- **Color Palette:**
  - Primary Purple: `#635bff` (signature brand color)
  - Accent Cyan: `#00d4ff`
  - Deep Navy: `#0a2540` (text primary)
  - Clean whites and soft grays for backgrounds
- **Visual Style:**
  - Gradient backgrounds with blur effects
  - Soft shadows with hover interactions
  - Glassmorphic navigation with backdrop blur
  - Smooth transitions and scale animations
- **Typography:**
  - Display/Headlines: Geist (400, 600, 700, 800)
  - Body Text: Inter (400, 500, 600)
  - Enhanced font smoothing for crisp rendering

- **Components:**
  - Responsive navigation with fixed header
  - Hero section with gradient backgrounds
  - Feature cards with hover effects
  - Semantic alert indicators (Safe, Warning, Critical, Overtime)
  - Gradient CTA sections with decorative elements

## Landing Page Sections

The user app (`/apps/user`) features a comprehensive landing page with:

1. **Navigation** — Fixed header with logo and CTA button
2. **Hero Section** — Product headline with mockup screenshot and gradient background
3. **Semantic Alerts** — Visual showcase of timer states (Safe, Warning, Critical, Overtime)
4. **Problem Section** — Common coordination challenges addressed by the platform
5. **Features Grid** — 6-card layout showcasing platform capabilities
6. **CTA Section** — Full-width gradient call-to-action with decorative blur effects
7. **Footer** — Minimal branding and copyright

All sections are mobile-responsive with Tailwind breakpoints and feature smooth hover interactions.

## Features (MVP)

- ✅ **Marketing & Landing Page**
  - Stripe-inspired design with gradient backgrounds
  - Responsive navigation with fixed header
  - Hero section with product mockup (screen.png)
  - Semantic alerts showcase (Safe, Warning, Critical, Overtime)
  - Problem statement and solution sections
  - Features grid with icon badges
  - Gradient CTA section
  - Mobile-first responsive design

- ✅ **Session Management**
  - Session creation with title, speaker name, and duration
  - Host controls: start, pause, resume, end, adjust time
  - Unique share links for public viewers
  - Session persistence and state machine validation
  - Control token authorization for host operations

- ✅ **Real-time Experience**
  - Real-time countdown across all connected viewers
  - Visual urgency indicators (green/yellow/red/overtime)
  - WebSocket-powered synchronization
  - Instant updates to all participants

## Documentation

- [Setup Guide](docs/setup.md) — Installation, development, Docker, CI/CD
- [API Structure](docs/api.md) — REST endpoints and WebSocket protocol
- [Architecture](docs/architecture.md) — System design and data flow
- [CI/CD & Deployment](docs/cicd.md) — GitHub Actions, Docker, production deployment

## Status

✅ **MVP Complete**

- Both frontend apps fully implemented in TypeScript
- Backend fully implemented with REST + WebSocket
- End-to-end integration tested and validated
- Ready for development and testing
