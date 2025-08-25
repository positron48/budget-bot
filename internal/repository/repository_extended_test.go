package repository

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUserSession_Extended(t *testing.T) {
	// Test UserSession with different values
	now := time.Now()
	session1 := &UserSession{
		TelegramID:              123456,
		UserID:                  "user123",
		TenantID:                "tenant123",
		AccessToken:             "access_token_123",
		RefreshToken:            "refresh_token_123",
		AccessTokenExpiresAt:    now.Add(time.Hour),
		RefreshTokenExpiresAt:   now.Add(24 * time.Hour),
		CreatedAt:               now,
		UpdatedAt:               now,
	}
	
	session2 := &UserSession{
		TelegramID:              789012,
		UserID:                  "user456",
		TenantID:                "tenant456",
		AccessToken:             "access_token_456",
		RefreshToken:            "refresh_token_456",
		AccessTokenExpiresAt:    now.Add(2 * time.Hour),
		RefreshTokenExpiresAt:   now.Add(48 * time.Hour),
		CreatedAt:               now,
		UpdatedAt:               now,
	}
	
	assert.Equal(t, int64(123456), session1.TelegramID)
	assert.Equal(t, "user123", session1.UserID)
	assert.Equal(t, "tenant123", session1.TenantID)
	assert.Equal(t, "access_token_123", session1.AccessToken)
	assert.Equal(t, "refresh_token_123", session1.RefreshToken)
	
	assert.Equal(t, int64(789012), session2.TelegramID)
	assert.Equal(t, "user456", session2.UserID)
	assert.Equal(t, "tenant456", session2.TenantID)
	assert.Equal(t, "access_token_456", session2.AccessToken)
	assert.Equal(t, "refresh_token_456", session2.RefreshToken)
}

func TestTokenPair_Extended(t *testing.T) {
	// Test TokenPair with different values
	now := time.Now()
	tokens1 := &TokenPair{
		AccessToken:           "access_token_123",
		RefreshToken:          "refresh_token_123",
		AccessTokenExpiresAt:  now.Add(time.Hour),
		RefreshTokenExpiresAt: now.Add(24 * time.Hour),
	}
	
	tokens2 := &TokenPair{
		AccessToken:           "access_token_456",
		RefreshToken:          "refresh_token_456",
		AccessTokenExpiresAt:  now.Add(2 * time.Hour),
		RefreshTokenExpiresAt: now.Add(48 * time.Hour),
	}
	
	assert.Equal(t, "access_token_123", tokens1.AccessToken)
	assert.Equal(t, "refresh_token_123", tokens1.RefreshToken)
	
	assert.Equal(t, "access_token_456", tokens2.AccessToken)
	assert.Equal(t, "refresh_token_456", tokens2.RefreshToken)
}

func TestCategoryMapping_Extended(t *testing.T) {
	// Test CategoryMapping with different values
	mapping1 := &CategoryMapping{
		TenantID:   "tenant123",
		Keyword:    "groceries",
		CategoryID: "cat123",
	}
	
	mapping2 := &CategoryMapping{
		TenantID:   "tenant456",
		Keyword:    "transport",
		CategoryID: "cat456",
	}
	
	assert.Equal(t, "tenant123", mapping1.TenantID)
	assert.Equal(t, "groceries", mapping1.Keyword)
	assert.Equal(t, "cat123", mapping1.CategoryID)
	
	assert.Equal(t, "tenant456", mapping2.TenantID)
	assert.Equal(t, "transport", mapping2.Keyword)
	assert.Equal(t, "cat456", mapping2.CategoryID)
}

