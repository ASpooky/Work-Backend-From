package goal

import "github.com/ASpooky/Work-Backend-From/src/entity"

type ListGoalsInput struct {
	WorkspaceID string
}

type ListGoalsUsecase struct {
	repo Repository
}

func NewListGoalsUsecase(repo Repository) *ListGoalsUsecase {
	return &ListGoalsUsecase{repo: repo}
}

func (u *ListGoalsUsecase) Execute(input ListGoalsInput) ([]*entity.Goal, error) {
	return u.repo.FindByWorkspaceID(input.WorkspaceID)
}
