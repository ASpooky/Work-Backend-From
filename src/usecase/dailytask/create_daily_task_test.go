package dailytask

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
	saved *entity.DailyTask
}

func (m *mockRepository) Save(task *entity.DailyTask) error {
	m.saved = task
	return nil
}

func TestCreateDailyTaskUsecase_Execute(t *testing.T) {
	fixedID := "task-001"
	fixedNow := time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)
	date := time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)

	repo := &mockRepository{}
	uc := NewCreateDailyTaskUsecase(repo, stubIDGenerator{id: fixedID}, stubClock{now: fixedNow})

	input := CreateDailyTaskInput{
		GoalID:  "goal-001",
		Date:    date,
		Content: "Run 5km",
	}

	got, err := uc.Execute(input)
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}

	want := entity.NewDailyTask(fixedID, input.GoalID, input.Date, input.Content, fixedNow)

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
