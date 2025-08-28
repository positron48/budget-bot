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

// FakeAuthClient implements AuthClient for testing
type FakeAuthClient struct{}

func (f *FakeAuthClient) Register(_ context.Context, _, _, _ string) (string, string, string, string, time.Time, time.Time, error) {
	return "user_123", "tenant_123", "access_token_123", "refresh_token_123", time.Now().Add(15*time.Minute), time.Now().Add(720*time.Hour), nil
}

func (f *FakeAuthClient) Login(_ context.Context, _, _ string) (string, string, string, string, time.Time, time.Time, error) {
	return "user_123", "tenant_123", "access_token_123", "refresh_token_123", time.Now().Add(15*time.Minute), time.Now().Add(720*time.Hour), nil
}

func (f *FakeAuthClient) RefreshToken(_ context.Context, _ string) (string, string, time.Time, time.Time, error) {
	return "new_access_token_123", "new_refresh_token_123", time.Now().Add(15*time.Minute), time.Now().Add(720*time.Hour), nil
}

func (f *FakeOAuthClient) GenerateAuthLink(_ context.Context, _ string, _ int64, _, _ string) (string, string, time.Time, error) {
	return "https://example.com/auth?token=test", "auth_token_123", time.Now().Add(5*time.Minute), nil
}

func (f *FakeOAuthClient) VerifyAuthCode(_ context.Context, _, _ string, _ int64) (*VerifyAuthCodeResult, error) {
	return &VerifyAuthCodeResult{
		Tokens: &pb.TokenPair{
			AccessToken:           "access_token_123",
			RefreshToken:          "refresh_token_123",
			AccessTokenExpiresAt:  nil,
			RefreshTokenExpiresAt: nil,
			TokenType:             "Bearer",
		},
		SessionID: "session_123",
		User: &pb.User{
			Id:    "user_123",
			Email: "test@example.com",
		},
		Memberships: []*pb.TenantMembership{
			{
				Tenant: &pb.Tenant{
					Id: "tenant_123",
				},
				Role: pb.TenantRole_TENANT_ROLE_OWNER,
			},
		},
	}, nil
}

func (f *FakeOAuthClient) CancelAuth(_ context.Context, _ string, _ int64) error {
	return nil
}

func (f *FakeOAuthClient) GetAuthStatus(_ context.Context, _ string) (string, string, time.Time, error) {
	return "pending", "test@example.com", time.Now().Add(5*time.Minute), nil
}

func (f *FakeOAuthClient) GetTelegramSession(_ context.Context, sessionID string) (*pb.GetTelegramSessionResponse, error) {
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

func (f *FakeOAuthClient) RevokeTelegramSession(_ context.Context, _ string, _ int64) error {
	return nil
}

func (f *FakeOAuthClient) ListTelegramSessions(_ context.Context, telegramUserID int64) ([]*pb.TelegramSession, error) {
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

func (f *FakeOAuthClient) GetAuthLogs(_ context.Context, telegramUserID int64, _, _ int32) ([]*pb.AuthLogEntry, int32, error) {
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
func WireClients(_ *zap.Logger) (CategoryClient, ReportClient, TenantClient, TransactionClient, OAuthClient, AuthClientInterface) {
    return nil, nil, nil, nil, &FakeOAuthClient{}, &FakeAuthClient{}
}


