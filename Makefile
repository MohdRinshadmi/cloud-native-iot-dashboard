# =============================================================================
# Cloud-Native IoT Analytics Dashboard — Developer Command Center
# =============================================================================
SHELL := /bin/bash
.DEFAULT_GOAL := help

COMPOSE := docker compose -f docker-compose.yml
BACKEND_DIR := backend
FRONTEND_DIR := frontend

## ----------------------------------------------------------------------------
## Help
## ----------------------------------------------------------------------------
.PHONY: help
help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@echo ""

## ----------------------------------------------------------------------------
## Bootstrap
## ----------------------------------------------------------------------------
.PHONY: setup
setup: ## Install all dependencies (backend + frontend) and copy env
	@test -f .env || cp .env.example .env
	cd $(BACKEND_DIR) && go mod download
	cd $(FRONTEND_DIR) && npm install

## ----------------------------------------------------------------------------
## Infrastructure (Postgres, Redis, Mosquitto, Prometheus, Grafana)
## ----------------------------------------------------------------------------
.PHONY: infra-up
infra-up: ## Start backing services (db, cache, mqtt, observability)
	$(COMPOSE) up -d postgres redis mosquitto prometheus grafana

.PHONY: infra-down
infra-down: ## Stop backing services
	$(COMPOSE) down

.PHONY: infra-logs
infra-logs: ## Tail infra logs
	$(COMPOSE) logs -f

.PHONY: up
up: ## Start the full stack (infra + api + web) in containers
	$(COMPOSE) up -d --build

.PHONY: down
down: ## Stop the full stack and remove volumes
	$(COMPOSE) down -v

## ----------------------------------------------------------------------------
## Backend
## ----------------------------------------------------------------------------
.PHONY: be-run
be-run: ## Run the Gin API locally (hot infra must be up)
	cd $(BACKEND_DIR) && go run ./cmd/api

.PHONY: be-build
be-build: ## Compile the API binary
	cd $(BACKEND_DIR) && go build -o bin/api ./cmd/api

.PHONY: be-test
be-test: ## Run backend tests
	cd $(BACKEND_DIR) && go test ./... -race -count=1

.PHONY: be-tidy
be-tidy: ## Tidy go modules
	cd $(BACKEND_DIR) && go mod tidy

.PHONY: be-lint
be-lint: ## Vet + format check
	cd $(BACKEND_DIR) && go vet ./... && gofmt -l .

.PHONY: sim
sim: ## Run the IoT device simulator (publishes MQTT telemetry)
	cd $(BACKEND_DIR) && go run ./cmd/simulator -interval 2s

## ----------------------------------------------------------------------------
## Frontend
## ----------------------------------------------------------------------------
.PHONY: fe-run
fe-run: ## Run the Vite dev server
	cd $(FRONTEND_DIR) && npm run dev

.PHONY: fe-build
fe-build: ## Production build of the web app
	cd $(FRONTEND_DIR) && npm run build

.PHONY: fe-test
fe-test: ## Run frontend tests
	cd $(FRONTEND_DIR) && npm run test

.PHONY: fe-lint
fe-lint: ## Lint + typecheck the web app
	cd $(FRONTEND_DIR) && npm run lint && npm run typecheck

## ----------------------------------------------------------------------------
## Quality gates (run before pushing)
## ----------------------------------------------------------------------------
.PHONY: check
check: be-lint be-test fe-lint ## Run all quality gates
	@echo "✅ all checks passed"

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf $(BACKEND_DIR)/bin $(FRONTEND_DIR)/dist
