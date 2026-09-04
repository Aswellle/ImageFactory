<div align="center">

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="frontend/public/logo.svg">
  <source media="(prefers-color-scheme: light)" srcset="frontend/public/logo.svg">
  <img alt="ImageForge" src="frontend/public/logo.svg" width="120" height="120">
</picture>

# ImageForge

**Commercial AI Image-Production Workspace**

[中文](README.md) | [English](README_EN.md)

[![Go Version](https://img.shields.io/badge/Go-1.26-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![Vue Version](https://img.shields.io/badge/Vue-3.5-4FC08D?style=flat-square&logo=vuedotjs)](https://vuejs.org/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.7-3178C6?style=flat-square&logo=typescript)](https://www.typescriptlang.org/)
[![Tailwind CSS](https://img.shields.io/badge/Tailwind-4-06B6D4?style=flat-square&logo=tailwindcss)](https://tailwindcss.com/)
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=flat-square&logo=docker)](https://www.docker.com/)
[![License](https://img.shields.io/badge/License-Apache%202.0-2EC5FF?style=flat-square)](LICENSE)

---

> Create, edit, organize, manage commercialized visual assets, and provide capabilities via API.

</div>

---

## 📸 Screenshots

| Dashboard | Gallery | Generation |
|:---:|:---:|:---:|
| ![Dashboard](docs/screenshots/dashboard.png) | ![Gallery](docs/screenshots/gallery.png) | ![Generation](docs/screenshots/generation.png) |

> Note: Save screenshots to `docs/screenshots/` to auto-display.

---

## ✨ Features

### 🎨 Image Production
- **AI Image Generation** — Integrate with Gemini, OpenAI and other providers via Sub2API gateway
- **Image Editing** — Edit and version-control your visual assets
- **Async Job Queue** — Background processing with real-time status tracking
- **OpenAI-Compatible API** — Drop-in compatible with OpenAI SDK for image generation

### 🗂 Asset Management
- **Projects & Collections** — Organize assets into hierarchical structures
- **Version Control** — Non-destructive editing with full version history
- **Tags & Favorites** — Flexible categorization and quick access
- **Prompt Templates** — Reusable generation prompts with variables

### 🔐 Enterprise Security
- **JWT + API Key Auth** — Dual authentication paths for users and programs
- **Role-Based Access** — User / Admin roles with middleware enforcement
- **Resource Ownership** — Server-side ownership verification on every request
- **Rate Limiting** — Redis-backed distributed rate limiting

### 🚀 Production Ready
- **Docker Compose** — One-command deployment with Caddy auto-HTTPS
- **Graceful Shutdown** — Zero-downtime restarts with 15s drain timeout
- **Health Checks** — Liveness + deep readiness probes (DB/Redis/storage)
- **Auto Migration** — Ent schema migration on startup

---

## 🏗 Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Browser / Client                      │
└──────────────────────────┬──────────────────────────────────┘
                           │ HTTPS
                           ▼
┌──────────────────────────────────────────────────────────────┐
│                    Caddy (80/443, Auto-HTTPS)                 │
│                    Reverse Proxy + Security Headers          │
└──────────────────────────┬───────────────────────────────────┘
                           │
                           ▼
┌──────────────────────────────────────────────────────────────┐
│                   ImageForge Backend (Go + Gin)              │
│  ┌─────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────────┐ │
│  │  Auth   │ │  Asset   │ │Generation│ │   Job Queue      │ │
│  │ Service │ │ Service  │ │ Service  │ │   Processor      │ │
│  └────┬────┘ └────┬─────┘ └────┬─────┘ └────────┬─────────┘ │
│       │           │            │                │           │
│  ┌────┴───────────┴────────────┴────────────────┴─────────┐ │
│  │                    Ent ORM (PostgreSQL)                 │ │
│  └────────────────────────────────────────────────────────┘ │
└──────────────────────────┬───────────────────────────────────┘
                           │
          ┌────────────────┼────────────────┐
          ▼                ▼                ▼
     ┌─────────┐     ┌─────────┐     ┌──────────┐
     │PostgreSQL│     │  Redis  │     │ R2/MinIO │
     └─────────┘     └─────────┘     └──────────┘
                                             │
                                             ▼
                                    ┌───────────────┐
                                    │   Sub2API     │
                                    │   Gateway     │
                                    └───────┬───────┘
                                            │
                              ┌─────────────┼─────────────┐
                              ▼             ▼             ▼
                           Gemini       OpenAI        Others
```

---

## 🛠 Tech Stack

| Layer | Technology |
|-------|------------|
| **Backend** | Go 1.26 · Gin · Ent ORM · Viper · Zap |
| **Frontend** | Vue 3 · Vite · Tailwind CSS 4 · Pinia · Vue Router |
| **Database** | PostgreSQL 18 · Redis 8 |
| **Storage** | Cloudflare R2 · MinIO · S3-Compatible |
| **Infrastructure** | Docker Compose · Caddy · Let's Encrypt |

---

## 🚀 Quick Start

### Prerequisites

- [Docker](https://www.docker.com/) 24+ & Docker Compose v2
- [Go](https://golang.org/) 1.26+ (local development)
- [Node](https://nodejs.org/) 22+ & pnpm 10+ (local development)

### Docker Compose (Recommended)

```bash
# Clone
git clone https://github.com/Aswellle/ImageFactory.git
cd ImageForge

# Configure
cp .env.example .env
# Edit .env — set JWT_SECRET, DB_PASSWORD, etc.

# Start
docker compose -f deploy/docker-compose.prod.yml up -d

# Verify
curl http://localhost/v1/health
curl http://localhost/v1/health?deep=true
```

### Local Development

```bash
# 1. Install frontend deps
make frontend-install

# 2. Configure environment
cp .env.example .env

# 3. Start infrastructure
docker compose -f deploy/docker-compose.yml up -d db redis minio

# 4. Run backend (terminal 1)
cd backend && go run ./cmd/server

# 5. Run frontend (terminal 2)
make frontend-dev
```

Then open http://127.0.0.1:5173

---

## 📦 Production Deployment

```bash
# 1. Clone
git clone https://github.com/Aswellle/ImageFactory.git && cd ImageForge

# 2. Configure (REQUIRED variables)
cp .env.example .env
#   IF_AUTH_JWT_SECRET    — JWT signing secret (long random string)
#   DB_PASSWORD           — Database password
#   JWT_SECRET            — Same as above
#   IF_CADDY_HOST         — Your domain (e.g. imageforge.example.com)
#   IF_ALLOWED_ORIGINS    — Frontend origin (e.g. https://imageforge.example.com)
#   STORAGE_SECRET_KEY    — R2/MinIO secret key

# 3. Deploy
docker compose -f deploy/docker-compose.prod.yml up -d

# 4. Verify
curl https://your-domain/v1/health
```

---

## 📡 API Reference

All endpoints live under `/v1`. The API is **provider-independent**.

### Authentication

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/v1/auth/register` | — | Create account |
| POST | `/v1/auth/login` | — | Sign in |
| POST | `/v1/auth/send-reset-code` | — | Send reset code |
| POST | `/v1/auth/reset-password` | — | Reset password |
| GET | `/v1/auth/me` | JWT | Current user |

### Image Generation

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/v1/images/generations` | Public | OpenAI-compatible generation |
| POST | `/v1/images/generate` | API Key | Programmatic generation |
| POST | `/v1/images/edits` | JWT | Image editing |
| GET | `/v1/images/jobs` | JWT | List jobs |
| GET | `/v1/images/jobs/:id` | JWT | Job detail |
| POST | `/v1/images/tasks` | JWT | Submit async task |
| GET | `/v1/images/tasks/:id` | JWT | Task status |

### Projects & Assets

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/v1/projects` | JWT | Create project |
| GET | `/v1/projects` | JWT | List projects |
| GET | `/v1/projects/:id` | JWT | Project detail |
| GET | `/v1/assets` | JWT | List assets |
| GET | `/v1/assets/:id` | JWT | Asset detail |
| DELETE | `/v1/assets/:id` | JWT | Delete asset |
| GET | `/v1/assets/:id/content` | JWT | Asset content |
| GET | `/v1/assets/:id/versions` | JWT | Version history |

### Prompt Templates

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/v1/prompt-templates` | JWT | Create template |
| GET | `/v1/prompt-templates` | JWT | List templates |
| GET | `/v1/prompt-templates/:id` | JWT | Template detail |
| PUT | `/v1/prompt-templates/:id` | JWT | Update template |
| DELETE | `/v1/prompt-templates/:id` | JWT | Delete template |
| POST | `/v1/prompt-templates/:id/apply` | JWT | Apply template |

### API Keys & Usage

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/v1/api-keys` | JWT | Create API key |
| GET | `/v1/api-keys` | JWT | List API keys |
| DELETE | `/v1/api-keys/:id` | JWT | Revoke API key |
| GET | `/v1/usage` | JWT | Usage stats |
| GET | `/v1/usage/history` | JWT | Usage history |

### Admin

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/v1/admin/dashboard` | Admin | Dashboard stats |
| GET | `/v1/admin/users` | Admin | User list |
| GET | `/v1/admin/jobs` | Admin | Job management |
| GET | `/v1/admin/api-keys` | Admin | API key management |

---

## ⚙️ Configuration

All configuration is environment-driven with `IF_` prefix.

| Variable | Default | Description |
|----------|---------|-------------|
| `IF_SERVER_HOST` | `0.0.0.0` | Listen host |
| `IF_SERVER_PORT` | `8080` | Listen port |
| `IF_SERVER_MODE` | `release` | Run mode (debug/release) |
| `IF_SERVER_ALLOWED_ORIGINS` | — | CORS origins (comma-separated) |
| `IF_SERVER_TRUSTED_PROXIES` | private CIDRs | Trusted proxy CIDRs |
| `IF_DATABASE_HOST` | `localhost` | Database host |
| `IF_DATABASE_AUTO_MIGRATE` | `true` | Auto-migrate |
| `IF_AUTH_JWT_SECRET` | — | **JWT secret (required)** |
| `IF_STORAGE_PROVIDER` | `filesystem` | Storage type (r2/minio/filesystem) |
| `IF_FORCE_HTTPS` | `false` | Force HTTPS |
| `IF_CADDY_HOST` | — | Public domain |

See `.env.example` for the full list.

---

## 📁 Project Structure

```
ImageForge/
├── backend/                    # Go API server
│   ├── cmd/
│   │   ├── server/             # Main entry + DI wiring
│   │   └── admin-cli/          # Admin CLI tool
│   ├── ent/                    # Ent ORM generated code
│   │   ├── schema/             # Entity schema definitions
│   │   └── *.go                # Generated code
│   ├── internal/
│   │   ├── config/             # Viper config loading
│   │   ├── server/             # Router + Gin middleware
│   │   ├── handler/            # HTTP handlers
│   │   ├── service/            # Business logic layer
│   │   ├── repository/         # Data access layer
│   │   ├── job/                # Async job queue
│   │   ├── batchimage/         # Batch image pipeline
│   │   ├── storage/            # Object storage abstraction
│   │   ├── domain/             # Domain constants
│   │   └── pkg/                # errors, logger, response
│   └── migrations/             # SQL migration scripts
├── frontend/                   # Vue 3 SPA
│   ├── src/
│   │   ├── api/                # Axios client + modules
│   │   ├── views/              # Page components
│   │   ├── components/         # Shared UI components
│   │   ├── stores/             # Pinia state management
│   │   ├── router/             # Route config
│   │   ├── i18n/               # i18n (en/zh)
│   │   ├── styles/             # Tailwind styles
│   │   └── types/              # TypeScript types
│   └── public/                 # Static assets + Logo
├── deploy/
│   ├── docker-compose.yml      # Development
│   ├── docker-compose.prod.yml # Production
│   └── Caddyfile               # Caddy reverse proxy config
├── tests/                      # Playwright E2E tests
├── Dockerfile                  # Multi-stage build
├── Makefile                    # Dev shortcuts
└── docs/                       # PRD, engineering rules, screenshots
```

---

## 🤝 Contributing

Contributions are welcome! Please read our engineering guidelines before submitting PRs.

```bash
# Fork and clone
git clone https://github.com/Aswellle/ImageFactory.git

# Create branch
git checkout -b feat/your-feature

# Commit
git commit -m "feat(module): add new feature"

# Push and create PR
git push origin feat/your-feature
```

---

## 📄 License

This project is licensed under the [Apache License 2.0](LICENSE).

---

## 🙏 Acknowledgments

- [Sub2API](https://github.com/Wei-Shaw/sub2api) — Upstream AI gateway infrastructure
- [Ent](https://entgo.io/) — Entity-relational mapping for Go
- [Gin](https://gin-gonic.com/) — HTTP web framework
- [Vue.js](https://vuejs.org/) — Progressive JavaScript framework

---

<div align="center">

**[⬆ Back to Top](#imageforge)**

Made with ❤️ by the ImageForge Team

</div>
