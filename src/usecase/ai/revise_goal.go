package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
	"github.com/ASpooky/Work-Backend-From/src/usecase"
)

const reviseGoalSystemPromptTemplate = `あなたはユーザーの目標達成を支援するコーチです。今日の日付は %s です。
これまでの目標見直しの会話をもとに、更新後の目標(goal)を提案してください。
会話の中で変更の言及がなかった項目は、現在の値をそのまま使ってください。
現在の値: タイトル=%s / 詳細=%s / 達成条件=%s / 期限=%s / モード=%s
「1ヶ月延長」のような相対的な期限は、現在の期限を基準に計算した実際の日付にしてください。
過去の日付を出力してはいけません。日付はすべて YYYY-MM-DD 形式で出力してください。`

const reviseGoalRequestMessage = "これまでの内容をもとに、見直し後の目標案を出力してください。"

var reviseGoalSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"title":                 map[string]any{"type": "string"},
		"detail":                map[string]any{"type": "string"},
		"achievement_condition": map[string]any{"type": "string"},
		"end_date":              map[string]any{"type": "string", "description": "YYYY-MM-DD"},
		"mode":                  map[string]any{"type": "string", "enum": []string{"strict", "want"}},
	},
	"required": []string{"title", "detail", "achievement_condition", "end_date", "mode"},
}

type ReviseGoalInput struct {
	GoalID         string
	ConversationID string
}

type reviseGoalWire struct {
	Title                string `json:"title"`
	Detail               string `json:"detail"`
	AchievementCondition string `json:"achievement_condition"`
	EndDate              string `json:"end_date"`
	Mode                 string `json:"mode"`
}

// ReviseGoalUsecase mirrors PlanGoalUsecase but proposes an update to an
// existing goal instead of a brand-new one, so it returns a bare
// PlannedGoal with no tasks.
type ReviseGoalUsecase struct {
	ai       PlanGenerator
	goals    SingleGoalReader
	messages ConversationMessageReader
	clock    usecase.Clock
}

func NewReviseGoalUsecase(ai PlanGenerator, goals SingleGoalReader, messages ConversationMessageReader, clock usecase.Clock) *ReviseGoalUsecase {
	return &ReviseGoalUsecase{ai: ai, goals: goals, messages: messages, clock: clock}
}

func (u *ReviseGoalUsecase) Execute(ctx context.Context, input ReviseGoalInput) (PlannedGoal, error) {
	goal, err := u.goals.FindByID(input.GoalID)
	if err != nil {
		return PlannedGoal{}, err
	}
	if goal == nil {
		return PlannedGoal{}, fmt.Errorf("goal not found: %s", input.GoalID)
	}

	history, err := u.messages.FindByConversationID(input.ConversationID)
	if err != nil {
		return PlannedGoal{}, err
	}

	prompt := fmt.Sprintf(reviseGoalSystemPromptTemplate,
		u.clock.Now().Format(planDateLayout),
		goal.Title, goal.Detail, goal.AchievementCondition, goal.EndDate.Format(planDateLayout), goal.Mode,
	)

	// Same Gemini constraint as PlanGoalUsecase: a conversation always ends
	// on the AI's last reply, and generateContent rejects that.
	messages := make([]entity.ChatMessage, 0, len(history)+1)
	for _, m := range history {
		messages = append(messages, entity.ChatMessage{Role: m.Role, Content: m.Content})
	}
	messages = append(messages, entity.ChatMessage{Role: entity.ChatRoleUser, Content: reviseGoalRequestMessage})

	raw, err := u.ai.GenerateJSON(ctx, prompt, messages, reviseGoalSchema)
	if err != nil {
		return PlannedGoal{}, err
	}

	var wire reviseGoalWire
	if err := json.Unmarshal([]byte(raw), &wire); err != nil {
		return PlannedGoal{}, fmt.Errorf("failed to parse AI revision response: %w", err)
	}

	endDate, err := time.Parse(planDateLayout, wire.EndDate)
	if err != nil {
		return PlannedGoal{}, fmt.Errorf("invalid end_date %q from AI: %w", wire.EndDate, err)
	}

	return PlannedGoal{
		Title:                wire.Title,
		Detail:               wire.Detail,
		AchievementCondition: wire.AchievementCondition,
		EndDate:              endDate,
		Mode:                 entity.GoalMode(wire.Mode),
	}, nil
}
