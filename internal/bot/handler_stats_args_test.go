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

func TestHandler_Stats_Top_WithArgs(t *testing.T) {
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
    chatID := int64(8400)
    userID := int64(12)

    // login
    updLogin := tgbotapi.Update{UpdateID: 1, Message: &tgbotapi.Message{Chat:&tgbotapi.Chat{ID:chatID}, From:&tgbotapi.User{ID:userID}, Text:"/login"}}
    updLogin.Message.Entities = []tgbotapi.MessageEntity{{Type:"bot_command", Offset:0, Length:6}}
    h.HandleUpdate(ctx, updLogin)
    updEmail := updLogin; updEmail.UpdateID = 2; updEmail.Message.Text = "u@e"; updEmail.Message.Entities = nil
    h.HandleUpdate(ctx, updEmail)
    updPass := updLogin; updPass.UpdateID = 3; updPass.Message.Text = "p"; updPass.Message.Entities = nil
    h.HandleUpdate(ctx, updPass)

    // /stats 2025-08
    updStatsMonth := updLogin; updStatsMonth.UpdateID = 4; updStatsMonth.Message.Text = "/stats 2025-08"; updStatsMonth.Message.Entities = []tgbotapi.MessageEntity{{Type:"bot_command", Offset:0, Length:6}}
    h.HandleUpdate(ctx, updStatsMonth)
    // /stats week
    updStatsWeek := updLogin; updStatsWeek.UpdateID = 5; updStatsWeek.Message.Text = "/stats week"; updStatsWeek.Message.Entities = []tgbotapi.MessageEntity{{Type:"bot_command", Offset:0, Length:6}}
    h.HandleUpdate(ctx, updStatsWeek)

    // /top_categories 2025-08 10
    updTop := updLogin; updTop.UpdateID = 6; updTop.Message.Text = "/top_categories 2025-08 10"; updTop.Message.Entities = []tgbotapi.MessageEntity{{Type:"bot_command", Offset:0, Length:14}}
    h.HandleUpdate(ctx, updTop)
    // /top_categories week 100 (should clamp)
    updTopWeek := updLogin; updTopWeek.UpdateID = 7; updTopWeek.Message.Text = "/top_categories week 100"; updTopWeek.Message.Entities = []tgbotapi.MessageEntity{{Type:"bot_command", Offset:0, Length:14}}
    h.HandleUpdate(ctx, updTopWeek)
}


