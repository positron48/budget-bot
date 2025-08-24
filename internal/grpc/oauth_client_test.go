package grpc

import (
	"context"
	"net"
	"testing"
	"time"

	pb "budget-bot/internal/pb/budget/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/timestamppb"
	"go.uber.org/zap"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func init() {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	pb.RegisterOAuthServiceServer(s, &mockOAuthServer{})
	go func() {
		if err := s.Serve(lis); err != nil {
			panic(err)
		}
	}()
}

type mockOAuthServer struct {
	pb.UnimplementedOAuthServiceServer
}

func (m *mockOAuthServer) GenerateAuthLink(ctx context.Context, req *pb.GenerateAuthLinkRequest) (*pb.GenerateAuthLinkResponse, error) {
	return &pb.GenerateAuthLinkResponse{
		AuthUrl:   "https://example.com/auth?token=" + req.Email,
		AuthToken: "auth_token_123",
		ExpiresAt: timestamppb.New(time.Now().Add(5 * time.Minute)),
	}, nil
}

func (m *mockOAuthServer) VerifyAuthCode(ctx context.Context, req *pb.VerifyAuthCodeRequest) (*pb.VerifyAuthCodeResponse, error) {
	return &pb.VerifyAuthCodeResponse{
		Tokens: &pb.TokenPair{
			AccessToken:           "access_token_123",
			RefreshToken:          "refresh_token_123",
			AccessTokenExpiresAt:  timestamppb.New(time.Now().Add(time.Hour)),
			RefreshTokenExpiresAt: timestamppb.New(time.Now().Add(24 * time.Hour)),
			TokenType:             "Bearer",
		},
		SessionId: "session_123",
		User: &pb.User{
			Id:    "user_123",
			Email: "test@example.com",
		},
		Memberships: []*pb.TenantMembership{
			{
				Tenant: &pb.Tenant{
					Id:   "tenant_123",
					Name: "Test Tenant",
				},
				Role: pb.TenantRole_TENANT_ROLE_OWNER,
			},
		},
	}, nil
}

func (m *mockOAuthServer) CancelAuth(ctx context.Context, req *pb.CancelAuthRequest) (*pb.CancelAuthResponse, error) {
	return &pb.CancelAuthResponse{}, nil
}

func (m *mockOAuthServer) GetAuthStatus(ctx context.Context, req *pb.GetAuthStatusRequest) (*pb.GetAuthStatusResponse, error) {
	return &pb.GetAuthStatusResponse{
		Status:    pb.GetAuthStatusResponse_STATUS_PENDING,
		Email:     "test@example.com",
		ExpiresAt: timestamppb.New(time.Now().Add(5 * time.Minute)),
	}, nil
}

func (m *mockOAuthServer) GetTelegramSession(ctx context.Context, req *pb.GetTelegramSessionRequest) (*pb.GetTelegramSessionResponse, error) {
	return &pb.GetTelegramSessionResponse{
		Session: &pb.TelegramSession{
			SessionId:        req.SessionId,
			UserId:           "user_123",
			TelegramUserId:   "12345",
			TenantId:         "tenant_123",
			CreatedAt:        timestamppb.New(time.Now()),
			ExpiresAt:        timestamppb.New(time.Now().Add(24 * time.Hour)),
			IsActive:         true,
		},
		User:   nil,
		Tenant: nil,
	}, nil
}

func (m *mockOAuthServer) RevokeTelegramSession(ctx context.Context, req *pb.RevokeTelegramSessionRequest) (*pb.RevokeTelegramSessionResponse, error) {
	return &pb.RevokeTelegramSessionResponse{}, nil
}

func (m *mockOAuthServer) ListTelegramSessions(ctx context.Context, req *pb.ListTelegramSessionsRequest) (*pb.ListTelegramSessionsResponse, error) {
	return &pb.ListTelegramSessionsResponse{
		Sessions: []*pb.TelegramSession{
			{
				SessionId:        "session_1",
				UserId:           "user_123",
				TelegramUserId:   req.TelegramUserId,
				TenantId:         "tenant_123",
				CreatedAt:        timestamppb.New(time.Now()),
				ExpiresAt:        timestamppb.New(time.Now().Add(24 * time.Hour)),
				IsActive:         true,
			},
		},
	}, nil
}

func (m *mockOAuthServer) GetAuthLogs(ctx context.Context, req *pb.GetAuthLogsRequest) (*pb.GetAuthLogsResponse, error) {
	return &pb.GetAuthLogsResponse{
		Logs: []*pb.AuthLogEntry{
			{
				Id:             "log_1",
				TelegramUserId: req.TelegramUserId,
				Email:          "test@example.com",
				Action:         "generate_link",
				Status:         "success",
				IpAddress:      "127.0.0.1",
				UserAgent:      "TelegramBot/1.0",
				CreatedAt:      timestamppb.New(time.Now()),
			},
		},
		TotalCount: 1,
	}, nil
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestOAuthClient_GenerateAuthLink(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := pb.NewOAuthServiceClient(conn)
	oauthClient := NewOAuthClient(client, zap.NewNop())

	authURL, authToken, expiresAt, err := oauthClient.GenerateAuthLink(ctx, "test@example.com", 12345, "TelegramBot/1.0", "127.0.0.1")
	if err != nil {
		t.Fatalf("GenerateAuthLink failed: %v", err)
	}

	if authURL == "" {
		t.Error("Expected non-empty auth URL")
	}
	if authToken == "" {
		t.Error("Expected non-empty auth token")
	}
	if expiresAt.IsZero() {
		t.Error("Expected non-zero expiration time")
	}
}

func TestOAuthClient_VerifyAuthCode(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := pb.NewOAuthServiceClient(conn)
	oauthClient := NewOAuthClient(client, zap.NewNop())

	result, err := oauthClient.VerifyAuthCode(ctx, "auth_token_123", "123456", 12345)
	if err != nil {
		t.Fatalf("VerifyAuthCode failed: %v", err)
	}

	if result == nil {
		t.Error("Expected non-nil result")
	}
	if result.Tokens == nil {
		t.Error("Expected non-nil tokens")
	}
	if result.SessionID == "" {
		t.Error("Expected non-empty session ID")
	}
	if result.Tokens.AccessToken == "" {
		t.Error("Expected non-empty access token")
	}
	if result.Tokens.RefreshToken == "" {
		t.Error("Expected non-empty refresh token")
	}
}

func TestOAuthClient_GetAuthStatus(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := pb.NewOAuthServiceClient(conn)
	oauthClient := NewOAuthClient(client, zap.NewNop())

	status, email, expiresAt, err := oauthClient.GetAuthStatus(ctx, "auth_token_123")
	if err != nil {
		t.Fatalf("GetAuthStatus failed: %v", err)
	}

	if status != "STATUS_PENDING" {
		t.Errorf("Expected status 'STATUS_PENDING', got '%s'", status)
	}
	if email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", email)
	}
	if expiresAt.IsZero() {
		t.Error("Expected non-zero expiration time")
	}
}

func TestOAuthClient_GetTelegramSession(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := pb.NewOAuthServiceClient(conn)
	oauthClient := NewOAuthClient(client, zap.NewNop())

	session, err := oauthClient.GetTelegramSession(ctx, "session_123")
	if err != nil {
		t.Fatalf("GetTelegramSession failed: %v", err)
	}

	if session == nil {
		t.Error("Expected non-nil session")
	}
	if session.Session.SessionId != "session_123" {
		t.Errorf("Expected session ID 'session_123', got '%s'", session.Session.SessionId)
	}
	if !session.Session.IsActive {
		t.Error("Expected active session")
	}
}

func TestOAuthClient_ListTelegramSessions(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := pb.NewOAuthServiceClient(conn)
	oauthClient := NewOAuthClient(client, zap.NewNop())

	sessions, err := oauthClient.ListTelegramSessions(ctx, 12345)
	if err != nil {
		t.Fatalf("ListTelegramSessions failed: %v", err)
	}

	if len(sessions) == 0 {
		t.Error("Expected non-empty sessions list")
	}
	if sessions[0].SessionId != "session_1" {
		t.Errorf("Expected session ID 'session_1', got '%s'", sessions[0].SessionId)
	}
}

func TestOAuthClient_GetAuthLogs(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := pb.NewOAuthServiceClient(conn)
	oauthClient := NewOAuthClient(client, zap.NewNop())

	logs, totalCount, err := oauthClient.GetAuthLogs(ctx, 12345, 10, 0)
	if err != nil {
		t.Fatalf("GetAuthLogs failed: %v", err)
	}

	if len(logs) == 0 {
		t.Error("Expected non-empty logs list")
	}
	if totalCount != 1 {
		t.Errorf("Expected total count 1, got %d", totalCount)
	}
	if logs[0].Action != "generate_link" {
		t.Errorf("Expected action 'generate_link', got '%s'", logs[0].Action)
	}
}
