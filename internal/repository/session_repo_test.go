package repository

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func openTempDB(t *testing.T) *sql.DB {
	t.Helper()
	f, err := os.CreateTemp("", "botdb-*.sqlite")
	if err != nil { t.Fatalf("temp file: %v", err) }
	_ = f.Close()
	dsn := "file:" + f.Name() + "?_foreign_keys=on"
	db, err := sql.Open("sqlite3", dsn)
	if err != nil { t.Fatalf("open db: %v", err) }
	// minimal schema
	_, err = db.Exec(`CREATE TABLE user_sessions (
		telegram_id INTEGER PRIMARY KEY,
		user_id TEXT NOT NULL,
		tenant_id TEXT NOT NULL,
		access_token TEXT NOT NULL,
		refresh_token TEXT NOT NULL,
		access_token_expires_at TIMESTAMP NOT NULL,
		refresh_token_expires_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`)
	if err != nil { t.Fatalf("migrate: %v", err) }
	t.Cleanup(func(){ _ = os.Remove(f.Name()); _ = db.Close() })
	return db
}

func TestSQLiteSessionRepository_CRUD(t *testing.T) {
	db := openTempDB(t)
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
