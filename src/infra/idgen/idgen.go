package idgen

import "github.com/google/uuid"

// UUIDGenerator satisfies usecase.IDGenerator via random UUIDs.
type UUIDGenerator struct{}

func (UUIDGenerator) NewID() string {
	return uuid.NewString()
}
