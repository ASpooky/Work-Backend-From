package dailytask

import (
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
	"github.com/ASpooky/Work-Backend-From/src/usecase"
)

type CreateDailyTaskInput struct {
	GoalID  string
	Date    time.Time
	Content string
}

type Repository interface {
	Save(task *entity.DailyTask) error
}

type CreateDailyTaskUsecase struct {
	repo  Repository
	idGen usecase.IDGenerator
	clock usecase.Clock
}

func NewCreateDailyTaskUsecase(repo Repository, idGen usecase.IDGenerator, clock usecase.Clock) *CreateDailyTaskUsecase {
	return &CreateDailyTaskUsecase{repo: repo, idGen: idGen, clock: clock}
}

func (u *CreateDailyTaskUsecase) Execute(input CreateDailyTaskInput) (*entity.DailyTask, error) {
	task := entity.NewDailyTask(u.idGen.NewID(), input.GoalID, input.Date, input.Content, u.clock.Now())

	if err := u.repo.Save(task); err != nil {
		return nil, err
	}

	return task, nil
}
