package bot

import (
    "context"
    "testing"

    "budget-bot/internal/repository"
    _ "github.com/mattn/go-sqlite3"
)

func TestCategoryMatcher_ExactBeatsPartial(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    repo := repository.NewSQLiteCategoryMappingRepository(db)
    // exact and partial both exist; exact should win regardless of priority
    _ = repo.AddMapping(context.Background(), &repository.CategoryMapping{ID:"1", TenantID:"t1", Keyword:"такси", CategoryID:"cat-taxi", Priority:0})
    _ = repo.AddMapping(context.Background(), &repository.CategoryMapping{ID:"2", TenantID:"t1", Keyword:"так", CategoryID:"cat-partial", Priority:10})
    cm := NewCategoryMatcher(repo)
    m, err := cm.FindCategory(context.Background(), "t1", "вечернее такси домой")
    if err != nil || m == nil || m.CategoryID != "cat-taxi" { t.Fatalf("exact should win: %+v %v", m, err) }
}


