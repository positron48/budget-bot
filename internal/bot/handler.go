package bot

import (
	"context"
	"fmt"
	"strings"
	"time"

	"budget-bot/internal/bot/ui"
	"budget-bot/internal/metrics"
	grpcclient "budget-bot/internal/grpc"
	"budget-bot/internal/repository"
	"github.com/google/uuid"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type Handler struct {
	bot        *tgbotapi.BotAPI
	states     repository.DialogStateRepository
	auth       *AuthManager
	logger     *zap.Logger
	parser     *MessageParser
	categories grpcclient.CategoryClient
	mappings   repository.CategoryMappingRepository
	matcher    *CategoryMatcher
	txClient   grpcclient.TransactionClient
	prefs      repository.PreferencesRepository
	report     grpcclient.ReportClient
}

func NewHandler(bot *tgbotapi.BotAPI, states repository.DialogStateRepository, auth *AuthManager, mappings repository.CategoryMappingRepository, categories grpcclient.CategoryClient, logger *zap.Logger) *Handler {
	if categories == nil {
		categories = &grpcclient.StaticCategoryClient{}
	}
	return &Handler{bot: bot, states: states, auth: auth, logger: logger, parser: NewMessageParser(), categories: categories, mappings: mappings, matcher: NewCategoryMatcher(mappings), txClient: &grpcclient.FakeTransactionClient{}, report: &grpcclient.FakeReportClient{}}
}

// WithPreferences allows injecting a preferences repository after construction.
func (h *Handler) WithPreferences(p repository.PreferencesRepository) *Handler {
	h.prefs = p
	return h
}

// WithReportClient allows injecting a report client.
func (h *Handler) WithReportClient(rc grpcclient.ReportClient) *Handler {
	if rc != nil {
		h.report = rc
	}
	return h
}

// WithCategoryClient allows injecting a category client.
func (h *Handler) WithCategoryClient(cc grpcclient.CategoryClient) *Handler {
	if cc != nil {
		h.categories = cc
	}
	return h
}

// WithTransactionClient allows injecting a transaction client.
func (h *Handler) WithTransactionClient(tc grpcclient.TransactionClient) *Handler {
	if tc != nil {
		h.txClient = tc
	}
	return h
}

func (h *Handler) HandleUpdate(ctx context.Context, update tgbotapi.Update) {
	if update.CallbackQuery != nil {
		h.handleCallback(ctx, update)
		return
	}
	if update.Message == nil {
		return
	}

	metrics.IncUpdate()

	if update.Message.IsCommand() {
		h.handleCommand(ctx, update)
		return
	}

	rec, _ := h.states.GetState(ctx, update.Message.From.ID)
	if rec != nil {
		switch rec.State {
		case repository.StateWaitingForEmail:
			h.handleLoginEmail(ctx, update)
			return
		case repository.StateWaitingForPassword:
			h.handleLoginPassword(ctx, update)
			return
		case repository.StateWaitingForRegisterEmail:
			h.handleRegisterEmail(ctx, update)
			return
		case repository.StateWaitingForRegisterPassword:
			h.handleRegisterPassword(ctx, update)
			return
		case repository.StateWaitingForRegisterName:
			h.handleRegisterName(ctx, update)
			return
		}
	}

	// Try parse transaction
	parsed, _ := h.parser.ParseMessage(update.Message.Text)
	if parsed != nil && parsed.IsValid {
		// Default currency from preferences if missing
		cur := parsed.Currency
		if cur == "" && h.prefs != nil {
			if pref, err := h.prefs.GetPreferences(ctx, update.Message.From.ID); err == nil && pref != nil && pref.DefaultCurrency != "" {
				cur = pref.DefaultCurrency
			}
			if cur == "" {
				cur = "RUB"
			}
		}
		amt := float64(parsed.Amount.AmountMinor) / 100.0
		// Try suggest category if session present
		if sess, err := h.auth.GetSession(ctx, update.Message.From.ID); err == nil && sess != nil {
			var catID string
			if h.matcher != nil {
				if m, err := h.matcher.FindCategory(ctx, sess.TenantID, parsed.Description); err == nil && m != nil {
					catID = m.CategoryID
				}
			}
			// If no mapping -> ask for category
			if catID == "" {
				list, err := h.categories.ListCategories(ctx, sess.TenantID)
				if err != nil || len(list) == 0 {
					_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Не удалось получить категории"))
					return
				}
				kb := ui.CreateCategoryKeyboard(list)
				_ = h.states.SetState(ctx, update.Message.From.ID, repository.StateWaitingForCategory, map[string]any{
					"type":         string(parsed.Type),
					"amount_minor": parsed.Amount.AmountMinor,
					"currency":     cur,
					"desc":         parsed.Description,
					"occurred_at":  occurredUnix(parsed.OccurredAt),
				}, nil)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите категорию")
				msg.ReplyMarkup = kb
				_, _ = h.bot.Send(msg)
				return
			}
			// Have category -> ask to confirm
			kb := ui.CreateConfirmationKeyboard()
			msg := tgbotapi.NewMessage(update.Message.Chat.ID,
				fmt.Sprintf("Сохранить: %s %.2f %s — %s (категория: %s)?", string(parsed.Type), amt, cur, parsed.Description, catID))
			msg.ReplyMarkup = kb
			_ = h.states.SetState(ctx, update.Message.From.ID, repository.StateConfirmingTransaction, map[string]any{
				"type":         string(parsed.Type),
				"amount_minor": parsed.Amount.AmountMinor,
				"currency":     cur,
				"desc":         parsed.Description,
				"category_id":  catID,
				"occurred_at":  occurredUnix(parsed.OccurredAt),
			}, nil)
			_, _ = h.bot.Send(msg)
			return
		}
		// No session; just echo parse
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			fmt.Sprintf("Распознано: %s %.2f %s — %s", string(parsed.Type), amt, cur, parsed.Description))
		_, _ = h.bot.Send(msg)
		return
	}
}

