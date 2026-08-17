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
	tasks := stubTaskRangeReader{}

	gen := &stubPlanGenerator{json: `{
		"goal": {
			"title": "フルマラソン完走(延長)",
			"detail": "detail",
			"achievement_condition": "42.195kmを完走する",
			"end_date": "2026-09-20",
			"mode": "strict"
		},
		"remove_task_ids": [],
		"new_tasks": []
	}`}
	messages := &spyMessageRepo{existing: []*entity.ConversationMessage{
		entity.NewConversationMessage("msg-001", "conv-001", entity.ChatRoleUser, "そろそろ厳しいので期限を延ばしたい", time.Now()),
		entity.NewConversationMessage("msg-002", "conv-001", entity.ChatRoleModel, "1ヶ月延ばしましょうか？", time.Now()),
	}}
	fixedNow := time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC)
	uc := NewReviseGoalUsecase(gen, goals, tasks, messages, stubClock{now: fixedNow})

	got, err := uc.Execute(context.Background(), ReviseGoalInput{GoalID: "goal-001", ConversationID: "conv-001"})
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}

	if got.Goal.Title != "フルマラソン完走(延長)" {
		t.Errorf("Goal.Title = %q, want %q", got.Goal.Title, "フルマラソン完走(延長)")
	}
	wantEndDate := time.Date(2026, 9, 20, 0, 0, 0, 0, time.UTC)
	if !got.Goal.EndDate.Equal(wantEndDate) {
		t.Errorf("Goal.EndDate = %v, want %v", got.Goal.EndDate, wantEndDate)
	}
	if len(got.RemovedTasks) != 0 || len(got.NewTasks) != 0 {
		t.Errorf("RemovedTasks/NewTasks = %+v/%+v, want both empty when the AI proposed no task changes", got.RemovedTasks, got.NewTasks)
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

func TestReviseGoalUsecase_Execute_ProposesTaskChanges(t *testing.T) {
	created := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 8, 21, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC)
	goal := entity.NewGoal("goal-001", "workspace-001", "フルマラソン完走", "detail", "42.195kmを完走する", endDate, entity.ModeStrict, created)
	goals := stubSingleGoalReader{goal: goal}

	pendingTask := entity.NewDailyTask("task-pending", goal.ID, now.AddDate(0, 0, 2), "10kmラン", created)
	doneTask := entity.NewDailyTask("task-done", goal.ID, now.AddDate(0, 0, -1), "5kmラン", created)
	doneTask.Done = true
	tasks := stubTaskRangeReader{tasks: []*entity.DailyTask{pendingTask, doneTask}}

	gen := &stubPlanGenerator{json: `{
		"goal": {
			"title": "フルマラソン完走",
			"detail": "detail",
			"achievement_condition": "42.195kmを完走する",
			"end_date": "2026-09-20",
			"mode": "strict"
		},
		"remove_task_ids": ["task-pending"],
		"new_tasks": [
			{"content": "軽いジョグ", "start_date": "2026-08-22", "end_date": "2026-09-20", "rule_type": "interval", "interval_days": 3}
		]
	}`}
	messages := &spyMessageRepo{existing: []*entity.ConversationMessage{
		entity.NewConversationMessage("msg-001", "conv-001", entity.ChatRoleUser, "頻度を減らしたい", time.Now()),
	}}
	uc := NewReviseGoalUsecase(gen, goals, tasks, messages, stubClock{now: now})

	got, err := uc.Execute(context.Background(), ReviseGoalInput{GoalID: "goal-001", ConversationID: "conv-001"})
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}

	if len(got.RemovedTasks) != 1 || got.RemovedTasks[0].ID != "task-pending" {
		t.Fatalf("RemovedTasks = %+v, want only task-pending", got.RemovedTasks)
	}
	if len(got.NewTasks) != 1 || got.NewTasks[0].Content != "軽いジョグ" {
		t.Fatalf("NewTasks = %+v, want the proposed 軽いジョグ task", got.NewTasks)
	}

	// The prompt should have given the AI the candidate task's id to
	// reference, and only future/pending tasks — not the already-done one.
	if !strings.Contains(gen.capturedSystem, "task-pending") {
		t.Errorf("system instruction = %q, want it to list the candidate task id", gen.capturedSystem)
	}
	if strings.Contains(gen.capturedSystem, "task-done") {
		t.Errorf("system instruction = %q, want it to exclude already-done tasks from removal candidates", gen.capturedSystem)
	}
}

func TestReviseGoalUsecase_Execute_IgnoresUnknownRemoveTaskIDs(t *testing.T) {
	created := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 8, 21, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC)
	goal := entity.NewGoal("goal-001", "workspace-001", "フルマラソン完走", "detail", "cond", endDate, entity.ModeStrict, created)
	goals := stubSingleGoalReader{goal: goal}
	tasks := stubTaskRangeReader{tasks: []*entity.DailyTask{
		entity.NewDailyTask("task-real", goal.ID, now.AddDate(0, 0, 1), "x", created),
	}}

	// The AI hallucinated an id that was never offered as a candidate — this
	// must not panic or silently pass through a removal for a task the
	// caller can't identify.
	gen := &stubPlanGenerator{json: `{
		"goal": {"title": "x", "detail": "x", "achievement_condition": "x", "end_date": "2026-09-20", "mode": "strict"},
		"remove_task_ids": ["task-real", "task-hallucinated"],
		"new_tasks": []
	}`}
	messages := &spyMessageRepo{}
	uc := NewReviseGoalUsecase(gen, goals, tasks, messages, stubClock{now: now})

	got, err := uc.Execute(context.Background(), ReviseGoalInput{GoalID: "goal-001", ConversationID: "conv-001"})
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}

	if len(got.RemovedTasks) != 1 || got.RemovedTasks[0].ID != "task-real" {
		t.Errorf("RemovedTasks = %+v, want only the known task-real, hallucinated id silently dropped", got.RemovedTasks)
	}
}

func TestReviseGoalUsecase_Execute_GoalNotFound(t *testing.T) {
	goals := stubSingleGoalReader{goal: nil}
	tasks := stubTaskRangeReader{}
	gen := &stubPlanGenerator{json: `{}`}
	messages := &spyMessageRepo{}

	uc := NewReviseGoalUsecase(gen, goals, tasks, messages, stubClock{now: time.Now()})
	_, err := uc.Execute(context.Background(), ReviseGoalInput{GoalID: "missing", ConversationID: "conv-001"})
	if err == nil {
		t.Fatal("Execute() with a missing goal returned no error, want one instead of panicking downstream")
	}
}
