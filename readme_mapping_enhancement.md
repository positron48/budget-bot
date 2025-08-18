# Улучшения системы маппинга описаний и категорий

## 🎯 Обзор

Данный документ содержит предложения по улучшению системы автоматической категоризации транзакций в Telegram боте. Система должна быть умной, обучаемой и адаптивной к привычкам пользователя.

## 🏗️ Текущая архитектура

### Базовая система сопоставлений

```sql
CREATE TABLE category_mappings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    keyword TEXT NOT NULL,
    category_id UUID NOT NULL,
    priority INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, keyword)
);
```

### Простой алгоритм поиска

1. Точные совпадения ключевых слов
2. Частичные совпадения (подстроки)
3. Учет приоритета сопоставлений
4. Возврат наиболее подходящей категории

## 🚀 Предложения по улучшению

### 1. Расширенная система сопоставлений

#### 1.1 Регулярные выражения

```sql
CREATE TABLE category_mappings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    keyword TEXT NOT NULL,
    pattern TEXT, -- регулярное выражение
    category_id UUID NOT NULL,
    priority INTEGER DEFAULT 0,
    match_type TEXT DEFAULT 'exact', -- exact, partial, regex, fuzzy
    confidence DECIMAL(3,2) DEFAULT 1.0,
    language TEXT DEFAULT 'ru', -- язык ключевого слова
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, keyword, language)
);
```

**Примеры использования:**
```go
// Точное совпадение (русский)
"макдональдс" -> "Питание"

// Точное совпадение (английский)
"mcdonalds" -> "Food"

// Регулярное выражение
"мак.*|mcd.*" -> "Питание"

// Частичное совпадение
"такси" -> "Транспорт" (находит "яндекс такси", "uber такси")

// Нечеткое совпадение
"продукты" -> "Питание" (находит "продукт", "продуктовый")
```

#### 1.2 Контекстные сопоставления

```sql
CREATE TABLE contextual_mappings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    keyword TEXT NOT NULL,
    category_id UUID NOT NULL,
    context_type TEXT, -- time_of_day, day_of_week, location, amount_range
    context_value TEXT, -- "morning", "monday", "moscow", "1000-5000"
    confidence DECIMAL(3,2) DEFAULT 1.0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Примеры:**
- "кофе" + утро (6-12) -> "Питание"
- "кофе" + вечер (18-24) -> "Развлечения"
- "такси" + ночь (22-6) -> "Транспорт" (с повышенным приоритетом)

### 2. Машинное обучение

#### 2.1 Простая модель на основе частот

```go
type FrequencyModel struct {
    tenantID string
    wordFrequencies map[string]map[string]int // word -> category -> frequency
    totalTransactions int
}

func (fm *FrequencyModel) PredictCategory(description string) (string, float64) {
    words := tokenize(description)
    categoryScores := make(map[string]float64)
    
    for _, word := range words {
        if frequencies, exists := fm.wordFrequencies[word]; exists {
            for category, freq := range frequencies {
                categoryScores[category] += float64(freq) / float64(fm.totalTransactions)
            }
        }
    }
    
    // Возвращаем категорию с наивысшим скором
    return findMaxScore(categoryScores)
}
```

#### 2.2 Нейронная сеть (опционально)

```go
type NeuralModel struct {
    model *tensorflow.SavedModel
    tokenizer *Tokenizer
    categories []string
}

func (nm *NeuralModel) PredictCategory(description string) (string, float64) {
    // Токенизация описания
    tokens := nm.tokenizer.Tokenize(description)
    
    // Предсказание через модель
    prediction := nm.model.Predict(tokens)
    
    // Возврат категории с наивысшей вероятностью
    return nm.categories[prediction.ArgMax()], prediction.Max()
}
```

### 3. Анализ истории транзакций

#### 3.1 Анализ паттернов

```sql
CREATE TABLE transaction_patterns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    user_id UUID NOT NULL,
    pattern_type TEXT, -- amount_pattern, time_pattern, location_pattern
    pattern_data JSONB,
    category_id UUID NOT NULL,
    confidence DECIMAL(3,2) DEFAULT 1.0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Типы паттернов:**
