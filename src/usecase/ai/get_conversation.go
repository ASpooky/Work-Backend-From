package ai

import "github.com/ASpooky/Work-Backend-From/src/entity"

type GetConversationUsecase struct {
	messages ConversationMessageReader
}

func NewGetConversationUsecase(messages ConversationMessageReader) *GetConversationUsecase {
	return &GetConversationUsecase{messages: messages}
}

func (u *GetConversationUsecase) Execute(conversationID string) ([]*entity.ConversationMessage, error) {
	return u.messages.FindByConversationID(conversationID)
}
