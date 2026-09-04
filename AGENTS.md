# Repository Guidelines

## Project Overview

**ImageForge** is a commercial AI image-production workspace. It lets users create, edit, organize, manage, and programmatically access visual assets. The frontend is a Vue 3 SPA; the backend is a Go API server. ImageForge is built on top of a local **Sub2API** codebase (an upstream AI gateway), which it ports and adapts rather than calling as an external service.

> Product mission: "Create, edit, organize, manage commercialized visual assets, and provide capabilities via API."

The UI must be image-first, minimal, professional, and Apple-style. No neon aesthetics, no dense dashboards, no excessive decoration.

---

## Architecture & Data Flow

### Request lifecycle

```
Browser
  → ImageForge Web (Vue 3 SPA, Vite)
  → ImageForge API (Gin, /v1)
    → Auth (JWT middleware or API-key middleware)
    → Handler (parse + validate)
    → Service (business logic)
    → Repository (Ent ORM → PostgreSQL)
    → Batch-image pipeline (ported from Sub2API)
      → Provider (Gemini / OpenAI / etc.)
    → Job queue (in-memory now, Redis later)
    → Object storage (R2 / MinIO / filesystem)
    → Asset record
    → Browser
```

### Key architectural decisions

| Concern | Choice |
|---------|--------|
| Backend framework | Gin (Go 1.26) |
| ORM | Ent (schema in `ent/schema/`, generated code in `ent/`) |
| DI | Hand-written initializer in `cmd/server/wire.go` (Wire planned for Phase 2+) |
| Frontend | Vue 3 + Vite + Tailwind CSS 4 + Pinia + vue-router |
| Database | PostgreSQL 18 |
| Cache | Redis 8 (optional — falls back to in-memory) |
| Storage | S3-compatible via `storage.Storage` interface (R2, MinIO, or local filesystem) |
| Async jobs | `job.Queue` + `job.Processor` (in-memory `MemoryQueue` in Phase 1) |
| Config | Viper, env-driven with `IF_` prefix |
| Logging | zap (JSON in production, console in debug) |
| Auth | JWT (`golang-jwt/jwt/v5`) + bcrypt passwords; API keys for programmatic access |

### Sub2API relationship

Sub2API is the **base codebase**, not an external service. ImageForge reuses its source code:

- **Ent models**: import directly from `github.com/Wei-Shaw/sub2api/ent/...` via Go workspace
- **Service logic**: copy source files → adapt package name → replace imports → use ent types → compile + test
- **Patterns**: follow Sub2API's conventions (error codes, middleware chains, worker queues)

The Go workspace is at `D:\Git-Clone\SRC\Sub2API\go.work` and links:
- `github.com/Wei-Shaw/sub2api` → `D:\Git-Clone\SRC\Sub2API\sub2api\backend`
- `github.com/imageforge/imageforge` → this project's `backend/`

---

## Key Directories