1. **Суммовые паттерны**: определенные суммы часто относятся к одной категории
2. **Временные паттерны**: транзакции в определенное время относятся к определенным категориям
3. **Локационные паттерны**: транзакции в определенных местах относятся к определенным категориям
4. **Комбинированные паттерны**: сочетание нескольких факторов

#### 3.2 Анализ последовательностей

```go
type SequenceAnalyzer struct {
    tenantID string
    maxSequenceLength int
}

func (sa *SequenceAnalyzer) AnalyzeSequences(transactions []Transaction) map[string]float64 {
    // Анализ последовательностей транзакций
    // Например: "продукты" часто следует за "зарплата"
    // Или: "такси" часто следует за "ресторан"
    
    sequences := make(map[string]int)
    for i := 0; i < len(transactions)-1; i++ {
        seq := fmt.Sprintf("%s->%s", 
            transactions[i].CategoryID, 
            transactions[i+1].CategoryID)
        sequences[seq]++
    }
    
    return calculateSequenceProbabilities(sequences)
}
```

### 4. Умные подсказки

#### 4.1 Контекстные подсказки

```go
type ContextualSuggestions struct {
    timeOfDay time.Time
    dayOfWeek time.Weekday
    location  *Location
    amount    *Money
    recentTransactions []Transaction
}

func (cs *ContextualSuggestions) GetSuggestions() []CategorySuggestion {
    suggestions := []CategorySuggestion{}
    
    // Подсказки на основе времени
    if cs.timeOfDay.Hour() >= 6 && cs.timeOfDay.Hour() <= 12 {
        suggestions = append(suggestions, CategorySuggestion{
            CategoryID: "breakfast",
            Confidence: 0.8,
            Reason: "Утреннее время - вероятно завтрак",
        })
    }
    
    // Подсказки на основе дня недели
    if cs.dayOfWeek == time.Friday {
        suggestions = append(suggestions, CategorySuggestion{
            CategoryID: "entertainment",
            Confidence: 0.6,
            Reason: "Пятница - вероятно развлечения",
        })
    }
    
    // Подсказки на основе суммы
    if cs.amount != nil && cs.amount.MinorUnits > 500000 { // > 5000 руб
        suggestions = append(suggestions, CategorySuggestion{
            CategoryID: "shopping",
            Confidence: 0.7,
            Reason: "Крупная сумма - вероятно покупки",
        })
    }
    
    return suggestions
}
```

#### 4.2 Персонализированные подсказки

```go
type PersonalizedSuggestions struct {
    userID string
    userPreferences map[string]float64
    userHabits      map[string]Habit
}

type Habit struct {
    CategoryID string
    Frequency  float64 // частота в неделю
    Amount     *Money
    TimeRange  *TimeRange
}

func (ps *PersonalizedSuggestions) GetPersonalizedSuggestions(description string) []CategorySuggestion {
    suggestions := []CategorySuggestion{}
    
    // Анализ привычек пользователя
    for habitID, habit := range ps.userHabits {
        if ps.isHabitApplicable(description, habit) {
            suggestions = append(suggestions, CategorySuggestion{
                CategoryID: habit.CategoryID,
                Confidence: habit.Frequency / 7.0, // нормализованная частота
                Reason: "Основано на ваших привычках",
            })
        }
    }
    
    return suggestions
}
```

### 5. Обучающая система

#### 5.1 Обратная связь от пользователя

```go
type LearningSystem struct {
    feedbackRepo repository.FeedbackRepository
    modelRepo    repository.ModelRepository
}

func (ls *LearningSystem) ProcessFeedback(feedback *CategoryFeedback) error {
    // Обработка обратной связи пользователя
    // Корректировка модели на основе фидбека
    
    if feedback.IsCorrect {
        // Усиление связи между описанием и категорией
        ls.strengthenMapping(feedback.Description, feedback.CategoryID)
    } else {
        // Ослабление связи и обучение на правильной категории
        ls.weakenMapping(feedback.Description, feedback.SuggestedCategoryID)
        ls.strengthenMapping(feedback.Description, feedback.CorrectCategoryID)
    }
    
    // Переобучение модели
    return ls.retrainModel()
}
```

#### 5.2 Автоматическое обучение

