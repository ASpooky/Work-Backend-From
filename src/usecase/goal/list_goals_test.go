package goal

import (
	"reflect"
	"testing"
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
)

func TestListGoalsUsecase_Execute(t *testing.T) {
	createdAt := time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	want := []*entity.Goal{
		entity.NewGoal("goal-001", "workspace-001", "Run a marathon", "detail", "condition", endDate, entity.ModeStrict, createdAt),
	}

	repo := &mockRepository{findByWorkspaceID: want}
	uc := NewListGoalsUsecase(repo)

	got, err := uc.Execute(ListGoalsInput{WorkspaceID: "workspace-001"})
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Execute() = %+v, want %+v", got, want)
	}
}

func TestListGoalsUsecase_Execute_AllWorkspaces(t *testing.T) {
	want := []*entity.Goal{
		entity.NewGoal("goal-a", "workspace-a", "A", "detail", "cond", time.Now(), entity.ModeStrict, time.Now()),
		entity.NewGoal("goal-b", "workspace-b", "B", "detail", "cond", time.Now(), entity.ModeStrict, time.Now()),
	}

	repo := &mockRepository{findAll: want}
	uc := NewListGoalsUsecase(repo)

	got, err := uc.Execute(ListGoalsInput{WorkspaceID: ""})
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("Execute() with empty WorkspaceID = %+v, want all goals across workspaces", got)
	}
}
