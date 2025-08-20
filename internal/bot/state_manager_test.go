package bot

import (
    "context"
    "database/sql"
    "testing"

    "budget-bot/internal/repository"
    _ "modernc.org/sqlite"
    "go.uber.org/zap"
)

func setupDialogStateDB(t *testing.T) *sql.DB {
    t.Helper()
    db, err := sql.Open("sqlite", ":memory:")
    if err != nil { t.Fatalf("open sqlite: %v", err) }
    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS dialog_states (
            telegram_id INTEGER PRIMARY KEY,
            state TEXT NOT NULL,
            draft_id TEXT,
            context TEXT,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
    `)
    if err != nil { t.Fatalf("create table: %v", err) }
    return db
}

func TestStateManager_SetGetClear(t *testing.T) {
    db := setupDialogStateDB(t)
    defer func(){ _ = db.Close() }()
    repo := repository.NewSQLiteDialogStateRepository(db)
    sm := NewStateManager(repo, zap.NewNop())
    ctx := context.Background()

    tg := int64(123)
    st := repository.StateWaitingForEmail
    ctxMap := map[string]any{"foo": "bar"}

    if err := sm.SetState(ctx, tg, st, ctxMap); err != nil {
        t.Fatalf("set state: %v", err)
    }
    rec, err := sm.GetState(ctx, tg)
    if err != nil { t.Fatalf("get state: %v", err) }
    if rec.State != st { t.Fatalf("expected %s, got %s", st, rec.State) }
    if rec.Context == nil || rec.Context["foo"].(string) != "bar" {
        t.Fatalf("expected context foo=bar, got %+v", rec.Context)
    }

    if err := sm.ClearState(ctx, tg); err != nil { t.Fatalf("clear: %v", err) }
    if _, err := sm.GetState(ctx, tg); err == nil {
        t.Fatalf("expected error after clear")
    }
}


