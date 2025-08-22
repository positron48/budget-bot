//go:build !withgrpc
// +build !withgrpc

package grpc

import (
	"context"
	"time"

	pb "budget-bot/internal/pb/budget/v1"
	"go.uber.org/zap"
)

// FakeOAuthClient implements OAuthClient for testing
type FakeOAuthClient struct{}

func (f *FakeOAuthClient) GenerateAuthLink(ctx context.Context, email string, telegramUserID int64, userAgent, ipAddress string) (string, string, time.Time, error) {
	return "https://example.com/auth?token=test", "auth_token_123", time.Now().Add(5*time.Minute), nil
}

func (f *FakeOAuthClient) VerifyAuthCode(ctx context.Context, authToken, verificationCode string, telegramUserID int64) (*pb.TokenPair, string, error) {
	return &pb.TokenPair{
		AccessToken:           "access_token_123",
		RefreshToken:          "refresh_token_123",
		AccessTokenExpiresAt:  nil,
		RefreshTokenExpiresAt: nil,
		TokenType:             "Bearer",
	}, "session_123", nil
}

func (f *FakeOAuthClient) CancelAuth(ctx context.Context, authToken string, telegramUserID int64) error {
	return nil
}

func (f *FakeOAuthClient) GetAuthStatus(ctx context.Context, authToken string) (string, string, time.Time, error) {
	return "pending", "test@example.com", time.Now().Add(5*time.Minute), nil
}

func (f *FakeOAuthClient) GetTelegramSession(ctx context.Context, sessionID string) (*pb.GetTelegramSessionResponse, error) {
	return &pb.GetTelegramSessionResponse{
		SessionId:        sessionID,
		UserId:           "user_123",
		TenantId:         "tenant_123",
		AccessTokenHash:  "hash_123",
		RefreshTokenHash: "hash_456",
		CreatedAt:        nil,
		ExpiresAt:        nil,
		IsActive:         true,
	}, nil
}

func (f *FakeOAuthClient) RevokeTelegramSession(ctx context.Context, sessionID string, telegramUserID int64) error {
	return nil
}

func (f *FakeOAuthClient) ListTelegramSessions(ctx context.Context, telegramUserID int64) ([]*pb.GetTelegramSessionResponse, error) {
	return []*pb.GetTelegramSessionResponse{
		{
			SessionId:        "session_1",
			UserId:           "user_123",
			TenantId:         "tenant_123",
			AccessTokenHash:  "hash_123",
			RefreshTokenHash: "hash_456",
			CreatedAt:        nil,
			ExpiresAt:        nil,
			IsActive:         true,
		},
	}, nil
}

func (f *FakeOAuthClient) GetAuthLogs(ctx context.Context, telegramUserID int64, limit, offset int32) ([]*pb.AuthLogEntry, int32, error) {
	return []*pb.AuthLogEntry{
		{
			Id:             "log_1",
			TelegramUserId: telegramUserID,
			Email:          "test@example.com",
			Action:         "generate_link",
			Status:         "success",
			IpAddress:      "127.0.0.1",
			UserAgent:      "TelegramBot/1.0",
			CreatedAt:      nil,
		},
	}, 1, nil
}

func (f *FakeOAuthClient) RefreshToken(ctx context.Context, refreshToken string) (string, string, time.Time, time.Time, error) {
	return "access_token_new", "refresh_token_new", time.Now().Add(time.Hour), time.Now().Add(24*time.Hour), nil
}

// WireClients (default build) returns nil clients so the app uses fakes.
// To enable real clients, build with -tags withgrpc and ensure proto is generated.
func WireClients(_ *zap.Logger) (CategoryClient, ReportClient, TenantClient, TransactionClient, OAuthClient) {
    return nil, nil, nil, nil, &FakeOAuthClient{}
}


