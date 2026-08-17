package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
	"github.com/ASpooky/Work-Backend-From/src/usecase"
	"github.com/ASpooky/Work-Backend-From/src/usecase/dailytask"
)

const reviseGoalSystemPromptTemplate = `あなたはユーザーの目標達成を支援するコーチです。今日の日付は %s です。
これまでの目標見直しの会話をもとに、更新後の目標(goal)を提案してください。
会話の中で変更の言及がなかった項目は、現在の値をそのまま使ってください。
現在の値: タイトル=%s / 詳細=%s / 達成条件=%s / 期限=%s / モード=%s
「1ヶ月延長」のような相対的な期限は、現在の期限を基準に計算した実際の日付にしてください。
過去の日付を出力してはいけません。日付はすべて YYYY-MM-DD 形式で出力してください。

さらに、必要であれば日々のタスクの見直しも提案してください。
- 会話の中でタスクの頻度・内容・期間の変更が明確に求められた場合のみ、
  remove_task_ids(削除する既存タスクのID)とnew_tasks(追加する新しいタスク)を使ってください。
- 目標の期限や達成条件だけの変更で、タスク自体の見直しが求められていない場合は、
  remove_task_idsとnew_tasksは両方とも空配列のままにしてください。
- remove_task_idsには、下記の「見直し候補タスク」に実際に存在するIDのみを指定してください。
%s`

const reviseGoalRequestMessage = "これまでの内容をもとに、見直し後の目標案を出力してください。"

var reviseGoalSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"goal": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"title":                 map[string]any{"type": "string"},
				"detail":                map[string]any{"type": "string"},
				"achievement_condition": map[string]any{"type": "string"},
				"end_date":              map[string]any{"type": "string", "description": "YYYY-MM-DD"},
				"mode":                  map[string]any{"type": "string", "enum": []string{"strict", "want"}},
			},
			"required": []string{"title", "detail", "achievement_condition", "end_date", "mode"},
		},
		"remove_task_ids": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
		"new_tasks":       map[string]any{"type": "array", "items": planTaskSchema},
	},
	"required": []string{"goal", "remove_task_ids", "new_tasks"},
}

type ReviseGoalInput struct {
	GoalID         string
	ConversationID string
}

// GoalRevision is ReviseGoalUsecase's proposal: an updated goal, existing
// tasks proposed for removal (full entities, not just ids, so the caller
// can render what's actually being removed), and new tasks to add.
type GoalRevision struct {
	Goal         PlannedGoal
	RemovedTasks []*entity.DailyTask
	NewTasks     []PlannedTask
}

type reviseGoalGoalWire struct {
	Title                string `json:"title"`
	Detail               string `json:"detail"`
	AchievementCondition string `json:"achievement_condition"`
	EndDate              string `json:"end_date"`
	Mode                 string `json:"mode"`
}

type reviseGoalWire struct {
	Goal          reviseGoalGoalWire `json:"goal"`
	RemoveTaskIDs []string           `json:"remove_task_ids"`
	NewTasks      []planTaskWire     `json:"new_tasks"`
}

// ReviseGoalUsecase mirrors PlanGoalUsecase but proposes an update to an
// existing goal instead of a brand-new one: a revised goal, plus optional
// changes to its already-scheduled tasks (add/remove), rather than just the
// goal fields alone.
type ReviseGoalUsecase struct {
	ai       PlanGenerator
	goals    SingleGoalReader
	tasks    TaskRangeReader
	messages ConversationMessageReader
	clock    usecase.Clock
}

func NewReviseGoalUsecase(ai PlanGenerator, goals SingleGoalReader, tasks TaskRangeReader, messages ConversationMessageReader, clock usecase.Clock) *ReviseGoalUsecase {
	return &ReviseGoalUsecase{ai: ai, goals: goals, tasks: tasks, messages: messages, clock: clock}
}

