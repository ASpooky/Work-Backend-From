package entity

import (
	"reflect"
	"testing"
	"time"
)

func TestNewDailyTask(t *testing.T) {
	createdAt := time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)
	date := time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)

	got := NewDailyTask("task-001", "goal-001", date, "Run 5km", createdAt)

	want := &DailyTask{
		ID:        "task-001",
		GoalID:    "goal-001",
		Date:      date,
		Content:   "Run 5km",
		Done:      false,
		CreatedAt: createdAt,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("NewDailyTask() = %+v, want %+v", got, want)
	}
}
