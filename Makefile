# ===== config =====
PROJECT ?= wb-l0
DEV     ?=                 
APPS    := web consumer emulator
DEV_APPS := kafka-ui pgadmin   
INFRA   := postgres kafka

# Базовая команда compose
DC_BASE := docker compose -p $(PROJECT) -f docker-compose.yaml
# Если DEV задан — добавляем профиль dev
DC      := $(DC_BASE) $(if $(DEV),--profile dev,)

# Миграции живут в профиле tooling
MIG := $(DC_BASE) --profile tooling $(if $(DEV),--profile dev,) run --rm migrate

# Масштабирование из параметров: make up consumer=3 emulator=2
SCALE := $(if $(consumer),--scale consumer=$(consumer)) \
         $(if $(emulator),--scale emulator=$(emulator))

.PHONY: up ps down down-v restart logs config wait-db mig-up mig-down mig-down-all mig-version mig-new \
        infra apps up-all help

# ===== common =====
up:
	$(DC) up -d --build $(SCALE)

ps:
	$(DC) ps

down:
	$(DC) down

down-v:
	$(DC) down -v

restart:
	@if [ -z "$(S)" ]; then echo "Usage: make restart S=<service>"; exit 1; fi
	$(DC) restart $(S)

logs:
	@if [ -z "$(S)" ]; then echo "Usage: make logs S=<service>"; exit 1; fi
	$(DC) logs -f $(S)

config:
	$(DC) config

help:
	@echo "make up [DEV=1] [consumer=N emulator=M]   - поднять стек (dev-профиль по флагу)"
	@echo "make infra [DEV=1]                         - поднять только postgres+kafka"
	@echo "make apps [DEV=1] [consumer=N emulator=M]  - поднять только приложения"
	@echo "make up-all [DEV=1]                        - infra -> wait-db -> mig-up -> apps"
	@echo "make down / down-v                         - остановить (down-v удалит volumes!)"
	@echo "make logs S=<service>                      - логи сервиса"
	@echo "make restart S=<service>                   - рестарт сервиса"
	@echo "make mig-up|mig-down|mig-down-all|mig-version|mig-new name=...  - миграции"
	@echo "make config                                 - показать финальный compose"

# ===== migrations =====
mig-up:
	$(MIG) up

mig-down:
	$(MIG) down 1

mig-down-all:
	$(MIG) down

mig-version:
	$(MIG) version

# usage: make mig-new name=create_orders_table
mig-new:
	@if [ -z "$(name)" ]; then echo "Usage: make mig-new name=<snake_case_name>"; exit 1; fi
	$(MIG) create -ext sql -dir /migrations -seq "$(name)"

# ===== wait db =====
wait-db:
	@$(DC) exec -T postgres bash -lc 'for i in {1..30}; do pg_isready -U $${POSTGRES_USER:-demo_user} -d $${POSTGRES_DB:-demo_db} >/dev/null 2>&1 && exit 0; sleep 1; done; echo "DB not ready"; exit 1'

# ===== infra + apps =====
infra:
	$(DC) up -d --build $(INFRA)

APPS_ALL := $(APPS) $(if $(DEV),$(DEV_APPS),)

apps:
	$(DC) up -d --build $(APPS_ALL) $(SCALE)

up-all:
	$(MAKE) infra DEV=$(DEV)
	$(MAKE) wait-db DEV=$(DEV)
	$(MAKE) mig-up DEV=$(DEV)
	$(MAKE) apps DEV=$(DEV) consumer=$(consumer) emulator=$(emulator)


# ==== demonstation =====
up-load:
	$(MAKE) infra DEV=
	$(MAKE) wait-db DEV=
	$(MAKE) mig-up DEV=
	$(MAKE) apps DEV= consumer=3 emulator=2