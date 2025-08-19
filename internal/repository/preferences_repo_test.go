package repository

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func openTempPrefsDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil { t.Fatalf("open: %v", err) }
	_, err = db.Exec(`CREATE TABLE user_preferences (
		telegram_id INTEGER PRIMARY KEY,
		language TEXT DEFAULT 'ru',
		default_currency TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`)
	if err != nil { t.Fatalf("migrate: %v", err) }
	return db
}

func TestSQLitePreferencesRepository_CRUD(t *testing.T) {
	db := openTempPrefsDB(t)
	defer db.Close()
	repo := NewSQLitePreferencesRepository(db)
	ctx := context.Background()
	p := &UserPreferences{TelegramID: 77, Language: "en", DefaultCurrency: "USD"}
	if err := repo.SavePreferences(ctx, p); err != nil { t.Fatalf("save: %v", err) }
	got, err := repo.GetPreferences(ctx, 77)
	if err != nil { t.Fatalf("get: %v", err) }
	if got.Language != "en" || got.DefaultCurrency != "USD" { t.Fatalf("unexpected: %+v", got) }
	if err := repo.UpdateLanguage(ctx, 77, "ru"); err != nil { t.Fatalf("upd lang: %v", err) }
	if err := repo.UpdateDefaultCurrency(ctx, 77, "RUB"); err != nil { t.Fatalf("upd cur: %v", err) }
	got, _ = repo.GetPreferences(ctx, 77)
	if got.Language != "ru" || got.DefaultCurrency != "RUB" { t.Fatalf("unexpected after upd: %+v", got) }
}
