SHELL := /bin/sh

# ----------------------------------------------------------------------
# Services
# ----------------------------------------------------------------------

API        := UniQUE-API
AUTH       := UniQUE-Auth
DISCORD    := UniQUE-Discord
MAILSERVER := UniQUE-MailServer


# ----------------------------------------------------------------------
# Help
# ----------------------------------------------------------------------

.PHONY: help
help:
	@echo "Available targets:"
	@echo ""
	@echo "  Build"
	@echo "    make build          Build all services"
	@echo "    make api-build      Build UniQUE-API"
	@echo "    make auth-build     Build UniQUE-Auth"
	@echo "    make discord-build  Build UniQUE-Discord"
	@echo "    make mail-build     Build UniQUE-MailServer"
	@echo ""
	@echo "  Development"
	@echo "    make api-dev        Run UniQUE-API in development mode"
	@echo "    make auth-dev       Run UniQUE-Auth in development mode"
	@echo ""
	@echo "  Generate"
	@echo "    make gen            Run generation scripts"
	@echo "    make swag           Generate Swagger documents"
	@echo ""
	@echo "  Docker Compose"
	@echo "    make up             Start containers"
	@echo "    make down           Stop containers"
	@echo "    make logs           Show container logs"


# ----------------------------------------------------------------------
# UniQUE-API
# ----------------------------------------------------------------------

.PHONY: api-build api-dev api-gen api-swag

api-build:
	cd $(API)/src && ../scripts/build.sh

api-dev:
	cd $(API)/src && ../scripts/build.sh --dev

api-gen:
	cd $(API)/src && ../scripts/gen.sh

api-swag:
	cd $(API)/src && ../scripts/swag.sh


# ----------------------------------------------------------------------
# UniQUE-Auth
# ----------------------------------------------------------------------

.PHONY: auth-build auth-dev auth-gen auth-swag auth-keys auth-run

auth-build:
	cd $(AUTH)/src && ../scripts/build.sh

auth-dev:
	cd $(AUTH)/src && ../scripts/build.sh --dev

auth-gen:
	cd $(AUTH)/src && ../scripts/gen.sh

auth-swag:
	cd $(AUTH)/src && ../scripts/swag.sh

auth-keys:
	cd $(AUTH)/src && ../scripts/keys.sh

auth-run:
	cd $(AUTH)/src && ../scripts/run.sh


# ----------------------------------------------------------------------
# UniQUE-Discord
# ----------------------------------------------------------------------

.PHONY: discord-build

discord-build:
	cd $(DISCORD)/src && ../scripts/build.sh

discord-dev:
	cd $(DISCORD)/src && ../scripts/build.sh --dev

# ----------------------------------------------------------------------
# UniQUE-MailServer
# ----------------------------------------------------------------------

.PHONY: mail-build mail-gen

mail-build:
	cd $(MAILSERVER)/src && ../scripts/build.sh

mail-dev:
	cd $(MAILSERVER)/src && ../scripts/build.sh --dev

mail-gen:
	cd $(MAILSERVER)/src && ../scripts/gen.sh


# ----------------------------------------------------------------------
# All
# ----------------------------------------------------------------------

.PHONY: build gen swag

build: api-build auth-build discord-build mail-build

gen: api-gen auth-gen mail-gen

swag: api-swag auth-swag


# ----------------------------------------------------------------------
# Docker Compose
# ----------------------------------------------------------------------

.PHONY: up down logs ps

up:
	docker compose up -d

up-build:
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f

ps:
	docker compose ps
