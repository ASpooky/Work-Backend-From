package goal

import (
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
)

type UpdateGoalInput struct {
	ID                   string
	Title                string
	Detail               string
	AchievementCondition string
	EndDate              time.Time
	Mode                 entity.GoalMode
}

type Updater interface {
	Update(goal *entity.Goal) error
}

type UpdateGoalUsecase struct {
	repo Updater
}

func NewUpdateGoalUsecase(repo Updater) *UpdateGoalUsecase {
	return &UpdateGoalUsecase{repo: repo}
}

func (u *UpdateGoalUsecase) Execute(input UpdateGoalInput) error {
	goal := &entity.Goal{
		ID:                   input.ID,
		Title:                input.Title,
		Detail:               input.Detail,
		AchievementCondition: input.AchievementCondition,
		EndDate:              input.EndDate,
		Mode:                 input.Mode,
	}
	return u.repo.Update(goal)
}
