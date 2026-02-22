# Добавление новой записи в `budget-bot`

## Цель
Документ описывает, как бот принимает сообщение пользователя, выбирает категорию, выполняет маппинг и отправляет запрос на создание транзакции в API сервиса бюджета.

## Основной сценарий
1. Пользователь отправляет текст в свободной форме, например: `1000 кофе` или `+50000 зарплата`.
2. `MessageParser` разбирает сообщение в структуру: тип (`income/expense`), сумма (`minor units`), валюта, описание, дата.
3. Бот подтягивает сессию OAuth и tenant (`OAuthManager.GetSession`).
4. Бот пытается автоматически определить категорию по маппингам (`CategoryMatcher.FindCategory`).
5. Если категория найдена, бот сразу вызывает gRPC `CreateTransaction`.
6. Если категория не найдена, бот показывает inline-клавиатуру категорий, ждет callback и только после выбора вызывает `CreateTransaction`.

Код: `internal/bot/handler.go`, `internal/bot/parser.go`, `internal/grpc/transaction_client.go`.

## Разбор сообщения
`internal/bot/parser.go`:
- Поддерживает даты (`сегодня`, `вчера`, `DD.MM`, `DD.MM.YYYY`).
- Определяет валюту (символы и коды) через `CurrencyParser`.
- Определяет тип операции:
- `+` -> `income`.
- без `+` и `-` -> `expense` (по умолчанию).
- `-` -> `expense`.
- Конвертирует сумму в `AmountMinor` (копейки/центы).
- Остаток строки становится `Description`.

Если валюта не указана, берется `default_currency` из `user_preferences`, иначе fallback `RUB`.

## Автоматическая привязка категории (маппинг)
`internal/bot/category_matcher.go` + `internal/repository/category_mapping_repo.go`:
1. Сначала точное совпадение по словам из описания (`keyword == word`).
2. Затем частичное совпадение (`description CONTAINS keyword`) среди всех маппингов tenant.
3. Если несколько частичных совпадений, выбирается запись с максимальным `priority`.

Хранилище маппингов:
- таблица `category_mappings` в SQLite (`tenant_id`, `keyword`, `category_id`, `priority`, `UNIQUE(tenant_id, keyword)`).

Управление маппингом из бота:
- `/map слово = название_категории`:
- бот резолвит `название_категории` в `category_id` через API (`CategoryNameMapper.GetCategoryIDByName` -> `ListCategories`).
- сохраняет `keyword -> category_id` в SQLite.
- `/map слово` и `/map --all` читают сохраненные маппинги.
- `/unmap слово` удаляет маппинг.

## Ручной выбор категории, если маппинг не найден
`internal/bot/handler.go`:
1. Бот запрашивает категории через API: `CategoryClient.ListCategories(tenant, token, transactionType, locale)`.
2. Показывает inline-клавиатуру категорий (`cat:<name>`).
3. Сохраняет контекст в `dialog_states` (`type`, `amount_minor`, `currency`, `desc`, `occurred_at`) и драфт в `transaction_drafts`.
4. По callback бот резолвит имя категории обратно в `category_id` (`CategoryNameMapper.GetCategoryIDByName`).
5. Формирует `CreateTransactionRequest` и вызывает `TransactionClient.CreateTransaction`.
6. После успеха очищает состояние (`dialog_states`).

Важно по текущей реализации:
- callback передает имя категории (`cat:<name>`), а затем делается повторный lookup имени в список категорий.
- у драфта создается `draft_id`, но при установке `dialog_states` `draft_id` сейчас явно не прокидывается в поле `draft_id` записи состояния; cleanup опирается на `context["draft_id"]`.

## Что уходит в API при создании
`internal/grpc/transaction_client.go`:
- В metadata передается `Authorization: Bearer <token>`.
- В `CreateTransactionRequest` отправляются:
- `type`
- `category_id`
- `amount { currency_code, minor_units }`
- `occurred_at`
- `comment` (описание из сообщения)

`tenant_id` в protobuf-запрос не уходит: tenant берется на стороне сервиса из auth/tenant контекста.

## Где категория может "сломаться"
1. Нет маппинга и API не вернул категории -> транзакция не создается.
2. В `/map` указано имя категории, которого нет в текущем locale -> маппинг не сохранится.
3. Тип операции и категория разного kind (income/expense) -> сервис отклонит запрос.
4. Категория удалена/деактивирована после сохранения маппинга -> сервис вернет ошибку при создании.

## Ключевые файлы
- `internal/bot/handler.go`
- `internal/bot/parser.go`
- `internal/bot/category_matcher.go`
- `internal/bot/category_name_mapper.go`
- `internal/grpc/category_client.go`
- `internal/grpc/transaction_client.go`
- `internal/repository/category_mapping_repo.go`
- `internal/repository/dialog_state_repo.go`
- `internal/repository/draft_repo.go`
- `migrations/0001_init.up.sql`
- `migrations/0002_drafts.up.sql`
