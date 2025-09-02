package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserSession_NewUserSession(t *testing.T) {
	// Test creating a new UserSession
	session := &UserSession{
		TelegramID: 123456,
		UserID:     "user123",
		TenantID:   "tenant123",
	}
	
	assert.Equal(t, int64(123456), session.TelegramID)
	assert.Equal(t, "user123", session.UserID)
	assert.Equal(t, "tenant123", session.TenantID)
}

func TestUserSession_EmptySession(t *testing.T) {
	// Test creating an empty UserSession
	session := &UserSession{}
	
	assert.Equal(t, int64(0), session.TelegramID)
	assert.Equal(t, "", session.UserID)
	assert.Equal(t, "", session.TenantID)
}

func TestTokenPair_NewTokenPair(t *testing.T) {
	// Test creating a new TokenPair
	tokens := &TokenPair{
		AccessToken:  "access123",
		RefreshToken: "refresh123",
	}
	
	assert.Equal(t, "access123", tokens.AccessToken)
	assert.Equal(t, "refresh123", tokens.RefreshToken)
}

func TestTokenPair_EmptyTokens(t *testing.T) {
	// Test creating empty TokenPair
	tokens := &TokenPair{}
	
	assert.Equal(t, "", tokens.AccessToken)
	assert.Equal(t, "", tokens.RefreshToken)
}

func TestCategoryMapping_NewCategoryMapping(t *testing.T) {
	// Test creating a new CategoryMapping
	mapping := &CategoryMapping{
		TenantID:   "tenant123",
		Keyword:    "groceries",
		CategoryID: "cat123",
	}
	
	assert.Equal(t, "tenant123", mapping.TenantID)
	assert.Equal(t, "groceries", mapping.Keyword)
	assert.Equal(t, "cat123", mapping.CategoryID)
}

func TestCategoryMapping_EmptyMapping(t *testing.T) {
	// Test creating empty CategoryMapping
	mapping := &CategoryMapping{}
	
	assert.Equal(t, "", mapping.TenantID)
	assert.Equal(t, "", mapping.Keyword)
	assert.Equal(t, "", mapping.CategoryID)
}

func TestDialogStateRecord_NewDialogStateRecord(t *testing.T) {
	// Test creating a new DialogStateRecord
	state := &DialogStateRecord{
		TelegramID: 123456,
		State:      StateWaitingForEmail,
	}
	
	assert.Equal(t, int64(123456), state.TelegramID)
	assert.Equal(t, StateWaitingForEmail, state.State)
}

func TestDialogStateRecord_EmptyState(t *testing.T) {
	// Test creating empty DialogStateRecord
	state := &DialogStateRecord{}
	
	assert.Equal(t, int64(0), state.TelegramID)
	assert.Equal(t, DialogState(""), state.State)
}

func TestDialogState_Constants(t *testing.T) {
	// Test dialog state constants
	assert.Equal(t, DialogState("idle"), StateIdle)
	assert.Equal(t, DialogState("waiting_for_email"), StateWaitingForEmail)
	assert.Equal(t, DialogState("waiting_for_password"), StateWaitingForPassword)
	assert.Equal(t, DialogState("waiting_for_register_email"), StateWaitingForRegisterEmail)
	assert.Equal(t, DialogState("waiting_for_register_password"), StateWaitingForRegisterPassword)
	assert.Equal(t, DialogState("waiting_for_register_name"), StateWaitingForRegisterName)

	assert.Equal(t, DialogState("waiting_for_category"), StateWaitingForCategory)
	assert.Equal(t, DialogState("waiting_for_oauth_email"), StateWaitingForOAuthEmail)
	assert.Equal(t, DialogState("waiting_for_oauth_code"), StateWaitingForOAuthCode)
}

func TestTransactionDraft_NewTransactionDraft(t *testing.T) {
	// Test creating a new TransactionDraft
	draft := &TransactionDraft{
		TelegramID:  123456,
		AmountMinor: 10000,
		Currency:    "USD",
		Description: "Test transaction",
	}
	
	assert.Equal(t, int64(123456), draft.TelegramID)
	assert.Equal(t, int64(10000), draft.AmountMinor)
	assert.Equal(t, "USD", draft.Currency)
	assert.Equal(t, "Test transaction", draft.Description)
}

func TestTransactionDraft_EmptyDraft(t *testing.T) {
	// Test creating empty TransactionDraft
	draft := &TransactionDraft{}
	
	assert.Equal(t, int64(0), draft.TelegramID)
	assert.Equal(t, int64(0), draft.AmountMinor)
	assert.Equal(t, "", draft.Currency)
	assert.Equal(t, "", draft.Description)
}

func TestUserPreferences_NewUserPreferences(t *testing.T) {
	// Test creating new UserPreferences
	prefs := &UserPreferences{
		TelegramID:      123456,
		Language:        "en",
		DefaultCurrency: "USD",
	}
	
	assert.Equal(t, int64(123456), prefs.TelegramID)
	assert.Equal(t, "en", prefs.Language)
	assert.Equal(t, "USD", prefs.DefaultCurrency)
}

func TestUserPreferences_EmptyPreferences(t *testing.T) {
	// Test creating empty UserPreferences
	prefs := &UserPreferences{}
	
	assert.Equal(t, int64(0), prefs.TelegramID)
	assert.Equal(t, "", prefs.Language)
	assert.Equal(t, "", prefs.DefaultCurrency)
}
