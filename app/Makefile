SHELL := /usr/bin/env bash

GO ?= go
DOCKER_COMPOSE ?= docker compose

APP_COMPOSE_FILE := deploy/docker-compose.yml
DB_COMPOSE_FILE := deploy/docker-compose.postgres.yml
DOCKER_ENV_FILE := .env

.DEFAULT_GOAL := help

.PHONY: help ensure-env run test test-integration lint migrate-up migrate-down docker-up docker-down docker-db-up docker-db-down

help: ## List available commands
	@grep -E '^[a-zA-Z0-9_.-]+:.*## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*## "}; {printf "\033[36m%-18s\033[0m %s\n", $$1, $$2}'

ensure-env: ## Create .env from .env.example when missing
	@if [ ! -f "$(DOCKER_ENV_FILE)" ]; then \
		cp .env.example "$(DOCKER_ENV_FILE)"; \
		echo "Created $(DOCKER_ENV_FILE) from .env.example"; \
	fi

run: ## Run the API locally
	$(GO) run ./cmd/api

test: ## Run all unit/API tests
	$(GO) test ./...

test-integration: ## Run integration tests (requires a running database)
	$(GO) test -tags=integration ./tests/integration/...

lint: ## Run static checks
	$(GO) vet ./...

migrate-up: ## Apply SQL migrations using DATABASE_URL from the environment
	./scripts/migrate-up.sh

migrate-down: ## Roll back SQL migrations using DATABASE_URL from the environment
	./scripts/migrate-down.sh

docker-up: ensure-env ## Start full app stack (app + postgres)
	$(DOCKER_COMPOSE) --env-file $(DOCKER_ENV_FILE) -f $(APP_COMPOSE_FILE) up -d --build

docker-down: ## Stop full app stack and remove volumes
	$(DOCKER_COMPOSE) --env-file $(DOCKER_ENV_FILE) -f $(APP_COMPOSE_FILE) down -v --remove-orphans

docker-db-up: ensure-env ## Start only PostgreSQL (auto-init via deploy/init.sql)
	$(DOCKER_COMPOSE) --env-file $(DOCKER_ENV_FILE) -f $(DB_COMPOSE_FILE) up -d

docker-db-down: ## Stop PostgreSQL and remove volumes
	$(DOCKER_COMPOSE) --env-file $(DOCKER_ENV_FILE) -f $(DB_COMPOSE_FILE) down -v --remove-orphans
