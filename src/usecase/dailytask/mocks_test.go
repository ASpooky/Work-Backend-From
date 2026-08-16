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
	saved                    *entity.DailyTask
	findByDate               []*entity.DailyTask
	findByDateAndWorkspaceID []*entity.DailyTask
	capturedWorkspaceID      string
	updatedID                string
	updatedDone              bool
	updatedCompletedAt       *time.Time
}

func (m *mockRepository) Save(task *entity.DailyTask) error {
	m.saved = task
	return nil
}

func (m *mockRepository) FindByDate(date time.Time) ([]*entity.DailyTask, error) {
	return m.findByDate, nil
}

func (m *mockRepository) FindByDateAndWorkspaceID(date time.Time, workspaceID string) ([]*entity.DailyTask, error) {
	m.capturedWorkspaceID = workspaceID
	return m.findByDateAndWorkspaceID, nil
}

func (m *mockRepository) UpdateDone(id string, done bool, completedAt *time.Time) error {
	m.updatedID = id
	m.updatedDone = done
	m.updatedCompletedAt = completedAt
	return nil
}

type recordingRepository struct {
	saved []*entity.DailyTask
}

func (r *recordingRepository) Save(task *entity.DailyTask) error {
	r.saved = append(r.saved, task)
	return nil
}

func (r *recordingRepository) FindByDate(date time.Time) ([]*entity.DailyTask, error) {
	return nil, nil
}

func (r *recordingRepository) FindByDateAndWorkspaceID(date time.Time, workspaceID string) ([]*entity.DailyTask, error) {
	return nil, nil
}
