# ImageForge Makefile — local development shortcuts.
# Production deploys use Docker (see deploy/docker-compose.yml).

SHELL := /bin/sh
BACKEND := backend
FRONTEND := frontend

.PHONY: help backend frontend dev down migrate

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-16s\033[0m %s\n", $$1, $$2}'

# --- Backend ---

backend: ## Build the backend binary
	cd $(BACKEND) && go build -o ../build/imageforge ./cmd/server

backend-generate: ## Regenerate Ent code after schema changes
	cd $(BACKEND) && go generate ./ent

backend-lint: ## Lint the backend
	cd $(BACKEND) && golangci-lint run ./...

backend-test: ## Run backend unit tests (build tag `unit`)
	cd $(BACKEND) && go test -tags=unit ./...

# --- Frontend ---

frontend: ## Install deps and build the frontend
	cd $(FRONTEND) && pnpm install && pnpm build

frontend-install: ## Install frontend deps only
	cd $(FRONTEND) && pnpm install

frontend-dev: ## Run the frontend dev server
	cd $(FRONTEND) && pnpm dev

frontend-typecheck: ## Type-check the frontend
	cd $(FRONTEND) && pnpm typecheck

# --- Combined ---

dev: ## Run backend + frontend dev servers (requires both toolsets)
	@echo "Backend:  make backend-dev  (or: cd backend && go run ./cmd/server)"
	@echo "Frontend: make frontend-dev"

docker-up: ## Start full stack via Docker Compose
	docker compose -f deploy/docker-compose.yml up --build -d

docker-down: ## Stop Docker Compose stack
	docker compose -f deploy/docker-compose.yml down

docker-logs: ## Tail the backend logs
	docker compose -f deploy/docker-compose.yml logs -f backend