func occurredUnix(t *time.Time) int64 {
	if t == nil {
		return 0
	}
	return t.Unix()
}

func (h *Handler) handleCallback(ctx context.Context, update tgbotapi.Update) {
	cb := update.CallbackQuery
	if cb == nil {
		return
	}
	data := cb.Data
	if strings.HasPrefix(data, "confirm:") {
		choice := strings.TrimPrefix(data, "confirm:")
		if choice == "yes" {
			// create transaction via txClient using stored state
			rec, _ := h.states.GetState(ctx, cb.From.ID)
			if rec != nil && rec.Context != nil {
				typeStr, _ := rec.Context["type"].(string)
				var amountMinor int64
				switch v := rec.Context["amount_minor"].(type) {
				case float64:
					amountMinor = int64(v)
				case int64:
					amountMinor = v
				case int:
					amountMinor = int64(v)
				}
				currency, _ := rec.Context["currency"].(string)
				desc, _ := rec.Context["desc"].(string)
				catID, _ := rec.Context["category_id"].(string)
				var occurred time.Time
				if ts, ok := rec.Context["occurred_at"].(float64); ok && ts > 0 {
					occurred = time.Unix(int64(ts), 0)
				} else {
					occurred = time.Now()
				}
				sess, err := h.auth.GetSession(ctx, cb.From.ID)
				if err == nil {
					_, _ = h.txClient.CreateTransaction(ctx, &grpcclient.CreateTransactionRequest{
						TenantID:    sess.TenantID,
						Type:        typeStr,
						AmountMinor: amountMinor,
						Currency:    currency,
						Description: desc,
						CategoryID:  catID,
						OccurredAt:  occurred,
					}, sess.AccessToken)
					metrics.IncTransactionsSaved("ok")
				}
			}
			_ = h.states.ClearState(ctx, cb.From.ID)
			_, _ = h.bot.Request(tgbotapi.NewCallback(cb.ID, "Сохранено"))
			_, _ = h.bot.Send(tgbotapi.NewMessage(cb.Message.Chat.ID, "Транзакция сохранена"))
			return
		}
		if choice == "no" {
			_ = h.states.ClearState(ctx, cb.From.ID)
			_, _ = h.bot.Request(tgbotapi.NewCallback(cb.ID, "Отменено"))
			_, _ = h.bot.Send(tgbotapi.NewMessage(cb.Message.Chat.ID, "Отменено"))
			return
		}
	}
	if strings.HasPrefix(data, "cat:") {
		categoryID := strings.TrimPrefix(data, "cat:")
		rec, _ := h.states.GetState(ctx, cb.From.ID)
		if rec == nil || rec.Context == nil {
			_, _ = h.bot.Request(tgbotapi.NewCallback(cb.ID, "Нет контекста"))
			return
		}
		rec.Context["category_id"] = categoryID
		// Ask confirmation now
		var amountMinor int64
		switch v := rec.Context["amount_minor"].(type) {
		case float64:
			amountMinor = int64(v)
		case int64:
			amountMinor = v
		case int:
			amountMinor = int64(v)
		}
		amt := float64(amountMinor) / 100.0
		typeStr, _ := rec.Context["type"].(string)
		currency, _ := rec.Context["currency"].(string)
		desc, _ := rec.Context["desc"].(string)
		kb := ui.CreateConfirmationKeyboard()
		msg := tgbotapi.NewMessage(cb.Message.Chat.ID,
			fmt.Sprintf("Сохранить: %s %.2f %s — %s (категория: %s)?", typeStr, amt, currency, desc, categoryID))
		msg.ReplyMarkup = kb
		_ = h.states.SetState(ctx, cb.From.ID, repository.StateConfirmingTransaction, rec.Context, nil)
		_, _ = h.bot.Request(tgbotapi.NewCallback(cb.ID, "Категория выбрана"))
		_, _ = h.bot.Send(msg)
		return
	}
	if strings.HasPrefix(data, "lang:") {
		lang := strings.TrimPrefix(data, "lang:")
		if h.prefs != nil {
			// preserve currency
			var cur string
			if pref, err := h.prefs.GetPreferences(ctx, cb.From.ID); err == nil && pref != nil {
				cur = pref.DefaultCurrency
			}
			_ = h.prefs.SavePreferences(ctx, &repository.UserPreferences{TelegramID: cb.From.ID, Language: lang, DefaultCurrency: cur})
		}
		_, _ = h.bot.Request(tgbotapi.NewCallback(cb.ID, "Язык: "+lang))
		_, _ = h.bot.Send(tgbotapi.NewMessage(cb.Message.Chat.ID, "Язык обновлён"))
		return
	}
	if strings.HasPrefix(data, "cur:") {
		cur := strings.TrimPrefix(data, "cur:")
		if h.prefs != nil {
			var lang string
			if pref, err := h.prefs.GetPreferences(ctx, cb.From.ID); err == nil && pref != nil {
				lang = pref.Language
			}
			_ = h.prefs.SavePreferences(ctx, &repository.UserPreferences{TelegramID: cb.From.ID, Language: lang, DefaultCurrency: cur})
		}
		_, _ = h.bot.Request(tgbotapi.NewCallback(cb.ID, "Валюта: "+cur))
		_, _ = h.bot.Send(tgbotapi.NewMessage(cb.Message.Chat.ID, "Валюта по умолчанию обновлена"))
		return
	}
}

