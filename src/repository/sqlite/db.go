package sqlite

import (
	"database/sql"
	_ "embed"
	"fmt"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schema string

// DefaultUserID is the fixed User.id seeded by schema.sql. entity.md allows
// a fixed id for single-user deployments, so callers can use this directly
// instead of looking it up.
const DefaultUserID = "00000000-0000-0000-0000-000000000001"

// Open connects to the sqlite database at path and applies schema.sql.
func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	// SQLite allows only one writer at a time; database/sql's default pool
	// hands out multiple connections to concurrent goroutines, and a second
	// writer trying to go through its own connection at the same moment
	// gets an immediate "database is locked" (SQLITE_BUSY) instead of
	// waiting its turn. A single shared connection serializes all access
	// through this *sql.DB, which is the simplest correct fix for a
	// single-user app with no meaningful concurrent-write throughput needs.
	// busy_timeout is set too, as defense-in-depth against a second,
	// separate process (e.g. cmd/seed run manually against the same file
	// while the server is up) still holding the file lock briefly.
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`PRAGMA busy_timeout = 5000`); err != nil {
		db.Close()
		return nil, err
	}

	if err := migrate(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func migrate(db *sql.DB) error {
	for _, stmt := range strings.Split(schema, ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}

	// CREATE TABLE IF NOT EXISTS above only applies to brand-new databases;
	// a database created before a column existed needs an explicit ALTER
	// TABLE, applied only if the column is actually missing (ALTER TABLE
	// ADD COLUMN isn't safe to re-run unconditionally).
	if err := addColumnIfMissing(db, "goals", "postpone_count", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if err := addColumnIfMissing(db, "conversations", "goal_id", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	if err := addColumnIfMissing(db, "daily_tasks", "completed_at", "TEXT"); err != nil {
		return err
	}
	if err := addColumnIfMissing(db, "goals", "priority", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	return addColumnIfMissing(db, "daily_tasks", "memo", "TEXT")
}

func addColumnIfMissing(db *sql.DB, table, column, definition string) error {
	exists, err := columnExists(db, table, column)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	_, err = db.Exec(fmt.Sprintf(`ALTER TABLE %s ADD COLUMN %s %s`, table, column, definition))
	return err
}

func columnExists(db *sql.DB, table, column string) (bool, error) {
	rows, err := db.Query(fmt.Sprintf(`PRAGMA table_info(%s)`, table))
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, ctype string
		var notNull, pk int
		var dflt any
		if err := rows.Scan(&cid, &name, &ctype, &notNull, &dflt, &pk); err != nil {
			return false, err
		}
		if name == column {
			return true, nil
		}
	}
	return false, rows.Err()
}
