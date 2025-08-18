package bot

import (
	"context"
	"fmt"
	"strings"

	"budget-bot/internal/repository"
	"budget-bot/internal/bot/ui"
	grpcclient "budget-bot/internal/grpc"
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
}

func NewHandler(bot *tgbotapi.BotAPI, states repository.DialogStateRepository, auth *AuthManager, logger *zap.Logger) *Handler {
	return &Handler{bot: bot, states: states, auth: auth, logger: logger, parser: NewMessageParser(), categories: &grpcclient.StaticCategoryClient{}}
}

func (h *Handler) HandleUpdate(ctx context.Context, update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

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
		amt := float64(parsed.Amount.AmountMinor) / 100.0
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			fmt.Sprintf("Распознано: %s %.2f %s — %s", string(parsed.Type), amt, parsed.Amount.CurrencyCode, parsed.Description))
		_, _ = h.bot.Send(msg)
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
	default:
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Unknown command")
		_, _ = h.bot.Send(msg)
	}
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
	if len(parts) != 2 {
		_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Формат: /map слово = category_id"))
		return
	}
	_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Сопоставление будет реализовано после подключения репозитория в handler"))
}

func (h *Handler) handleUnmap(ctx context.Context, update tgbotapi.Update) {
	keyword := strings.TrimSpace(update.Message.CommandArguments())
	if keyword == "" {
		_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Формат: /unmap слово"))
		return
	}
	_, _ = h.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Удаление сопоставления будет реализовано после подключения репозитория в handler"))
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


