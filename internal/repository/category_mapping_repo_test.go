package repository

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func openTempDBForMappings(t *testing.T) *sql.DB {
	t.Helper()
	f, err := os.CreateTemp("", "botdb-*.sqlite")
	if err != nil { t.Fatalf("temp file: %v", err) }
	_ = f.Close()
	dsn := "file:" + f.Name() + "?_foreign_keys=on"
	db, err := sql.Open("sqlite3", dsn)
	if err != nil { t.Fatalf("open db: %v", err) }
	_, err = db.Exec(`CREATE TABLE category_mappings (
		id TEXT PRIMARY KEY,
		tenant_id TEXT NOT NULL,
		keyword TEXT NOT NULL,
		category_id TEXT NOT NULL,
		priority INTEGER DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(tenant_id, keyword)
	);`)
	if err != nil { t.Fatalf("migrate: %v", err) }
	t.Cleanup(func(){ _ = os.Remove(f.Name()); _ = db.Close() })
	return db
}

func TestSQLiteCategoryMappingRepository_CRUD(t *testing.T) {
	db := openTempDBForMappings(t)
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