func TestDialogStateRecord_Extended(t *testing.T) {
	// Test DialogStateRecord with different states
	state1 := &DialogStateRecord{
		TelegramID: 123456,
		State:      StateWaitingForEmail,
	}
	
	state2 := &DialogStateRecord{
		TelegramID: 789012,
		State:      StateWaitingForPassword,
	}
	
	state3 := &DialogStateRecord{
		TelegramID: 345678,
		State:      StateConfirmingTransaction,
	}
	
	assert.Equal(t, int64(123456), state1.TelegramID)
	assert.Equal(t, StateWaitingForEmail, state1.State)
	
	assert.Equal(t, int64(789012), state2.TelegramID)
	assert.Equal(t, StateWaitingForPassword, state2.State)
	
	assert.Equal(t, int64(345678), state3.TelegramID)
	assert.Equal(t, StateConfirmingTransaction, state3.State)
}

func TestTransactionDraft_Extended(t *testing.T) {
	// Test TransactionDraft with different values
	draft1 := &TransactionDraft{
		TelegramID:  123456,
		AmountMinor: 10000,
		Currency:    "USD",
		Description: "Groceries",
		Type:        "expense",
	}
	
	draft2 := &TransactionDraft{
		TelegramID:  789012,
		AmountMinor: 5000,
		Currency:    "EUR",
		Description: "Salary",
		Type:        "income",
	}
	
	assert.Equal(t, int64(123456), draft1.TelegramID)
	assert.Equal(t, int64(10000), draft1.AmountMinor)
	assert.Equal(t, "USD", draft1.Currency)
	assert.Equal(t, "Groceries", draft1.Description)
	assert.Equal(t, "expense", draft1.Type)
	
	assert.Equal(t, int64(789012), draft2.TelegramID)
	assert.Equal(t, int64(5000), draft2.AmountMinor)
	assert.Equal(t, "EUR", draft2.Currency)
	assert.Equal(t, "Salary", draft2.Description)
	assert.Equal(t, "income", draft2.Type)
}

func TestUserPreferences_Extended(t *testing.T) {
	// Test UserPreferences with different values
	prefs1 := &UserPreferences{
		TelegramID:      123456,
		Language:        "en",
		DefaultCurrency: "USD",
	}
	
	prefs2 := &UserPreferences{
		TelegramID:      789012,
		Language:        "ru",
		DefaultCurrency: "RUB",
	}
	
	assert.Equal(t, int64(123456), prefs1.TelegramID)
	assert.Equal(t, "en", prefs1.Language)
	assert.Equal(t, "USD", prefs1.DefaultCurrency)
	
	assert.Equal(t, int64(789012), prefs2.TelegramID)
	assert.Equal(t, "ru", prefs2.Language)
	assert.Equal(t, "RUB", prefs2.DefaultCurrency)
}

func TestUserSession_WithNegativeTelegramID(t *testing.T) {
	// Test UserSession with negative telegram ID
	now := time.Now()
	session := &UserSession{
		TelegramID:              -123456,
		UserID:                  "user123",
		TenantID:                "tenant123",
		AccessToken:             "access_token_123",
		RefreshToken:            "refresh_token_123",
		AccessTokenExpiresAt:    now.Add(time.Hour),
		RefreshTokenExpiresAt:   now.Add(24 * time.Hour),
		CreatedAt:               now,
		UpdatedAt:               now,
	}
	
	assert.Equal(t, int64(-123456), session.TelegramID)
	assert.Equal(t, "user123", session.UserID)
	assert.Equal(t, "tenant123", session.TenantID)
}

func TestUserSession_WithEmptyStrings(t *testing.T) {
	// Test UserSession with empty strings
	now := time.Now()
	session := &UserSession{
		TelegramID:              123456,
		UserID:                  "",
		TenantID:                "",
		AccessToken:             "",
		RefreshToken:            "",
		AccessTokenExpiresAt:    now.Add(time.Hour),
		RefreshTokenExpiresAt:   now.Add(24 * time.Hour),
		CreatedAt:               now,
		UpdatedAt:               now,
	}
	
	assert.Equal(t, int64(123456), session.TelegramID)
	assert.Equal(t, "", session.UserID)
	assert.Equal(t, "", session.TenantID)
	assert.Equal(t, "", session.AccessToken)
	assert.Equal(t, "", session.RefreshToken)
}

