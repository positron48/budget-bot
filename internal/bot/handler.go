// Package bot contains the core Telegram bot business logic.
package bot

import (
	"context"
	"fmt"
	"strings"
	"time"

	"budget-bot/internal/bot/ui"
	"budget-bot/internal/domain"
	"budget-bot/internal/metrics"
	grpcclient "budget-bot/internal/grpc"
	"budget-bot/internal/repository"
	"github.com/google/uuid"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
	pb "budget-bot/internal/pb/budget/v1"
)

// Handler wires bot dependencies and handles Telegram updates.
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
	drafts     repository.DraftRepository
	tenants    grpcclient.TenantClient
	fmt        *ui.MessageFormatter
}

// NewHandler constructs a Handler.
func NewHandler(bot *tgbotapi.BotAPI, states repository.DialogStateRepository, auth *AuthManager, mappings repository.CategoryMappingRepository, categories grpcclient.CategoryClient, logger *zap.Logger) *Handler {
	if categories == nil {
		categories = &grpcclient.StaticCategoryClient{}
	}
	return &Handler{bot: bot, states: states, auth: auth, logger: logger, parser: NewMessageParser(), categories: categories, mappings: mappings, matcher: NewCategoryMatcher(mappings), txClient: &grpcclient.FakeTransactionClient{}, report: &grpcclient.FakeReportClient{}, tenants: &grpcclient.FakeTenantClient{}, fmt: ui.NewMessageFormatter()}
}

// WithPreferences allows injecting a preferences repository after construction.
func (h *Handler) WithPreferences(p repository.PreferencesRepository) *Handler {
	h.prefs = p
	return h
}

