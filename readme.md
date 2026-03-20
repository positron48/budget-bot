# Telegram Bot для учета расходов и доходов

## 📋 Описание

Telegram бот для быстрого добавления транзакций в систему учета личных финансов. Бот выступает как самостоятельный клиент для gRPC сервиса бюджета, обеспечивая удобный интерфейс для добавления транзакций через Telegram.

### 🎯 Возможности

- **Быстрое добавление транзакций** через Telegram без необходимости открывать веб-интерфейс
- **Интеграция с существующей системой** через gRPC API
- **Умная категоризация** транзакций на основе описания
- **Многопользовательность** с привязкой Telegram аккаунтов к пользователям системы
- **Простота использования** - минимальное количество действий для добавления транзакции

## 🚀 Быстрый старт

### Предварительные требования

- Go 1.23+
- SQLite
- Telegram Bot Token (получить у [@BotFather](https://t.me/BotFather))

### Установка и запуск

1. **Клонирование репозитория**:
   ```bash
   git clone <repository-url>
   cd budget-bot
   ```

2. **Настройка переменных окружения**:
   ```bash
   cp env.example .env
   # Отредактируйте .env файл, указав ваш TELEGRAM_BOT_TOKEN
   ```

3. **Генерация protobuf клиентов**:
   ```bash
   make proto
   ```

4. **Сборка и запуск**:
   ```bash
   # Для локальной разработки (с фейковыми gRPC клиентами)
   make build-fake
   make run
   
   # Для продакшена (с реальными gRPC клиентами)
   make build
   make run
   
   # Или напрямую
   go build -o bin/budget-bot ./cmd/bot
   ./bin/budget-bot
   ```

### Развертывание в продакшене

#### Systemd сервис (рекомендуется)

Создайте файл `/etc/systemd/system/budget-bot.service`:

```ini
[Unit]
Description=Budget Bot
After=network.target

[Service]
Type=simple
User=bot
WorkingDirectory=/opt/budget-bot
ExecStart=/opt/budget-bot/budget-bot
Restart=always
RestartSec=10
Environment=TELEGRAM_BOT_TOKEN=your_token_here
Environment=TELEGRAM_WEBHOOK_ENABLE=true
Environment=TELEGRAM_WEBHOOK_DOMAIN=https://your-domain.com
Environment=GRPC_SERVER_ADDRESS=your_grpc_server:8081

[Install]
WantedBy=multi-user.target
```

Затем:
```bash
sudo systemctl daemon-reload
sudo systemctl enable budget-bot
sudo systemctl start budget-bot
sudo systemctl status budget-bot
```

#### Запуск в фоне

```bash
nohup ./budget-bot > bot.log 2>&1 &
```



## 💬 Как пользоваться

### Регистрация и вход

1. **Начало работы**: Отправьте `/start` боту
2. **Регистрация**: Используйте `/register` для создания нового аккаунта
3. **Вход**: Используйте `/login` для входа в существующий аккаунт
4. **Выход**: Используйте `/logout` для выхода из системы

### Добавление транзакций

Бот принимает сообщения в следующем формате:

```
[дата] [+]сумма[валюта] описание
```

#### Примеры сообщений:

```
1000 продукты                    # Расход 1000 в валюте по умолчанию
+50000 зарплата                  # Доход 50000
01.12 5000 подарок              # Расход с датой
12.12.2023 1234.56 такси        # Расход с точной датой
вчера 100 кофе                   # Расход за вчера
сегодня +1000 возврат            # Доход за сегодня

# С указанием валюты
1000₽ продукты                   # Расход в рублях
+50000$ зарплата                 # Доход в долларах
01.12 5000 EUR подарок          # Расход в евро
```

#### Поддерживаемые форматы дат:
- `сегодня`, `вчера`, `позавчера`
- `DD.MM.YYYY` (например, `15.12.2023`)
- `DD.MM` (например, `15.12` - текущий год)

#### Поддерживаемые валюты:
- Символы: ₽, $, €, £, ¥
- Коды: RUB, USD, EUR, GBP, JPY

### Управление категориями

   - `/categories` - список доступных категорий
- `/map слово = категория` - добавить сопоставление слова с категорией
   - `/map слово` - показать текущее сопоставление
   - `/map --all` - показать все сопоставления
   - `/unmap слово` - удалить сопоставление

### Статистика и отчеты

- `/stats` - статистика за текущий месяц
- `/stats 2023-12` - статистика за конкретный месяц
- `/stats week` - статистика за текущую неделю
- `/top_categories` - топ категорий по расходам
- `/recent` - последние транзакции
- `/export` - экспорт данных в CSV

### Настройки

- `/switch_tenant` - переключение между организациями
- `/language` - выбор языка интерфейса
- `/currency` - настройка валюты по умолчанию
- `/settings` - общие настройки бота

## ⚙️ Конфигурация

### Основные переменные окружения

```bash
# Telegram
TELEGRAM_BOT_TOKEN=your_bot_token_here
TELEGRAM_API_BASE_URL=https://api.telegram.org/
TELEGRAM_DEBUG=false
TELEGRAM_SOCKS5_PROXY=

# Webhook (опционально)
TELEGRAM_WEBHOOK_ENABLE=false
TELEGRAM_WEBHOOK_DOMAIN=https://your-domain.com
TELEGRAM_WEBHOOK_PATH=/tg

# gRPC сервер
GRPC_SERVER_ADDRESS=127.0.0.1:8081
GRPC_INSECURE=true

# База данных
DATABASE_DRIVER=sqlite
DATABASE_DSN=file:./data/bot.sqlite?_foreign_keys=on

# Логирование
LOG_LEVEL=info

# Метрики
METRICS_ENABLED=false
METRICS_ADDRESS=:9090
```

### Конфигурационный файл

Создайте `configs/config.yaml` для дополнительных настроек:

```yaml
telegram:
  token: ${TELEGRAM_BOT_TOKEN}
  timeout: 30s

grpc:
  server_address: ${GRPC_SERVER_ADDRESS}
  timeout: 10s
  retry_attempts: 3

database:
  url: ${DATABASE_DSN}
  max_connections: 10

logging:
  level: ${LOG_LEVEL}
  format: json
```

### Webhook режим

Бот поддерживает два режима работы с Telegram API:

1. **Long Polling** (по умолчанию) - бот сам запрашивает обновления
2. **Webhook** - Telegram отправляет обновления на указанный URL

Для включения webhook режима:

```bash
# Включить webhook
TELEGRAM_WEBHOOK_ENABLE=true

# Указать домен с протоколом
TELEGRAM_WEBHOOK_DOMAIN=https://your-domain.com

# Путь для webhook (по умолчанию /tg)
TELEGRAM_WEBHOOK_PATH=/tg

# API для управления webhook (может быть эмулятор)
TELEGRAM_API_BASE_URL=http://127.0.0.1:3001/bot%s/%s
```

**Важно:** Webhook устанавливается и удаляется автоматически через API, указанный в `TELEGRAM_API_BASE_URL`.

Подробная документация по настройке webhook: [readme_webhook_setup.md](readme_webhook_setup.md)

## 🏗️ Архитектура

### Компоненты системы

- **Telegram Bot Service** - основной сервис бота
- **Message Parser** - парсинг сообщений пользователя
- **Category Matcher** - сопоставление описаний с категориями
- **State Manager** - управление состоянием диалога
- **Auth Manager** - управление аутентификацией
- **gRPC Client** - клиент для взаимодействия с основным сервисом

### База данных

Бот использует SQLite для хранения:
- Пользовательских сессий
- Состояний диалогов
- Сопоставлений категорий
- Пользовательских настроек

## 🔧 Разработка

### Структура проекта

```
budget-bot/
├── cmd/bot/                 # Точка входа приложения
├── internal/
│   ├── bot/                # Основная логика бота
│   ├── grpc/               # gRPC клиенты
│   ├── domain/             # Доменные модели
│   ├── repository/         # Репозитории для работы с БД
│   └── pkg/                # Общие пакеты
├── migrations/             # Миграции базы данных
├── proto/                  # Protobuf определения
└── configs/                # Конфигурационные файлы
```

### Тестирование

```bash
# Запуск всех тестов
go test ./...

# Запуск тестов с покрытием
go test -cover ./...

# Запуск конкретного теста
go test ./internal/bot -v
```

#### Анализ покрытия тестами

```bash
# Базовый анализ покрытия с рекомендациями
make coverage

# Детальный анализ с приоритетами для покрытия
make coverage-detail

# Создание HTML отчета для визуального анализа
make coverage-html
```

**Команды покрытия предоставляют:**

- **`make coverage`** - общий анализ покрытия с топ-10 функциями для тестирования
- **`make coverage-detail`** - детальные рекомендации по приоритетам покрытия
- **`make coverage-html`** - HTML отчет для просмотра покрытия по строкам кода

**Приоритеты покрытия:**
1. **Приоритет 1 (0-20%)** - функции с очень низким покрытием (высший приоритет)
2. **Приоритет 2 (21-50%)** - функции с низким покрытием
3. **Приоритет 3 (51-80%)** - функции со средним покрытием
4. **Приоритет 4 (>80%)** - функции с высоким покрытием (низкий приоритет)

### Сборка

```bash
# Локальная разработка
make build

# Продакшен сборка
make build-prod

# Сборка Docker образа
make docker-build
```

## 📊 Мониторинг

### Метрики Prometheus

При включенных метриках (`METRICS_ENABLED=true`) бот предоставляет метрики на порту `METRICS_ADDRESS`:

- `bot_transactions_total` - общее количество транзакций
- `bot_response_time_seconds` - время ответа бота
- `bot_active_users` - количество активных пользователей

### Логирование

Бот использует структурированное логирование с помощью Zap. Логи включают:
- Информацию о транзакциях
- Ошибки парсинга
- Проблемы с gRPC соединениями
- Действия пользователей

## 🔒 Безопасность

- JWT токены хранятся в зашифрованном виде
- Все соединения с gRPC сервером защищены
- Пользовательские данные изолированы между аккаунтами
   - Автоматическое обновление токенов

## 🤝 Поддержка

- **Документация**: 
  - `readme_quick_start.md` - быстрый старт
  - `readme_commands.md` - полный список команд бота
  - `readme_mappings.md` - система маппингов категорий
  - `readme_development_plan.md` - план разработки
- **Тестирование**: См. `readme_testing.md` для информации о тестировании
- **Issues**: Создавайте issues для багов и предложений

## 📄 Лицензия

[Укажите лицензию проекта]
