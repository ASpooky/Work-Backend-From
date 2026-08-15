package dailytask

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
	saved      *entity.DailyTask
	findByDate []*entity.DailyTask
}

func (m *mockRepository) Save(task *entity.DailyTask) error {
	m.saved = task
	return nil
}

func (m *mockRepository) FindByDate(date time.Time) ([]*entity.DailyTask, error) {
	return m.findByDate, nil
}