func (h *Handler) handleCommand(ctx context.Context, update tgbotapi.Update) {
	switch update.Message.Command() {
	case "start":
		h.handleStart(ctx, update)
	case "login":
		h.startLogin(ctx, update)
	case "register":
		h.startRegister(ctx, update)
	case "logout":
		h.handleLogout(ctx, update)
	case "map":
		h.handleMap(ctx, update)
	case "unmap":
		h.handleUnmap(ctx, update)
	case "categories":
		h.handleCategories(ctx, update)
	case "language":
		h.handleLanguage(ctx, update)
	case "currency":
		h.handleCurrency(ctx, update)
	case "stats":
		h.handleStats(ctx, update)
	case "top_categories":
		h.handleTopCategories(ctx, update)
	case "recent":
		h.handleRecent(ctx, update)
	case "cancel":
		h.handleCancel(ctx, update)
	default:
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Unknown command")
		_, _ = h.bot.Send(msg)
	}
}

func (h *Handler) handleCancel(ctx context.Context, update tgbotapi.Update) {
	_ = h.states.ClearState(ctx, update.Message.From.ID)
	_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Текущая операция отменена"))
}

func (h *Handler) handleStart(ctx context.Context, update tgbotapi.Update) {
	// Greet and show basic commands
	var b strings.Builder
	b.WriteString("Привет! Я бот учёта бюджета.\n")
	b.WriteString("/login — вход\n")
	b.WriteString("/register — регистрация\n")
	b.WriteString("/logout — выход\n")
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, b.String())
	_, _ = h.bot.Send(msg)
}

func (h *Handler) startLogin(ctx context.Context, update tgbotapi.Update) {
	_ = h.states.SetState(ctx, update.Message.From.ID, repository.StateWaitingForEmail, nil, nil)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите email:")
	_, _ = h.bot.Send(msg)
}

func (h *Handler) handleLoginEmail(ctx context.Context, update tgbotapi.Update) {
	ctxMap := map[string]any{"email": strings.TrimSpace(update.Message.Text)}
	_ = h.states.SetState(ctx, update.Message.From.ID, repository.StateWaitingForPassword, ctxMap, nil)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите пароль:")
	_, _ = h.bot.Send(msg)
}

