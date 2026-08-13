package entity

import (
	"reflect"
	"testing"
	"time"
)

func TestNewGoal(t *testing.T) {
	createdAt := time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)

	got := NewGoal("goal-001", "workspace-001", "Run a marathon", "Complete a full marathon under 4 hours", "Finish 42.195km within 4 hours", endDate, ModeStrict, createdAt)

	want := &Goal{
		ID:                   "goal-001",
		WorkspaceID:          "workspace-001",
		Title:                "Run a marathon",
		Detail:               "Complete a full marathon under 4 hours",
		AchievementCondition: "Finish 42.195km within 4 hours",
		EndDate:              endDate,
		Mode:                 ModeStrict,
		Status:               StatusActive,
		CreatedAt:            createdAt,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("NewGoal() = %+v, want %+v", got, want)
	}
}

func TestGoal_Postpone_StrictMode(t *testing.T) {
	endDate := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	goal := &Goal{Mode: ModeStrict, EndDate: endDate}

	goal.Postpone()

	want := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
	if !goal.EndDate.Equal(want) {
		t.Errorf("EndDate = %v, want %v", goal.EndDate, want)
	}
}

func TestGoal_Postpone_WantMode(t *testing.T) {
	endDate := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	goal := &Goal{Mode: ModeWant, EndDate: endDate}

	goal.Postpone()

	if !goal.EndDate.Equal(endDate) {
		t.Errorf("EndDate = %v, want unchanged %v", goal.EndDate, endDate)
	}
}

func TestGoal_UpdateAchievement_AllDone(t *testing.T) {
	goal := &Goal{ID: "goal-001", Status: StatusActive}
	tasks := []DailyTask{
		{GoalID: "goal-001", Done: true},
		{GoalID: "goal-001", Done: true},
	}

	goal.UpdateAchievement(tasks)

	if goal.Status != StatusAchieved {
		t.Errorf("Status = %v, want %v", goal.Status, StatusAchieved)
	}
}

func TestGoal_UpdateAchievement_NotAllDone(t *testing.T) {
	goal := &Goal{ID: "goal-001", Status: StatusActive}
	tasks := []DailyTask{
		{GoalID: "goal-001", Done: true},
		{GoalID: "goal-001", Done: false},
	}

	goal.UpdateAchievement(tasks)

	if goal.Status != StatusActive {
		t.Errorf("Status = %v, want %v", goal.Status, StatusActive)
	}
}
