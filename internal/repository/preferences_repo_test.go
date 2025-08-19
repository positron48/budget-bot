package repository

import (
	"context"
	"testing"

	"budget-bot/internal/testutil"
)

func TestSQLitePreferencesRepository_CRUD(t *testing.T) {
	db := testutil.OpenMigratedSQLite(t)
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
