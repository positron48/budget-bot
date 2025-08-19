package repository

import (
	"context"
	"testing"
	"time"

	"budget-bot/internal/testutil"
)

func TestSQLiteDraftRepository_CRUD(t *testing.T) {
	db := testutil.OpenMigratedSQLite(t)
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
