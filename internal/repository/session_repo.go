package repository

import (
	"context"
	"database/sql"
	"time"
)

type UserSession struct {
	TelegramID              int64
	UserID                  string
	TenantID                string
	AccessToken             string
	RefreshToken            string
	AccessTokenExpiresAt    time.Time
	RefreshTokenExpiresAt   time.Time
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

type TokenPair struct {
	AccessToken           string
	RefreshToken          string
	AccessTokenExpiresAt  time.Time
	RefreshTokenExpiresAt time.Time
}

type SessionRepository interface {
	SaveSession(ctx context.Context, session *UserSession) error
	GetSession(ctx context.Context, telegramID int64) (*UserSession, error)
	DeleteSession(ctx context.Context, telegramID int64) error
	UpdateTokens(ctx context.Context, telegramID int64, tokens *TokenPair) error
	UpdateTenantID(ctx context.Context, telegramID int64, tenantID string) error
}

type SQLiteSessionRepository struct {
	db *sql.DB
}

func NewSQLiteSessionRepository(db *sql.DB) *SQLiteSessionRepository {
	return &SQLiteSessionRepository{db: db}
}

func (r *SQLiteSessionRepository) SaveSession(ctx context.Context, s *UserSession) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO user_sessions (
			telegram_id, user_id, tenant_id, access_token, refresh_token, access_token_expires_at, refresh_token_expires_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(telegram_id) DO UPDATE SET
			user_id = excluded.user_id,
			tenant_id = excluded.tenant_id,
			access_token = excluded.access_token,
			refresh_token = excluded.refresh_token,
			access_token_expires_at = excluded.access_token_expires_at,
			refresh_token_expires_at = excluded.refresh_token_expires_at,
			updated_at = CURRENT_TIMESTAMP
	`, s.TelegramID, s.UserID, s.TenantID, s.AccessToken, s.RefreshToken, s.AccessTokenExpiresAt, s.RefreshTokenExpiresAt)
	return err
}

func (r *SQLiteSessionRepository) GetSession(ctx context.Context, telegramID int64) (*UserSession, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT telegram_id, user_id, tenant_id, access_token, refresh_token, access_token_expires_at, refresh_token_expires_at, created_at, updated_at
		FROM user_sessions WHERE telegram_id = ?
	`, telegramID)
	var s UserSession
	if err := row.Scan(&s.TelegramID, &s.UserID, &s.TenantID, &s.AccessToken, &s.RefreshToken, &s.AccessTokenExpiresAt, &s.RefreshTokenExpiresAt, &s.CreatedAt, &s.UpdatedAt); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SQLiteSessionRepository) DeleteSession(ctx context.Context, telegramID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM user_sessions WHERE telegram_id = ?`, telegramID)
	return err
}

func (r *SQLiteSessionRepository) UpdateTokens(ctx context.Context, telegramID int64, t *TokenPair) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE user_sessions SET
			access_token = ?,
			refresh_token = ?,
			access_token_expires_at = ?,
			refresh_token_expires_at = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE telegram_id = ?
	`, t.AccessToken, t.RefreshToken, t.AccessTokenExpiresAt, t.RefreshTokenExpiresAt, telegramID)
	return err
}

func (r *SQLiteSessionRepository) UpdateTenantID(ctx context.Context, telegramID int64, tenantID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE user_sessions SET
			tenant_id = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE telegram_id = ?
	`, tenantID, telegramID)
	return err
}


