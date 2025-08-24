# Многоэтапная сборка для оптимизации размера образа
FROM golang:1.23.1-alpine AS builder

# Установка необходимых пакетов для сборки
RUN apk add --no-cache \
    git \
    protobuf \
    protobuf-dev \
    make

# Установка Go инструментов для protobuf
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.1 && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0

# Установка рабочей директории
WORKDIR /app

# Копирование файлов зависимостей
COPY go.mod go.sum ./

# Загрузка зависимостей
RUN go mod download

# Копирование исходного кода
COPY . .

# Генерация protobuf файлов
RUN make proto

# Сборка приложения
RUN CGO_ENABLED=0 GOOS=linux go build -tags withgrpc -a -installsuffix cgo -o bin/budget-bot ./cmd/bot

# Финальный образ
FROM alpine:latest

# Установка необходимых пакетов для runtime
RUN apk add --no-cache \
    ca-certificates \
    sqlite \
    tzdata \
    && rm -rf /var/cache/apk/*

# Создание пользователя для безопасности
RUN addgroup -g 1001 -S budgetbot && \
    adduser -u 1001 -S budgetbot -G budgetbot

# Создание необходимых директорий
RUN mkdir -p /app/bin /app/data /app/configs /app/logs && \
    chown -R budgetbot:budgetbot /app

# Копирование бинарного файла из builder
COPY --from=builder /app/bin/budget-bot /app/bin/budget-bot

# Копирование конфигурационных файлов
COPY --chown=budgetbot:budgetbot configs/ /app/configs/
COPY --chown=budgetbot:budgetbot migrations/ /app/migrations/

# Установка рабочей директории
WORKDIR /app

# Переключение на пользователя budgetbot
USER budgetbot

# Экспорт портов
EXPOSE 8088 9090

# Проверка здоровья
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8088/health || exit 1

# Запуск приложения
CMD ["./bin/budget-bot"]
