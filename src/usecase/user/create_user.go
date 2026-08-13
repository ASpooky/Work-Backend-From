package user

import (
	"github.com/ASpooky/Work-Backend-From/src/entity"
	"github.com/ASpooky/Work-Backend-From/src/usecase"
)

type CreateUserInput struct {
	Name string
}

type Repository interface {
	Save(user *entity.User) error
}

type CreateUserUsecase struct {
	repo  Repository
	idGen usecase.IDGenerator
	clock usecase.Clock
}

func NewCreateUserUsecase(repo Repository, idGen usecase.IDGenerator, clock usecase.Clock) *CreateUserUsecase {
	return &CreateUserUsecase{repo: repo, idGen: idGen, clock: clock}
}

func (u *CreateUserUsecase) Execute(input CreateUserInput) (*entity.User, error) {
	user := entity.NewUser(u.idGen.NewID(), input.Name, u.clock.Now())

	if err := u.repo.Save(user); err != nil {
		return nil, err
	}

	return user, nil
}
