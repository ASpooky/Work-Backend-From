package ai

import "github.com/ASpooky/Work-Backend-From/src/entity"

type ConversationLister interface {
	FindByWorkspaceID(workspaceID string) ([]*entity.Conversation, error)
}

type ListConversationsUsecase struct {
	repo ConversationLister
}

func NewListConversationsUsecase(repo ConversationLister) *ListConversationsUsecase {
	return &ListConversationsUsecase{repo: repo}
}

func (u *ListConversationsUsecase) Execute(workspaceID string) ([]*entity.Conversation, error) {
	return u.repo.FindByWorkspaceID(workspaceID)
}
