// Package grpc contains gRPC client facades used by the bot.
package grpc

import (
	"context"
	"time"

	pb "budget-bot/internal/pb/budget/v1"
	"go.uber.org/zap"
)

// AuthClient wraps the gRPC Auth service with helper methods.
type AuthClient struct {
	client pb.AuthServiceClient
	log    *zap.Logger
}

// NewAuthClient constructs an AuthClient.
func NewAuthClient(client pb.AuthServiceClient, log *zap.Logger) *AuthClient {
	return &AuthClient{client: client, log: log}
}

// Register registers a new user and returns ids and token expirations.
func (a *AuthClient) Register(ctx context.Context, email, password, name string) (string, string, string, string, time.Time, time.Time, error) {
	res, err := a.client.Register(ctx, &pb.RegisterRequest{Email: email, Password: password, Name: name})
	if err != nil {
		return "", "", "", "", time.Time{}, time.Time{}, err
	}
	userID := ""
	tenantID := ""
	if res.User != nil {
		userID = res.User.Id
	}
	if res.Tenant != nil {
		tenantID = res.Tenant.Id
	}
	tokens := res.Tokens
	return userID, tenantID, tokens.AccessToken, tokens.RefreshToken, tokens.AccessTokenExpiresAt.AsTime(), tokens.RefreshTokenExpiresAt.AsTime(), nil
}

// Login authenticates and returns ids and token expirations.
func (a *AuthClient) Login(ctx context.Context, email, password string) (string, string, string, string, time.Time, time.Time, error) {
	res, err := a.client.Login(ctx, &pb.LoginRequest{Email: email, Password: password})
	if err != nil {
		return "", "", "", "", time.Time{}, time.Time{}, err
	}
	tokens := res.Tokens
	
	// Логируем весь gRPC ответ, скрывая только сами токены
	a.log.Info("Full gRPC Login response",
		zap.Int("membershipsCount", len(res.Memberships)),
		zap.Any("memberships", res.Memberships),
		zap.String("tokens.accessToken", "[HIDDEN]"),
		zap.String("tokens.refreshToken", "[HIDDEN]"),
		zap.String("tokens.tokenType", tokens.TokenType),
		zap.Any("tokens.accessTokenExpiresAt", tokens.AccessTokenExpiresAt),
		zap.Any("tokens.refreshTokenExpiresAt", tokens.RefreshTokenExpiresAt))
	
	// Логируем время токенов от auth сервиса
	accessExp := tokens.AccessTokenExpiresAt.AsTime()
	refreshExp := tokens.RefreshTokenExpiresAt.AsTime()
	now := time.Now()
	
	a.log.Info("Auth service returned tokens", 
		zap.Time("accessTokenExpiresAt", accessExp),
		zap.Time("refreshTokenExpiresAt", refreshExp),
		zap.Time("now", now),
		zap.Duration("accessTokenTTL", accessExp.Sub(now)),
		zap.Duration("refreshTokenTTL", refreshExp.Sub(now)))
	
	tenantID := ""
	for _, m := range res.Memberships {
		if m.IsDefault {
			tenantID = m.Tenant.Id
			break
		}
	}
	if tenantID == "" && len(res.Memberships) > 0 {
		tenantID = res.Memberships[0].Tenant.Id
	}
	return "", tenantID, tokens.AccessToken, tokens.RefreshToken, accessExp, refreshExp, nil
}

// RefreshToken exchanges refresh token for a new access/refresh pair.
func (a *AuthClient) RefreshToken(ctx context.Context, refreshToken string) (string, string, time.Time, time.Time, error) {
	res, err := a.client.RefreshToken(ctx, &pb.RefreshTokenRequest{RefreshToken: refreshToken})
	if err != nil {
		return "", "", time.Time{}, time.Time{}, err
	}
	tokens := res.Tokens
	
	// Логируем весь gRPC ответ, скрывая только сами токены
	a.log.Info("Full gRPC RefreshToken response",
		zap.String("tokens.accessToken", "[HIDDEN]"),
		zap.String("tokens.refreshToken", "[HIDDEN]"),
		zap.String("tokens.tokenType", tokens.TokenType),
		zap.Any("tokens.accessTokenExpiresAt", tokens.AccessTokenExpiresAt),
		zap.Any("tokens.refreshTokenExpiresAt", tokens.RefreshTokenExpiresAt))
	
	// Логируем время токенов от auth сервиса
	accessExp := tokens.AccessTokenExpiresAt.AsTime()
	refreshExp := tokens.RefreshTokenExpiresAt.AsTime()
	now := time.Now()
	
	a.log.Info("Auth service refreshed tokens", 
		zap.Time("accessTokenExpiresAt", accessExp),
		zap.Time("refreshTokenExpiresAt", refreshExp),
		zap.Time("now", now),
		zap.Duration("accessTokenTTL", accessExp.Sub(now)),
		zap.Duration("refreshTokenTTL", refreshExp.Sub(now)))
	
	return tokens.AccessToken, tokens.RefreshToken, accessExp, refreshExp, nil
}


