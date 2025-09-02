package bot

import (
	"context"
	"strings"
	"testing"
	"time"

	"budget-bot/internal/repository"
	"go.uber.org/zap"
)

func TestOAuthManager_GetSession_WithAuthClient(t *testing.T) {
	db := setupOAuthSessionDB(t)
	defer func() { _ = db.Close() }()
	sessions := repository.NewSQLiteSessionRepository(db)
	
	// Create manager with auth client
	om := NewOAuthManagerWithAuthClient(&fakeOAuthClient{}, &fakeAuthClient{}, sessions, zap.NewNop(), "http://localhost:3000")
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
		return
	}
	if session.AccessToken != "access_token_123" {
		t.Errorf("Expected access token 'access_token_123', got '%s'", session.AccessToken)
	}
}

func TestOAuthManager_GetSession_ExpiredAccessToken_WithAuthClient(t *testing.T) {
	db := setupOAuthSessionDB(t)
	defer func() { _ = db.Close() }()
	sessions := repository.NewSQLiteSessionRepository(db)
	
	// Create manager with auth client
	om := NewOAuthManagerWithAuthClient(&fakeOAuthClient{}, &fakeAuthClient{}, sessions, zap.NewNop(), "http://localhost:3000")
	ctx := context.Background()

	// First verify auth code to create session
	err := om.VerifyAuthCode(ctx, 12345, "auth_token_123", "123456")
	if err != nil {
		t.Fatalf("VerifyAuthCode: %v", err)
	}

	// Manually update session to have expired access token but valid refresh token
	expiredAccessToken := time.Now().Add(-time.Hour) // Expired 1 hour ago
	validRefreshToken := time.Now().Add(24 * time.Hour) // Valid for 24 hours
	
	err = sessions.UpdateTokens(ctx, 12345, &repository.TokenPair{
		AccessToken:           "expired_access_token",
		RefreshToken:          "valid_refresh_token",
		AccessTokenExpiresAt:  expiredAccessToken,
		RefreshTokenExpiresAt: validRefreshToken,
	})
	if err != nil {
		t.Fatalf("Failed to update tokens: %v", err)
	}

	// Now get session - should trigger token refresh
	session, err := om.GetSession(ctx, 12345)
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}

	if session == nil {
		t.Error("Expected non-nil session")
		return
	}
	
	// Should have new tokens after refresh
	if session.AccessToken == "expired_access_token" {
		t.Error("Expected new access token after refresh")
	}
}

func TestOAuthManager_GetSession_ExpiredAccessToken_NoAuthClient(t *testing.T) {
	db := setupOAuthSessionDB(t)
	defer func() { _ = db.Close() }()
	sessions := repository.NewSQLiteSessionRepository(db)
	
	// Create manager without auth client
	om := NewOAuthManager(&fakeOAuthClient{}, sessions, zap.NewNop(), "http://localhost:3000")
	ctx := context.Background()

	// First verify auth code to create session
	err := om.VerifyAuthCode(ctx, 12345, "auth_token_123", "123456")
	if err != nil {
		t.Fatalf("VerifyAuthCode: %v", err)
	}

	// Manually update session to have expired access token
	expiredAccessToken := time.Now().Add(-time.Hour) // Expired 1 hour ago
	validRefreshToken := time.Now().Add(24 * time.Hour) // Valid for 24 hours
	
	err = sessions.UpdateTokens(ctx, 12345, &repository.TokenPair{
		AccessToken:           "expired_access_token",
		RefreshToken:          "valid_refresh_token",
		AccessTokenExpiresAt:  expiredAccessToken,
		RefreshTokenExpiresAt: validRefreshToken,
	})
	if err != nil {
		t.Fatalf("Failed to update tokens: %v", err)
	}

	// Now get session - should log warning but not panic
	session, err := om.GetSession(ctx, 12345)
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}

	if session == nil {
		t.Error("Expected non-nil session")
		return
	}
	
	// Should still have expired token since no auth client to refresh
	if session.AccessToken != "expired_access_token" {
		t.Error("Expected expired access token when no auth client available")
	}
}

