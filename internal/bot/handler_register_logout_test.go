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

func TestHandler_Register_Then_Logout(t *testing.T) {
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
    chatID := int64(2000)
    userID := int64(77)

    // /register -> email -> password -> name
    upd := tgbotapi.Update{UpdateID: 1, Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: chatID}, From: &tgbotapi.User{ID: userID}, Text: "/register"}}
    upd.Message.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 9}}
    h.HandleUpdate(ctx, upd)
    updEmail := upd; updEmail.UpdateID = 2; updEmail.Message.Text = "user@example.com"; updEmail.Message.Entities = nil
    h.HandleUpdate(ctx, updEmail)
    updPass := upd; updPass.UpdateID = 3; updPass.Message.Text = "pass"; updPass.Message.Entities = nil
    h.HandleUpdate(ctx, updPass)
    updName := upd; updName.UpdateID = 4; updName.Message.Text = "John"; updName.Message.Entities = nil
    h.HandleUpdate(ctx, updName)

    if _, err := sessions.GetSession(ctx, userID); err != nil {
        t.Fatalf("expected session after register: %v", err)
    }

    // /logout
    updLo := upd; updLo.UpdateID = 5; updLo.Message.Text = "/logout"; updLo.Message.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 7}}
    h.HandleUpdate(ctx, updLo)
    if _, err := sessions.GetSession(ctx, userID); err == nil {
        t.Fatalf("expected no session after logout")
    }
    _ = time.Second
}


