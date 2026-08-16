package ai

import (
	"context"
	"fmt"

	"github.com/ASpooky/Work-Backend-From/src/usecase"
)

const goalReviewSystemPromptTemplate = `あなたはユーザーの既存の目標の見直しを手伝うコーチです。今日の日付は %s です。

現在の目標:
- タイトル: %s
- 詳細: %s
- 達成条件: %s
- 期限: %s
- モード: %s
- 進捗: 予定%d件中%d件完了(達成率%d%%)、先延ばし%d回、残り%d日

ユーザーと対話しながら、この目標のタイトル・詳細・達成条件・期限・モードの見直しを手伝ってください。
「1ヶ月延長」のような相対的な期限が出てきたら、現在の期限を基準に具体的な日付に置き換えて
確認してください。期限やタスクの時期を変更する際は、下記の予定状況も踏まえて、他の目標と
負荷が集中しない時期を提案してください。見直しの方向性が固まってきたら、
「この内容で見直しを反映できそうです」とユーザーに伝えてください。回答は日本語、簡潔に。
%s`

// GoalReviewChatUsecase is ChatUsecase's counterpart for an existing goal:
// same Chatter shape (so SendMessageUsecase can host either one), but its
// system prompt is built from the goal's current fields and progress
// instead of being fixed, so the AI is discussing this specific goal from
// the first message rather than starting from a blank slate.
type GoalReviewChatUsecase struct {
	ai             ChatCompleter
	goals          SingleGoalReader
	tasks          TaskRangeReader
	workspaceTasks WorkspaceTaskRangeReader
	clock          usecase.Clock
}

func NewGoalReviewChatUsecase(ai ChatCompleter, goals SingleGoalReader, tasks TaskRangeReader, workspaceTasks WorkspaceTaskRangeReader, clock usecase.Clock) *GoalReviewChatUsecase {
	return &GoalReviewChatUsecase{ai: ai, goals: goals, tasks: tasks, workspaceTasks: workspaceTasks, clock: clock}
}

func (u *GoalReviewChatUsecase) Execute(ctx context.Context, input ChatInput) (string, error) {
	goal, err := u.goals.FindByID(input.GoalID)
	if err != nil {
		return "", err
	}
	if goal == nil {
		return "", fmt.Errorf("goal not found: %s", input.GoalID)
	}

	now := u.clock.Now()
	stats, err := computeGoalStats(u.tasks, goal, now)
	if err != nil {
		return "", err
	}

	lookaheadEnd := now.AddDate(0, 0, busyDaysLookaheadDays)
	busyTasks, err := u.workspaceTasks.FindByWorkspaceIDAndDateRange(goal.WorkspaceID, now, lookaheadEnd)
	if err != nil {
		return "", err
	}
	busyDays := summarizeBusyDays(busyTasks, goal.ID)

	prompt := fmt.Sprintf(goalReviewSystemPromptTemplate,
		now.Format(planDateLayout),
		goal.Title, goal.Detail, goal.AchievementCondition, goal.EndDate.Format(planDateLayout), goal.Mode,
		stats.ScheduledCount, stats.DoneCount, int(stats.AchievementRate*100), stats.PostponeCount, stats.DaysRemaining,
		busyDays,
	)

	return u.ai.Chat(ctx, prompt, input.Messages)
}
