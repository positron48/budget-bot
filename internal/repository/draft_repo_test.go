package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func openTempDraftDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil { t.Fatalf("open: %v", err) }
	_, err = db.Exec(`CREATE TABLE transaction_drafts (
		id TEXT PRIMARY KEY,
		telegram_id INTEGER NOT NULL,
		type TEXT,
		amount_minor INTEGER,
		currency TEXT,
		description TEXT,
		category_id TEXT,
		occurred_at TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`)
	if err != nil { t.Fatalf("migrate: %v", err) }
	return db
}

func TestSQLiteDraftRepository_CRUD(t *testing.T) {
	db := openTempDraftDB(t)
	defer db.Close()
	repo := NewSQLiteDraftRepository(db)
	ctx := context.Background()
	id := "d1"
	now := time.Now()
	d := &TransactionDraft{ID: id, TelegramID: 5, Type: "expense", AmountMinor: 123, Currency: "RUB", Description: "t", CategoryID: "c", OccurredAt: &now}
	if err := repo.Create(ctx, d); err != nil { t.Fatalf("create: %v", err) }
	got, err := repo.Get(ctx, id)
	if err != nil { t.Fatalf("get: %v", err) }
	if got == nil || got.ID != id || got.AmountMinor != 123 { t.Fatalf("unexpected: %+v", got) }
	if err := repo.Delete(ctx, id); err != nil { t.Fatalf("delete: %v", err) }
	if _, err := repo.Get(ctx, id); err == nil { t.Fatalf("expected error after delete") }
}
