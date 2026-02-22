# План доработки `budget-bot`: LLM-подбор категории, управление запоминанием, смена категории

## 0. Цель изменений
Сделать процесс выбора категории при добавлении транзакции управляемым и прозрачным:
1. Если маппинг не найден, пробовать подобрать категорию через LLM (OpenRouter) из списка категорий tenant.
2. Автоматически применять LLM-категорию только при `probability >= 50%`.
3. После автоприменения показывать кнопки `Запомнить выбор` и `Сменить категорию`.
4. Для категории, примененной из маппинга, показывать `Забыть выбор` и `Сменить категорию`.
5. Для ручного выбора также поддержать `Запомнить выбор` (по точному описанию) и `Сменить категорию`.
6. После выбора категории удалять предыдущее сообщение со списком категорий и выводить компактное подтверждение с выбранной категорией и кнопкой `Сменить категорию`.

Документ описывает реализацию так, чтобы можно было выполнить работу без дополнительного контекста.

---

## 1. Что есть сейчас (база)

### 1.1 Основной флоу
- `internal/bot/handler.go`:
- парсинг сообщения -> попытка маппинга через `CategoryMatcher`.
- если маппинг есть, сразу `CreateTransaction`.
- если нет маппинга, показывается inline-список категорий и ожидание callback `cat:<name>`.

### 1.2 Текущие компоненты
- `MessageParser` (`internal/bot/parser.go`) разбирает сумму/тип/дату/описание.
- `CategoryMatcher` (`internal/bot/category_matcher.go`) работает по `category_mappings` SQLite.
- `CategoryNameMapper` (`internal/bot/category_name_mapper.go`) маппит name <-> id через API категорий.
- `CategoryClient` (`internal/grpc/category_client.go`) получает список категорий с сервиса.
- `TransactionClient` (`internal/grpc/transaction_client.go`) создает транзакцию.
- `DialogStateRepository` и `DraftRepository` хранят контекст выбора категории.

### 1.3 Ограничения текущей реализации
- Нет LLM fallback.
- Callback хранит имя категории, не id (`cat:<name>`), что хрупко при локали/переименовании.
- Нет единого post-selection UI с `запомнить/забыть/сменить`.
- Не управляется жизненный цикл message-id сообщения со списком категорий.

---

## 2. Целевая архитектура

### 2.1 Новый источник автоподбора
Добавить слой `LLMCategorySuggester`:
- вход: `tenant_id`, `tx_type`, `description`, `locale`, список категорий (`id`, `name`, `kind`).
- выход: `suggested_category_id`, `suggested_category_name`, `probability`, `reason`.

Источник LLM:
- HTTP вызов OpenRouter.
- Модель задается в `.env`.

### 2.2 Приоритет выбора категории
При создании транзакции в боте:
1. `Mapping` (локальный `category_mappings`) - приоритет #1.
2. `LLM` (если mapping не найден) - приоритет #2, применять только если `probability >= 0.5`.
3. `Manual` (клавиатура категорий) - fallback.

### 2.3 Единый post-selection state
Ввести расширенный контекст выбора для конкретной попытки добавления:
- `selection_source`: `mapping | llm | manual`
- `description_original`: точный текст для последующего запоминания
- `category_id_selected`
- `category_name_selected`
- `tx_payload` (type/amount/currency/occurred_at/comment)
- `category_list_message_id` (если отправляли список)
- `confirmation_message_id` (опционально)
- `can_remember` / `can_forget` (производные флаги)

### 2.4 Callback actions
Стандартизировать callback data (с version-префиксом):
- `v1:cat_select:<category_id>`
- `v1:remember:<draft_or_op_id>`
- `v1:forget:<draft_or_op_id>`
- `v1:change:<draft_or_op_id>`

---

## 3. Изменения по слоям

## 3.1 Конфигурация
Файлы:
- `env.example`
- `internal/pkg/config/config.go`
- `internal/pkg/config/config_test.go`

Добавить новые env:
- `OPENROUTER_API_KEY` - ключ OpenRouter.
- `OPENROUTER_MODEL` - имя модели (пользователь задает сам).
- `OPENROUTER_BASE_URL` - default `https://openrouter.ai/api/v1`.
- `OPENROUTER_TIMEOUT` - timeout HTTP (например `10s`).
- `OPENROUTER_ENABLE` - feature-flag (`true/false`).