// WithDrafts allows injecting a draft repository.
func (h *Handler) WithDrafts(d repository.DraftRepository) *Handler {
	h.drafts = d
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

// WithTenantClient allows injecting a tenant client.
func (h *Handler) WithTenantClient(tc grpcclient.TenantClient) *Handler {
	if tc != nil {
		h.tenants = tc
	}
	return h
}

// HandleUpdate processes a single Telegram update.
func (h *Handler) HandleUpdate(ctx context.Context, update tgbotapi.Update) {
	if update.CallbackQuery != nil {
		h.handleCallback(ctx, update)
		return
	}
	if update.Message == nil {
		return
	}

	metrics.IncUpdate()

	// Debug logging for command detection
	if strings.HasPrefix(update.Message.Text, "/") {
		h.logger.Debug("potential command detected", 
			zap.String("text", update.Message.Text),
			zap.Bool("is_command", update.Message.IsCommand()),
			zap.Any("entities", update.Message.Entities))
	}

	if update.Message.IsCommand() {
		h.handleCommand(ctx, update)
		return
	}

	// Fallback for commands that start with / but are not recognized as commands
	if strings.HasPrefix(update.Message.Text, "/") {
		h.logger.Warn("command not recognized by IsCommand()", 
			zap.String("text", update.Message.Text))
		// Try to handle as command anyway
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
			// If no mapping -> ask for category (persist as draft)
			if catID == "" {
				pref, _ := h.prefs.GetPreferences(ctx, update.Message.From.ID)
				locale := ""
				if pref != nil && pref.Language != "" { locale = pref.Language }
				list, err := h.categories.ListCategories(ctx, sess.TenantID, sess.AccessToken, parsed.Type, locale)
				if err != nil || len(list) == 0 {
					_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Не удалось получить категории"))
					return
				}
				kb := ui.CreateCategoryKeyboard(list)
				if h.drafts != nil {
					draftID := uuid.NewString()
					_ = h.drafts.Create(ctx, &repository.TransactionDraft{ID: draftID, TelegramID: update.Message.From.ID, Type: string(parsed.Type), AmountMinor: parsed.Amount.AmountMinor, Currency: cur, Description: parsed.Description, OccurredAt: parsed.OccurredAt})
				}
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
		msgText := fmt.Sprintf("Распознано: %s %.2f %s — %s", string(parsed.Type), amt, cur, parsed.Description)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
		_, err := h.bot.Send(msg)
		if err != nil {
			h.logger.Error("failed to send parse result", zap.Error(err), zap.String("text", msgText))
		}
		return
	}

	if parsed != nil && !parsed.IsValid {
		// Provide simple validation feedback
		msgText := "Не удалось распознать сообщение. Убедитесь, что указана сумма (например: 100 кофе)"
		if len(parsed.Errors) > 0 {
			// Show first error in a user-friendly way
			msgText = "Ошибка: " + parsed.Errors[0]
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
		_, err := h.bot.Send(msg)
		if err != nil {
			h.logger.Error("failed to send validation error", zap.Error(err), zap.String("text", msgText))
		}
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
					// Refresh tokens if expiring soon
					if time.Until(sess.AccessTokenExpiresAt) < 30*time.Second {
						_ = h.auth.RefreshTokens(ctx, cb.From.ID)
						// reload session
						sess, _ = h.auth.GetSession(ctx, cb.From.ID)
					}
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
				// Cleanup draft if present
				if h.drafts != nil {
					if dID, ok := rec.Context["draft_id"].(string); ok && dID != "" {
						_ = h.drafts.Delete(ctx, dID)
					}
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
	if strings.HasPrefix(data, "tenant:") {
		tenantID := strings.TrimPrefix(data, "tenant:")
		_, _ = h.bot.Request(tgbotapi.NewCallback(cb.ID, "Организация выбрана"))
		if err := h.auth.sessionRepo.UpdateTenantID(ctx, cb.From.ID, tenantID); err == nil {
			_, _ = h.bot.Send(tgbotapi.NewMessage(cb.Message.Chat.ID, "Организация переключена"))
			return
		}
		_, _ = h.bot.Send(tgbotapi.NewMessage(cb.Message.Chat.ID, "Не удалось переключить организацию"))
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
	case "export":
		h.handleExport(ctx, update)
	case "create_category":
		h.handleCreateCategory(ctx, update)
	case "rename_category":
		h.handleRenameCategory(ctx, update)
	case "delete_category":
		h.handleDeleteCategory(ctx, update)
	case "switch_tenant":
		h.handleSwitchTenant(ctx, update)
	case "profile":
		h.handleProfile(ctx, update)
	case "cancel":
		h.handleCancel(ctx, update)
	default:
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда")
		_, err := h.bot.Send(msg)
		if err != nil {
			h.logger.Error("failed to send unknown command message", zap.Error(err))
		}
	}
}

func (h *Handler) handleSwitchTenant(ctx context.Context, update tgbotapi.Update) {
	sess, err := h.auth.GetSession(ctx, update.Message.From.ID)
	if err != nil {
		_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Сначала выполните вход: /login"))
		return
	}
	list, err := h.tenants.ListTenants(ctx, sess.AccessToken)
	if err != nil || len(list) == 0 {
		_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Не удалось получить организации"))
		return
	}
	kb := ui.CreateTenantKeyboard(list)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите организацию")
	msg.ReplyMarkup = kb
	_, _ = h.bot.Send(msg)
}

func (h *Handler) handleCancel(ctx context.Context, update tgbotapi.Update) {
	_ = h.states.ClearState(ctx, update.Message.From.ID)
	_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Текущая операция отменена"))
}

func (h *Handler) handleStart(_ context.Context, update tgbotapi.Update) {
	// Greet and show basic commands
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Привет! Я бот учёта бюджета.\n\n"+
		"/login — вход\n"+
		"/register — регистрация\n"+
		"/logout — выход\n\n"+
		"Отправьте сумму и описание для добавления транзакции, например:\n"+
		"1000 продукты\n"+
		"+50000 зарплата")
	
	menu := ui.CreateMainMenuKeyboard()
	msg.ReplyMarkup = menu
	
	_, err := h.bot.Send(msg)
	if err != nil {
		h.logger.Error("failed to send start message", zap.Error(err))
	}
}

func (h *Handler) startLogin(ctx context.Context, update tgbotapi.Update) {
	_ = h.states.SetState(ctx, update.Message.From.ID, repository.StateWaitingForEmail, nil, nil)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите email:")
	_, err := h.bot.Send(msg)
	if err != nil {
		h.logger.Error("failed to send login email prompt", zap.Error(err))
	}
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
	pref, _ := h.prefs.GetPreferences(ctx, update.Message.From.ID)
	locale := ""
	if pref != nil && pref.Language != "" { locale = pref.Language }
	// Default to expense categories for /categories command
	list, err := h.categories.ListCategories(ctx, sess.TenantID, sess.AccessToken, domain.TransactionExpense, locale)
	if err != nil {
		_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Не удалось получить категории"))
		return
	}
	kb := ui.CreateCategoryKeyboard(list)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите категорию")
	msg.ReplyMarkup = kb
	_, _ = h.bot.Send(msg)
}

func (h *Handler) handleLanguage(_ context.Context, update tgbotapi.Update) {
	kb := ui.CreateLanguageKeyboard()
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите язык интерфейса")
	msg.ReplyMarkup = kb
	_, _ = h.bot.Send(msg)
}

func (h *Handler) handleCurrency(_ context.Context, update tgbotapi.Update) {
	kb := ui.CreateCurrencyKeyboard()
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите валюту по умолчанию")
	msg.ReplyMarkup = kb
	_, _ = h.bot.Send(msg)
}

func (h *Handler) handleStats(ctx context.Context, update tgbotapi.Update) {
	h.logger.Debug("handleStats called", 
		zap.Int64("userID", update.Message.From.ID),
		zap.String("commandArgs", update.Message.CommandArguments()))
	
	sess, err := h.auth.GetSession(ctx, update.Message.From.ID)
	if err != nil {
		h.logger.Warn("handleStats: no session found", 
			zap.Int64("userID", update.Message.From.ID),
			zap.Error(err))
		_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Сначала выполните вход: /login"))
		return
	}
	
	h.logger.Debug("handleStats: session found", 
		zap.String("tenantID", sess.TenantID),
		zap.String("accessToken", sess.AccessToken[:10] + "..."))
	
	// Current month (overridden by optional arg)
	now := time.Now()
	from := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	to := from.AddDate(0, 1, -1)
	if arg := strings.TrimSpace(update.Message.CommandArguments()); arg != "" {
		if arg == "week" {
			wd := int(now.Weekday())
			if wd == 0 { wd = 7 }
			from = time.Date(now.Year(), now.Month(), now.Day()-(wd-1), 0, 0, 0, 0, now.Location())
			to = from.AddDate(0, 0, 6)
		} else if len(arg) == 7 {
			var y, m int
			if _, e := fmt.Sscanf(arg, "%d-%d", &y, &m); e == nil && m >= 1 && m <= 12 {
				from = time.Date(y, time.Month(m), 1, 0, 0, 0, 0, now.Location())
				to = from.AddDate(0, 1, -1)
			}
		}
	}
	
	h.logger.Debug("handleStats: calling GetStats", 
		zap.Time("from", from),
		zap.Time("to", to))
	
	st, err := h.report.GetStats(ctx, sess.TenantID, from, to, sess.AccessToken)
	if err != nil {
		h.logger.Error("handleStats: GetStats failed", 
			zap.String("tenantID", sess.TenantID),
			zap.Time("from", from),
			zap.Time("to", to),
			zap.Error(err))
		_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Не удалось получить статистику"))
		return
	}
	
	h.logger.Debug("handleStats: GetStats successful", 
		zap.String("period", st.Period),
		zap.Int64("totalIncome", st.TotalIncome),
		zap.Int64("totalExpense", st.TotalExpense),
		zap.String("currency", st.Currency))
	
	text := h.fmt.FormatStats(st)
	_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, text))
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
	limit := 5
	if arg := strings.TrimSpace(update.Message.CommandArguments()); arg != "" {
		parts := strings.Fields(arg)
		for _, p := range parts {
			if p == "week" {
				wd := int(now.Weekday())
				if wd == 0 { wd = 7 }
				from = time.Date(now.Year(), now.Month(), now.Day()-(wd-1), 0, 0, 0, 0, now.Location())
				to = from.AddDate(0, 0, 6)
				continue
			}
			if len(p) == 7 {
				var y, m int
				if _, e := fmt.Sscanf(p, "%d-%d", &y, &m); e == nil && m >= 1 && m <= 12 {
					from = time.Date(y, time.Month(m), 1, 0, 0, 0, 0, now.Location())
					to = from.AddDate(0, 1, -1)
					continue
				}
			}
			if v, e := fmt.Sscanf(p, "%d", &limit); e == nil && v >= 0 {
				// limit parsed via Sscanf above; ensure sensible bounds
				if limit <= 0 { limit = 5 }
				if limit > 50 { limit = 50 }
			}
		}
	}
	items, err := h.report.TopCategories(ctx, sess.TenantID, from, to, limit, sess.AccessToken)
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
	limit := 10
	if arg := strings.TrimSpace(update.Message.CommandArguments()); arg != "" {
		var parsed int
		if _, e := fmt.Sscanf(arg, "%d", &parsed); e == nil {
			if parsed > 0 && parsed <= 100 { limit = parsed }
		}
	}
	txs, err := h.txClient.ListRecent(ctx, sess.TenantID, limit, sess.AccessToken)
	if err != nil {
		_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Не удалось получить последние транзакции"))
		return
	}
	if len(txs) == 0 {
		_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Нет данных"))
		return
	}
	var b strings.Builder
	b.WriteString("Последние транзакции:\n")
	for _, t := range txs {
		sign := "-"
		if t.GetType() == pb.TransactionType_TRANSACTION_TYPE_INCOME {
			sign = "+"
		}
		amt := float64(t.GetAmount().GetMinorUnits()) / 100.0
		curr := t.GetAmount().GetCurrencyCode()
		b.WriteString(fmt.Sprintf("- %s%.2f %s %s\n", sign, amt, curr, t.GetComment()))
	}
	_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, b.String()))
}

