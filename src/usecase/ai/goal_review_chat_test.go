package ai

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
)

type stubWorkspaceTaskRangeReader struct {
	tasks []*entity.DailyTask
}

func (s stubWorkspaceTaskRangeReader) FindByWorkspaceIDAndDateRange(workspaceID string, from, to time.Time) ([]*entity.DailyTask, error) {
	return s.tasks, nil
}

func TestGoalReviewChatUsecase_Execute(t *testing.T) {
	created := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 8, 21, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC)

	goal := entity.NewGoal("goal-001", "workspace-001", "フルマラソン完走", "detail", "42.195kmを完走する", endDate, entity.ModeStrict, created)
	goal.PostponeCount = 2

	goals := stubSingleGoalReader{goal: goal}
	tasks := stubTaskRangeReader{}
	workspaceTasks := stubWorkspaceTaskRangeReader{}
	completer := &stubChatCompleter{reply: "期限を1ヶ月延ばしましょうか？"}

	uc := NewGoalReviewChatUsecase(completer, goals, tasks, workspaceTasks, stubClock{now: now})

	messages := []entity.ChatMessage{{Role: entity.ChatRoleUser, Content: "そろそろ厳しいので期限を延ばしたい"}}
	got, err := uc.Execute(context.Background(), ChatInput{Messages: messages, GoalID: "goal-001"})
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}
	if got != "期限を1ヶ月延ばしましょうか？" {
		t.Errorf("Execute() = %q, want the stub reply", got)
	}

	if !strings.Contains(completer.capturedSystem, "フルマラソン完走") {
		t.Errorf("system prompt = %q, want it to mention the current goal title", completer.capturedSystem)
	}
	if !strings.Contains(completer.capturedSystem, "先延ばし2回") {
		t.Errorf("system prompt = %q, want it to mention the current postpone count", completer.capturedSystem)
	}
	if len(completer.capturedMessages) != 1 || completer.capturedMessages[0].Content != messages[0].Content {
		t.Errorf("Chat() was called with messages = %+v, want %+v", completer.capturedMessages, messages)
	}
}

func TestGoalReviewChatUsecase_Execute_MentionsOtherGoalsBusyDays(t *testing.T) {
	created := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 8, 21, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC)

	goal := entity.NewGoal("goal-001", "workspace-001", "フルマラソン完走", "detail", "cond", endDate, entity.ModeStrict, created)

	goals := stubSingleGoalReader{goal: goal}
	tasks := stubTaskRangeReader{}
	workspaceTasks := stubWorkspaceTaskRangeReader{tasks: []*entity.DailyTask{
		entity.NewDailyTask("t1", "other-goal", time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC), "x", time.Now()),
		entity.NewDailyTask("t2", "other-goal", time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC), "y", time.Now()),
		// Same goal's own tasks shouldn't count as "competition" against itself.
		entity.NewDailyTask("t3", goal.ID, time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC), "z", time.Now()),
	}}
	completer := &stubChatCompleter{reply: "承知しました"}

	uc := NewGoalReviewChatUsecase(completer, goals, tasks, workspaceTasks, stubClock{now: now})

	if _, err := uc.Execute(context.Background(), ChatInput{
		Messages: []entity.ChatMessage{{Role: entity.ChatRoleUser, Content: "期限を伸ばしたい"}},
		GoalID:   "goal-001",
	}); err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}

	if !strings.Contains(completer.capturedSystem, "2026-08-20") {
		t.Errorf("system prompt = %q, want it to mention the busy day 2026-08-20 from the other goal", completer.capturedSystem)
	}
}

func TestGoalReviewChatUsecase_Execute_GoalNotFound(t *testing.T) {
	goals := stubSingleGoalReader{goal: nil}
	tasks := stubTaskRangeReader{}
	workspaceTasks := stubWorkspaceTaskRangeReader{}
	completer := &stubChatCompleter{reply: "unused"}

	uc := NewGoalReviewChatUsecase(completer, goals, tasks, workspaceTasks, stubClock{now: time.Now()})
	_, err := uc.Execute(context.Background(), ChatInput{GoalID: "missing"})
	if err == nil {
		t.Fatal("Execute() with a missing goal returned no error, want one instead of panicking downstream")
	}
}
