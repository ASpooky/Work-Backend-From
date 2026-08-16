package dailytask

import (
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
	"github.com/ASpooky/Work-Backend-From/src/usecase"
)

type RecurrenceType string

const (
	// RecurrenceInterval matches every Nth day starting from StartDate
	// (IntervalDays=1 is daily, 2 is every other day, and so on).
	RecurrenceInterval RecurrenceType = "interval"
	// RecurrenceWeekly matches any day whose weekday is in Weekdays.
	RecurrenceWeekly RecurrenceType = "weekly"
)

type RecurrenceRule struct {
	Type         RecurrenceType
	IntervalDays int
	Weekdays     []time.Weekday
}

func (r RecurrenceRule) matches(d time.Time, dayIndex int) bool {
	switch r.Type {
	case RecurrenceWeekly:
		for _, wd := range r.Weekdays {
			if d.Weekday() == wd {
				return true
			}
		}
		return false
	default:
		interval := r.IntervalDays
		if interval < 1 {
			interval = 1
		}
		return dayIndex%interval == 0
	}
}

type CreateRecurringDailyTasksInput struct {
	GoalID    string
	Content   string
	StartDate time.Time
	EndDate   time.Time
	Rule      RecurrenceRule
}

type CreateRecurringDailyTasksUsecase struct {
	repo  Repository
	idGen usecase.IDGenerator
	clock usecase.Clock
}

func NewCreateRecurringDailyTasksUsecase(repo Repository, idGen usecase.IDGenerator, clock usecase.Clock) *CreateRecurringDailyTasksUsecase {
	return &CreateRecurringDailyTasksUsecase{repo: repo, idGen: idGen, clock: clock}
}

func (u *CreateRecurringDailyTasksUsecase) Execute(input CreateRecurringDailyTasksInput) ([]*entity.DailyTask, error) {
	tasks := []*entity.DailyTask{}

	dayIndex := 0
	for d := input.StartDate; !d.After(input.EndDate); d = d.AddDate(0, 0, 1) {
		if input.Rule.matches(d, dayIndex) {
			task := entity.NewDailyTask(u.idGen.NewID(), input.GoalID, d, input.Content, u.clock.Now())
			if err := u.repo.Save(task); err != nil {
				return nil, err
			}
			tasks = append(tasks, task)
		}
		dayIndex++
	}

	return tasks, nil
}
