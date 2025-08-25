package bot

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	grpcclient "budget-bot/internal/grpc"
	pb "budget-bot/internal/pb/budget/v1"
	"budget-bot/internal/repository"
	"google.golang.org/protobuf/types/known/timestamppb"
	_ "modernc.org/sqlite"
	"go.uber.org/zap"
)

type fakeOAuthClient struct{}

func (f *fakeOAuthClient) GenerateAuthLink(ctx context.Context, email string, telegramUserID int64, userAgent, ipAddress string) (string, string, time.Time, error) {
	return "https://example.com/auth?token=test", "auth_token_123", time.Now().Add(5*time.Minute), nil
}

func (f *fakeOAuthClient) VerifyAuthCode(ctx context.Context, authToken, verificationCode string, telegramUserID int64) (*grpcclient.VerifyAuthCodeResult, error) {
	return &grpcclient.VerifyAuthCodeResult{
		Tokens: &pb.TokenPair{
			AccessToken:           "access_token_123",
			RefreshToken:          "refresh_token_123",
			AccessTokenExpiresAt:  timestamppb.New(time.Now().Add(time.Hour)),
			RefreshTokenExpiresAt: timestamppb.New(time.Now().Add(24*time.Hour)),
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

func (f *fakeOAuthClient) CancelAuth(ctx context.Context, authToken string, telegramUserID int64) error {
	return nil
}

func (f *fakeOAuthClient) GetAuthStatus(ctx context.Context, authToken string) (string, string, time.Time, error) {
	return "pending", "test@example.com", time.Now().Add(5*time.Minute), nil
}

func (f *fakeOAuthClient) GetTelegramSession(ctx context.Context, sessionID string) (*pb.GetTelegramSessionResponse, error) {
	return &pb.GetTelegramSessionResponse{
		Session: &pb.TelegramSession{
			SessionId:        sessionID,
			UserId:           "user_123",
			TelegramUserId:   "12345",
			TenantId:         "tenant_123",
			CreatedAt:        timestamppb.New(time.Now()),
			ExpiresAt:        timestamppb.New(time.Now().Add(24*time.Hour)),
			IsActive:         true,
		},
		User:   nil,
		Tenant: nil,
	}, nil
}

func (f *fakeOAuthClient) RevokeTelegramSession(ctx context.Context, sessionID string, telegramUserID int64) error {
	return nil
}

func (f *fakeOAuthClient) ListTelegramSessions(ctx context.Context, telegramUserID int64) ([]*pb.TelegramSession, error) {
	return []*pb.TelegramSession{
		{
			SessionId:        "session_1",
			UserId:           "user_123",
			TelegramUserId:   fmt.Sprintf("%d", telegramUserID),
			TenantId:         "tenant_123",
			CreatedAt:        timestamppb.New(time.Now()),
			ExpiresAt:        timestamppb.New(time.Now().Add(24*time.Hour)),
			IsActive:         true,
		},
	}, nil
}

func (f *fakeOAuthClient) GetAuthLogs(ctx context.Context, telegramUserID int64, limit, offset int32) ([]*pb.AuthLogEntry, int32, error) {
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
			CreatedAt:      timestamppb.New(time.Now()),
		},
	}, 1, nil
}



func setupOAuthSessionDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS user_sessions (
			telegram_id INTEGER PRIMARY KEY,
			user_id TEXT NOT NULL,
			tenant_id TEXT NOT NULL,
			access_token TEXT NOT NULL,
			refresh_token TEXT NOT NULL,
			access_token_expires_at TIMESTAMP NOT NULL,
			refresh_token_expires_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		t.Fatalf("create table: %v", err)
	}
	return db
}

