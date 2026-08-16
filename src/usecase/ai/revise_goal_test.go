package ai

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
)

func TestReviseGoalUsecase_Execute(t *testing.T) {
	created := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 8, 21, 0, 0, 0, 0, time.UTC)
	goal := entity.NewGoal("goal-001", "workspace-001", "フルマラソン完走", "detail", "42.195kmを完走する", endDate, entity.ModeStrict, created)
	goals := stubSingleGoalReader{goal: goal}

	gen := &stubPlanGenerator{json: `{
		"title": "フルマラソン完走(延長)",
		"detail": "detail",
		"achievement_condition": "42.195kmを完走する",
		"end_date": "2026-09-20",
		"mode": "strict"
	}`}
	messages := &spyMessageRepo{existing: []*entity.ConversationMessage{
		entity.NewConversationMessage("msg-001", "conv-001", entity.ChatRoleUser, "そろそろ厳しいので期限を延ばしたい", time.Now()),
		entity.NewConversationMessage("msg-002", "conv-001", entity.ChatRoleModel, "1ヶ月延ばしましょうか？", time.Now()),
	}}
	fixedNow := time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC)
	uc := NewReviseGoalUsecase(gen, goals, messages, stubClock{now: fixedNow})

	got, err := uc.Execute(context.Background(), ReviseGoalInput{GoalID: "goal-001", ConversationID: "conv-001"})
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}

	if got.Title != "フルマラソン完走(延長)" {
		t.Errorf("Title = %q, want %q", got.Title, "フルマラソン完走(延長)")
	}
	wantEndDate := time.Date(2026, 9, 20, 0, 0, 0, 0, time.UTC)
	if !got.EndDate.Equal(wantEndDate) {
		t.Errorf("EndDate = %v, want %v", got.EndDate, wantEndDate)
	}

	if !strings.Contains(gen.capturedSystem, "フルマラソン完走") {
		t.Errorf("system instruction = %q, want it to mention the current goal title", gen.capturedSystem)
	}

	// Same Gemini "no trailing model turn" constraint as PlanGoalUsecase.
	got2 := gen.capturedMessages
	if len(got2) == 0 || got2[len(got2)-1].Role != entity.ChatRoleUser {
		t.Fatalf("GenerateJSON() called with messages ending in role %q, want %q", got2[len(got2)-1].Role, entity.ChatRoleUser)
	}
	if len(got2) != len(messages.existing)+1 {
		t.Errorf("GenerateJSON() received %d messages, want %d (history + 1 appended trailing user turn)", len(got2), len(messages.existing)+1)
	}
}

func TestReviseGoalUsecase_Execute_GoalNotFound(t *testing.T) {
	goals := stubSingleGoalReader{goal: nil}
	gen := &stubPlanGenerator{json: `{}`}
	messages := &spyMessageRepo{}

	uc := NewReviseGoalUsecase(gen, goals, messages, stubClock{now: time.Now()})
	_, err := uc.Execute(context.Background(), ReviseGoalInput{GoalID: "missing", ConversationID: "conv-001"})
	if err == nil {
		t.Fatal("Execute() with a missing goal returned no error, want one instead of panicking downstream")
	}
}
