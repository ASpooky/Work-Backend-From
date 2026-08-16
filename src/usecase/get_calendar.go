package usecase

import (
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
)

const calendarDateLayout = "2006-01-02"

type DayStatus string

const (
	DayStatusNoTask  DayStatus = "no_task"
	DayStatusNotDone DayStatus = "not_done"
	DayStatusDone    DayStatus = "done"
)

type DayEntry struct {
	Status  DayStatus `json:"status"`
	Content string    `json:"content"`
}

type GoalCalendar struct {
	Goal *entity.Goal        `json:"goal"`
	Days map[string]DayEntry `json:"days"`
}

type GetCalendarInput struct {
	WorkspaceID string
	From        time.Time
	To          time.Time
}

type GoalReader interface {
	FindByWorkspaceID(workspaceID string) ([]*entity.Goal, error)
	FindAll() ([]*entity.Goal, error)
}

type DailyTaskRangeReader interface {
	FindByGoalIDAndDateRange(goalID string, from, to time.Time) ([]*entity.DailyTask, error)
}

type GetCalendarUsecase struct {
	goals GoalReader
	tasks DailyTaskRangeReader
}

func NewGetCalendarUsecase(goals GoalReader, tasks DailyTaskRangeReader) *GetCalendarUsecase {
	return &GetCalendarUsecase{goals: goals, tasks: tasks}
}

func (u *GetCalendarUsecase) Execute(input GetCalendarInput) ([]GoalCalendar, error) {
	var goals []*entity.Goal
	var err error
	if input.WorkspaceID == "" {
		goals, err = u.goals.FindAll()
	} else {
		goals, err = u.goals.FindByWorkspaceID(input.WorkspaceID)
	}
	if err != nil {
		return nil, err
	}

	result := make([]GoalCalendar, 0, len(goals))
	for _, g := range goals {
		tasks, err := u.tasks.FindByGoalIDAndDateRange(g.ID, input.From, input.To)
		if err != nil {
			return nil, err
		}

		byDate := make(map[string][]*entity.DailyTask)
		for _, t := range tasks {
			key := t.Date.Format(calendarDateLayout)
			byDate[key] = append(byDate[key], t)
		}

		days := make(map[string]DayEntry)
		for d := input.From; !d.After(input.To); d = d.AddDate(0, 0, 1) {
			key := d.Format(calendarDateLayout)
			days[key] = deriveDayEntry(byDate[key])
		}

		result = append(result, GoalCalendar{Goal: g, Days: days})
	}

	return result, nil
}

func deriveDayEntry(tasks []*entity.DailyTask) DayEntry {
	if len(tasks) == 0 {
		return DayEntry{Status: DayStatusNoTask}
	}

	content := tasks[0].Content
	for _, t := range tasks {
		if !t.Done {
			return DayEntry{Status: DayStatusNotDone, Content: content}
		}
	}

	return DayEntry{Status: DayStatusDone, Content: content}
}
