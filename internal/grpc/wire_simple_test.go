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
