package entity

import "time"

type DailyTask struct {
	ID        string    `json:"id"`
	GoalID    string    `json:"goal_id"`
	Date      time.Time `json:"date"`
	Content   string    `json:"content"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
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
