# Budget Bot

[![CI](https://github.com/positron48/budget-bot/actions/workflows/ci.yml/badge.svg)](https://github.com/positron48/budget-bot/actions/workflows/ci.yml)
![Coverage](https://raw.githubusercontent.com/positron48/budget-bot/master/badge-coverage.svg)
[![PHPStan](https://img.shields.io/badge/PHPStan-level%208-brightgreen.svg?style=flat)](https://phpstan.org/)
[![PHP Version](https://img.shields.io/badge/php-%3E%3D8.2-8892BF.svg)](https://www.php.net/)

Telegram бот для учета расходов и доходов в Google таблицах.

## Установка

1. Склонируйте репозиторий
2. Скопируйте `.env.example` в `.env` и заполните необходимые переменные окружения
3. Запустите `docker-compose up -d`
4. Выполните миграции: `docker-compose exec php bin/console doctrine:migrations:migrate`

## Разработка

### Тесты

```bash
make test
```

### Code Style

```bash
make cs-fix
```

### Статический анализ

```bash
make phpstan
```

### CI

```bash
make ci
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

## Разработка с помощью Cursor

Этот проект полностью разработан с использованием [Cursor](https://cursor.sh/) - AI-powered IDE. Все файлы проекта, кроме `cursor-log.md` и файлов с секретами, были написаны с помощью Cursor.

История разработки и все запросы к Cursor доступны в файле [cursor-log.md](cursor-log.md).

## Использование бота

### Настройка Google Sheets

1. Создайте проект в [Google Cloud Console](https://console.cloud.google.com/)
2. Включите Google Sheets API для проекта
3. Создайте Service Account и скачайте JSON с учетными данными
4. Поместите JSON файл в `config/credentials/google-sheets.json`
5. Создайте новую Google таблицу для учета расходов со следующей структурой на листе "Транзакции":
   - Строка 1: заголовки (любые)
   - Колонки для расходов:
     - B: Дата
     - C: Сумма
     - D: Описание
     - E: Категория
   - Колонки для доходов:
     - G: Дата
     - H: Сумма
     - I: Описание
     - J: Категория
6. Предоставьте доступ к таблице для Service Account (email указан в JSON файле)

### Команды бота

- `/start` - начало работы с ботом
- `/list` - список доступных таблиц
- `/categories` - список доступных категорий

### Формат сообщений

Бот принимает сообщения в следующем формате:
```
[дата] [+]сумма описание
```

Где:
- `[дата]` - необязательное поле, может быть:
  - Пропущено (будет использована текущая дата)
  - "сегодня"
  - "вчера"
  - В формате "DD.MM.YYYY" или "DD.MM"
- `[+]` - необязательный знак для доходов
- `сумма` - число с точкой или запятой для копеек
- `описание` - текстовое описание транзакции

Примеры:
```
1000 продукты
+50000 зарплата
01.12 5000 подарок
12.12.2023 1234.56 такси
вчера 100 кофе
сегодня +1000 возврат
```

### Категории

Бот автоматически определяет категорию по описанию. Если категория не определена, бот предложит выбрать её из списка.

Доступные категории по умолчанию:

Расходы:
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

Доходы:
- Зарплата
- Премия
- Кешбек, др. бонусы
- Процентный доход
- Инвестиции
- Другое

### Работа с таблицами

Для начала работы с ботом нужно добавить существующую таблицу Google Sheets:

1. Подготовьте таблицу со следующей структурой:
   - Лист "Транзакции":
     - Колонки для расходов:
       - B: Дата
       - C: Сумма
       - D: Описание
       - E: Категория
     - Колонки для доходов:
       - G: Дата
       - H: Сумма
       - I: Описание
       - J: Категория
   - Лист "Сводка":
     - Строка 28: начало списка категорий
     - Колонка B: категории расходов
     - Колонка H: категории доходов

2. Добавьте таблицу в бота:
   - Используйте команду `/add`
   - Отправьте ID таблицы (можно найти в URL: https://docs.google.com/spreadsheets/d/ID/...)
   - Если у бота нет доступа, он отправит ссылку для предоставления доступа и email сервисного аккаунта
   - После получения доступа, укажите за какой месяц эта таблица

3. Переключение между таблицами:
   - Используйте команду `/list` для просмотра списка доступных таблиц
   - Выберите нужную таблицу из списка