# Setup & Development Guide

This guide covers installing dependencies, running the platform locally, and development workflows.

## Prerequisites

### Required

- **Node.js 18+** (for frontend)
- **Go 1.20+** (for backend)
- **npm** or **pnpm** (frontend package manager; pnpm recommended for monorepo)
- **git**

### Optional

- **Make** — convenience targets for running services (included on macOS/Linux)
- **tmux** — for running multiple services in one terminal session
- **VS Code** — recommended editor with TypeScript support

---

## Quick Start (5 minutes)

### 1. Clone Repository

```bash
git clone https://github.com/your-org/realtime-session-coordination.git
cd realtime-session-coordination
```

### 2. Install Dependencies

```bash
make install
```

Or manually:

```bash
cd frontend && npm install
cd ../backend && go mod download
cd ..
```

### 3. Start All Services

**Option A: Three Terminal Tabs**

```bash
# Terminal 1: Backend
make backend

# Terminal 2: User App
make frontend-user

# Terminal 3: Admin App
make frontend-admin
```

**Option B: Single Terminal with tmux**

```bash
tmux new-session -d -s realtime -n backend
tmux send-keys -t realtime:backend 'make backend' Enter
tmux new-window -t realtime -n user 'make frontend-user'
tmux new-window -t realtime -n admin 'make frontend-admin'
tmux attach -t realtime
```

### 4. Open Browser

- **Admin App:** http://localhost:3002
- **User App:** http://localhost:3001 (accessed via share link from admin)

---

## Detailed Setup

### Node.js / npm Setup

#### macOS (Homebrew)

```bash
brew install node
node --version    # v18.0.0 or later
npm --version     # 9.0.0 or later
```

#### Windows / Linux

Download from https://nodejs.org/ (LTS version recommended).

#### Install pnpm (Optional, Recommended)

```bash
npm install -g pnpm
pnpm --version
```

---

### Go Setup

#### macOS (Homebrew)

```bash
brew install go
go version    # go1.20 or later
```

#### Windows / Linux

Download from https://golang.org/dl/

Verify installation:

```bash
go version
```

---

### Frontend Monorepo

The frontend uses **pnpm workspaces** to manage two independent Next.js applications.

#### Directory Structure

```
frontend/
├── package.json          # Root workspace config
├── pnpm-workspace.yaml   # Workspace definition
├── apps/
│   ├── user/             # User countdown app
│   │   ├── package.json
│   │   ├── src/
│   │   ├── public/
│   │   └── next.config.js
│   └── admin/            # Admin control panel
│       ├── package.json
│       ├── src/
│       ├── public/
│       └── next.config.js
└── node_modules/
```

#### Install Frontend Dependencies

```bash
cd frontend
npm install
# or
pnpm install
```

Verify installation:

```bash
npm ls            # List all packages
npm ls @realtime/user
npm ls @realtime/admin
```

---

### Backend Setup

The backend is a standalone Go service.

#### Directory Structure

```
backend/
├── go.mod              # Go module definition
├── go.sum              # Dependency lockfile
├── cmd/
│   └── api/
│       └── main.go     # Entry point
└── internal/
    ├── api/            # HTTP handlers
    ├── session/        # Domain model
    └── ws/             # WebSocket hub
```

#### Install Backend Dependencies

```bash
cd backend
go mod download
```

Verify installation:

```bash
go list ./...     # List all packages
```

---

## Running Services

### Backend (Go + Gin)

**Development Mode:**

```bash
cd backend
go run ./cmd/api
```

**Output:**

```
[GIN-debug] Loaded HTML Templates (0):
[GIN-debug] Listening and serving HTTP on :8080
[GIN-debug] POST   /api/v1/sessions
[GIN-debug] GET    /api/v1/sessions/:id
[GIN-debug] POST   /api/v1/sessions/:id/start
[GIN-debug] POST   /api/v1/sessions/:id/pause
[GIN-debug] POST   /api/v1/sessions/:id/resume
[GIN-debug] POST   /api/v1/sessions/:id/end
[GIN-debug] POST   /api/v1/sessions/:id/adjust-time
[GIN-debug] GET    /ws/sessions/:id
```

