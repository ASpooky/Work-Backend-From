package sqlite

import (
	"database/sql"
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
)

type ConversationMessageRepository struct {
	db *sql.DB
}

func NewConversationMessageRepository(db *sql.DB) *ConversationMessageRepository {
	return &ConversationMessageRepository{db: db}
}

func (r *ConversationMessageRepository) Save(m *entity.ConversationMessage) error {
	_, err := r.db.Exec(
		`INSERT INTO conversation_messages (id, conversation_id, role, content, created_at) VALUES (?, ?, ?, ?, ?)`,
		m.ID, m.ConversationID, string(m.Role), m.Content, m.CreatedAt.Format(time.RFC3339),
	)
	return err
}

func (r *ConversationMessageRepository) FindByConversationID(conversationID string) ([]*entity.ConversationMessage, error) {
	rows, err := r.db.Query(
		`SELECT id, conversation_id, role, content, created_at FROM conversation_messages
		 WHERE conversation_id = ? ORDER BY created_at ASC`,
		conversationID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := []*entity.ConversationMessage{}
	for rows.Next() {
		var id, convID, role, content, createdAt string
		if err := rows.Scan(&id, &convID, &role, &content, &createdAt); err != nil {
			return nil, err
		}

		parsedCreatedAt, err := time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return nil, err
		}

		messages = append(messages, entity.NewConversationMessage(id, convID, entity.ChatRole(role), content, parsedCreatedAt))
	}
	return messages, rows.Err()
}
