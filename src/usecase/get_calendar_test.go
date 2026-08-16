package usecase

import (
	"testing"
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
)

type stubGoalReader struct {
	goals []*entity.Goal
}

// FindByWorkspaceID only returns goals whose WorkspaceID actually matches,
// so a test can tell "scoped to one workspace" apart from "all workspaces"
// (FindAll) — a stub that ignored workspaceID would make that distinction
// untestable.
func (s stubGoalReader) FindByWorkspaceID(workspaceID string) ([]*entity.Goal, error) {
	matched := []*entity.Goal{}
	for _, g := range s.goals {
		if g.WorkspaceID == workspaceID {
			matched = append(matched, g)
		}
	}
	return matched, nil
}

func (s stubGoalReader) FindAll() ([]*entity.Goal, error) {
	return s.goals, nil
}

type stubDailyTaskRangeReader struct {
	tasksByGoal map[string][]*entity.DailyTask
}

// FindByGoalIDAndDateRange actually filters by [from, to] (inclusive), like
// the real DailyTaskRepository — a stub that ignored the bounds would make
// range-dependent behavior (e.g. "hasn't started yet" filtering) untestable.
func (s stubDailyTaskRangeReader) FindByGoalIDAndDateRange(goalID string, from, to time.Time) ([]*entity.DailyTask, error) {
	matched := []*entity.DailyTask{}
	for _, t := range s.tasksByGoal[goalID] {
		if !t.Date.Before(from) && !t.Date.After(to) {
			matched = append(matched, t)
		}
	}
	return matched, nil
}

func TestGetCalendarUsecase_Execute(t *testing.T) {
	from := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC)

	goalA := entity.NewGoal("goal-001", "workspace-001", "Run", "detail", "cond", time.Now(), entity.ModeStrict, from)

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

func TestGetCalendarUsecase_Execute_AllWorkspaces(t *testing.T) {
	from := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	goalA := entity.NewGoal("goal-a", "workspace-a", "Aのgoal", "detail", "cond", time.Now(), entity.ModeStrict, from)
	goalB := entity.NewGoal("goal-b", "workspace-b", "Bのgoal", "detail", "cond", time.Now(), entity.ModeStrict, from)

	goals := stubGoalReader{goals: []*entity.Goal{goalA, goalB}}
	tasks := stubDailyTaskRangeReader{tasksByGoal: map[string][]*entity.DailyTask{
		goalA.ID: {entity.NewDailyTask("task-a", goalA.ID, from, "x", time.Now())},
		goalB.ID: {entity.NewDailyTask("task-b", goalB.ID, from, "x", time.Now())},
	}}

	u := NewGetCalendarUsecase(goals, tasks)
	got, err := u.Execute(GetCalendarInput{WorkspaceID: "", From: from, To: to})
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("Execute() with empty WorkspaceID returned %d goal calendars, want 2 (across all workspaces)", len(got))
	}
}

// TestGetCalendarUsecase_Execute_HidesGoalsThatHaventStartedYet covers a gap
// exposed by multi-phase AI planning: creating goal phases 1-4 as a batch
// leaves phases 2-4 with tasks entirely in the future relative to "today".
// Those shouldn't clutter this week's calendar until they're actually
// relevant — a goal with zero tasks on or before the visible week's end
// hasn't started yet and should be excluded from the result entirely.
func TestGetCalendarUsecase_Execute_HidesGoalsThatHaventStartedYet(t *testing.T) {
	from := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 8, 7, 0, 0, 0, 0, time.UTC)

	started := entity.NewGoal("goal-started", "workspace-001", "土台作り期", "detail", "cond", time.Now(), entity.ModeStrict, from)
	notStarted := entity.NewGoal("goal-not-started", "workspace-001", "追い込み期", "detail", "cond", time.Now(), entity.ModeStrict, from)

	goals := stubGoalReader{goals: []*entity.Goal{started, notStarted}}
	tasks := stubDailyTaskRangeReader{tasksByGoal: map[string][]*entity.DailyTask{
		started.ID:    {entity.NewDailyTask("task-started", started.ID, from, "土台ジョグ", time.Now())},
		notStarted.ID: {entity.NewDailyTask("task-future", notStarted.ID, to.AddDate(0, 1, 0), "ロング走", time.Now())},
	}}

	u := NewGetCalendarUsecase(goals, tasks)
	got, err := u.Execute(GetCalendarInput{WorkspaceID: "workspace-001", From: from, To: to})
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("Execute() returned %d goal calendars, want 1 (the not-yet-started phase excluded)", len(got))
	}
	if got[0].Goal.ID != started.ID {
		t.Errorf("Execute()[0].Goal.ID = %v, want %v", got[0].Goal.ID, started.ID)
	}
}
