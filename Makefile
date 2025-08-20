APP_NAME := budget-bot
GO := go

-include .env
.EXPORT_ALL_VARIABLES:

.PHONY: all tidy build run test lint setup up coverage

all: build

tidy:
	$(GO) mod tidy

build:
	$(GO) build -o bin/$(APP_NAME) ./cmd/bot

run: tidy build
	./bin/$(APP_NAME)

test:
	$(GO) test ./...

GOLANGCI := $(shell if [ -x ./bin/golangci-lint ]; then echo ./bin/golangci-lint; else echo golangci-lint; fi)
LINT_TOOLCHAIN ?= go1.23.1

lint:
	GOTOOLCHAIN=$(LINT_TOOLCHAIN) $(GOLANGCI) run --timeout=3m

.PHONY: lint-install
lint-install:
	@echo "Installing golangci-lint v1.61.0 into ./bin..."
	@mkdir -p bin
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ./bin v1.61.0

setup:
	mkdir -p data

up: run

coverage:
	@PKGS=$$(go list ./... | grep -v "/internal/pb/" | grep -v "/cmd/"); \
	COVERPKG=$$(echo $$PKGS | tr ' ' ','); \
	go test -coverpkg=$$COVERPKG $$PKGS -coverprofile=coverage.out; \
	go tool cover -func=coverage.out | tail -n 1

.PHONY: docker-build compose-up compose-down
docker-build:
	docker build -t budget-bot:latest .

compose-up:
	docker compose up -d --build

compose-down:
	docker compose down

# --- Proto generation ---
PROTO_DIR ?= ./proto
PB_OUT := internal/pb
PROTO_FILES := $(shell find $(PROTO_DIR) -name '*.proto')

.PHONY: proto-tools proto

proto-tools:
	$(GO) install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.1
	$(GO) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0

proto: proto-tools
	rm -rf $(PB_OUT)
	mkdir -p $(PB_OUT)
	protoc -I $(PROTO_DIR) \
		--go_out=paths=source_relative:$(PB_OUT) \
		--go-grpc_out=paths=source_relative:$(PB_OUT) \
		$(PROTO_FILES)


