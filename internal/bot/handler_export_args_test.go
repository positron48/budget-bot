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

func TestHandler_Export_WithArgs(t *testing.T) {
    log := zap.NewNop()
    db := testutil.OpenMigratedSQLite(t)
    sessions := repository.NewSQLiteSessionRepository(db)
    states := repository.NewSQLiteDialogStateRepository(db)
    mappings := repository.NewSQLiteCategoryMappingRepository(db)
    prefs := repository.NewSQLitePreferencesRepository(db)
    drafts := repository.NewSQLiteDraftRepository(db)
    auth := NewAuthManager(&fakeAuthClient{}, sessions, log)
    bot := testutil.NewTestBot(t)

    h := NewHandler(bot, states, auth, mappings, nil, log).
        WithPreferences(prefs).
        WithDrafts(drafts).
        WithReportClient(&grpcclient.FakeReportClient{}).
        WithTransactionClient(&grpcclient.FakeTransactionClient{}).
        WithTenantClient(&grpcclient.FakeTenantClient{})
    h.fmt = ui.NewMessageFormatter()

    ctx := context.Background()
    chatID := int64(8700)
    userID := int64(3)
    // login
    updLogin := tgbotapi.Update{UpdateID: 1, Message:&tgbotapi.Message{Chat:&tgbotapi.Chat{ID:chatID}, From:&tgbotapi.User{ID:userID}, Text:"/login"}}; updLogin.Message.Entities = []tgbotapi.MessageEntity{{Type:"bot_command", Offset:0, Length:6}}
    h.HandleUpdate(ctx, updLogin)
    updEmail := updLogin; updEmail.UpdateID = 2; updEmail.Message.Text = "u@e"; updEmail.Message.Entities = nil; h.HandleUpdate(ctx, updEmail)
    updPass := updLogin; updPass.UpdateID = 3; updPass.Message.Text = "p"; updPass.Message.Entities = nil; h.HandleUpdate(ctx, updPass)

    // /export 2025-08 500
    updExp := updLogin; updExp.UpdateID = 4; updExp.Message.Text = "/export 2025-08 500"; updExp.Message.Entities = []tgbotapi.MessageEntity{{Type:"bot_command", Offset:0, Length:7}}
    h.HandleUpdate(ctx, updExp)
    // /export week 99999 (должно ограничить по максимуму)
    updExp2 := updLogin; updExp2.UpdateID = 5; updExp2.Message.Text = "/export week 99999"; updExp2.Message.Entities = []tgbotapi.MessageEntity{{Type:"bot_command", Offset:0, Length:7}}
    h.HandleUpdate(ctx, updExp2)
}


