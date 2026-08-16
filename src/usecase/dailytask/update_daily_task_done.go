package dailytask

import (
	"time"

	"github.com/ASpooky/Work-Backend-From/src/usecase"
)

type DoneUpdater interface {
	UpdateDone(id string, done bool, completedAt *time.Time) error
}

type UpdateDailyTaskDoneInput struct {
	ID   string
	Done bool
}

type UpdateDailyTaskDoneUsecase struct {
	repo  DoneUpdater
	clock usecase.Clock
}

func NewUpdateDailyTaskDoneUsecase(repo DoneUpdater, clock usecase.Clock) *UpdateDailyTaskDoneUsecase {
	return &UpdateDailyTaskDoneUsecase{repo: repo, clock: clock}
}

// Execute records not just whether a task is done but when — re-toggling a
// task back to not-done clears CompletedAt rather than leaving a stale
// timestamp, so it always reflects the most recent transition.
func (u *UpdateDailyTaskDoneUsecase) Execute(input UpdateDailyTaskDoneInput) error {
	var completedAt *time.Time
	if input.Done {
		now := u.clock.Now()
		completedAt = &now
	}
	return u.repo.UpdateDone(input.ID, input.Done, completedAt)
}