func (h *Handler) handleExport(ctx context.Context, update tgbotapi.Update) {
    sess, err := h.auth.GetSession(ctx, update.Message.From.ID)
    if err != nil {
        _, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Сначала выполните вход: /login"))
        return
    }
    // Export current month by default; supports args: YYYY-MM|week [limit]
    now := time.Now()
    from := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
    to := from.AddDate(0, 1, -1)
    limit := 100
    if arg := strings.TrimSpace(update.Message.CommandArguments()); arg != "" {
        parts := strings.Fields(arg)
        for _, p := range parts {
            if p == "week" {
                wd := int(now.Weekday())
                if wd == 0 { wd = 7 }
                from = time.Date(now.Year(), now.Month(), now.Day()-(wd-1), 0, 0, 0, 0, now.Location())
                to = from.AddDate(0, 0, 6)
                continue
            }
            if len(p) == 7 {
                var y, m int
                if _, e := fmt.Sscanf(p, "%d-%d", &y, &m); e == nil && m >= 1 && m <= 12 {
                    from = time.Date(y, time.Month(m), 1, 0, 0, 0, 0, now.Location())
                    to = from.AddDate(0, 1, -1)
                    continue
                }
            }
            var v int
            if _, e := fmt.Sscanf(p, "%d", &v); e == nil {
                if v > 0 && v <= 5000 { limit = v }
            }
        }
    }
    txs, err := h.txClient.ListForExport(ctx, sess.TenantID, from, to, limit, sess.AccessToken)
    if err != nil {
        _, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Не удалось выгрузить транзакции"))
        return
    }
    var b strings.Builder
    b.WriteString("date,type,amount,currency,category_id,comment\n")
    for _, t := range txs {
        dt := t.GetOccurredAt().AsTime().Format("2006-01-02")
        typ := "expense"
        if t.GetType() == pb.TransactionType_TRANSACTION_TYPE_INCOME { typ = "income" }
        amt := float64(t.GetAmount().GetMinorUnits())/100.0
        curr := t.GetAmount().GetCurrencyCode()
        b.WriteString(fmt.Sprintf("%s,%s,%.2f,%s,%s,%s\n", dt, typ, amt, curr, t.GetCategoryId(), strings.ReplaceAll(t.GetComment(), ",", " ")))
    }
    file := tgbotapi.FileBytes{Name: "export.csv", Bytes: []byte(b.String())}
    msg := tgbotapi.NewDocument(update.Message.Chat.ID, file)
    msg.Caption = "Экспорт за текущий месяц"
    _, _ = h.bot.Send(msg)
}