Требования:
- если `OPENROUTER_ENABLE=false`, LLM шаг пропускается.
- если включено, но не хватает ключа/модели, логируем warning и fallback на ручной выбор.

## 3.2 Новый LLM-клиент
Новые файлы (предложение):
- `internal/llm/openrouter_client.go`
- `internal/llm/types.go`
- `internal/llm/openrouter_client_test.go`

Контракт интерфейса:
- `SuggestCategory(ctx, req) (resp, error)`.

`req`:
- `description`
- `transaction_type`
- `locale`
- `categories[]: {id,name}`

`resp`:
- `category_id`
- `probability` (0..1)
- `reason` (short text)

Формат ответа LLM:
- Требовать строго JSON без markdown.
- Пример:
```json
{"category_id":"cat-food","probability":0.82,"reason":"кофе и еда"}
```

Валидация после ответа:
- `category_id` обязан быть в переданном списке категорий.
- `probability` clamp в [0,1], invalid -> ошибка.
- при ошибке парсинга -> fallback.

## 3.3 Промпт для LLM
Новый файл:
- `prompts.md` (добавить секцию) или `internal/llm/prompt_builder.go`.

Содержимое промпта:
- Явно описать задачу классификации описания транзакции.
- Передать список доступных категорий текущего tenant и текущего типа (`income` или `expense`).
- Попросить вернуть только JSON.
- Попросить оценку уверенности (вероятность) с учетом неоднозначности.
- Указать правило: если не уверен, вероятность должна быть низкой.

Пример логики:
- System: "Ты классификатор банковских/бытовых транзакций".
- User: описание + категории + желаемый JSON-формат.

## 3.4 Доработка `handler.go` (главная)

### 3.4.1 Изменение флоу в `HandleUpdate`
Текущий блок `if parsed != nil && parsed.IsValid` переписать в последовательность:
1. Подготовка `tx draft payload`.
2. Попытка `mapping`.
3. Если mapping найден:
- выбрать категорию.
- отправить/обновить сообщение подтверждения с кнопками:
- `Забыть выбор`
- `Сменить категорию`
- создать транзакцию (как сейчас).
4. Если mapping не найден:
- если LLM включен: сделать suggest.
- если `probability >= 0.5`:
- автоприменить категорию.
- создать транзакцию.
- показать подтверждение с кнопками:
- `Запомнить выбор`
- `Сменить категорию`
- если `< 0.5` или ошибка:
- отправить сообщение: "Автоматически определить категорию не получилось".
- показать список категорий для ручного выбора.

### 3.4.2 Поддержка ручного выбора
При ручном выборе (`v1:cat_select:<id>`):
- взять категорию по id (без name->id lookup).
- создать транзакцию.
- удалить сообщение со списком категорий (если есть `category_list_message_id`).
- отправить подтверждение:
- выбранная категория
- кнопки: `Запомнить выбор` и `Сменить категорию`

### 3.4.3 Логика remember/forget
`Запомнить выбор`:
- добавить mapping:
- `keyword = description_original` (в точности, как ввел пользователь, trim only)
- `category_id = category_id_selected`
- `tenant_id = session.tenant_id`
- после успеха обновить/перерисовать кнопки (можно скрыть кнопку remember и оставить `Сменить категорию`).

`Забыть выбор`:
- удалить mapping по `keyword = description_original` (точный текст, используемый при remember).
- после успеха обновить кнопки (можно показать `Запомнить выбор` обратно или оставить только `Сменить категорию`).

Важно:
- в текущей БД `category_mappings.keyword` уникален в рамках tenant.
- для точного соответствия учитывать регистр/пробелы согласно выбранной нормализации (рекомендуется: хранить как есть, но при поиске оставить текущую lowercasing-логику matcher, либо нормализовать при сохранении консистентно).

### 3.4.4 Смена категории для конкретной записи
Поведение кнопки `Сменить категорию`:
1. Открыть список категорий для конкретной уже созданной транзакции.
2. После выбора категории:
- вызвать API обновления транзакции (`UpdateTransaction`) c новым `category_id`.
3. Удалить старое сообщение со списком.
4. Обновить подтверждение с новой категорией и той же кнопкой `Сменить категорию`.

