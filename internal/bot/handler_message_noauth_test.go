package bot

import (
    "context"
    "testing"

    "budget-bot/internal/bot/ui"
    "budget-bot/internal/repository"
    "budget-bot/internal/testutil"
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
    "go.uber.org/zap"
)

func TestHandler_MessageParse_NoAuth_Echo(t *testing.T) {
    log := zap.NewNop()
    db := testutil.OpenMigratedSQLite(t)
    states := repository.NewSQLiteDialogStateRepository(db)
    sessions := repository.NewSQLiteSessionRepository(db)
    mappings := repository.NewSQLiteCategoryMappingRepository(db)
    bot := testutil.NewTestBot(t)
    auth := NewOAuthManager(&TestOAuthClient{}, sessions, log, "http://localhost:3000")
    h := NewHandler(bot, states, auth, mappings, nil, log)
    h.fmt = ui.NewMessageFormatter()

    ctx := context.Background()
    chatID := int64(8600)
    userID := int64(1)
    upd := tgbotapi.Update{UpdateID: 1, Message:&tgbotapi.Message{Chat:&tgbotapi.Chat{ID:chatID}, From:&tgbotapi.User{ID:userID}, Text:"100 кофе"}}
    h.HandleUpdate(ctx, upd)
}

func TestHandler_MessageParse_Invalid_Feedback(t *testing.T) {
    log := zap.NewNop()
    db := testutil.OpenMigratedSQLite(t)
    states := repository.NewSQLiteDialogStateRepository(db)
    sessions := repository.NewSQLiteSessionRepository(db)
    mappings := repository.NewSQLiteCategoryMappingRepository(db)
    bot := testutil.NewTestBot(t)
    auth := NewOAuthManager(&TestOAuthClient{}, sessions, log, "http://localhost:3000")
    h := NewHandler(bot, states, auth, mappings, nil, log)
    h.fmt = ui.NewMessageFormatter()

    ctx := context.Background()
    chatID := int64(8610)
    userID := int64(2)
    upd := tgbotapi.Update{UpdateID: 1, Message:&tgbotapi.Message{Chat:&tgbotapi.Chat{ID:chatID}, From:&tgbotapi.User{ID:userID}, Text:"просто слова без суммы"}}
    h.HandleUpdate(ctx, upd)
}


