package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
	"github.com/ASpooky/Work-Backend-From/src/usecase"
	"github.com/ASpooky/Work-Backend-From/src/usecase/dailytask"
)

const planDateLayout = "2006-01-02"

const planSystemPromptTemplate = `あなたはユーザーの目標達成を支援するプランナーです。今日の日付は %s です。
これまでの会話の内容をもとに、目標(goals)と、それぞれを確実に達成するための日々のタスク(tasks)を
提案してください。タスクは無理のない頻度にし、対応するgoalの達成条件と矛盾しないようにしてください。
「3ヶ月後」のような相対的な期限は、今日の日付を基準に計算した実際の日付にしてください。
過去の日付を出力してはいけません。日付はすべて YYYY-MM-DD 形式で出力してください。

会話の内容が、性質の異なる複数の段階（フェーズ）に分けたほうが自然な場合は、
goalsに複数の目標を、時系列順に並べて提案してください(例: 「土台作り期」→「追い込み期」のように、
それぞれ異なる達成条件・期限・タスクを持つ別の目標として)。単一の目標で十分な内容であれば、
goalsには1件だけを含めてください。無理に分割する必要はありません。`

const planRequestMessage = "これまでの内容をもとに、目標と日々のタスクの提案を出力してください。"

var planTaskSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"content":       map[string]any{"type": "string"},
		"start_date":    map[string]any{"type": "string", "description": "YYYY-MM-DD"},
		"end_date":      map[string]any{"type": "string", "description": "YYYY-MM-DD"},
		"rule_type":     map[string]any{"type": "string", "enum": []string{"interval", "weekly"}},
		"interval_days": map[string]any{"type": "integer"},
		"weekdays":      map[string]any{"type": "array", "items": map[string]any{"type": "integer"}},
	},
	"required": []string{"content", "start_date", "end_date", "rule_type"},
}

var planSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"goals": map[string]any{
			"type": "array",
			"items": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"title":                 map[string]any{"type": "string"},
					"detail":                map[string]any{"type": "string"},
					"achievement_condition": map[string]any{"type": "string"},
					"end_date":              map[string]any{"type": "string", "description": "YYYY-MM-DD"},
					"mode":                  map[string]any{"type": "string", "enum": []string{"strict", "want"}},
					"tasks":                 map[string]any{"type": "array", "items": planTaskSchema},
				},
				"required": []string{"title", "detail", "achievement_condition", "end_date", "mode", "tasks"},
			},
		},
	},
	"required": []string{"goals"},
}

type PlanGenerator interface {
	GenerateJSON(ctx context.Context, systemInstruction string, messages []entity.ChatMessage, schema map[string]any) (string, error)
}

type ConversationMessageReader interface {
	FindByConversationID(conversationID string) ([]*entity.ConversationMessage, error)
}

type PlanGoalInput struct {
	ConversationID string
}

type PlannedGoal struct {
	Title                string
	Detail               string
	AchievementCondition string
	EndDate              time.Time
	Mode                 entity.GoalMode
}

type PlannedTask struct {
	Content   string
	StartDate time.Time
	EndDate   time.Time
	Rule      dailytask.RecurrenceRule
}

// PlannedGoalPlan pairs one proposed goal with its own tasks — a Plan holds
// one or more of these, since a long-running goal may naturally split into
// sequential phases (each with its own achievement condition and deadline)
// rather than a single goal trying to cover the whole arc.
type PlannedGoalPlan struct {
	Goal  PlannedGoal
	Tasks []PlannedTask
}

type Plan struct {
	Goals []PlannedGoalPlan
}

type planTaskWire struct {
	Content      string `json:"content"`
	StartDate    string `json:"start_date"`
	EndDate      string `json:"end_date"`
	RuleType     string `json:"rule_type"`
	IntervalDays int    `json:"interval_days"`
	Weekdays     []int  `json:"weekdays"`
}

type planGoalWire struct {
	Title                string         `json:"title"`
	Detail               string         `json:"detail"`
	AchievementCondition string         `json:"achievement_condition"`
	EndDate              string         `json:"end_date"`
	Mode                 string         `json:"mode"`
	Tasks                []planTaskWire `json:"tasks"`
}

type planWire struct {
	Goals []planGoalWire `json:"goals"`
}

type PlanGoalUsecase struct {
	ai       PlanGenerator
	messages ConversationMessageReader
	clock    usecase.Clock
}

func NewPlanGoalUsecase(ai PlanGenerator, messages ConversationMessageReader, clock usecase.Clock) *PlanGoalUsecase {
	return &PlanGoalUsecase{ai: ai, messages: messages, clock: clock}
}

func (u *PlanGoalUsecase) Execute(ctx context.Context, input PlanGoalInput) (Plan, error) {
	history, err := u.messages.FindByConversationID(input.ConversationID)
	if err != nil {
		return Plan{}, err
	}

	prompt := fmt.Sprintf(planSystemPromptTemplate, u.clock.Now().Format(planDateLayout))

	// The Gemini API rejects a `contents` array whose last turn has role
	// "model" ("Requests ending with a model turn are not supported."),
	// which the conversation always does at this point (it just replied).
	// Append an explicit trailing user turn requesting the plan.
	messages := make([]entity.ChatMessage, 0, len(history)+1)
	for _, m := range history {
		messages = append(messages, entity.ChatMessage{Role: m.Role, Content: m.Content})
	}
	messages = append(messages, entity.ChatMessage{Role: entity.ChatRoleUser, Content: planRequestMessage})

	raw, err := u.ai.GenerateJSON(ctx, prompt, messages, planSchema)
	if err != nil {
		return Plan{}, err
	}

	var wire planWire
	if err := json.Unmarshal([]byte(raw), &wire); err != nil {
		return Plan{}, fmt.Errorf("failed to parse AI plan response: %w", err)
	}
	if len(wire.Goals) == 0 {
		return Plan{}, fmt.Errorf("AI plan response contained no goals")
	}

	plan := Plan{Goals: make([]PlannedGoalPlan, 0, len(wire.Goals))}
	for _, g := range wire.Goals {
		goalEndDate, err := time.Parse(planDateLayout, g.EndDate)
		if err != nil {
			return Plan{}, fmt.Errorf("invalid goal end_date %q from AI: %w", g.EndDate, err)
		}

		planned := PlannedGoalPlan{
			Goal: PlannedGoal{
				Title:                g.Title,
				Detail:               g.Detail,
				AchievementCondition: g.AchievementCondition,
				EndDate:              goalEndDate,
				Mode:                 entity.GoalMode(g.Mode),
			},
		}

		for _, t := range g.Tasks {
			start, err := time.Parse(planDateLayout, t.StartDate)
			if err != nil {
				return Plan{}, fmt.Errorf("invalid task start_date %q from AI: %w", t.StartDate, err)
			}
			end, err := time.Parse(planDateLayout, t.EndDate)
			if err != nil {
				return Plan{}, fmt.Errorf("invalid task end_date %q from AI: %w", t.EndDate, err)
			}

			rule := dailytask.RecurrenceRule{Type: dailytask.RecurrenceType(t.RuleType), IntervalDays: t.IntervalDays}
			for _, wd := range t.Weekdays {
				rule.Weekdays = append(rule.Weekdays, time.Weekday(wd))
			}

			planned.Tasks = append(planned.Tasks, PlannedTask{
				Content:   t.Content,
				StartDate: start,
				EndDate:   end,
				Rule:      rule,
			})
		}

		plan.Goals = append(plan.Goals, planned)
	}

	return plan, nil
}
