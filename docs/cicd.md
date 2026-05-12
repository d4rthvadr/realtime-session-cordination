# CI/CD & Deployment Guide

This guide covers the complete CI/CD pipeline, deployment strategies, and production considerations.

---

## Overview

The project uses **GitHub Actions** for automated testing, building, and deployment.

**Key Components:**

- **CI Pipeline** (`.github/workflows/ci.yml`) — Runs on every commit/PR
- **Build & Deploy** (`.github/workflows/build-deploy.yml`) — Builds Docker images and deploys
- **Docker** — Multi-stage builds for production-ready images
- **Docker Compose** — Local development orchestration

---

## CI/CD Workflows

### 1. CI Pipeline (`ci.yml`)

**Trigger:** Push or pull request to `main`/`develop`

**Jobs:**

#### Frontend: Lint & Build

- Install dependencies with npm
- Run TypeScript type checking
- Run ESLint (optional, configurable)
- Build both Next.js apps
- Upload `.next` build artifacts

#### Backend: Lint & Build

- Install Go dependencies
- Run `go fmt` (formatting check)
- Run `go vet` (static analysis)
- Run `golangci-lint` (comprehensive linting)
- Build Go binary
- Upload binary to artifacts

#### Backend: Tests

- Run `go test -v -race -coverprofile=coverage.out ./...`
- Generate coverage report
- Upload coverage to Codecov (optional)

#### Security Scan

- Run `npm audit` (npm packages)
- Run Trivy vulnerability scanner (container image scanner)
- Upload SARIF results to GitHub Security tab

#### Integration Tests (Optional)

- Start backend service
- Run health check
- Create sample session via REST API
- Verify session retrieval
- Report success/failure

#### CI Status

- Aggregate results from all jobs
- Post summary comment on PR
- Fail if any critical job fails

**Output:**

- ✅ All checks pass → PR can be merged
- ❌ Any check fails → PR blocked until fixed

---

### 2. Build & Deploy (`build-deploy.yml`)

**Trigger:**

- Push to `main` branch (builds, no deploy)
- Git tag matching `v*.*.*` (builds, publishes release, deploys)
- Manual workflow dispatch

**Jobs:**

#### Build Backend Docker

- Multi-stage build (small final image)
- Install dependencies
- Compile Go binary
- Create Alpine Linux runtime image
- Push to container registry

**Output:** `ghcr.io/your-org/realtime-session-coordination/backend:main`

#### Build Frontend: User App

- Multi-stage build (Next.js standalone)
- Install dependencies
- Build Next.js app
- Create Node.js Alpine runtime
- Push to container registry

**Output:** `ghcr.io/your-org/realtime-session-coordination/frontend-user:main`

#### Build Frontend: Admin App

- Similar to user app
- Push to container registry

**Output:** `ghcr.io/your-org/realtime-session-coordination/frontend-admin:main`

#### Publish Release

- **Condition:** On git tag `v*.*.*`
- Auto-generate release notes from commits
- Create GitHub Release
- Mark as pre-release if `alpha`, `beta`, or `rc` in version

#### Deploy to Staging

- **Condition:** Push to `develop` branch
- Requires approval (environment protection rule)
- Deploy containers to staging environment
- Run smoke tests

#### Deploy to Production

- **Condition:** Git tag `v*.*.*`
- Requires approval (environment protection rule)
- Deploy containers to production environment
- Send Slack notification
- Create deployment record

---

## Docker

### Dockerfile Strategy

**Multi-stage builds** for optimal image sizes:

#### Backend (`infra/docker/backend/Dockerfile`)

```dockerfile
# Stage 1: Builder
FROM golang:1.25-alpine AS builder
# Download dependencies
# Compile binary

# Stage 2: Runtime
FROM alpine:latest
# Copy binary from builder
# Non-root user
# Health check
# CMD
```

**Size:** ~10-15 MB (vs ~800 MB with full Go image)

#### Frontend (`infra/docker/frontend/Dockerfile.user` and `.admin`)

```dockerfile
# Stage 1: Dependencies
FROM node:18-alpine AS deps
# Install pnpm
# Install all dependencies

# Stage 2: Builder
FROM node:18-alpine AS builder
# Copy dependencies from stage 1
# Build Next.js app

# Stage 3: Runtime
FROM node:18-alpine
# Copy built files from builder (standalone mode)
# Non-root user
# Health check
# CMD
```

