package db

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestOpenAndMigrate_ErrorScenarios(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name           string
		dsn            string
		migrationsDir  string
		expectError    bool
		errorContains  string
	}{
		{
			name:          "invalid dsn",
			dsn:           "invalid://dsn",
			migrationsDir: "/tmp/migrations",
			expectError:   true,
			errorContains: "ping sqlite",
		},
		{
			name:          "non-existent migrations directory",
			dsn:           ":memory:",
			migrationsDir: "/non/existent/path",
			expectError:   true,
			errorContains: "migrate init",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := OpenAndMigrate(tt.dsn, tt.migrationsDir, logger)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, db)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, db)
				if db != nil {
					db.Close()
				}
			}
		})
	}
}

func TestOpenAndMigrate_WithRealMigrations(t *testing.T) {
	logger := zap.NewNop()
	
	// Use the actual migrations directory from the project
	migrationsDir := "../../migrations"
	
	// Check if migrations directory exists
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		t.Skip("Migrations directory not found, skipping test")
	}

	// Test with memory database and real migrations
	db, err := OpenAndMigrate(":memory:", migrationsDir, logger)
	
	assert.NoError(t, err)
	assert.NotNil(t, db)
	if db != nil {
		db.Close()
	}
}

func TestOpenAndMigrate_WithMemoryDatabase(t *testing.T) {
	logger := zap.NewNop()
	
	// Create a temporary directory for migrations
	tempDir, err := os.MkdirTemp("", "migrations")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a simple migration file to avoid the "file does not exist" error
	migrationFile := tempDir + "/0001_test.up.sql"
	err = os.WriteFile(migrationFile, []byte("CREATE TABLE test (id INTEGER PRIMARY KEY);"), 0644)
	require.NoError(t, err)

	// Test with memory database and temp migrations directory
	db, err := OpenAndMigrate(":memory:", tempDir, logger)
	
	// Should succeed with valid migration file
	assert.NoError(t, err)
	assert.NotNil(t, db)
	if db != nil {
		db.Close()
	}
}
