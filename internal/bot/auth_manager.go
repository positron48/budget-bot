// Package bot contains the core Telegram bot business logic.
package bot

import (
	"context"
	"time"

	"budget-bot/internal/repository"
	"go.uber.org/zap"
)

// AuthClient defines auth server operations used by the bot.
type AuthClient interface {
	Register(ctx context.Context, email, password, name string) (userID string, tenantID string, accessToken string, refreshToken string, accessExp time.Time, refreshExp time.Time, err error)
	Login(ctx context.Context, email, password string) (userID string, tenantID string, accessToken string, refreshToken string, accessExp time.Time, refreshExp time.Time, err error)
	RefreshToken(ctx context.Context, refreshToken string) (accessToken string, refreshTokenNew string, accessExp time.Time, refreshExp time.Time, err error)
}

// AuthManager coordinates auth flows and session persistence.
type AuthManager struct {
	authClient  AuthClient
	sessionRepo repository.SessionRepository
	logger     *zap.Logger
}

// NewAuthManager constructs an AuthManager.
func NewAuthManager(authClient AuthClient, sessionRepo repository.SessionRepository, logger *zap.Logger) *AuthManager {
	return &AuthManager{authClient: authClient, sessionRepo: sessionRepo, logger: logger}
}

// Register registers and stores session tokens for a user.
func (am *AuthManager) Register(ctx context.Context, telegramID int64, email, password, name string) error {
	userID, tenantID, accessToken, refreshToken, accessExp, refreshExp, err := am.authClient.Register(ctx, email, password, name)
	if err != nil {
		return err
	}
	return am.sessionRepo.SaveSession(ctx, &repository.UserSession{
		TelegramID:            telegramID,
		UserID:                userID,
		TenantID:              tenantID,
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresAt:  accessExp,
		RefreshTokenExpiresAt: refreshExp,
	})
}

// Login authenticates and stores session tokens for a user.
func (am *AuthManager) Login(ctx context.Context, telegramID int64, email, password string) error {
	userID, tenantID, accessToken, refreshToken, accessExp, refreshExp, err := am.authClient.Login(ctx, email, password)
	if err != nil {
		return err
	}
	return am.sessionRepo.SaveSession(ctx, &repository.UserSession{
		TelegramID:            telegramID,
		UserID:                userID,
		TenantID:              tenantID,
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresAt:  accessExp,
		RefreshTokenExpiresAt: refreshExp,
	})
}

// Logout removes stored session for a user.
func (am *AuthManager) Logout(ctx context.Context, telegramID int64) error {
	return am.sessionRepo.DeleteSession(ctx, telegramID)
}

// GetSession returns current session for a user.
func (am *AuthManager) GetSession(ctx context.Context, telegramID int64) (*repository.UserSession, error) {
	session, err := am.sessionRepo.GetSession(ctx, telegramID)
	if err != nil {
		return nil, err
	}
	
	// Проверяем, не истек ли access token
	if time.Now().After(session.AccessTokenExpiresAt) {
		am.logger.Debug("Access token expired, attempting refresh", 
			zap.Int64("telegramID", telegramID),
			zap.Time("expiresAt", session.AccessTokenExpiresAt))
		
		// Пытаемся обновить токены
		err := am.RefreshTokens(ctx, telegramID)
		if err != nil {
			am.logger.Error("Failed to refresh tokens", 
				zap.Int64("telegramID", telegramID),
				zap.Error(err))
			return nil, err
		}
		
		// Получаем обновленную сессию
		session, err = am.sessionRepo.GetSession(ctx, telegramID)
		if err != nil {
			return nil, err
		}
		
		am.logger.Debug("Tokens refreshed successfully", 
			zap.Int64("telegramID", telegramID),
			zap.Time("newExpiresAt", session.AccessTokenExpiresAt))
	}
	
	return session, nil
}

// RefreshTokens refreshes auth tokens and stores them.
func (am *AuthManager) RefreshTokens(ctx context.Context, telegramID int64) error {
	s, err := am.sessionRepo.GetSession(ctx, telegramID)
	if err != nil {
		return err
	}
	access, refresh, accessExp, refreshExp, err := am.authClient.RefreshToken(ctx, s.RefreshToken)
	if err != nil {
		return err
	}
	return am.sessionRepo.UpdateTokens(ctx, telegramID, &repository.TokenPair{
		AccessToken:           access,
		RefreshToken:          refresh,
		AccessTokenExpiresAt:  accessExp,
		RefreshTokenExpiresAt: refreshExp,
	})
}


