.PHONY: help backend run-backend frontend-user run-user frontend-admin run-admin install build dev

help:
	@echo "Realtime Session Coordination - Available Commands"
	@echo ""
	@echo "Setup:"
	@echo "  make install          Install all dependencies (npm + go)"
	@echo "  make build            Build all apps (frontend + backend)"
	@echo ""
	@echo "Development:"
	@echo "  make backend          Start backend on :8080"
	@echo "  make run-backend      (alias for 'make backend')"
	@echo "  make frontend-user    Start user app on :3001"
	@echo "  make run-user         (alias for 'make frontend-user')"
	@echo "  make frontend-admin   Start admin app on :3002"
	@echo "  make run-admin        (alias for 'make frontend-admin')"
	@echo ""
	@echo "Notes:"
	@echo "  - Backend requires Go and will listen on http://localhost:8080"
	@echo "  - User app requires Node.js and will listen on http://localhost:3001"
	@echo "  - Admin app requires Node.js and will listen on http://localhost:3002"
	@echo "  - Environment: NEXT_PUBLIC_BACKEND_BASE_URL (default: http://localhost:8080)"
	@echo ""

install:
	@echo "Installing dependencies..."
	cd frontend && npm install
	cd backend && go mod download
	@echo "✓ Dependencies installed"

build:
	@echo "Building frontend apps..."
	cd frontend && npm run build -w @realtime/user
	cd frontend && npm run build -w @realtime/admin
	@echo "Building backend..."
	cd backend && go build -o bin/api ./cmd/api
	@echo "✓ All builds complete"

backend: run-backend
run-backend:
	@echo "Starting backend on :8080..."
	cd backend && go run ./cmd/api

frontend-user: run-user
run-user:
	@echo "Starting user app on :3001..."
	cd frontend && npm run dev -w @realtime/user

frontend-admin: run-admin
run-admin:
	@echo "Starting admin app on :3002..."
	cd frontend && npm run dev -w @realtime/admin

dev:
	@echo "To run all services, use three separate terminal tabs:"
	@echo "  Tab 1: make backend"
	@echo "  Tab 2: make frontend-user"
	@echo "  Tab 3: make frontend-admin"
	@echo ""
	@echo "Or use tmux:"
	@echo "  tmux new-session -d -s realtime -n backend"
	@echo "  tmux send-keys -t realtime:backend 'make backend' Enter"
	@echo "  tmux new-window -t realtime -n user 'make frontend-user'"
	@echo "  tmux new-window -t realtime -n admin 'make frontend-admin'"
	@echo "  tmux attach -t realtime"
