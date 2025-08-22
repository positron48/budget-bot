package bot

import (
    "context"
    "testing"

    "budget-bot/internal/bot/ui"
    grpcclient "budget-bot/internal/grpc"
    "budget-bot/internal/repository"
    "budget-bot/internal/testutil"
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
    "go.uber.org/zap"
)

func setupAuthedHandler(t *testing.T) (*Handler, int64, int64) {
    log := zap.NewNop()
    db := testutil.OpenMigratedSQLite(t)
    sessions := repository.NewSQLiteSessionRepository(db)
    states := repository.NewSQLiteDialogStateRepository(db)
    mappings := repository.NewSQLiteCategoryMappingRepository(db)
    prefs := repository.NewSQLitePreferencesRepository(db)
    drafts := repository.NewSQLiteDraftRepository(db)
    auth := NewOAuthManager(&TestOAuthClient{}, sessions, log, "http://localhost:3000")
    bot := testutil.NewTestBot(t)
    h := NewHandler(bot, states, auth, mappings, nil, log).
        WithPreferences(prefs).
        WithDrafts(drafts).
        WithReportClient(&grpcclient.FakeReportClient{}).
        WithTransactionClient(&grpcclient.FakeTransactionClient{}).
        WithTenantClient(&grpcclient.FakeTenantClient{})
    h.fmt = ui.NewMessageFormatter()
    chatID := int64(8800)
    userID := int64(21)
    // login
    ctx := context.Background()
    updLogin := tgbotapi.Update{UpdateID: 1, Message:&tgbotapi.Message{Chat:&tgbotapi.Chat{ID:chatID}, From:&tgbotapi.User{ID:userID}, Text:"/login"}}; updLogin.Message.Entities = []tgbotapi.MessageEntity{{Type:"bot_command", Offset:0, Length:6}}
    h.HandleUpdate(ctx, updLogin)
    updEmail := updLogin; updEmail.UpdateID = 2; updEmail.Message.Text = "u@e"; updEmail.Message.Entities = nil; h.HandleUpdate(ctx, updEmail)
    updPass := updLogin; updPass.UpdateID = 3; updPass.Message.Text = "p"; updPass.Message.Entities = nil; h.HandleUpdate(ctx, updPass)
    return h, chatID, userID
}

func TestHandler_TopCategories_Default(t *testing.T) {
    h, chatID, userID := setupAuthedHandler(t)
    ctx := context.Background()
    upd := tgbotapi.Update{UpdateID: 10, Message:&tgbotapi.Message{Chat:&tgbotapi.Chat{ID:chatID}, From:&tgbotapi.User{ID:userID}, Text:"/top_categories"}}
    upd.Message.Entities = []tgbotapi.MessageEntity{{Type:"bot_command", Offset:0, Length:14}}
    h.HandleUpdate(ctx, upd)
}

func TestHandler_Recent_WithLimit(t *testing.T) {
    h, chatID, userID := setupAuthedHandler(t)
    ctx := context.Background()
    upd := tgbotapi.Update{UpdateID: 11, Message:&tgbotapi.Message{Chat:&tgbotapi.Chat{ID:chatID}, From:&tgbotapi.User{ID:userID}, Text:"/recent 20"}}
    upd.Message.Entities = []tgbotapi.MessageEntity{{Type:"bot_command", Offset:0, Length:7}}
    h.HandleUpdate(ctx, upd)
}

func TestHandler_Profile(t *testing.T) {
    h, chatID, userID := setupAuthedHandler(t)
    ctx := context.Background()
    upd := tgbotapi.Update{UpdateID: 12, Message:&tgbotapi.Message{Chat:&tgbotapi.Chat{ID:chatID}, From:&tgbotapi.User{ID:userID}, Text:"/profile"}}
    upd.Message.Entities = []tgbotapi.MessageEntity{{Type:"bot_command", Offset:0, Length:8}}
    h.HandleUpdate(ctx, upd)
}

func TestHandler_Unmap_NoArgs_Err(t *testing.T) {
    h, chatID, userID := setupAuthedHandler(t)
    ctx := context.Background()
    upd := tgbotapi.Update{UpdateID: 13, Message:&tgbotapi.Message{Chat:&tgbotapi.Chat{ID:chatID}, From:&tgbotapi.User{ID:userID}, Text:"/unmap"}}
    upd.Message.Entities = []tgbotapi.MessageEntity{{Type:"bot_command", Offset:0, Length:6}}
    h.HandleUpdate(ctx, upd)
}


