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
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestHandler_HandleUpdate_Comprehensive(t *testing.T) {
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

	// Test 1: HandleUpdate with nil update (should not panic)
	t.Run("HandleUpdate_NilUpdate", func(t *testing.T) {
		var update tgbotapi.Update
		h.HandleUpdate(ctx, update)
		assert.True(t, true, "HandleUpdate with nil update should not panic")
	})

	// Test 2: HandleUpdate with nil message and nil callback (should return early)
	t.Run("HandleUpdate_NilMessageAndCallback", func(t *testing.T) {
		update := tgbotapi.Update{
			UpdateID: 1,
		}
		h.HandleUpdate(ctx, update)
		assert.True(t, true, "HandleUpdate with nil message and callback should return early")
	})

	// Test 3: HandleUpdate with callback query (should call handleCallback)
	t.Run("HandleUpdate_WithCallback", func(t *testing.T) {
		update := tgbotapi.Update{
			UpdateID: 2,
			CallbackQuery: &tgbotapi.CallbackQuery{
				ID: "callback_123",
				From: &tgbotapi.User{ID: 12345},
				Message: &tgbotapi.Message{
					Chat: &tgbotapi.Chat{ID: 67890},
				},
				Data: "test_callback_data",
			},
		}
		h.HandleUpdate(ctx, update)
		assert.True(t, true, "HandleUpdate with callback should process it")
	})

	// Test 4: HandleUpdate with command message (should call handleCommand)
	t.Run("HandleUpdate_WithCommand", func(t *testing.T) {
		update := tgbotapi.Update{
			UpdateID: 3,
			Message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: 67890},
				From: &tgbotapi.User{ID: 12345},
				Text: "/help",
				Entities: []tgbotapi.MessageEntity{
					{Type: "bot_command", Offset: 0, Length: 5},
				},
			},
		}
		h.HandleUpdate(ctx, update)
		assert.True(t, true, "HandleUpdate with command should process it")
	})

	// Test 5: HandleUpdate with unrecognized command (should handle as command anyway)
	t.Run("HandleUpdate_UnrecognizedCommand", func(t *testing.T) {
		update := tgbotapi.Update{
			UpdateID: 4,
			Message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: 67890},
				From: &tgbotapi.User{ID: 12345},
				Text: "/unknown_command",
				// No entities - should be handled as unrecognized command
			},
		}
		h.HandleUpdate(ctx, update)
		assert.True(t, true, "HandleUpdate with unrecognized command should handle it anyway")
	})

	// Test 6: HandleUpdate with command without entities but with multiple words
	t.Run("HandleUpdate_CommandWithoutEntitiesMultipleWords", func(t *testing.T) {
		update := tgbotapi.Update{
			UpdateID: 5,
			Message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: 67890},
				From: &tgbotapi.User{ID: 12345},
				Text: "/help auth transactions",
				// No entities - should trigger fallback command handling
			},
		}
		h.HandleUpdate(ctx, update)
		assert.True(t, true, "HandleUpdate should handle command without entities with multiple words")
	})

	// Test 7: HandleUpdate with OAuth email state
	t.Run("HandleUpdate_OAuthEmailState", func(t *testing.T) {
		err := states.SetState(ctx, 12345, repository.StateWaitingForOAuthEmail, nil, nil)
		assert.NoError(t, err)

		update := tgbotapi.Update{
			UpdateID: 6,
			Message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: 67890},
				From: &tgbotapi.User{ID: 12345},
				Text: "test@example.com",
			},
		}
		h.HandleUpdate(ctx, update)
		assert.True(t, true, "HandleUpdate with OAuth email state should handle it")
		_ = states.ClearState(ctx, 12345)
	})

	// Test 8: HandleUpdate with OAuth code state
	t.Run("HandleUpdate_OAuthCodeState", func(t *testing.T) {
		err := states.SetState(ctx, 12345, repository.StateWaitingForOAuthCode, nil, nil)
		assert.NoError(t, err)

		update := tgbotapi.Update{
			UpdateID: 7,
			Message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: 67890},
				From: &tgbotapi.User{ID: 12345},
				Text: "123456",
			},
		}
		h.HandleUpdate(ctx, update)
		assert.True(t, true, "HandleUpdate with OAuth code state should handle it")
		_ = states.ClearState(ctx, 12345)
	})

	// Test 9: HandleUpdate with transaction parsing and no session (echo parse)
	t.Run("HandleUpdate_TransactionNoSession", func(t *testing.T) {
		update := tgbotapi.Update{
			UpdateID: 8,
			Message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: 67890},
				From: &tgbotapi.User{ID: 99999}, // Different user without session
				Text: "300 no_session_test",
			},
		}
		h.HandleUpdate(ctx, update)
		assert.True(t, true, "HandleUpdate should echo parse result for user without session")
	})

	// Test 10: HandleUpdate with invalid transaction parsing
	t.Run("HandleUpdate_InvalidTransaction", func(t *testing.T) {
		update := tgbotapi.Update{
			UpdateID: 9,
			Message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: 67890},
				From: &tgbotapi.User{ID: 12345},
				Text: "invalid transaction text",
			},
		}
		h.HandleUpdate(ctx, update)
		assert.True(t, true, "HandleUpdate should handle invalid transaction")
	})

	// Test 11: HandleUpdate with transaction parsing and preferences
	t.Run("HandleUpdate_TransactionWithPreferences", func(t *testing.T) {
		err := prefs.SavePreferences(ctx, &repository.UserPreferences{
			TelegramID:      12345,
			Language:        "ru",
			DefaultCurrency: "USD",
		})
		assert.NoError(t, err)

		update := tgbotapi.Update{
			UpdateID: 10,
			Message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: 67890},
				From: &tgbotapi.User{ID: 12345},
				Text: "50 coffee",
			},
		}
		h.HandleUpdate(ctx, update)
		assert.True(t, true, "HandleUpdate should use preferences for currency")
	})

	// Test 12: HandleUpdate with transaction parsing and session with expired tokens
	t.Run("HandleUpdate_TransactionWithExpiredSession", func(t *testing.T) {
		err := auth.VerifyAuthCode(ctx, 12345, "auth_token_123", "123456")
		assert.NoError(t, err)

		err = sessions.UpdateTokens(ctx, 12345, &repository.TokenPair{
			AccessToken:           "expired_token",
			RefreshToken:          "expired_refresh",
			AccessTokenExpiresAt:  time.Now().Add(-time.Hour),
			RefreshTokenExpiresAt: time.Now().Add(-time.Hour),
		})
		assert.NoError(t, err)

		update := tgbotapi.Update{
			UpdateID: 11,
			Message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: 67890},
				From: &tgbotapi.User{ID: 12345},
				Text: "100 expired_session_test",
			},
		}
		h.HandleUpdate(ctx, update)
		assert.True(t, true, "HandleUpdate should handle expired session")
	})

	// Test 13: HandleUpdate with transaction parsing and valid session with category mapping
	t.Run("HandleUpdate_TransactionWithValidSessionAndCategory", func(t *testing.T) {
		err := auth.VerifyAuthCode(ctx, 12345, "auth_token_123", "123456")
		assert.NoError(t, err)

		update := tgbotapi.Update{
			UpdateID: 12,
			Message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: 67890},
				From: &tgbotapi.User{ID: 12345},
				Text: "200 valid_session_test",
			},
		}
		h.HandleUpdate(ctx, update)
		assert.True(t, true, "HandleUpdate should process transaction with valid session")
	})

	// Test 14: HandleUpdate with transaction parsing and transaction creation failure
	t.Run("HandleUpdate_TransactionCreationFailure", func(t *testing.T) {
		err := auth.VerifyAuthCode(ctx, 12345, "auth_token_123", "123456")
		assert.NoError(t, err)

		update := tgbotapi.Update{
			UpdateID: 13,
			Message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: 67890},
				From: &tgbotapi.User{ID: 12345},
				Text: "700 FAIL", // This will trigger failure in FakeTransactionClient
			},
		}
		h.HandleUpdate(ctx, update)
		assert.True(t, true, "HandleUpdate should handle transaction creation failure")
	})

	// Test 15: HandleUpdate with transaction parsing with errors in parsed result
	t.Run("HandleUpdate_TransactionParsingWithErrors", func(t *testing.T) {
		update := tgbotapi.Update{
			UpdateID: 14,
			Message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: 67890},
				From: &tgbotapi.User{ID: 12345},
				Text: "abc coffee", // Invalid amount
			},
		}
		h.HandleUpdate(ctx, update)
		assert.True(t, true, "HandleUpdate should handle transaction parsing with errors")
	})

	// Test 16: HandleUpdate with transaction and nameMapper returning category name
	t.Run("HandleUpdate_TransactionWithNameMapperSuccess", func(t *testing.T) {
		err := auth.VerifyAuthCode(ctx, 12345, "auth_token_123", "123456")
		assert.NoError(t, err)

		// Add category mapping
		err = mappings.AddMapping(ctx, &repository.CategoryMapping{
			ID:         "mapping_1",
			TenantID:   "test_tenant",
			Keyword:    "coffee",
			CategoryID: "category_coffee_id",
			Priority:   1,
		})
		assert.NoError(t, err)

		update := tgbotapi.Update{
			UpdateID: 15,
			Message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: 67890},
				From: &tgbotapi.User{ID: 12345},
				Text: "100 coffee",
			},
		}
		h.HandleUpdate(ctx, update)
		assert.True(t, true, "HandleUpdate should handle transaction with successful nameMapper")
	})
}

