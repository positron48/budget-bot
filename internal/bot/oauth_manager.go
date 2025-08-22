// Package bot contains the core Telegram bot business logic.
package bot

import (
	"context"
	"fmt"
	"time"

	pb "budget-bot/internal/pb/budget/v1"
	"budget-bot/internal/repository"
	grpcclient "budget-bot/internal/grpc"
	"go.uber.org/zap"
)



// OAuthManager coordinates OAuth flows and session persistence.
type OAuthManager struct {
	oauthClient  grpcclient.OAuthClient
	sessionRepo  repository.SessionRepository
	logger       *zap.Logger
	webBaseURL   string
}

// NewOAuthManager constructs an OAuthManager.
func NewOAuthManager(oauthClient grpcclient.OAuthClient, sessionRepo repository.SessionRepository, logger *zap.Logger, webBaseURL string) *OAuthManager {
	return &OAuthManager{
		oauthClient: oauthClient,
		sessionRepo: sessionRepo,
		logger:      logger,
		webBaseURL:  webBaseURL,
	}
}

// GenerateAuthLink generates an OAuth authorization link for the user.
func (om *OAuthManager) GenerateAuthLink(ctx context.Context, telegramID int64, email, userAgent, ipAddress string) (string, string, time.Time, error) {
	om.logger.Info("Generating OAuth auth link",
		zap.Int64("telegramID", telegramID),
		zap.String("email", email),
		zap.String("ipAddress", ipAddress))

	authURL, authToken, expiresAt, err := om.oauthClient.GenerateAuthLink(ctx, email, telegramID, userAgent, ipAddress)
	if err != nil {
		om.logger.Error("Failed to generate auth link",
			zap.Int64("telegramID", telegramID),
			zap.String("email", email),
			zap.Error(err))
		return "", "", time.Time{}, fmt.Errorf("failed to generate auth link: %w", err)
	}

	om.logger.Info("Auth link generated successfully",
		zap.Int64("telegramID", telegramID),
		zap.String("email", email),
		zap.Time("expiresAt", expiresAt))

	return authURL, authToken, expiresAt, nil
}

// VerifyAuthCode verifies the OAuth verification code and creates a session.
func (om *OAuthManager) VerifyAuthCode(ctx context.Context, telegramID int64, authToken, verificationCode string) error {
	om.logger.Info("Verifying OAuth auth code",
		zap.Int64("telegramID", telegramID),
		zap.String("authToken", authToken),
		zap.String("verificationCode", verificationCode))

	tokens, sessionID, err := om.oauthClient.VerifyAuthCode(ctx, authToken, verificationCode, telegramID)
	if err != nil {
		om.logger.Error("Failed to verify auth code",
			zap.Int64("telegramID", telegramID),
			zap.String("authToken", authToken),
			zap.String("verificationCode", verificationCode),
			zap.Error(err))
		return fmt.Errorf("failed to verify auth code: %w", err)
	}

	om.logger.Info("Received successful response from gRPC",
		zap.Int64("telegramID", telegramID),
		zap.String("sessionID", sessionID),
		zap.String("accessToken", tokens.AccessToken[:10]+"..."), // Логируем только первые 10 символов токена
		zap.String("refreshToken", tokens.RefreshToken[:10]+"..."))

	// Save session to local database
	session := &repository.UserSession{
		TelegramID:            telegramID,
		UserID:                "", // Will be filled from session later
		TenantID:              "", // Will be filled from session later
		AccessToken:           tokens.AccessToken,
		RefreshToken:          tokens.RefreshToken,
		AccessTokenExpiresAt:  tokens.AccessTokenExpiresAt.AsTime(),
		RefreshTokenExpiresAt: tokens.RefreshTokenExpiresAt.AsTime(),
	}

	if err := om.sessionRepo.SaveSession(ctx, session); err != nil {
		om.logger.Error("Failed to save session",
			zap.Int64("telegramID", telegramID),
			zap.String("sessionID", sessionID),
			zap.Error(err))
		return fmt.Errorf("failed to save session: %w", err)
	}

	om.logger.Info("Auth code verified successfully",
		zap.Int64("telegramID", telegramID),
		zap.String("sessionID", sessionID))

	return nil
}

