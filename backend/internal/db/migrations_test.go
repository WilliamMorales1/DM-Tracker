package db

import (
	"path/filepath"
	"testing"
)

func TestMigrationsCanResetAndReapply(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	migrationsDir := filepath.Join("..", "..", "migrations")

	if err := RunMigrations(migrationsDir, dbPath); err != nil {
		t.Fatalf("run migrations: %v", err)
	}
	assertMigrationVersion(t, dbPath, 1)

	if err := ResetMigrations(migrationsDir, dbPath); err != nil {
		t.Fatalf("reset migrations: %v", err)
	}
	assertMigrationVersion(t, dbPath, 1)
}

func assertMigrationVersion(t *testing.T, dbPath string, want int) {
	t.Helper()

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()

	var got int
	if err := db.QueryRow("SELECT version FROM schema_migrations").Scan(&got); err != nil {
		t.Fatalf("read migration version: %v", err)
	}
	if got != want {
		t.Errorf("migration version = %d, want %d", got, want)
	}
}
