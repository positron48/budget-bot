// Package bot contains the core Telegram bot business logic.
package bot

import (
	"context"
	"fmt"
	"math"
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
		am.logger.Debug("Failed to get session from repository", 
			zap.Int64("telegramID", telegramID),
			zap.Error(err))
		return nil, err
	}
	
	am.logger.Debug("Retrieved session from repository", 
		zap.Int64("telegramID", telegramID),
		zap.Time("accessTokenExpiresAt", session.AccessTokenExpiresAt),
		zap.Time("refreshTokenExpiresAt", session.RefreshTokenExpiresAt),
		zap.Time("now", time.Now()))
	
	// Проверяем, не истек ли access token
	if time.Now().After(session.AccessTokenExpiresAt) {
		am.logger.Info("Access token expired, attempting refresh", 
			zap.Int64("telegramID", telegramID),
			zap.Time("expiresAt", session.AccessTokenExpiresAt))
		
		// Проверяем, не истек ли refresh token
		if time.Now().After(session.RefreshTokenExpiresAt) {
			am.logger.Error("Refresh token also expired, cannot refresh access token", 
				zap.Int64("telegramID", telegramID),
				zap.Time("refreshExpiresAt", session.RefreshTokenExpiresAt))
			return nil, fmt.Errorf("refresh token expired")
		}
		
		// Пытаемся обновить токены, передавая текущую сессию
		err := am.RefreshTokensWithSession(ctx, session)
		if err != nil {
			am.logger.Error("Failed to refresh tokens", 
				zap.Int64("telegramID", telegramID),
				zap.Error(err))
			return nil, err
		}
		
		// Получаем обновленную сессию
		session, err = am.sessionRepo.GetSession(ctx, telegramID)
		if err != nil {
			am.logger.Error("Failed to get updated session after refresh", 
				zap.Int64("telegramID", telegramID),
				zap.Error(err))
			return nil, err
		}
		
		am.logger.Info("Tokens refreshed successfully", 
			zap.Int64("telegramID", telegramID),
			zap.Time("newExpiresAt", session.AccessTokenExpiresAt))
	} else {
		am.logger.Debug("Access token is still valid", 
			zap.Int64("telegramID", telegramID),
			zap.Time("expiresAt", session.AccessTokenExpiresAt))
	}
	
	return session, nil
}

// RefreshTokens refreshes auth tokens and stores them.
func (am *AuthManager) RefreshTokens(ctx context.Context, telegramID int64) error {
	s, err := am.sessionRepo.GetSession(ctx, telegramID)
	if err != nil {
		return err
	}
	return am.RefreshTokensWithSession(ctx, s)
}

// RefreshTokensWithSession refreshes auth tokens using the provided session.
func (am *AuthManager) RefreshTokensWithSession(ctx context.Context, session *repository.UserSession) error {
	am.logger.Debug("Starting token refresh", zap.Int64("telegramID", session.TelegramID))
	
	am.logger.Debug("Calling auth client RefreshToken", 
		zap.Int64("telegramID", session.TelegramID),
		zap.String("refreshToken", session.RefreshToken[:int(math.Min(float64(len(session.RefreshToken)), 10))] + "..."))
	
	access, refresh, accessExp, refreshExp, err := am.authClient.RefreshToken(ctx, session.RefreshToken)
	if err != nil {
		am.logger.Error("Auth client RefreshToken failed", 
			zap.Int64("telegramID", session.TelegramID),
			zap.Error(err))
		return err
	}
	
	am.logger.Debug("Auth client RefreshToken succeeded", 
		zap.Int64("telegramID", session.TelegramID),
		zap.String("newAccessToken", access[:int(math.Min(float64(len(access)), 10))] + "..."),
		zap.String("newRefreshToken", refresh[:int(math.Min(float64(len(refresh)), 10))] + "..."),
		zap.Time("newAccessExp", accessExp),
		zap.Time("newRefreshExp", refreshExp))
	
	err = am.sessionRepo.UpdateTokens(ctx, session.TelegramID, &repository.TokenPair{
		AccessToken:           access,
		RefreshToken:          refresh,
		AccessTokenExpiresAt:  accessExp,
		RefreshTokenExpiresAt: refreshExp,
	})
	if err != nil {
		am.logger.Error("Failed to update tokens in repository", 
			zap.Int64("telegramID", session.TelegramID),
			zap.Error(err))
		return err
	}
	
	am.logger.Info("Token refresh completed successfully", 
		zap.Int64("telegramID", session.TelegramID),
		zap.Time("newAccessExp", accessExp),
		zap.Time("newRefreshExp", refreshExp))
	
	return nil
}


