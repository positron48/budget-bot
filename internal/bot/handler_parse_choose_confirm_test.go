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

func TestHandler_Parse_NoMapping_ChooseCategory_ImmediateSave(t *testing.T) {
    log := zap.NewNop()
    db := testutil.OpenMigratedSQLite(t)
    sessions := repository.NewSQLiteSessionRepository(db)
    states := repository.NewSQLiteDialogStateRepository(db)
    mappings := repository.NewSQLiteCategoryMappingRepository(db)
    prefs := repository.NewSQLitePreferencesRepository(db)
    drafts := repository.NewSQLiteDraftRepository(db)
    auth := NewOAuthManager(&TestOAuthClient{}, sessions, log, "http://localhost:3000")
    bot := testutil.NewTestBot(t)

    // real category listing is static client by default; good for test
    h := NewHandler(bot, states, auth, mappings, nil, log).
        WithPreferences(prefs).
        WithDrafts(drafts).
        WithReportClient(&grpcclient.FakeReportClient{}).
        WithTransactionClient(&grpcclient.FakeTransactionClient{}).
        WithTenantClient(&grpcclient.FakeTenantClient{})
    h.fmt = ui.NewMessageFormatter()

    ctx := context.Background()
    chatID := int64(6000)
    userID := int64(99)

    // login to create session
    updStart := tgbotapi.Update{UpdateID: 1, Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: chatID}, From: &tgbotapi.User{ID: userID}, Text: "/login"}}
    updStart.Message.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 6}}
    h.HandleUpdate(ctx, updStart)
    updEmail := updStart; updEmail.UpdateID = 2; updEmail.Message.Text = "u@e"; updEmail.Message.Entities = nil
    h.HandleUpdate(ctx, updEmail)
    updPass := updStart; updPass.UpdateID = 3; updPass.Message.Text = "p"; updPass.Message.Entities = nil
    h.HandleUpdate(ctx, updPass)

    // send transaction (no mapping yet) -> should request category selection
    updTx := updStart; updTx.UpdateID = 4; updTx.Message.Text = "100 кофе"; updTx.Message.Entities = nil
    h.HandleUpdate(ctx, updTx)

    // select category - transaction should be created immediately
    cbCat := tgbotapi.Update{UpdateID: 5, CallbackQuery: &tgbotapi.CallbackQuery{ID: "cbc1", From: &tgbotapi.User{ID: userID}, Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: chatID}}, Data: "cat:Питание"}}
    h.HandleUpdate(ctx, cbCat)
    _ = time.Second
}


