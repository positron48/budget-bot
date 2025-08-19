package repository

import (
	"context"
	"testing"
	"time"

	"budget-bot/internal/testutil"
)

func TestSQLiteSessionRepository_CRUD(t *testing.T) {
	db := testutil.OpenMigratedSQLite(t)
	repo := NewSQLiteSessionRepository(db)
	ctx := context.Background()

	s := &UserSession{
		TelegramID:  123,
		UserID:      "user-1",
		TenantID:    "tenant-1",
		AccessToken: "a1",
		RefreshToken: "r1",
		AccessTokenExpiresAt:  time.Now().Add(1*time.Hour),
		RefreshTokenExpiresAt: time.Now().Add(24*time.Hour),
	}
	if err := repo.SaveSession(ctx, s); err != nil { t.Fatalf("save: %v", err) }

	got, err := repo.GetSession(ctx, 123)
	if err != nil { t.Fatalf("get: %v", err) }
	if got.UserID != s.UserID || got.TenantID != s.TenantID { t.Fatalf("mismatch: %+v", got) }

	if err := repo.UpdateTenantID(ctx, 123, "tenant-2"); err != nil { t.Fatalf("update tenant: %v", err) }
	got, _ = repo.GetSession(ctx, 123)
	if got.TenantID != "tenant-2" { t.Fatalf("tenant not updated: %+v", got) }

	if err := repo.UpdateTokens(ctx, 123, &TokenPair{AccessToken: "a2", RefreshToken: "r2", AccessTokenExpiresAt: time.Now().Add(2*time.Hour), RefreshTokenExpiresAt: time.Now().Add(48*time.Hour)}); err != nil { t.Fatalf("update tokens: %v", err) }
	got, _ = repo.GetSession(ctx, 123)
	if got.AccessToken != "a2" || got.RefreshToken != "r2" { t.Fatalf("tokens not updated: %+v", got) }

	if err := repo.DeleteSession(ctx, 123); err != nil { t.Fatalf("delete: %v", err) }
	if _, err := repo.GetSession(ctx, 123); err == nil { t.Fatalf("expected error after delete") }
}
