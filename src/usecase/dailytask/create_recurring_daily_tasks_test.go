package dailytask

import (
	"testing"
	"time"
)

func TestCreateRecurringDailyTasksUsecase_Execute_Interval(t *testing.T) {
	repo := &recordingRepository{}
	uc := NewCreateRecurringDailyTasksUsecase(repo, stubIDGenerator{id: "task"}, stubClock{now: time.Now()})

	start := time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 6) // 7-day window

	got, err := uc.Execute(CreateRecurringDailyTasksInput{
		GoalID:    "goal-001",
		Content:   "Run 5km",
		StartDate: start,
		EndDate:   end,
		Rule:      RecurrenceRule{Type: RecurrenceInterval, IntervalDays: 2},
	})
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}

	wantDates := []time.Time{start, start.AddDate(0, 0, 2), start.AddDate(0, 0, 4), start.AddDate(0, 0, 6)}
	if len(got) != len(wantDates) {
		t.Fatalf("Execute() returned %d tasks, want %d", len(got), len(wantDates))
	}
	for i, task := range got {
		if !task.Date.Equal(wantDates[i]) {
			t.Errorf("task[%d].Date = %v, want %v", i, task.Date, wantDates[i])
		}
	}
	if len(repo.saved) != len(wantDates) {
		t.Errorf("repository.Save called %d times, want %d", len(repo.saved), len(wantDates))
	}
}

func TestCreateRecurringDailyTasksUsecase_Execute_Weekly(t *testing.T) {
	repo := &recordingRepository{}
	uc := NewCreateRecurringDailyTasksUsecase(repo, stubIDGenerator{id: "task"}, stubClock{now: time.Now()})

	start := time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC)
	secondWeekday := start.AddDate(0, 0, 2)
	end := start.AddDate(0, 0, 13) // two-week window

	got, err := uc.Execute(CreateRecurringDailyTasksInput{
		GoalID:    "goal-001",
		Content:   "Read 30 minutes",
		StartDate: start,
		EndDate:   end,
		Rule:      RecurrenceRule{Type: RecurrenceWeekly, Weekdays: []time.Weekday{start.Weekday(), secondWeekday.Weekday()}},
	})
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}

	wantDates := []time.Time{
		start,
		secondWeekday,
		start.AddDate(0, 0, 7),
		start.AddDate(0, 0, 9),
	}
	if len(got) != len(wantDates) {
		t.Fatalf("Execute() returned %d tasks, want %d", len(got), len(wantDates))
	}
	for i, task := range got {
		if !task.Date.Equal(wantDates[i]) {
			t.Errorf("task[%d].Date = %v, want %v", i, task.Date, wantDates[i])
		}
	}
}
