# Infrastructure

This directory contains all infrastructure-related files for deployment and containerization.

## Structure

```
infra/
├── docker/
│   ├── backend/
│   │   ├── Dockerfile           # Multi-stage Go binary build
│   │   └── .dockerignore
│   ├── frontend/
│   │   ├── Dockerfile.user      # User app container image
│   │   ├── Dockerfile.admin     # Admin app container image
│   │   └── .dockerignore
│   └── README.md                # Docker-specific documentation (this file)
└── docker-compose.yml           # Local development orchestration
```

## Docker Images

### Backend (`docker/backend/Dockerfile`)

**Features:**

- Multi-stage build (small Alpine-based final image)
- Non-root user (uid 1000)
- Health check enabled
- ~10-15 MB final image size

**Build:**

```bash
docker build -t realtime-backend -f ./infra/docker/backend/Dockerfile ./backend
```

### Frontend: User App (`docker/frontend/Dockerfile.user`)

**Features:**

- Multi-stage Next.js standalone build
- Non-root user (uid 1000)
- Health check enabled
- Configurable via environment variables
- ~100-150 MB image size

**Build:**

```bash
docker build -t realtime-user -f ./infra/docker/frontend/Dockerfile.user ./frontend
```

**Environment Variables:**

- `NEXT_PUBLIC_BACKEND_BASE_URL` — Backend API endpoint (default: http://localhost:8080)

### Frontend: Admin App (`docker/frontend/Dockerfile.admin`)

**Features:**

- Multi-stage Next.js standalone build
- Non-root user (uid 1000)
- Health check enabled
- Configurable via environment variables
- ~100-150 MB image size

**Build:**

```bash
docker build -t realtime-admin -f ./infra/docker/frontend/Dockerfile.admin ./frontend
```

**Environment Variables:**

- `NEXT_PUBLIC_BACKEND_BASE_URL` — Backend API endpoint (default: http://localhost:8080)

---

## Docker Compose

### Local Development

The `docker-compose.yml` orchestrates all three services locally.

**Start services:**

```bash
# From root or infra/ directory
docker-compose up --build
```

**Available services:**

- Backend: http://localhost:8080
- User App: http://localhost:3001
- Admin App: http://localhost:3002

**Stop services:**

```bash
docker-compose down
```

**View logs:**

```bash
docker-compose logs -f backend
docker-compose logs -f user-app
docker-compose logs -f admin-app
```

**Rebuild specific service:**

```bash
docker-compose build backend
docker-compose up backend
```

---

## Image Registry

When deploying, images are pushed to a container registry:

### GitHub Container Registry (GHCR)

**Default registry:** `ghcr.io`

**Image naming:**

```
ghcr.io/{owner}/{repo}/backend:main
ghcr.io/{owner}/{repo}/frontend-user:main
ghcr.io/{owner}/{repo}/frontend-admin:main
```

**Configure in:** `.github/workflows/build-deploy.yml`

### Alternative Registries

**Docker Hub:**

```yaml
env:
  REGISTRY: docker.io
```

**AWS ECR:**

```yaml
env:
  REGISTRY: 123456789.dkr.ecr.us-east-1.amazonaws.com
```

---

## Security Notes

- All images run as non-root user (uid 1000)
- Using Alpine base images to minimize attack surface
- Health checks enabled for automatic restart on failure
- No secrets stored in Docker images
- Environment variables used for runtime configuration

---

## Performance

### Image Sizes

| Image     | Size       | Notes                   |
| --------- | ---------- | ----------------------- |
| Backend   | 10-15 MB   | Go binary + Alpine base |
| User App  | 100-150 MB | Node.js + Next.js       |
| Admin App | 100-150 MB | Node.js + Next.js       |

### Build Time Optimization

- Multi-stage builds reduce final image size
- Docker layer caching speeds up rebuilds
- GitHub Actions uses `type=gha` cache for faster CI builds

---

## Development Workflow

### Local Development (Without Docker)

```bash
make install
make backend  # Terminal 1
make frontend-user  # Terminal 2
make frontend-admin  # Terminal 3
```

### Local Development (With Docker)

```bash
docker-compose up --build
```

### Production Deployment

```bash
# Create git tag (triggers GitHub Actions)
git tag v1.0.0
git push --tags

# GitHub Actions:
# 1. Builds images
# 2. Pushes to registry
# 3. Creates release
# 4. Deploys (if configured)
```

---

## Troubleshooting

### Docker build fails

```bash
# Clear Docker cache
docker system prune
docker builder prune

# Rebuild without cache
docker-compose build --no-cache
```

### Port already in use

```bash
# Find process on port
lsof -i :8080

# Kill process
kill -9 <PID>

# Or use different port
docker run -p 8888:8080 realtime-backend
```

### Image too large

Check `.dockerignore` files:

```bash
cat ./infra/docker/backend/.dockerignore
cat ./infra/docker/frontend/.dockerignore
```

Add unnecessary files to reduce context size.

### Container won't start

```bash
# Check logs
docker logs <container-id>

# Run interactively to debug
docker run -it realtime-backend /bin/sh
```

---

For CI/CD details, see [../docs/cicd.md](../docs/cicd.md)  
For setup instructions, see [../docs/setup.md](../docs/setup.md)
