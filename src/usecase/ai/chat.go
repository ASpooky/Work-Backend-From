package ai

import (
	"context"

	"github.com/ASpooky/Work-Backend-From/src/entity"
)

const chatSystemPrompt = `あなたはユーザーの目標設定を手伝うコーチです。ユーザーと対話しながら、
目標の強度(必達/努力目標)と達成期限を一緒に明確にしてください。曖昧な点があれば
質問し、達成条件も具体的にすり合わせてください。会話の内容が固まってきたら、
「この内容でタスクに落とし込めそうです」とユーザーに伝えてください。
回答は日本語、簡潔に。`

type ChatCompleter interface {
	Chat(ctx context.Context, systemInstruction string, messages []entity.ChatMessage) (string, error)
}

type ChatInput struct {
	Messages []entity.ChatMessage
}

type ChatUsecase struct {
	ai ChatCompleter
}

func NewChatUsecase(ai ChatCompleter) *ChatUsecase {
	return &ChatUsecase{ai: ai}
}

func (u *ChatUsecase) Execute(ctx context.Context, input ChatInput) (string, error) {
	return u.ai.Chat(ctx, chatSystemPrompt, input.Messages)
}