// CancelAuth cancels the OAuth authorization process.
func (om *OAuthManager) CancelAuth(ctx context.Context, telegramID int64, authToken string) error {
	om.logger.Info("Cancelling OAuth auth",
		zap.Int64("telegramID", telegramID),
		zap.String("authToken", authToken))

	err := om.oauthClient.CancelAuth(ctx, authToken, telegramID)
	if err != nil {
		om.logger.Error("Failed to cancel auth",
			zap.Int64("telegramID", telegramID),
			zap.String("authToken", authToken),
			zap.Error(err))
		return fmt.Errorf("failed to cancel auth: %w", err)
	}

	om.logger.Info("Auth cancelled successfully",
		zap.Int64("telegramID", telegramID))

	return nil
}

// GetAuthStatus gets the status of an OAuth authorization.
func (om *OAuthManager) GetAuthStatus(ctx context.Context, authToken string) (string, string, time.Time, error) {
	status, email, expiresAt, err := om.oauthClient.GetAuthStatus(ctx, authToken)
	if err != nil {
		om.logger.Error("Failed to get auth status",
			zap.String("authToken", authToken),
			zap.Error(err))
		return "", "", time.Time{}, fmt.Errorf("failed to get auth status: %w", err)
	}

	return status, email, expiresAt, nil
}

// GetSession returns current session for a user.
func (om *OAuthManager) GetSession(ctx context.Context, telegramID int64) (*repository.UserSession, error) {
	session, err := om.sessionRepo.GetSession(ctx, telegramID)
	if err != nil {
		return nil, err
	}
	

	
	return session, nil
}



// Logout removes stored session for a user.
func (om *OAuthManager) Logout(ctx context.Context, telegramID int64) error {
	om.logger.Info("Logging out user",
		zap.Int64("telegramID", telegramID))

	err := om.sessionRepo.DeleteSession(ctx, telegramID)
	if err != nil {
		om.logger.Error("Failed to delete session",
			zap.Int64("telegramID", telegramID),
			zap.Error(err))
		return fmt.Errorf("failed to delete session: %w", err)
	}

	om.logger.Info("User logged out successfully",
		zap.Int64("telegramID", telegramID))

	return nil
}

// ListSessions lists all sessions for a user.
func (om *OAuthManager) ListSessions(ctx context.Context, telegramID int64) ([]*pb.TelegramSession, error) {
	sessions, err := om.oauthClient.ListTelegramSessions(ctx, telegramID)
	if err != nil {
		om.logger.Error("Failed to list sessions",
			zap.Int64("telegramID", telegramID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	return sessions, nil
}

// RevokeSession revokes a specific session.
func (om *OAuthManager) RevokeSession(ctx context.Context, telegramID int64, sessionID string) error {
	om.logger.Info("Revoking session",
		zap.Int64("telegramID", telegramID),
		zap.String("sessionID", sessionID))

	err := om.oauthClient.RevokeTelegramSession(ctx, sessionID, telegramID)
	if err != nil {
		om.logger.Error("Failed to revoke session",
			zap.Int64("telegramID", telegramID),
			zap.String("sessionID", sessionID),
			zap.Error(err))
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	om.logger.Info("Session revoked successfully",
		zap.Int64("telegramID", telegramID),
		zap.String("sessionID", sessionID))

	return nil
}

// GetAuthLogs gets authentication logs for a user.
func (om *OAuthManager) GetAuthLogs(ctx context.Context, telegramID int64, limit, offset int32) ([]*pb.AuthLogEntry, int32, error) {
	logs, totalCount, err := om.oauthClient.GetAuthLogs(ctx, telegramID, limit, offset)
	if err != nil {
		om.logger.Error("Failed to get auth logs",
			zap.Int64("telegramID", telegramID),
			zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get auth logs: %w", err)
	}

	return logs, totalCount, nil
}
