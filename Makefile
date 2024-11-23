#GOLANGCI_LINT_VERSION := "v1.43.0" # Optional configuration to pinpoint golangci-lint version.

# The head of Makefile determines location of dev-go to include standard targets.
GO ?= go
export GO111MODULE = on

ifneq "$(GOFLAGS)" ""
  $(info GOFLAGS: ${GOFLAGS})
endif

ifneq "$(wildcard ./vendor )" ""
  $(info Using vendor)
  modVendor =  -mod=vendor
  ifeq (,$(findstring -mod,$(GOFLAGS)))
      export GOFLAGS := ${GOFLAGS} ${modVendor}
  endif
  ifneq "$(wildcard ./vendor/github.com/bool64/dev)" ""
  	DEVGO_PATH := ./vendor/github.com/bool64/dev
  endif
  # adding github.com/dohernandez/dev-grpc
  ifneq "$(wildcard ./vendor/github.com/dohernandez/dev-grpc)" ""
  	DEVGRPCGO_PATH := ./vendor/github.com/dohernandez/dev-grpc
  endif
endif

ifeq ($(DEVGO_PATH),)
	DEVGO_PATH := $(shell GO111MODULE=on $(GO) list ${modVendor} -f '{{.Dir}}' -m github.com/bool64/dev)
	ifeq ($(DEVGO_PATH),)
    	$(info Module github.com/bool64/dev not found, downloading.)
    	DEVGO_PATH := $(shell export GO111MODULE=on && $(GO) get github.com/bool64/dev && $(GO) list -f '{{.Dir}}' -m github.com/bool64/dev)
	endif
endif

# defining DEVGRPCGO_PATH
ifeq ($(DEVGRPCGO_PATH),)
	DEVGRPCGO_PATH := $(shell GO111MODULE=on $(GO) list ${modVendor} -f '{{.Dir}}' -m github.com/bool64/dev)
	ifeq ($(DEVGRPCGO_PATH),)
    	$(info Module github.com/dohernandez/dev-grpc not found, downloading.)
    	DEVGRPCGO_PATH := $(shell export GO111MODULE=on && $(GO) get github.com/dohernandez/dev-grpc && $(GO) list -f '{{.Dir}}' -m github.com/dohernandez/dev-grpc)
	endif
endif

-include $(DEVGO_PATH)/makefiles/main.mk
-include $(DEVGO_PATH)/makefiles/build.mk
-include $(DEVGO_PATH)/makefiles/lint.mk
-include $(DEVGO_PATH)/makefiles/test-unit.mk
-include $(DEVGO_PATH)/makefiles/test-integration.mk
-include $(DEVGO_PATH)/makefiles/bench.mk
-include $(DEVGO_PATH)/makefiles/reset-ci.mk

# Add your custom targets here.
BUILD_LDFLAGS="-s -w"
BUILD_PKG = ./cmd/vio
INTEGRATION_DOCKER_COMPOSE = ./docker-compose.integration-test.yml

APP_PATH = $(shell pwd)
APP_SCRIPTS = $(APP_PATH)/resources/app/scripts
SRC_PROTO_PATH = $(APP_PATH)/resources/proto
GO_PROTO_PATH = $(APP_PATH)/pkg/proto
SWAGGER_PATH = $(APP_PATH)/resources/swagger

-include $(APP_PATH)/resources/app/makefiles/database.mk
-include $(APP_PATH)/resources/app/makefiles/dep.mk
-include $(DEVGRPCGO_PATH)/makefiles/protoc.mk

## Run tests
test: test-unit test-integration

## Generate code from proto file(s)
proto-gen: proto-gen-code-swagger
	@cat $(SWAGGER_PATH)/service.swagger.json | jq del\(.paths[][].responses.'"default"'\) > $(SWAGGER_PATH)/service.swagger.json.tmp
	@mv $(SWAGGER_PATH)/service.swagger.json.tmp $(SWAGGER_PATH)/service.swagger.json

## Run integration benchmark
bench-integration:
	@make start-deps
	@echo ">> running integration benchmark"
	@$(GO) test -tags bench -bench=. -count=5 -benchtime=20000x -run=^a  benchmark_test.go | tee /dev/tty >bench-$(shell git symbolic-ref HEAD --short | tr / - 2>/dev/null).txt
	@test -s $(GOPATH)/bin/benchstat || GO111MODULE=off GOFLAGS= GOBIN=$(GOPATH)/bin $(GO) get -u golang.org/x/perf/cmd/benchstat
	@benchstat bench-$(shell git symbolic-ref HEAD --short | tr / - 2>/dev/null).txt