Следствие:
- боту нужен клиент `UpdateTransaction` (сейчас есть только Create/List).

## 3.5 gRPC-клиент транзакций
Файл:
- `internal/grpc/transaction_client.go`

Добавить:
- `UpdateTransactionCategory(ctx, txID, categoryID, accessToken) error`.

Реализация:
- `pb.UpdateTransactionRequest`:
- `id = txID`
- `transaction.category_id = new_id`
- `update_mask.paths = ["category_id"]`

Также для работы "сменить категорию" нужен `tx_id` после create:
- сохранить возвращенный `transaction_id` в контексте операции, чтобы потом можно было обновить.

## 3.6 UI-клавиатуры
Файл:
- `internal/bot/ui/keyboards.go`

Добавить фабрики:
- `CreatePostSelectionKeyboard(source, opID)`: 
- source `mapping` -> `Забыть выбор` + `Сменить категорию`.
- source `llm|manual` -> `Запомнить выбор` + `Сменить категорию`.
- `CreateChangeCategoryKeyboard(categories, opID)` c callback `v1:cat_select:<category_id>:<opID>`.

## 3.7 Хранилище контекста операции
Текущий `dialog_states` используется только для шагов выбора.
Для надежного управления кнопками после создания транзакции нужен отдельный storage.

Новая сущность (рекомендуется): `operation_contexts` в SQLite.

Миграция `migrations/0003_operation_contexts.up.sql`:
- `op_id TEXT PRIMARY KEY`
- `telegram_id INTEGER NOT NULL`
- `tenant_id TEXT NOT NULL`
- `transaction_id TEXT` (nullable до create)
- `description_original TEXT NOT NULL`
- `category_id_selected TEXT`
- `category_name_selected TEXT`
- `selection_source TEXT NOT NULL` (`mapping|llm|manual`)
- `tx_type TEXT NOT NULL`
- `amount_minor INTEGER NOT NULL`
- `currency TEXT NOT NULL`
- `occurred_at TIMESTAMP`
- `category_list_message_id INTEGER`
- `confirmation_message_id INTEGER`
- `created_at`, `updated_at`

Репозиторий:
- `internal/repository/operation_context_repo.go`

Почему так:
- callback-кнопки асинхронны, нужны устойчивые данные для remember/forget/change.
- не перегружает `dialog_states` временными ключами.

## 3.8 Удаление сообщения со списком категорий
Требование: после любого выбора категории удалить предыдущее сообщение со списком.

Реализация:
- при отправке списка категорий сохранять `message_id` в `operation_context`.
- после выбора (manual/llm/mapping при наличии списка) вызывать `DeleteMessage(chat_id, message_id)`.
- если удаление не удалось (expired, not found) - только логируем, без срыва сценария.

## 3.9 Поведение текста сообщений
Стандартизировать UX тексты:
- если LLM < 50%: 
- `"Категорию автоматически определить не получилось. Выберите вручную:"`
- при автоприменении:
- `"Выбрана категория: <name>"`
- при mapping:
- `"Применено сохраненное сопоставление: <name>"`
- после смены категории:
- `"Категория обновлена: <name>"`

И всегда с кнопкой `Сменить категорию`.

## 3.10 Логирование и метрики
- Логи:
- источник категории (`mapping/llm/manual`)
- вероятность LLM
- факт auto-apply (`true/false`)
- действия кнопок (`remember/forget/change`)

- Метрики (при наличии общего паттерна):
- `bot_category_selected_total{source=...}`
- `bot_llm_suggestion_total{result=applied|rejected|error}`
- `bot_mapping_mutation_total{action=remember|forget}`

---

## 4. Пошаговый план внедрения (очередность)

### Этап 1: инфраструктура LLM
1. Добавить env + config parsing + тесты.
2. Реализовать OpenRouter client + typed response + тесты парсинга/валидации.
3. Добавить prompt builder и golden tests для prompt output.

### Этап 2: callback-протокол и хранилище операций
1. Ввести новый callback формат `v1:*`.
2. Добавить `operation_contexts` (миграция + repo + тесты).
3. Прокинуть `op_id` во все релевантные callback.

### Этап 3: новый флоу выбора категории
1. В `HandleUpdate` реализовать приоритет `mapping -> llm -> manual`.
2. Добавить порог `0.5` и сообщение о неуспехе автоопределения.
3. Реализовать удаление сообщения списка категорий после выбора.