**Size:** ~100-150 MB per app

### Building Images

```bash
# Backend
docker build -t realtime-backend:latest -f ./infra/docker/backend/Dockerfile ./backend

# User app
docker build -t realtime-user:latest -f ./infra/docker/frontend/Dockerfile.user ./frontend

# Admin app
docker build -t realtime-admin:latest -f ./infra/docker/frontend/Dockerfile.admin ./frontend

# Or use docker-compose (builds all)
docker-compose build
```

### Image Registry

**Default:** GitHub Container Registry (GHCR)

**Images pushed to:**

- `ghcr.io/your-org/realtime-session-coordination/backend:main`
- `ghcr.io/your-org/realtime-session-coordination/backend:v1.2.3`
- `ghcr.io/your-org/realtime-session-coordination/frontend-user:main`
- `ghcr.io/your-org/realtime-session-coordination/frontend-admin:main`

**To use different registry:**

Edit `.github/workflows/build-deploy.yml`:

```yaml
env:
  REGISTRY: docker.io  # Docker Hub
  # or
  REGISTRY: 123456789.dkr.ecr.us-east-1.amazonaws.com  # AWS ECR
```

---

## Deployment Strategies

### Local Development

```bash
docker-compose up --build
# Services available:
# - Backend: http://localhost:8080
# - User: http://localhost:3001
# - Admin: http://localhost:3002
```

### Staging Deployment

**Manual (via GitHub Actions):**

1. Push code to `develop` branch
2. GitHub Actions runs CI pipeline
3. If passing, wait for approval prompt
4. Approve deployment → deployed to staging
5. Verify at `https://staging.realtime.example.com`

**Kubernetes:**

```bash
kubectl apply -f k8s/staging/
```

**Docker Compose on VM:**

```bash
ssh deploy@staging.example.com
cd /opt/realtime
docker-compose pull
docker-compose up -d
```

### Production Deployment

**Trigger:** Create git tag with semantic version

```bash
git tag v1.0.0
git push --tags
```

**Workflow:**

1. GitHub Actions builds images
2. Creates GitHub Release with notes
3. Pushes images to registry
4. Waits for approval (environment protection)
5. Deploys to production
6. Sends Slack notification

**Kubernetes:**

```bash
# Edit k8s/production/ manifests with new image version
kubectl apply -f k8s/production/

# Or use helm
helm upgrade realtime ./helm/realtime --values values-prod.yaml
```

**Docker Compose on VM:**

```bash
ssh deploy@prod.example.com
cd /opt/realtime
# Update docker-compose.yml with new image tags
docker-compose pull
docker-compose up -d --no-build
```

---

## Environment Configuration

### GitHub Secrets

Add via **Settings → Secrets and variables → Actions**:

```
GITHUB_TOKEN              # Auto-provided
SLACK_WEBHOOK_URL         # Slack notifications
DOCKERHUB_USERNAME        # Alternative registry
DOCKERHUB_TOKEN           # Alternative registry
AWS_ACCESS_KEY_ID         # AWS ECR
AWS_SECRET_ACCESS_KEY     # AWS ECR
DEPLOY_KEY_STAGING        # SSH key for staging
DEPLOY_KEY_PRODUCTION     # SSH key for production
KUBE_CONFIG_PROD          # Kubernetes config
```

### Environment Variables

**Backend:**

```env
PORT=8080
GIN_MODE=release
DATABASE_URL=...          # Future: database
```

**Frontend:**

```env
NEXT_PUBLIC_BACKEND_BASE_URL=https://api.example.com
NEXT_PUBLIC_BACKEND_WS_URL=wss://api.example.com
```

---

## Production Checklist

Before deploying to production:

- [ ] All CI checks passing
- [ ] Code reviewed and approved
- [ ] Security scan passed
- [ ] Integration tests passed
- [ ] Database migrations tested (if applicable)
- [ ] Environment variables configured
- [ ] SSL/TLS certificates installed
- [ ] Monitoring and logging configured
- [ ] Backup strategy in place
- [ ] Rollback plan documented

