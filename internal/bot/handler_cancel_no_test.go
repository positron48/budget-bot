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



func TestHandler_Cancel_And_ConfirmNo(t *testing.T) {
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
    chatID := int64(8100)
    userID := int64(98)

    // cancel should clear state
    updCancel := tgbotapi.Update{UpdateID: 1, Message: &tgbotapi.Message{Chat:&tgbotapi.Chat{ID:chatID}, From:&tgbotapi.User{ID:userID}, Text:"/cancel"}}
    updCancel.Message.Entities = []tgbotapi.MessageEntity{{Type:"bot_command", Offset:0, Length:7}}
    h.HandleUpdate(ctx, updCancel)

    // simulate pending confirmation and then press no
    updLogin := tgbotapi.Update{UpdateID: 2, Message: &tgbotapi.Message{Chat:&tgbotapi.Chat{ID:chatID}, From:&tgbotapi.User{ID:userID}, Text:"/login"}}
    updLogin.Message.Entities = []tgbotapi.MessageEntity{{Type:"bot_command", Offset:0, Length:6}}
    h.HandleUpdate(ctx, updLogin)
    updEmail := updLogin; updEmail.UpdateID = 3; updEmail.Message.Text = "u@e"; updEmail.Message.Entities = nil
    h.HandleUpdate(ctx, updEmail)
    updPass := updLogin; updPass.UpdateID = 4; updPass.Message.Text = "p"; updPass.Message.Entities = nil
    h.HandleUpdate(ctx, updPass)
    updTx := updLogin; updTx.UpdateID = 5; updTx.Message.Text = "100 кофе"; updTx.Message.Entities = nil
    h.HandleUpdate(ctx, updTx)
    // choose category
    cbCat := tgbotapi.Update{UpdateID: 6, CallbackQuery: &tgbotapi.CallbackQuery{ID:"cb", From:&tgbotapi.User{ID:userID}, Message:&tgbotapi.Message{Chat:&tgbotapi.Chat{ID:chatID}}, Data:"cat:Питание"}}
    h.HandleUpdate(ctx, cbCat)
    // press no
    cbNo := tgbotapi.Update{UpdateID: 7, CallbackQuery: &tgbotapi.CallbackQuery{ID:"cb2", From:&tgbotapi.User{ID:userID}, Message:&tgbotapi.Message{Chat:&tgbotapi.Chat{ID:chatID}}, Data:"confirm:no"}}
    h.HandleUpdate(ctx, cbNo)
    _ = time.Second
}


