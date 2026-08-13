package goal

import (
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
	"github.com/ASpooky/Work-Backend-From/src/usecase"
)

type CreateGoalInput struct {
	WorkspaceID          string
	Title                string
	Detail               string
	AchievementCondition string
	EndDate              time.Time
	Mode                 entity.GoalMode
}

type Repository interface {
	Save(goal *entity.Goal) error
}

type CreateGoalUsecase struct {
	repo  Repository
	idGen usecase.IDGenerator
	clock usecase.Clock
}

func NewCreateGoalUsecase(repo Repository, idGen usecase.IDGenerator, clock usecase.Clock) *CreateGoalUsecase {
	return &CreateGoalUsecase{repo: repo, idGen: idGen, clock: clock}
}

func (u *CreateGoalUsecase) Execute(input CreateGoalInput) (*entity.Goal, error) {
	goal := entity.NewGoal(
		u.idGen.NewID(),
		input.WorkspaceID,
		input.Title,
		input.Detail,
		input.AchievementCondition,
		input.EndDate,
		input.Mode,
		u.clock.Now(),
	)

	if err := u.repo.Save(goal); err != nil {
		return nil, err
	}

	return goal, nil
}
