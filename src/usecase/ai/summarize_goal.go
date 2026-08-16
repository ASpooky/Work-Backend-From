package ai

import (
	"context"
	"fmt"
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
	"github.com/ASpooky/Work-Backend-From/src/usecase"
)

const goalSummarySystemPrompt = `あなたはユーザーの目標達成を支援するコーチです。以下の目標の進捗データをもとに、
現在のペースで期限までに達成できそうか、達成率、先延ばしの回数などを踏まえて、
簡潔に(3〜4文程度)日本語で状況を評価してください。単なる数字の読み上げではなく、
具体的なアドバイスや励ましを含めてください。`

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

	prompt := fmt.Sprintf(
		`今日の日付: %s
目標: %s
達成条件: %s
モード: %s
期限まで残り: %d日
これまでの達成率: %d%%(%d件中%d件完了)
先延ばし回数: %d回`,
		nowDate.Format(planDateLayout), goal.Title, goal.AchievementCondition, goal.Mode,
		stats.DaysRemaining, int(stats.AchievementRate*100), stats.ScheduledCount, stats.DoneCount, stats.PostponeCount,
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
