package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestFakeOAuthClient_GenerateAuthLink(t *testing.T) {
	client := &FakeOAuthClient{}
	ctx := context.Background()
	
	link, token, expires, err := client.GenerateAuthLink(ctx, "test@example.com", 12345, "TelegramBot/1.0", "127.0.0.1")
	
	assert.NoError(t, err)
	assert.Equal(t, "https://example.com/auth?token=test", link)
	assert.Equal(t, "auth_token_123", token)
	assert.True(t, expires.After(time.Now()))
}

func TestFakeOAuthClient_VerifyAuthCode(t *testing.T) {
	client := &FakeOAuthClient{}
	ctx := context.Background()
	
	result, err := client.VerifyAuthCode(ctx, "auth_token_123", "123456", 12345)
	
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "session_123", result.SessionID)
}

func TestFakeOAuthClient_CancelAuth(t *testing.T) {
	client := &FakeOAuthClient{}
	ctx := context.Background()
	
	err := client.CancelAuth(ctx, "auth_token_123", 12345)
	
	assert.NoError(t, err)
}

func TestFakeOAuthClient_GetAuthStatus(t *testing.T) {
	client := &FakeOAuthClient{}
	ctx := context.Background()
	
	status, email, expires, err := client.GetAuthStatus(ctx, "auth_token_123")
	
	assert.NoError(t, err)
	assert.Equal(t, "pending", status)
	assert.Equal(t, "test@example.com", email)
	assert.True(t, expires.After(time.Now()))
}

func TestFakeOAuthClient_GetTelegramSession(t *testing.T) {
	client := &FakeOAuthClient{}
	ctx := context.Background()
	
	result, err := client.GetTelegramSession(ctx, "session_123")
	
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Session)
	assert.Equal(t, "session_123", result.Session.SessionId)
}

func TestFakeOAuthClient_RevokeTelegramSession(t *testing.T) {
	client := &FakeOAuthClient{}
	ctx := context.Background()
	
	err := client.RevokeTelegramSession(ctx, "session_123", 12345)
	
	assert.NoError(t, err)
}

func TestFakeOAuthClient_ListTelegramSessions(t *testing.T) {
	client := &FakeOAuthClient{}
	ctx := context.Background()
	
	sessions, err := client.ListTelegramSessions(ctx, 12345)
	
	assert.NoError(t, err)
	assert.Len(t, sessions, 1)
	assert.Equal(t, "session_1", sessions[0].SessionId)
}

func TestFakeOAuthClient_GetAuthLogs(t *testing.T) {
	client := &FakeOAuthClient{}
	ctx := context.Background()
	
	logs, total, err := client.GetAuthLogs(ctx, 12345, 10, 0)
	
	assert.NoError(t, err)
	assert.Equal(t, int32(1), total)
	assert.Len(t, logs, 1)
	assert.Equal(t, "log_1", logs[0].Id)
}

func TestWireClients(t *testing.T) {
	logger := zap.NewNop()
	
	categoryClient, reportClient, tenantClient, transactionClient, oauthClient, authClient := WireClients(logger)
	
	assert.Nil(t, categoryClient)
	assert.Nil(t, reportClient)
	assert.Nil(t, tenantClient)
	assert.Nil(t, transactionClient)
	assert.NotNil(t, oauthClient)
	assert.IsType(t, &FakeOAuthClient{}, oauthClient)
	assert.NotNil(t, authClient)
	assert.IsType(t, &FakeAuthClient{}, authClient)
}

func TestFakeAuthClient_Register(t *testing.T) {
	client := &FakeAuthClient{}
	ctx := context.Background()
	
	userID, tenantID, accessToken, refreshToken, accessExp, refreshExp, err := client.Register(ctx, "test@example.com", "password123", "Test User")
	
	assert.NoError(t, err)
	assert.Equal(t, "user_123", userID)
	assert.Equal(t, "tenant_123", tenantID)
	assert.Equal(t, "access_token_123", accessToken)
	assert.Equal(t, "refresh_token_123", refreshToken)
	assert.True(t, accessExp.After(time.Now()))
	assert.True(t, refreshExp.After(time.Now()))
	
	// Проверяем, что время истечения access token примерно 15 минут
	expectedAccessExp := time.Now().Add(15 * time.Minute)
	assert.WithinDuration(t, expectedAccessExp, accessExp, 2*time.Second)
	
	// Проверяем, что время истечения refresh token примерно 720 часов (30 дней)
	expectedRefreshExp := time.Now().Add(720 * time.Hour)
	assert.WithinDuration(t, expectedRefreshExp, refreshExp, 2*time.Second)
}

func TestFakeAuthClient_Login(t *testing.T) {
	client := &FakeAuthClient{}
	ctx := context.Background()
	
	userID, tenantID, accessToken, refreshToken, accessExp, refreshExp, err := client.Login(ctx, "test@example.com", "password123")
	
	assert.NoError(t, err)
	assert.Equal(t, "user_123", userID)
	assert.Equal(t, "tenant_123", tenantID)
	assert.Equal(t, "access_token_123", accessToken)
	assert.Equal(t, "refresh_token_123", refreshToken)
	assert.True(t, accessExp.After(time.Now()))
	assert.True(t, refreshExp.After(time.Now()))
}

func TestFakeAuthClient_RefreshToken(t *testing.T) {
	client := &FakeAuthClient{}
	ctx := context.Background()
	
	accessToken, refreshTokenNew, accessExp, refreshExp, err := client.RefreshToken(ctx, "old_refresh_token")
	
	assert.NoError(t, err)
	assert.Equal(t, "new_access_token_123", accessToken)
	assert.Equal(t, "new_refresh_token_123", refreshTokenNew)
	assert.True(t, accessExp.After(time.Now()))
	assert.True(t, refreshExp.After(time.Now()))
}
