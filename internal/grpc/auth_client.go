package grpc

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// Placeholder interfaces for generated pb clients. Replace with actual pb imports when proto is added.
type pbAuthClient interface {
	Register(ctx context.Context, in *RegisterRequest) (*RegisterResponse, error)
	Login(ctx context.Context, in *LoginRequest) (*LoginResponse, error)
	RefreshToken(ctx context.Context, in *RefreshTokenRequest) (*RefreshTokenResponse, error)
}

type RegisterRequest struct{ Email, Password, Name string }
type RegisterResponse struct {
	UserId      string
	TenantId    string
	AccessToken string
	RefreshToken string
	AccessExpiresAt  int64
	RefreshExpiresAt int64
}

type LoginRequest struct{ Email, Password string }
type LoginResponse struct {
	UserId      string
	TenantId    string
	AccessToken string
	RefreshToken string
	AccessExpiresAt  int64
	RefreshExpiresAt int64
}

type RefreshTokenRequest struct{ RefreshToken string }
type RefreshTokenResponse struct {
	AccessToken string
	RefreshToken string
	AccessExpiresAt  int64
	RefreshExpiresAt int64
}

type AuthClient struct {
	client pbAuthClient
	log    *zap.Logger
}

func NewAuthClient(client pbAuthClient, log *zap.Logger) *AuthClient {
	return &AuthClient{client: client, log: log}
}

func (a *AuthClient) Register(ctx context.Context, email, password, name string) (string, string, string, string, time.Time, time.Time, error) {
	res, err := a.client.Register(ctx, &RegisterRequest{Email: email, Password: password, Name: name})
	if err != nil {
		return "", "", "", "", time.Time{}, time.Time{}, err
	}
	return res.UserId, res.TenantId, res.AccessToken, res.RefreshToken, time.Unix(res.AccessExpiresAt, 0), time.Unix(res.RefreshExpiresAt, 0), nil
}

func (a *AuthClient) Login(ctx context.Context, email, password string) (string, string, string, string, time.Time, time.Time, error) {
	res, err := a.client.Login(ctx, &LoginRequest{Email: email, Password: password})
	if err != nil {
		return "", "", "", "", time.Time{}, time.Time{}, err
	}
	return res.UserId, res.TenantId, res.AccessToken, res.RefreshToken, time.Unix(res.AccessExpiresAt, 0), time.Unix(res.RefreshExpiresAt, 0), nil
}

func (a *AuthClient) RefreshToken(ctx context.Context, refreshToken string) (string, string, time.Time, time.Time, error) {
	res, err := a.client.RefreshToken(ctx, &RefreshTokenRequest{RefreshToken: refreshToken})
	if err != nil {
		return "", "", time.Time{}, time.Time{}, err
	}
	return res.AccessToken, res.RefreshToken, time.Unix(res.AccessExpiresAt, 0), time.Unix(res.RefreshExpiresAt, 0), nil
}


