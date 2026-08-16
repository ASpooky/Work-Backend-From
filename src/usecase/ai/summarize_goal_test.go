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
