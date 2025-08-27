package bot

import (
	"context"
	"errors"
	"testing"
	"time"

	"budget-bot/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockAuthClient is a mock implementation of AuthClient
type MockAuthClient struct {
	mock.Mock
}

func (m *MockAuthClient) Register(ctx context.Context, email, password, name string) (userID string, tenantID string, accessToken string, refreshToken string, accessExp time.Time, refreshExp time.Time, err error) {
	args := m.Called(ctx, email, password, name)
	return args.String(0), args.String(1), args.String(2), args.String(3), args.Get(4).(time.Time), args.Get(5).(time.Time), args.Error(6)
}

func (m *MockAuthClient) Login(ctx context.Context, email, password string) (userID string, tenantID string, accessToken string, refreshToken string, accessExp time.Time, refreshExp time.Time, err error) {
	args := m.Called(ctx, email, password)
	return args.String(0), args.String(1), args.String(2), args.String(3), args.Get(4).(time.Time), args.Get(5).(time.Time), args.Error(6)
}

func (m *MockAuthClient) RefreshToken(ctx context.Context, refreshToken string) (accessToken string, refreshTokenNew string, accessExp time.Time, refreshExp time.Time, err error) {
	args := m.Called(ctx, refreshToken)
	return args.String(0), args.String(1), args.Get(2).(time.Time), args.Get(3).(time.Time), args.Error(4)
}

// MockSessionRepository is a mock implementation of SessionRepository
type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) SaveSession(ctx context.Context, session *repository.UserSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionRepository) GetSession(ctx context.Context, telegramID int64) (*repository.UserSession, error) {
	args := m.Called(ctx, telegramID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.UserSession), args.Error(1)
}

func (m *MockSessionRepository) DeleteSession(ctx context.Context, telegramID int64) error {
	args := m.Called(ctx, telegramID)
	return args.Error(0)
}

func (m *MockSessionRepository) UpdateTokens(ctx context.Context, telegramID int64, tokens *repository.TokenPair) error {
	args := m.Called(ctx, telegramID, tokens)
	return args.Error(0)
}

func (m *MockSessionRepository) UpdateTenantID(ctx context.Context, telegramID int64, tenantID string) error {
	args := m.Called(ctx, telegramID, tenantID)
	return args.Error(0)
}

func TestNewAuthManager(t *testing.T) {
	logger := zap.NewNop()
	authClient := &MockAuthClient{}
	sessionRepo := &MockSessionRepository{}

	am := NewAuthManager(authClient, sessionRepo, logger)

	assert.NotNil(t, am)
	assert.Equal(t, authClient, am.authClient)
	assert.Equal(t, sessionRepo, am.sessionRepo)
	assert.Equal(t, logger, am.logger)
}

func TestAuthManager_Register_Success(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	authClient := &MockAuthClient{}
	sessionRepo := &MockSessionRepository{}

	am := NewAuthManager(authClient, sessionRepo, logger)

	telegramID := int64(123)
	email := "test@example.com"
	password := "password"
	name := "Test User"

	expectedUserID := "user123"
	expectedTenantID := "tenant123"
	expectedAccessToken := "access123"
	expectedRefreshToken := "refresh123"
	expectedAccessExp := time.Now().Add(time.Hour)
	expectedRefreshExp := time.Now().Add(24 * time.Hour)

	authClient.On("Register", ctx, email, password, name).Return(
		expectedUserID, expectedTenantID, expectedAccessToken, expectedRefreshToken,
		expectedAccessExp, expectedRefreshExp, nil)

	sessionRepo.On("SaveSession", ctx, mock.MatchedBy(func(session *repository.UserSession) bool {
		return session.TelegramID == telegramID &&
			session.UserID == expectedUserID &&
			session.TenantID == expectedTenantID &&
			session.AccessToken == expectedAccessToken &&
			session.RefreshToken == expectedRefreshToken
	})).Return(nil)

	err := am.Register(ctx, telegramID, email, password, name)

	assert.NoError(t, err)
	authClient.AssertExpectations(t)
	sessionRepo.AssertExpectations(t)
}

