// Package bot contains the core Telegram bot business logic.
package bot

import (
	"context"
	"fmt"
	"math"
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
	auth       *OAuthManager
	logger     *zap.Logger
	parser     *MessageParser
	categories grpcclient.CategoryClient
	mappings   repository.CategoryMappingRepository
	matcher    *CategoryMatcher
	nameMapper *CategoryNameMapper
	txClient   grpcclient.TransactionClient
	prefs      repository.PreferencesRepository
	report     grpcclient.ReportClient
	drafts     repository.DraftRepository
	tenants    grpcclient.TenantClient
	fmt        *ui.MessageFormatter
}

// NewHandler constructs a Handler.
func NewHandler(bot *tgbotapi.BotAPI, states repository.DialogStateRepository, auth *OAuthManager, mappings repository.CategoryMappingRepository, categories grpcclient.CategoryClient, logger *zap.Logger) *Handler {
	if categories == nil {
		categories = &grpcclient.StaticCategoryClient{}
	}
	return &Handler{bot: bot, states: states, auth: auth, logger: logger, parser: NewMessageParser(), categories: categories, mappings: mappings, matcher: NewCategoryMatcher(mappings), nameMapper: NewCategoryNameMapper(categories), txClient: &grpcclient.FakeTransactionClient{}, report: &grpcclient.FakeReportClient{}, tenants: &grpcclient.FakeTenantClient{}, fmt: ui.NewMessageFormatter()}
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
		// Try to handle as command anyway - create a modified update with proper entities
		modifiedUpdate := update
		parts := strings.Fields(update.Message.Text)
		if len(parts) > 0 {
			// Create proper entities for the command
			commandLength := len(parts[0])
			modifiedUpdate.Message.Entities = []tgbotapi.MessageEntity{
				{
					Type:   "bot_command",
					Offset: 0,
					Length: commandLength,
				},
			}
		}
		h.handleCommand(ctx, modifiedUpdate)
		return
	}

	rec, _ := h.states.GetState(ctx, update.Message.From.ID)
	if rec != nil {
		switch rec.State {
		case repository.StateWaitingForOAuthEmail:
			h.handleOAuthEmail(ctx, update)
			return
		case repository.StateWaitingForOAuthCode:
			h.handleOAuthCode(ctx, update)
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
		sess, err := h.auth.GetSession(ctx, update.Message.From.ID)
		if err == nil && sess != nil {
			// Проверяем, что сессия действительно валидна (токены не истекли)
			if time.Now().After(sess.AccessTokenExpiresAt) {
				h.logger.Warn("Session has expired tokens, user needs to re-authenticate", 
					zap.Int64("telegramID", update.Message.From.ID),
					zap.Time("accessTokenExpiresAt", sess.AccessTokenExpiresAt),
					zap.Time("refreshTokenExpiresAt", sess.RefreshTokenExpiresAt))
				// Удаляем невалидную сессию
				if err := h.auth.Logout(ctx, update.Message.From.ID); err != nil {
					h.logger.Error("Failed to logout user with expired tokens", 
						zap.Int64("telegramID", update.Message.From.ID),
						zap.Error(err))
				}
				// Продолжаем как будто сессии нет
				sess = nil
			}
		}
		
		if sess != nil {
			h.logger.Debug("Got valid session for user", 
				zap.Int64("telegramID", update.Message.From.ID),
				zap.String("accessToken", sess.AccessToken[:int(math.Min(float64(len(sess.AccessToken)), 10))] + "..."),
				zap.String("refreshToken", sess.RefreshToken[:int(math.Min(float64(len(sess.RefreshToken)), 10))] + "..."),
				zap.Time("accessTokenExpiresAt", sess.AccessTokenExpiresAt),
				zap.Time("refreshTokenExpiresAt", sess.RefreshTokenExpiresAt),
				zap.Time("now", time.Now()),
				zap.Bool("accessTokenExpired", time.Now().After(sess.AccessTokenExpiresAt)),
				zap.Bool("refreshTokenExpired", time.Now().After(sess.RefreshTokenExpiresAt)))
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
				h.logger.Debug("Calling ListCategories with access token", 
					zap.Int64("telegramID", update.Message.From.ID),
					zap.String("accessToken", sess.AccessToken[:int(math.Min(float64(len(sess.AccessToken)), 10))] + "..."),
					zap.String("transactionType", string(parsed.Type)),
					zap.String("locale", locale))
				list, err := h.categories.ListCategories(ctx, sess.TenantID, sess.AccessToken, parsed.Type, locale)
				if err != nil || len(list) == 0 {
					h.logger.Error("Failed to get categories", 
						zap.Int64("telegramID", update.Message.From.ID),
						zap.Error(err))
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
			// Have category -> create transaction immediately
			// Get category name by ID for display
			var categoryDisplayName string
			if h.nameMapper != nil {
				pref, _ := h.prefs.GetPreferences(ctx, update.Message.From.ID)
				locale := "ru"
				if pref != nil && pref.Language != "" {
					locale = pref.Language
				}
				if name, err := h.nameMapper.GetCategoryNameByID(ctx, sess.TenantID, sess.AccessToken, catID, parsed.Type, locale); err == nil && name != "" {
					categoryDisplayName = name
				} else {
					categoryDisplayName = catID // fallback to ID if name not found
				}
			} else {
				categoryDisplayName = catID // fallback to ID if nameMapper not available
			}
			
			// Create transaction immediately
			_, err = h.txClient.CreateTransaction(ctx, &grpcclient.CreateTransactionRequest{
				TenantID:    sess.TenantID,
				Type:        string(parsed.Type),
				AmountMinor: parsed.Amount.AmountMinor,
				Currency:    cur,
				Description: parsed.Description,
				CategoryID:  catID,
				OccurredAt:  func() time.Time {
					if parsed.OccurredAt != nil {
						return *parsed.OccurredAt
					}
					return time.Now()
				}(),
			}, sess.AccessToken)
			
			if err != nil {
				h.logger.Error("Failed to create transaction", 
					zap.Int64("telegramID", update.Message.From.ID),
					zap.Error(err))
				_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Не удалось сохранить транзакцию"))
				return
			}
			
			// Send success message
			msg := tgbotapi.NewMessage(update.Message.Chat.ID,
				fmt.Sprintf("✅ Сохранено: %s %.2f %s — %s (категория: %s)", string(parsed.Type), amt, cur, parsed.Description, categoryDisplayName))
			_, _ = h.bot.Send(msg)
			metrics.IncTransactionsSaved("ok")
			return
		}
		// No session; just echo parse
		msgText := fmt.Sprintf("Распознано: %s %.2f %s — %s", string(parsed.Type), amt, cur, parsed.Description)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
		_, sendErr := h.bot.Send(msg)
		if sendErr != nil {
			h.logger.Error("failed to send parse result", zap.Error(sendErr), zap.String("text", msgText))
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
		_, sendErr := h.bot.Send(msg)
		if sendErr != nil {
			h.logger.Error("failed to send validation error", zap.Error(sendErr), zap.String("text", msgText))
		}
		return
	}
}

func occurredUnix(t *time.Time) int64 {
	if t == nil {
		return 0
	}
	// Time is already in UTC, just convert to Unix timestamp
	return t.Unix()
}

func (h *Handler) handleCallback(ctx context.Context, update tgbotapi.Update) {
	cb := update.CallbackQuery
	if cb == nil {
		return
	}
	data := cb.Data

	if strings.HasPrefix(data, "cat:") {
		categoryName := strings.TrimPrefix(data, "cat:")
		rec, _ := h.states.GetState(ctx, cb.From.ID)
		if rec == nil || rec.Context == nil {
			_, _ = h.bot.Request(tgbotapi.NewCallback(cb.ID, "Нет контекста"))
			return
		}
		
		// Get session for tenant and access token
		sess, err := h.auth.GetSession(ctx, cb.From.ID)
		if err != nil || sess == nil {
			_, _ = h.bot.Request(tgbotapi.NewCallback(cb.ID, "Нет сессии"))
			return
		}
		
		// Get preferences for locale
		pref, _ := h.prefs.GetPreferences(ctx, cb.From.ID)
		locale := "ru"
		if pref != nil && pref.Language != "" {
			locale = pref.Language
		}
		
		// Determine transaction type from context
		typeStr, _ := rec.Context["type"].(string)
		var transactionType domain.TransactionType
		if typeStr == "income" {
			transactionType = domain.TransactionIncome
		} else {
			transactionType = domain.TransactionExpense
		}
		
		// Map category name to ID
		categoryID, err := h.nameMapper.GetCategoryIDByName(ctx, sess.TenantID, sess.AccessToken, categoryName, transactionType, locale)
		if err != nil || categoryID == "" {
			_, _ = h.bot.Request(tgbotapi.NewCallback(cb.ID, "Категория не найдена"))
			return
		}
		
		rec.Context["category_id"] = categoryID
		// Create transaction immediately
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
		
		// Create transaction immediately
		_, err = h.txClient.CreateTransaction(ctx, &grpcclient.CreateTransactionRequest{
			TenantID:    sess.TenantID,
			Type:        typeStr,
			AmountMinor: amountMinor,
			Currency:    currency,
			Description: desc,
			CategoryID:  categoryID,
			OccurredAt:  time.Now(), // Use current time as fallback
		}, sess.AccessToken)
		
		if err != nil {
			h.logger.Error("Failed to create transaction", 
				zap.Int64("telegramID", cb.From.ID),
				zap.Error(err))
			_, _ = h.bot.Request(tgbotapi.NewCallback(cb.ID, "Ошибка"))
			_, _ = h.bot.Send(tgbotapi.NewMessage(cb.Message.Chat.ID, "Не удалось сохранить транзакцию"))
			_ = h.states.ClearState(ctx, cb.From.ID)
			return
		}
		
		// Cleanup draft if present
		if h.drafts != nil {
			if dID, ok := rec.Context["draft_id"].(string); ok && dID != "" {
				_ = h.drafts.Delete(ctx, dID)
			}
		}
		
		// Clear state and send success message
		_ = h.states.ClearState(ctx, cb.From.ID)
		_, _ = h.bot.Request(tgbotapi.NewCallback(cb.ID, "Сохранено"))
		_, _ = h.bot.Send(tgbotapi.NewMessage(cb.Message.Chat.ID, fmt.Sprintf("✅ Сохранено: %s %.2f %s — %s (категория: %s)", typeStr, float64(amountMinor)/100.0, currency, desc, categoryName)))
		metrics.IncTransactionsSaved("ok")
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
	if strings.HasPrefix(data, "help:") {
		helpSection := strings.TrimPrefix(data, "help:")
		_, _ = h.bot.Request(tgbotapi.NewCallback(cb.ID, "Раздел справки"))
		
		// Create proper message with command entities
		messageText := "/help " + helpSection
		message := &tgbotapi.Message{
			Chat: cb.Message.Chat,
			From: cb.From,
			Text: messageText,
			Entities: []tgbotapi.MessageEntity{
				{
					Type:   "bot_command",
					Offset: 0,
					Length: 5, // "/help"
				},
			},
		}
		
		h.handleHelp(ctx, tgbotapi.Update{Message: message})
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
	case "help":
		h.handleHelp(ctx, update)
	case "cancel":
		h.handleCancel(ctx, update)
	default:
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда. Используйте /help для получения справки.")
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
		"/login — вход через OAuth\n"+
		"/logout — выход\n"+
		"/help — подробная справка\n\n"+
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
	_ = h.states.SetState(ctx, update.Message.From.ID, repository.StateWaitingForOAuthEmail, nil, nil)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите email для авторизации через OAuth:")
	_, err := h.bot.Send(msg)
	if err != nil {
		h.logger.Error("failed to send OAuth login email prompt", zap.Error(err))
	}
}

func (h *Handler) handleOAuthEmail(ctx context.Context, update tgbotapi.Update) {
	email := strings.TrimSpace(update.Message.Text)
	
	// Простая валидация email на стороне клиента
	if !isValidEmail(email) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неверный формат email. Пожалуйста, введите корректный email адрес.\n\nПример: user@example.com")
		_, _ = h.bot.Send(msg)
		return
	}
	
	// Generate OAuth auth link
	userAgent := "TelegramBot/1.0"
	ipAddress := "127.0.0.1" // In real implementation, get from request context
	
	authURL, authToken, expiresAt, err := h.auth.GenerateAuthLink(ctx, update.Message.From.ID, email, userAgent, ipAddress)
	if err != nil {
		errorMsg := GetUserFriendlyError(err)
		if IsRetryableError(err) {
			errorMsg += "\n\nПопробуйте снова через несколько секунд."
		} else {
			errorMsg += "\n\nПопробуйте снова /login"
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, errorMsg)
		_, _ = h.bot.Send(msg)
		return
	}
	
	// Store auth token in context for later verification
	ctxMap := map[string]any{
		"email":      email,
		"authToken":  authToken,
		"expiresAt":  expiresAt,
	}
	_ = h.states.SetState(ctx, update.Message.From.ID, repository.StateWaitingForOAuthCode, ctxMap, nil)
	
	// Send auth link to user
	authMessage := fmt.Sprintf("Для авторизации перейдите по ссылке:\n%s\n\nПосле авторизации введите код подтверждения, который появится на странице.", authURL)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, authMessage)
	_, _ = h.bot.Send(msg)
}

func (h *Handler) handleOAuthCode(ctx context.Context, update tgbotapi.Update) {
	rec, _ := h.states.GetState(ctx, update.Message.From.ID)
	if rec == nil || rec.Context == nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Начните с /login")
		_, _ = h.bot.Send(msg)
		return
	}
	
	authToken, _ := rec.Context["authToken"].(string)
	verificationCode := strings.TrimSpace(update.Message.Text)
	
	h.logger.Info("User entered verification code",
		zap.Int64("telegramID", update.Message.From.ID),
		zap.String("verificationCode", verificationCode),
		zap.String("authToken", authToken))
	
	if err := h.auth.VerifyAuthCode(ctx, update.Message.From.ID, authToken, verificationCode); err != nil {
		errorMsg := GetUserFriendlyError(err)
		if IsRetryableError(err) {
			errorMsg += "\n\nПопробуйте снова через несколько секунд."
		} else {
			errorMsg += "\n\nПопробуйте снова /login"
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, errorMsg)
		_, _ = h.bot.Send(msg)
		return
	}
	
	_ = h.states.ClearState(ctx, update.Message.From.ID)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы успешно авторизованы через OAuth!")
	_, _ = h.bot.Send(msg)
}

func (h *Handler) handleLogout(ctx context.Context, update tgbotapi.Update) {
	_ = h.auth.Logout(ctx, update.Message.From.ID)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы вышли из системы")
	_, _ = h.bot.Send(msg)
}

// getSessionWithErrorHandling получает сессию пользователя с понятной обработкой ошибок
func (h *Handler) getSessionWithErrorHandling(ctx context.Context, chatID int64, userID int64) (*repository.UserSession, bool) {
	session, err := h.auth.GetSession(ctx, userID)
	if err != nil {
		errorMsg := GetUserFriendlyError(err)
		if errorMsg == "" {
			errorMsg = "Требуется авторизация"
		}
		errorMsg += "\n\nВыполните вход: /login"
		msg := tgbotapi.NewMessage(chatID, errorMsg)
		_, _ = h.bot.Send(msg)
		return nil, false
	}
	return session, true
}

// Registration is not supported in OAuth flow - users should register through the web interface
func (h *Handler) startRegister(_ context.Context, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Регистрация через бота не поддерживается. Пожалуйста, зарегистрируйтесь через веб-интерфейс.")
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
		
		// Get preferences for locale
		pref, _ := h.prefs.GetPreferences(ctx, update.Message.From.ID)
		locale := "ru"
		if pref != nil && pref.Language != "" {
			locale = pref.Language
		}
		
		var b strings.Builder
		for _, m := range items {
			// Try to get category name by ID
			categoryName, err := h.nameMapper.GetCategoryNameByID(ctx, sess.TenantID, sess.AccessToken, m.CategoryID, domain.TransactionExpense, locale)
			if err != nil || categoryName == "" {
				// Fallback to ID if name not found
				b.WriteString(fmt.Sprintf("%s = %s\n", m.Keyword, m.CategoryID))
			} else {
				b.WriteString(fmt.Sprintf("%s = %s\n", m.Keyword, categoryName))
			}
		}
		_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, b.String()))
		return
	}
	if len(parts) == 1 {
		// show mapping for keyword
		keyword := strings.TrimSpace(parts[0])
		if keyword == "" {
			_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Формат: /map слово = название_категории"))
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
		
		// Get preferences for locale
		pref, _ := h.prefs.GetPreferences(ctx, update.Message.From.ID)
		locale := "ru"
		if pref != nil && pref.Language != "" {
			locale = pref.Language
		}
		
		// Try to get category name by ID
		categoryName, err := h.nameMapper.GetCategoryNameByID(ctx, sess.TenantID, sess.AccessToken, m.CategoryID, domain.TransactionExpense, locale)
		if err != nil || categoryName == "" {
			// Fallback to ID if name not found
			_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("%s = %s", m.Keyword, m.CategoryID)))
		} else {
			_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("%s = %s", m.Keyword, categoryName)))
		}
		return
	}
	if len(parts) == 2 {
		keyword := strings.TrimSpace(parts[0])
		categoryName := strings.TrimSpace(parts[1])
		if keyword == "" || categoryName == "" {
			_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Формат: /map слово = название_категории"))
			return
		}
		sess, err := h.auth.GetSession(ctx, update.Message.From.ID)
		if err != nil {
			_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Сначала выполните вход: /login"))
			return
		}
		
		// Get preferences for locale
		pref, _ := h.prefs.GetPreferences(ctx, update.Message.From.ID)
		locale := "ru"
		if pref != nil && pref.Language != "" {
			locale = pref.Language
		}
		
		// Map category name to ID
		categoryID, err := h.nameMapper.GetCategoryIDByName(ctx, sess.TenantID, sess.AccessToken, categoryName, domain.TransactionExpense, locale)
		if err != nil || categoryID == "" {
			_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Категория не найдена"))
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
	_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Формат: /map слово = название_категории"))
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

func (h *Handler) handleHelp(ctx context.Context, update tgbotapi.Update) {
	args := strings.TrimSpace(update.Message.CommandArguments())
	
	h.logger.Debug("Help command called", 
		zap.String("text", update.Message.Text),
		zap.String("args", args),
		zap.Int("entities_count", len(update.Message.Entities)))
	
	switch args {
	case "auth", "аутентификация":
		h.showAuthHelp(ctx, update)
	case "transactions", "транзакции":
		h.showTransactionsHelp(ctx, update)
	case "categories", "категории":
		h.showCategoriesHelp(ctx, update)
	case "stats", "статистика":
		h.showStatsHelp(ctx, update)
	case "settings", "настройки":
		h.showSettingsHelp(ctx, update)
	case "admin", "админ":
		h.showAdminHelp(ctx, update)
	default:
		h.showMainHelp(ctx, update)
	}
}

func (h *Handler) showMainHelp(_ context.Context, update tgbotapi.Update) {
	text := "🤖 *Справка по командам бота*\n\n" +
		"Выберите раздел для получения подробной информации:\n\n" +
		"🔐 *Аутентификация* - `/help auth`\n" +
		"Вход, регистрация, управление профилем\n\n" +
		"💰 *Транзакции* - `/help transactions`\n" +
		"Добавление транзакций, форматы сообщений\n\n" +
		"🏷️ *Категории* - `/help categories`\n" +
		"Управление категориями и маппингами\n\n" +
		"📊 *Статистика* - `/help stats`\n" +
		"Отчеты и аналитика\n\n" +
		"⚙️ *Настройки* - `/help settings`\n" +
		"Язык, валюта, профиль\n\n" +
		"👨‍💼 *Админ* - `/help admin`\n" +
		"Управление категориями (только в сборке withgrpc)\n\n" +
		"💡 *Быстрый старт:*\n" +
		"1. /start - Начало работы\n" +
		"2. /login - Вход в систему\n" +
		"3. Отправьте транзакцию: \"1000 продукты\""

	kb := ui.CreateHelpKeyboard()
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = kb
	_, _ = h.bot.Send(msg)
}

func (h *Handler) showAuthHelp(_ context.Context, update tgbotapi.Update) {
	h.logger.Debug("showAuthHelp called", 
		zap.Int64("chatID", update.Message.Chat.ID),
		zap.Int64("userID", update.Message.From.ID))
	
	text := `🔐 *Аутентификация и профиль*

/start - Начало работы
Приветственное сообщение с главным меню

/login - Вход в систему
Запускает OAuth аутентификацию через email

/register - Регистрация
Создание нового аккаунта через OAuth

/logout - Выход из системы
Завершение текущей сессии

/profile - Профиль пользователя
Информация о пользователе, настройки

/switch\\_tenant - Переключение организации
Выбор организации для работы

💡 *Для начала работы:*
1\\. /start
2\\. /login
3\\. Следуйте инструкциям для авторизации`

	kb := ui.CreateBackToHelpKeyboard()
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = kb
	
	_, err := h.bot.Send(msg)
	if err != nil {
		h.logger.Error("failed to send auth help message", 
			zap.Int64("chatID", update.Message.Chat.ID),
			zap.Error(err))
	} else {
		h.logger.Debug("auth help message sent successfully", 
			zap.Int64("chatID", update.Message.Chat.ID))
	}
}

func (h *Handler) showTransactionsHelp(_ context.Context, update tgbotapi.Update) {
	text := `💰 *Добавление транзакций*

Бот автоматически распознает сообщения в формате транзакций.

*Формат:* ` + "`[дата] [+]сумма[валюта] описание`" + `

*Примеры:*
• ` + "`1000 продукты`" + ` - Расход 1000 в валюте по умолчанию
• ` + "`+50000 зарплата`" + ` - Доход 50000
• ` + "`01.12 5000 подарок`" + ` - Расход с датой
• ` + "`вчера 100 кофе`" + ` - Расход за вчера
• ` + "`1000₽ продукты`" + ` - Расход в рублях
• ` + "`+50000$ зарплата`" + ` - Доход в долларах

*Поддерживаемые даты:*
• ` + "`сегодня`" + `, ` + "`вчера`" + `, ` + "`позавчера`" + `
• ` + "`DD.MM.YYYY`" + ` (например, ` + "`15.12.2023`" + `)
• ` + "`DD.MM`" + ` (например, ` + "`15.12`" + ` - текущий год)

*Поддерживаемые валюты:*
• Символы: ₽, $, €, £, ¥
• Коды: RUB, USD, EUR, GBP, JPY

*Процесс добавления:*
1. Отправьте транзакцию в нужном формате
2. Если категория не найдена автоматически, выберите из списка
3. Транзакция сохраняется автоматически`

	kb := ui.CreateBackToHelpKeyboard()
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = kb
	_, _ = h.bot.Send(msg)
}

func (h *Handler) showCategoriesHelp(_ context.Context, update tgbotapi.Update) {
	text := "🏷️ *Управление категориями*\n\n" +
		"/categories - Список категорий\n" +
		"Показывает доступные категории для выбора\n\n" +
		"`/map слово = название_категории` - Добавить сопоставление\n" +
		"Создает связь между словом и категорией для автоматической категоризации\n\n" +
		"*Примеры маппингов:*\n" +
		"• `/map кофе = Питание`\n" +
		"• `/map такси = Транспорт`\n" +
		"• `/map продукты = Продукты`\n\n" +
		"`/map слово` - Показать сопоставление\n" +
		"Показывает текущее сопоставление для слова\n\n" +
		"`/map --all` - Показать все сопоставления\n" +
		"Список всех созданных сопоставлений\n\n" +
		"`/unmap слово` - Удалить сопоставление\n" +
		"Удаляет сопоставление для указанного слова\n\n" +
		"*Как работают маппинги:*\n" +
		"1. Бот ищет точные совпадения ключевых слов\n" +
		"2. Если не найдено, ищет частичные совпадения\n" +
		"3. При нескольких совпадениях выбирает с наивысшим приоритетом\n\n" +
		"💡 *Совет:* Создайте маппинги для часто используемых слов"

	kb := ui.CreateBackToHelpKeyboard()
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = kb
	_, _ = h.bot.Send(msg)
}

func (h *Handler) showStatsHelp(_ context.Context, update tgbotapi.Update) {
	h.logger.Debug("showStatsHelp called", 
		zap.Int64("chatID", update.Message.Chat.ID),
		zap.Int64("userID", update.Message.From.ID))
	
	text := "📊 *Статистика и отчеты*\n\n" +
		"`/stats [период]` - Общая статистика\n" +
		"Показывает доходы и расходы за период\n\n" +
		"*Варианты периода:*\n" +
		"• /stats - Текущий месяц\n" +
		"• `/stats 2023\\-12` - Конкретный месяц \\(YYYY\\-MM\\)\n" +
		"• `/stats week` - Текущая неделя\n\n" +
		"`/top\\_categories [период] [лимит]` - Топ категорий\n" +
		"Показывает категории с наибольшими расходами\n\n" +
		"*Примеры:*\n" +
		"• /top\\_categories - Топ\\-5 за текущий месяц\n" +
		"• `/top\\_categories 2023\\-12` - Топ\\-5 за декабрь 2023\n" +
		"• `/top\\_categories week 10` - Топ\\-10 за неделю\n\n" +
		"`/recent [лимит]` - Последние транзакции\n" +
		"Показывает последние транзакции\n\n" +
		"*Примеры:*\n" +
		"• /recent - Последние 10 транзакций\n" +
		"• `/recent 20` - Последние 20 транзакций\n\n" +
		"`/export [период] [лимит]` - Экспорт данных\n" +
		"Экспортирует транзакции в CSV формат\n\n" +
		"*Примеры:*\n" +
		"• /export - Экспорт за текущий месяц\n" +
		"• `/export 2023\\-12` - Экспорт за декабрь 2023\n" +
		"• `/export week 100` - Экспорт 100 транзакций за неделю"

	kb := ui.CreateBackToHelpKeyboard()
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = kb
	
	_, err := h.bot.Send(msg)
	if err != nil {
		h.logger.Error("failed to send stats help message", 
			zap.Int64("chatID", update.Message.Chat.ID),
			zap.Error(err))
	} else {
		h.logger.Debug("stats help message sent successfully", 
			zap.Int64("chatID", update.Message.Chat.ID))
	}
}

func (h *Handler) showSettingsHelp(_ context.Context, update tgbotapi.Update) {
	text := `⚙️ *Настройки*

/language - Выбор языка
Показывает клавиатуру для выбора языка интерфейса
• 🇷🇺 Русский
• 🇺🇸 English

/currency - Настройка валюты
Показывает клавиатуру для выбора валюты по умолчанию
• ₽ RUB
• $ USD
• € EUR
• £ GBP
• ¥ JPY

/profile - Профиль пользователя
Показывает информацию о пользователе:
• UserID и TenantID
• Язык интерфейса
• Валюта по умолчанию
• Статус авторизации

/settings - Общие настройки
Аналогично /profile

💡 *Рекомендации:*
• Установите удобный язык интерфейса
• Выберите основную валюту для транзакций
• Регулярно проверяйте профиль`

	kb := ui.CreateBackToHelpKeyboard()
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = kb
	_, _ = h.bot.Send(msg)
}

func (h *Handler) showAdminHelp(_ context.Context, update tgbotapi.Update) {
	text := "👨‍💼 *Административные команды*\n\n" +
		"*Доступно только в сборке withgrpc*\n\n" +
		"`/create_category code название` - Создать категорию\n" +
		"Создает новую категорию в системе\n\n" +
		"*Пример:*\n" +
		"• `/create_category cat-entertainment Развлечения`\n\n" +
		"`/rename_category category_id новое_название` - Переименовать категорию\n" +
		"Изменяет название существующей категории\n\n" +
		"*Пример:*\n" +
		"• `/rename_category cat-food Еда и напитки`\n\n" +
		"`/delete_category category_id` - Удалить категорию\n" +
		"Удаляет категорию из системы\n\n" +
		"*Пример:*\n" +
		"• `/delete_category cat-old-category`\n\n" +
		"⚠️ *Внимание:* Эти команды доступны только в специальной сборке бота с поддержкой gRPC.\n\n" +
		"💡 *Для обычных пользователей:*\n" +
		"Используйте команды /categories и /map для работы с категориями"

	kb := ui.CreateBackToHelpKeyboard()
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = kb
	_, _ = h.bot.Send(msg)
}

