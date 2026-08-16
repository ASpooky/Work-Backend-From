package entity

import "time"

// Conversation is a persisted AI planning chat thread, so past 壁打ち
// context isn't lost between sessions. GoalID is empty for a general
// new-goal-creation chat (scoped only to a workspace) and set for a
// conversation reviewing one specific existing goal.
type Conversation struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	GoalID      string    `json:"goal_id"`
	Title       string    `json:"title"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func NewConversation(id, workspaceID, goalID, title string, createdAt time.Time) *Conversation {
	return &Conversation{
		ID:          id,
		WorkspaceID: workspaceID,
		GoalID:      goalID,
		Title:       title,
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
	}
}
