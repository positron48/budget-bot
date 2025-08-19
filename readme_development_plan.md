# 📌 Прогресс (кратко)

- Этап 1 (база) — выполнен, кроме golangci-lint
- Этап 2 (аутентификация/организации/настройки) — реализованы `/start`, `/login`, `/register`, `/logout`, `/switch_tenant`, `/language`, `/currency`
- Этап 3 (парсер) — реализован, есть unit‑тесты
- Этап 4 (категоризация) — команды `/categories`, `/map`, `/unmap` + сопоставления
- Этап 5 (состояние) — state manager, черновики, подтверждение, `/cancel`
- Этап 6 (UI) — форматтер сообщений и клавиатуры
- Этап 7 (статистика) — `/stats` (Report.GetMonthlySummary), `/top_categories` (по сводке), `/recent` (Transaction.ListTransactions сортировка по дате desc)
- В процессе/дальше: `/export`, golangci-lint, интеграционные и e2e тесты

---

# Telegram Bot - Пошаговый план разработки

## 📅 Общий план (10-12 недель)

### Этап 1: Базовая инфраструктура (1-2 недели)
### Этап 2: Аутентификация и пользователи (1 неделя)
### Этап 3: Парсинг сообщений (1 неделя)
### Этап 4: Система категоризации (1-2 недели)
### Этап 5: Управление состоянием (1 неделя)
### Этап 6: Пользовательский интерфейс (1 неделя)
### Этап 7: Статистика и аналитика (1 неделя)
### Этап 8: Тестирование и оптимизация (1 неделя)

---

## 🏗️ Этап 1: Базовая инфраструктура (1-2 недели)

### Неделя 1: Настройка проекта

#### День 1-2: Инициализация проекта
- [ ] Создание структуры проекта `telegram-bot/`
- [ ] Инициализация Go модуля (`go mod init`)
- [ ] Настройка зависимостей в `go.mod`:
  ```go
  require (
      github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.5.1
      google.golang.org/grpc v1.62.0
      google.golang.org/protobuf v1.33.0
      github.com/spf13/viper v1.18.2
      go.uber.org/zap v1.26.0
      github.com/prometheus/client_golang v1.18.0
      github.com/mattn/go-sqlite3 v1.14.19
      github.com/golang-migrate/migrate/v4 v4.17.0
  )
  ```
- [ ] Создание базовой структуры директорий
- [ ] Настройка `.gitignore`
- [ ] Создание `Makefile` с базовыми командами

#### День 3-4: Конфигурация и логирование
- [ ] Создание `internal/pkg/config/config.go`:
  ```go
  type Config struct {
      Telegram TelegramConfig
      GRPC     GRPCConfig
      Database DatabaseConfig
      Logging  LoggingConfig
      Metrics  MetricsConfig
  }
  ```
- [ ] Интеграция с Viper для загрузки конфигурации
- [ ] Настройка структурированного логирования с Zap
- [ ] Создание конфигурационных файлов (`configs/config.yaml`, `.env.example`)

#### День 5: Docker и CI/CD
- [ ] Создание `Dockerfile` для многоэтапной сборки
- [ ] Настройка `docker-compose.yml` для локальной разработки
- [ ] Создание GitHub Actions workflow (`.github/workflows/ci.yml`)
- [ ] Настройка линтеров (golangci-lint)

### Неделя 2: База данных и gRPC клиенты

#### День 1-2: Схема базы данных
- [ ] Создание `migrations/0001_init.up.sql`:
  ```sql
  CREATE TABLE user_sessions (
      telegram_id BIGINT PRIMARY KEY,
      user_id UUID NOT NULL,
      tenant_id UUID NOT NULL,
      access_token TEXT NOT NULL,
      refresh_token TEXT NOT NULL,
      access_token_expires_at TIMESTAMP NOT NULL,
      refresh_token_expires_at TIMESTAMP NOT NULL,
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
      updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  );

  CREATE TABLE category_mappings (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      tenant_id UUID NOT NULL,
      keyword TEXT NOT NULL,
      category_id UUID NOT NULL,
      priority INTEGER DEFAULT 0,
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
      UNIQUE(tenant_id, keyword)
  );

  CREATE TABLE dialog_states (
    telegram_id BIGINT PRIMARY KEY,
    state TEXT NOT NULL,
    draft_id UUID,
    context JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE user_preferences (
    telegram_id BIGINT PRIMARY KEY,
    language TEXT DEFAULT 'ru',
    default_currency TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
  ```
- [ ] Создание `migrations/0001_init.down.sql`
- [ ] Настройка golang-migrate для управления миграциями