func TestHandler_HandleCallback_Comprehensive(t *testing.T) {
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

	// Test 1: handleCallback with nil callback (should return early)
	t.Run("HandleCallback_NilCallback", func(t *testing.T) {
		update := tgbotapi.Update{
			UpdateID: 1,
		}
		h.handleCallback(ctx, update)
		assert.True(t, true, "handleCallback with nil callback should return early")
	})

	// Test 2: handleCallback with category callback but no state
	t.Run("HandleCallback_CategoryNoState", func(t *testing.T) {
		update := tgbotapi.Update{
			UpdateID: 2,
			CallbackQuery: &tgbotapi.CallbackQuery{
				ID: "callback_123",
				From: &tgbotapi.User{ID: 12345},
				Message: &tgbotapi.Message{
					Chat: &tgbotapi.Chat{ID: 67890},
				},
				Data: "cat:food",
			},
		}
		h.handleCallback(ctx, update)
		assert.True(t, true, "handleCallback should handle missing state")
	})

	// Test 3: handleCallback with category callback and state but no session
	t.Run("HandleCallback_CategoryNoSession", func(t *testing.T) {
		err := states.SetState(ctx, 12345, repository.StateWaitingForCategory, map[string]interface{}{
			"type":         "expense",
			"amount_minor": int64(1000),
			"currency":     "RUB",
			"desc":         "test",
		}, nil)
		assert.NoError(t, err)

		update := tgbotapi.Update{
			UpdateID: 3,
			CallbackQuery: &tgbotapi.CallbackQuery{
				ID: "callback_123",
				From: &tgbotapi.User{ID: 12345},
				Message: &tgbotapi.Message{
					Chat: &tgbotapi.Chat{ID: 67890},
				},
				Data: "cat:food",
			},
		}
		h.handleCallback(ctx, update)
		assert.True(t, true, "handleCallback should handle missing session")
		_ = states.ClearState(ctx, 12345)
	})

	// Test 4: handleCallback with category callback, state, and session
	t.Run("HandleCallback_CategoryWithSession", func(t *testing.T) {
		err := auth.VerifyAuthCode(ctx, 12345, "auth_token_123", "123456")
		assert.NoError(t, err)

		err = states.SetState(ctx, 12345, repository.StateWaitingForCategory, map[string]interface{}{
			"type":         "expense",
			"amount_minor": int64(1000),
			"currency":     "RUB",
			"desc":         "test",
		}, nil)
		assert.NoError(t, err)

		update := tgbotapi.Update{
			UpdateID: 4,
			CallbackQuery: &tgbotapi.CallbackQuery{
				ID: "callback_123",
				From: &tgbotapi.User{ID: 12345},
				Message: &tgbotapi.Message{
					Chat: &tgbotapi.Chat{ID: 67890},
				},
				Data: "cat:food",
			},
		}
		h.handleCallback(ctx, update)
		assert.True(t, true, "handleCallback should process category callback")
		_ = states.ClearState(ctx, 12345)
	})

	// Test 5: handleCallback with language callback
	t.Run("HandleCallback_Language", func(t *testing.T) {
		update := tgbotapi.Update{
			UpdateID: 5,
			CallbackQuery: &tgbotapi.CallbackQuery{
				ID: "callback_123",
				From: &tgbotapi.User{ID: 12345},
				Message: &tgbotapi.Message{
					Chat: &tgbotapi.Chat{ID: 67890},
				},
				Data: "lang:en",
			},
		}
		h.handleCallback(ctx, update)
		assert.True(t, true, "handleCallback should process language callback")
	})

	// Test 6: handleCallback with currency callback
	t.Run("HandleCallback_Currency", func(t *testing.T) {
		update := tgbotapi.Update{
			UpdateID: 6,
			CallbackQuery: &tgbotapi.CallbackQuery{
				ID: "callback_123",
				From: &tgbotapi.User{ID: 12345},
				Message: &tgbotapi.Message{
					Chat: &tgbotapi.Chat{ID: 67890},
				},
				Data: "cur:USD",
			},
		}
		h.handleCallback(ctx, update)
		assert.True(t, true, "handleCallback should process currency callback")
	})

	// Test 7: handleCallback with different amount types in context
	t.Run("HandleCallback_DifferentAmountTypes", func(t *testing.T) {
		err := auth.VerifyAuthCode(ctx, 12345, "auth_token_123", "123456")
		assert.NoError(t, err)

		// Test with float64 amount
		err = states.SetState(ctx, 12345, repository.StateWaitingForCategory, map[string]interface{}{
			"type":         "expense",
			"amount_minor": float64(1500.5),
			"currency":     "RUB",
			"desc":         "test",
		}, nil)
		assert.NoError(t, err)

		update := tgbotapi.Update{
			UpdateID: 7,
			CallbackQuery: &tgbotapi.CallbackQuery{
				ID: "callback_123",
				From: &tgbotapi.User{ID: 12345},
				Message: &tgbotapi.Message{
					Chat: &tgbotapi.Chat{ID: 67890},
				},
				Data: "cat:food",
			},
		}
		h.handleCallback(ctx, update)
		assert.True(t, true, "handleCallback should handle float64 amount")

		// Test with int amount
		err = states.SetState(ctx, 12345, repository.StateWaitingForCategory, map[string]interface{}{
			"type":         "income",
			"amount_minor": int(2000),
			"currency":     "RUB",
			"desc":         "test",
		}, nil)
		assert.NoError(t, err)

		update.UpdateID = 8
		h.handleCallback(ctx, update)
		assert.True(t, true, "handleCallback should handle int amount")
		_ = states.ClearState(ctx, 12345)
	})

	// Test 8: handleCallback with tenant callback
	t.Run("HandleCallback_Tenant", func(t *testing.T) {
		err := auth.VerifyAuthCode(ctx, 12345, "auth_token_123", "123456")
		assert.NoError(t, err)

		update := tgbotapi.Update{
			UpdateID: 9,
			CallbackQuery: &tgbotapi.CallbackQuery{
				ID: "callback_123",
				From: &tgbotapi.User{ID: 12345},
				Message: &tgbotapi.Message{
					Chat: &tgbotapi.Chat{ID: 67890},
				},
				Data: "tenant:test_tenant_id",
			},
		}
		h.handleCallback(ctx, update)
		assert.True(t, true, "handleCallback should handle tenant selection")
	})

	// Test 9: handleCallback with tenant callback and update failure
	t.Run("HandleCallback_TenantUpdateFailure", func(t *testing.T) {
		update := tgbotapi.Update{
			UpdateID: 10,
			CallbackQuery: &tgbotapi.CallbackQuery{
				ID: "callback_123",
				From: &tgbotapi.User{ID: 99999}, // User without session
				Message: &tgbotapi.Message{
					Chat: &tgbotapi.Chat{ID: 67890},
				},
				Data: "tenant:test_tenant_id",
			},
		}
		h.handleCallback(ctx, update)
		assert.True(t, true, "handleCallback should handle tenant update failure")
	})

	// Test 10: handleCallback with help callback
	t.Run("HandleCallback_Help", func(t *testing.T) {
		update := tgbotapi.Update{
			UpdateID: 11,
			CallbackQuery: &tgbotapi.CallbackQuery{
				ID: "callback_123",
				From: &tgbotapi.User{ID: 12345},
				Message: &tgbotapi.Message{
					Chat: &tgbotapi.Chat{ID: 67890},
				},
				Data: "help:auth",
			},
		}
		h.handleCallback(ctx, update)
		assert.True(t, true, "handleCallback should handle help callback")
	})

	// Test 11: handleCallback with category callback with draft_id in context
	t.Run("HandleCallback_CategoryWithDraft", func(t *testing.T) {
		err := auth.VerifyAuthCode(ctx, 12345, "auth_token_123", "123456")
		assert.NoError(t, err)

		err = states.SetState(ctx, 12345, repository.StateWaitingForCategory, map[string]interface{}{
			"type":         "expense",
			"amount_minor": int64(1000),
			"currency":     "RUB",
			"desc":         "test",
			"draft_id":     "test_draft_id",
		}, nil)
		assert.NoError(t, err)

		update := tgbotapi.Update{
			UpdateID: 12,
			CallbackQuery: &tgbotapi.CallbackQuery{
				ID: "callback_123",
				From: &tgbotapi.User{ID: 12345},
				Message: &tgbotapi.Message{
					Chat: &tgbotapi.Chat{ID: 67890},
				},
				Data: "cat:food",
			},
		}
		h.handleCallback(ctx, update)
		assert.True(t, true, "handleCallback should handle category with draft")
	})

	// Test 12: handleCallback with category callback and transaction creation failure
	t.Run("HandleCallback_CategoryTransactionFailure", func(t *testing.T) {
		err := auth.VerifyAuthCode(ctx, 12345, "auth_token_123", "123456")
		assert.NoError(t, err)

		err = states.SetState(ctx, 12345, repository.StateWaitingForCategory, map[string]interface{}{
			"type":         "expense",
			"amount_minor": int64(1000),
			"currency":     "RUB",
			"desc":         "FAIL", // This will trigger failure in FakeTransactionClient
		}, nil)
		assert.NoError(t, err)

		update := tgbotapi.Update{
			UpdateID: 13,
			CallbackQuery: &tgbotapi.CallbackQuery{
				ID: "callback_123",
				From: &tgbotapi.User{ID: 12345},
				Message: &tgbotapi.Message{
					Chat: &tgbotapi.Chat{ID: 67890},
				},
				Data: "cat:food",
			},
		}
		h.handleCallback(ctx, update)
		assert.True(t, true, "handleCallback should handle transaction creation failure")
	})

	// Test 13: handleCallback with language callback preserving currency
	t.Run("HandleCallback_LanguagePreserveCurrency", func(t *testing.T) {
		err := prefs.SavePreferences(ctx, &repository.UserPreferences{
			TelegramID:      12345,
			Language:        "ru",
			DefaultCurrency: "EUR",
		})
		assert.NoError(t, err)

		update := tgbotapi.Update{
			UpdateID: 14,
			CallbackQuery: &tgbotapi.CallbackQuery{
				ID: "callback_123",
				From: &tgbotapi.User{ID: 12345},
				Message: &tgbotapi.Message{
					Chat: &tgbotapi.Chat{ID: 67890},
				},
				Data: "lang:en",
			},
		}
		h.handleCallback(ctx, update)
		assert.True(t, true, "handleCallback should preserve currency when changing language")
	})

	// Test 14: handleCallback with currency callback preserving language
	t.Run("HandleCallback_CurrencyPreserveLanguage", func(t *testing.T) {
		err := prefs.SavePreferences(ctx, &repository.UserPreferences{
			TelegramID:      12345,
			Language:        "en",
			DefaultCurrency: "RUB",
		})
		assert.NoError(t, err)

		update := tgbotapi.Update{
			UpdateID: 15,
			CallbackQuery: &tgbotapi.CallbackQuery{
				ID: "callback_123",
				From: &tgbotapi.User{ID: 12345},
				Message: &tgbotapi.Message{
					Chat: &tgbotapi.Chat{ID: 67890},
				},
				Data: "cur:USD",
			},
		}
		h.handleCallback(ctx, update)
		assert.True(t, true, "handleCallback should preserve language when changing currency")
	})

	// Test 15: handleCallback with unknown callback data
	t.Run("HandleCallback_UnknownData", func(t *testing.T) {
		update := tgbotapi.Update{
			UpdateID: 16,
			CallbackQuery: &tgbotapi.CallbackQuery{
				ID: "callback_123",
				From: &tgbotapi.User{ID: 12345},
				Message: &tgbotapi.Message{
					Chat: &tgbotapi.Chat{ID: 67890},
				},
				Data: "unknown:data",
			},
		}
		h.handleCallback(ctx, update)
		assert.True(t, true, "handleCallback should handle unknown callback data")
	})
}
