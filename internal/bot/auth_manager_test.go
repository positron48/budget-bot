package bot

import (
    "context"
    "database/sql"
    "testing"
    "time"

    "budget-bot/internal/repository"
    _ "github.com/mattn/go-sqlite3"
    "go.uber.org/zap"
)

type fakeAuth struct{}

func (f *fakeAuth) Register(ctx context.Context, email, password, name string) (string, string, string, string, time.Time, time.Time, error) {
    return "user-1", "tenant-1", "acc-1", "ref-1", time.Now().Add(time.Hour), time.Now().Add(24 * time.Hour), nil
}

func (f *fakeAuth) Login(ctx context.Context, email, password string) (string, string, string, string, time.Time, time.Time, error) {
    return "user-2", "tenant-2", "acc-2", "ref-2", time.Now().Add(time.Hour), time.Now().Add(24 * time.Hour), nil
}

func (f *fakeAuth) RefreshToken(ctx context.Context, refreshToken string) (string, string, time.Time, time.Time, error) {
    return "acc-3", "ref-3", time.Now().Add(2 * time.Hour), time.Now().Add(48 * time.Hour), nil
}

func setupSessionDB(t *testing.T) *sql.DB {
    t.Helper()
    db, err := sql.Open("sqlite3", ":memory:")
    if err != nil { t.Fatalf("open sqlite: %v", err) }
    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS user_sessions (
            telegram_id INTEGER PRIMARY KEY,
            user_id TEXT NOT NULL,
            tenant_id TEXT NOT NULL,
            access_token TEXT NOT NULL,
            refresh_token TEXT NOT NULL,
            access_token_expires_at TIMESTAMP NOT NULL,
            refresh_token_expires_at TIMESTAMP NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
    `)
    if err != nil { t.Fatalf("create table: %v", err) }
    return db
}

func TestAuthManager_Register_Login_Refresh(t *testing.T) {
    db := setupSessionDB(t)
    defer db.Close()
    sessions := repository.NewSQLiteSessionRepository(db)
    am := NewAuthManager(&fakeAuth{}, sessions, zap.NewNop())
    ctx := context.Background()
    tg := int64(1001)

    // Register
    if err := am.Register(ctx, tg, "a@b.c", "p", "n"); err != nil {
        t.Fatalf("register: %v", err)
    }
    s, err := sessions.GetSession(ctx, tg)
    if err != nil { t.Fatalf("get after register: %v", err) }
    if s.UserID != "user-1" || s.TenantID != "tenant-1" || s.AccessToken != "acc-1" {
        t.Fatalf("unexpected session after register: %+v", s)
    }

    // Login overwrites
    if err := am.Login(ctx, tg, "a@b.c", "p"); err != nil {
        t.Fatalf("login: %v", err)
    }
    s, err = sessions.GetSession(ctx, tg)
    if err != nil { t.Fatalf("get after login: %v", err) }
    if s.TenantID != "tenant-2" || s.AccessToken != "acc-2" {
        t.Fatalf("unexpected session after login: %+v", s)
    }

    // Refresh updates tokens
    if err := am.RefreshTokens(ctx, tg); err != nil {
        t.Fatalf("refresh: %v", err)
    }
    s, err = sessions.GetSession(ctx, tg)
    if err != nil { t.Fatalf("get after refresh: %v", err) }
    if s.AccessToken != "acc-3" || s.RefreshToken != "ref-3" {
        t.Fatalf("unexpected tokens after refresh: %+v", s)
    }
}


