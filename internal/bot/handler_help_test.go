package bot

import (
	"context"
	"testing"

	"budget-bot/internal/repository"
	"budget-bot/internal/testutil"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

func TestHandler_HelpCommands(t *testing.T) {
	log := zap.NewNop()
	db := testutil.OpenMigratedSQLite(t)
	sessions := repository.NewSQLiteSessionRepository(db)
	states := repository.NewSQLiteDialogStateRepository(db)
	mappings := repository.NewSQLiteCategoryMappingRepository(db)
	prefs := repository.NewSQLitePreferencesRepository(db)
	auth := NewOAuthManager(&TestOAuthClient{}, sessions, log, "http://localhost:3000")
	bot := testutil.NewTestBot(t)

	h := NewHandler(bot, states, auth, mappings, nil, log).
		WithPreferences(prefs)

	ctx := context.Background()
	chatID := int64(6000)
	userID := int64(99)

	// Test main help
	upd := tgbotapi.Update{
		UpdateID: 1,
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: chatID},
			From: &tgbotapi.User{ID: userID},
			Text: "/help",
		},
	}
	upd.Message.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 5}}
	h.HandleUpdate(ctx, upd)

	// Test help with section
	updAuth := tgbotapi.Update{
		UpdateID: 2,
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: chatID},
			From: &tgbotapi.User{ID: userID},
			Text: "/help auth",
		},
	}
	updAuth.Message.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 5}}
	h.HandleUpdate(ctx, updAuth)

	// Test help with transactions section
	updTx := tgbotapi.Update{
		UpdateID: 3,
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: chatID},
			From: &tgbotapi.User{ID: userID},
			Text: "/help transactions",
		},
	}
	updTx.Message.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 5}}
	h.HandleUpdate(ctx, updTx)

	// Test help with categories section
	updCat := tgbotapi.Update{
		UpdateID: 4,
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: chatID},
			From: &tgbotapi.User{ID: userID},
			Text: "/help categories",
		},
	}
	updCat.Message.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 5}}
	h.HandleUpdate(ctx, updCat)

	// Test help with stats section
	updStats := tgbotapi.Update{
		UpdateID: 5,
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: chatID},
			From: &tgbotapi.User{ID: userID},
			Text: "/help stats",
		},
	}
	updStats.Message.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 5}}
	h.HandleUpdate(ctx, updStats)

	// Test help with settings section
	updSettings := tgbotapi.Update{
		UpdateID: 6,
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: chatID},
			From: &tgbotapi.User{ID: userID},
			Text: "/help settings",
		},
	}
	updSettings.Message.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 5}}
	h.HandleUpdate(ctx, updSettings)

	// Test help with admin section
	updAdmin := tgbotapi.Update{
		UpdateID: 7,
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: chatID},
			From: &tgbotapi.User{ID: userID},
			Text: "/help admin",
		},
	}
	updAdmin.Message.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 5}}
	h.HandleUpdate(ctx, updAdmin)

	// Test help with unknown section (should show main help)
	updUnknown := tgbotapi.Update{
		UpdateID: 8,
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: chatID},
			From: &tgbotapi.User{ID: userID},
			Text: "/help unknown",
		},
	}
	updUnknown.Message.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 5}}
	h.HandleUpdate(ctx, updUnknown)
}

func TestHandler_HelpCallbacks(t *testing.T) {
	log := zap.NewNop()
	db := testutil.OpenMigratedSQLite(t)
	sessions := repository.NewSQLiteSessionRepository(db)
	states := repository.NewSQLiteDialogStateRepository(db)
	mappings := repository.NewSQLiteCategoryMappingRepository(db)
	prefs := repository.NewSQLitePreferencesRepository(db)
	auth := NewOAuthManager(&TestOAuthClient{}, sessions, log, "http://localhost:3000")
	bot := testutil.NewTestBot(t)

	h := NewHandler(bot, states, auth, mappings, nil, log).
		WithPreferences(prefs)

	ctx := context.Background()
	chatID := int64(6000)
	userID := int64(99)

	// Test help callback for auth section
	cbAuth := tgbotapi.Update{
		UpdateID: 1,
		CallbackQuery: &tgbotapi.CallbackQuery{
			ID:   "cb1",
			From: &tgbotapi.User{ID: userID},
			Message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: chatID},
			},
			Data: "help:auth",
		},
	}
	h.HandleUpdate(ctx, cbAuth)

	// Test help callback for transactions section
	cbTx := tgbotapi.Update{
		UpdateID: 2,
		CallbackQuery: &tgbotapi.CallbackQuery{
			ID:   "cb2",
			From: &tgbotapi.User{ID: userID},
			Message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: chatID},
			},
			Data: "help:transactions",
		},
	}
	h.HandleUpdate(ctx, cbTx)

	// Test help callback for back to main help
	cbBack := tgbotapi.Update{
		UpdateID: 3,
		CallbackQuery: &tgbotapi.CallbackQuery{
			ID:   "cb3",
			From: &tgbotapi.User{ID: userID},
			Message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: chatID},
			},
			Data: "help:",
		},
	}
	h.HandleUpdate(ctx, cbBack)
}

func TestHandler_HelpCallbackProcessing(t *testing.T) {
	log := zap.NewNop()
	db := testutil.OpenMigratedSQLite(t)
	sessions := repository.NewSQLiteSessionRepository(db)
	states := repository.NewSQLiteDialogStateRepository(db)
	mappings := repository.NewSQLiteCategoryMappingRepository(db)
	prefs := repository.NewSQLitePreferencesRepository(db)
	auth := NewOAuthManager(&TestOAuthClient{}, sessions, log, "http://localhost:3000")
	bot := testutil.NewTestBot(t)

	h := NewHandler(bot, states, auth, mappings, nil, log).
		WithPreferences(prefs)

	ctx := context.Background()
	chatID := int64(6000)
	userID := int64(99)

	// Test help callback for auth section - this should work
	cbAuth := tgbotapi.Update{
		UpdateID: 1,
		CallbackQuery: &tgbotapi.CallbackQuery{
			ID:   "cb1",
			From: &tgbotapi.User{ID: userID},
			Message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: chatID},
			},
			Data: "help:auth",
		},
	}
	h.HandleUpdate(ctx, cbAuth)

	// Test help callback for empty section (back to main)
	cbMain := tgbotapi.Update{
		UpdateID: 2,
		CallbackQuery: &tgbotapi.CallbackQuery{
			ID:   "cb2",
			From: &tgbotapi.User{ID: userID},
			Message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: chatID},
			},
			Data: "help:",
		},
	}
	h.HandleUpdate(ctx, cbMain)

	// Test help callback for transactions section
	cbTx := tgbotapi.Update{
		UpdateID: 3,
		CallbackQuery: &tgbotapi.CallbackQuery{
			ID:   "cb3",
			From: &tgbotapi.User{ID: userID},
			Message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: chatID},
			},
			Data: "help:transactions",
		},
	}
	h.HandleUpdate(ctx, cbTx)
}
