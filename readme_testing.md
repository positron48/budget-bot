## Тестирование: подход, структура и практики

### Уровни тестов
- Юнит‑тесты: без БД, без сети, тестируем чистую логику (парсинг, форматирование, матчинг ключевых слов).
- Интеграционные тесты: реальная SQLite, реальные миграции, gRPC клиенты через локальные fake‑сервера, Telegram эмулируется `httptest.Server`.

### Инструменты
- Go testing + `httptest` для HTTP/Telegram.
- Локальные fake gRPC‑сервера (реализуют интерфейс pb.*ServiceServer).
- SQLite in‑memory/временный файл. Схема поднимается реальными миграциями (или установочными DDL на время перехода).

### Best practices
- Изоляция: каждый тест использует отдельную БД (in‑memory или temp‑файл) и свои фабрики данных. Очистка через `t.Cleanup`.
- Подготовка данных: через репозитории/фабрики, а не raw SQL. Raw SQL допустим только для быстрой очистки.
- Детерминизм: никаких внешних сетей. Telegram/gRPC — внутри процесса.
- Структура теста: Given (prepare) → When (act) → Then (assert). Повторяющиеся шаги — выносить в helpers.

### Хелперы (план)
- `internal/testutil/sqlite.go`: `OpenMigratedSQLite(t)`, `CloseAndRemove(t, db)` — запуск миграций и закрытие DB.
- `internal/testutil/factory.go`: `SeedSession`, `SeedMapping`, `SeedPreferences`, `SeedDraft` — создают данные через репозитории.
- `internal/testutil/telegram.go`: `NewTestBot(t)` — поднимает `httptest.Server` и возвращает `*tgbotapi.BotAPI` с кастомным endpoint.
- `internal/testutil/ctx.go`: `Ctx(t)` — контекст с таймаутом.

### Покрытие
- Цель: ≥ 80% по проекту без учета сгенерированного `internal/pb` и исполняемого `cmd`.
- Запуск: `make coverage` (исключает `internal/pb` и `cmd`).
- В CI — добавить `coverage-check` (фейл, если < 80%).

### Что покрываем в интеграционных тестах
- `internal/bot/handler.go`:
  - `/start`, `/login` (email→password), `/register`, `/logout`
  - `/language` и `/currency` через callback
  - `/categories` (успех/ошибка)
  - `/map`, `/unmap`
  - `/switch_tenant`
  - Сообщение: parse → категория найдена → подтверждение yes/no
  - Сообщение: parse → категории нет → выбор → подтверждение yes
  - `/stats`, `/top_categories`, `/recent`, `/export` (через fake Report/Transaction)
  - Ошибочные сообщения → валидационный фидбек
- `internal/bot/parser.go`, `currency_parser.go`: крайние кейсы, отрицательные сценарии.
- `internal/bot/category_matcher.go`: приоритеты и смешанные совпадения.
- `internal/bot/ui/*`: форматирование и клавиатуры.
- `internal/grpc/*_client.go`: проверка метаданных (authorization), корректных маппингов.
- `internal/pkg/config`, `internal/pkg/logger`: значения по умолчанию, env overrides, уровни логирования.
- Репозитории: happy‑path и ошибки.

### Примеры
- Интеграционный тест handler: см. `internal/bot/handler_test.go` — локальный Telegram-мок, SQLite и фейковые gRPC клиенты.
- gRPC клиенты: см. `internal/grpc/*_test.go` — локальные fake‑сервера на `127.0.0.1:0`.
- UI: `internal/bot/ui/*_test.go` — проверка разметки клавиатур и форматирования.

### Дальнейшие шаги
1) Ввести `internal/testutil/*` и перевести тесты на миграции + фабрики.
2) Расширить тесты `handler` для всех команд/сценариев.
3) Добавить `coverage-check` и включить его в CI.


