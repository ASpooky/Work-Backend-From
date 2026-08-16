package usecase

import (
	"testing"
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
)

type stubClock struct {
	now time.Time
}

func (s stubClock) Now() time.Time {
	return s.now
}

type stubCatchUpGoalReader struct {
	goals []*entity.Goal
}

func (s stubCatchUpGoalReader) FindByWorkspaceID(workspaceID string) ([]*entity.Goal, error) {
	return s.goals, nil
}

type recordingGoalEndDateUpdater struct {
	calls              map[string][]time.Time
	postponeCountCalls map[string][]int
}

func (r *recordingGoalEndDateUpdater) UpdatePostponement(id string, endDate time.Time, postponeCount int) error {
	if r.calls == nil {
		r.calls = make(map[string][]time.Time)
		r.postponeCountCalls = make(map[string][]int)
	}
	r.calls[id] = append(r.calls[id], endDate)
	r.postponeCountCalls[id] = append(r.postponeCountCalls[id], postponeCount)
	return nil
}

// queueMissedTaskFinder simulates a repository that has a fixed-size backlog
// of pending tasks per goal; each call to FindOldestPendingBefore pops the
// next one, returning nil once the queue is empty.
type queueMissedTaskFinder struct {
	remaining map[string]int
}

func (q *queueMissedTaskFinder) FindOldestPendingBefore(goalID string, before time.Time) (*entity.DailyTask, error) {
	if q.remaining[goalID] <= 0 {
		return nil, nil
	}
	q.remaining[goalID]--
	return entity.NewDailyTask("task-"+goalID, goalID, before.AddDate(0, 0, -1), "content", before), nil
}

type recordingShifter struct {
	calls []string
}

func (r *recordingShifter) ShiftPendingForward(goalID string, fromDate time.Time) error {
	r.calls = append(r.calls, goalID)
	return nil
}

func TestCatchUpMissedTasksUsecase_Execute(t *testing.T) {
	today := time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC)

	strictGoal := entity.NewGoal("goal-strict", "workspace-001", "Run", "detail", "cond", today, entity.ModeStrict, today)
	wantGoal := entity.NewGoal("goal-want", "workspace-001", "Read", "detail", "cond", today, entity.ModeWant, today)

	goals := stubCatchUpGoalReader{goals: []*entity.Goal{strictGoal, wantGoal}}
	finder := &queueMissedTaskFinder{remaining: map[string]int{"goal-strict": 2, "goal-want": 3}}
	shifter := &recordingShifter{}
	goalUpdater := &recordingGoalEndDateUpdater{}

	u := NewCatchUpMissedTasksUsecase(goals, goalUpdater, finder, shifter, stubClock{now: today})

	if err := u.Execute("workspace-001"); err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}

	if len(shifter.calls) != 2 {
		t.Fatalf("shifter called %d times, want 2 (once per missed day on the strict goal)", len(shifter.calls))
	}
	for _, id := range shifter.calls {
		if id != "goal-strict" {
			t.Errorf("shifter called for goal %q, want only goal-strict (want-mode goals must be skipped)", id)
		}
	}

	if calls := goalUpdater.calls["goal-strict"]; len(calls) != 2 {
		t.Fatalf("UpdatePostponement called %d times for goal-strict, want 2", len(calls))
	} else {
		want := today.AddDate(0, 0, 2)
		if !calls[len(calls)-1].Equal(want) {
			t.Errorf("final EndDate = %v, want %v (extended by 2 days for 2 missed days)", calls[len(calls)-1], want)
		}
	}

	if counts := goalUpdater.postponeCountCalls["goal-strict"]; len(counts) != 2 || counts[len(counts)-1] != 2 {
		t.Errorf("final postponeCount = %v, want last call to be 2", counts)
	}

	if _, ok := goalUpdater.calls["goal-want"]; ok {
		t.Errorf("UpdatePostponement was called for goal-want, want-mode goals must never be postponed")
	}
}
