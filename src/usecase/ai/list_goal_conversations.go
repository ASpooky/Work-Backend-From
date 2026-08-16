package ai

import "github.com/ASpooky/Work-Backend-From/src/entity"

type GoalConversationLister interface {
	FindByGoalID(goalID string) ([]*entity.Conversation, error)
}

type ListGoalConversationsUsecase struct {
	repo GoalConversationLister
}

func NewListGoalConversationsUsecase(repo GoalConversationLister) *ListGoalConversationsUsecase {
	return &ListGoalConversationsUsecase{repo: repo}
}

func (u *ListGoalConversationsUsecase) Execute(goalID string) ([]*entity.Conversation, error) {
	return u.repo.FindByGoalID(goalID)
}