#### День 3-4: Репозитории
- [ ] Создание `internal/repository/session_repo.go`:
  ```go
  type SessionRepository interface {
      SaveSession(ctx context.Context, session *UserSession) error
      GetSession(ctx context.Context, telegramID int64) (*UserSession, error)
      DeleteSession(ctx context.Context, telegramID int64) error
      UpdateTokens(ctx context.Context, telegramID int64, tokens *TokenPair) error
  }
  ```
- [ ] Создание `internal/repository/category_mapping_repo.go`
- [ ] Создание `internal/repository/user_repo.go`
- [ ] Создание `internal/repository/preferences_repo.go`:
  ```go
  type PreferencesRepository interface {
      SavePreferences(ctx context.Context, preferences *UserPreferences) error
      GetPreferences(ctx context.Context, telegramID int64) (*UserPreferences, error)
      UpdateLanguage(ctx context.Context, telegramID int64, language string) error
      UpdateDefaultCurrency(ctx context.Context, telegramID int64, currency string) error
  }
  ```
- [ ] Реализация SQLite адаптеров для всех репозиториев

#### День 5: gRPC клиенты
- [ ] Копирование proto файлов из основного проекта https://github.com/positron48/budget/tree/master/proto
- [ ] Генерация Go кода из proto файлов
- [ ] Создание `internal/grpc/client.go` - базовый gRPC клиент
- [ ] Создание `internal/grpc/auth_client.go`:
  ```go
  type AuthClient interface {
      Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error)
      Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error)
      RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error)
  }
  ```

---

## 🔐 Этап 2: Аутентификация и пользователи (1 неделя)

### Неделя 3: Система аутентификации

#### День 1-2: Auth Manager
- [ ] Создание `internal/bot/auth_manager.go`:
  ```go
  type AuthManager struct {
      authClient    grpc.AuthClient
      sessionRepo   repository.SessionRepository
      logger        *zap.Logger
  }

  func (am *AuthManager) Register(ctx context.Context, telegramID int64, email, password, name string) error
  func (am *AuthManager) Login(ctx context.Context, telegramID int64, email, password string) error
  func (am *AuthManager) Logout(ctx context.Context, telegramID int64) error
  func (am *AuthManager) GetSession(ctx context.Context, telegramID int64) (*UserSession, error)
  func (am *AuthManager) RefreshTokens(ctx context.Context, telegramID int64) error
  ```
- [ ] Реализация регистрации через бота
- [ ] Реализация входа в систему
- [ ] Управление JWT токенами
- [ ] Выбор языка при первом входе
- [ ] Сохранение языковых предпочтений

#### День 3-4: Обработчики команд аутентификации
- [ ] Создание `internal/bot/handler.go` - основной обработчик
- [ ] Реализация команды `/start`:
  ```go
  func (h *Handler) handleStart(ctx context.Context, update tgbotapi.Update) error {
      // Проверка существующей сессии
      // Предложение регистрации/входа
      // Отправка приветственного сообщения
  }
  ```
- [ ] Реализация команды `/login` с диалогом ввода email/password
- [ ] Реализация команды `/register` с диалогом регистрации
- [ ] Реализация команды `/logout`

#### День 5: Управление организациями и настройки
- [ ] Создание `internal/grpc/tenant_client.go`
- [ ] Реализация команды `/switch_tenant` для переключения между организациями
- [ ] Получение списка доступных организаций
- [ ] Сохранение выбранной организации в сессии
- [ ] Реализация команды `/language` для выбора языка
- [ ] Реализация команды `/currency` для настройки валюты
- [ ] Получение валюты тенанта по умолчанию

---

## 💬 Этап 3: Парсинг сообщений (1 неделя)

### Неделя 4: Парсер сообщений

#### День 1-2: Базовый парсер
- [ ] Создание `internal/bot/parser.go`:
  ```go
  type MessageParser struct {
      logger *zap.Logger
  }

  type ParsedTransaction struct {
      Type        TransactionType
      Amount      *Money
      Currency    string
      Description string
      OccurredAt  *time.Time
      IsValid     bool
      Errors      []string
  }

  func (p *MessageParser) ParseMessage(text string) (*ParsedTransaction, error)
  ```
- [ ] Реализация парсинга дат:
  - Ключевые слова: "сегодня", "вчера", "позавчера"
  - Форматы: "DD.MM.YYYY", "DD.MM"
  - Валидация корректности дат
- [ ] Реализация парсинга сумм:
  - Поиск чисел с точкой/запятой
  - Определение типа транзакции по знаку "+"
  - Конвертация в minor units
