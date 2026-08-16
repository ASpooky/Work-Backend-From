package sqlite

import (
	"database/sql"
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
)

type ConversationRepository struct {
	db *sql.DB
}

func NewConversationRepository(db *sql.DB) *ConversationRepository {
	return &ConversationRepository{db: db}
}

func (r *ConversationRepository) Save(c *entity.Conversation) error {
	_, err := r.db.Exec(
		`INSERT INTO conversations (id, workspace_id, title, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		c.ID, c.WorkspaceID, c.Title, c.CreatedAt.Format(time.RFC3339), c.UpdatedAt.Format(time.RFC3339),
	)
	return err
}

func (r *ConversationRepository) FindByWorkspaceID(workspaceID string) ([]*entity.Conversation, error) {
	rows, err := r.db.Query(
		`SELECT id, workspace_id, title, created_at, updated_at FROM conversations
		 WHERE workspace_id = ? ORDER BY updated_at DESC`,
		workspaceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	conversations := []*entity.Conversation{}
	for rows.Next() {
		var id, wsID, title, createdAt, updatedAt string
		if err := rows.Scan(&id, &wsID, &title, &createdAt, &updatedAt); err != nil {
			return nil, err
		}

		parsedCreatedAt, err := time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return nil, err
		}
		parsedUpdatedAt, err := time.Parse(time.RFC3339, updatedAt)
		if err != nil {
			return nil, err
		}

		c := entity.NewConversation(id, wsID, title, parsedCreatedAt)
		c.UpdatedAt = parsedUpdatedAt
		conversations = append(conversations, c)
	}
	return conversations, rows.Err()
}

func (r *ConversationRepository) Touch(id string, updatedAt time.Time) error {
	_, err := r.db.Exec(`UPDATE conversations SET updated_at = ? WHERE id = ?`, updatedAt.Format(time.RFC3339), id)
	return err
}
