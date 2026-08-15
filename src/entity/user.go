package entity

import "time"

type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

func NewUser(id, name string, createdAt time.Time) *User {
	return &User{
		ID:        id,
		Name:      name,
		CreatedAt: createdAt,
	}
}