func (h *Handler) handleProfile(ctx context.Context, update tgbotapi.Update) {
	sess, _ := h.auth.GetSession(ctx, update.Message.From.ID)
	pref, _ := h.prefs.GetPreferences(ctx, update.Message.From.ID)
	var b strings.Builder
	b.WriteString("Профиль:\n")
	if sess != nil {
		b.WriteString(fmt.Sprintf("UserID: %s\nTenantID: %s\n", sess.UserID, sess.TenantID))
	} else {
		b.WriteString("Не авторизован\n")
	}
	if pref != nil {
		b.WriteString(fmt.Sprintf("Язык: %s\nВалюта по умолчанию: %s\n", pref.Language, pref.DefaultCurrency))
	}
	_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, b.String()))
}


func (h *Handler) handleCreateCategory(ctx context.Context, update tgbotapi.Update) {
    sess, err := h.auth.GetSession(ctx, update.Message.From.ID)
    if err != nil {
        _, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Сначала выполните вход: /login"))
        return
    }
    args := strings.TrimSpace(update.Message.CommandArguments())
    if args == "" {
        _, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Формат: /create_category code название"))
        return
    }
    parts := strings.Fields(args)
    if len(parts) < 2 {
        _, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Формат: /create_category code название"))
        return
    }
    code := parts[0]
    name := strings.TrimSpace(strings.TrimPrefix(args, code))
    if name == "" {
        _, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Укажите название категории"))
        return
    }
    pref, _ := h.prefs.GetPreferences(ctx, update.Message.From.ID)
    locale := ""
    if pref != nil && pref.Language != "" { locale = pref.Language }
    cat, err := h.categories.CreateCategory(ctx, sess.AccessToken, code, name, locale)
    if err != nil {
        _, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Не удалось создать категорию (доступно в сборке withgrpc)"))
        return
    }
    _, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Категория создана: %s (%s)", cat.Name, cat.ID)))
}

