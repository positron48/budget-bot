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

func TestHandler_handleStats(t *testing.T) {
	log := zap.NewNop()
	db := testutil.OpenMigratedSQLite(t)
	states := repository.NewSQLiteDialogStateRepository(db)
	sessions := repository.NewSQLiteSessionRepository(db)
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
	chatID := int64(1000)
	userID := int64(42)

	t.Run("no session", func(t *testing.T) {
		upd := tgbotapi.Update{
			UpdateID: 1,
			Message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: chatID},
				From: &tgbotapi.User{ID: userID},
				Text: "/stats",
			},
		}
		upd.Message.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 6}}

		// Should not panic
		h.handleStats(ctx, upd)
	})

	t.Run("with session - current month", func(t *testing.T) {
		// Create session first
		sess := &repository.UserSession{
			TelegramID:   userID,
			UserID:       "test-user",
			TenantID:     "test-tenant",
			AccessToken:  "test-token",
			RefreshToken: "test-refresh",
		}
		err := sessions.SaveSession(ctx, sess)
		if err != nil {
			t.Fatalf("Failed to save session: %v", err)
		}

		upd := tgbotapi.Update{
			UpdateID: 2,
			Message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: chatID},
				From: &tgbotapi.User{ID: userID},
				Text: "/stats",
			},
		}
		upd.Message.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 6}}

		// Should not panic
		h.handleStats(ctx, upd)
	})

	t.Run("with session - specific month", func(t *testing.T) {
		// Create session first
		sess := &repository.UserSession{
			TelegramID:   userID,
			UserID:       "test-user",
			TenantID:     "test-tenant",
			AccessToken:  "test-token",
			RefreshToken: "test-refresh",
		}
		err := sessions.SaveSession(ctx, sess)
		if err != nil {
			t.Fatalf("Failed to save session: %v", err)
		}

		upd := tgbotapi.Update{
			UpdateID: 3,
			Message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: chatID},
				From: &tgbotapi.User{ID: userID},
				Text: "/stats 2023-12",
			},
		}
		upd.Message.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 6}}

		// Should not panic
		h.handleStats(ctx, upd)
	})

	t.Run("with session - week", func(t *testing.T) {
		// Create session first
		sess := &repository.UserSession{
			TelegramID:   userID,
			UserID:       "test-user",
			TenantID:     "test-tenant",
			AccessToken:  "test-token",
			RefreshToken: "test-refresh",
		}
		err := sessions.SaveSession(ctx, sess)
		if err != nil {
			t.Fatalf("Failed to save session: %v", err)
		}

		upd := tgbotapi.Update{
			UpdateID: 4,
			Message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: chatID},
				From: &tgbotapi.User{ID: userID},
				Text: "/stats week",
			},
		}
		upd.Message.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 6}}

		// Should not panic
		h.handleStats(ctx, upd)
	})

	t.Run("with session - invalid month format", func(t *testing.T) {
		// Create session first
		sess := &repository.UserSession{
			TelegramID:   userID,
			UserID:       "test-user",
			TenantID:     "test-tenant",
			AccessToken:  "test-token",
			RefreshToken: "test-refresh",
		}
		err := sessions.SaveSession(ctx, sess)
		if err != nil {
			t.Fatalf("Failed to save session: %v", err)
		}

		upd := tgbotapi.Update{
			UpdateID: 5,
			Message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: chatID},
				From: &tgbotapi.User{ID: userID},
				Text: "/stats invalid-format",
			},
		}
		upd.Message.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 6}}

		// Should not panic
		h.handleStats(ctx, upd)
	})
}