func (h *Handler) handleLoginPassword(ctx context.Context, update tgbotapi.Update) {
	rec, _ := h.states.GetState(ctx, update.Message.From.ID)
	if rec == nil || rec.Context == nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Начните с /login")
		_, _ = h.bot.Send(msg)
		return
	}
	email, _ := rec.Context["email"].(string)
	password := strings.TrimSpace(update.Message.Text)
	if err := h.auth.Login(ctx, update.Message.From.ID, email, password); err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка входа. Попробуйте снова /login")
		_, _ = h.bot.Send(msg)
		return
	}
	_ = h.states.ClearState(ctx, update.Message.From.ID)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы успешно вошли")
	_, _ = h.bot.Send(msg)
}

func (h *Handler) handleLogout(ctx context.Context, update tgbotapi.Update) {
	_ = h.auth.Logout(ctx, update.Message.From.ID)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы вышли из системы")
	_, _ = h.bot.Send(msg)
}

func (h *Handler) startRegister(ctx context.Context, update tgbotapi.Update) {
	_ = h.states.SetState(ctx, update.Message.From.ID, repository.StateWaitingForRegisterEmail, nil, nil)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите email для регистрации:")
	_, _ = h.bot.Send(msg)
}

func (h *Handler) handleRegisterEmail(ctx context.Context, update tgbotapi.Update) {
	ctxMap := map[string]any{"email": strings.TrimSpace(update.Message.Text)}
	_ = h.states.SetState(ctx, update.Message.From.ID, repository.StateWaitingForRegisterPassword, ctxMap, nil)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите пароль:")
	_, _ = h.bot.Send(msg)
}

func (h *Handler) handleRegisterPassword(ctx context.Context, update tgbotapi.Update) {
	rec, _ := h.states.GetState(ctx, update.Message.From.ID)
	if rec == nil || rec.Context == nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Начните с /register")
		_, _ = h.bot.Send(msg)
		return
	}
	email, _ := rec.Context["email"].(string)
	ctxMap := map[string]any{"email": email, "password": strings.TrimSpace(update.Message.Text)}
	_ = h.states.SetState(ctx, update.Message.From.ID, repository.StateWaitingForRegisterName, ctxMap, nil)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите имя:")
	_, _ = h.bot.Send(msg)
}

func (h *Handler) handleRegisterName(ctx context.Context, update tgbotapi.Update) {
	rec, _ := h.states.GetState(ctx, update.Message.From.ID)
	if rec == nil || rec.Context == nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Начните с /register")
		_, _ = h.bot.Send(msg)
		return
	}
	email, _ := rec.Context["email"].(string)
	password, _ := rec.Context["password"].(string)
	name := strings.TrimSpace(update.Message.Text)
	if err := h.auth.Register(ctx, update.Message.From.ID, email, password, name); err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка регистрации. Попробуйте снова /register")
		_, _ = h.bot.Send(msg)
		return
	}
	_ = h.states.ClearState(ctx, update.Message.From.ID)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Регистрация успешна. Вы вошли в систему.")
	_, _ = h.bot.Send(msg)
}

func (h *Handler) handleMap(ctx context.Context, update tgbotapi.Update) {
	parts := strings.SplitN(strings.TrimSpace(update.Message.CommandArguments()), "=", 2)
	args := strings.TrimSpace(update.Message.CommandArguments())
	if args == "--all" {
		sess, err := h.auth.GetSession(ctx, update.Message.From.ID)
		if err != nil {
			_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Сначала выполните вход: /login"))
			return
		}
		items, err := h.mappings.ListMappings(ctx, sess.TenantID)
		if err != nil {
			_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Не удалось получить сопоставления"))
			return
		}
		if len(items) == 0 {
			_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Сопоставлений нет"))
			return
		}
		var b strings.Builder
		for _, m := range items {
			b.WriteString(fmt.Sprintf("%s = %s\n", m.Keyword, m.CategoryID))
		}
		_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, b.String()))
		return
	}
	if len(parts) == 1 {
		// show mapping for keyword
		keyword := strings.TrimSpace(parts[0])
		if keyword == "" {
			_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Формат: /map слово = category_id"))
			return
		}
		sess, err := h.auth.GetSession(ctx, update.Message.From.ID)
		if err != nil {
			_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Сначала выполните вход: /login"))
			return
		}
		m, err := h.mappings.FindMapping(ctx, sess.TenantID, keyword)
		if err != nil || m == nil {
			_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Сопоставление не найдено"))
			return
		}
		_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("%s = %s", m.Keyword, m.CategoryID)))
		return
	}
	if len(parts) == 2 {
		keyword := strings.TrimSpace(parts[0])
		categoryID := strings.TrimSpace(parts[1])
		if keyword == "" || categoryID == "" {
			_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Формат: /map слово = category_id"))
			return
		}
		sess, err := h.auth.GetSession(ctx, update.Message.From.ID)
		if err != nil {
			_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Сначала выполните вход: /login"))
			return
		}
		id := uuid.NewString()
		if err := h.mappings.AddMapping(ctx, &repository.CategoryMapping{ID: id, TenantID: sess.TenantID, Keyword: keyword, CategoryID: categoryID, Priority: 0}); err != nil {
			_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Не удалось сохранить сопоставление"))
			return
		}
		_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Сопоставление сохранено"))
		return
	}
	_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Формат: /map слово = category_id"))
}

