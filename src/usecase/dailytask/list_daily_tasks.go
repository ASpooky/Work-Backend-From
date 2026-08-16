package dailytask

import (
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
)

type ListDailyTasksInput struct {
	Date        time.Time
	WorkspaceID string // empty means all workspaces
}

type ListDailyTasksUsecase struct {
	repo Repository
}

func NewListDailyTasksUsecase(repo Repository) *ListDailyTasksUsecase {
	return &ListDailyTasksUsecase{repo: repo}
}

func (u *ListDailyTasksUsecase) Execute(input ListDailyTasksInput) ([]*entity.DailyTask, error) {
	if input.WorkspaceID == "" {
		return u.repo.FindByDate(input.Date)
	}
	return u.repo.FindByDateAndWorkspaceID(input.Date, input.WorkspaceID)
}
