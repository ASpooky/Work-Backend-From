package goal

import "testing"

type spyDeleteRepository struct {
	deletedID string
	err       error
}

func (s *spyDeleteRepository) Delete(id string) error {
	s.deletedID = id
	return s.err
}

func TestDeleteGoalUsecase_Execute(t *testing.T) {
	repo := &spyDeleteRepository{}
	uc := NewDeleteGoalUsecase(repo)

	if err := uc.Execute("goal-001"); err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}
	if repo.deletedID != "goal-001" {
		t.Errorf("Delete() called with id = %q, want %q", repo.deletedID, "goal-001")
	}
}
