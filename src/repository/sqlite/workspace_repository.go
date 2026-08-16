package sqlite

import (
	"database/sql"
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
)

type WorkspaceRepository struct {
	db *sql.DB
}

func NewWorkspaceRepository(db *sql.DB) *WorkspaceRepository {
	return &WorkspaceRepository{db: db}
}

func (r *WorkspaceRepository) Save(workspace *entity.WorkSpace) error {
	_, err := r.db.Exec(
		`INSERT INTO workspaces (id, user_id, name, created_at) VALUES (?, ?, ?, ?)`,
		workspace.ID, workspace.UserID, workspace.Name, workspace.CreatedAt.Format(time.RFC3339),
	)
	return err
}

func (r *WorkspaceRepository) UpdateName(id, name string) error {
	_, err := r.db.Exec(`UPDATE workspaces SET name = ? WHERE id = ?`, name, id)
	return err
}

// Delete removes a workspace and everything scoped to it (goals, their
// daily tasks, AI conversations, and those conversations' messages) in a
// single transaction, since SQLite foreign keys aren't enforced/cascaded
// here (no PRAGMA foreign_keys, no ON DELETE CASCADE in schema.sql).
func (r *WorkspaceRepository) Delete(id string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	statements := []struct {
		query string
		args  []any
	}{
		{`DELETE FROM conversation_messages WHERE conversation_id IN (SELECT id FROM conversations WHERE workspace_id = ?)`, []any{id}},
		{`DELETE FROM conversations WHERE workspace_id = ?`, []any{id}},
		{`DELETE FROM daily_tasks WHERE goal_id IN (SELECT id FROM goals WHERE workspace_id = ?)`, []any{id}},
		{`DELETE FROM goals WHERE workspace_id = ?`, []any{id}},
		{`DELETE FROM workspaces WHERE id = ?`, []any{id}},
	}
	for _, s := range statements {
		if _, err := tx.Exec(s.query, s.args...); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *WorkspaceRepository) FindAll() ([]*entity.WorkSpace, error) {
	rows, err := r.db.Query(`SELECT id, user_id, name, created_at FROM workspaces ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	workspaces := []*entity.WorkSpace{}
	for rows.Next() {
		var id, userID, name, createdAt string
		if err := rows.Scan(&id, &userID, &name, &createdAt); err != nil {
			return nil, err
		}

		parsedCreatedAt, err := time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return nil, err
		}

		workspaces = append(workspaces, entity.NewWorkSpace(id, userID, name, parsedCreatedAt))
	}
	return workspaces, rows.Err()
}
