package sqlite

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
)

func TestWorkspaceRepository_UpdateName(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	db, err := Open(path)
	if err != nil {
		t.Fatalf("Open() returned unexpected error: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	repo := NewWorkspaceRepository(db)
	ws := entity.NewWorkSpace("workspace-001", DefaultUserID, "private", time.Now())
	if err := repo.Save(ws); err != nil {
		t.Fatalf("Save() returned unexpected error: %v", err)
	}

	if err := repo.UpdateName(ws.ID, "仕事"); err != nil {
		t.Fatalf("UpdateName() returned unexpected error: %v", err)
	}

	got, err := repo.FindAll()
	if err != nil {
		t.Fatalf("FindAll() returned unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].Name != "仕事" {
		t.Fatalf("FindAll() after UpdateName() = %+v, want Name=仕事", got)
	}
}

func TestWorkspaceRepository_Delete_CascadesToGoalsTasksAndConversations(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	db, err := Open(path)
	if err != nil {
		t.Fatalf("Open() returned unexpected error: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	workspaceRepo := NewWorkspaceRepository(db)
	doomed := entity.NewWorkSpace("workspace-doomed", DefaultUserID, "消える方", time.Now())
	survivor := entity.NewWorkSpace("workspace-survivor", DefaultUserID, "残る方", time.Now())
	for _, ws := range []*entity.WorkSpace{doomed, survivor} {
		if err := workspaceRepo.Save(ws); err != nil {
			t.Fatalf("Save() returned unexpected error: %v", err)
		}
	}

	goalRepo := NewGoalRepository(db)
	doomedGoal := entity.NewGoal("goal-doomed", doomed.ID, "消えるgoal", "detail", "cond", time.Now(), entity.ModeStrict, time.Now())
	survivorGoal := entity.NewGoal("goal-survivor", survivor.ID, "残るgoal", "detail", "cond", time.Now(), entity.ModeStrict, time.Now())
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
	doomedConv := entity.NewConversation("conv-doomed", doomed.ID, "", "消える会話", time.Now())
	survivorConv := entity.NewConversation("conv-survivor", survivor.ID, "", "残る会話", time.Now())
	for _, c := range []*entity.Conversation{doomedConv, survivorConv} {
		if err := convRepo.Save(c); err != nil {
			t.Fatalf("Save() returned unexpected error: %v", err)
		}
	}

	msgRepo := NewConversationMessageRepository(db)
	doomedMsg := entity.NewConversationMessage("msg-doomed", doomedConv.ID, entity.ChatRoleUser, "消えるメッセージ", time.Now())
	survivorMsg := entity.NewConversationMessage("msg-survivor", survivorConv.ID, entity.ChatRoleUser, "残るメッセージ", time.Now())
	for _, m := range []*entity.ConversationMessage{doomedMsg, survivorMsg} {
		if err := msgRepo.Save(m); err != nil {
			t.Fatalf("Save() returned unexpected error: %v", err)
		}
	}

	if err := workspaceRepo.Delete(doomed.ID); err != nil {
		t.Fatalf("Delete() returned unexpected error: %v", err)
	}

	workspaces, err := workspaceRepo.FindAll()
	if err != nil {
		t.Fatalf("FindAll() returned unexpected error: %v", err)
	}
	if len(workspaces) != 1 || workspaces[0].ID != survivor.ID {
		t.Errorf("FindAll() after Delete() = %+v, want only %s", workspaces, survivor.ID)
	}

	goals, err := goalRepo.FindByWorkspaceID(doomed.ID)
	if err != nil {
		t.Fatalf("FindByWorkspaceID(doomed) returned unexpected error: %v", err)
	}
	if len(goals) != 0 {
		t.Errorf("goals for deleted workspace = %+v, want none", goals)
	}

	survivorGoals, err := goalRepo.FindByWorkspaceID(survivor.ID)
	if err != nil {
		t.Fatalf("FindByWorkspaceID(survivor) returned unexpected error: %v", err)
	}
	if len(survivorGoals) != 1 {
		t.Errorf("goals for surviving workspace = %+v, want 1 untouched", survivorGoals)
	}

	doomedTasks, err := taskRepo.FindByGoalIDAndDateRange(doomedGoal.ID, time.Now().AddDate(0, 0, -1), time.Now().AddDate(0, 0, 1))
	if err != nil {
		t.Fatalf("FindByGoalIDAndDateRange(doomed) returned unexpected error: %v", err)
	}
	if len(doomedTasks) != 0 {
		t.Errorf("tasks for deleted workspace's goal = %+v, want none", doomedTasks)
	}

	survivorTasks, err := taskRepo.FindByGoalIDAndDateRange(survivorGoal.ID, time.Now().AddDate(0, 0, -1), time.Now().AddDate(0, 0, 1))
	if err != nil {
		t.Fatalf("FindByGoalIDAndDateRange(survivor) returned unexpected error: %v", err)
	}
	if len(survivorTasks) != 1 {
		t.Errorf("tasks for surviving workspace's goal = %+v, want 1 untouched", survivorTasks)
	}

	convs, err := convRepo.FindByWorkspaceID(doomed.ID)
	if err != nil {
		t.Fatalf("FindByWorkspaceID(doomed) returned unexpected error: %v", err)
	}
	if len(convs) != 0 {
		t.Errorf("conversations for deleted workspace = %+v, want none", convs)
	}

	doomedMessages, err := msgRepo.FindByConversationID(doomedConv.ID)
	if err != nil {
		t.Fatalf("FindByConversationID(doomed) returned unexpected error: %v", err)
	}
	if len(doomedMessages) != 0 {
		t.Errorf("messages for deleted workspace's conversation = %+v, want none", doomedMessages)
	}

	survivorMessages, err := msgRepo.FindByConversationID(survivorConv.ID)
	if err != nil {
		t.Fatalf("FindByConversationID(survivor) returned unexpected error: %v", err)
	}
	if len(survivorMessages) != 1 {
		t.Errorf("messages for surviving workspace's conversation = %+v, want 1 untouched", survivorMessages)
	}
}
