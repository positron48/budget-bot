//go:build !withgrpc
// +build !withgrpc

package grpc

import (
	"context"
	"fmt"
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
		Session: &pb.TelegramSession{
			SessionId:        sessionID,
			UserId:           "user_123",
			TelegramUserId:   "12345",
			TenantId:         "tenant_123",
			CreatedAt:        nil,
			ExpiresAt:        nil,
			IsActive:         true,
		},
		User:   nil,
		Tenant: nil,
	}, nil
}

func (f *FakeOAuthClient) RevokeTelegramSession(ctx context.Context, sessionID string, telegramUserID int64) error {
	return nil
}

func (f *FakeOAuthClient) ListTelegramSessions(ctx context.Context, telegramUserID int64) ([]*pb.TelegramSession, error) {
	return []*pb.TelegramSession{
		{
			SessionId:        "session_1",
			UserId:           "user_123",
			TelegramUserId:   fmt.Sprintf("%d", telegramUserID),
			TenantId:         "tenant_123",
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
			Email:          "test@example.com",
			TelegramUserId: fmt.Sprintf("%d", telegramUserID),
			IpAddress:      "127.0.0.1",
			UserAgent:      "TelegramBot/1.0",
			Action:         "generate_link",
			Status:         "success",
			ErrorMessage:   "",
			CreatedAt:      nil,
		},
	}, 1, nil
}



// WireClients (default build) returns nil clients so the app uses fakes.
// To enable real clients, build with -tags withgrpc and ensure proto is generated.
func WireClients(_ *zap.Logger) (CategoryClient, ReportClient, TenantClient, TransactionClient, OAuthClient) {
    return nil, nil, nil, nil, &FakeOAuthClient{}
}