```
ImageForge/
├── backend/                    # Go API server
│   ├── cmd/
│   │   ├── server/             # Main entrypoint (main.go) + DI wiring (wire.go)
│   │   └── admin-cli/          # CLI for admin tasks (promote/demote users)
│   ├── ent/
│   │   ├── schema/             # Ent entity definitions (edit these)
│   │   └── *.go                # Generated Ent code (run `go generate ./ent` after schema changes)
│   ├── internal/
│   │   ├── config/             # Viper config loading (env-driven, IF_ prefix)
│   │   ├── server/
│   │   │   ├── router.go       # Gin engine setup, dependency wiring, route registration
│   │   │   ├── middleware/      # Auth, API-key auth, CORS, request ID, max body
│   │   │   └── routes/         # Route grouping (admin routes)
│   │   ├── handler/            # HTTP handlers (auth, asset, generation, tag, etc.)
│   │   ├── service/            # Business logic (auth, JWT, password, asset, generation, usage, etc.)
│   │   ├── repository/         # Data access (Ent repositories, DB connector, rate limiter)
│   │   ├── job/                # Async job model, Queue interface, MemoryQueue, Processor
│   │   ├── batchimage/         # Ported Sub2API pipeline (provider interface, Gemini provider, public service)
│   │   ├── storage/            # Object storage (Storage interface, R2, filesystem)
│   │   ├── domain/             # Domain-wide constants (roles, statuses, error codes)
│   │   ├── web/                // Embedded frontend assets via go:embed
│   │   └── pkg/
│   │       ├── errors/         # Stable, provider-independent error codes
│   │       ├── response/       # Uniform JSON envelope helpers
│   │       ├── logger/         # zap logger factory
│   │       └── image/          # Image processing (thumbnail)
│   ├── integration/            # Integration + e2e tests (build tags `integration`, `e2e`)
│   └── migrations/             # SQL migration files
├── frontend/                   # Vue 3 SPA
│   ├── src/
│   │   ├── api/               # Axios client + per-domain API modules
│   │   ├── views/             # Page components (Dashboard, Login, Gallery, Generation, etc.)
│   │   ├── components/        # Shared UI components (Toast, Skeleton, TagBadge, etc.)
│   │   ├── stores/            # Pinia stores (auth, asset, generation, etc.)
│   │   ├── router/            # vue-router config + auth guard
│   │   ├── types/             # Shared TypeScript interfaces
│   │   ├── composables/       # Vue composables (useKeyboard, useDesktopLauncher)
│   │   ├── i18n/              # vue-i18n localization
│   │   ├── styles/            # Tailwind CSS entry (main.css)
│   │   └── utils/             # Utilities (debounce, password rules)
│   └── public/                 # Static assets
├── deploy/                     # Docker Compose + Caddyfile
├── tests/                      # Python Playwright tests
│   ├── test_webapp.py          # Standalone web UI tests
│   └── web/                    # pytest + Playwright tests (auth, landing, views)
│   └── e2e/                    # End-to-end tests with screenshots
├── docs/                       # PRD (Chinese), RULES.md
├── Dockerfile                  # Multi-stage build (frontend → Go binary → alpine runtime)
└── Makefile                    # Dev shortcuts
```

---

## Development Commands

### Backend

```bash
cd backend

# Run the server directly
go run ./cmd/server

# Build the binary
go build -o ../build/imageforge ./cmd/server

# Run unit tests (build tag `unit`)
go test -tags=unit ./...

# Run all tests
go test ./...

# Regenerate Ent code after schema changes
go generate ./ent

# Lint
golangci-lint run ./...
```

### Frontend

```bash
cd frontend

# Install deps
pnpm install

# Run dev server (proxies /v1 to backend at 127.0.0.1:8080)
pnpm dev

# Type-check
pnpm typecheck

# Build for production
pnpm build

# Lint
pnpm lint
```

### Combined (Makefile)

```bash
make backend              # Build backend binary
make backend-generate     # Regenerate Ent code
make backend-lint         # Lint backend
make backend-test         # Run backend unit tests
make frontend             # Install + build frontend
make frontend-dev         # Run frontend dev server
make frontend-typecheck   # Type-check frontend
make docker-up            # Start full stack via Docker Compose
make docker-down          # Stop Docker stack
make docker-logs          # Tail backend logs
```

### Docker

```bash
# Full stack
docker compose -f deploy/docker-compose.yml up --build

# Infrastructure only (db, redis, minio)
docker compose -f deploy/docker-compose.yml up -d db redis minio
```

### CLI admin tool

```bash
cd backend
go run ./cmd/admin-cli promote <email>    # Promote user to admin
go run ./cmd/admin-cli demote <email>     # Demote admin to user
go run ./cmd/admin-cli list-admins        # List all admins
```

---

## Code Conventions & Common Patterns

### Backend layering

The backend follows a strict layered architecture. **Never** let a handler bypass the service layer or let a service write HTTP responses.

