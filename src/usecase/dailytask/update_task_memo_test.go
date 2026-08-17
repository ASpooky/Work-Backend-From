package dailytask

import "testing"

type mockMemoUpdater struct {
	updatedID   string
	updatedMemo *string
}

func (m *mockMemoUpdater) UpdateMemo(id string, memo *string) error {
	m.updatedID = id
	m.updatedMemo = memo
	return nil
}

func TestUpdateTaskMemoUsecase_Execute_SetsMemo(t *testing.T) {
	repo := &mockMemoUpdater{}
	u := NewUpdateTaskMemoUsecase(repo)

	memo := "腰が痛かった"
	if err := u.Execute(UpdateTaskMemoInput{ID: "task-001", Memo: &memo}); err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}

	if repo.updatedID != "task-001" {
		t.Errorf("updatedID = %q, want %q", repo.updatedID, "task-001")
	}
	if repo.updatedMemo == nil || *repo.updatedMemo != memo {
		t.Errorf("updatedMemo = %v, want %q", repo.updatedMemo, memo)
	}
}

func TestUpdateTaskMemoUsecase_Execute_ClearsMemo(t *testing.T) {
	repo := &mockMemoUpdater{}
	u := NewUpdateTaskMemoUsecase(repo)

	if err := u.Execute(UpdateTaskMemoInput{ID: "task-001", Memo: nil}); err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}

	if repo.updatedMemo != nil {
		t.Errorf("updatedMemo = %v, want nil", repo.updatedMemo)
	}
}
