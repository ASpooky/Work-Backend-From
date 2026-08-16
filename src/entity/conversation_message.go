package entity

import "time"

// ConversationMessage is one persisted turn within a Conversation. Unlike
// ChatMessage (a transient DTO for talking to the AI API), this is stored
// so a conversation's history survives across requests.
type ConversationMessage struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	Role           ChatRole  `json:"role"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"created_at"`
}

func NewConversationMessage(id, conversationID string, role ChatRole, content string, createdAt time.Time) *ConversationMessage {
	return &ConversationMessage{
		ID:             id,
		ConversationID: conversationID,
		Role:           role,
		Content:        content,
		CreatedAt:      createdAt,
	}
}
