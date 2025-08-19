package repository

import (
	"context"
	"testing"

	"budget-bot/internal/testutil"
)

func TestSQLiteCategoryMappingRepository_CRUD(t *testing.T) {
	db := testutil.OpenMigratedSQLite(t)
	repo := NewSQLiteCategoryMappingRepository(db)
	ctx := context.Background()

	m := &CategoryMapping{ID: "id1", TenantID: "t1", Keyword: "кофе", CategoryID: "cat-food", Priority: 1}
	if err := repo.AddMapping(ctx, m); err != nil { t.Fatalf("add: %v", err) }

	got, err := repo.FindMapping(ctx, "t1", "кофе")
	if err != nil || got == nil { t.Fatalf("find: %v %v", got, err) }
	if got.CategoryID != "cat-food" { t.Fatalf("unexpected: %+v", got) }

	m2 := &CategoryMapping{ID: "id1", TenantID: "t1", Keyword: "кофе", CategoryID: "cat-drinks", Priority: 2}
	if err := repo.AddMapping(ctx, m2); err != nil { t.Fatalf("update via upsert: %v", err) }
	got, _ = repo.FindMapping(ctx, "t1", "кофе")
	if got.CategoryID != "cat-drinks" || got.Priority != 2 { t.Fatalf("not updated: %+v", got) }

	list, err := repo.ListMappings(ctx, "t1")
	if err != nil || len(list) == 0 { t.Fatalf("list: %v %v", len(list), err) }

	if err := repo.RemoveMapping(ctx, "t1", "кофе"); err != nil { t.Fatalf("remove: %v", err) }
	if _, err := repo.FindMapping(ctx, "t1", "кофе"); err == nil { t.Fatalf("expected error after delete") }
}
