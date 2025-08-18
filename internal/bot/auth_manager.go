package bot

import (
	"context"
	"time"

	"budget-bot/internal/repository"
	"go.uber.org/zap"
)

type AuthClient interface {
	Register(ctx context.Context, email, password, name string) (userID string, tenantID string, accessToken string, refreshToken string, accessExp time.Time, refreshExp time.Time, err error)
	Login(ctx context.Context, email, password string) (userID string, tenantID string, accessToken string, refreshToken string, accessExp time.Time, refreshExp time.Time, err error)
	RefreshToken(ctx context.Context, refreshToken string) (accessToken string, refreshTokenNew string, accessExp time.Time, refreshExp time.Time, err error)
}

type AuthManager struct {
	authClient  AuthClient
	sessionRepo repository.SessionRepository
	logger     *zap.Logger
}

func NewAuthManager(authClient AuthClient, sessionRepo repository.SessionRepository, logger *zap.Logger) *AuthManager {
	return &AuthManager{authClient: authClient, sessionRepo: sessionRepo, logger: logger}
}

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

func (am *AuthManager) Logout(ctx context.Context, telegramID int64) error {
	return am.sessionRepo.DeleteSession(ctx, telegramID)
}

func (am *AuthManager) GetSession(ctx context.Context, telegramID int64) (*repository.UserSession, error) {
	return am.sessionRepo.GetSession(ctx, telegramID)
}

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


