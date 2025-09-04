# ====== SETTINGS ======
COMPOSE        ?= docker compose
BASE_FILE      ?= docker-compose.yaml
DEV_FILE       ?= docker-compose.dev.yaml
ENV_FILE       ?= .env

# если нужно, переопредели список "инфраструктуры", которую надо поднять до миграций
INFRA_SERVICES ?= postgres kafka

DC := $(COMPOSE) -f $(BASE_FILE) -f $(DEV_FILE) --env-file $(ENV_FILE)

# ====== HELP ======
.PHONY: help
help:
	@echo "Targets:"
	@echo "  up-dev           - поднять инфру, прогнать миграции, запустить весь стек"
	@echo "  up-db            - поднять только БД (и инфру), без остальных сервисов"
	@echo "  down-dev         - остановить и удалить контейнеры + тома"
	@echo "  logs             - хвост логов всех сервисов"
	@echo "  ps               - статус сервисов"
	@echo "  migrate-up       - применить все новые миграции"
	@echo "  migrate-up-one   - применить ровно одну миграцию"
	@echo "  migrate-steps    - N шагов вверх/вниз (make migrate-steps n=-1)"
	@echo "  migrate-goto     - перейти к версии (make migrate-goto v=20250901_0001)"
	@echo "  migrate-version  - показать текущую версию"
	@echo "  migrate-force    - форснуть версию (аварийно)"
	@echo "  migrate-create   - создать файлы миграции (make migrate-create name=add_x)"

# ====== HIGH-LEVEL (ONE BUTTON) ======
.PHONY: up-dev up-db down-dev logs ps
up-dev: ensure-env
	$(DC) up -d $(INFRA_SERVICES)
	$(DC) run --rm  migrate up
	$(DC) up -d

up-db: ensure-env
	$(DC) up -d $(INFRA_SERVICES)

down-dev: ensure-env
	$(DC) down -v --remove-orphans


logs: ensure-env
	$(DC) logs -f --tail=200

ps: ensure-env
	$(DC) ps

# ====== MIGRATIONS ======
.PHONY: migrate-up migrate-up-one migrate-steps migrate-goto migrate-version migrate-force migrate-create
migrate-up: ensure-env
	$(DC) run --rm -e MIGRATE_CMD=up migrate

migrate-up-one: ensure-env
	$(DC) run --rm -e MIGRATE_CMD="up 1" migrate

migrate-steps: ensure-env
	@test -n "$(n)" || (echo "Usage: make migrate-steps n=±N"; exit 1)
	$(DC) run --rm -e MIGRATE_CMD="steps $(n)" migrate

migrate-goto: ensure-env
	@test -n "$(v)" || (echo "Usage: make migrate-goto v=<version>"; exit 1)
	$(DC) run --rm -e MIGRATE_CMD="goto $(v)" migrate

migrate-version: ensure-env
	$(DC) run --rm -e MIGRATE_CMD=version migrate

migrate-force: ensure-env
	@test -n "$(v)" || (echo "Usage: make migrate-force v=<version>"; exit 1)
	$(DC) run --rm -e MIGRATE_CMD="force $(v)" migrate

# создание пары *.up.sql/*.down.sql (без локальной установки CLI)
MIGR_DIR      := db/migrations
MIGRATE_IMAGE := migrate/migrate:4
migrate-create:
	@test -n "$(name)" || (echo "Usage: make migrate-create name=<snake_case>"; exit 1)
	docker run --rm -v "$(PWD)/$(MIGR_DIR):/migrations" $(MIGRATE_IMAGE) \
	  create -ext sql -dir /migrations -seq $(name)

# ====== UTIL ======
.PHONY: ensure-env
ensure-env:
	@test -f "$(ENV_FILE)" || (cp .env.example "$(ENV_FILE)" && echo ">> Created $(ENV_FILE) from .env.example")
