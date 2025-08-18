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
	@echo "lint placeholder; integrate golangci-lint later"

setup:
	mkdir -p data

up: run


