package bot

import (
	"context"
	"time"

	pb "budget-bot/internal/pb/budget/v1"
)

// TestOAuthClient is a fake OAuth client for testing
type TestOAuthClient struct{}

func (f *TestOAuthClient) GenerateAuthLink(ctx context.Context, email string, telegramUserID int64, userAgent, ipAddress string) (string, string, time.Time, error) {
	return "https://example.com/auth", "auth_token", time.Now().Add(5*time.Minute), nil
}

func (f *TestOAuthClient) VerifyAuthCode(ctx context.Context, authToken, verificationCode string, telegramUserID int64) (*pb.TokenPair, string, error) {
	return &pb.TokenPair{
		AccessToken:           "access_token",
		RefreshToken:          "refresh_token",
		AccessTokenExpiresAt:  nil,
		RefreshTokenExpiresAt: nil,
		TokenType:             "Bearer",
	}, "session_id", nil
}

func (f *TestOAuthClient) CancelAuth(ctx context.Context, authToken string, telegramUserID int64) error {
	return nil
}

func (f *TestOAuthClient) GetAuthStatus(ctx context.Context, authToken string) (string, string, time.Time, error) {
	return "completed", "test@example.com", time.Now().Add(5*time.Minute), nil
}

func (f *TestOAuthClient) GetTelegramSession(ctx context.Context, sessionID string) (*pb.GetTelegramSessionResponse, error) {
	return &pb.GetTelegramSessionResponse{
		Session: &pb.TelegramSession{
			SessionId:        sessionID,
			UserId:           "user_id",
			TelegramUserId:   "12345",
			TenantId:         "tenant_id",
			CreatedAt:        nil,
			ExpiresAt:        nil,
			IsActive:         true,
		},
		User:   nil,
		Tenant: nil,
	}, nil
}

func (f *TestOAuthClient) RevokeTelegramSession(ctx context.Context, sessionID string, telegramUserID int64) error {
	return nil
}

func (f *TestOAuthClient) ListTelegramSessions(ctx context.Context, telegramUserID int64) ([]*pb.TelegramSession, error) {
	return []*pb.TelegramSession{}, nil
}

func (f *TestOAuthClient) GetAuthLogs(ctx context.Context, telegramUserID int64, limit, offset int32) ([]*pb.AuthLogEntry, int32, error) {
	return []*pb.AuthLogEntry{}, 0, nil
}
