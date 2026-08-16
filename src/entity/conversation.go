package entity

import "time"

// Conversation is a persisted AI planning chat thread, so past 壁打ち
// context isn't lost between sessions.
type Conversation struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	Title       string    `json:"title"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func NewConversation(id, workspaceID, title string, createdAt time.Time) *Conversation {
	return &Conversation{
		ID:          id,
		WorkspaceID: workspaceID,
		Title:       title,
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
	}
}
