package goal

import (
	"testing"
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
)

type spyUpdater struct {
	updated *entity.Goal
}

func (s *spyUpdater) Update(goal *entity.Goal) error {
	s.updated = goal
	return nil
}

func TestUpdateGoalUsecase_Execute(t *testing.T) {
	repo := &spyUpdater{}
	uc := NewUpdateGoalUsecase(repo)

	endDate := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)
	input := UpdateGoalInput{
		ID:                   "goal-001",
		Title:                "フルマラソン完走",
		Detail:               "detail",
		AchievementCondition: "cond",
		EndDate:              endDate,
		Mode:                 entity.ModeWant,
	}

	if err := uc.Execute(input); err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}

	if repo.updated == nil {
		t.Fatal("expected repository.Update to be called")
	}
	if repo.updated.ID != input.ID || repo.updated.Title != input.Title || repo.updated.Mode != input.Mode {
		t.Errorf("repository.Update called with = %+v, want fields from %+v", repo.updated, input)
	}
	if !repo.updated.EndDate.Equal(endDate) {
		t.Errorf("EndDate = %v, want %v", repo.updated.EndDate, endDate)
	}
}
