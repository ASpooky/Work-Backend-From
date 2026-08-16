package ai

import (
	"context"
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
	"github.com/ASpooky/Work-Backend-From/src/usecase"
)

const conversationTitleMaxRunes = 30

type ConversationCreator interface {
	Save(c *entity.Conversation) error
}

type ConversationToucher interface {
	Touch(id string, updatedAt time.Time) error
}

type MessageStore interface {
	Save(m *entity.ConversationMessage) error
	FindByConversationID(conversationID string) ([]*entity.ConversationMessage, error)
}

type SendMessageInput struct {
	WorkspaceID    string
	ConversationID string // empty starts a new conversation
	Content        string
}

type SendMessageOutput struct {
	ConversationID string
	Reply          string
}

// SendMessageUsecase is the persisted counterpart to ChatUsecase: it owns
// conversation lifecycle (create-or-continue) and durably records both
// sides of the exchange, so a 壁打ち thread survives across sessions
// instead of living only in frontend state.
type SendMessageUsecase struct {
	conversations ConversationCreator
	toucher       ConversationToucher
	messages      MessageStore
	chat          *ChatUsecase
	idGen         usecase.IDGenerator
	clock         usecase.Clock
}

func NewSendMessageUsecase(
	conversations ConversationCreator,
	toucher ConversationToucher,
	messages MessageStore,
	chat *ChatUsecase,
	idGen usecase.IDGenerator,
	clock usecase.Clock,
) *SendMessageUsecase {
	return &SendMessageUsecase{
		conversations: conversations,
		toucher:       toucher,
		messages:      messages,
		chat:          chat,
		idGen:         idGen,
		clock:         clock,
	}
}

func (u *SendMessageUsecase) Execute(ctx context.Context, input SendMessageInput) (SendMessageOutput, error) {
	conversationID := input.ConversationID
	var history []*entity.ConversationMessage

	if conversationID == "" {
		conversationID = u.idGen.NewID()
		conv := entity.NewConversation(conversationID, input.WorkspaceID, deriveTitle(input.Content), u.clock.Now())
		if err := u.conversations.Save(conv); err != nil {
			return SendMessageOutput{}, err
		}
	} else {
		existing, err := u.messages.FindByConversationID(conversationID)
		if err != nil {
			return SendMessageOutput{}, err
		}
		history = existing
	}

	userMsg := entity.NewConversationMessage(u.idGen.NewID(), conversationID, entity.ChatRoleUser, input.Content, u.clock.Now())
	if err := u.messages.Save(userMsg); err != nil {
		return SendMessageOutput{}, err
	}

	chatMessages := make([]entity.ChatMessage, 0, len(history)+1)
	for _, m := range history {
		chatMessages = append(chatMessages, entity.ChatMessage{Role: m.Role, Content: m.Content})
	}
	chatMessages = append(chatMessages, entity.ChatMessage{Role: entity.ChatRoleUser, Content: input.Content})

	reply, err := u.chat.Execute(ctx, ChatInput{Messages: chatMessages})
	if err != nil {
		return SendMessageOutput{}, err
	}

	modelMsg := entity.NewConversationMessage(u.idGen.NewID(), conversationID, entity.ChatRoleModel, reply, u.clock.Now())
	if err := u.messages.Save(modelMsg); err != nil {
		return SendMessageOutput{}, err
	}

	if err := u.toucher.Touch(conversationID, u.clock.Now()); err != nil {
		return SendMessageOutput{}, err
	}

	return SendMessageOutput{ConversationID: conversationID, Reply: reply}, nil
}

func deriveTitle(firstMessage string) string {
	runes := []rune(firstMessage)
	if len(runes) <= conversationTitleMaxRunes {
		return firstMessage
	}
	return string(runes[:conversationTitleMaxRunes]) + "…"
}
