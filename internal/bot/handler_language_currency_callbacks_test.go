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

func TestHandler_Language_Currency_Callbacks_NoExistingPrefs(t *testing.T) {
    log := zap.NewNop()
    db := testutil.OpenMigratedSQLite(t)
    sessions := repository.NewSQLiteSessionRepository(db)
    states := repository.NewSQLiteDialogStateRepository(db)
    mappings := repository.NewSQLiteCategoryMappingRepository(db)
    prefs := repository.NewSQLitePreferencesRepository(db)
    auth := NewOAuthManager(&TestOAuthClient{}, sessions, log, "http://localhost:3000")
    bot := testutil.NewTestBot(t)
    h := NewHandler(bot, states, auth, mappings, nil, log).WithPreferences(prefs)
    h.fmt = ui.NewMessageFormatter()

    ctx := context.Background()
    chatID := int64(9000)
    userID := int64(77)

    // language
    updLang := tgbotapi.Update{UpdateID: 1, Message:&tgbotapi.Message{Chat:&tgbotapi.Chat{ID:chatID}, From:&tgbotapi.User{ID:userID}, Text:"/language"}}
    updLang.Message.Entities = []tgbotapi.MessageEntity{{Type:"bot_command", Offset:0, Length:9}}
    h.HandleUpdate(ctx, updLang)
    cbLang := tgbotapi.Update{UpdateID: 2, CallbackQuery:&tgbotapi.CallbackQuery{ID:"cb1", From:&tgbotapi.User{ID:userID}, Message:&tgbotapi.Message{Chat:&tgbotapi.Chat{ID:chatID}}, Data:"lang:en"}}
    h.HandleUpdate(ctx, cbLang)

    // currency
    updCur := tgbotapi.Update{UpdateID: 3, Message:&tgbotapi.Message{Chat:&tgbotapi.Chat{ID:chatID}, From:&tgbotapi.User{ID:userID}, Text:"/currency"}}
    updCur.Message.Entities = []tgbotapi.MessageEntity{{Type:"bot_command", Offset:0, Length:9}}
    h.HandleUpdate(ctx, updCur)
    cbCur := tgbotapi.Update{UpdateID: 4, CallbackQuery:&tgbotapi.CallbackQuery{ID:"cb2", From:&tgbotapi.User{ID:userID}, Message:&tgbotapi.Message{Chat:&tgbotapi.Chat{ID:chatID}}, Data:"cur:USD"}}
    h.HandleUpdate(ctx, cbCur)
}


