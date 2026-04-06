SHELL := /usr/bin/env bash

GO ?= go
DOCKER_COMPOSE ?= docker compose
ENV_FILE := .env
ENV_EXAMPLE := .env.example
DB_COMPOSE_FILE := deploy/docker-compose.postgres.yml

BLUE := \033[1;34m
GREEN := \033[1;32m
YELLOW := \033[1;33m
RED := \033[1;31m
RESET := \033[0m

.DEFAULT_GOAL := help

.PHONY: help ensure-env remind-external run test lint test-integration docker-up docker-down migrate-up migrate-down

help: ## List commands
	@grep -E '^[a-zA-Z0-9_.-]+:.*## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*## "}; {printf "\033[36m%-18s\033[0m %s\n", $$1, $$2}'

ensure-env: ## Create .env from .env.example if missing
	@if [ ! -f "$(ENV_FILE)" ]; then \
		cp "$(ENV_EXAMPLE)" "$(ENV_FILE)"; \
		printf "$(YELLOW)[WARN]$(RESET) %s\n" "Criei $(ENV_FILE) a partir de $(ENV_EXAMPLE)."; \
	else \
		printf "$(GREEN)[OK]$(RESET) %s\n" "$(ENV_FILE) encontrado."; \
	fi

remind-external: ## Reminder about Besu/deploy/env
	@printf "$(YELLOW)[WARN]$(RESET) %s\n" "Este Makefile NAO sobe Besu nem faz deploy de contrato."
	@printf "$(BLUE)[INFO]$(RESET) %s\n" "Antes de rodar a API, suba Besu/deploy pela raiz e atualize CONTRACT_ADDRESS/CHAIN_ID no $(ENV_FILE)."

run: ensure-env remind-external ## Run API locally
	@printf "$(BLUE)[INFO]$(RESET) %s\n" "Iniciando API local..."
	@set -a; . "$(ENV_FILE)"; set +a; $(GO) run ./cmd/api

test: ## Run tests
	@printf "$(BLUE)[INFO]$(RESET) %s\n" "Executando testes..."
	@$(GO) test ./...
	@printf "$(GREEN)[OK]$(RESET) %s\n" "Testes finalizados."

lint: ## Run go vet
	@printf "$(BLUE)[INFO]$(RESET) %s\n" "Executando lint..."
	@$(GO) vet ./...
	@printf "$(GREEN)[OK]$(RESET) %s\n" "Lint finalizado."

test-integration: ## Run integration tests (DB must be running)
	@printf "$(YELLOW)[WARN]$(RESET) %s\n" "Garanta que o banco esteja rodando: make docker-up"
	@$(GO) test -tags=integration ./tests/integration/...

docker-up: ensure-env remind-external ## Start PostgreSQL only
	@printf "$(BLUE)[INFO]$(RESET) %s\n" "Subindo PostgreSQL..."
	@set +e; \
	out="$$( $(DOCKER_COMPOSE) --env-file $(ENV_FILE) -f $(DB_COMPOSE_FILE) up -d 2>&1 )"; \
	code=$$?; \
	if [ $$code -ne 0 ]; then printf "$(RED)[ERROR]$(RESET) %s\n%s\n" "Falha ao subir PostgreSQL." "$$out"; exit $$code; fi; \
	printf "%s\n" "$$out"; \
	printf "$(GREEN)[OK]$(RESET) %s\n" "PostgreSQL iniciado."

docker-down: ## Stop PostgreSQL and remove volumes
	@printf "$(YELLOW)[WARN]$(RESET) %s\n" "Parando PostgreSQL e removendo volumes (dados serão apagados)."
	@set +e; \
	env_args=""; [ -f "$(ENV_FILE)" ] && env_args="--env-file $(ENV_FILE)"; \
	out="$$( $(DOCKER_COMPOSE) $$env_args -f $(DB_COMPOSE_FILE) down -v --remove-orphans 2>&1 )"; \
	code=$$?; \
	if [ $$code -ne 0 ]; then printf "$(RED)[ERROR]$(RESET) %s\n%s\n" "Falha ao remover PostgreSQL." "$$out"; exit $$code; fi; \
	printf "%s\n" "$$out"; \
	printf "$(GREEN)[OK]$(RESET) %s\n" "PostgreSQL removido."

migrate-up: ensure-env ## Apply SQL migrations
	@printf "$(BLUE)[INFO]$(RESET) %s\n" "Aplicando migrations up..."
	@set -euo pipefail; \
	set -a; . "$(ENV_FILE)"; set +a; \
	files=(internal/infrastructure/persistence/postgres/migrations/*up.sql); \
	if [ $${#files[@]} -eq 0 ]; then printf "$(RED)[ERROR]$(RESET) %s\n" "Nenhuma migration up encontrada."; exit 1; fi; \
	if command -v psql >/dev/null 2>&1; then \
		for f in "$${files[@]}"; do psql "$$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$$f"; done; \
	else \
		for f in "$${files[@]}"; do cat "$$f" | docker exec -i goledger-postgres psql -v ON_ERROR_STOP=1 -U "$${POSTGRES_USER:-postgres}" -d "$${POSTGRES_DB:-goledger}"; done; \
	fi
	@printf "$(GREEN)[OK]$(RESET) %s\n" "Migrations up concluídas."

migrate-down: ensure-env ## Apply SQL migrations down
	@printf "$(YELLOW)[WARN]$(RESET) %s\n" "Aplicando migrations down..."
	@set -euo pipefail; \
	set -a; . "$(ENV_FILE)"; set +a; \
	files=($$(ls -1 internal/infrastructure/persistence/postgres/migrations/*down.sql 2>/dev/null | sort -r)); \
	if [ $${#files[@]} -eq 0 ]; then printf "$(RED)[ERROR]$(RESET) %s\n" "Nenhuma migration down encontrada."; exit 1; fi; \
	if command -v psql >/dev/null 2>&1; then \
		for f in "$${files[@]}"; do psql "$$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$$f"; done; \
	else \
		for f in "$${files[@]}"; do cat "$$f" | docker exec -i goledger-postgres psql -v ON_ERROR_STOP=1 -U "$${POSTGRES_USER:-postgres}" -d "$${POSTGRES_DB:-goledger}"; done; \
	fi
	@printf "$(GREEN)[OK]$(RESET) %s\n" "Migrations down concluídas."