func (h *Handler) handleUnmap(ctx context.Context, update tgbotapi.Update) {
	keyword := strings.TrimSpace(update.Message.CommandArguments())
	if keyword == "" {
		_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Формат: /unmap слово"))
		return
	}
	sess, err := h.auth.GetSession(ctx, update.Message.From.ID)
	if err != nil {
		_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Сначала выполните вход: /login"))
		return
	}
	if err := h.mappings.RemoveMapping(ctx, sess.TenantID, keyword); err != nil {
		_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Не удалось удалить сопоставление"))
		return
	}
	_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Сопоставление удалено"))
}

func (h *Handler) handleCategories(ctx context.Context, update tgbotapi.Update) {
	sess, err := h.auth.GetSession(ctx, update.Message.From.ID)
	if err != nil {
		_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Сначала выполните вход: /login"))
		return
	}
	list, err := h.categories.ListCategories(ctx, sess.TenantID)
	if err != nil {
		_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Не удалось получить категории"))
		return
	}
	kb := ui.CreateCategoryKeyboard(list)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите категорию")
	msg.ReplyMarkup = kb
	_, _ = h.bot.Send(msg)
}

func (h *Handler) handleLanguage(ctx context.Context, update tgbotapi.Update) {
	kb := ui.CreateLanguageKeyboard()
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите язык интерфейса")
	msg.ReplyMarkup = kb
	_, _ = h.bot.Send(msg)
}

func (h *Handler) handleCurrency(ctx context.Context, update tgbotapi.Update) {
	kb := ui.CreateCurrencyKeyboard()
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите валюту по умолчанию")
	msg.ReplyMarkup = kb
	_, _ = h.bot.Send(msg)
}

func (h *Handler) handleStats(ctx context.Context, update tgbotapi.Update) {
	sess, err := h.auth.GetSession(ctx, update.Message.From.ID)
	if err != nil {
		_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Сначала выполните вход: /login"))
		return
	}
	// Current month
	now := time.Now()
	from := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	to := from.AddDate(0, 1, -1)
	st, err := h.report.GetStats(ctx, sess.TenantID, from, to)
	if err != nil {
		_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Не удалось получить статистику"))
		return
	}
	msg := fmt.Sprintf("Статистика %s\nДоход: %.2f %s\nРасход: %.2f %s", st.Period, float64(st.TotalIncome)/100.0, st.Currency, float64(st.TotalExpense)/100.0, st.Currency)
	_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
}

func (h *Handler) handleTopCategories(ctx context.Context, update tgbotapi.Update) {
	sess, err := h.auth.GetSession(ctx, update.Message.From.ID)
	if err != nil {
		_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Сначала выполните вход: /login"))
		return
	}
	now := time.Now()
	from := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	to := from.AddDate(0, 1, -1)
	items, err := h.report.TopCategories(ctx, sess.TenantID, from, to, 5)
	if err != nil {
		_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Не удалось получить топ категорий"))
		return
	}
	if len(items) == 0 {
		_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Нет данных"))
		return
	}
	var b strings.Builder
	b.WriteString("Топ категорий:\n")
	for i, it := range items {
		b.WriteString(fmt.Sprintf("%d) %s — %.2f %s\n", i+1, it.Name, float64(it.SumMinor)/100.0, it.Currency))
	}
	_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, b.String()))
}

func (h *Handler) handleRecent(ctx context.Context, update tgbotapi.Update) {
	sess, err := h.auth.GetSession(ctx, update.Message.From.ID)
	if err != nil {
		_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Сначала выполните вход: /login"))
		return
	}
	items, err := h.report.Recent(ctx, sess.TenantID, 10)
	if err != nil {
		_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Не удалось получить последние транзакции"))
		return
	}
	if len(items) == 0 {
		_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Нет данных"))
		return
	}
	var b strings.Builder
	b.WriteString("Последние транзакции:\n")
	for _, it := range items {
		b.WriteString("- " + it + "\n")
	}
	_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, b.String()))
}