func TestAuthManager_Register_AuthError(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	authClient := &MockAuthClient{}
	sessionRepo := &MockSessionRepository{}

	am := NewAuthManager(authClient, sessionRepo, logger)

	telegramID := int64(123)
	email := "test@example.com"
	password := "password"
	name := "Test User"

	expectedError := errors.New("auth failed")

	authClient.On("Register", ctx, email, password, name).Return(
		"", "", "", "", time.Time{}, time.Time{}, expectedError)

	err := am.Register(ctx, telegramID, email, password, name)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	authClient.AssertExpectations(t)
	sessionRepo.AssertNotCalled(t, "SaveSession")
}

func TestAuthManager_Register_SaveSessionError(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	authClient := &MockAuthClient{}
	sessionRepo := &MockSessionRepository{}

	am := NewAuthManager(authClient, sessionRepo, logger)

	telegramID := int64(123)
	email := "test@example.com"
	password := "password"
	name := "Test User"

	expectedUserID := "user123"
	expectedTenantID := "tenant123"
	expectedAccessToken := "access123"
	expectedRefreshToken := "refresh123"
	expectedAccessExp := time.Now().Add(time.Hour)
	expectedRefreshExp := time.Now().Add(24 * time.Hour)

	authClient.On("Register", ctx, email, password, name).Return(
		expectedUserID, expectedTenantID, expectedAccessToken, expectedRefreshToken,
		expectedAccessExp, expectedRefreshExp, nil)

	expectedError := errors.New("save session failed")
	sessionRepo.On("SaveSession", ctx, mock.AnythingOfType("*repository.UserSession")).Return(expectedError)

	err := am.Register(ctx, telegramID, email, password, name)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	authClient.AssertExpectations(t)
	sessionRepo.AssertExpectations(t)
}

func TestAuthManager_Login_Success(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	authClient := &MockAuthClient{}
	sessionRepo := &MockSessionRepository{}

	am := NewAuthManager(authClient, sessionRepo, logger)

	telegramID := int64(123)
	email := "test@example.com"
	password := "password"

	expectedUserID := "user123"
	expectedTenantID := "tenant123"
	expectedAccessToken := "access123"
	expectedRefreshToken := "refresh123"
	expectedAccessExp := time.Now().Add(time.Hour)
	expectedRefreshExp := time.Now().Add(24 * time.Hour)

	authClient.On("Login", ctx, email, password).Return(
		expectedUserID, expectedTenantID, expectedAccessToken, expectedRefreshToken,
		expectedAccessExp, expectedRefreshExp, nil)

	sessionRepo.On("SaveSession", ctx, mock.MatchedBy(func(session *repository.UserSession) bool {
		return session.TelegramID == telegramID &&
			session.UserID == expectedUserID &&
			session.TenantID == expectedTenantID &&
			session.AccessToken == expectedAccessToken &&
			session.RefreshToken == expectedRefreshToken
	})).Return(nil)

	err := am.Login(ctx, telegramID, email, password)

	assert.NoError(t, err)
	authClient.AssertExpectations(t)
	sessionRepo.AssertExpectations(t)
}

func TestAuthManager_Login_AuthError(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	authClient := &MockAuthClient{}
	sessionRepo := &MockSessionRepository{}

	am := NewAuthManager(authClient, sessionRepo, logger)

	telegramID := int64(123)
	email := "test@example.com"
	password := "password"

	expectedError := errors.New("auth failed")

	authClient.On("Login", ctx, email, password).Return(
		"", "", "", "", time.Time{}, time.Time{}, expectedError)

	err := am.Login(ctx, telegramID, email, password)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	authClient.AssertExpectations(t)
	sessionRepo.AssertNotCalled(t, "SaveSession")
}

func TestAuthManager_Logout(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	authClient := &MockAuthClient{}
	sessionRepo := &MockSessionRepository{}

	am := NewAuthManager(authClient, sessionRepo, logger)

	telegramID := int64(123)

	sessionRepo.On("DeleteSession", ctx, telegramID).Return(nil)

	err := am.Logout(ctx, telegramID)

	assert.NoError(t, err)
	sessionRepo.AssertExpectations(t)
}

