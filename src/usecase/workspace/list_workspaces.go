package workspace

import "github.com/ASpooky/Work-Backend-From/src/entity"

type ListWorkspacesUsecase struct {
	repo Repository
}

func NewListWorkspacesUsecase(repo Repository) *ListWorkspacesUsecase {
	return &ListWorkspacesUsecase{repo: repo}
}

func (u *ListWorkspacesUsecase) Execute() ([]*entity.WorkSpace, error) {
	return u.repo.FindAll()
}
