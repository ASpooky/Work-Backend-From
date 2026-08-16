package usecase

import (
	"testing"
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
)

type stubSingleGoalReader struct {
	goal *entity.Goal
}

func (s stubSingleGoalReader) FindByID(id string) (*entity.Goal, error) {
	return s.goal, nil
}

func TestGetGoalStatsUsecase_Execute(t *testing.T) {
	created := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 8, 21, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC)

	goal := entity.NewGoal("goal-001", "workspace-001", "Run", "detail", "cond", endDate, entity.ModeStrict, created)
	goal.PostponeCount = 2

	pastDone := entity.NewDailyTask("t1", goal.ID, now.AddDate(0, 0, -3), "x", created)
	pastDone.Done = true
	pastNotDone := entity.NewDailyTask("t2", goal.ID, now.AddDate(0, 0, -2), "x", created)
	future := entity.NewDailyTask("t3", goal.ID, now.AddDate(0, 0, 1), "x", created) // not yet due, must not count

	goals := stubSingleGoalReader{goal: goal}
	tasks := stubDailyTaskRangeReader{tasksByGoal: map[string][]*entity.DailyTask{
		goal.ID: {pastDone, pastNotDone, future},
	}}

	u := NewGetGoalStatsUsecase(goals, tasks, stubClock{now: now})
	gotGoal, stats, err := u.Execute("goal-001")
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}

	if gotGoal.ID != goal.ID {
		t.Errorf("Goal.ID = %v, want %v", gotGoal.ID, goal.ID)
	}
	if stats.ScheduledCount != 2 {
		t.Errorf("ScheduledCount = %d, want 2 (future task excluded)", stats.ScheduledCount)
	}
	if stats.DoneCount != 1 {
		t.Errorf("DoneCount = %d, want 1", stats.DoneCount)
	}
	if stats.AchievementRate != 0.5 {
		t.Errorf("AchievementRate = %v, want 0.5", stats.AchievementRate)
	}
	if stats.PostponeCount != 2 {
		t.Errorf("PostponeCount = %d, want 2", stats.PostponeCount)
	}
	if stats.DaysRemaining != 5 {
		t.Errorf("DaysRemaining = %d, want 5", stats.DaysRemaining)
	}
}

func TestGetGoalStatsUsecase_Execute_NoScheduledTasks(t *testing.T) {
	goal := entity.NewGoal("goal-001", "workspace-001", "Run", "detail", "cond", time.Now().AddDate(0, 0, 10), entity.ModeStrict, time.Now())
	goals := stubSingleGoalReader{goal: goal}
	tasks := stubDailyTaskRangeReader{tasksByGoal: map[string][]*entity.DailyTask{}}

	u := NewGetGoalStatsUsecase(goals, tasks, stubClock{now: time.Now()})
	_, stats, err := u.Execute("goal-001")
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}
	if stats.AchievementRate != 0 {
		t.Errorf("AchievementRate with zero scheduled tasks = %v, want 0 (no division by zero)", stats.AchievementRate)
	}
}
