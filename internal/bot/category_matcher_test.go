package bot

import (
	"context"
	"database/sql"
	"testing"

	"budget-bot/internal/repository"
	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil { t.Fatalf("open db: %v", err) }
	// create required tables
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS category_mappings (
			id TEXT PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			keyword TEXT NOT NULL,
			category_id TEXT NOT NULL,
			priority INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(tenant_id, keyword)
		);
	`)
	if err != nil { t.Fatalf("create table: %v", err) }
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
	if err != nil { t.Fatalf("create dialog_states: %v", err) }
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
	if err != nil { t.Fatalf("create user_sessions: %v", err) }
	return db
}

func TestCategoryMatcher_ExactMatch(t *testing.T) {
	db := setupTestDB(t)
	defer func(){ _ = db.Close() }()
	repo := repository.NewSQLiteCategoryMappingRepository(db)
	cm := NewCategoryMatcher(repo)
	ctx := context.Background()
	// seed mapping
	_ = repo.AddMapping(ctx, &repository.CategoryMapping{ID: "1", TenantID: "t1", Keyword: "кофе", CategoryID: "cat-coffee", Priority: 1})

	m, err := cm.FindCategory(ctx, "t1", "утренний кофе")
	if err != nil { t.Fatalf("find: %v", err) }
	if m == nil || m.CategoryID != "cat-coffee" {
		t.Fatalf("expected cat-coffee, got %+v", m)
	}
}

func TestCategoryMatcher_PartialMatchWithPriority(t *testing.T) {
	db := setupTestDB(t)
	defer func(){ _ = db.Close() }()
	repo := repository.NewSQLiteCategoryMappingRepository(db)
	cm := NewCategoryMatcher(repo)
	ctx := context.Background()
	// seed mappings with different priorities, both partial only
	_ = repo.AddMapping(ctx, &repository.CategoryMapping{ID: "1", TenantID: "t1", Keyword: "так", CategoryID: "cat-tak", Priority: 1})
	_ = repo.AddMapping(ctx, &repository.CategoryMapping{ID: "2", TenantID: "t1", Keyword: "работ", CategoryID: "cat-work", Priority: 2})

	m, err := cm.FindCategory(ctx, "t1", "утреннее такси до работы")
	if err != nil { t.Fatalf("find: %v", err) }
	if m == nil || m.CategoryID != "cat-work" {
		t.Fatalf("expected cat-work (higher priority among partials), got %+v", m)
	}
}

func TestCategoryMatcher_NoMatch(t *testing.T) {
	db := setupTestDB(t)
	defer func(){ _ = db.Close() }()
	repo := repository.NewSQLiteCategoryMappingRepository(db)
	cm := NewCategoryMatcher(repo)
	m, err := cm.FindCategory(context.Background(), "t1", "без совпадений")
	if err != nil { t.Fatalf("find: %v", err) }
	if m != nil { t.Fatalf("expected nil, got %+v", m) }
}


