package sqlite

import (
	"database/sql"
	_ "embed"
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
	return nil
}
