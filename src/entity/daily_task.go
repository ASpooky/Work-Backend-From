package entity

import "time"

type DailyTask struct {
	ID        string
	GoalID    string
	Date      time.Time
	Content   string
	Done      bool
	CreatedAt time.Time
}

func NewDailyTask(id, goalID string, date time.Time, content string, createdAt time.Time) *DailyTask {
	return &DailyTask{
		ID:        id,
		GoalID:    goalID,
		Date:      date,
		Content:   content,
		Done:      false,
		CreatedAt: createdAt,
	}
}
