// Package grpc contains gRPC client facades used by the bot.
package grpc

import (
	"context"
	"fmt"
	"time"

	pb "budget-bot/internal/pb/budget/v1"
	"go.uber.org/zap"
)

// OAuthGRPCClient wraps the gRPC OAuth service with helper methods.
type OAuthGRPCClient struct {
	client pb.OAuthServiceClient
	log    *zap.Logger
}

// NewOAuthClient constructs an OAuthGRPCClient.
func NewOAuthClient(client pb.OAuthServiceClient, log *zap.Logger) OAuthClient {
	return &OAuthGRPCClient{client: client, log: log}
}

// GenerateAuthLink generates an OAuth authorization link.
func (o *OAuthGRPCClient) GenerateAuthLink(ctx context.Context, email string, telegramUserID int64, userAgent, ipAddress string) (string, string, time.Time, error) {
	res, err := o.client.GenerateAuthLink(ctx, &pb.GenerateAuthLinkRequest{
		Email:          email,
		TelegramUserId: fmt.Sprintf("%d", telegramUserID),
		UserAgent:      userAgent,
		IpAddress:      ipAddress,
	})
	if err != nil {
		return "", "", time.Time{}, err
	}
	return res.AuthUrl, res.AuthToken, res.ExpiresAt.AsTime(), nil
}

// VerifyAuthCode verifies the OAuth verification code.
func (o *OAuthGRPCClient) VerifyAuthCode(ctx context.Context, authToken, verificationCode string, telegramUserID int64) (*VerifyAuthCodeResult, error) {
	o.log.Info("Sending VerifyAuthCode request to gRPC",
		zap.String("authToken", authToken),
		zap.String("verificationCode", verificationCode),
		zap.Int64("telegramUserID", telegramUserID))

	res, err := o.client.VerifyAuthCode(ctx, &pb.VerifyAuthCodeRequest{
		AuthToken:        authToken,
		VerificationCode: verificationCode,
		TelegramUserId:   fmt.Sprintf("%d", telegramUserID),
	})
	if err != nil {
		o.log.Error("gRPC VerifyAuthCode failed",
			zap.String("authToken", authToken),
			zap.String("verificationCode", verificationCode),
			zap.Int64("telegramUserID", telegramUserID),
			zap.Error(err))
		return nil, err
	}

	accessTokenLog := ""
	refreshTokenLog := ""
	if res.Tokens != nil {
		if len(res.Tokens.AccessToken) > 10 {
			accessTokenLog = res.Tokens.AccessToken[:10] + "..."
		} else {
			accessTokenLog = res.Tokens.AccessToken
		}
		if len(res.Tokens.RefreshToken) > 10 {
			refreshTokenLog = res.Tokens.RefreshToken[:10] + "..."
		} else {
			refreshTokenLog = res.Tokens.RefreshToken
		}
	}
	
	o.log.Info("gRPC VerifyAuthCode succeeded",
		zap.String("sessionID", res.SessionId),
		zap.String("accessToken", accessTokenLog),
		zap.String("refreshToken", refreshTokenLog),
		zap.String("userID", res.User.Id),
		zap.Int("membershipsCount", len(res.Memberships)))

	result := &VerifyAuthCodeResult{
		Tokens:      res.Tokens,
		SessionID:   res.SessionId,
		User:        res.User,
		Memberships: res.Memberships,
	}

	return result, nil
}

// CancelAuth cancels the OAuth authorization process.
func (o *OAuthGRPCClient) CancelAuth(ctx context.Context, authToken string, telegramUserID int64) error {
	_, err := o.client.CancelAuth(ctx, &pb.CancelAuthRequest{
		AuthToken:      authToken,
		TelegramUserId: fmt.Sprintf("%d", telegramUserID),
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
	return res.Status.String(), res.Email, res.ExpiresAt.AsTime(), nil
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
		TelegramUserId: fmt.Sprintf("%d", telegramUserID),
	})
	return err
}

// ListTelegramSessions lists all Telegram sessions for a user.
func (o *OAuthGRPCClient) ListTelegramSessions(ctx context.Context, telegramUserID int64) ([]*pb.TelegramSession, error) {
	res, err := o.client.ListTelegramSessions(ctx, &pb.ListTelegramSessionsRequest{
		TelegramUserId: fmt.Sprintf("%d", telegramUserID),
	})
	if err != nil {
		return nil, err
	}
	return res.Sessions, nil
}

// GetAuthLogs gets authentication logs for a user.
func (o *OAuthGRPCClient) GetAuthLogs(ctx context.Context, telegramUserID int64, limit, offset int32) ([]*pb.AuthLogEntry, int32, error) {
	res, err := o.client.GetAuthLogs(ctx, &pb.GetAuthLogsRequest{
		TelegramUserId: fmt.Sprintf("%d", telegramUserID),
		Limit:          limit,
		Offset:         offset,
	})
	if err != nil {
		return nil, 0, err
	}
	return res.Logs, res.TotalCount, nil
}


