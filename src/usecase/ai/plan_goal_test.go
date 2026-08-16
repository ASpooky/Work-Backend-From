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

func TestPlanGoalUsecase_Execute_SingleGoal(t *testing.T) {
	gen := &stubPlanGenerator{json: `{
		"goals": [
			{
				"title": "5kmを走れるようになる",
				"detail": "毎日のランニング習慣をつける",
				"achievement_condition": "5km以上を休まず走り切る",
				"end_date": "2026-09-30",
				"mode": "strict",
				"tasks": [
					{
						"content": "2kmランニング",
						"start_date": "2026-08-17",
						"end_date": "2026-09-30",
						"rule_type": "weekly",
						"weekdays": [1, 3, 5]
					}
				]
			}
		]
	}`}
	messages := &spyMessageRepo{existing: []*entity.ConversationMessage{
		entity.NewConversationMessage("msg-001", "conv-001", entity.ChatRoleUser, "5km走れるようになりたい", time.Now()),
	}}
	fixedNow := time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC)
	uc := NewPlanGoalUsecase(gen, messages, stubWorkspaceTaskRangeReader{}, stubClock{now: fixedNow})

	got, err := uc.Execute(context.Background(), PlanGoalInput{ConversationID: "conv-001"})
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}

	if len(got.Goals) != 1 {
		t.Fatalf("len(Goals) = %d, want 1", len(got.Goals))
	}
	planned := got.Goals[0]

	if planned.Goal.Title != "5kmを走れるようになる" {
		t.Errorf("Goal.Title = %q, want %q", planned.Goal.Title, "5kmを走れるようになる")
	}
	if planned.Goal.Mode != entity.ModeStrict {
		t.Errorf("Goal.Mode = %q, want %q", planned.Goal.Mode, entity.ModeStrict)
	}
	wantEndDate := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)
	if !planned.Goal.EndDate.Equal(wantEndDate) {
		t.Errorf("Goal.EndDate = %v, want %v", planned.Goal.EndDate, wantEndDate)
	}

	if len(planned.Tasks) != 1 {
		t.Fatalf("len(Tasks) = %d, want 1", len(planned.Tasks))
	}
	task := planned.Tasks[0]
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

func TestPlanGoalUsecase_Execute_MultiplePhaseGoals(t *testing.T) {
	gen := &stubPlanGenerator{json: `{
		"goals": [
			{
				"title": "土台作り期",
				"detail": "怪我なく走れる体を作る",
				"achievement_condition": "週20km走る",
				"end_date": "2026-09-30",
				"mode": "want",
				"tasks": [
					{"content": "軽いジョグ", "start_date": "2026-08-17", "end_date": "2026-09-30", "rule_type": "interval", "interval_days": 2}
				]
			},
			{
				"title": "追い込み期",
				"detail": "本番に向けた走り込み",
				"achievement_condition": "週40km走る",
				"end_date": "2026-11-15",
				"mode": "strict",
				"tasks": [
					{"content": "ロング走", "start_date": "2026-10-01", "end_date": "2026-11-15", "rule_type": "weekly", "weekdays": [6]}
				]
			}
		]
	}`}
	messages := &spyMessageRepo{existing: []*entity.ConversationMessage{
		entity.NewConversationMessage("msg-001", "conv-001", entity.ChatRoleUser, "3ヶ月後のフルマラソンに向けて段階的に鍛えたい", time.Now()),
	}}
	uc := NewPlanGoalUsecase(gen, messages, stubWorkspaceTaskRangeReader{}, stubClock{now: time.Now()})

	got, err := uc.Execute(context.Background(), PlanGoalInput{ConversationID: "conv-001"})
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}

	if len(got.Goals) != 2 {
		t.Fatalf("len(Goals) = %d, want 2 (multi-phase plan)", len(got.Goals))
	}
	if got.Goals[0].Goal.Title != "土台作り期" || got.Goals[0].Goal.Mode != entity.ModeWant {
		t.Errorf("Goals[0] = %+v, want phase 1 (土台作り期, want mode)", got.Goals[0].Goal)
	}
	if got.Goals[1].Goal.Title != "追い込み期" || got.Goals[1].Goal.Mode != entity.ModeStrict {
		t.Errorf("Goals[1] = %+v, want phase 2 (追い込み期, strict mode)", got.Goals[1].Goal)
	}
	if len(got.Goals[0].Tasks) != 1 || len(got.Goals[1].Tasks) != 1 {
		t.Errorf("each phase should keep its own tasks, got %d and %d", len(got.Goals[0].Tasks), len(got.Goals[1].Tasks))
	}
}

