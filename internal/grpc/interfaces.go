// Package grpc contains gRPC client facades used by the bot.
package grpc

import (
	"context"
	"time"

	pb "budget-bot/internal/pb/budget/v1"
)

// VerifyAuthCodeResult contains the result of OAuth verification
type VerifyAuthCodeResult struct {
	Tokens     *pb.TokenPair
	SessionID  string
	User       *pb.User
	Memberships []*pb.TenantMembership
}

// OAuthClient defines OAuth operations used by the bot.
type OAuthClient interface {
	GenerateAuthLink(ctx context.Context, email string, telegramUserID int64, userAgent, ipAddress string) (string, string, time.Time, error)
	VerifyAuthCode(ctx context.Context, authToken, verificationCode string, telegramUserID int64) (*VerifyAuthCodeResult, error)
	CancelAuth(ctx context.Context, authToken string, telegramUserID int64) error
	GetAuthStatus(ctx context.Context, authToken string) (string, string, time.Time, error)
	GetTelegramSession(ctx context.Context, sessionID string) (*pb.GetTelegramSessionResponse, error)
	RevokeTelegramSession(ctx context.Context, sessionID string, telegramUserID int64) error
	ListTelegramSessions(ctx context.Context, telegramUserID int64) ([]*pb.TelegramSession, error)
	GetAuthLogs(ctx context.Context, telegramUserID int64, limit, offset int32) ([]*pb.AuthLogEntry, int32, error)
}
