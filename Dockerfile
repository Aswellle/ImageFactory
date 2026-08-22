# ImageForge — multi-stage build.
#
# Stages:
#   1. frontend: build the Vue SPA with pnpm -> frontend/dist
#   2. backend:  compile the Go binary, embedding dist/ via go:embed
#   3. runtime:  minimal final image
#
# Build args:
#   BUILD_FRONTEND=false  skip the frontend stage (backend-only/dev image)
#   GOLANG_IMAGE / ALPINE_IMAGE / NODE_IMAGE  pin base images

ARG GOLANG_IMAGE=golang:1.22-alpine
ARG ALPINE_IMAGE=alpine:3.21
ARG NODE_IMAGE=node:22-alpine

# =============================================================================
# Stage 1: Frontend build (Vue + Vite)
# =============================================================================
FROM ${NODE_IMAGE} AS frontend
WORKDIR /src/frontend

# Install pnpm via corepack (bundled with Node 22).
RUN corepack enable

# Layer-cache dependency install behind lockfile checksum.
COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

# Build static assets.
COPY frontend/ ./
RUN pnpm build

# =============================================================================
# Stage 2: Backend build (Go, CGO disabled -> static binary)
# =============================================================================
FROM ${GOLANG_IMAGE} AS builder

# Install git + CA certs so `go mod download` works against HTTPS module proxies.
RUN apk add --no-cache ca-certificates git

WORKDIR /src/backend

# Layer-cache module download.
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy backend source.
COPY backend/ ./
# Drop the placeholder web/dist so the embed picks up the real frontend build.
RUN rm -rf internal/web/dist
# Copy compiled frontend into the embed root (internal/web/dist).
COPY --from=frontend /src/frontend/dist ./internal/web/dist

# Static binary, trimmed for size. -s -w strip debug info; trimpath removes
# local FS paths from the binary.
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /imageforge \
    ./cmd/server

# =============================================================================
# Stage 3: Runtime — minimal image
# =============================================================================
FROM ${ALPINE_IMAGE} AS runtime

RUN apk add --no-cache ca-certificates tzdata postgresql16-client

# Unprivileged runtime user.
RUN addgroup -S iforge && adduser -S iforge -G iforge
USER iforge

WORKDIR /app
COPY --from=builder /imageforge /app/imageforge

EXPOSE 8080
ENTRYPOINT ["/app/imageforge"]
