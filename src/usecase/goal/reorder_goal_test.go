package goal

import "testing"

type spyPriorityUpdater struct {
	id       string
	priority int
	err      error
}

func (s *spyPriorityUpdater) UpdatePriority(id string, priority int) error {
	s.id = id
	s.priority = priority
	return s.err
}

func TestReorderGoalUsecase_Execute(t *testing.T) {
	repo := &spyPriorityUpdater{}
	uc := NewReorderGoalUsecase(repo)

	if err := uc.Execute(ReorderGoalInput{ID: "goal-001", Priority: -1}); err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}
	if repo.id != "goal-001" || repo.priority != -1 {
		t.Errorf("UpdatePriority() called with (%q, %d), want (%q, %d)", repo.id, repo.priority, "goal-001", -1)
	}
}