func (h *Handler) handleRenameCategory(ctx context.Context, update tgbotapi.Update) {
    sess, err := h.auth.GetSession(ctx, update.Message.From.ID)
    if err != nil {
        _, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Сначала выполните вход: /login"))
        return
    }
    args := strings.TrimSpace(update.Message.CommandArguments())
    parts := strings.Fields(args)
    if len(parts) < 2 {
        _, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Формат: /rename_category category_id новое_название"))
        return
    }
    id := parts[0]
    name := strings.TrimSpace(strings.TrimPrefix(args, id))
    if name == "" {
        _, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Укажите новое название"))
        return
    }
    pref, _ := h.prefs.GetPreferences(ctx, update.Message.From.ID)
    locale := ""
    if pref != nil && pref.Language != "" { locale = pref.Language }
    cat, err := h.categories.UpdateCategoryName(ctx, sess.AccessToken, id, name, locale)
    if err != nil {
        _, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Не удалось обновить категорию (доступно в сборке withgrpc)"))
        return
    }
    _, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Категория обновлена: %s (%s)", cat.Name, cat.ID)))
}

func (h *Handler) handleDeleteCategory(ctx context.Context, update tgbotapi.Update) {
    sess, err := h.auth.GetSession(ctx, update.Message.From.ID)
    if err != nil {
        _, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Сначала выполните вход: /login"))
        return
    }
    id := strings.TrimSpace(update.Message.CommandArguments())
    if id == "" {
        _, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Формат: /delete_category category_id"))
        return
    }
    if err := h.categories.DeleteCategory(ctx, sess.AccessToken, id); err != nil {
        _, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Не удалось удалить категорию (доступно в сборке withgrpc)"))
        return
    }
    _, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Категория удалена"))
}

