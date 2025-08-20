// Package testutil provides helpers for tests (SQLite DB, Telegram fakes, etc.).
package testutil

import (
    "database/sql"
    "os"
    "path/filepath"
    "testing"

    appdb "budget-bot/internal/pkg/db"
    "go.uber.org/zap"
)

// OpenMigratedSQLite opens a temporary SQLite DB file and runs real migrations from repo root.
// It registers a cleanup to remove the temp file and close the DB.
func OpenMigratedSQLite(t *testing.T) *sql.DB {
    t.Helper()
    tmp, err := os.CreateTemp("", "botdb-*.sqlite")
    if err != nil { t.Fatalf("temp file: %v", err) }
    _ = tmp.Close()

    root := findRepoRoot(t)
    migrations := filepath.Join(root, "migrations")
    dsn := "file:" + tmp.Name() + "?_foreign_keys=on"

    log, _ := zap.NewDevelopment()
    db, err := appdb.OpenAndMigrate(dsn, migrations, log)
    if err != nil { t.Fatalf("open and migrate: %v", err) }
    t.Cleanup(func(){ _ = db.Close(); _ = os.Remove(tmp.Name()) })
    return db
}

// findRepoRoot walks upwards to find a directory containing the migrations folder.
func findRepoRoot(t *testing.T) string {
    t.Helper()
    wd, err := os.Getwd()
    if err != nil { t.Fatalf("getwd: %v", err) }
    dir := wd
    for i := 0; i < 10; i++ {
        if _, err := os.Stat(filepath.Join(dir, "migrations")); err == nil {
            return dir
        }
        parent := filepath.Dir(dir)
        if parent == dir { break }
        dir = parent
    }
    t.Fatalf("migrations folder not found upwards from %s", wd)
    return ""
}


