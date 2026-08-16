package ai

import (
	"strings"
	"testing"
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
)

func TestSummarizeBusyDays_ReportsDaysWithSeveralTasks(t *testing.T) {
	tasks := []*entity.DailyTask{
		entity.NewDailyTask("t1", "goal-a", time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC), "x", time.Now()),
		entity.NewDailyTask("t2", "goal-b", time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC), "y", time.Now()),
		entity.NewDailyTask("t3", "goal-a", time.Date(2026, 8, 21, 0, 0, 0, 0, time.UTC), "z", time.Now()), // only 1 task this day
	}

	got := summarizeBusyDays(tasks, "")
	if !strings.Contains(got, "2026-08-20") {
		t.Errorf("summarizeBusyDays() = %q, want it to mention 2026-08-20 (2 tasks)", got)
	}
	if strings.Contains(got, "2026-08-21") {
		t.Errorf("summarizeBusyDays() = %q, want it to omit 2026-08-21 (only 1 task, below the threshold)", got)
	}
}

func TestSummarizeBusyDays_ExcludesGivenGoal(t *testing.T) {
	tasks := []*entity.DailyTask{
		entity.NewDailyTask("t1", "goal-self", time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC), "x", time.Now()),
		entity.NewDailyTask("t2", "goal-self", time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC), "y", time.Now()),
	}

	got := summarizeBusyDays(tasks, "goal-self")
	if got != "" {
		t.Errorf("summarizeBusyDays() with excludeGoalID matching every task = %q, want empty (nothing competing)", got)
	}
}

func TestSummarizeBusyDays_EmptyWhenNothingBusy(t *testing.T) {
	tasks := []*entity.DailyTask{
		entity.NewDailyTask("t1", "goal-a", time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC), "x", time.Now()),
	}

	got := summarizeBusyDays(tasks, "")
	if got != "" {
		t.Errorf("summarizeBusyDays() with only 1 task on any day = %q, want empty", got)
	}
}