func TestAuthManager_Logout_Error(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	authClient := &MockAuthClient{}
	sessionRepo := &MockSessionRepository{}

	am := NewAuthManager(authClient, sessionRepo, logger)

	telegramID := int64(123)
	expectedError := errors.New("delete session failed")

	sessionRepo.On("DeleteSession", ctx, telegramID).Return(expectedError)

	err := am.Logout(ctx, telegramID)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	sessionRepo.AssertExpectations(t)
}

func TestAuthManager_GetSession_Success(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	authClient := &MockAuthClient{}
	sessionRepo := &MockSessionRepository{}

	am := NewAuthManager(authClient, sessionRepo, logger)

	telegramID := int64(123)
	expectedSession := &repository.UserSession{
		TelegramID:            telegramID,
		UserID:                "user123",
		TenantID:              "tenant123",
		AccessToken:           "access123",
		RefreshToken:          "refresh123",
		AccessTokenExpiresAt:  time.Now().Add(time.Hour),
		RefreshTokenExpiresAt: time.Now().Add(24 * time.Hour),
	}

	sessionRepo.On("GetSession", ctx, telegramID).Return(expectedSession, nil)

	session, err := am.GetSession(ctx, telegramID)

	assert.NoError(t, err)
	assert.Equal(t, expectedSession, session)
	sessionRepo.AssertExpectations(t)
}

func TestAuthManager_GetSession_ExpiredToken(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	authClient := &MockAuthClient{}
	sessionRepo := &MockSessionRepository{}

	am := NewAuthManager(authClient, sessionRepo, logger)

	telegramID := int64(123)
	expiredSession := &repository.UserSession{
		TelegramID:            telegramID,
		UserID:                "user123",
		TenantID:              "tenant123",
		AccessToken:           "access123",
		RefreshToken:          "refresh123",
		AccessTokenExpiresAt:  time.Now().Add(-time.Hour), // Expired
		RefreshTokenExpiresAt: time.Now().Add(24 * time.Hour),
	}

	refreshedSession := &repository.UserSession{
		TelegramID:            telegramID,
		UserID:                "user123",
		TenantID:              "tenant123",
		AccessToken:           "new_access123",
		RefreshToken:          "new_refresh123",
		AccessTokenExpiresAt:  time.Now().Add(time.Hour),
		RefreshTokenExpiresAt: time.Now().Add(24 * time.Hour),
	}

	// First call returns expired session
	sessionRepo.On("GetSession", ctx, telegramID).Return(expiredSession, nil).Once()

	// Refresh token call - should be called with the old refresh token
	authClient.On("RefreshToken", ctx, expiredSession.RefreshToken).Return(
		"new_access123", "new_refresh123",
		time.Now().Add(time.Hour), time.Now().Add(24 * time.Hour), nil)

	// Update tokens call
	sessionRepo.On("UpdateTokens", ctx, telegramID, mock.AnythingOfType("*repository.TokenPair")).Return(nil)

	// Second call returns refreshed session (with valid tokens)
	sessionRepo.On("GetSession", ctx, telegramID).Return(refreshedSession, nil).Once()

	// Test GetSession with automatic refresh
	session, err := am.GetSession(ctx, telegramID)

	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, "new_access123", session.AccessToken)
	assert.Equal(t, "new_refresh123", session.RefreshToken)
	authClient.AssertExpectations(t)
	sessionRepo.AssertExpectations(t)
}

