package repository

import (
	"context"
	"testing"

	"budget-bot/internal/testutil"
)

func TestOperationContextRepository_CreateGetUpdate(t *testing.T) {
	db := testutil.OpenMigratedSQLite(t)
	r := NewSQLiteOperationContextRepository(db)
	ctx := context.Background()

	op := &OperationContext{
		OpID:                "op-1",
		TelegramID:          1,
		TenantID:            "tenant-1",
		DescriptionOriginal: "кофе",
		SelectionSource:     "manual",
		TxType:              "expense",
		AmountMinor:         100,
		Currency:            "RUB",
	}
	if err := r.Create(ctx, op); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := r.UpdateSelection(ctx, "op-1", "cat-1", "Еда", "llm"); err != nil {
		t.Fatalf("update selection: %v", err)
	}
	if err := r.SetTransactionID(ctx, "op-1", "tx-1"); err != nil {
		t.Fatalf("set tx id: %v", err)
	}
	got, err := r.Get(ctx, "op-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.SelectionSource != "llm" || got.TransactionID == nil || *got.TransactionID != "tx-1" {
		t.Fatalf("unexpected context: %+v", got)
	}
}