func (u *ReviseGoalUsecase) Execute(ctx context.Context, input ReviseGoalInput) (GoalRevision, error) {
	goal, err := u.goals.FindByID(input.GoalID)
	if err != nil {
		return GoalRevision{}, err
	}
	if goal == nil {
		return GoalRevision{}, fmt.Errorf("goal not found: %s", input.GoalID)
	}

	now := u.clock.Now()

	history, err := u.messages.FindByConversationID(input.ConversationID)
	if err != nil {
		return GoalRevision{}, err
	}

	// Candidates for removal: not-yet-done tasks from today through the
	// goal's current deadline. Already-done tasks are historical record and
	// shouldn't be offered for removal.
	found, err := u.tasks.FindByGoalIDAndDateRange(goal.ID, now, goal.EndDate)
	if err != nil {
		return GoalRevision{}, err
	}
	candidates := make([]*entity.DailyTask, 0, len(found))
	candidatesByID := make(map[string]*entity.DailyTask, len(found))
	for _, t := range found {
		if t.Done {
			continue
		}
		candidates = append(candidates, t)
		candidatesByID[t.ID] = t
	}

	prompt := fmt.Sprintf(reviseGoalSystemPromptTemplate,
		now.Format(planDateLayout),
		goal.Title, goal.Detail, goal.AchievementCondition, goal.EndDate.Format(planDateLayout), goal.Mode,
		summarizeRemovalCandidates(candidates),
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
		return GoalRevision{}, err
	}

	var wire reviseGoalWire
	if err := json.Unmarshal([]byte(raw), &wire); err != nil {
		return GoalRevision{}, fmt.Errorf("failed to parse AI revision response: %w", err)
	}

	endDate, err := time.Parse(planDateLayout, wire.Goal.EndDate)
	if err != nil {
		return GoalRevision{}, fmt.Errorf("invalid end_date %q from AI: %w", wire.Goal.EndDate, err)
	}

	revision := GoalRevision{
		Goal: PlannedGoal{
			Title:                wire.Goal.Title,
			Detail:               wire.Goal.Detail,
			AchievementCondition: wire.Goal.AchievementCondition,
			EndDate:              endDate,
			Mode:                 entity.GoalMode(wire.Goal.Mode),
		},
	}

	// Only accept ids that were actually offered as candidates — an AI-
	// hallucinated id must never silently pass through to a caller that
	// might act on it (e.g. delete an unrelated task).
	for _, id := range wire.RemoveTaskIDs {
		if task, ok := candidatesByID[id]; ok {
			revision.RemovedTasks = append(revision.RemovedTasks, task)
		}
	}

	for _, t := range wire.NewTasks {
		start, err := time.Parse(planDateLayout, t.StartDate)
		if err != nil {
			return GoalRevision{}, fmt.Errorf("invalid task start_date %q from AI: %w", t.StartDate, err)
		}
		end, err := time.Parse(planDateLayout, t.EndDate)
		if err != nil {
			return GoalRevision{}, fmt.Errorf("invalid task end_date %q from AI: %w", t.EndDate, err)
		}

		rule := dailytask.RecurrenceRule{Type: dailytask.RecurrenceType(t.RuleType), IntervalDays: t.IntervalDays}
		for _, wd := range t.Weekdays {
			rule.Weekdays = append(rule.Weekdays, time.Weekday(wd))
		}

		revision.NewTasks = append(revision.NewTasks, PlannedTask{
			Content:   t.Content,
			StartDate: start,
			EndDate:   end,
			Rule:      rule,
		})
	}

	return revision, nil
}

// summarizeRemovalCandidates lists not-yet-done tasks with their ids, so the
// AI can reference a specific existing task if the conversation calls for
// changing it. Returns a short "no changeable tasks" note rather than an
// empty string when there's nothing to offer, so the prompt is explicit
// either way.
func summarizeRemovalCandidates(candidates []*entity.DailyTask) string {
	if len(candidates) == 0 {
		return "\n見直し候補タスク: なし"
	}

	var b strings.Builder
	b.WriteString("\n見直し候補タスク(未完了・今後の予定分のみ):\n")
	for _, t := range candidates {
		fmt.Fprintf(&b, "- id: %s, 日付: %s, 内容: %s\n", t.ID, t.Date.Format(planDateLayout), t.Content)
	}
	return b.String()
}
