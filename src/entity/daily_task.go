package entity

import "time"

type DailyTask struct {
	ID          string     `json:"id"`
	GoalID      string     `json:"goal_id"`
	Date        time.Time  `json:"date"`
	Content     string     `json:"content"`
	Done        bool       `json:"done"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	// Memo is a free-text note the user can attach/edit at any time,
	// independent of Done/CompletedAt — e.g. "腰が痛かった" on a workout day.
	// Feeds SummarizeGoalUsecase's AI prompt so it has more than raw
	// done/scheduled counts to comment on.
	Memo *string `json:"memo,omitempty"`
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
