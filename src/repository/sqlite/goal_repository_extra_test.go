package sqlite

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
)

func TestGoalRepository_Delete_CascadesToTasksAndGoalScopedConversations(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	db, err := Open(path)
	if err != nil {
		t.Fatalf("Open() returned unexpected error: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	workspaceRepo := NewWorkspaceRepository(db)
	workspace := entity.NewWorkSpace("workspace-001", DefaultUserID, "private", time.Now())
	if err := workspaceRepo.Save(workspace); err != nil {
		t.Fatalf("workspaceRepo.Save() returned unexpected error: %v", err)
	}

	goalRepo := NewGoalRepository(db)
	doomedGoal := entity.NewGoal("goal-doomed", workspace.ID, "消えるgoal", "detail", "cond", time.Now(), entity.ModeStrict, time.Now())
	survivorGoal := entity.NewGoal("goal-survivor", workspace.ID, "残るgoal", "detail", "cond", time.Now(), entity.ModeStrict, time.Now())
	for _, g := range []*entity.Goal{doomedGoal, survivorGoal} {
		if err := goalRepo.Save(g); err != nil {
			t.Fatalf("Save() returned unexpected error: %v", err)
		}
	}

	taskRepo := NewDailyTaskRepository(db)
	doomedTask := entity.NewDailyTask("task-doomed", doomedGoal.ID, time.Now(), "消えるtask", time.Now())
	survivorTask := entity.NewDailyTask("task-survivor", survivorGoal.ID, time.Now(), "残るtask", time.Now())
	for _, task := range []*entity.DailyTask{doomedTask, survivorTask} {
		if err := taskRepo.Save(task); err != nil {
			t.Fatalf("Save() returned unexpected error: %v", err)
		}
	}

	convRepo := NewConversationRepository(db)
	// A goal-scoped review conversation for the doomed goal (must go away),
	// a goal-scoped one for the survivor goal, and a general workspace-wide
	// conversation (goal_id == "") that must survive untouched either way.
	doomedGoalConv := entity.NewConversation("conv-doomed", workspace.ID, doomedGoal.ID, "消える見直し会話", time.Now())
	survivorGoalConv := entity.NewConversation("conv-survivor-goal", workspace.ID, survivorGoal.ID, "残る見直し会話", time.Now())
	generalConv := entity.NewConversation("conv-general", workspace.ID, "", "一般の会話", time.Now())
	for _, c := range []*entity.Conversation{doomedGoalConv, survivorGoalConv, generalConv} {
		if err := convRepo.Save(c); err != nil {
			t.Fatalf("Save() returned unexpected error: %v", err)
		}
	}

	msgRepo := NewConversationMessageRepository(db)
	doomedMsg := entity.NewConversationMessage("msg-doomed", doomedGoalConv.ID, entity.ChatRoleUser, "消えるメッセージ", time.Now())
	if err := msgRepo.Save(doomedMsg); err != nil {
		t.Fatalf("Save() returned unexpected error: %v", err)
	}

	if err := goalRepo.Delete(doomedGoal.ID); err != nil {
		t.Fatalf("Delete() returned unexpected error: %v", err)
	}

	remaining, err := goalRepo.FindByWorkspaceID(workspace.ID)
	if err != nil {
		t.Fatalf("FindByWorkspaceID() returned unexpected error: %v", err)
	}
	if len(remaining) != 1 || remaining[0].ID != survivorGoal.ID {
		t.Errorf("FindByWorkspaceID() after Delete() = %+v, want only %s", remaining, survivorGoal.ID)
	}

	doomedTasks, err := taskRepo.FindByGoalIDAndDateRange(doomedGoal.ID, time.Now().AddDate(0, 0, -1), time.Now().AddDate(0, 0, 1))
	if err != nil {
		t.Fatalf("FindByGoalIDAndDateRange(doomed) returned unexpected error: %v", err)
	}
	if len(doomedTasks) != 0 {
		t.Errorf("tasks for deleted goal = %+v, want none", doomedTasks)
	}

	survivorTasks, err := taskRepo.FindByGoalIDAndDateRange(survivorGoal.ID, time.Now().AddDate(0, 0, -1), time.Now().AddDate(0, 0, 1))
	if err != nil {
		t.Fatalf("FindByGoalIDAndDateRange(survivor) returned unexpected error: %v", err)
	}
	if len(survivorTasks) != 1 {
		t.Errorf("tasks for surviving goal = %+v, want 1 untouched", survivorTasks)
	}

	doomedConvs, err := convRepo.FindByGoalID(doomedGoal.ID)
	if err != nil {
		t.Fatalf("FindByGoalID(doomed) returned unexpected error: %v", err)
	}
	if len(doomedConvs) != 0 {
		t.Errorf("goal-scoped conversations for deleted goal = %+v, want none", doomedConvs)
	}

	doomedMessages, err := msgRepo.FindByConversationID(doomedGoalConv.ID)
	if err != nil {
		t.Fatalf("FindByConversationID(doomed) returned unexpected error: %v", err)
	}
	if len(doomedMessages) != 0 {
		t.Errorf("messages for the deleted goal's conversation = %+v, want none", doomedMessages)
	}

	survivorGoalConvs, err := convRepo.FindByGoalID(survivorGoal.ID)
	if err != nil {
		t.Fatalf("FindByGoalID(survivor) returned unexpected error: %v", err)
	}
	if len(survivorGoalConvs) != 1 {
		t.Errorf("goal-scoped conversations for surviving goal = %+v, want 1 untouched", survivorGoalConvs)
	}

	generalConvs, err := convRepo.FindByWorkspaceID(workspace.ID)
	if err != nil {
		t.Fatalf("FindByWorkspaceID() returned unexpected error: %v", err)
	}
	if len(generalConvs) != 1 || generalConvs[0].ID != generalConv.ID {
		t.Errorf("general workspace conversations = %+v, want only %s untouched", generalConvs, generalConv.ID)
	}
}