func TestAuthManager_GetSession_RefreshTokenError(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	authClient := &MockAuthClient{}
	sessionRepo := &MockSessionRepository{}

	am := NewAuthManager(authClient, sessionRepo, logger)

	telegramID := int64(123)
	expiredSession := &repository.UserSession{
		TelegramID:            telegramID,
		UserID:                "user123",
		TenantID:              "tenant123",
		AccessToken:           "access123",
		RefreshToken:          "refresh123",
		AccessTokenExpiresAt:  time.Now().Add(-time.Hour), // Expired
		RefreshTokenExpiresAt: time.Now().Add(24 * time.Hour),
	}

	sessionRepo.On("GetSession", ctx, telegramID).Return(expiredSession, nil)

	expectedError := errors.New("refresh token failed")
	authClient.On("RefreshToken", ctx, expiredSession.RefreshToken).Return(
		"", "", time.Time{}, time.Time{}, expectedError)

	session, err := am.GetSession(ctx, telegramID)

	assert.Error(t, err)
	assert.Nil(t, session)
	assert.Equal(t, expectedError, err)
	authClient.AssertExpectations(t)
	sessionRepo.AssertExpectations(t)
}

func TestAuthManager_RefreshTokens_Success(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	authClient := &MockAuthClient{}
	sessionRepo := &MockSessionRepository{}

	am := NewAuthManager(authClient, sessionRepo, logger)

	telegramID := int64(123)
	existingSession := &repository.UserSession{
		TelegramID:   telegramID,
		RefreshToken: "old_refresh123",
	}

	newAccessToken := "new_access123"
	newRefreshToken := "new_refresh123"
	newAccessExp := time.Now().Add(time.Hour)
	newRefreshExp := time.Now().Add(24 * time.Hour)

	sessionRepo.On("GetSession", ctx, telegramID).Return(existingSession, nil)

	authClient.On("RefreshToken", ctx, existingSession.RefreshToken).Return(
		newAccessToken, newRefreshToken, newAccessExp, newRefreshExp, nil)

	sessionRepo.On("UpdateTokens", ctx, telegramID, mock.MatchedBy(func(tokens *repository.TokenPair) bool {
		return tokens.AccessToken == newAccessToken &&
			tokens.RefreshToken == newRefreshToken &&
			tokens.AccessTokenExpiresAt == newAccessExp &&
			tokens.RefreshTokenExpiresAt == newRefreshExp
	})).Return(nil)

	err := am.RefreshTokens(ctx, telegramID)

	assert.NoError(t, err)
	authClient.AssertExpectations(t)
	sessionRepo.AssertExpectations(t)
}

func TestAuthManager_RefreshTokens_GetSessionError(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	authClient := &MockAuthClient{}
	sessionRepo := &MockSessionRepository{}

	am := NewAuthManager(authClient, sessionRepo, logger)

	telegramID := int64(123)
	expectedError := errors.New("get session failed")

	sessionRepo.On("GetSession", ctx, telegramID).Return(nil, expectedError)

	err := am.RefreshTokens(ctx, telegramID)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	sessionRepo.AssertExpectations(t)
	authClient.AssertNotCalled(t, "RefreshToken")
}

func TestAuthManager_RefreshTokens_RefreshError(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	authClient := &MockAuthClient{}
	sessionRepo := &MockSessionRepository{}

	am := NewAuthManager(authClient, sessionRepo, logger)

	telegramID := int64(123)
	existingSession := &repository.UserSession{
		TelegramID:   telegramID,
		RefreshToken: "old_refresh123",
	}

	expectedError := errors.New("refresh token failed")

	sessionRepo.On("GetSession", ctx, telegramID).Return(existingSession, nil)

	authClient.On("RefreshToken", ctx, existingSession.RefreshToken).Return(
		"", "", time.Time{}, time.Time{}, expectedError)

	err := am.RefreshTokens(ctx, telegramID)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	authClient.AssertExpectations(t)
	sessionRepo.AssertExpectations(t)
	sessionRepo.AssertNotCalled(t, "UpdateTokens")
}

