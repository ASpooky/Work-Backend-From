package sqlite

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"sync"
	"testing"

	_ "modernc.org/sqlite"
)

// TestOpen_SurvivesConcurrentWrites reproduces a real bug: the frontend's
// goal-reorder feature fires two concurrent PATCH requests per swap, and a
// user clicking ▲/▼ a few times in a row stacks up several more on top of
// that. sql.Open with no MaxOpenConns/busy_timeout lets Go hand out multiple
// connections to the same SQLite file, and a second concurrent writer gets
// an immediate "database is locked" (SQLITE_BUSY) instead of waiting its
// turn — reproduced live via concurrent curl PATCH /goals/{id}/priority
// calls before this fix. Open() must configure the pool so concurrent
// writers queue instead of failing.
func TestOpen_SurvivesConcurrentWrites(t *testing.T) {
	path := filepath.Join(t.TempDir(), "concurrent.db")

	db, err := Open(path)
	if err != nil {
		t.Fatalf("Open() returned unexpected error: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if _, err := db.Exec(
		`INSERT INTO workspaces (id, user_id, name, created_at) VALUES ('w1', ?, 'private', '2026-01-01T00:00:00Z')`,
		DefaultUserID,
	); err != nil {
		t.Fatalf("failed to seed workspace: %v", err)
	}
	if _, err := db.Exec(
		`INSERT INTO goals (id, workspace_id, title, detail, achievement_condition, end_date, mode, status, created_at)
		 VALUES ('g1', 'w1', 't', 'd', 'c', '2026-01-01', 'strict', 'active', '2026-01-01T00:00:00Z')`,
	); err != nil {
		t.Fatalf("failed to seed goal: %v", err)
	}

	const concurrency = 20
	var wg sync.WaitGroup
	errs := make(chan error, concurrency)
	for i := range concurrency {
		wg.Add(1)
		go func(priority int) {
			defer wg.Done()
			if _, err := db.Exec(`UPDATE goals SET priority = ? WHERE id = 'g1'`, priority); err != nil {
				errs <- fmt.Errorf("write %d: %w", priority, err)
			}
		}(i)
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("concurrent write failed: %v", err)
	}
}

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

// TestOpen_MigratesLegacyDailyTasksTableMissingCompletedAt mirrors the
// postpone_count migration test above, for the completed_at column added
// to daily_tasks after some databases already existed.
func TestOpen_MigratesLegacyDailyTasksTableMissingCompletedAt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.db")

	legacyDB, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("sql.Open() returned unexpected error: %v", err)
	}
	_, err = legacyDB.Exec(`CREATE TABLE daily_tasks (
		id TEXT PRIMARY KEY,
		goal_id TEXT NOT NULL,
		date TEXT NOT NULL,
		content TEXT NOT NULL,
		done INTEGER NOT NULL,
		created_at TEXT NOT NULL
	)`)
	if err != nil {
		t.Fatalf("failed to create legacy daily_tasks table: %v", err)
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
		`INSERT INTO daily_tasks (id, goal_id, date, content, done, created_at, completed_at)
		 VALUES ('t1', 'g1', '2026-01-01', 'c', 1, '2026-01-01T00:00:00Z', '2026-01-01T09:00:00Z')`,
	); err != nil {
		t.Fatalf("insert using completed_at after migration returned unexpected error: %v", err)
	}

	var got string
	if err := db.QueryRow(`SELECT completed_at FROM daily_tasks WHERE id = 't1'`).Scan(&got); err != nil {
		t.Fatalf("failed to read back completed_at: %v", err)
	}
	if got != "2026-01-01T09:00:00Z" {
		t.Errorf("completed_at = %q, want %q", got, "2026-01-01T09:00:00Z")
	}

	// Re-opening (idempotency) must not fail with "duplicate column name".
	db2, err := Open(path)
	if err != nil {
		t.Fatalf("second Open() on an already-migrated database returned unexpected error: %v", err)
	}
	db2.Close()
}

// TestOpen_MigratesLegacyGoalsTableMissingPriority mirrors the
// postpone_count/completed_at migration tests above, for the priority
// column added to goals after some databases already existed.
func TestOpen_MigratesLegacyGoalsTableMissingPriority(t *testing.T) {
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
		created_at TEXT NOT NULL,
		postpone_count INTEGER NOT NULL DEFAULT 0
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
		`INSERT INTO goals (id, workspace_id, title, detail, achievement_condition, end_date, mode, status, created_at, postpone_count, priority)
		 VALUES ('g1', 'w1', 't', 'd', 'c', '2026-01-01', 'strict', 'active', '2026-01-01T00:00:00Z', 0, 5)`,
	); err != nil {
		t.Fatalf("insert using priority after migration returned unexpected error: %v", err)
	}

	var got int
	if err := db.QueryRow(`SELECT priority FROM goals WHERE id = 'g1'`).Scan(&got); err != nil {
		t.Fatalf("failed to read back priority: %v", err)
	}
	if got != 5 {
		t.Errorf("priority = %d, want 5", got)
	}

	// Re-opening (idempotency) must not fail with "duplicate column name".
	db2, err := Open(path)
	if err != nil {
		t.Fatalf("second Open() on an already-migrated database returned unexpected error: %v", err)
	}
	db2.Close()
}
