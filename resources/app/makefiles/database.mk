GO ?= go

## Create database migration, usage: "make create-migration NAME=<migration-name>"
create-migration: migrate-cli
	@migrate create -ext=sql -dir=./resources/migrations/ "$(NAME)" && echo ">> new migration created"
	@git add ./resources/migrations

## Apply migrations
migrate: migrate-cli
	@echo ">> running migrations"
	@migrate -source=file://./resources/migrations/ -database="$(DATABASE_DSN)" up

## Rollback migrations
migrate-down: migrate-cli
	@echo ">> rolling back migrations"
	@migrate -source=file://./resources/migrations/ -database="$(DATABASE_DSN)" down


## Check/install migrations tool
migrate-cli:
	@bash $(APP_SCRIPTS)/migrate-cli.sh

.PHONY: create-migration migrate migrate-cli
