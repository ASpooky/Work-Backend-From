package ai

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
	"github.com/ASpooky/Work-Backend-From/src/usecase/dailytask"
)

type stubPlanGenerator struct {
	capturedSystem   string
	capturedMessages []entity.ChatMessage
	capturedSchema   map[string]any
	json             string
}

func (s *stubPlanGenerator) GenerateJSON(ctx context.Context, systemInstruction string, messages []entity.ChatMessage, schema map[string]any) (string, error) {
	s.capturedSystem = systemInstruction
	s.capturedMessages = messages
	s.capturedSchema = schema
	return s.json, nil
}

func TestPlanGoalUsecase_Execute(t *testing.T) {
	gen := &stubPlanGenerator{json: `{
		"goal": {
			"title": "5kmを走れるようになる",
			"detail": "毎日のランニング習慣をつける",
			"achievement_condition": "5km以上を休まず走り切る",
			"end_date": "2026-09-30",
			"mode": "strict"
		},
		"tasks": [
			{
				"content": "2kmランニング",
				"start_date": "2026-08-17",
				"end_date": "2026-09-30",
				"rule_type": "weekly",
				"weekdays": [1, 3, 5]
			}
		]
	}`}
	messages := &spyMessageRepo{existing: []*entity.ConversationMessage{
		entity.NewConversationMessage("msg-001", "conv-001", entity.ChatRoleUser, "5km走れるようになりたい", time.Now()),
	}}
	fixedNow := time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC)
	uc := NewPlanGoalUsecase(gen, messages, stubClock{now: fixedNow})

	got, err := uc.Execute(context.Background(), PlanGoalInput{ConversationID: "conv-001"})
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}

	if got.Goal.Title != "5kmを走れるようになる" {
		t.Errorf("Goal.Title = %q, want %q", got.Goal.Title, "5kmを走れるようになる")
	}
	if got.Goal.Mode != entity.ModeStrict {
		t.Errorf("Goal.Mode = %q, want %q", got.Goal.Mode, entity.ModeStrict)
	}
	wantEndDate := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)
	if !got.Goal.EndDate.Equal(wantEndDate) {
		t.Errorf("Goal.EndDate = %v, want %v", got.Goal.EndDate, wantEndDate)
	}

	if len(got.Tasks) != 1 {
		t.Fatalf("len(Tasks) = %d, want 1", len(got.Tasks))
	}
	task := got.Tasks[0]
	if task.Content != "2kmランニング" {
		t.Errorf("Tasks[0].Content = %q, want %q", task.Content, "2kmランニング")
	}
	if task.Rule.Type != dailytask.RecurrenceWeekly {
		t.Errorf("Tasks[0].Rule.Type = %q, want %q", task.Rule.Type, dailytask.RecurrenceWeekly)
	}
	wantWeekdays := []time.Weekday{time.Monday, time.Wednesday, time.Friday}
	if len(task.Rule.Weekdays) != len(wantWeekdays) {
		t.Fatalf("Tasks[0].Rule.Weekdays = %v, want %v", task.Rule.Weekdays, wantWeekdays)
	}
	for i, wd := range wantWeekdays {
		if task.Rule.Weekdays[i] != wd {
			t.Errorf("Tasks[0].Rule.Weekdays[%d] = %v, want %v", i, task.Rule.Weekdays[i], wd)
		}
	}

	if gen.capturedSchema == nil {
		t.Errorf("GenerateJSON() was called with a nil schema, want the plan schema")
	}
	if gen.capturedSystem == "" {
		t.Errorf("GenerateJSON() was called with an empty system instruction")
	}
	if !strings.Contains(gen.capturedSystem, "2026-08-16") {
		t.Errorf("GenerateJSON() system instruction = %q, want it to mention today's date so relative deadlines resolve correctly", gen.capturedSystem)
	}
}

func TestPlanGoalUsecase_Execute_AppendsTrailingUserTurnWhenConversationEndsWithModel(t *testing.T) {
	gen := &stubPlanGenerator{json: `{
		"goal": {"title": "x", "detail": "x", "achievement_condition": "x", "end_date": "2026-09-30", "mode": "strict"},
		"tasks": []
	}`}
	messages := &spyMessageRepo{existing: []*entity.ConversationMessage{
		entity.NewConversationMessage("msg-001", "conv-001", entity.ChatRoleUser, "3ヶ月後のフルマラソンで完走したい", time.Now()),
		entity.NewConversationMessage("msg-002", "conv-001", entity.ChatRoleModel, "必達ですか、努力目標ですか？", time.Now()),
	}}
	uc := NewPlanGoalUsecase(gen, messages, stubClock{now: time.Now()})

	if _, err := uc.Execute(context.Background(), PlanGoalInput{ConversationID: "conv-001"}); err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}

	// The Gemini API rejects a `contents` array whose last turn has role
	// "model" ("Requests ending with a model turn are not supported.").
	// Reproduced live: every real "落とし込む" click failed with exactly
	// that 400 once the conversation's last message was the AI's reply.
	got := gen.capturedMessages
	if len(got) == 0 || got[len(got)-1].Role != entity.ChatRoleUser {
		t.Fatalf("GenerateJSON() was called with messages ending in role %q, want the last message to be role %q",
			got[len(got)-1].Role, entity.ChatRoleUser)
	}
	if len(got) != len(messages.existing)+1 {
		t.Errorf("GenerateJSON() received %d messages, want %d (history + 1 appended trailing user turn)", len(got), len(messages.existing)+1)
	}
}

func TestPlanGoalUsecase_Execute_InvalidDate(t *testing.T) {
	gen := &stubPlanGenerator{json: `{
		"goal": {
			"title": "x", "detail": "x", "achievement_condition": "x",
			"end_date": "not-a-date", "mode": "strict"
		},
		"tasks": []
	}`}
	messages := &spyMessageRepo{}
	uc := NewPlanGoalUsecase(gen, messages, stubClock{now: time.Now()})

	_, err := uc.Execute(context.Background(), PlanGoalInput{ConversationID: "conv-001"})
	if err == nil {
		t.Fatal("Execute() with an invalid end_date returned nil error, want non-nil")
	}
}
