package usecase

import (
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
)

// CatchUpGoalReader is satisfied by any repository that can list a
// workspace's goals.
type CatchUpGoalReader interface {
	FindByWorkspaceID(workspaceID string) ([]*entity.Goal, error)
}

// GoalEndDateUpdater persists a goal's postponed EndDate.
type GoalEndDateUpdater interface {
	UpdateEndDate(id string, endDate time.Time) error
}

// MissedTaskFinder locates the single oldest not-done task before a cutoff.
type MissedTaskFinder interface {
	FindOldestPendingBefore(goalID string, before time.Time) (*entity.DailyTask, error)
}

// PendingTaskShifter pushes a goal's not-done tasks forward by one day.
type PendingTaskShifter interface {
	ShiftPendingForward(goalID string, fromDate time.Time) error
}

// CatchUpMissedTasksUsecase reframes a missed day as a postponement rather
// than a failure: for every strict-mode goal, each undone task left behind
// in the past pushes that goal's entire remaining schedule (and its
// deadline) forward by one day, repeated until no backlog remains.
// want-mode goals are never touched, matching entity.Goal.Postpone()'s
// existing strict-only semantics.
type CatchUpMissedTasksUsecase struct {
	goals       CatchUpGoalReader
	goalUpdater GoalEndDateUpdater
	finder      MissedTaskFinder
	shifter     PendingTaskShifter
	clock       Clock
}

func NewCatchUpMissedTasksUsecase(
	goals CatchUpGoalReader,
	goalUpdater GoalEndDateUpdater,
	finder MissedTaskFinder,
	shifter PendingTaskShifter,
	clock Clock,
) *CatchUpMissedTasksUsecase {
	return &CatchUpMissedTasksUsecase{
		goals:       goals,
		goalUpdater: goalUpdater,
		finder:      finder,
		shifter:     shifter,
		clock:       clock,
	}
}

func (u *CatchUpMissedTasksUsecase) Execute(workspaceID string) error {
	goals, err := u.goals.FindByWorkspaceID(workspaceID)
	if err != nil {
		return err
	}

	today := truncateToDate(u.clock.Now())

	for _, g := range goals {
		if g.Mode != entity.ModeStrict {
			continue
		}

		for {
			oldest, err := u.finder.FindOldestPendingBefore(g.ID, today)
			if err != nil {
				return err
			}
			if oldest == nil {
				break
			}

			if err := u.shifter.ShiftPendingForward(g.ID, oldest.Date); err != nil {
				return err
			}

			g.Postpone()
			if err := u.goalUpdater.UpdateEndDate(g.ID, g.EndDate); err != nil {
				return err
			}
		}
	}

	return nil
}

func truncateToDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
