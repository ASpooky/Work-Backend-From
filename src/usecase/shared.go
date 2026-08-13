package usecase

import "time"

type IDGenerator interface {
	NewID() string
}

type Clock interface {
	Now() time.Time
}