func TestPlanGoalUsecase_Execute_AppendsTrailingUserTurnWhenConversationEndsWithModel(t *testing.T) {
	gen := &stubPlanGenerator{json: `{
		"goals": [
			{"title": "x", "detail": "x", "achievement_condition": "x", "end_date": "2026-09-30", "mode": "strict", "tasks": []}
		]
	}`}
	messages := &spyMessageRepo{existing: []*entity.ConversationMessage{
		entity.NewConversationMessage("msg-001", "conv-001", entity.ChatRoleUser, "3ヶ月後のフルマラソンで完走したい", time.Now()),
		entity.NewConversationMessage("msg-002", "conv-001", entity.ChatRoleModel, "必達ですか、努力目標ですか？", time.Now()),
	}}
	uc := NewPlanGoalUsecase(gen, messages, stubWorkspaceTaskRangeReader{}, stubClock{now: time.Now()})

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
		"goals": [
			{"title": "x", "detail": "x", "achievement_condition": "x", "end_date": "not-a-date", "mode": "strict", "tasks": []}
		]
	}`}
	messages := &spyMessageRepo{}
	uc := NewPlanGoalUsecase(gen, messages, stubWorkspaceTaskRangeReader{}, stubClock{now: time.Now()})

	_, err := uc.Execute(context.Background(), PlanGoalInput{ConversationID: "conv-001"})
	if err == nil {
		t.Fatal("Execute() with an invalid end_date returned nil error, want non-nil")
	}
}

func TestPlanGoalUsecase_Execute_NoGoals(t *testing.T) {
	gen := &stubPlanGenerator{json: `{"goals": []}`}
	messages := &spyMessageRepo{}
	uc := NewPlanGoalUsecase(gen, messages, stubWorkspaceTaskRangeReader{}, stubClock{now: time.Now()})

	_, err := uc.Execute(context.Background(), PlanGoalInput{ConversationID: "conv-001"})
	if err == nil {
		t.Fatal("Execute() with zero goals from the AI returned nil error, want non-nil (nothing usable to review/save)")
	}
}

func TestPlanGoalUsecase_Execute_MentionsOtherGoalsBusyDays(t *testing.T) {
	gen := &stubPlanGenerator{json: `{
		"goals": [
			{"title": "x", "detail": "x", "achievement_condition": "x", "end_date": "2026-09-30", "mode": "strict", "tasks": []}
		]
	}`}
	messages := &spyMessageRepo{}
	workspaceTasks := stubWorkspaceTaskRangeReader{tasks: []*entity.DailyTask{
		entity.NewDailyTask("t1", "other-goal", time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC), "x", time.Now()),
		entity.NewDailyTask("t2", "other-goal", time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC), "y", time.Now()),
	}}
	fixedNow := time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC)
	uc := NewPlanGoalUsecase(gen, messages, workspaceTasks, stubClock{now: fixedNow})

	if _, err := uc.Execute(context.Background(), PlanGoalInput{ConversationID: "conv-001", WorkspaceID: "workspace-001"}); err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}

	if !strings.Contains(gen.capturedSystem, "2026-08-20") {
		t.Errorf("system instruction = %q, want it to mention the busy day 2026-08-20", gen.capturedSystem)
	}
}

func TestPlanGoalUsecase_Execute_NoWorkspaceIDSkipsBusyDaysLookup(t *testing.T) {
	gen := &stubPlanGenerator{json: `{
		"goals": [
			{"title": "x", "detail": "x", "achievement_condition": "x", "end_date": "2026-09-30", "mode": "strict", "tasks": []}
		]
	}`}
	messages := &spyMessageRepo{}
	uc := NewPlanGoalUsecase(gen, messages, stubWorkspaceTaskRangeReader{}, stubClock{now: time.Now()})

	// No WorkspaceID (e.g. an older/malformed request) shouldn't error out —
	// it should just skip the busy-days context rather than querying with an
	// empty workspace id.
	if _, err := uc.Execute(context.Background(), PlanGoalInput{ConversationID: "conv-001"}); err != nil {
		t.Fatalf("Execute() with no WorkspaceID returned unexpected error: %v", err)
	}
}
