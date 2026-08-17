package ai

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
)

type stubSingleGoalReader struct {
	goal *entity.Goal
}

func (s stubSingleGoalReader) FindByID(id string) (*entity.Goal, error) {
	return s.goal, nil
}

type stubTaskRangeReader struct {
	tasks []*entity.DailyTask
}

func (s stubTaskRangeReader) FindByGoalIDAndDateRange(goalID string, from, to time.Time) ([]*entity.DailyTask, error) {
	return s.tasks, nil
}

func TestSummarizeGoalUsecase_Execute(t *testing.T) {
	created := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 8, 21, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC)

	goal := entity.NewGoal("goal-001", "workspace-001", "フルマラソン完走", "detail", "42.195kmを完走する", endDate, entity.ModeStrict, created)
	goal.PostponeCount = 1

	doneTask := entity.NewDailyTask("t1", goal.ID, now.AddDate(0, 0, -1), "x", created)
	doneTask.Done = true

	goals := stubSingleGoalReader{goal: goal}
	tasks := stubTaskRangeReader{tasks: []*entity.DailyTask{doneTask}}
	completer := &stubChatCompleter{reply: "順調です。このペースなら期限までに達成できそうです。"}

	uc := NewSummarizeGoalUsecase(goals, tasks, completer, stubClock{now: now})

	got, err := uc.Execute(context.Background(), GoalSummaryInput{GoalID: "goal-001"})
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}

	if got.Summary != "順調です。このペースなら期限までに達成できそうです。" {
		t.Errorf("Summary = %q, want the stub reply", got.Summary)
	}
	if got.Stats.DoneCount != 1 || got.Stats.ScheduledCount != 1 {
		t.Errorf("Stats = %+v, want ScheduledCount=1 DoneCount=1", got.Stats)
	}
	if got.Stats.PostponeCount != 1 {
		t.Errorf("Stats.PostponeCount = %d, want 1", got.Stats.PostponeCount)
	}

	if !strings.Contains(completer.capturedMessages[0].Content, "フルマラソン完走") {
		t.Errorf("prompt sent to AI = %q, want it to mention the goal title", completer.capturedMessages[0].Content)
	}
	if !strings.Contains(completer.capturedMessages[0].Content, "先延ばし") {
		t.Errorf("prompt sent to AI = %q, want it to mention postpone count", completer.capturedMessages[0].Content)
	}
}

func TestSummarizeGoalUsecase_Execute_IncludesRecentMemos(t *testing.T) {
	created := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 8, 21, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC)

	goal := entity.NewGoal("goal-001", "workspace-001", "フルマラソン完走", "detail", "42.195kmを完走する", endDate, entity.ModeStrict, created)

	memo := "腰が痛かった"
	doneTask := entity.NewDailyTask("t1", goal.ID, now.AddDate(0, 0, -1), "10kmラン", created)
	doneTask.Done = true
	doneTask.Memo = &memo
	// A task with no memo shouldn't produce an empty/blank line in the prompt.
	noMemoTask := entity.NewDailyTask("t2", goal.ID, now.AddDate(0, 0, -2), "ストレッチ", created)
	noMemoTask.Done = true

	goals := stubSingleGoalReader{goal: goal}
	tasks := stubTaskRangeReader{tasks: []*entity.DailyTask{doneTask, noMemoTask}}
	completer := &stubChatCompleter{reply: "順調です"}

	uc := NewSummarizeGoalUsecase(goals, tasks, completer, stubClock{now: now})

	if _, err := uc.Execute(context.Background(), GoalSummaryInput{GoalID: "goal-001"}); err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}

	prompt := completer.capturedMessages[0].Content
	if !strings.Contains(prompt, "腰が痛かった") {
		t.Errorf("prompt = %q, want it to include the task memo", prompt)
	}
	if !strings.Contains(prompt, "10kmラン") {
		t.Errorf("prompt = %q, want the memo attributed to its task content", prompt)
	}
}

func TestSummarizeGoalUsecase_Execute_NoMemosOmitsSection(t *testing.T) {
	created := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 8, 21, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC)

	goal := entity.NewGoal("goal-001", "workspace-001", "フルマラソン完走", "detail", "cond", endDate, entity.ModeStrict, created)
	task := entity.NewDailyTask("t1", goal.ID, now.AddDate(0, 0, -1), "10kmラン", created)
	task.Done = true

	goals := stubSingleGoalReader{goal: goal}
	tasks := stubTaskRangeReader{tasks: []*entity.DailyTask{task}}
	completer := &stubChatCompleter{reply: "順調です"}

	uc := NewSummarizeGoalUsecase(goals, tasks, completer, stubClock{now: now})

	if _, err := uc.Execute(context.Background(), GoalSummaryInput{GoalID: "goal-001"}); err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}

	if strings.Contains(completer.capturedMessages[0].Content, "メモ") {
		t.Errorf("prompt = %q, want no memo section when no task has a memo", completer.capturedMessages[0].Content)
	}
}

func TestSummarizeGoalUsecase_Execute_GoalNotFound(t *testing.T) {
	goals := stubSingleGoalReader{goal: nil}
	tasks := stubTaskRangeReader{}
	completer := &stubChatCompleter{reply: "unused"}

	uc := NewSummarizeGoalUsecase(goals, tasks, completer, stubClock{now: time.Now()})
	_, err := uc.Execute(context.Background(), GoalSummaryInput{GoalID: "missing"})
	if err == nil {
		t.Fatal("Execute() with a missing goal returned no error, want one instead of panicking downstream")
	}
}