### Этап 4: remember/forget/change
1. Добавить кнопки post-selection.
2. Реализовать `remember` (точный `description_original`).
3. Реализовать `forget`.
4. Реализовать `change category` через `UpdateTransaction`.

### Этап 5: полировка
1. Единые тексты ответов и локализация (если нужно).
2. Логи/метрики.
3. Рефакторинг handler на мелкие функции (чтобы избежать регрессий).

---

## 5. Изменяемые файлы (чеклист)

Обязательно:
- `env.example`
- `internal/pkg/config/config.go`
- `internal/pkg/config/config_test.go`
- `internal/bot/handler.go`
- `internal/bot/ui/keyboards.go`
- `internal/grpc/transaction_client.go`
- `internal/repository/*` (новый repo operation context)
- `migrations/0003_operation_contexts.up.sql`
- `migrations/0003_operation_contexts.down.sql`

Новые:
- `internal/llm/openrouter_client.go`
- `internal/llm/types.go`
- `internal/llm/prompt_builder.go`
- `internal/llm/openrouter_client_test.go`
- `internal/repository/operation_context_repo.go`
- тесты по новым сценариям в `internal/bot/*_test.go`

Опционально:
- `prompts.md` (описание прод-подсказки)
- `docs/readme_commands.md` (обновить пользовательское поведение)

---

## 6. Тест-план

## 6.1 Unit
1. LLM client:
- happy path валидный JSON.
- невалидный JSON.
- category_id вне списка.
- probability отсутствует/вне диапазона.

2. Handler category flow:
- mapping found -> create -> кнопки `forget+change`.
- mapping miss + llm prob 0.9 -> auto create -> `remember+change`.
- mapping miss + llm prob 0.3 -> сообщение о неуспехе -> ручной список.
- manual select -> create -> удаление списка -> `remember+change`.
- remember -> запись mapping с точным description.
- forget -> удаление mapping.
- change -> update transaction category.

3. UI keyboards:
- корректные callback data для всех кнопок.

## 6.2 Integration-ish (с fake clients)
1. Полный сценарий: ввод -> llm low confidence -> ручной выбор -> remember -> следующий ввод автопопадание mapping.
2. Сценарий mapping -> forget -> следующий ввод больше не матчит mapping.
3. Сценарий change category после create.

## 6.3 Regression
- старые команды `/map`, `/unmap`, `/categories`, `/recent`, `/stats`, `/export` не ломаются.

---

## 7. Риски и решения
1. Риск: LLM задержка увеличит время ответа.
- Решение: timeout + fallback на ручной выбор.

2. Риск: неоднозначные категории/локали.
- Решение: LLM всегда получает категории только нужного `transactionType` и текущей локали.

3. Риск: callback data превышает лимит Telegram.
- Решение: хранить все тяжелые данные в `operation_contexts`, в callback передавать короткий `op_id`.

4. Риск: race при двойном нажатии кнопок.
- Решение: optimistic lock через `updated_at` или idempotency-проверки (если действие уже применено, возвращать мягкий ответ).

5. Риск: удаление старого message может падать.
- Решение: best-effort delete, не блокировать бизнес-флоу.

---

## 8. Критерии готовности (Definition of Done)
1. Включенный OpenRouter корректно участвует в выборе категории при отсутствии mapping.
2. Порог 50% строго соблюдается.
3. Кнопки `Запомнить выбор/Забыть выбор/Сменить категорию` работают согласно источнику выбора.
4. После выбора категории сообщение со списком категорий удаляется (best effort), отображается актуальная категория и `Сменить категорию`.
5. Ручная смена категории изменяет категорию уже созданной транзакции через API.
6. Добавлены тесты на новые сценарии и они проходят.
7. `env.example` и документация обновлены.

---

## 9. Минимальный MVP и расширения

MVP:
- LLM fallback + threshold + remember/forget + change category + delete list message.

Расширения после MVP:
1. Кеш LLM ответов по `(tenant, type, normalized description)`.
2. A/B разных моделей OpenRouter.
3. Explainability в UI (почему выбрана категория).
4. Автозапоминание при высокой вероятности (например >= 0.9) через флаг.

