package dailytask

import "testing"

type mockDeleteRepository struct {
	deletedID string
}

func (m *mockDeleteRepository) Delete(id string) error {
	m.deletedID = id
	return nil
}

func TestDeleteDailyTaskUsecase_Execute(t *testing.T) {
	repo := &mockDeleteRepository{}
	u := NewDeleteDailyTaskUsecase(repo)

	if err := u.Execute("task-001"); err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}
	if repo.deletedID != "task-001" {
		t.Errorf("Delete() called with id = %q, want %q", repo.deletedID, "task-001")
	}
}
