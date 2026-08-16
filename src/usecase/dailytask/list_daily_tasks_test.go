package dailytask

import (
	"reflect"
	"testing"
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
)

func TestListDailyTasksUsecase_Execute(t *testing.T) {
	createdAt := time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)
	date := time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)
	want := []*entity.DailyTask{
		entity.NewDailyTask("task-001", "goal-001", date, "Run 5km", createdAt),
	}

	repo := &mockRepository{findByDate: want}
	uc := NewListDailyTasksUsecase(repo)

	got, err := uc.Execute(ListDailyTasksInput{Date: date})
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Execute() = %+v, want %+v", got, want)
	}
}

func TestListDailyTasksUsecase_Execute_ScopedToWorkspace(t *testing.T) {
	createdAt := time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)
	date := time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)
	want := []*entity.DailyTask{
		entity.NewDailyTask("task-001", "goal-001", date, "Run 5km", createdAt),
	}

	repo := &mockRepository{findByDateAndWorkspaceID: want}
	uc := NewListDailyTasksUsecase(repo)

	got, err := uc.Execute(ListDailyTasksInput{Date: date, WorkspaceID: "workspace-001"})
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Execute() = %+v, want %+v", got, want)
	}
	if repo.capturedWorkspaceID != "workspace-001" {
		t.Errorf("FindByDateAndWorkspaceID called with workspaceID = %q, want %q", repo.capturedWorkspaceID, "workspace-001")
	}
}
