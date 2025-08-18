package repository

import (
	"context"
	"database/sql"
)

type UserPreferences struct {
	TelegramID      int64
	Language        string
	DefaultCurrency string
}

type PreferencesRepository interface {
	SavePreferences(ctx context.Context, preferences *UserPreferences) error
	GetPreferences(ctx context.Context, telegramID int64) (*UserPreferences, error)
	UpdateLanguage(ctx context.Context, telegramID int64, language string) error
	UpdateDefaultCurrency(ctx context.Context, telegramID int64, currency string) error
}

type SQLitePreferencesRepository struct {
	db *sql.DB
}

func NewSQLitePreferencesRepository(db *sql.DB) *SQLitePreferencesRepository {
	return &SQLitePreferencesRepository{db: db}
}

func (r *SQLitePreferencesRepository) SavePreferences(ctx context.Context, p *UserPreferences) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO user_preferences (telegram_id, language, default_currency)
		VALUES (?, ?, ?)
		ON CONFLICT(telegram_id) DO UPDATE SET
			language = excluded.language,
			default_currency = excluded.default_currency
	`, p.TelegramID, p.Language, p.DefaultCurrency)
	return err
}

func (r *SQLitePreferencesRepository) GetPreferences(ctx context.Context, telegramID int64) (*UserPreferences, error) {
	row := r.db.QueryRowContext(ctx, `SELECT telegram_id, language, default_currency FROM user_preferences WHERE telegram_id = ?`, telegramID)
	var p UserPreferences
	if err := row.Scan(&p.TelegramID, &p.Language, &p.DefaultCurrency); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *SQLitePreferencesRepository) UpdateLanguage(ctx context.Context, telegramID int64, language string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE user_preferences SET language = ? WHERE telegram_id = ?`, language, telegramID)
	return err
}

func (r *SQLitePreferencesRepository) UpdateDefaultCurrency(ctx context.Context, telegramID int64, currency string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE user_preferences SET default_currency = ? WHERE telegram_id = ?`, currency, telegramID)
	return err
}


