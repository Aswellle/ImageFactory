# ImageForge

A commercial AI image-production workspace. Create, edit, organize, manage,
and programmatically access visual assets.

ImageForge is **product** — a polished, image-first creative workspace.
Sub2API is **infrastructure** — the upstream AI gateway that handles provider
accounts, routing, auth, and failover. ImageForge never exposes Sub2API
credentials to browsers.

---

## Stack

| Layer    | Choice                                           |
|----------|--------------------------------------------------|
| Backend  | Go 1.26, Gin, Ent (ORM), Wire (DI)              |
| Frontend | Vue 3, Vite, Tailwind CSS, Pinia, vue-router    |
| Database | PostgreSQL 18                                    |
| Cache    | Redis 8                                          |
| Storage  | S3-compatible (Cloudflare R2 / MinIO)            |
| Gateway  | Sub2API (upstream AI provider routing + auth)    |

## Repository layout

```
ImageForge/
├── backend/                # Go API server
│   ├── cmd/server/         # entrypoint + DI wiring
│   ├── ent/schema/         # Ent entity definitions
│   ├── internal/
│   │   ├── config/         # env-driven config (Viper)
│   │   ├── server/         # router + Gin middleware
│   │   ├── handler/        # HTTP handlers
│   │   ├── service/        # business logic (auth, JWT, password)
│   │   ├── repository/     # data access (Ent) + DB connector
│   │   ├── job/            # async image job model, queue, worker
│   │   ├── storage/        # object storage (R2 / MinIO / filesystem)
│   │   ├── domain/         # domain constants
│   │   └── pkg/            # errors, logger, response, storage interface
│   └── migrations/         # SQL migrations
├── frontend/               # Vue 3 SPA
├── deploy/                 # Docker Compose (db, redis, minio, app)
├── Dockerfile              # multi-stage backend build
├── Makefile                # dev shortcuts
└── docs/                   # PRD, engineering rules, integration map
```

## Quick start

### Prerequisites

- Go 1.26+
- Node 22+ and pnpm 10+
- Docker + Docker Compose (for the full local stack)

### Local development

```bash
# 1. Install frontend deps
make frontend-install

# 2. Configure environment
cp .env.example .env

# 3. Start infrastructure (Postgres, Redis, MinIO)
docker compose -f deploy/docker-compose.yml up -d db redis minio

# 4. Run the backend (in one terminal)
cd backend && go run ./cmd/server

# 5. Run the frontend (in another terminal)
make frontend-dev
```

Then open http://127.0.0.1:5173.

### Full Docker stack

```bash
docker compose -f deploy/docker-compose.yml up --build
```

## Production deployment

ImageForge ships a production-ready Docker Compose stack with Caddy reverse
proxy, automatic HTTPS, isolated networks, and resource limits.

### Prerequisites

- Docker 24+ and Docker Compose v2
- A domain pointing to your server (for Let's Encrypt)
- At least 2 GB RAM, 2 CPU cores

### Steps

```bash
# 1. Clone the repository
git clone <repo-url> && cd ImageForge

# 2. Create production environment file
cp .env.example .env

# 3. Edit .env and set required values:
#    - IF_AUTH_JWT_SECRET (long random string)
#    - DB_PASSWORD / IF_DATABASE_PASSWORD
#    - IF_CADDY_HOST (your domain, e.g. imageforge.example.com)
#    - IF_ALLOWED_ORIGINS (your frontend origin, e.g. https://imageforge.example.com)
#    - STORAGE credentials (R2 or MinIO)

# 4. Start the production stack
docker compose -f deploy/docker-compose.prod.yml up -d

# 5. Verify health
curl https://your-domain/v1/health
curl https://your-domain/v1/health?deep=true
```

### Production architecture

```
Internet → Caddy (80/443, auto-HTTPS) → backend:8080
                                              ↓
                                    ┌─────────┼─────────┐
                                    ↓         ↓         ↓
                                 Postgres   Redis    R2/MinIO
```

### Key production features

- **Automatic HTTPS**: Caddy obtains and renews Let's Encrypt certificates
- **Isolated networks**: `backend_net` is internal-only (no outbound)
- **Resource limits**: CPU/memory limits on all services
- **Graceful shutdown**: SIGTERM handling with 15s drain timeout
- **Deep health checks**: `/v1/health?deep=true` verifies DB/Redis connectivity
- **Auto-migration**: Ent schema migration runs on startup (configurable)
- **CORS**: Configurable via `IF_ALLOWED_ORIGINS`
- **Trusted proxies**: Configure `IF_SERVER_TRUSTED_PROXIES` for correct client IPs

### Required environment variables

| Variable | Purpose | Required |
|----------|---------|----------|
| `IF_AUTH_JWT_SECRET` | JWT signing secret | Yes |
| `DB_PASSWORD` | PostgreSQL password | Yes |
| `JWT_SECRET` | Alias for JWT secret (compose) | Yes |
| `IF_CADDY_HOST` | Public domain for HTTPS | Yes |
| `IF_ALLOWED_ORIGINS` | CORS allowed origins | Recommended |
| `STORAGE_SECRET_KEY` | R2/MinIO secret key | Yes |


All configuration is environment-driven (see `.env.example`). The `IF_` prefix
is used throughout. Secrets (`IF_AUTH_JWT_SECRET`, `IF_DATABASE_PASSWORD`,
storage keys, `IF_SUB2API_ADMIN_API_KEY`) must never be committed.

## API

The public ImageForge API is provider-independent. It lives under `/v1`:

| Method | Path                  | Auth | Purpose              |
|--------|-----------------------|------|----------------------|
| POST   | /v1/auth/register     | —    | Create account       |
| POST   | /v1/auth/login        | —    | Sign in              |
| GET    | /v1/auth/me           | JWT  | Current user         |
| GET    | /v1/health            | —    | Liveness check       |

Generation endpoints (`/v1/images/*`, `/v1/projects/*`, `/v1/assets/*`,
`/v1/api-keys/*`) are added in Phase 2+.

## Architecture

```
Browser → ImageForge Web → ImageForge API → AuthZ → Generation Job
                                                       ↓
                                                  Sub2API → Upstream
                                                       ↓
                                                  Worker → Storage → Asset → UI
```

ImageForge owns UX, auth, projects, assets, jobs, storage, API keys, and
metadata. Sub2API owns upstream accounts, model routing, provider protocols,
failover, and subscription scheduling. Browsers never touch Sub2API directly.

## Documentation

- `docs/项目总体定义prd.md` — product requirements (Chinese)
- `AGENTS.md` — engineering rules and development workflow
- `docs/RULES.md` — non-negotiable stop conditions
- `docs/IMAGEFORGE_INTEGRATION_MAP.md` — Sub2API integration map (Phase 0 output)
