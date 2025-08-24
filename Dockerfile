# Базовый образ с Go и необходимыми инструментами
FROM golang:1.23.1-alpine

# Установка необходимых пакетов
RUN apk add --no-cache \
    git \
    make \
    ca-certificates \
    sqlite \
    tzdata \
    wget \
    && rm -rf /var/cache/apk/*

# Создание пользователя для безопасности
RUN addgroup -g 1001 -S budgetbot && \
    adduser -u 1001 -S budgetbot -G budgetbot

# Создание необходимых директорий
RUN mkdir -p /app/bin /app/data /app/configs /app/logs /app/migrations && \
    chown -R budgetbot:budgetbot /app && \
    chmod 755 /app/data



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
CMD ["sh", "-c", "touch /app/data/bot.sqlite && chmod 666 /app/data/bot.sqlite && go build -mod=vendor -o bin/budget-bot ./cmd/bot && ./bin/budget-bot"]
