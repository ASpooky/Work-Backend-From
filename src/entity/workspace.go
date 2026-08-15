package entity

import "time"

type WorkSpace struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

func NewWorkSpace(id, userID, name string, createdAt time.Time) *WorkSpace {
	return &WorkSpace{
		ID:        id,
		UserID:    userID,
		Name:      name,
		CreatedAt: createdAt,
	}
}
