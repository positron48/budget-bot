package bot

import (
	"context"
	"time"

	grpcclient "budget-bot/internal/grpc"
	pb "budget-bot/internal/pb/budget/v1"
)

// TestOAuthClient is a fake OAuth client for testing
type TestOAuthClient struct{}

func (f *TestOAuthClient) GenerateAuthLink(_ context.Context, _ string, _ int64, _, _ string) (string, string, time.Time, error) {
	return "https://example.com/auth", "auth_token", time.Now().Add(5*time.Minute), nil
}

func (f *TestOAuthClient) VerifyAuthCode(_ context.Context, _, _ string, _ int64) (*grpcclient.VerifyAuthCodeResult, error) {
	return &grpcclient.VerifyAuthCodeResult{
		Tokens: &pb.TokenPair{
			AccessToken:           "access_token",
			RefreshToken:          "refresh_token",
			AccessTokenExpiresAt:  nil,
			RefreshTokenExpiresAt: nil,
			TokenType:             "Bearer",
		},
		SessionID: "session_id",
		User: &pb.User{
			Id:    "test_user_id",
			Email: "test@example.com",
		},
		Memberships: []*pb.TenantMembership{
			{
				Tenant: &pb.Tenant{
					Id: "test_tenant_id",
				},
				Role: pb.TenantRole_TENANT_ROLE_OWNER,
			},
		},
	}, nil
}

func (f *TestOAuthClient) CancelAuth(_ context.Context, _ string, _ int64) error {
	return nil
}

func (f *TestOAuthClient) GetAuthStatus(_ context.Context, _ string) (string, string, time.Time, error) {
	return "completed", "test@example.com", time.Now().Add(5*time.Minute), nil
}

func (f *TestOAuthClient) GetTelegramSession(_ context.Context, sessionID string) (*pb.GetTelegramSessionResponse, error) {
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

func (f *TestOAuthClient) RevokeTelegramSession(_ context.Context, _ string, _ int64) error {
	return nil
}

func (f *TestOAuthClient) ListTelegramSessions(_ context.Context, _ int64) ([]*pb.TelegramSession, error) {
	return []*pb.TelegramSession{}, nil
}

func (f *TestOAuthClient) GetAuthLogs(_ context.Context, _ int64, _, _ int32) ([]*pb.AuthLogEntry, int32, error) {
	return []*pb.AuthLogEntry{}, 0, nil
}
