GO ?= go

# Override in app Makefile to control build target, example ENV_FILE=./cmd/.env
ENV_FILE ?= .env

# Override in app Makefile to control build target, example BUILD_PKG=./cmd/my-app
BUILD_PKG ?= .

# Override in app Makefile to control build artifact destination.
BUILD_DIR ?= ./bin

BINARY_NAME ?= $(shell basename $(BUILD_PKG))

## Ensure dependencies according to lock file
deps:
	@echo ">> ensuring dependencies"
	@CGO_ENABLED=1 $(GO) mod vendor

## Run with .env vars
env: envfile
	@echo ">> running with .env"
	@bash $(APP_SCRIPTS)/env-run.sh make $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
	@kill -3 $$PPID
	@echo "Job done, stopping make, please disregard following 'make: *** [env] Error 1'"
	@exit 1

envfile:
	@test -s ./$(ENV_FILE) || (echo ">> copying .env.template to .env" && cp .env.template $(ENV_FILE))

## Run application with CompileDaemon (automatic rebuild on code change)
run-compile-daemon:
	@test -s $(shell $(GO) env GOPATH)/bin/CompileDaemon || (echo ">> installing CompileDaemon" && $(GO) get -u github.com/githubnemo/CompileDaemon)
	@echo ">> running app with CompileDaemon"
	@$(shell $(GO) env GOPATH)/bin/CompileDaemon -exclude-dir=vendor -build='make build migrate' -command='$(BUILD_DIR)/$(BINARY_NAME)' -graceful-kill


.PHONY: deps env envfile run-compile-daemon