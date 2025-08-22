package bot

import (
    "context"
    "testing"

    "budget-bot/internal/bot/ui"
    "budget-bot/internal/domain"
    "budget-bot/internal/repository"
    "budget-bot/internal/testutil"
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
    "go.uber.org/zap"
)

type emptyCatClient struct{}
func (e *emptyCatClient) ListCategories(_ context.Context, _ string, _ string, _ domain.TransactionType, _ ...string) ([]*domain.Category, error) { return []*domain.Category{}, nil }
func (e *emptyCatClient) CreateCategory(_ context.Context, _ string, _ string, name string, _ string) (*domain.Category, error) { return &domain.Category{ID:"id", Name:name}, nil }
func (e *emptyCatClient) UpdateCategoryName(_ context.Context, _ string, id string, name string, _ string) (*domain.Category, error) { return &domain.Category{ID:id, Name:name}, nil }
func (e *emptyCatClient) DeleteCategory(_ context.Context, _ string, _ string) error { return nil }

func TestHandler_Categories_EmptyList(t *testing.T) {
    log := zap.NewNop()
    db := testutil.OpenMigratedSQLite(t)
    sessions := repository.NewSQLiteSessionRepository(db)
    states := repository.NewSQLiteDialogStateRepository(db)
    mappings := repository.NewSQLiteCategoryMappingRepository(db)
    prefs := repository.NewSQLitePreferencesRepository(db)
    auth := NewOAuthManager(&TestOAuthClient{}, sessions, log, "http://localhost:3000")
    bot := testutil.NewTestBot(t)
    h := NewHandler(bot, states, auth, mappings, &emptyCatClient{}, log).WithPreferences(prefs)
    h.fmt = ui.NewMessageFormatter()

    ctx := context.Background()
    chatID := int64(9100)
    userID := int64(88)
    // login
    updLogin := tgbotapi.Update{UpdateID: 1, Message:&tgbotapi.Message{Chat:&tgbotapi.Chat{ID:chatID}, From:&tgbotapi.User{ID:userID}, Text:"/login"}}; updLogin.Message.Entities = []tgbotapi.MessageEntity{{Type:"bot_command", Offset:0, Length:6}}
    h.HandleUpdate(ctx, updLogin)
    updEmail := updLogin; updEmail.UpdateID = 2; updEmail.Message.Text = "u@e"; updEmail.Message.Entities = nil; h.HandleUpdate(ctx, updEmail)
    updPass := updLogin; updPass.UpdateID = 3; updPass.Message.Text = "p"; updPass.Message.Entities = nil; h.HandleUpdate(ctx, updPass)

    upd := updLogin; upd.UpdateID = 4; upd.Message.Text = "/categories"; upd.Message.Entities = []tgbotapi.MessageEntity{{Type:"bot_command", Offset:0, Length:11}}
    h.HandleUpdate(ctx, upd)
}


