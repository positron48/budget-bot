package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func openTempDialogDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil { t.Fatalf("open: %v", err) }
	_, err = db.Exec(`CREATE TABLE dialog_states (
		telegram_id INTEGER PRIMARY KEY,
		state TEXT NOT NULL,
		draft_id TEXT,
		context TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`)
	if err != nil { t.Fatalf("migrate: %v", err) }
	return db
}

func TestSQLiteDialogStateRepository_SetGetClear(t *testing.T) {
	db := openTempDialogDB(t)
	defer db.Close()
	repo := NewSQLiteDialogStateRepository(db)
	ctx := context.Background()
	ctxMap := map[string]any{"a": 1, "b": "x"}
	if err := repo.SetState(ctx, 100, StateWaitingForPassword, ctxMap, nil); err != nil { t.Fatalf("set: %v", err) }
	rec, err := repo.GetState(ctx, 100)
	if err != nil { t.Fatalf("get: %v", err) }
	if rec.State != StateWaitingForPassword { t.Fatalf("state mismatch: %s", rec.State) }
	// Ensure context round-trip
	if rec.Context == nil { t.Fatalf("nil context") }
	b, _ := json.Marshal(rec.Context)
	if len(b) == 0 { t.Fatalf("empty context json") }
	if err := repo.ClearState(ctx, 100); err != nil { t.Fatalf("clear: %v", err) }
	if _, err := repo.GetState(ctx, 100); err == nil { t.Fatalf("expected error after clear") }
}
