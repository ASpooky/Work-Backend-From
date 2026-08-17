package ai

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
	"github.com/ASpooky/Work-Backend-From/src/usecase"
)

const recentMemosLimit = 5

const goalSummarySystemPrompt = `あなたはユーザーの目標達成を支援するコーチです。以下の目標の進捗データをもとに、
現在のペースで期限までに達成できそうか、達成率、先延ばしの回数などを踏まえて、
簡潔に(3〜4文程度)日本語で状況を評価してください。単なる数字の読み上げではなく、
具体的なアドバイスや励ましを含めてください。直近のメモが提供されている場合は、
その内容(体調・気づきなど)にも触れ、数字だけでは分からない具体的なフィードバックを
含めてください。`

type GoalSummaryInput struct {
	GoalID string
}

type GoalStats struct {
	ScheduledCount  int
	DoneCount       int
	AchievementRate float64
	PostponeCount   int
	DaysRemaining   int
}

type GoalSummaryOutput struct {
	Stats   GoalStats
	Summary string
}

type SingleGoalReader interface {
	FindByID(id string) (*entity.Goal, error)
}

type TaskRangeReader interface {
	FindByGoalIDAndDateRange(goalID string, from, to time.Time) ([]*entity.DailyTask, error)
}

// SummarizeGoalUsecase computes the same plain stats as
// usecase.GetGoalStatsUsecase (duplicated rather than depending on that
// usecase directly, so usecase/ai's dependencies stay pure repository-
// shaped interfaces like everywhere else in this package) and asks the AI
// to turn them into a short natural-language assessment.
type SummarizeGoalUsecase struct {
	goals SingleGoalReader
	tasks TaskRangeReader
	ai    ChatCompleter
	clock usecase.Clock
}

func NewSummarizeGoalUsecase(goals SingleGoalReader, tasks TaskRangeReader, ai ChatCompleter, clock usecase.Clock) *SummarizeGoalUsecase {
	return &SummarizeGoalUsecase{goals: goals, tasks: tasks, ai: ai, clock: clock}
}

func (u *SummarizeGoalUsecase) Execute(ctx context.Context, input GoalSummaryInput) (GoalSummaryOutput, error) {
	goal, err := u.goals.FindByID(input.GoalID)
	if err != nil {
		return GoalSummaryOutput{}, err
	}
	if goal == nil {
		return GoalSummaryOutput{}, fmt.Errorf("goal not found: %s", input.GoalID)
	}

	now := u.clock.Now()
	nowDate := truncateToDate(now)

	stats, err := computeGoalStats(u.tasks, goal, now)
	if err != nil {
		return GoalSummaryOutput{}, err
	}

	tasks, err := u.tasks.FindByGoalIDAndDateRange(goal.ID, goal.CreatedAt, goal.EndDate)
	if err != nil {
		return GoalSummaryOutput{}, err
	}

	prompt := fmt.Sprintf(
		`今日の日付: %s
目標: %s
達成条件: %s
モード: %s
期限まで残り: %d日
これまでの達成率: %d%%(%d件中%d件完了)
先延ばし回数: %d回%s`,
		nowDate.Format(planDateLayout), goal.Title, goal.AchievementCondition, goal.Mode,
		stats.DaysRemaining, int(stats.AchievementRate*100), stats.ScheduledCount, stats.DoneCount, stats.PostponeCount,
		summarizeRecentMemos(tasks),
	)

	summary, err := u.ai.Chat(ctx, goalSummarySystemPrompt, []entity.ChatMessage{
		{Role: entity.ChatRoleUser, Content: prompt},
	})
	if err != nil {
		return GoalSummaryOutput{}, err
	}

	return GoalSummaryOutput{Stats: stats, Summary: summary}, nil
}

// computeGoalStats duplicates usecase.GetGoalStatsUsecase's math (see that
// type's doc comment for why usecase/ai doesn't just depend on it), shared
// within this package by SummarizeGoalUsecase and GoalReviewChatUsecase.
func computeGoalStats(tasks TaskRangeReader, goal *entity.Goal, now time.Time) (GoalStats, error) {
	nowDate := truncateToDate(now)

	found, err := tasks.FindByGoalIDAndDateRange(goal.ID, goal.CreatedAt, goal.EndDate)
	if err != nil {
		return GoalStats{}, err
	}

	scheduled, done := 0, 0
	for _, t := range found {
		if t.Date.After(nowDate) {
			continue
		}
		scheduled++
		if t.Done {
			done++
		}
	}

	rate := 0.0
	if scheduled > 0 {
		rate = float64(done) / float64(scheduled)
	}

	return GoalStats{
		ScheduledCount:  scheduled,
		DoneCount:       done,
		AchievementRate: rate,
		PostponeCount:   goal.PostponeCount,
		DaysRemaining:   int(goal.EndDate.Sub(nowDate).Hours() / 24),
	}, nil
}

func truncateToDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// summarizeRecentMemos pulls the most recent tasks that have a user-written
// memo attached, so the AI summary can reference specifics ("腰が痛かった"
// on leg-day entries) instead of only raw done/scheduled counts. Returns ""
// when nothing has a memo, so the prompt doesn't grow a section for no
// reason.
func summarizeRecentMemos(tasks []*entity.DailyTask) string {
	withMemo := make([]*entity.DailyTask, 0, len(tasks))
	for _, t := range tasks {
		if t.Memo != nil && strings.TrimSpace(*t.Memo) != "" {
			withMemo = append(withMemo, t)
		}
	}
	if len(withMemo) == 0 {
		return ""
	}

	sort.Slice(withMemo, func(i, j int) bool { return withMemo[i].Date.After(withMemo[j].Date) })
	if len(withMemo) > recentMemosLimit {
		withMemo = withMemo[:recentMemosLimit]
	}

	var b strings.Builder
	b.WriteString("\n直近のメモ:\n")
	for _, t := range withMemo {
		fmt.Fprintf(&b, "- %s(%s): %s\n", t.Date.Format(planDateLayout), t.Content, *t.Memo)
	}
	return b.String()
}
