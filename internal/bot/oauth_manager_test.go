package bot

import (
	"context"
	"database/sql"
	"testing"
	"time"

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

func (f *fakeOAuthClient) VerifyAuthCode(ctx context.Context, authToken, verificationCode string, telegramUserID int64) (*pb.TokenPair, string, error) {
	return &pb.TokenPair{
		AccessToken:           "access_token_123",
		RefreshToken:          "refresh_token_123",
		AccessTokenExpiresAt:  timestamppb.New(time.Now().Add(time.Hour)),
		RefreshTokenExpiresAt: timestamppb.New(time.Now().Add(24*time.Hour)),
		TokenType:             "Bearer",
	}, "session_123", nil
}

func (f *fakeOAuthClient) CancelAuth(ctx context.Context, authToken string, telegramUserID int64) error {
	return nil
}

func (f *fakeOAuthClient) GetAuthStatus(ctx context.Context, authToken string) (string, string, time.Time, error) {
	return "pending", "test@example.com", time.Now().Add(5*time.Minute), nil
}

func (f *fakeOAuthClient) GetTelegramSession(ctx context.Context, sessionID string) (*pb.GetTelegramSessionResponse, error) {
	return &pb.GetTelegramSessionResponse{
		SessionId:        sessionID,
		UserId:           "user_123",
		TenantId:         "tenant_123",
		AccessTokenHash:  "hash_123",
		RefreshTokenHash: "hash_456",
		CreatedAt:        timestamppb.New(time.Now()),
		ExpiresAt:        timestamppb.New(time.Now().Add(24*time.Hour)),
		IsActive:         true,
	}, nil
}

func (f *fakeOAuthClient) RevokeTelegramSession(ctx context.Context, sessionID string, telegramUserID int64) error {
	return nil
}

func (f *fakeOAuthClient) ListTelegramSessions(ctx context.Context, telegramUserID int64) ([]*pb.GetTelegramSessionResponse, error) {
	return []*pb.GetTelegramSessionResponse{
		{
			SessionId:        "session_1",
			UserId:           "user_123",
			TenantId:         "tenant_123",
			AccessTokenHash:  "hash_123",
			RefreshTokenHash: "hash_456",
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
			TelegramUserId: telegramUserID,
			Email:          "test@example.com",
			Action:         "generate_link",
			Status:         "success",
			IpAddress:      "127.0.0.1",
			UserAgent:      "TelegramBot/1.0",
			CreatedAt:      timestamppb.New(time.Now()),
		},
	}, 1, nil
}

func (f *fakeOAuthClient) RefreshToken(ctx context.Context, refreshToken string) (string, string, time.Time, time.Time, error) {
	return "access_token_new", "refresh_token_new", time.Now().Add(time.Hour), time.Now().Add(24*time.Hour), nil
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
