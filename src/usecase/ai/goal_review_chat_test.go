package ai

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
)

func TestGoalReviewChatUsecase_Execute(t *testing.T) {
	created := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 8, 21, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC)

	goal := entity.NewGoal("goal-001", "workspace-001", "フルマラソン完走", "detail", "42.195kmを完走する", endDate, entity.ModeStrict, created)
	goal.PostponeCount = 2

	goals := stubSingleGoalReader{goal: goal}
	tasks := stubTaskRangeReader{}
	completer := &stubChatCompleter{reply: "期限を1ヶ月延ばしましょうか？"}

	uc := NewGoalReviewChatUsecase(completer, goals, tasks, stubClock{now: now})

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

func TestGoalReviewChatUsecase_Execute_GoalNotFound(t *testing.T) {
	goals := stubSingleGoalReader{goal: nil}
	tasks := stubTaskRangeReader{}
	completer := &stubChatCompleter{reply: "unused"}

	uc := NewGoalReviewChatUsecase(completer, goals, tasks, stubClock{now: time.Now()})
	_, err := uc.Execute(context.Background(), ChatInput{GoalID: "missing"})
	if err == nil {
		t.Fatal("Execute() with a missing goal returned no error, want one instead of panicking downstream")
	}
}
