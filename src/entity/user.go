package entity

import "time"

type User struct {
	ID        string
	Name      string
	CreatedAt time.Time
}

func NewUser(id, name string, createdAt time.Time) *User {
	return &User{
		ID:        id,
		Name:      name,
		CreatedAt: createdAt,
	}
}
