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

func TestHandler_SwitchTenant(t *testing.T) {
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
    chatID := int64(5000)
    userID := int64(55)

    // login to create session
    updStart := tgbotapi.Update{UpdateID: 1, Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: chatID}, From: &tgbotapi.User{ID: userID}, Text: "/login"}}
    updStart.Message.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 6}}
    h.HandleUpdate(ctx, updStart)
    updEmail := updStart; updEmail.UpdateID = 2; updEmail.Message.Text = "u@e"; updEmail.Message.Entities = nil
    h.HandleUpdate(ctx, updEmail)
    updPass := updStart; updPass.UpdateID = 3; updPass.Message.Text = "p"; updPass.Message.Entities = nil
    h.HandleUpdate(ctx, updPass)

    // /switch_tenant shows tenants
    updSw := updStart; updSw.UpdateID = 4; updSw.Message.Text = "/switch_tenant"; updSw.Message.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 14}}
    h.HandleUpdate(ctx, updSw)

    // choose tenant callback
    cb := tgbotapi.Update{UpdateID: 5, CallbackQuery: &tgbotapi.CallbackQuery{ID: "cbt", From: &tgbotapi.User{ID: userID}, Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: chatID}}, Data: "tenant:tenant-2"}}
    h.HandleUpdate(ctx, cb)

    s, err := sessions.GetSession(ctx, userID)
    if err != nil { t.Fatalf("session: %v", err) }
    if s.TenantID != "tenant-2" { t.Fatalf("tenant not switched: %s", s.TenantID) }
    _ = time.Second
}