func TestOAuthManager_GenerateAuthLink(t *testing.T) {
	db := setupOAuthSessionDB(t)
	defer func() { _ = db.Close() }()
	sessions := repository.NewSQLiteSessionRepository(db)
	om := NewOAuthManager(&fakeOAuthClient{}, sessions, zap.NewNop(), "http://localhost:3000")
	ctx := context.Background()

	authURL, authToken, expiresAt, err := om.GenerateAuthLink(ctx, 12345, "test@example.com", "TelegramBot/1.0", "127.0.0.1")
	if err != nil {
		t.Fatalf("GenerateAuthLink: %v", err)
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

func TestOAuthManager_VerifyAuthCode(t *testing.T) {
	db := setupOAuthSessionDB(t)
	defer func() { _ = db.Close() }()
	sessions := repository.NewSQLiteSessionRepository(db)
	om := NewOAuthManager(&fakeOAuthClient{}, sessions, zap.NewNop(), "http://localhost:3000")
	ctx := context.Background()

	err := om.VerifyAuthCode(ctx, 12345, "auth_token_123", "123456")
	if err != nil {
		t.Fatalf("VerifyAuthCode: %v", err)
	}

	// Verify session was saved
	session, err := sessions.GetSession(ctx, 12345)
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if session.AccessToken != "access_token_123" {
		t.Errorf("Expected access token 'access_token_123', got '%s'", session.AccessToken)
	}
	if session.RefreshToken != "refresh_token_123" {
		t.Errorf("Expected refresh token 'refresh_token_123', got '%s'", session.RefreshToken)
	}
}

func TestOAuthManager_GetSession(t *testing.T) {
	db := setupOAuthSessionDB(t)
	defer func() { _ = db.Close() }()
	sessions := repository.NewSQLiteSessionRepository(db)
	om := NewOAuthManager(&fakeOAuthClient{}, sessions, zap.NewNop(), "http://localhost:3000")
	ctx := context.Background()

	// First verify auth code to create session
	err := om.VerifyAuthCode(ctx, 12345, "auth_token_123", "123456")
	if err != nil {
		t.Fatalf("VerifyAuthCode: %v", err)
	}

	// Then get session
	session, err := om.GetSession(ctx, 12345)
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}

	if session == nil {
		t.Error("Expected non-nil session")
	}
	if session.AccessToken != "access_token_123" {
		t.Errorf("Expected access token 'access_token_123', got '%s'", session.AccessToken)
	}
}

func TestOAuthManager_Logout(t *testing.T) {
	db := setupOAuthSessionDB(t)
	defer func() { _ = db.Close() }()
	sessions := repository.NewSQLiteSessionRepository(db)
	om := NewOAuthManager(&fakeOAuthClient{}, sessions, zap.NewNop(), "http://localhost:3000")
	ctx := context.Background()

	// First verify auth code to create session
	err := om.VerifyAuthCode(ctx, 12345, "auth_token_123", "123456")
	if err != nil {
		t.Fatalf("VerifyAuthCode: %v", err)
	}

	// Then logout
	err = om.Logout(ctx, 12345)
	if err != nil {
		t.Fatalf("Logout: %v", err)
	}

	// Verify session was deleted
	_, err = sessions.GetSession(ctx, 12345)
	if err == nil {
		t.Error("Expected error when getting deleted session")
	}
}

func TestOAuthManager_ListSessions(t *testing.T) {
	db := setupOAuthSessionDB(t)
	defer func() { _ = db.Close() }()
	sessions := repository.NewSQLiteSessionRepository(db)
	om := NewOAuthManager(&fakeOAuthClient{}, sessions, zap.NewNop(), "http://localhost:3000")
	ctx := context.Background()

	sessionList, err := om.ListSessions(ctx, 12345)
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}

	if len(sessionList) == 0 {
		t.Error("Expected non-empty sessions list")
	}
	if sessionList[0].SessionId != "session_1" {
		t.Errorf("Expected session ID 'session_1', got '%s'", sessionList[0].SessionId)
	}
}

func TestOAuthManager_GetAuthLogs(t *testing.T) {
	db := setupOAuthSessionDB(t)
	defer func() { _ = db.Close() }()
	sessions := repository.NewSQLiteSessionRepository(db)
	om := NewOAuthManager(&fakeOAuthClient{}, sessions, zap.NewNop(), "http://localhost:3000")
	ctx := context.Background()

	logs, totalCount, err := om.GetAuthLogs(ctx, 12345, 10, 0)
	if err != nil {
		t.Fatalf("GetAuthLogs: %v", err)
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

func TestOAuthManager_CancelAuth(t *testing.T) {
	db := setupOAuthSessionDB(t)
	defer func() { _ = db.Close() }()
	sessions := repository.NewSQLiteSessionRepository(db)
	om := NewOAuthManager(&fakeOAuthClient{}, sessions, zap.NewNop(), "http://localhost:3000")
	ctx := context.Background()

	err := om.CancelAuth(ctx, 12345, "auth_token_123")
	if err != nil {
		t.Fatalf("CancelAuth: %v", err)
	}
}

func TestOAuthManager_GetAuthStatus(t *testing.T) {
	db := setupOAuthSessionDB(t)
	defer func() { _ = db.Close() }()
	sessions := repository.NewSQLiteSessionRepository(db)
	om := NewOAuthManager(&fakeOAuthClient{}, sessions, zap.NewNop(), "http://localhost:3000")
	ctx := context.Background()

	status, email, expiresAt, err := om.GetAuthStatus(ctx, "auth_token_123")
	if err != nil {
		t.Fatalf("GetAuthStatus: %v", err)
	}

	if status != "pending" {
		t.Errorf("Expected status 'pending', got '%s'", status)
	}
	if email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", email)
	}
	if expiresAt.IsZero() {
		t.Error("Expected non-zero expiresAt")
	}
}

func TestOAuthManager_RevokeSession(t *testing.T) {
	db := setupOAuthSessionDB(t)
	defer func() { _ = db.Close() }()
	sessions := repository.NewSQLiteSessionRepository(db)
	om := NewOAuthManager(&fakeOAuthClient{}, sessions, zap.NewNop(), "http://localhost:3000")
	ctx := context.Background()

	err := om.RevokeSession(ctx, 12345, "session_123")
	if err != nil {
		t.Fatalf("RevokeSession: %v", err)
	}
}
