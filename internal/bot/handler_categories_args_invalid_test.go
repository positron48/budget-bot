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

// Validate invalid arguments paths for category commands
func TestHandler_CategoryCommands_InvalidArgs(t *testing.T) {
    log := zap.NewNop()
    db := testutil.OpenMigratedSQLite(t)
    sessions := repository.NewSQLiteSessionRepository(db)
    states := repository.NewSQLiteDialogStateRepository(db)
    mappings := repository.NewSQLiteCategoryMappingRepository(db)
    auth := NewOAuthManager(&TestOAuthClient{}, sessions, log, "http://localhost:3000")
    bot := testutil.NewTestBot(t)
    h := NewHandler(bot, states, auth, mappings, nil, log)
    h.fmt = ui.NewMessageFormatter()

    ctx := context.Background()
    chatID := int64(8900)
    userID := int64(50)
    // login
    updLogin := tgbotapi.Update{UpdateID: 1, Message:&tgbotapi.Message{Chat:&tgbotapi.Chat{ID:chatID}, From:&tgbotapi.User{ID:userID}, Text:"/login"}}; updLogin.Message.Entities = []tgbotapi.MessageEntity{{Type:"bot_command", Offset:0, Length:6}}
    h.HandleUpdate(ctx, updLogin)
    updEmail := updLogin; updEmail.UpdateID = 2; updEmail.Message.Text = "u@e"; updEmail.Message.Entities = nil; h.HandleUpdate(ctx, updEmail)
    updPass := updLogin; updPass.UpdateID = 3; updPass.Message.Text = "p"; updPass.Message.Entities = nil; h.HandleUpdate(ctx, updPass)

    // invalid create: missing args
    updCreate := updLogin; updCreate.UpdateID = 4; updCreate.Message.Text = "/create_category"; updCreate.Message.Entities = []tgbotapi.MessageEntity{{Type:"bot_command", Offset:0, Length:16}}
    h.HandleUpdate(ctx, updCreate)
    // invalid rename: missing new name
    updRename := updLogin; updRename.UpdateID = 5; updRename.Message.Text = "/rename_category onlyid"; updRename.Message.Entities = []tgbotapi.MessageEntity{{Type:"bot_command", Offset:0, Length:16}}
    h.HandleUpdate(ctx, updRename)
    // invalid delete: missing id
    updDelete := updLogin; updDelete.UpdateID = 6; updDelete.Message.Text = "/delete_category"; updDelete.Message.Entities = []tgbotapi.MessageEntity{{Type:"bot_command", Offset:0, Length:16}}
    h.HandleUpdate(ctx, updDelete)
}