func TestOAuthManager_GetSession_ExpiredRefreshToken_WithAuthClient(t *testing.T) {
	db := setupOAuthSessionDB(t)
	defer func() { _ = db.Close() }()
	sessions := repository.NewSQLiteSessionRepository(db)
	
	// Create manager with auth client
	om := NewOAuthManagerWithAuthClient(&fakeOAuthClient{}, &fakeAuthClient{}, sessions, zap.NewNop(), "http://localhost:3000")
	ctx := context.Background()

	// First verify auth code to create session
	err := om.VerifyAuthCode(ctx, 12345, "auth_token_123", "123456")
	if err != nil {
		t.Fatalf("VerifyAuthCode: %v", err)
	}

	// Manually update session to have both tokens expired
	expiredAccessToken := time.Now().Add(-time.Hour) // Expired 1 hour ago
	expiredRefreshToken := time.Now().Add(-time.Hour) // Expired 1 hour ago
	
	err = sessions.UpdateTokens(ctx, 12345, &repository.TokenPair{
		AccessToken:           "expired_access_token",
		RefreshToken:          "expired_refresh_token",
		AccessTokenExpiresAt:  expiredAccessToken,
		RefreshTokenExpiresAt: expiredRefreshToken,
	})
	if err != nil {
		t.Fatalf("Failed to update tokens: %v", err)
	}

	// Now get session - should fail because refresh token is expired
	_, err = om.GetSession(ctx, 12345)
	if err == nil {
		t.Error("Expected error when refresh token is expired")
	}
	
	expectedError := "refresh token expired"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error containing '%s', got '%s'", expectedError, err.Error())
	}
}

func TestOAuthManager_GetSession_ValidTokens_WithAuthClient(t *testing.T) {
	db := setupOAuthSessionDB(t)
	defer func() { _ = db.Close() }()
	sessions := repository.NewSQLiteSessionRepository(db)
	
	// Create manager with auth client
	om := NewOAuthManagerWithAuthClient(&fakeOAuthClient{}, &fakeAuthClient{}, sessions, zap.NewNop(), "http://localhost:3000")
	ctx := context.Background()

	// First verify auth code to create session
	err := om.VerifyAuthCode(ctx, 12345, "auth_token_123", "123456")
	if err != nil {
		t.Fatalf("VerifyAuthCode: %v", err)
	}

	// Manually update session to have valid tokens
	validAccessToken := time.Now().Add(time.Hour) // Valid for 1 hour
	validRefreshToken := time.Now().Add(24 * time.Hour) // Valid for 24 hours
	
	err = sessions.UpdateTokens(ctx, 12345, &repository.TokenPair{
		AccessToken:           "valid_access_token",
		RefreshToken:          "valid_refresh_token",
		AccessTokenExpiresAt:  validAccessToken,
		RefreshTokenExpiresAt: validRefreshToken,
	})
	if err != nil {
		t.Fatalf("Failed to update tokens: %v", err)
	}

	// Now get session - should return valid session without refresh
	session, err := om.GetSession(ctx, 12345)
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}

	if session == nil {
		t.Error("Expected non-nil session")
		return
	}
	
	// Should have the same valid tokens
	if session.AccessToken != "valid_access_token" {
		t.Error("Expected same valid access token")
	}
}

func TestOAuthManager_GetSession_RefreshFailure(t *testing.T) {
	db := setupOAuthSessionDB(t)
	defer func() { _ = db.Close() }()
	sessions := repository.NewSQLiteSessionRepository(db)
	
	// Create manager with auth client that will fail refresh
	failingAuthClient := &failingAuthClient{}
	om := NewOAuthManagerWithAuthClient(&fakeOAuthClient{}, failingAuthClient, sessions, zap.NewNop(), "http://localhost:3000")
	ctx := context.Background()

	// First verify auth code to create session
	err := om.VerifyAuthCode(ctx, 12345, "auth_token_123", "123456")
	if err != nil {
		t.Fatalf("VerifyAuthCode: %v", err)
	}

	// Manually update session to have expired access token but valid refresh token
	expiredAccessToken := time.Now().Add(-time.Hour) // Expired 1 hour ago
	validRefreshToken := time.Now().Add(24 * time.Hour) // Valid for 24 hours
	
	err = sessions.UpdateTokens(ctx, 12345, &repository.TokenPair{
		AccessToken:           "expired_access_token",
		RefreshToken:          "valid_refresh_token",
		AccessTokenExpiresAt:  expiredAccessToken,
		RefreshTokenExpiresAt: validRefreshToken,
	})
	if err != nil {
		t.Fatalf("Failed to update tokens: %v", err)
	}

	// Now get session - should fail during refresh
	_, err = om.GetSession(ctx, 12345)
	if err == nil {
		t.Error("Expected error when token refresh fails")
	}
}
