// Package grpc contains gRPC client facades used by the bot.
package grpc

import (
	"context"
	"time"

	pb "budget-bot/internal/pb/budget/v1"
	"go.uber.org/zap"
)

// OAuthGRPCClient wraps the gRPC OAuth service with helper methods.
type OAuthGRPCClient struct {
	client pb.AuthServiceClient
	log    *zap.Logger
}

// NewOAuthClient constructs an OAuthGRPCClient.
func NewOAuthClient(client pb.AuthServiceClient, log *zap.Logger) OAuthClient {
	return &OAuthGRPCClient{client: client, log: log}
}

// GenerateAuthLink generates an OAuth authorization link.
func (o *OAuthGRPCClient) GenerateAuthLink(ctx context.Context, email string, telegramUserID int64, userAgent, ipAddress string) (string, string, time.Time, error) {
	res, err := o.client.GenerateAuthLink(ctx, &pb.GenerateAuthLinkRequest{
		Email:          email,
		TelegramUserId: telegramUserID,
		UserAgent:      userAgent,
		IpAddress:      ipAddress,
	})
	if err != nil {
		return "", "", time.Time{}, err
	}
	return res.AuthUrl, res.AuthToken, res.ExpiresAt.AsTime(), nil
}

// VerifyAuthCode verifies the OAuth verification code.
func (o *OAuthGRPCClient) VerifyAuthCode(ctx context.Context, authToken, verificationCode string, telegramUserID int64) (*pb.TokenPair, string, error) {
	res, err := o.client.VerifyAuthCode(ctx, &pb.VerifyAuthCodeRequest{
		AuthToken:        authToken,
		VerificationCode: verificationCode,
		TelegramUserId:   telegramUserID,
	})
	if err != nil {
		return nil, "", err
	}
	return res.Tokens, res.SessionId, nil
}

// CancelAuth cancels the OAuth authorization process.
func (o *OAuthGRPCClient) CancelAuth(ctx context.Context, authToken string, telegramUserID int64) error {
	_, err := o.client.CancelAuth(ctx, &pb.CancelAuthRequest{
		AuthToken:      authToken,
		TelegramUserId: telegramUserID,
	})
	return err
}

// GetAuthStatus gets the status of an OAuth authorization.
func (o *OAuthGRPCClient) GetAuthStatus(ctx context.Context, authToken string) (string, string, time.Time, error) {
	res, err := o.client.GetAuthStatus(ctx, &pb.GetAuthStatusRequest{
		AuthToken: authToken,
	})
	if err != nil {
		return "", "", time.Time{}, err
	}
	return res.Status, res.Email, res.ExpiresAt.AsTime(), nil
}

// GetTelegramSession gets a Telegram session by session ID.
func (o *OAuthGRPCClient) GetTelegramSession(ctx context.Context, sessionID string) (*pb.GetTelegramSessionResponse, error) {
	return o.client.GetTelegramSession(ctx, &pb.GetTelegramSessionRequest{
		SessionId: sessionID,
	})
}

// RevokeTelegramSession revokes a Telegram session.
func (o *OAuthGRPCClient) RevokeTelegramSession(ctx context.Context, sessionID string, telegramUserID int64) error {
	_, err := o.client.RevokeTelegramSession(ctx, &pb.RevokeTelegramSessionRequest{
		SessionId:      sessionID,
		TelegramUserId: telegramUserID,
	})
	return err
}

// ListTelegramSessions lists all Telegram sessions for a user.
func (o *OAuthGRPCClient) ListTelegramSessions(ctx context.Context, telegramUserID int64) ([]*pb.GetTelegramSessionResponse, error) {
	res, err := o.client.ListTelegramSessions(ctx, &pb.ListTelegramSessionsRequest{
		TelegramUserId: telegramUserID,
	})
	if err != nil {
		return nil, err
	}
	return res.Sessions, nil
}

// GetAuthLogs gets authentication logs for a user.
func (o *OAuthGRPCClient) GetAuthLogs(ctx context.Context, telegramUserID int64, limit, offset int32) ([]*pb.AuthLogEntry, int32, error) {
	res, err := o.client.GetAuthLogs(ctx, &pb.GetAuthLogsRequest{
		TelegramUserId: telegramUserID,
		Limit:          limit,
		Offset:         offset,
	})
	if err != nil {
		return nil, 0, err
	}
	return res.Logs, res.TotalCount, nil
}

// RefreshToken exchanges refresh token for a new access/refresh pair.
func (o *OAuthGRPCClient) RefreshToken(ctx context.Context, refreshToken string) (string, string, time.Time, time.Time, error) {
	res, err := o.client.RefreshToken(ctx, &pb.RefreshTokenRequest{RefreshToken: refreshToken})
	if err != nil {
		return "", "", time.Time{}, time.Time{}, err
	}
	tokens := res.Tokens
	return tokens.AccessToken, tokens.RefreshToken, tokens.AccessTokenExpiresAt.AsTime(), tokens.RefreshTokenExpiresAt.AsTime(), nil
}
