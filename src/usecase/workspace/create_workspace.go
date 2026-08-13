package workspace

import (
	"github.com/ASpooky/Work-Backend-From/src/entity"
	"github.com/ASpooky/Work-Backend-From/src/usecase"
)

type CreateWorkspaceInput struct {
	UserID string
	Name   string
}

type Repository interface {
	Save(workspace *entity.WorkSpace) error
}

type CreateWorkspaceUsecase struct {
	repo  Repository
	idGen usecase.IDGenerator
	clock usecase.Clock
}

func NewCreateWorkspaceUsecase(repo Repository, idGen usecase.IDGenerator, clock usecase.Clock) *CreateWorkspaceUsecase {
	return &CreateWorkspaceUsecase{repo: repo, idGen: idGen, clock: clock}
}

func (u *CreateWorkspaceUsecase) Execute(input CreateWorkspaceInput) (*entity.WorkSpace, error) {
	workspace := entity.NewWorkSpace(u.idGen.NewID(), input.UserID, input.Name, u.clock.Now())

	if err := u.repo.Save(workspace); err != nil {
		return nil, err
	}

	return workspace, nil
}