func TestAuthManager_RefreshTokens_UpdateTokensError(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	authClient := &MockAuthClient{}
	sessionRepo := &MockSessionRepository{}

	am := NewAuthManager(authClient, sessionRepo, logger)

	telegramID := int64(123)
	existingSession := &repository.UserSession{
		TelegramID:   telegramID,
		RefreshToken: "old_refresh123",
	}

	newAccessToken := "new_access123"
	newRefreshToken := "new_refresh123"
	newAccessExp := time.Now().Add(time.Hour)
	newRefreshExp := time.Now().Add(24 * time.Hour)

	sessionRepo.On("GetSession", ctx, telegramID).Return(existingSession, nil)

	authClient.On("RefreshToken", ctx, existingSession.RefreshToken).Return(
		newAccessToken, newRefreshToken, newAccessExp, newRefreshExp, nil)

	expectedError := errors.New("update tokens failed")
	sessionRepo.On("UpdateTokens", ctx, telegramID, mock.AnythingOfType("*repository.TokenPair")).Return(expectedError)

	err := am.RefreshTokens(ctx, telegramID)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	authClient.AssertExpectations(t)
	sessionRepo.AssertExpectations(t)
}

func TestAuthManager_GetSession_RefreshTokenExpired(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	authClient := &MockAuthClient{}
	sessionRepo := &MockSessionRepository{}

	am := NewAuthManager(authClient, sessionRepo, logger)

	telegramID := int64(123)
	expiredSession := &repository.UserSession{
		TelegramID:            telegramID,
		UserID:                "user123",
		TenantID:              "tenant123",
		AccessToken:           "access123",
		RefreshToken:          "refresh123",
		AccessTokenExpiresAt:  time.Now().Add(-time.Hour), // Expired
		RefreshTokenExpiresAt: time.Now().Add(-time.Hour), // Also expired
	}

	// Mock the session repository to return expired session
	sessionRepo.On("GetSession", ctx, telegramID).Return(expiredSession, nil)

	// Test GetSession with expired refresh token
	session, err := am.GetSession(ctx, telegramID)

	assert.Error(t, err)
	assert.Nil(t, session)
	assert.Contains(t, err.Error(), "refresh token expired")
	authClient.AssertNotCalled(t, "RefreshToken")
	sessionRepo.AssertExpectations(t)
}

func TestAuthManager_GetSession_RealWorldScenario(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	authClient := &MockAuthClient{}
	sessionRepo := &MockSessionRepository{}

	am := NewAuthManager(authClient, sessionRepo, logger)

	telegramID := int64(123)
	
	// Scenario: User has a session with expired access token but valid refresh token
	expiredSession := &repository.UserSession{
		TelegramID:            telegramID,
		UserID:                "user123",
		TenantID:              "tenant123",
		AccessToken:           "old_access_token",
		RefreshToken:          "valid_refresh_token",
		AccessTokenExpiresAt:  time.Now().Add(-time.Minute), // Just expired
		RefreshTokenExpiresAt: time.Now().Add(24 * time.Hour), // Still valid
	}

	refreshedSession := &repository.UserSession{
		TelegramID:            telegramID,
		UserID:                "user123",
		TenantID:              "tenant123",
		AccessToken:           "new_access_token",
		RefreshToken:          "new_refresh_token",
		AccessTokenExpiresAt:  time.Now().Add(time.Hour), // Valid for 1 hour
		RefreshTokenExpiresAt: time.Now().Add(24 * time.Hour),
	}

	// First call returns expired session
	sessionRepo.On("GetSession", ctx, telegramID).Return(expiredSession, nil).Once()

	// Refresh token call succeeds
	authClient.On("RefreshToken", ctx, expiredSession.RefreshToken).Return(
		"new_access_token", "new_refresh_token",
		time.Now().Add(time.Hour), time.Now().Add(24 * time.Hour), nil)

	// Update tokens call succeeds
	sessionRepo.On("UpdateTokens", ctx, telegramID, mock.AnythingOfType("*repository.TokenPair")).Return(nil)

	// Second call returns refreshed session
	sessionRepo.On("GetSession", ctx, telegramID).Return(refreshedSession, nil).Once()

	// Test GetSession - should automatically refresh and return valid session
	session, err := am.GetSession(ctx, telegramID)

	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, "new_access_token", session.AccessToken)
	assert.Equal(t, "new_refresh_token", session.RefreshToken)
	assert.True(t, session.AccessTokenExpiresAt.After(time.Now()))
	
	authClient.AssertExpectations(t)
	sessionRepo.AssertExpectations(t)
}
