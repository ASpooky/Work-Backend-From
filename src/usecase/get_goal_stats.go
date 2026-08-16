package usecase

import "github.com/ASpooky/Work-Backend-From/src/entity"

type GoalStats struct {
	ScheduledCount  int
	DoneCount       int
	AchievementRate float64
	PostponeCount   int
	DaysRemaining   int
}

type SingleGoalReader interface {
	FindByID(id string) (*entity.Goal, error)
}

// GetGoalStatsUsecase computes plain, AI-free numbers for a goal's detail
// page: how much of its (already-due) schedule has actually been done,
// how many times it's been postponed, and how many days remain. Only tasks
// dated on or before "now" count toward ScheduledCount/DoneCount — a task
// that hasn't happened yet isn't a missed one.
type GetGoalStatsUsecase struct {
	goals SingleGoalReader
	tasks DailyTaskRangeReader
	clock Clock
}

func NewGetGoalStatsUsecase(goals SingleGoalReader, tasks DailyTaskRangeReader, clock Clock) *GetGoalStatsUsecase {
	return &GetGoalStatsUsecase{goals: goals, tasks: tasks, clock: clock}
}

func (u *GetGoalStatsUsecase) Execute(goalID string) (*entity.Goal, GoalStats, error) {
	goal, err := u.goals.FindByID(goalID)
	if err != nil {
		return nil, GoalStats{}, err
	}

	now := truncateToDate(u.clock.Now())

	tasks, err := u.tasks.FindByGoalIDAndDateRange(goal.ID, goal.CreatedAt, goal.EndDate)
	if err != nil {
		return nil, GoalStats{}, err
	}

	scheduled, done := 0, 0
	for _, t := range tasks {
		if t.Date.After(now) {
			continue
		}
		scheduled++
		if t.Done {
			done++
		}
	}

	rate := 0.0
	if scheduled > 0 {
		rate = float64(done) / float64(scheduled)
	}

	daysRemaining := int(goal.EndDate.Sub(now).Hours() / 24)

	return goal, GoalStats{
		ScheduledCount:  scheduled,
		DoneCount:       done,
		AchievementRate: rate,
		PostponeCount:   goal.PostponeCount,
		DaysRemaining:   daysRemaining,
	}, nil
}
