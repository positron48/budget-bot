package db

import (
    "os"
    "path/filepath"
    "testing"

    _ "modernc.org/sqlite"
    "go.uber.org/zap"
)

func TestOpenAndMigrate_Works(t *testing.T) {
    tmp, err := os.CreateTemp("", "dbtest-*.sqlite")
    if err != nil { t.Fatalf("temp: %v", err) }
    _ = tmp.Close()
    t.Cleanup(func(){ _ = os.Remove(tmp.Name()) })

    // migrations dir: from internal/pkg/db to repo root => ../../../migrations
    migrations := filepath.Join("..", "..", "..", "migrations")
    if _, err := os.Stat(migrations); err != nil {
        t.Fatalf("migrations not found: %s: %v", migrations, err)
    }
    dsn := "file:" + tmp.Name() + "?_foreign_keys=on"
    log, _ := zap.NewDevelopment()
    dbc, err := OpenAndMigrate(dsn, migrations, log)
    if err != nil { t.Fatalf("OpenAndMigrate: %v", err) }
    defer dbc.Close()

    // ensure table exists
    var name string
    row := dbc.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='user_sessions'")
    if err := row.Scan(&name); err != nil { t.Fatalf("scan: %v", err) }
    if name != "user_sessions" { t.Fatalf("table missing: %s", name) }
}