- [ ] Реализация парсинга валют:
  - Поиск символов валют (₽, $, €, £, ¥)
  - Поиск кодов валют (RUB, USD, EUR, GBP, JPY)
  - Валидация корректности валют

#### День 3-4: Валидация и обработка ошибок
- [ ] Создание системы валидации:
  ```go
  type ValidationError struct {
      Field   string
      Message string
  }

  func (p *MessageParser) Validate(parsed *ParsedTransaction) []ValidationError
  ```
- [ ] Обработка различных форматов ввода
- [ ] Поддержка отрицательных сумм
- [ ] Обработка специальных символов и эмодзи
- [ ] Парсинг валют (символы и коды)
- [ ] Валидация корректности валют

#### День 5: Интеграция с обработчиком
- [ ] Интеграция парсера в основной обработчик
- [ ] Обработка сообщений в формате транзакций
- [ ] Создание черновиков транзакций
- [ ] Тестирование различных форматов сообщений

---

## 🏷️ Этап 4: Система категоризации (1-2 недели)

### Неделя 5: База сопоставлений

#### День 1-2: Category Matcher
- [ ] Создание `internal/bot/category_matcher.go`:
  ```go
  type CategoryMatcher struct {
      categoryClient     grpc.CategoryClient
      mappingRepo        repository.CategoryMappingRepository
      logger            *zap.Logger
  }

  func (cm *CategoryMatcher) FindCategory(ctx context.Context, tenantID, description string) (*Category, error)
  func (cm *CategoryMatcher) AddMapping(ctx context.Context, tenantID, keyword, categoryID string) error
  func (cm *CategoryMatcher) RemoveMapping(ctx context.Context, tenantID, keyword string) error
  ```
- [ ] Алгоритм поиска категорий:
  - Точные совпадения ключевых слов
  - Частичные совпадения (подстроки)
  - Учет приоритета сопоставлений
  - Возврат наиболее подходящей категории

#### День 3-4: Команды управления категориями
- [ ] Реализация команды `/categories`:
  ```go
  func (h *Handler) handleCategories(ctx context.Context, update tgbotapi.Update) error {
      // Получение списка категорий через gRPC
      // Форматирование списка с эмодзи
      // Отправка inline клавиатуры для выбора
  }
  ```
- [ ] Реализация команды `/map`:
  - `/map слово = категория` - добавление сопоставления
  - `/map слово` - показ текущего сопоставления
  - `/map --all` - показ всех сопоставлений
- [ ] Реализация команды `/unmap` для удаления сопоставлений

#### День 5: Умные подсказки
- [ ] Анализ истории транзакций для предложения категорий
- [ ] Контекстные подсказки на основе времени и дня недели
- [ ] Кеширование часто используемых категорий

### Неделя 6: Мультивалютность и конвертация

#### День 1-2: Конвертация валют
- [ ] Создание `internal/bot/currency_converter.go`:
  ```go
  type CurrencyConverter struct {
      fxClient grpc.FxClient
      logger   *zap.Logger
  }

  func (cc *CurrencyConverter) ConvertToBaseCurrency(amount *Money, fromCurrency, toCurrency string, date time.Time) (*Money, error)
  func (cc *CurrencyConverter) GetExchangeRate(fromCurrency, toCurrency string, date time.Time) (float64, error)
  ```
- [ ] Интеграция с gRPC Fx Service
- [ ] Кеширование курсов валют
- [ ] Обработка ошибок конвертации

#### День 3-4: Парсинг валют
- [ ] Создание `internal/bot/currency_parser.go`:
  ```go
  type CurrencyParser struct {
      symbolToCode map[string]string
      codeToSymbol map[string]string
  }

  func (cp *CurrencyParser) ParseCurrency(text string) (string, string, error)
  func (cp *CurrencyParser) ValidateCurrency(currency string) bool
  ```
- [ ] Поддержка символов валют (₽, $, €, £, ¥)
- [ ] Поддержка кодов валют (RUB, USD, EUR, GBP, JPY)
- [ ] Валидация корректности валют

#### День 5: Интеграция с парсером
- [ ] Интеграция парсера валют в основной парсер сообщений
- [ ] Обновление структуры ParsedTransaction для поддержки валют
- [ ] Тестирование парсинга сообщений с валютой

### Неделя 7: Продвинутая категоризация

#### День 1-2: Машинное обучение (опционально)
- [ ] Интеграция с простыми ML алгоритмами для улучшения точности
- [ ] Анализ паттернов в описаниях транзакций
- [ ] Обучение на исторических данных