```
Handler (HTTP) → Service (business logic) → Repository (data access) → Ent (ORM) → PostgreSQL
```

### Handler pattern

Handlers are thin: bind JSON, extract user from context, call service, write response.

```go
func (h *XHandler) Create(c *gin.Context) {
    var req requestRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, 400, string(errors.ErrorCodeInvalidRequest), "invalid request", requestID(c))
        return
    }
    uid, ok := userID(c)
    if !ok {
        response.Error(c, http.StatusUnauthorized, string(errors.ErrorCodeUnauthorized), "authentication required", requestID(c))
        return
    }
    result, err := h.svc.Create(c.Request.Context(), service.CreateInput{...})
    if err != nil {
        writeError(c, err)   // maps *errors.Error to HTTP response
        return
    }
    response.Created(c, result)
}
```

Key conventions:
- Request structs use `binding:` tags for validation (e.g., `binding:"required,email"`, `binding:"required,min=8"`)
- Extract user ID via `userID(c)` helper (from `if.user_id` context key set by auth middleware)
- Extract request ID via `requestID(c)` helper (from `if.request_id` context key)
- Use `response.OK`, `response.Created`, `response.Paginated`, `response.Error` for all responses
- Use `writeError(c, err)` to map `*errors.Error` to HTTP responses

### Service pattern

Services are structs with constructor functions. They never touch HTTP or Gin.

```go
type AssetService struct {
    db    *ent.Client
    store storage.Storage
}

func NewAssetService(db *ent.Client, store storage.Storage) *AssetService {
    return &AssetService{db: db, store: store}
}
```

