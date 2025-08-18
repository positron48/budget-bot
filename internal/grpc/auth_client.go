package grpc

import (
	"context"
	"time"

	pb "budget-bot/internal/pb/budget/v1"
	"go.uber.org/zap"
)

type AuthClient struct {
	client pb.AuthServiceClient
	log    *zap.Logger
}

func NewAuthClient(client pb.AuthServiceClient, log *zap.Logger) *AuthClient {
	return &AuthClient{client: client, log: log}
}

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

func (a *AuthClient) Login(ctx context.Context, email, password string) (string, string, string, string, time.Time, time.Time, error) {
	res, err := a.client.Login(ctx, &pb.LoginRequest{Email: email, Password: password})
	if err != nil {
		return "", "", "", "", time.Time{}, time.Time{}, err
	}
	tokens := res.Tokens
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
	return "", tenantID, tokens.AccessToken, tokens.RefreshToken, tokens.AccessTokenExpiresAt.AsTime(), tokens.RefreshTokenExpiresAt.AsTime(), nil
}

func (a *AuthClient) RefreshToken(ctx context.Context, refreshToken string) (string, string, time.Time, time.Time, error) {
	res, err := a.client.RefreshToken(ctx, &pb.RefreshTokenRequest{RefreshToken: refreshToken})
	if err != nil {
		return "", "", time.Time{}, time.Time{}, err
	}
	tokens := res.Tokens
	return tokens.AccessToken, tokens.RefreshToken, tokens.AccessTokenExpiresAt.AsTime(), tokens.RefreshTokenExpiresAt.AsTime(), nil
}


