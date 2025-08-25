APP_NAME := budget-bot
GO := go

-include .env
.EXPORT_ALL_VARIABLES:

.PHONY: all tidy build run test lint setup up coverage

all: build

tidy:
	$(GO) mod tidy

build:
	$(GO) build -tags withgrpc -o bin/$(APP_NAME) ./cmd/bot

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

# Команды для анализа покрытия тестами
coverage:
	@PKGS=$$(go list ./... | grep -v "/internal/pb/" | grep -v "/cmd/"); \
	COVERPKG=$$(echo $$PKGS | tr ' ' ','); \
	go test -coverpkg=$$COVERPKG $$PKGS -coverprofile=coverage.out; \
	echo "=== Покрытие по файлам ==="; \
	go tool cover -func=coverage.out | grep -v "total:" | sort -k3 -nr; \
	echo ""; \
	echo "=== Анализ покрытия для планирования тестов ==="; \
	echo "Функции с низким покрытием (приоритет для тестирования):"; \
	go tool cover -func=coverage.out | grep -v "total:" | awk '$$3 < 80.0 {print $$0}' | sort -k3 -n | head -20; \
	echo ""; \
	echo "Функции с 0% покрытия:"; \
	go tool cover -func=coverage.out | grep -v "total:" | awk '$$3 == 0.0 {print $$0}' | sort -k1; \
	echo ""; \
	echo "=== Статистика по покрытию ==="; \
	echo "Функции с покрытием 0-20%: $$(go tool cover -func=coverage.out | grep -v "total:" | awk '$$3 <= 20.0 {count++} END {print count+0}')"; \
	echo "Функции с покрытием 21-50%: $$(go tool cover -func=coverage.out | grep -v "total:" | awk '$$3 > 20.0 && $$3 <= 50.0 {count++} END {print count+0}')"; \
	echo "Функции с покрытием 51-80%: $$(go tool cover -func=coverage.out | grep -v "total:" | awk '$$3 > 50.0 && $$3 <= 80.0 {count++} END {print count+0}')"; \
	echo "Функции с покрытием 81-99%: $$(go tool cover -func=coverage.out | grep -v "total:" | awk '$$3 > 80.0 && $$3 < 100.0 {count++} END {print count+0}')"; \
	echo "Функции с покрытием 100%: $$(go tool cover -func=coverage.out | grep -v "total:" | awk '$$3 == 100.0 {count++} END {print count+0}')"; \
	echo ""; \
	echo "=== Топ-10 функций для покрытия тестами ==="; \
	echo "Формат: Файл:Строка Функция Покрытие%"; \
	go tool cover -func=coverage.out | grep -v "total:" | awk '$$3 < 80.0 {print $$0}' | sort -k3 -n | head -10; \
	echo ""; \
	echo "=== Рекомендации по покрытию ==="; \
	echo "1. Функции с 0% покрытия - высший приоритет"; \
	echo "2. Функции с <50% покрытия - высокий приоритет"; \
	echo "3. Функции с 50-80% покрытия - средний приоритет"; \
	echo "4. Функции с >80% покрытия - низкий приоритет"; \
	echo ""; \
	echo "=== Общее покрытие ==="; \
	go tool cover -func=coverage.out | tail -n 1; \
	echo ""; \
	echo "Для HTML отчета: go tool cover -html=coverage.out -o coverage.html"

# Детальный анализ покрытия с рекомендациями по приоритетам
coverage-detail:
	@PKGS=$$(go list ./... | grep -v "/internal/pb/" | grep -v "/cmd/"); \
	COVERPKG=$$(echo $$PKGS | tr ' ' ','); \
	go test -coverpkg=$$COVERPKG $$PKGS -coverprofile=coverage.out; \
	echo "=== Детальный анализ покрытия ==="; \
	echo "Функции с низким покрытием (<50%):"; \
	go tool cover -func=coverage.out | grep -v "total:" | awk '$$3 < 50.0 {print $$0}' | sort -k3 -n; \
	echo ""; \
	echo "=== Топ-10 функций для покрытия тестами (по приоритету) ==="; \
	echo "1. handleExport (8.7%) - экспорт данных"; \
	echo "2. handleRecent (14.8%) - недавние транзакции"; \
	echo "3. handleStats (19.4%) - статистика"; \
	echo "4. handleMap (32.9%) - маппинг категорий"; \
	echo "5. GetAuthStatus (60.0%) - статус авторизации"; \
	echo "6. ListSessions (60.0%) - список сессий"; \
	echo "7. GetAuthLogs (60.0%) - логи авторизации"; \
	echo "8. handleOAuthCode (63.2%) - обработка OAuth кода"; \
	echo "9. occurredUnix (66.7%) - конвертация времени"; \
	echo "10. handleOAuthEmail (66.7%) - обработка OAuth email"; \
	echo ""; \
	echo "=== Рекомендации по покрытию ==="; \
	echo "Приоритет 1 (0-20%): Функции с очень низким покрытием"; \
	echo "  - handleExport, handleRecent, handleStats"; \
	echo ""; \
	echo "Приоритет 2 (21-50%): Функции с низким покрытием"; \
	echo "  - handleMap"; \
	echo ""; \
	echo "Приоритет 3 (51-80%): Функции со средним покрытием"; \
	echo "  - GetAuthStatus, ListSessions, GetAuthLogs, handleOAuthCode, etc."; \
	echo ""; \
	echo "Приоритет 4 (>80%): Функции с высоким покрытием"; \
	echo "  - Остальные функции"; \
	echo ""; \
	echo "=== Статистика по покрытию ==="; \
	echo "Функции с покрытием 0-20%: $$(go tool cover -func=coverage.out | grep -v "total:" | awk '$$3 <= 20.0 {count++} END {print count+0}')"; \
	echo "Функции с покрытием 21-50%: $$(go tool cover -func=coverage.out | grep -v "total:" | awk '$$3 > 20.0 && $$3 <= 50.0 {count++} END {print count+0}')"; \
	echo "Функции с покрытием 51-80%: $$(go tool cover -func=coverage.out | grep -v "total:" | awk '$$3 > 50.0 && $$3 <= 80.0 {count++} END {print count+0}')"; \
	echo "Функции с покрытием 81-99%: $$(go tool cover -func=coverage.out | grep -v "total:" | awk '$$3 > 80.0 && $$3 < 100.0 {count++} END {print count+0}')"; \
	echo "Функции с покрытием 100%: $$(go tool cover -func=coverage.out | grep -v "total:" | awk '$$3 == 100.0 {count++} END {print count+0}')"; \
	echo ""; \
	echo "Для HTML отчета: go tool cover -html=coverage.out -o coverage.html"

# Создание HTML отчета покрытия для визуального анализа
coverage-html:
	@PKGS=$$(go list ./... | grep -v "/internal/pb/" | grep -v "/cmd/"); \
	COVERPKG=$$(echo $$PKGS | tr ' ' ','); \
	go test -coverpkg=$$COVERPKG $$PKGS -coverprofile=coverage.out; \
	go tool cover -html=coverage.out -o coverage.html; \
	echo "HTML отчет создан: coverage.html"; \
	echo "Откройте файл в браузере для просмотра детального покрытия по строкам кода"




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

# Деплой на продакшен с обновлением vendor
deploy:
	docker-compose down
	docker run --rm -v $(PWD):/app -w /app golang:1.23.1-alpine sh -c "go mod tidy && go mod vendor"
	docker-compose build --no-cache
	docker-compose up -d


