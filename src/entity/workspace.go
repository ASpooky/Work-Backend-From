package entity

import "time"

type WorkSpace struct {
	ID        string
	UserID    string
	Name      string
	CreatedAt time.Time
}

func NewWorkSpace(id, userID, name string, createdAt time.Time) *WorkSpace {
	return &WorkSpace{
		ID:        id,
		UserID:    userID,
		Name:      name,
		CreatedAt: createdAt,
	}
}
