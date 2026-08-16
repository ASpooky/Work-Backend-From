package sqlite

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

// TestOpen_MigratesLegacyGoalsTableMissingPostponeCount simulates a
// pre-existing app.db created before postpone_count existed: CREATE TABLE
// IF NOT EXISTS alone can't add a column to an already-existing table, so
// Open() needs its own ALTER TABLE step for databases created by an older
// version of schema.sql.
func TestOpen_MigratesLegacyGoalsTableMissingPostponeCount(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.db")

	legacyDB, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("sql.Open() returned unexpected error: %v", err)
	}
	_, err = legacyDB.Exec(`CREATE TABLE goals (
		id TEXT PRIMARY KEY,
		workspace_id TEXT NOT NULL,
		title TEXT NOT NULL,
		detail TEXT NOT NULL,
		achievement_condition TEXT NOT NULL,
		end_date TEXT NOT NULL,
		mode TEXT NOT NULL,
		status TEXT NOT NULL,
		created_at TEXT NOT NULL
	)`)
	if err != nil {
		t.Fatalf("failed to create legacy goals table: %v", err)
	}
	if err := legacyDB.Close(); err != nil {
		t.Fatalf("failed to close legacy db: %v", err)
	}

	db, err := Open(path)
	if err != nil {
		t.Fatalf("Open() on a legacy database returned unexpected error: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if _, err := db.Exec(
		`INSERT INTO goals (id, workspace_id, title, detail, achievement_condition, end_date, mode, status, created_at, postpone_count)
		 VALUES ('g1', 'w1', 't', 'd', 'c', '2026-01-01', 'strict', 'active', '2026-01-01T00:00:00Z', 3)`,
	); err != nil {
		t.Fatalf("insert using postpone_count after migration returned unexpected error: %v", err)
	}

	var got int
	if err := db.QueryRow(`SELECT postpone_count FROM goals WHERE id = 'g1'`).Scan(&got); err != nil {
		t.Fatalf("failed to read back postpone_count: %v", err)
	}
	if got != 3 {
		t.Errorf("postpone_count = %d, want 3", got)
	}

	// Re-opening (idempotency) must not fail with "duplicate column name".
	db2, err := Open(path)
	if err != nil {
		t.Fatalf("second Open() on an already-migrated database returned unexpected error: %v", err)
	}
	db2.Close()
}
