package ai

import (
	"context"
	"fmt"

	"github.com/ASpooky/Work-Backend-From/src/entity"
	"github.com/ASpooky/Work-Backend-From/src/usecase"
)

const chatSystemPromptTemplate = `あなたはユーザーの目標設定を手伝うコーチです。今日の日付は %s です。
ユーザーと対話しながら、目標の強度(必達/努力目標)と達成期限を一緒に明確にしてください。
「3ヶ月後」のような相対的な期限が出てきたら、今日の日付を基準に具体的な日付に置き換えて
確認してください。曖昧な点があれば質問し、達成条件も具体的にすり合わせてください。
会話の内容が固まってきたら、「この内容でタスクに落とし込めそうです」とユーザーに伝えてください。
回答は日本語、簡潔に。`

type ChatCompleter interface {
	Chat(ctx context.Context, systemInstruction string, messages []entity.ChatMessage) (string, error)
}

type ChatInput struct {
	Messages []entity.ChatMessage
	// GoalID is ignored by ChatUsecase (the new-goal-creation coach); it
	// only matters to GoalReviewChatUsecase, which shares this same input
	// shape so both can be injected into SendMessageUsecase as a Chatter.
	GoalID string
}

type ChatUsecase struct {
	ai    ChatCompleter
	clock usecase.Clock
}

func NewChatUsecase(ai ChatCompleter, clock usecase.Clock) *ChatUsecase {
	return &ChatUsecase{ai: ai, clock: clock}
}

func (u *ChatUsecase) Execute(ctx context.Context, input ChatInput) (string, error) {
	prompt := fmt.Sprintf(chatSystemPromptTemplate, u.clock.Now().Format(planDateLayout))
	return u.ai.Chat(ctx, prompt, input.Messages)
}
