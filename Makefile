APP_NAME := budget-bot
GO := go

-include .env
.EXPORT_ALL_VARIABLES:

.PHONY: all tidy build run test lint setup up

all: build

tidy:
	$(GO) mod tidy

build:
	$(GO) build -o bin/$(APP_NAME) ./cmd/bot

run: tidy build
	./bin/$(APP_NAME)

test:
	$(GO) test ./...

lint:
	golangci-lint run --timeout=3m

setup:
	mkdir -p data

up: run

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


