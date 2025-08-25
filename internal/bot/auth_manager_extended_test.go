package bot

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"budget-bot/internal/repository"
	"go.uber.org/zap"
)

func TestAuthManager_GetSession_EmptySession(t *testing.T) {
	// Test GetSession with empty session
	mockAuthClient := &MockAuthClient{}
	mockSessionRepo := &MockSessionRepository{}
	
	am := &AuthManager{
		authClient: mockAuthClient,
		sessionRepo: mockSessionRepo,
		logger: zap.NewNop(),
	}
	
	// Test with empty session - mock returns error
	mockSessionRepo.On("GetSession", context.Background(), int64(123)).Return(nil, assert.AnError)
	
	session, err := am.GetSession(context.Background(), 123)
	assert.Nil(t, session)
	assert.Error(t, err)
}

func TestAuthManager_GetSession_ValidSession(t *testing.T) {
	// Test GetSession with valid session
	mockAuthClient := &MockAuthClient{}
	mockSessionRepo := &MockSessionRepository{}
	
	am := &AuthManager{
		authClient: mockAuthClient,
		sessionRepo: mockSessionRepo,
		logger: zap.NewNop(),
	}
	
	// Create a valid session
	validSession := &repository.UserSession{
		TelegramID: 123,
		UserID:     "user123",
		TenantID:   "tenant123",
		AccessToken: "access_token",
		RefreshToken: "refresh_token",
		AccessTokenExpiresAt: time.Now().Add(time.Hour), // Valid for 1 hour
		RefreshTokenExpiresAt: time.Now().Add(24 * time.Hour),
	}
	
	// Mock the session repository to return valid session
	mockSessionRepo.On("GetSession", context.Background(), int64(123)).Return(validSession, nil)
	
	session, err := am.GetSession(context.Background(), 123)
	assert.NotNil(t, session)
	assert.NoError(t, err)
	assert.Equal(t, int64(123), session.TelegramID)
	assert.Equal(t, "user123", session.UserID)
	assert.Equal(t, "tenant123", session.TenantID)
}

func TestAuthManager_Struct(t *testing.T) {
	// Test creating an AuthManager struct
	mockAuthClient := &MockAuthClient{}
	mockSessionRepo := &MockSessionRepository{}
	logger := zap.NewNop()
	
	am := &AuthManager{
		authClient: mockAuthClient,
		sessionRepo: mockSessionRepo,
		logger: logger,
	}
	
	assert.NotNil(t, am)
	assert.Equal(t, mockAuthClient, am.authClient)
	assert.Equal(t, mockSessionRepo, am.sessionRepo)
	assert.Equal(t, logger, am.logger)
}

func TestAuthManager_NewAuthManager(t *testing.T) {
	// Test NewAuthManager function
	mockAuthClient := &MockAuthClient{}
	mockSessionRepo := &MockSessionRepository{}
	logger := zap.NewNop()
	
	am := NewAuthManager(mockAuthClient, mockSessionRepo, logger)
	assert.NotNil(t, am)
	assert.Equal(t, mockAuthClient, am.authClient)
	assert.Equal(t, mockSessionRepo, am.sessionRepo)
	assert.Equal(t, logger, am.logger)
}