**Production Build:**

```bash
cd backend
go build -o bin/api ./cmd/api
./bin/api
```

**Environment Variables:**

- `PORT` — server port (default: 8080)
- `GIN_MODE` — `debug` or `release` (default: debug)

### User App (Next.js)

**Development Mode:**

```bash
cd frontend
npm run dev -w @realtime/user
```

**Output:**

```
▲ Next.js 14.2.3
  - Local:        http://localhost:3001
  - Environments: .env.local

✓ Ready in 2.5s
```

**Production Build:**

```bash
cd frontend
npm run build -w @realtime/user
npm run start -w @realtime/user
```

### Admin App (Next.js)

**Development Mode:**

```bash
cd frontend
npm run dev -w @realtime/admin
```

**Output:**

```
▲ Next.js 14.2.3
  - Local:        http://localhost:3002
  - Environments: .env.local

✓ Ready in 2.5s
```

**Production Build:**

```bash
cd frontend
npm run build -w @realtime/admin
npm run start -w @realtime/admin
```

---

## Environment Variables

### Frontend

**File:** `frontend/.env.local` (create if doesn't exist)

```env
# Backend API base URL
NEXT_PUBLIC_BACKEND_BASE_URL=http://localhost:8080

# WebSocket URL (auto-derived from BACKEND_BASE_URL)
# NEXT_PUBLIC_BACKEND_WS_URL=ws://localhost:8080
```

### Backend

**File:** `backend/.env.local` (optional, Go reads from environment)

```env
# Server port
PORT=8080

# Gin mode (debug or release)
GIN_MODE=debug
```

---

## Makefile Targets

Convenience targets for local development. Requires `make` (standard on macOS/Linux).

```bash
make help              # Display all targets
make install           # Install dependencies
make build             # Build frontend + backend
make backend           # Run backend
make frontend-user     # Run user app
make frontend-admin    # Run admin app
```

---

## Development Workflow

### 1. Making Code Changes

#### Frontend

```bash
# Terminal 1: Watch user app
cd frontend
npm run dev -w @realtime/user

# Terminal 2: Watch admin app
cd frontend
npm run dev -w @realtime/admin
```

Next.js hot-reloads on file changes automatically.

#### Backend

```bash
# Terminal 3: Watch backend
cd backend
go run ./cmd/api
```

To auto-reload on changes, use **air** or **nodemon** equivalents:

```bash
go install github.com/cosmtrek/air@latest
cd backend
air
```

Or use VS Code's built-in Run and Debug features.

### 2. Testing Workflows

#### Manual Testing via Browser

1. Start all three services (backend, user app, admin app)
2. Open http://localhost:3002 (admin)
3. Create a session:
   - Fill form with title, speaker name, duration
   - Click "Create Session"
   - Note the control token in sessionStorage
4. Click "Copy Viewer Link" and paste in new tab
5. In admin tab, click "Start Session"
6. Observe countdown in user tab updates in real-time

#### REST API Testing

Use `curl`, Postman, or `httpie`:

```bash
# Create session
curl -X POST http://localhost:8080/api/v1/sessions \
  -H "Content-Type: application/json" \
  -d '{
    "title": "My Presentation",
    "speakerName": "Alice",
    "durationSeconds": 600
  }'

# Response includes controlToken
# Save it for next requests

CONTROL_TOKEN="token_xyz"
SESSION_ID="sess_abc123"

# Get session
curl http://localhost:8080/api/v1/sessions/$SESSION_ID

# Start session
curl -X POST http://localhost:8080/api/v1/sessions/$SESSION_ID/start \
  -H "X-Control-Token: $CONTROL_TOKEN"

# Pause session
curl -X POST http://localhost:8080/api/v1/sessions/$SESSION_ID/pause \
  -H "X-Control-Token: $CONTROL_TOKEN"
```

#### WebSocket Testing

Use `wscat` or browser DevTools:

```bash
# Install wscat
npm install -g wscat

# Connect to session
wscat -c ws://localhost:8080/ws/sessions/$SESSION_ID
```

Browser DevTools (Console):

```javascript
const ws = new WebSocket("ws://localhost:8080/ws/sessions/sess_abc123");
ws.onmessage = (e) => console.log(JSON.parse(e.data));
```

### 3. Building for Production

#### Frontend

```bash
cd frontend
npm run build -w @realtime/user
npm run build -w @realtime/admin
# Output: .next/ directories in each app
```

#### Backend

```bash
cd backend
go build -o bin/api ./cmd/api
# Output: bin/api binary
```

---

## Troubleshooting

### Port Already in Use

**Problem:** "Address already in use :3001" or similar

**Solution:** Kill process on port or use different port:

```bash
# macOS/Linux: find and kill process
lsof -i :3001
kill -9 <PID>

# Or use different port
PORT=3003 npm run dev -w @realtime/user
```

### Node Modules Corrupted

**Problem:** "Cannot find module @realtime/user" or TypeScript errors

**Solution:** Clean and reinstall:

```bash
cd frontend
rm -rf node_modules pnpm-lock.yaml
pnpm install
```

### Go Modules Out of Sync

**Problem:** "missing go.sum entry"

**Solution:**

```bash
cd backend
go mod tidy
go mod download
```

### Backend Won't Connect to Frontend

**Problem:** Frontend shows "connection refused" or WebSocket errors

**Verify:**

1. Backend is running: `curl http://localhost:8080/healthz`
2. Frontend env var: Check `NEXT_PUBLIC_BACKEND_BASE_URL` in browser console
3. CORS: Backend allows all origins in dev; shouldn't be an issue

### TypeScript Errors in Frontend

**Problem:** Red squiggles, build errors

**Solution:**

```bash
cd frontend
npm run build -w @realtime/user
npm run build -w @realtime/admin
```

Ensure no JavaScript `.js` files remain alongside new TypeScript `.ts` files.

---

## IDE Setup (VS Code)

### Recommended Extensions

- **ES7+ React/Redux/React-Native Snippets** (dsznajder.es7-react-js-snippets)
- **Tailwind CSS IntelliSense** (bradlc.vscode-tailwindcss)
- **Go** (golang.go)
- **Thunder Client** or **REST Client** (for API testing)

### Workspace Settings

Create `.vscode/settings.json`:

```json
{
  "typescript.enablePromptUseWorkspaceTypeScriptVersion": true,
  "typescript.tsdk": "frontend/node_modules/typescript/lib",
  "editor.defaultFormatter": "esbenp.prettier-vscode",
  "editor.formatOnSave": true,
  "[go]": {
    "editor.defaultFormatter": "golang.go",
    "editor.formatOnSave": true
  }
}
```

### Debug Configurations

Create `.vscode/launch.json`:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Go: Backend API",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/backend/cmd/api",
      "preLaunchTask": "go: build"
    },
    {
      "name": "Node: User App",
      "type": "node",
      "request": "launch",
      "program": "${workspaceFolder}/frontend/node_modules/.bin/next",
      "args": ["dev", "-w", "@realtime/user"],
      "cwd": "${workspaceFolder}/frontend"
    }
  ]
}
```

---

## CI/CD (Optional)

### GitHub Actions Example

Create `.github/workflows/ci.yml`:

```yaml
name: CI

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
        with:
          node-version: 18
      - uses: actions/setup-go@v3
        with:
          go-version: 1.20

      - name: Install dependencies
        run: |
          cd frontend && npm install
          cd ../backend && go mod download

      - name: Build frontend
        run: cd frontend && npm run build -w @realtime/user && npm run build -w @realtime/admin

      - name: Build backend
        run: cd backend && go build -o bin/api ./cmd/api

      - name: Lint (frontend)
        run: cd frontend && npm run lint
```

---

## Next Steps

1. **Review Architecture** — Read [architecture.md](architecture.md)
2. **Explore API** — Read [api.md](api.md)
3. **Make Code Changes** — Follow development workflow above
4. **Test End-to-End** — Create session, start countdown, verify sync
5. **Deploy** — Build binaries and deploy to server (see deployment guides)

For questions or issues, check the main [README.md](../README.md) or open an issue on GitHub.
