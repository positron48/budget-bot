package repository

import (
	"context"
	"encoding/json"
	"testing"

	"budget-bot/internal/testutil"
)

func TestSQLiteDialogStateRepository_SetGetClear(t *testing.T) {
	db := testutil.OpenMigratedSQLite(t)
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
