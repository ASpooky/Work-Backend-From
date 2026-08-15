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

func (r *WorkspaceRepository) FindAll() ([]*entity.WorkSpace, error) {
	rows, err := r.db.Query(`SELECT id, user_id, name, created_at FROM workspaces ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workspaces []*entity.WorkSpace
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
