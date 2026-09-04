<div align="center">

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="frontend/public/logo.svg">
  <source media="(prefers-color-scheme: light)" srcset="frontend/public/logo.svg">
  <img alt="ImageForge" src="frontend/public/logo.svg" width="120" height="120">
</picture>

# ImageForge

**商业级 AI 图像生产工作空间 / Commercial AI Image-Production Workspace**

[![Go Version](https://img.shields.io/badge/Go-1.26-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![Vue Version](https://img.shields.io/badge/Vue-3.5-4FC08D?style=flat-square&logo=vuedotjs)](https://vuejs.org/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.7-3178C6?style=flat-square&logo=typescript)](https://www.typescriptlang.org/)
[![Tailwind CSS](https://img.shields.io/badge/Tailwind-4-06B6D4?style=flat-square&logo=tailwindcss)](https://tailwindcss.com/)
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=flat-square&logo=docker)](https://www.docker.com/)
[![License](https://img.shields.io/badge/License-Apache%202.0-2EC5FF?style=flat-square)](LICENSE)

[English](#english) · [中文](#中文)

---

> 创建、编辑、组织、管理商业化视觉资产，并通过 API 提供能力。
>
> *Create, edit, organize, manage commercialized visual assets, and provide capabilities via API.*

</div>

---

## 📸 Screenshots / 界面预览

| Dashboard | Gallery | Generation |
|:---:|:---:|:---:|
| ![Dashboard](docs/screenshots/dashboard.png) | ![Gallery](docs/screenshots/gallery.png) | ![Generation](docs/screenshots/generation.png) |

> 注：将界面截图保存至 `docs/screenshots/` 目录即可自动展示。
>
> *Save screenshots to `docs/screenshots/` to auto-display.*

---

## ✨ Features / 核心功能

### 🎨 Image Production / 图像生产
- **AI Image Generation** — Integrate with Gemini, OpenAI and other providers via Sub2API gateway
- **Image Editing** — Edit and version-control your visual assets
- **Async Job Queue** — Background processing with real-time status tracking
- **OpenAI-Compatible API** — Drop-in compatible with OpenAI SDK for image generation

- **AI 图像生成** — 通过 Sub2API 网关集成 Gemini、OpenAI 等供应商
- **图像编辑** — 支持编辑与版本控制的视觉资产管理
- **异步任务队列** — 后台处理与实时状态追踪
- **OpenAI 兼容 API** — 可直接使用 OpenAI SDK 接入

### 🗂 Asset Management / 资产管理
- **Projects & Collections** — Organize assets into hierarchical structures
- **Version Control** — Non-destructive editing with full version history
- **Tags & Favorites** — Flexible categorization and quick access
- **Prompt Templates** — Reusable generation prompts with variables

- **项目与合集** — 层级化结构组织资产
- **版本控制** — 非破坏性编辑，完整版本历史
- **标签与收藏** — 灵活分类与快速访问
- **提示词模板** — 可复用的生成提示词，支持变量

### 🔐 Enterprise Security / 企业级安全
- **JWT + API Key Auth** — Dual authentication paths for users and programs
- **Role-Based Access** — User / Admin roles with middleware enforcement
- **Resource Ownership** — Server-side ownership verification on every request
- **Rate Limiting** — Redis-backed distributed rate limiting

- **JWT + API Key 认证** — 用户与程序的双认证路径
- **基于角色的访问** — 用户/管理员角色，中间件强制校验
- **资源所有权** — 每次请求服务端验证所有权
- **限流保护** — Redis 支持的分布式限流

### 🚀 Production Ready / 生产就绪
- **Docker Compose** — One-command deployment with Caddy auto-HTTPS
- **Graceful Shutdown** — Zero-downtime restarts with 15s drain timeout
- **Health Checks** — Liveness + deep readiness probes (DB/Redis/storage)
- **Auto Migration** — Ent schema migration on startup

- **Docker Compose** — Caddy 自动 HTTPS 一键部署
- **优雅关闭** — 15 秒排空超时，零停机重启
- **健康检查** — 存活探针 + 深度就绪探针 (DB/Redis/存储)
- **自动迁移** — 启动时 Ent  schema 自动迁移

---

## 🏗 Architecture / 系统架构

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

## 🛠 Tech Stack / 技术栈

| Layer | Technology |
|-------|------------|
| **Backend** | Go 1.26 · Gin · Ent ORM · Viper · Zap |
| **Frontend** | Vue 3 · Vite · Tailwind CSS 4 · Pinia · Vue Router |
| **Database** | PostgreSQL 18 · Redis 8 |
| **Storage** | Cloudflare R2 · MinIO · S3-Compatible |
| **Infrastructure** | Docker Compose · Caddy · Let's Encrypt |

---

## 🚀 Quick Start / 快速开始

### Prerequisites / 环境要求

- [Docker](https://www.docker.com/) 24+ & Docker Compose v2
- [Go](https://golang.org/) 1.26+ (local development)
- [Node](https://nodejs.org/) 22+ & pnpm 10+ (local development)

### Docker Compose (Recommended / 推荐)

```bash
# Clone / 克隆
git clone https://github.com/Aswellle/ImageFactory.git
cd ImageForge

# Configure / 配置
cp .env.example .env
# Edit .env — set JWT_SECRET, DB_PASSWORD, etc.

# Start / 启动
docker compose -f deploy/docker-compose.prod.yml up -d

# Verify / 验证
curl http://localhost/v1/health
curl http://localhost/v1/health?deep=true
```

### Local Development / 本地开发

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

## 📦 Production Deployment / 生产部署

```bash
# 1. Clone
git clone https://github.com/Aswellle/ImageFactory.git && cd ImageForge

# 2. Configure (REQUIRED variables / 必填变量)
cp .env.example .env
#   IF_AUTH_JWT_SECRET    — JWT 签名密钥 (长随机字符串)
#   DB_PASSWORD           — 数据库密码
#   JWT_SECRET            — 同上
#   IF_CADDY_HOST         — 你的域名 (e.g. imageforge.example.com)
#   IF_ALLOWED_ORIGINS    — 前端域名 (e.g. https://imageforge.example.com)
#   STORAGE_SECRET_KEY    — R2/MinIO 密钥

# 3. Deploy / 部署
docker compose -f deploy/docker-compose.prod.yml up -d

# 4. Verify / 验证
curl https://your-domain/v1/health
```

See [Production Deployment Guide](#-production-ready--生产就绪) for details.

---

## 📡 API Reference / API 参考

All endpoints live under `/v1`. The API is **provider-independent**.

所有端点位于 `/v1` 下，API **与供应商无关**。

### Authentication / 认证

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/v1/auth/register` | — | 创建账户 / Create account |
| POST | `/v1/auth/login` | — | 登录 / Sign in |
| POST | `/v1/auth/send-reset-code` | — | 发送密码重置码 / Send reset code |
| POST | `/v1/auth/reset-password` | — | 重置密码 / Reset password |
| GET | `/v1/auth/me` | JWT | 当前用户 / Current user |

### Image Generation / 图像生成

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/v1/images/generations` | Public | OpenAI 兼容生成 / OpenAI-compatible generation |
| POST | `/v1/images/generate` | API Key | 程序化生成 / Programmatic generation |
| POST | `/v1/images/edits` | JWT | 图像编辑 / Image editing |
| GET | `/v1/images/jobs` | JWT | 任务列表 / List jobs |
| GET | `/v1/images/jobs/:id` | JWT | 任务详情 / Job detail |
| POST | `/v1/images/tasks` | JWT | 提交异步任务 / Submit async task |
| GET | `/v1/images/tasks/:id` | JWT | 查询任务状态 / Task status |

### Projects & Assets / 项目与资产

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/v1/projects` | JWT | 创建项目 / Create project |
| GET | `/v1/projects` | JWT | 项目列表 / List projects |
| GET | `/v1/projects/:id` | JWT | 项目详情 / Project detail |
| GET | `/v1/assets` | JWT | 资产列表 / List assets |
| GET | `/v1/assets/:id` | JWT | 资产详情 / Asset detail |
| DELETE | `/v1/assets/:id` | JWT | 删除资产 / Delete asset |
| GET | `/v1/assets/:id/content` | JWT | 资产内容 / Asset content |
| GET | `/v1/assets/:id/versions` | JWT | 版本列表 / Version history |

### Prompt Templates / 提示词模板

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/v1/prompt-templates` | JWT | 创建模板 / Create template |
| GET | `/v1/prompt-templates` | JWT | 模板列表 / List templates |
| GET | `/v1/prompt-templates/:id` | JWT | 模板详情 / Template detail |
| PUT | `/v1/prompt-templates/:id` | JWT | 更新模板 / Update template |
| DELETE | `/v1/prompt-templates/:id` | JWT | 删除模板 / Delete template |
| POST | `/v1/prompt-templates/:id/apply` | JWT | 应用模板 / Apply template |

### API Keys & Usage / API 密钥与用量

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/v1/api-keys` | JWT | 创建密钥 / Create API key |
| GET | `/v1/api-keys` | JWT | 密钥列表 / List API keys |
| DELETE | `/v1/api-keys/:id` | JWT | 撤销密钥 / Revoke API key |
| GET | `/v1/usage` | JWT | 用量统计 / Usage stats |
| GET | `/v1/usage/history` | JWT | 用量历史 / Usage history |

### Admin / 管理

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/v1/admin/dashboard` | Admin | 仪表盘数据 / Dashboard stats |
| GET | `/v1/admin/users` | Admin | 用户列表 / User list |
| GET | `/v1/admin/jobs` | Admin | 任务管理 / Job management |
| GET | `/v1/admin/api-keys` | Admin | 密钥管理 / API key management |

---

## ⚙️ Configuration / 配置

All configuration is environment-driven with `IF_` prefix.

所有配置通过环境变量驱动，统一使用 `IF_` 前缀。

| Variable | Default | Description |
|----------|---------|-------------|
| `IF_SERVER_HOST` | `0.0.0.0` | 监听地址 / Listen host |
| `IF_SERVER_PORT` | `8080` | 监听端口 / Listen port |
| `IF_SERVER_MODE` | `release` | 运行模式 (debug/release) / Run mode |
| `IF_SERVER_ALLOWED_ORIGINS` | — | CORS 来源 (逗号分隔) / CORS origins |
| `IF_SERVER_TRUSTED_PROXIES` | private CIDRs | 可信代理网段 / Trusted proxy CIDRs |
| `IF_DATABASE_HOST` | `localhost` | 数据库地址 / Database host |
| `IF_DATABASE_AUTO_MIGRATE` | `true` | 自动迁移 / Auto-migrate |
| `IF_AUTH_JWT_SECRET` | — | **JWT 密钥 (必填)** / **JWT secret (required)** |
| `IF_STORAGE_PROVIDER` | `filesystem` | 存储类型 (r2/minio/filesystem) |
| `IF_FORCE_HTTPS` | `false` | 强制 HTTPS / Force HTTPS |
| `IF_CADDY_HOST` | — | 公网域名 / Public domain |

See `.env.example` for the full list.

---

## 📁 Project Structure / 项目结构

```
ImageForge/
├── backend/                    # Go API 服务器
│   ├── cmd/
│   │   ├── server/             # 主入口 + DI  wiring
│   │   └── admin-cli/          # 管理员 CLI 工具
│   ├── ent/                    # Ent ORM 生成代码
│   │   ├── schema/             # 实体 schema 定义
│   │   └── *.go                # 生成代码
│   ├── internal/
│   │   ├── config/             # Viper 配置加载
│   │   ├── server/             # 路由器 + Gin 中间件
│   │   ├── handler/            # HTTP 处理器
│   │   ├── service/            # 业务逻辑层
│   │   ├── repository/         # 数据访问层
│   │   ├── job/                # 异步任务队列
│   │   ├── batchimage/         # 批量图像管线
│   │   ├── storage/            # 对象存储抽象
│   │   ├── domain/             # 领域常量
│   │   └── pkg/                # errors, logger, response
│   └── migrations/             # SQL 迁移脚本
├── frontend/                   # Vue 3 SPA
│   ├── src/
│   │   ├── api/                # Axios 客户端 + 模块
│   │   ├── views/              # 页面组件
│   │   ├── components/         # 通用 UI 组件
│   │   ├── stores/             # Pinia 状态管理
│   │   ├── router/             # 路由配置
│   │   ├── i18n/               # 国际化 (en/zh)
│   │   ├── styles/             # Tailwind 样式
│   │   └── types/              # TypeScript 类型
│   └── public/                 # 静态资源 + Logo
├── deploy/
│   ├── docker-compose.yml      # 开发环境
│   ├── docker-compose.prod.yml # 生产环境
│   └── Caddyfile               # Caddy 反向代理配置
├── tests/                      # Playwright 端到端测试
├── Dockerfile                  # 多阶段构建
├── Makefile                    # 开发快捷命令
└── docs/                       # PRD、工程规范、截图
```

---

## 🤝 Contributing / 贡献

Contributions are welcome! Please read our engineering guidelines before submitting PRs.

欢迎贡献代码！提交 PR 前请阅读工程规范。

```bash
# Fork and clone
git clone https://github.com/Aswellle/ImageFactory.git

# Create branch
git checkout -b feat/your-feature

# Commit (Chinese commit messages preferred)
git commit -m "feat(module): 添加新功能"

# Push and create PR
git push origin feat/your-feature
```

---

## 📄 License / 开源协议

This project is licensed under the [Apache License 2.0](LICENSE).

本项目基于 [Apache License 2.0](LICENSE) 协议开源。

---

## 🙏 Acknowledgments / 致谢

- [Sub2API](https://github.com/Wei-Shaw/sub2api) — Upstream AI gateway infrastructure
- [Ent](https://entgo.io/) — Entity-relational mapping for Go
- [Gin](https://gin-gonic.com/) — HTTP web framework
- [Vue.js](https://vuejs.org/) — Progressive JavaScript framework

---

<div align="center">

**[⬆ Back to Top / 返回顶部](#imageforge)**

Made with ❤️ by the ImageForge Team

</div>
