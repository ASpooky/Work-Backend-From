package workspace

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
	saved       *entity.WorkSpace
	findAll     []*entity.WorkSpace
	updatedID   string
	updatedName string
	deletedID   string
}

func (m *mockRepository) Save(workspace *entity.WorkSpace) error {
	m.saved = workspace
	return nil
}

func (m *mockRepository) FindAll() ([]*entity.WorkSpace, error) {
	return m.findAll, nil
}

func (m *mockRepository) UpdateName(id, name string) error {
	m.updatedID = id
	m.updatedName = name
	return nil
}

func (m *mockRepository) Delete(id string) error {
	m.deletedID = id
	return nil
}
