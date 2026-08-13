package workspace

import (
	"reflect"
	"testing"
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
	saved *entity.WorkSpace
}

func (m *mockRepository) Save(workspace *entity.WorkSpace) error {
	m.saved = workspace
	return nil
}

func TestCreateWorkspaceUsecase_Execute(t *testing.T) {
	fixedID := "workspace-001"
	fixedNow := time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)

	repo := &mockRepository{}
	uc := NewCreateWorkspaceUsecase(repo, stubIDGenerator{id: fixedID}, stubClock{now: fixedNow})

	input := CreateWorkspaceInput{
		UserID: "user-001",
		Name:   "private",
	}

	got, err := uc.Execute(input)
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}

	want := entity.NewWorkSpace(fixedID, input.UserID, input.Name, fixedNow)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Execute() = %+v, want %+v", got, want)
	}

	if repo.saved == nil {
		t.Fatal("expected repository.Save to be called")
	}
	if !reflect.DeepEqual(repo.saved, want) {
		t.Errorf("repository saved = %+v, want %+v", repo.saved, want)
	}
}
