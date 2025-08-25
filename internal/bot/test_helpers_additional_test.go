package bot

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTestOAuthClient_CancelAuth(t *testing.T) {
	client := &TestOAuthClient{}
	ctx := context.Background()
	
	err := client.CancelAuth(ctx, "auth_token", 12345)
	
	assert.NoError(t, err)
}

func TestTestOAuthClient_GetAuthStatus(t *testing.T) {
	client := &TestOAuthClient{}
	ctx := context.Background()
	
	status, email, expires, err := client.GetAuthStatus(ctx, "auth_token")
	
	assert.NoError(t, err)
	assert.Equal(t, "completed", status)
	assert.Equal(t, "test@example.com", email)
	assert.True(t, expires.After(time.Now()))
}

func TestTestOAuthClient_GetTelegramSession(t *testing.T) {
	client := &TestOAuthClient{}
	ctx := context.Background()
	
	result, err := client.GetTelegramSession(ctx, "session_123")
	
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Session)
	assert.Equal(t, "session_123", result.Session.SessionId)
	assert.Equal(t, "user_id", result.Session.UserId)
	assert.Equal(t, "12345", result.Session.TelegramUserId)
	assert.Equal(t, "tenant_id", result.Session.TenantId)
	assert.True(t, result.Session.IsActive)
}

func TestTestOAuthClient_RevokeTelegramSession(t *testing.T) {
	client := &TestOAuthClient{}
	ctx := context.Background()
	
	err := client.RevokeTelegramSession(ctx, "session_123", 12345)
	
	assert.NoError(t, err)
}

func TestTestOAuthClient_ListTelegramSessions(t *testing.T) {
	client := &TestOAuthClient{}
	ctx := context.Background()
	
	sessions, err := client.ListTelegramSessions(ctx, 12345)
	
	assert.NoError(t, err)
	assert.Empty(t, sessions)
}

func TestTestOAuthClient_GetAuthLogs(t *testing.T) {
	client := &TestOAuthClient{}
	ctx := context.Background()
	
	logs, total, err := client.GetAuthLogs(ctx, 12345, 10, 0)
	
	assert.NoError(t, err)
	assert.Equal(t, int32(0), total)
	assert.Empty(t, logs)
}
