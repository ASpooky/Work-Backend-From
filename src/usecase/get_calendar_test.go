package usecase

import (
	"testing"
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
)

type stubGoalReader struct {
	goals []*entity.Goal
}

func (s stubGoalReader) FindByWorkspaceID(workspaceID string) ([]*entity.Goal, error) {
	return s.goals, nil
}

type stubDailyTaskRangeReader struct {
	tasksByGoal map[string][]*entity.DailyTask
}

func (s stubDailyTaskRangeReader) FindByGoalIDAndDateRange(goalID string, from, to time.Time) ([]*entity.DailyTask, error) {
	return s.tasksByGoal[goalID], nil
}

func TestGetCalendarUsecase_Execute(t *testing.T) {
	goalA := entity.NewGoal("goal-001", "workspace-001", "Run", "detail", "cond", time.Now(), entity.ModeStrict, time.Now())

	from := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC)

	doneTask := entity.NewDailyTask("task-001", goalA.ID, from, "Run 5km", time.Now())
	doneTask.Done = true
	notDoneTask := entity.NewDailyTask("task-002", goalA.ID, from.AddDate(0, 0, 1), "Run 8km", time.Now())
	// 2026-08-03 intentionally has no task -> want DayStatusNoTask

	goals := stubGoalReader{goals: []*entity.Goal{goalA}}
	tasks := stubDailyTaskRangeReader{tasksByGoal: map[string][]*entity.DailyTask{
		goalA.ID: {doneTask, notDoneTask},
	}}

	u := NewGetCalendarUsecase(goals, tasks)
	got, err := u.Execute(GetCalendarInput{WorkspaceID: "workspace-001", From: from, To: to})
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("Execute() returned %d goal calendars, want 1", len(got))
	}
	if got[0].Goal.ID != goalA.ID {
		t.Errorf("Goal.ID = %v, want %v", got[0].Goal.ID, goalA.ID)
	}

	days := got[0].Days
	if days["2026-08-01"].Status != DayStatusDone {
		t.Errorf("2026-08-01 status = %v, want %v", days["2026-08-01"].Status, DayStatusDone)
	}
	if days["2026-08-01"].Content != "Run 5km" {
		t.Errorf("2026-08-01 content = %q, want %q", days["2026-08-01"].Content, "Run 5km")
	}
	if days["2026-08-02"].Status != DayStatusNotDone {
		t.Errorf("2026-08-02 status = %v, want %v", days["2026-08-02"].Status, DayStatusNotDone)
	}
	if days["2026-08-02"].Content != "Run 8km" {
		t.Errorf("2026-08-02 content = %q, want %q", days["2026-08-02"].Content, "Run 8km")
	}
	if days["2026-08-03"].Status != DayStatusNoTask {
		t.Errorf("2026-08-03 status = %v, want %v", days["2026-08-03"].Status, DayStatusNoTask)
	}
	if days["2026-08-03"].Content != "" {
		t.Errorf("2026-08-03 content = %q, want empty", days["2026-08-03"].Content)
	}
}
