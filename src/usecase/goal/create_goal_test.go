package goal

import (
	"reflect"
	"testing"
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
)

func TestCreateGoalUsecase_Execute(t *testing.T) {
	fixedID := "goal-001"
	fixedNow := time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)

	repo := &mockRepository{}
	uc := NewCreateGoalUsecase(repo, stubIDGenerator{id: fixedID}, stubClock{now: fixedNow})

	input := CreateGoalInput{
		WorkspaceID:          "workspace-001",
		Title:                "Run a marathon",
		Detail:               "Complete a full marathon under 4 hours",
		AchievementCondition: "Finish 42.195km within 4 hours",
		EndDate:              endDate,
		Mode:                 entity.ModeStrict,
	}

	got, err := uc.Execute(input)
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}

	want := entity.NewGoal(fixedID, input.WorkspaceID, input.Title, input.Detail, input.AchievementCondition, input.EndDate, input.Mode, fixedNow)

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