```go
func (ls *LearningSystem) AutoLearn() error {
    // Автоматическое обучение на основе истории транзакций
    
    // Получение транзакций за последний месяц
    transactions := ls.getRecentTransactions(time.Now().AddDate(0, -1, 0))
    
    // Анализ успешных предсказаний
    successfulPredictions := ls.analyzeSuccessfulPredictions(transactions)
    
    // Корректировка модели
    for _, prediction := range successfulPredictions {
        ls.strengthenMapping(prediction.Description, prediction.CategoryID)
    }
    
    // Обнаружение новых паттернов
    newPatterns := ls.discoverNewPatterns(transactions)
    for _, pattern := range newPatterns {
        ls.addPattern(pattern)
    }
    
    return nil
}
```

### 6. Мультиязычная категоризация

#### 6.1 Поддержка языков

```go
type MultilingualCategoryMatcher struct {
    matchers map[string]*CategoryMatcher // language -> matcher
    logger   *zap.Logger
}

func (mcm *MultilingualCategoryMatcher) FindCategory(
    ctx context.Context, 
    tenantID, 
    description string, 
    language string,
) (*Category, error) {
    // Поиск в указанном языке
    if matcher, exists := mcm.matchers[language]; exists {
        if category, err := matcher.FindCategory(ctx, tenantID, description); err == nil {
            return category, nil
        }
    }
    
    // Fallback на основной язык (русский)
    if matcher, exists := mcm.matchers["ru"]; exists {
        return matcher.FindCategory(ctx, tenantID, description)
    }
    
    return nil, fmt.Errorf("no category found")
}
```

#### 6.2 Перевод ключевых слов

```go
type KeywordTranslator struct {
    translations map[string]map[string]string // word -> language -> translation
}

func (kt *KeywordTranslator) TranslateKeyword(word, fromLang, toLang string) string {
    if translations, exists := kt.translations[word]; exists {
        if translation, exists := translations[toLang]; exists {
            return translation
        }
    }
    return word // возвращаем исходное слово если перевод не найден
}

// Примеры переводов
var keywordTranslations = map[string]map[string]string{
    "продукты": {
        "en": "groceries",
        "ru": "продукты",
    },
    "такси": {
        "en": "taxi",
        "ru": "такси",
    },
    "ресторан": {
        "en": "restaurant",
        "ru": "ресторан",
    },
}
```

#### 6.3 Автоматическое определение языка

```go
type LanguageDetector struct {
    patterns map[string][]string // language -> patterns
}

func (ld *LanguageDetector) DetectLanguage(text string) string {
    // Простое определение языка по символам
    if strings.ContainsAny(text, "абвгдеёжзийклмнопрстуфхцчшщъыьэюя") {
        return "ru"
    }
    
    // Определение по ключевым словам
    for lang, patterns := range ld.patterns {
        for _, pattern := range patterns {
            if strings.Contains(strings.ToLower(text), pattern) {
                return lang
            }
        }
    }
    
    return "en" // по умолчанию английский
}
```

### 7. Интеграция с внешними сервисами

#### 7.1 Геолокация

```go
type LocationBasedCategorization struct {
    geocodingService GeocodingService
    locationMappings map[string]string // place_id -> category_id
}

func (lbc *LocationBasedCategorization) CategorizeByLocation(lat, lng float64) (string, float64) {
    // Получение информации о месте
    place := lbc.geocodingService.GetPlaceInfo(lat, lng)
    
    // Поиск категории по месту
    if categoryID, exists := lbc.locationMappings[place.ID]; exists {
        return categoryID, 0.9
    }
    
    // Категоризация по типу места
    switch place.Type {
    case "restaurant", "cafe":
        return "food", 0.8
    case "store", "supermarket":
        return "shopping", 0.8
    case "gas_station":
        return "transport", 0.9
    }
    
    return "", 0.0
}
```

#### 7.2 Банковские API

```go
type BankIntegration struct {
    bankAPI BankAPI
    merchantMappings map[string]string // merchant_id -> category_id
}

func (bi *BankIntegration) CategorizeByMerchant(merchantID string) (string, float64) {
    if categoryID, exists := bi.merchantMappings[merchantID]; exists {
        return categoryID, 0.95
    }
    
    // Получение информации о мерчанте
    merchant := bi.bankAPI.GetMerchantInfo(merchantID)
    
    // Категоризация по типу мерчанта
    return bi.categorizeByMerchantType(merchant.Type), 0.7
}
```

