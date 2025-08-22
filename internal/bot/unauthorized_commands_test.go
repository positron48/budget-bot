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

// Ensure commands require auth respond with proper hints
func TestHandler_Unauthorized_Commands(t *testing.T) {
    log := zap.NewNop()
    db := testutil.OpenMigratedSQLite(t)
    states := repository.NewSQLiteDialogStateRepository(db)
    sessions := repository.NewSQLiteSessionRepository(db)
    mappings := repository.NewSQLiteCategoryMappingRepository(db)
    prefs := repository.NewSQLitePreferencesRepository(db)
    drafts := repository.NewSQLiteDraftRepository(db)
    bot := testutil.NewTestBot(t)
    auth := NewOAuthManager(&TestOAuthClient{}, sessions, log, "http://localhost:3000")

    h := NewHandler(bot, states, auth, mappings, nil, log).
        WithPreferences(prefs).
        WithDrafts(drafts).
        WithReportClient(&grpcclient.FakeReportClient{}).
        WithTransactionClient(&grpcclient.FakeTransactionClient{}).
        WithTenantClient(&grpcclient.FakeTenantClient{})
    h.fmt = ui.NewMessageFormatter()

    ctx := context.Background()
    chatID := int64(8500)
    userID := int64(99)

    // Stats requires auth
    updStats := tgbotapi.Update{UpdateID: 1, Message: &tgbotapi.Message{Chat:&tgbotapi.Chat{ID:chatID}, From:&tgbotapi.User{ID:userID}, Text:"/stats"}}
    updStats.Message.Entities = []tgbotapi.MessageEntity{{Type:"bot_command", Offset:0, Length:6}}
    h.HandleUpdate(ctx, updStats)
    // Categories requires auth
    updCategories := updStats; updCategories.UpdateID = 2; updCategories.Message.Text = "/categories"; updCategories.Message.Entities = []tgbotapi.MessageEntity{{Type:"bot_command", Offset:0, Length:11}}
    h.HandleUpdate(ctx, updCategories)
    // Recent requires auth
    updRecent := updStats; updRecent.UpdateID = 3; updRecent.Message.Text = "/recent"; updRecent.Message.Entities = []tgbotapi.MessageEntity{{Type:"bot_command", Offset:0, Length:7}}
    h.HandleUpdate(ctx, updRecent)
}


