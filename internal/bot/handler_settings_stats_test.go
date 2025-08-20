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

func TestHandler_Language_Currency_Stats_Top_Recent_Export(t *testing.T) {
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
    chatID := int64(3000)
    userID := int64(88)

    // login to create session
    updStart := tgbotapi.Update{UpdateID: 1, Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: chatID}, From: &tgbotapi.User{ID: userID}, Text: "/login"}}
    updStart.Message.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 6}}
    h.HandleUpdate(ctx, updStart)
    updEmail := updStart; updEmail.UpdateID = 2; updEmail.Message.Text = "u@e"; updEmail.Message.Entities = nil
    h.HandleUpdate(ctx, updEmail)
    updPass := updStart; updPass.UpdateID = 3; updPass.Message.Text = "p"; updPass.Message.Entities = nil
    h.HandleUpdate(ctx, updPass)

    // /language -> callback
    updLang := updStart; updLang.UpdateID = 4; updLang.Message.Text = "/language"; updLang.Message.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 9}}
    h.HandleUpdate(ctx, updLang)
    cbLang := tgbotapi.Update{UpdateID: 5, CallbackQuery: &tgbotapi.CallbackQuery{ID: "cbl", From: &tgbotapi.User{ID: userID}, Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: chatID}}, Data: "lang:en"}}
    h.HandleUpdate(ctx, cbLang)

    // /currency -> callback
    updCur := updStart; updCur.UpdateID = 6; updCur.Message.Text = "/currency"; updCur.Message.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 9}}
    h.HandleUpdate(ctx, updCur)
    cbCur := tgbotapi.Update{UpdateID: 7, CallbackQuery: &tgbotapi.CallbackQuery{ID: "cbc", From: &tgbotapi.User{ID: userID}, Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: chatID}}, Data: "cur:USD"}}
    h.HandleUpdate(ctx, cbCur)

    // stats; should use fake report
    updStats := updStart; updStats.UpdateID = 8; updStats.Message.Text = "/stats"; updStats.Message.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 6}}
    h.HandleUpdate(ctx, updStats)

    // top_categories
    updTop := updStart; updTop.UpdateID = 9; updTop.Message.Text = "/top_categories"; updTop.Message.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 14}}
    h.HandleUpdate(ctx, updTop)

    // recent
    updRecent := updStart; updRecent.UpdateID = 10; updRecent.Message.Text = "/recent"; updRecent.Message.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 7}}
    h.HandleUpdate(ctx, updRecent)

    // export
    updExport := updStart; updExport.UpdateID = 11; updExport.Message.Text = "/export"; updExport.Message.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 7}}
    h.HandleUpdate(ctx, updExport)
    _ = time.Second
}


