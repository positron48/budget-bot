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

build-fake:
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




# --- Docker commands ---
DOCKER_IMAGE := budget-bot
DOCKER_TAG := latest

.PHONY: docker-build docker-run docker-stop docker-logs docker-clean

docker-build:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-run:
	docker-compose up -d

docker-stop:
	docker-compose down

docker-logs:
	docker-compose logs -f

docker-clean:
	docker-compose down -v
	docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) || true

docker-rebuild: docker-clean docker-build docker-run

# Команды для разработки
docker-dev:
	docker-compose up -d

docker-dev-logs:
	docker-compose logs -f

docker-dev-restart:
	docker-compose restart budget-bot

# Установка зависимостей через Docker
docker-deps:
	docker run --rm -v $(PWD):/app -w /app golang:1.23.1-alpine sh -c "go mod download && go mod vendor"

# Очистка vendor
docker-clean-deps:
	rm -rf vendor/


