# Budget Bot

[![CI](https://github.com/positron48/budget-bot/actions/workflows/ci.yml/badge.svg)](https://github.com/positron48/budget-bot/actions/workflows/ci.yml)
[![PHP Version](https://img.shields.io/badge/php-%3E%3D8.2-8892BF.svg)](https://php.net/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Telegram-бот на Symfony для учета расходов и доходов с использованием Google Spreadsheets.

## Возможности

- Отслеживание расходов и доходов через сообщения в Telegram
- Интеграция с Google Spreadsheets для хранения данных
- Автоматическое определение категорий на основе описания
- Поддержка нескольких пользователей
- Гибкие форматы ввода даты
- Пользовательские категории для каждого пользователя
- Команды для управления таблицами и категориями

## Требования

- Docker
- Docker Compose
- Google Sheets API credentials
- Telegram Bot Token

## Установка

1. Клонируйте репозиторий:
```bash
git clone https://github.com/positron48/budget-bot.git
cd budget-bot
```

2. Скопируйте файл окружения:
```bash
cp .env.example .env
```

3. Настройте переменные окружения в файле `.env`:

```env
# Symfony
APP_ENV=dev
APP_SECRET=your_secret_key_here # Замените на случайную строку

# База данных
DATABASE_URL="sqlite:///%kernel.project_dir%/var/data.db"

# Telegram
TELEGRAM_BOT_TOKEN=your_bot_token_here # Получите у @BotFather
TELEGRAM_BOT_USERNAME=your_bot_username # Имя вашего бота без @

# Google
GOOGLE_SHEETS_CREDENTIALS_PATH=%kernel.project_dir%/config/google_credentials.json
```

### Получение необходимых токенов

#### Telegram Bot Token:
1. Откройте @BotFather в Telegram
2. Отправьте команду `/newbot`
3. Следуйте инструкциям для создания бота
4. Скопируйте полученный токен в `TELEGRAM_BOT_TOKEN`
5. Имя бота (без @) укажите в `TELEGRAM_BOT_USERNAME`

#### Google Sheets API:
1. Перейдите в [Google Cloud Console](https://console.cloud.google.com/)
2. Создайте новый проект
3. Включите Google Sheets API
4. Создайте Service Account
5. Скачайте JSON с учетными данными
6. Сохраните файл как `config/google_credentials.json`

4. Соберите и запустите Docker контейнеры:
```bash
docker-compose up -d --build
```

5. Установите зависимости:
```bash
docker-compose run composer install
```

6. Создайте базу данных и выполните миграции:
```bash
docker-compose exec php bin/console doctrine:database:create
docker-compose exec php bin/console doctrine:migrations:migrate
```

7. Настройте webhook (замените на ваш домен):
```bash
docker-compose exec php bin/console app:set-webhook https://your-domain.com/webhook
```

## Использование

### Основные команды

- `/start` - Инициализация бота
- `/list` - Получить список доступных таблиц
- `/categories` - Управление категориями

### Добавление транзакций

Формат: `[дата] [+]сумма описание`

Примеры:
- `1000 продукты` - Добавляет расход 1000 за продукты сегодня
- `вчера 500 ресторан` - Добавляет расход за вчера
- `12.12 +5000 зарплата` - Добавляет доход за 12 декабря
- `1500.50 такси` - Добавляет расход с копейками

### Категории

По умолчанию доступны следующие категории:

#### Расходы:
- Питание
- Подарки
- Здоровье/медицина
- Дом
- Транспорт
- Личные расходы
- Домашние животные
- Коммунальные услуги
- Путешествия
- Одежда
- Развлечения
- Кафе/Ресторан
- Алко
- Образование
- Услуги
- Авто

#### Доходы:
- Зарплата
- Премия
- Кешбек, др. бонусы
- Процентный доход
- Инвестиции
- Другое

## Разработка

### Структура проекта

- `src/Command/` - Команды Telegram бота
- `src/Service/` - Основные сервисы (Google Sheets, определение категорий и т.д.)
- `src/Entity/` - Сущности базы данных
- `src/Repository/` - Репозитории для работы с БД

### Проверка качества кода

В проекте настроены следующие инструменты для проверки качества кода:

#### PHP CS Fixer
Проверка и исправление стиля кода согласно стандартам Symfony и PSR-12.

```bash
# Проверка стиля
composer cs-check

# Исправление стиля
composer cs-fix
```

#### PHPStan
Статический анализ кода с максимальным уровнем строгости (level 8).

```bash
composer phpstan
```

#### PHPUnit
Модульное тестирование.

```bash
composer test
```

#### Запуск всех проверок

```bash
composer check-all
```

Все проверки автоматически запускаются в GitHub Actions при:
- Push в ветку main
- Создании Pull Request

### Добавление категорий

Каждый пользователь может иметь свои собственные категории и ключевые слова для их определения.

## Contributing

1. Создайте форк репозитория
2. Создайте ветку для новой функциональности
3. Внесите изменения
4. Убедитесь, что все проверки проходят успешно (`composer check-all`)
5. Отправьте pull request

## Лицензия

Этот проект распространяется под лицензией MIT.