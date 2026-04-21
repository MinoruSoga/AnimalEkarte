---
description: Docker development standards (multi-stage build, security, best practices)
alwaysApply: false
globs: ["**/Dockerfile*", "docker-compose.yml", "**/.dockerignore"]
---

# Docker Rules

Docker Compose development environment standards.

## Core Rules

### 1. Multi-Stage Build (Required)

```dockerfile
# ❌ Inefficient: Image size 1.2GB
FROM golang:1.25
COPY . .
RUN go build -o app ./cmd/api
EXPOSE 8080
CMD ["./app"]

# ✅ Optimized: Image size 50MB
# Build stage
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/api

# Runtime stage
FROM alpine:3.18
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /app/app .
RUN addgroup -g 1000 appuser && adduser -D -u 1000 -G appuser appuser
USER appuser
EXPOSE 8080
CMD ["./app"]
```

### 2. Layer Caching Optimization

```dockerfile
# ❌ Inefficient: Code change rebuilds all layers
FROM golang:1.25-alpine
COPY . .                 # ← Code (most frequently changed)
RUN go mod download
RUN go build -o app ./cmd/api

# ✅ Optimized: Low-change dependencies first
FROM golang:1.25-alpine
COPY go.mod go.sum ./   # ← Dependencies (rarely changed)
RUN go mod download
COPY . .                # ← Code (last)
RUN go build -o app ./cmd/api
```

### 3. .dockerignore File

```
# backend/.dockerignore
.git
.gitignore
README.md
Makefile
*.local
.DS_Store
__pycache__
*.pyc
node_modules
dist
build
coverage
.env.local

# frontend/.dockerignore
.git
.gitignore
README.md
node_modules
dist
build
.env.local
.DS_Store
cypress
coverage
```

### 4. Non-root User (Security)

```dockerfile
# ✅ Dockerfile
FROM alpine:3.18

# Create user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

COPY --from=builder --chown=appuser:appuser /app/app .

# Run as non-root
USER appuser

EXPOSE 8080
CMD ["./app"]
```

### 5. Frontend Nginx Optimization

```dockerfile
# Build stage
FROM node:20-alpine AS builder
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build

# Runtime stage
FROM nginx:1.25-alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/nginx.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

### 6. Docker Compose Commands (Project Standard)

```bash
# ✅ Use Make commands
make up          # Start containers
make down        # Stop containers
make logs        # View logs
make clean       # Clear cache and rebuild

# ✅ Execute commands in containers (npm/go prohibited locally)
docker compose exec frontend npm run build
docker compose exec backend go test ./... -v
docker compose exec backend golangci-lint run ./...

# ❌ Prohibited local execution
npm run build
go test ./...
```

### 7. Security Scanning

```bash
# Scan with trivy
docker build -t ekarte-backend:latest .
trivy image ekarte-backend:latest

# Maintain secure state
# Critical 0, High 0
```

## Checklist

- [ ] Multi-stage build used
- [ ] Layer order optimized (go.mod/sum → code)
- [ ] .dockerignore excludes unnecessary files
- [ ] Non-root user implemented
- [ ] Image size < 200MB
- [ ] Security scan (trivy): Critical 0
- [ ] npm/go prohibited locally (Docker only)

## Performance Targets

```
Build Time:    < 2 minutes
Image Size:    Backend < 50MB, Frontend < 100MB
Startup Time:  < 5s
Security:      Vulnerability Critical/High 0
```