- Business logic that spans multiple entities should use transactions (Ent's `Tx`)
- Services depend on interfaces (e.g., `storage.Storage`) not concrete types
- Services return domain errors (`*errors.Error`) not raw upstream errors

### Repository pattern

Repositories wrap Ent operations. They expose domain-meaningful methods, not raw Ent builders.

```go
type UserRepository struct {
    db *ent.Client
}

func NewUserRepository(db *ent.Client) *UserRepository {
    return &UserRepository{db: db}
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*ent.User, error) {
    return r.db.User.Query().Where(user.Email(email)).Only(ctx)
}
```

### Error handling

All API errors use stable, provider-independent codes defined in `internal/pkg/errors`:

```go
// Error codes (partial list)
ErrorCodeInvalidRequest      = "INVALID_REQUEST"           // 400
ErrorCodeUnauthorized        = "UNAUTHORIZED"              // 401
ErrorCodeForbidden           = "FORBIDDEN"                 // 403
ErrorCodeNotFound            = "NOT_FOUND"                 // 404
ErrorCodeConflict            = "CONFLICT"                  // 409
ErrorCodeImageGeneration     = "IMAGE_GENERATION_FAILED"   // 500
ErrorCodeImageEdit           = "IMAGE_EDIT_FAILED"         // 500
ErrorCodeUpstreamTimeout     = "UPSTREAM_TIMEOUT"          // 504
ErrorCodeUpstreamRateLimited = "UPSTREAM_RATE_LIMITED"     // 429
ErrorCodeModelUnavailable    = "MODEL_UNAVAILABLE"         // 503
ErrorCodeAuthExpired         = "AUTHENTICATION_EXPIRED"    // 401
ErrorCodeStorageFailed       = "STORAGE_FAILED"            // 500
ErrorCodeQuotaExceeded       = "QUOTA_EXCEEDED"            // 429
ErrorCodeInternal            = "INTERNAL_ERROR"            // 500
```

Rules:
- Never expose raw upstream errors (OpenAI, Gemini, etc.) to clients
- Use `errors.New(code, msg)` to create; `errors.Wrap(code, msg, cause)` to wrap
- Map internal errors to these stable codes at the handler layer via `writeError`

### Response envelope

All JSON responses use a uniform envelope:

```json
// Success
{ "data": { ... } }

// Created (201)
{ "data": { ... } }

// Error
{ "error": { "code": "NOT_FOUND", "message": "asset not found", "request_id": "req_xxx" } }

// Paginated
{ "data": [ ... ], "pagination": { "total": 100, "page": 1, "page_size": 20 } }
```

### Ent schema conventions

- Schemas live in `ent/schema/`; generated code in `ent/`
- Use `field.Enum(...)` for status fields with `.Values(...)` and `.Default(...)`
- Sensitive fields use `.Sensitive()` (e.g., `password_hash`)
- Edges use `edge.To(...)` / `edge.From(...).Ref(...)` with `.Unique()` and `.Required()` where appropriate
- Add indexes via `Indexes()` for query patterns
- After schema changes: run `go generate ./ent` to regenerate code

### Middleware

Global middleware order (set in `router.go`):
1. `RequestID()` — attaches `req_<uuid>` or reuses `X-Request-ID` header
2. `CORS()` — permits common dev origins
3. `MaxRequestBodySize(1 << 20)` — 1 MB limit
4. `gin.Recovery()` — panic recovery
5. `requestLogger(log)` — logs method, path, status, duration, request ID

Route groups apply additional middleware:
- `authMW.Require()` — JWT required
- `authMW.RequireAdmin()` — JWT + admin role required
- `apiKeyMW` — API key required (supports `x-api-key` header and `Authorization: Bearer <key>`)
- `adminAuth` — combined admin auth for admin route group

Context keys set by middleware:
- `if.user_id` — authenticated user ID (int64)
- `if.role` — user role ("user" | "admin")
- `if.request_id` — request correlation ID
- `if.auth_method` — "jwt" or "api_key"

### Naming conventions

- **Packages**: short, lowercase, singular (e.g., `service`, `handler`, `repository`, `storage`)
- **Constructors**: `New<X>(...) *<X>` (e.g., `NewAuthService`, `NewAssetHandler`)
- **Interfaces**: describe behavior (e.g., `Storage`, `Queue`, `Processor`); often with `var _ Interface = (*Impl)(nil)` compile-time check
- **Request/response types**: `<action>Request`, `<action>Result`, `<action>Input`
- **Files**: one handler/service per file, named after the entity (e.g., `asset_handler.go`, `auth_service.go`)
- **Ent schemas**: singular nouns (`User`, `Asset`, `GenerationJob`)

### Logging

- Use `go.uber.org/zap`; never log secrets (API keys, tokens, passwords, signed URLs, cookies, auth headers)
- Use `request_id` for correlation
- In debug mode: colored console output; in production: JSON with ISO 8601 timestamps

### Storage path layout

```
users/{user_id}/projects/{project_id}/{kind}/{name}
```

Where `kind` ∈ {`originals`, `thumbnails`, `mediums`, `versions`, `uploads`}.

### Async job pattern

Every image operation is an async job:

```go
// job.go
type Job struct {
    ID, UserID, ProjectID, Type, Status, Model, Prompt string/int
    Attempt, MaxAttempts int
    Sub2APITaskID string
    CreatedAt, ScheduledAt time.Time
    InputAssetIDs []int64
}

type Task struct {
    Job *Job
    Run func(ctx context.Context) error
}

type Queue interface {
    Submit(ctx context.Context, t *Task) error
    Stop() error
}
```

Job statuses: `pending`, `processing`, `completed`, `failed`, `cancelled`.

---

## Frontend Conventions

### State management

Pinia stores use the Composition API pattern:

```typescript
export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(getToken())
  const user = ref<User | null>(null)
  const isAuthenticated = computed(() => !!token.value)
  // ... actions
  return { token, user, isAuthenticated, login, logout }
})
```

### API client

- Single axios instance at `src/api/client.ts`
- Request interceptor attaches Bearer token + `X-Request-ID`
- Response interceptor surfaces 401/403/network errors as toasts
- Use `parseApiError(err)` to extract `{ code, message, requestID }` from failures

### Routing

- Lazy-loaded views keep the initial bundle small
- Auth guard enforces `requiresAuth` and `requiresAdmin` meta fields
- Authenticated users are redirected from landing to dashboard

### Path alias

`@/` maps to `src/` (configured in `vite.config.ts`).

### Dev proxy

Vite dev server proxies `/v1` to the backend at `http://127.0.0.1:8080` (override with `IF_API_TARGET`).

---

## Important Files

| File | Purpose |
|------|---------|
| `backend/cmd/server/main.go` | Application entrypoint, graceful shutdown |
| `backend/cmd/server/wire.go` | Dependency injection (hand-written; Wire later) |
| `backend/internal/server/router.go` | Gin engine setup, all route registration |
| `backend/internal/config/config.go` | Viper config loading, `Config` struct |
| `backend/internal/pkg/errors/errors.go` | Stable error codes, `Error` type |
| `backend/internal/pkg/response/response.go` | Uniform JSON envelope helpers |
| `backend/internal/domain/constants.go` | Domain constants (roles, statuses, error codes) |
| `backend/internal/web/web.go` | `go:embed` of frontend `dist/` |
| `backend/ent/schema/*.go` | Ent entity definitions (source of truth for models) |
| `backend/migrations/*.sql` | Versioned SQL migrations |
| `backend/go.mod` | Go module definition |
| `frontend/src/main.ts` | Vue app bootstrap |
| `frontend/src/router/index.ts` | Route definitions + auth guard |
| `frontend/src/api/client.ts` | Shared axios instance + interceptors |
| `frontend/src/types/index.ts` | Shared TypeScript interfaces |
| `frontend/vite.config.ts` | Vite config + dev proxy |
| `Dockerfile` | Multi-stage build (frontend → Go → alpine) |
| `deploy/docker-compose.yml` | Local infrastructure stack |
| `deploy/Caddyfile` | Production reverse proxy config |
| `.env.example` | Environment variable template |
| `Makefile` | Development shortcuts |

---

## Runtime/Tooling Preferences

| Concern | Requirement |
|---------|-------------|
| Go | 1.26+ |
| Node | 22+ |
| Package manager (frontend) | pnpm 10+ (corepack enabled in Dockerfile) |
| Database driver | pgx/v5 (stdlib) |
| ORM | Ent v0.14.4 |
| Linting (backend) | golangci-lint |
| Linting (frontend) | eslint |
| Type-checking (frontend) | vue-tsc |

> The `backend/go.sh` helper sets `GOROOT=/c/GoInstall/go`, `GOFLAGS=-mod=mod`, `GOSUMDB=off` for the local Windows Go install.

---

## Testing & QA

### Test layers

| Layer | Location | Build tag | Purpose |
|-------|----------|-----------|---------|
| Unit tests | `backend/` (alongside source) | `unit` | Test services/handlers in isolation |
| Integration tests | `backend/integration/` | `integration` | In-process HTTP tests against real DB |
| E2E tests | `backend/integration/` | `e2e` | Full request lifecycle (register → login → API key) |
| Web UI tests | `tests/web/` | — | Playwright tests against running dev server |
| Legacy web tests | `tests/test_webapp.py` | — | Standalone Playwright script |
| E2E UX tests | `tests/e2e/` | — | Full-app Playwright tests with screenshots |

### Running tests

```bash
# Backend unit tests
cd backend && go test -tags=unit ./...

# Backend integration tests (needs DB)
cd backend && go test -tags=integration ./...

# Backend e2e tests (needs DB + running infra)
cd backend && go test -tags=e2e ./...

# Frontend web UI tests (needs running frontend dev server)
cd tests/web && pytest

# Legacy web tests
python tests/test_webapp.py
```

### Integration test harness

`backend/integration/main_test.go` provides:
- `TestMain` that waits for infrastructure, runs migrations, boots the server
- `doRequest()` helper for in-process HTTP requests
- `mustSeedUser()` to provision authenticated test users
- Graceful skip when database is unavailable (so CI without Docker passes)

### Critical test paths for image generation

Must cover: success, timeout, 429, 500, 502, 503, invalid response, provider failure, storage failure.

---

## API Design Rules

- All public endpoints live under `/v1`
- API is provider-independent: never expose `provider_account_id`, `sub2api_account_id`, or Sub2API internals
- Use provider-neutral identifiers: `model`, `generation_id`, `task_id`
- Two auth paths: JWT (browser sessions) and API keys (programmatic access)
- OpenAI-compatible endpoint at `POST /v1/images/generations` for SDK access
- API-key-authed generation at `POST /v1/images/generate` (avoids conflict with public endpoint)
- Admin routes under `/v1/admin` with admin-only middleware

---

## Security Rules

- All protected resources verify: **authentication + authorization + resource ownership**
- Never trust client-submitted IDs; always resolve ownership server-side
- Never use client `user_id` for authorization — use the ID from the validated JWT/API key
- API keys are stored as SHA-256 hashes; plaintext shown only once at creation
- Passwords hashed with bcrypt (cost configurable, capped at 20)
- JWT secret must be set in production mode; debug mode auto-generates a temporary secret
- Never commit: API keys, OAuth tokens, refresh tokens, cookies, session secrets, production credentials
- Logs must never contain: API keys, tokens, cookies, auth headers, private image data, full signed URLs
- Token versioning invalidates all existing tokens on password change

---

## Image & Storage Rules

- Never store image binaries in the database — only keys, MIME type, dimensions, size, checksum, metadata
- Use object storage (S3/MinIO/R2) via the `storage.Storage` interface
- Private assets use presigned URLs
- Thumbnails generated asynchronously
- Editing an image creates a new version; never overwrite the original
- Storage path layout: `users/{user_id}/projects/{project_id}/{kind}/{name}`

---

## Porting Code from Sub2API

When porting Sub2API code, follow this exact sequence:

1. **Copy** the source file into the corresponding ImageForge package directory
2. **Rename** the package declaration to the ImageForge target package
3. **Replace imports**: `github.com/Wei-Shaw/sub2api/internal/...` → ImageForge local packages or direct `ent/` imports
4. **Remove internal dependencies**: replace Sub2API `internal/` types with local stubs or ent types
5. **Compile**: `go build ./...` must pass
6. **Test**: `go test ./...` must pass

> Always check Sub2API first before writing new code. **REUSE > REWRITE**, **PORT > REIMPLEMENT**.

---

## Decision Hierarchy

When facing a choice:

1. **Existing implementation reusable?** → REUSE (copy from Sub2API)
2. **Unclear what exists?** → INSPECT (read code) > GUESS
3. **Requirements ambiguous?** → ASK > ASSUME
4. **Large change scope?** → MINIMIZE > REFACTOR
5. **Security concern?** → STOP > SHIP
6. **Sub2API has it?** → PORT > REIMPLEMENT

---

## Non-Negotiable Stop Conditions

Stop and reassess immediately when:

1. Starting to reimplement something Sub2API already has
2. Treating Sub2API as an external HTTP service instead of a codebase to port from
3. Modifying Sub2API source files (only copy and adapt)
4. Implementing OAuth/upstream account scheduler/provider protocol from scratch
5. Exposing API keys or refresh tokens to the frontend
6. Storing image binaries in the database
7. Making image generation a permanent synchronous HTTP request
8. Adding excessive gradients, glassmorphism, or decoration
9. Creating abstraction layers for hypothetical future needs
10. Modifying files unrelated to the current task
11. Starting large-scale refactoring
12. Assuming APIs or database fields exist without verifying
13. Bypassing tests
14. Returning raw upstream errors to users
15. Ignoring the Go workspace's ability to import ent directly

---

## Documentation

- `docs/项目总体定义prd.md` — Product requirements (Chinese)
- `docs/RULES.md` — Non-negotiable stop conditions (Chinese)
- `README.md` — Stack overview, quick start, repository layout
