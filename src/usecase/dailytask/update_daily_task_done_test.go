package dailytask

import "testing"

func TestUpdateDailyTaskDoneUsecase_Execute(t *testing.T) {
	repo := &mockRepository{}
	u := NewUpdateDailyTaskDoneUsecase(repo)

	if err := u.Execute(UpdateDailyTaskDoneInput{ID: "task-001", Done: true}); err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}

	if repo.updatedID != "task-001" {
		t.Errorf("updatedID = %q, want %q", repo.updatedID, "task-001")
	}
	if !repo.updatedDone {
		t.Errorf("updatedDone = %v, want true", repo.updatedDone)
	}
}
