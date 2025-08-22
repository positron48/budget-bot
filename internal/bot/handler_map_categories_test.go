package bot

import (
    "context"
    "testing"
    "time"

    "budget-bot/internal/bot/ui"
    grpcclient "budget-bot/internal/grpc"
    "budget-bot/internal/repository"
    "budget-bot/internal/testutil"
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
    "go.uber.org/zap"
)

func TestHandler_Map_Unmap_Categories(t *testing.T) {
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

    ctx := context.Background()
    chatID := int64(4000)
    userID := int64(66)

    // login to create session
    updStart := tgbotapi.Update{UpdateID: 1, Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: chatID}, From: &tgbotapi.User{ID: userID}, Text: "/login"}}
    updStart.Message.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 6}}
    h.HandleUpdate(ctx, updStart)
    updEmail := updStart; updEmail.UpdateID = 2; updEmail.Message.Text = "u@e"; updEmail.Message.Entities = nil
    h.HandleUpdate(ctx, updEmail)
    updPass := updStart; updPass.UpdateID = 3; updPass.Message.Text = "p"; updPass.Message.Entities = nil
    h.HandleUpdate(ctx, updPass)

    // /map слово = cat-id
    updMap := updStart; updMap.UpdateID = 4; updMap.Message.Text = "/map кофе = cat-food"; updMap.Message.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 4}}
    h.HandleUpdate(ctx, updMap)

    // /map кофе (просмотр)
    updMapShow := updStart; updMapShow.UpdateID = 5; updMapShow.Message.Text = "/map кофе"; updMapShow.Message.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 4}}
    h.HandleUpdate(ctx, updMapShow)

    // /map --all
    updMapAll := updStart; updMapAll.UpdateID = 6; updMapAll.Message.Text = "/map --all"; updMapAll.Message.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 4}}
    h.HandleUpdate(ctx, updMapAll)

    // /unmap слово
    updUnmap := updStart; updUnmap.UpdateID = 7; updUnmap.Message.Text = "/unmap кофе"; updUnmap.Message.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 6}}
    h.HandleUpdate(ctx, updUnmap)
    _ = time.Second
}