---

## Monitoring & Observability

### Health Checks

**Backend:**

```bash
curl http://localhost:8080/healthz
# Response: { "status": "ok" }
```

**Frontend:**

```bash
curl http://localhost:3001/
```

### Logs

**Local (Docker Compose):**

```bash
docker-compose logs -f backend
docker-compose logs -f user-app
docker-compose logs -f admin-app
```

**Kubernetes:**

```bash
kubectl logs -f deployment/backend
kubectl logs -f deployment/frontend-user
```

**CloudWatch (AWS):**

```bash
aws logs tail /aws/ecs/realtime --follow
```

### Metrics

**Prometheus scrape targets** (optional):

- Backend: `http://localhost:8080/metrics`
- Frontend: `http://localhost:3001/metrics`

**Grafana dashboards** (optional):

- Request latency
- Error rates
- WebSocket connections
- Session lifecycle metrics

---

## Rollback Strategy

### Kubernetes Rollback

```bash
# View rollout history
kubectl rollout history deployment/backend

# Rollback to previous version
kubectl rollout undo deployment/backend

# Rollback to specific revision
kubectl rollout undo deployment/backend --to-revision=2
```

### Docker Compose Rollback

```bash
# Update image version in docker-compose.yml
docker-compose pull
docker-compose up -d

# Or use previous tag
docker-compose down
# Edit docker-compose.yml to use previous tag
docker-compose up -d
```

### Git Rollback

```bash
# Revert commit
git revert <commit-hash>
git push

# Create tag for rollback version
git tag v1.0.0-hotfix
git push --tags
# GitHub Actions deploys v1.0.0-hotfix
```

---

## Scaling Considerations

### Horizontal Scaling

**Multiple Backend Instances:**

- Use load balancer (e.g., nginx, AWS ALB)
- Replace in-memory session store with Redis
- Update WebSocket hub to use message broker

**Multiple Frontend Instances:**

- Stateless Next.js apps
- Use CDN for static assets
- Session storage in browser (no server state needed)

### Vertical Scaling

- Increase container resource limits
- Upgrade database tier (if added)
- Increase cache sizes (Redis, etc.)

### Auto-scaling

**Kubernetes:**

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: backend-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: backend
  minReplicas: 2
  maxReplicas: 10
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
```

---

## Security Best Practices

- [ ] Use non-root user in containers (UID 1000)
- [ ] Scan images for vulnerabilities (Trivy)
- [ ] Use private container registry or authentication
- [ ] Rotate secrets regularly
- [ ] Enable RBAC in Kubernetes
- [ ] Use network policies to restrict traffic
- [ ] Enable audit logging
- [ ] Use sealed secrets for sensitive data
- [ ] Enable TLS/HTTPS for all communication
- [ ] Implement rate limiting on APIs
- [ ] Log all authentication attempts

---

## Troubleshooting

### Workflow fails on build

1. Check logs: GitHub → Actions → select run → expand job
2. Common issues:
   - Out of memory → split jobs or use larger runner
   - Timeout → increase timeout or optimize build
   - Dependency issues → run locally to debug

### Docker image too large

- Use multi-stage builds
- Remove dev dependencies
- Use Alpine base images
- Prune unused layers

### Deployment fails

1. Verify environment secrets are set
2. Check deployment environment protection rules
3. Review SSH keys and credentials
4. Check Kubernetes cluster health (`kubectl cluster-info`)
5. Review resource limits

### Container won't start

1. Check image: `docker run -it image-name /bin/sh`
2. Check logs: `docker logs container-name`
3. Verify port mappings: `docker port container-name`
4. Check environment variables are set

---

## Next Steps

1. **Push to GitHub** — `git push origin main`
2. **Monitor workflow** — GitHub → Actions tab
3. **Configure registry** — Update `.github/workflows/build-deploy.yml` if using non-GHCR
4. **Setup environments** — GitHub → Settings → Environments
5. **Add secrets** — Add deployment credentials
6. **Deploy** — Create version tag to trigger production deployment

For more details, see:

- [Setup Guide](setup.md) — Local development
- [Architecture](architecture.md) — System design
- [API Documentation](api.md) — Endpoints