### 8. Интерфейс управления

#### 8.1 Команды бота

```bash
# Базовые команды
/map слово = категория
/map слово
/map --all
/unmap слово

# Мультиязычные команды
/map-ru "макдональдс" = Питание
/map-en "mcdonalds" = Food
/map-lang ru "продукты" = Питание
/map-lang en "groceries" = Food

# Расширенные команды
/map-regex "мак.*|mcd.*" = Питание
/map-context "кофе" утро = Питание
/map-context "кофе" вечер = Развлечения

# Управление паттернами
/patterns --show
/patterns --add "сумма > 5000" = Покупки
/patterns --remove "сумма > 5000"

# Обучение
/learn --auto
/learn --feedback "такси" Транспорт
/learn --reset

# Статистика
/map-stats
/map-accuracy
```

#### 8.2 Веб-интерфейс

Создание веб-интерфейса для управления маппингами:

1. **Просмотр всех маппингов** с возможностью редактирования
2. **Добавление новых маппингов** с различными типами (точное, regex, контекстное)
3. **Анализ точности** предсказаний
4. **Управление паттернами** и правилами
5. **Обучение модели** и просмотр результатов

### 9. Метрики и мониторинг

#### 9.1 Метрики точности

```go
type AccuracyMetrics struct {
    TotalPredictions    int64
    CorrectPredictions  int64
    AccuracyByCategory  map[string]float64
    AccuracyByTime      map[string]float64
    UserFeedback        map[string]int64
}

func (am *AccuracyMetrics) CalculateAccuracy() float64 {
    return float64(am.CorrectPredictions) / float64(am.TotalPredictions)
}

func (am *AccuracyMetrics) GetCategoryAccuracy(categoryID string) float64 {
    if total, exists := am.TotalPredictionsByCategory[categoryID]; exists {
        if correct, exists := am.CorrectPredictionsByCategory[categoryID]; exists {
            return float64(correct) / float64(total)
        }
    }
    return 0.0
}
```

#### 9.2 Prometheus метрики

```go
var (
    categoryPredictionTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "category_predictions_total",
            Help: "Total number of category predictions",
        },
        []string{"category", "correct"},
    )
    
    categoryPredictionAccuracy = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "category_prediction_accuracy",
            Help: "Category prediction accuracy",
        },
        []string{"category"},
    )
    
    mappingRulesTotal = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "mapping_rules_total",
            Help: "Total number of mapping rules",
        },
        []string{"type"},
    )
)
```

## 🚀 План внедрения

### Фаза 1: Базовые улучшения (2 недели)
1. Расширение схемы БД для поддержки regex и контекстных маппингов
2. Реализация улучшенного алгоритма поиска
3. Добавление команд для управления расширенными маппингами
4. Поддержка мультиязычности в маппингах

### Фаза 2: Машинное обучение (3 недели)
1. Реализация простой модели на основе частот
2. Система обратной связи от пользователя
3. Автоматическое обучение на основе истории

### Фаза 3: Продвинутые функции (2 недели)
1. Интеграция с геолокацией
2. Анализ паттернов и последовательностей
3. Персонализированные подсказки

### Фаза 4: Оптимизация и мониторинг (1 неделя)
1. Метрики точности и производительности
2. Оптимизация алгоритмов
3. Документация и тестирование

## 📊 Ожидаемые результаты

### Точность категоризации
- **Текущая система**: 60-70%
- **После улучшений**: 85-95%

### Скорость работы
- **Время предсказания**: < 100ms
- **Поддержка**: 1000+ одновременных пользователей

### Пользовательский опыт
- **Снижение ручного ввода**: на 80%
- **Удовлетворенность**: повышение на 40%
- **Время добавления транзакции**: сокращение на 60%

---

**Статус документа**: 📋 Готов к реализации

Данные улучшения значительно повысят точность автоматической категоризации и улучшат пользовательский опыт работы с ботом.
