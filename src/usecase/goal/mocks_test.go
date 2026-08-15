package goal

import (
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
)

type stubIDGenerator struct {
	id string
}

func (s stubIDGenerator) NewID() string {
	return s.id
}

type stubClock struct {
	now time.Time
}

func (s stubClock) Now() time.Time {
	return s.now
}

type mockRepository struct {
	saved             *entity.Goal
	findByWorkspaceID []*entity.Goal
}

func (m *mockRepository) Save(goal *entity.Goal) error {
	m.saved = goal
	return nil
}

func (m *mockRepository) FindByWorkspaceID(workspaceID string) ([]*entity.Goal, error) {
	return m.findByWorkspaceID, nil
}