func TestTokenPair_WithExpiredTokens(t *testing.T) {
	// Test TokenPair with expired tokens
	past := time.Now().Add(-time.Hour)
	tokens := &TokenPair{
		AccessToken:           "expired_access_token",
		RefreshToken:          "expired_refresh_token",
		AccessTokenExpiresAt:  past,
		RefreshTokenExpiresAt: past,
	}
	
	assert.Equal(t, "expired_access_token", tokens.AccessToken)
	assert.Equal(t, "expired_refresh_token", tokens.RefreshToken)
	assert.True(t, tokens.AccessTokenExpiresAt.Before(time.Now()))
	assert.True(t, tokens.RefreshTokenExpiresAt.Before(time.Now()))
}

func TestCategoryMapping_WithSpecialCharacters(t *testing.T) {
	// Test CategoryMapping with special characters
	mapping := &CategoryMapping{
		TenantID:   "tenant_with_special_chars_123",
		Keyword:    "groceries & food",
		CategoryID: "cat_with_special_chars_456",
	}
	
	assert.Equal(t, "tenant_with_special_chars_123", mapping.TenantID)
	assert.Equal(t, "groceries & food", mapping.Keyword)
	assert.Equal(t, "cat_with_special_chars_456", mapping.CategoryID)
}

func TestDialogStateRecord_WithAllStates(t *testing.T) {
	// Test DialogStateRecord with all possible states
	states := []DialogState{
		StateWaitingForEmail,
		StateWaitingForPassword,
		StateConfirmingTransaction,
	}
	
	for i, state := range states {
		record := &DialogStateRecord{
			TelegramID: int64(100000 + i),
			State:      state,
		}
		
		assert.Equal(t, int64(100000+i), record.TelegramID)
		assert.Equal(t, state, record.State)
	}
}

func TestTransactionDraft_WithLargeAmount(t *testing.T) {
	// Test TransactionDraft with large amount
	largeAmount := int64(999999999)
	draft := &TransactionDraft{
		TelegramID:  123456,
		AmountMinor: largeAmount,
		Currency:    "USD",
		Description: "Large transaction",
		Type:        "expense",
	}
	
	assert.Equal(t, int64(123456), draft.TelegramID)
	assert.Equal(t, largeAmount, draft.AmountMinor)
	assert.Equal(t, "USD", draft.Currency)
	assert.Equal(t, "Large transaction", draft.Description)
	assert.Equal(t, "expense", draft.Type)
}

func TestTransactionDraft_WithNegativeAmount(t *testing.T) {
	// Test TransactionDraft with negative amount
	negativeAmount := int64(-1000)
	draft := &TransactionDraft{
		TelegramID:  123456,
		AmountMinor: negativeAmount,
		Currency:    "USD",
		Description: "Negative transaction",
		Type:        "expense",
	}
	
	assert.Equal(t, int64(123456), draft.TelegramID)
	assert.Equal(t, negativeAmount, draft.AmountMinor)
	assert.Equal(t, "USD", draft.Currency)
	assert.Equal(t, "Negative transaction", draft.Description)
	assert.Equal(t, "expense", draft.Type)
}

func TestUserPreferences_WithDifferentLanguages(t *testing.T) {
	// Test UserPreferences with different languages
	languages := []string{"en", "ru", "es", "fr", "de"}
	currencies := []string{"USD", "RUB", "EUR", "GBP", "JPY"}
	
	for i, lang := range languages {
		prefs := &UserPreferences{
			TelegramID:      int64(100000 + i),
			Language:        lang,
			DefaultCurrency: currencies[i],
		}
		
		assert.Equal(t, int64(100000+i), prefs.TelegramID)
		assert.Equal(t, lang, prefs.Language)
		assert.Equal(t, currencies[i], prefs.DefaultCurrency)
	}
}
