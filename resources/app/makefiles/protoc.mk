GO ?= go

## Check/install protoc tool
protoc-cli:
	@bash $(APP_SCRIPTS)/protoc-gen-cli.sh

.PHONY: protoc-cli
