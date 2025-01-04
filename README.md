# Budget Bot

Telegram бот для учета доходов и расходов в Google Таблицах.

## Требования

- PHP 8.1+
- Docker
- Docker Compose
- Make

## Установка

1. Клонируйте репозиторий:
```bash
git clone git@github.com:your-username/budget-bot.git
cd budget-bot
```

2. Скопируйте файл `.env.example` в `.env` и настройте переменные окружения:
```bash
cp .env.example .env
```

Необходимо заполнить следующие переменные:
- `TELEGRAM_BOT_TOKEN` - токен вашего Telegram бота
- `TELEGRAM_BOT_USERNAME` - имя пользователя вашего бота
- `GOOGLE_SHEETS_CREDENTIALS_PATH` - путь к файлу с учетными данными Google Sheets API

3. Запустите сервисы:
```bash
make up
```

## Разработка

### Доступные команды

- `make up` - запуск сервисов
- `make down` - остановка сервисов
- `make build` - пересборка и запуск сервисов
- `make restart` - перезапуск сервисов
- `make tunnel` - создание SSH-туннеля для локальной разработки
- `make permissions` - установка прав доступа для директорий
- `make ci` - запуск всех проверок
- `make cs-check` - проверка стиля кода
- `make cs-fix` - исправление стиля кода
- `make phpstan` - статический анализ кода
- `make test` - запуск тестов

### Локальная разработка с использованием туннеля

Для локальной разработки можно использовать SSH-туннель, чтобы получать webhook-запросы от Telegram.

1. Настройте переменные окружения для туннеля в `.env`:
```
SSH_TUNNEL_HOST=your-server.com
SSH_TUNNEL_USER=username
SSH_TUNNEL_PORT=22
```

2. Запустите туннель:
```bash
make tunnel
```

3. Настройте webhook в Telegram на URL: `https://your-domain.com/webhook`

## Тестирование

Для запуска тестов используйте:
```bash
make test
```

## Проверка кода

Для проверки качества кода используйте:
```bash
make ci
```

Это запустит:
- PHP CS Fixer для проверки стиля кода
- PHPStan для статического анализа
- PHPUnit для тестов