#### День 3-4: Управление категориями
- [ ] Создание новых категорий через бота
- [ ] Редактирование существующих категорий
- [ ] Синхронизация категорий между ботом и основной системой

#### День 5: Тестирование и оптимизация
- [ ] Тестирование точности категоризации
- [ ] Оптимизация алгоритмов поиска
- [ ] Настройка производительности

---

## 🔄 Этап 5: Управление состоянием (1 неделя)

### Неделя 8: State Manager

#### День 1-2: Система состояний
- [ ] Создание `internal/bot/state_manager.go`:
  ```go
  type StateManager struct {
      stateRepo repository.DialogStateRepository
      logger    *zap.Logger
  }

  type DialogState string
  const (
      StateIdle DialogState = "idle"
      StateWaitingForAmount DialogState = "waiting_for_amount"
      StateWaitingForDescription DialogState = "waiting_for_description"
      StateWaitingForCategory DialogState = "waiting_for_category"
      StateWaitingForDate DialogState = "waiting_for_date"
      StateWaitingForEmail DialogState = "waiting_for_email"
      StateWaitingForPassword DialogState = "waiting_for_password"
      StateConfirmingTransaction DialogState = "confirming_transaction"
  )

  func (sm *StateManager) SetState(ctx context.Context, telegramID int64, state DialogState, context map[string]interface{}) error
  func (sm *StateManager) GetState(ctx context.Context, telegramID int64) (*DialogState, map[string]interface{}, error)
  func (sm *StateManager) ClearState(ctx context.Context, telegramID int64) error
  ```

#### День 3-4: Черновики транзакций
- [ ] Создание `internal/domain/transaction_draft.go`:
  ```go
  type TransactionDraft struct {
      ID          string
      TelegramID  int64
      Type        TransactionType
      Amount      *Money
      Description string
      CategoryID  string
      OccurredAt  *time.Time
      CreatedAt   time.Time
  }
  ```
- [ ] Управление черновиками в базе данных
- [ ] Сохранение и восстановление контекста диалога

#### День 5: Интерактивные диалоги
- [ ] Реализация диалогов для запроса недостающих данных
- [ ] Подтверждение транзакций перед сохранением
- [ ] Команда `/cancel` для отмены текущей операции
- [ ] Обработка таймаутов диалогов

---

## 🎨 Этап 6: Пользовательский интерфейс (1 неделя)

### Неделя 9: UI компоненты

#### День 1-2: Типы сообщений
- [ ] Создание `internal/bot/ui/message_formatter.go`:
  ```go
  type MessageFormatter struct {
      logger *zap.Logger
  }

  func (mf *MessageFormatter) FormatTransactionCreated(tx *Transaction, locale string) string
  func (mf *MessageFormatter) FormatCategoriesList(categories []*Category, locale string) string
  func (mf *MessageFormatter) FormatStats(stats *Stats, locale string) string
  func (mf *MessageFormatter) FormatMoney(amount *Money, locale string) string
  ```
- [ ] Форматирование сообщений о транзакциях с эмодзи
- [ ] Создание списков категорий
- [ ] Форматирование статистики
- [ ] Локализованные сообщения (русский/английский)
- [ ] Форматирование сумм в зависимости от локали

#### День 3-4: Inline клавиатуры
- [ ] Создание `internal/bot/ui/keyboards.go`:
  ```go
  func CreateCategoryKeyboard(categories []*Category) tgbotapi.InlineKeyboardMarkup
  func CreateConfirmationKeyboard() tgbotapi.InlineKeyboardMarkup
  func CreateMainMenuKeyboard() tgbotapi.ReplyKeyboardMarkup
  ```
- [ ] Клавиатура выбора категорий
- [ ] Клавиатура подтверждения действий
- [ ] Главное меню с быстрыми командами

#### День 5: Reply клавиатуры
- [ ] Быстрые команды для часто используемых сумм
- [ ] Шаблоны транзакций
- [ ] Навигация по меню

---

## 📈 Этап 7: Статистика и аналитика (1 неделя)

### Неделя 10: Статистика

#### День 1-2: Команды статистики
- [ ] Создание `internal/grpc/report_client.go`
- [ ] Реализация команды `/stats`:
  ```go
  func (h *Handler) handleStats(ctx context.Context, update tgbotapi.Update) error {
      // Получение статистики через gRPC Report Service
      // Форматирование отчета
      // Отправка с графиками (если возможно)
  }
  ```
- [ ] Реализация команды `/stats 2023-12` для конкретного периода
- [ ] Реализация команды `/stats week` для недельной статистики

