package workspace

import (
	"errors"

	"github.com/ASpooky/Work-Backend-From/src/entity"
)

type DeleteRepository interface {
	FindAll() ([]*entity.WorkSpace, error)
	Delete(id string) error
}

type DeleteWorkspaceUsecase struct {
	repo DeleteRepository
}

func NewDeleteWorkspaceUsecase(repo DeleteRepository) *DeleteWorkspaceUsecase {
	return &DeleteWorkspaceUsecase{repo: repo}
}

func (u *DeleteWorkspaceUsecase) Execute(id string) error {
	all, err := u.repo.FindAll()
	if err != nil {
		return err
	}
	if len(all) <= 1 {
		return errors.New("cannot delete the last remaining workspace")
	}

	return u.repo.Delete(id)
}
