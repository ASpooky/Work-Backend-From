package dailytask

import (
	"testing"
	"time"
)

func TestUpdateDailyTaskDoneUsecase_Execute_MarkDone(t *testing.T) {
	repo := &mockRepository{}
	fixedNow := time.Date(2026, 8, 17, 10, 0, 0, 0, time.UTC)
	u := NewUpdateDailyTaskDoneUsecase(repo, stubClock{now: fixedNow})

	if err := u.Execute(UpdateDailyTaskDoneInput{ID: "task-001", Done: true}); err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}

	if repo.updatedID != "task-001" {
		t.Errorf("updatedID = %q, want %q", repo.updatedID, "task-001")
	}
	if !repo.updatedDone {
		t.Errorf("updatedDone = %v, want true", repo.updatedDone)
	}
	if repo.updatedCompletedAt == nil || !repo.updatedCompletedAt.Equal(fixedNow) {
		t.Errorf("updatedCompletedAt = %v, want %v", repo.updatedCompletedAt, fixedNow)
	}
}

func TestUpdateDailyTaskDoneUsecase_Execute_MarkNotDone_ClearsCompletedAt(t *testing.T) {
	repo := &mockRepository{}
	u := NewUpdateDailyTaskDoneUsecase(repo, stubClock{now: time.Now()})

	if err := u.Execute(UpdateDailyTaskDoneInput{ID: "task-001", Done: false}); err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}

	if repo.updatedDone {
		t.Errorf("updatedDone = %v, want false", repo.updatedDone)
	}
	if repo.updatedCompletedAt != nil {
		t.Errorf("updatedCompletedAt = %v, want nil when marking not done", repo.updatedCompletedAt)
	}
}