#### День 3-4: Дополнительные команды
- [ ] Реализация команды `/top_categories` - топ категорий по расходам
- [ ] Реализация команды `/recent` - последние 10 транзакций
- [ ] Реализация команды `/export` - экспорт данных

#### День 5: Метрики и мониторинг
- [ ] Создание `internal/metrics/metrics.go`:
  ```go
  var (
      transactionsTotal = prometheus.NewCounterVec(...)
      responseTime = prometheus.NewHistogramVec(...)
      activeUsers = prometheus.NewGauge(...)
  )
  ```
- [ ] Интеграция метрик в обработчики
- [ ] Настройка Prometheus endpoint
- [ ] Создание базовых Grafana дашбордов

---

## 🧪 Этап 8: Тестирование и оптимизация (1 неделя)

### Неделя 11: Тестирование

#### День 1-2: Unit тесты
- [ ] Тесты для парсера сообщений:
  ```go
  func TestMessageParser_ParseMessage(t *testing.T) {
      tests := []struct {
          name    string
          input   string
          want    *ParsedTransaction
          wantErr bool
      }{
          {"simple expense", "1000 продукты", &ParsedTransaction{...}, false},
          {"income with plus", "+50000 зарплата", &ParsedTransaction{...}, false},
          {"with date", "01.12 5000 подарок", &ParsedTransaction{...}, false},
      }
      // ...
  }
  ```
- [ ] Тесты для Category Matcher
- [ ] Тесты для State Manager
- [ ] Тесты для Auth Manager

#### День 3-4: Integration тесты
- [ ] Тесты интеграции с gRPC сервисами
- [ ] Тесты работы с базой данных
- [ ] Тесты полного flow добавления транзакции

#### День 5: E2E тесты
- [ ] Создание тестового Telegram бота
- [ ] Автоматизированные тесты через Telegram Bot API
- [ ] Тестирование различных сценариев использования

### Неделя 12: Оптимизация и документация

#### День 1-2: Оптимизация производительности
- [ ] Профилирование приложения
- [ ] Оптимизация запросов к базе данных
- [ ] Кеширование часто используемых данных
- [ ] Оптимизация gRPC соединений

#### День 3-4: Документация
- [ ] Создание README.md с инструкциями по установке
- [ ] Документация API
- [ ] Руководство пользователя
- [ ] Документация по развертыванию

#### День 5: Финальная подготовка
- [ ] Создание релизных версий
- [ ] Настройка production конфигурации
- [ ] Подготовка к развертыванию
- [ ] Финальное тестирование

---

## 📋 Чек-лист готовности к релизу

### Функциональность
- [ ] Регистрация и аутентификация пользователей
- [ ] Парсинг сообщений в формате транзакций
- [ ] Автоматическая категоризация
- [ ] Управление сопоставлениями категорий
- [ ] Статистика и отчеты
- [ ] Управление организациями

### Техническое качество
- [ ] Все unit тесты проходят
- [ ] Integration тесты проходят
- [ ] E2E тесты проходят
- [ ] Код покрыт тестами на 80%+
- [ ] Линтеры не выдают ошибок
- [ ] Документация актуальна

### Безопасность
- [ ] JWT токены зашифрованы
- [ ] Валидация всех входных данных
- [ ] Логирование действий пользователей
- [ ] Защита от SQL инъекций
- [ ] Rate limiting настроен

### Производительность
- [ ] Время ответа < 2 секунд
- [ ] Использование памяти < 100MB
- [ ] Поддержка 100+ одновременных пользователей
- [ ] Мониторинг настроен

### Развертывание
- [ ] Docker образ создан
- [ ] Docker Compose настроен
- [ ] Kubernetes манифесты готовы
- [ ] CI/CD pipeline работает
- [ ] Production конфигурация готова

---

## 🚀 Команды для быстрого старта

### Локальная разработка
```bash
# Клонирование и настройка
git clone <repository>
cd telegram-bot
make setup

# Запуск с Docker Compose
make up

# Запуск только бота
make run

# Тесты
make test

# Линтинг
make lint

# Сборка
make build
```

### Production развертывание
```bash
# Сборка образа
docker build -t telegram-bot .

# Запуск с переменными окружения
docker run -d \
  --name telegram-bot \
  -e TELEGRAM_BOT_TOKEN=your_token \
  -e GRPC_SERVER_ADDRESS=your_grpc_server:8080 \
  telegram-bot
```

---

**Статус плана**: 📋 Готов к выполнению

Данный план содержит детальные задачи для каждого этапа разработки с временными рамками и конкретными техническими решениями.
