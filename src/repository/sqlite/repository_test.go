package sqlite

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
)

func TestWorkspaceRepository_SaveAndFindAll(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	db, err := Open(path)
	if err != nil {
		t.Fatalf("Open() returned unexpected error: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	repo := NewWorkspaceRepository(db)
	createdAt := time.Date(2026, 8, 14, 9, 0, 0, 0, time.UTC)
	workspace := entity.NewWorkSpace("workspace-001", DefaultUserID, "private", createdAt)

	if err := repo.Save(workspace); err != nil {
		t.Fatalf("Save() returned unexpected error: %v", err)
	}

	got, err := repo.FindAll()
	if err != nil {
		t.Fatalf("FindAll() returned unexpected error: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("FindAll() returned %d workspaces, want 1", len(got))
	}
	if got[0].ID != workspace.ID || got[0].UserID != workspace.UserID || got[0].Name != workspace.Name {
		t.Errorf("FindAll()[0] = %+v, want %+v", got[0], workspace)
	}
	if !got[0].CreatedAt.Equal(createdAt) {
		t.Errorf("CreatedAt = %v, want %v", got[0].CreatedAt, createdAt)
	}
}

func TestGoalRepository_SaveAndFindByWorkspaceID(t *testing.T) {
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

	repo := NewGoalRepository(db)
	createdAt := time.Date(2026, 8, 14, 9, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	goal := entity.NewGoal("goal-001", workspace.ID, "Run a marathon", "detail", "condition", endDate, entity.ModeStrict, createdAt)

	if err := repo.Save(goal); err != nil {
		t.Fatalf("Save() returned unexpected error: %v", err)
	}

	got, err := repo.FindByWorkspaceID(workspace.ID)
	if err != nil {
		t.Fatalf("FindByWorkspaceID() returned unexpected error: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("FindByWorkspaceID() returned %d goals, want 1", len(got))
	}
	if got[0].ID != goal.ID || got[0].Title != goal.Title || got[0].Status != entity.StatusActive || got[0].Mode != entity.ModeStrict {
		t.Errorf("FindByWorkspaceID()[0] = %+v, want %+v", got[0], goal)
	}
	if !got[0].EndDate.Equal(endDate) {
		t.Errorf("EndDate = %v, want %v", got[0].EndDate, endDate)
	}
}

func TestGoalRepository_UpdatePostponement(t *testing.T) {
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

	repo := NewGoalRepository(db)
	endDate := time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC)
	goal := entity.NewGoal("goal-001", workspace.ID, "Run a marathon", "detail", "condition", endDate, entity.ModeStrict, time.Now())
	if err := repo.Save(goal); err != nil {
		t.Fatalf("Save() returned unexpected error: %v", err)
	}

	newEndDate := time.Date(2026, 8, 21, 0, 0, 0, 0, time.UTC)
	if err := repo.UpdatePostponement(goal.ID, newEndDate, 1); err != nil {
		t.Fatalf("UpdatePostponement() returned unexpected error: %v", err)
	}

	got, err := repo.FindByWorkspaceID(workspace.ID)
	if err != nil {
		t.Fatalf("FindByWorkspaceID() returned unexpected error: %v", err)
	}
	if len(got) != 1 || !got[0].EndDate.Equal(newEndDate) || got[0].PostponeCount != 1 {
		t.Fatalf("FindByWorkspaceID() after UpdatePostponement = %+v, want EndDate=%v PostponeCount=1", got, newEndDate)
	}
}

func TestGoalRepository_FindByID(t *testing.T) {
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

	repo := NewGoalRepository(db)
	goal := entity.NewGoal("goal-001", workspace.ID, "Run a marathon", "detail", "condition", time.Now(), entity.ModeStrict, time.Now())
	if err := repo.Save(goal); err != nil {
		t.Fatalf("Save() returned unexpected error: %v", err)
	}

	got, err := repo.FindByID(goal.ID)
	if err != nil {
		t.Fatalf("FindByID() returned unexpected error: %v", err)
	}
	if got == nil || got.ID != goal.ID || got.Title != goal.Title {
		t.Errorf("FindByID() = %+v, want %+v", got, goal)
	}

	none, err := repo.FindByID("does-not-exist")
	if err != nil {
		t.Fatalf("FindByID() for a missing id returned unexpected error: %v", err)
	}
	if none != nil {
		t.Errorf("FindByID() for a missing id = %+v, want nil", none)
	}
}

func TestGoalRepository_Update(t *testing.T) {
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

	repo := NewGoalRepository(db)
	goal := entity.NewGoal("goal-001", workspace.ID, "Run a marathon", "detail", "condition", time.Now(), entity.ModeStrict, time.Now())
	if err := repo.Save(goal); err != nil {
		t.Fatalf("Save() returned unexpected error: %v", err)
	}

	newEndDate := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)
	updated := entity.NewGoal(goal.ID, workspace.ID, "フルマラソン完走", "新しいdetail", "新しい達成条件", newEndDate, entity.ModeWant, goal.CreatedAt)
	if err := repo.Update(updated); err != nil {
		t.Fatalf("Update() returned unexpected error: %v", err)
	}

	got, err := repo.FindByID(goal.ID)
	if err != nil {
		t.Fatalf("FindByID() returned unexpected error: %v", err)
	}
	if got.Title != "フルマラソン完走" || got.Detail != "新しいdetail" || got.AchievementCondition != "新しい達成条件" || got.Mode != entity.ModeWant {
		t.Errorf("FindByID() after Update() = %+v, want the updated fields", got)
	}
	if !got.EndDate.Equal(newEndDate) {
		t.Errorf("EndDate = %v, want %v", got.EndDate, newEndDate)
	}
	if got.Status != entity.StatusActive {
		t.Errorf("Status = %v, want unchanged %v (Update must not touch status)", got.Status, entity.StatusActive)
	}
}

func TestGoalRepository_FindAll(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	db, err := Open(path)
	if err != nil {
		t.Fatalf("Open() returned unexpected error: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	workspaceRepo := NewWorkspaceRepository(db)
	wsA := entity.NewWorkSpace("workspace-a", DefaultUserID, "A", time.Now())
	wsB := entity.NewWorkSpace("workspace-b", DefaultUserID, "B", time.Now())
	for _, ws := range []*entity.WorkSpace{wsA, wsB} {
		if err := workspaceRepo.Save(ws); err != nil {
			t.Fatalf("Save() returned unexpected error: %v", err)
		}
	}

	repo := NewGoalRepository(db)
	goalA := entity.NewGoal("goal-a", wsA.ID, "Aのgoal", "detail", "cond", time.Now(), entity.ModeStrict, time.Now())
	goalB := entity.NewGoal("goal-b", wsB.ID, "Bのgoal", "detail", "cond", time.Now(), entity.ModeStrict, time.Now())
	for _, g := range []*entity.Goal{goalA, goalB} {
		if err := repo.Save(g); err != nil {
			t.Fatalf("Save() returned unexpected error: %v", err)
		}
	}

	got, err := repo.FindAll()
	if err != nil {
		t.Fatalf("FindAll() returned unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("FindAll() returned %d goals, want 2 (across both workspaces)", len(got))
	}
}

func TestDailyTaskRepository_SaveAndFindByDate(t *testing.T) {
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
	goal := entity.NewGoal("goal-001", workspace.ID, "Run a marathon", "detail", "condition", time.Now(), entity.ModeStrict, time.Now())
	if err := goalRepo.Save(goal); err != nil {
		t.Fatalf("goalRepo.Save() returned unexpected error: %v", err)
	}

	repo := NewDailyTaskRepository(db)
	date := time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)
	task := entity.NewDailyTask("task-001", goal.ID, date, "Run 5km", time.Now())

	if err := repo.Save(task); err != nil {
		t.Fatalf("Save() returned unexpected error: %v", err)
	}

	got, err := repo.FindByDate(date)
	if err != nil {
		t.Fatalf("FindByDate() returned unexpected error: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("FindByDate() returned %d tasks, want 1", len(got))
	}
	if got[0].ID != task.ID || got[0].Content != task.Content || got[0].Done != false {
		t.Errorf("FindByDate()[0] = %+v, want %+v", got[0], task)
	}
	if !got[0].Date.Equal(date) {
		t.Errorf("Date = %v, want %v", got[0].Date, date)
	}
}

func TestDailyTaskRepository_FindByGoalIDAndDateRange(t *testing.T) {
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
	goalA := entity.NewGoal("goal-001", workspace.ID, "Run a marathon", "detail", "condition", time.Now(), entity.ModeStrict, time.Now())
	if err := goalRepo.Save(goalA); err != nil {
		t.Fatalf("goalRepo.Save() returned unexpected error: %v", err)
	}
	goalB := entity.NewGoal("goal-002", workspace.ID, "Read more", "detail", "condition", time.Now(), entity.ModeWant, time.Now())
	if err := goalRepo.Save(goalB); err != nil {
		t.Fatalf("goalRepo.Save() returned unexpected error: %v", err)
	}

	repo := NewDailyTaskRepository(db)
	inRange1 := entity.NewDailyTask("task-001", goalA.ID, time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC), "Run 5km", time.Now())
	inRange2 := entity.NewDailyTask("task-002", goalA.ID, time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC), "Run 5km", time.Now())
	beforeRange := entity.NewDailyTask("task-003", goalA.ID, time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC), "Run 5km", time.Now())
	afterRange := entity.NewDailyTask("task-004", goalA.ID, time.Date(2026, 8, 8, 0, 0, 0, 0, time.UTC), "Run 5km", time.Now())
	otherGoal := entity.NewDailyTask("task-005", goalB.ID, time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC), "Read a book", time.Now())
	for _, task := range []*entity.DailyTask{inRange1, inRange2, beforeRange, afterRange, otherGoal} {
		if err := repo.Save(task); err != nil {
			t.Fatalf("Save() returned unexpected error: %v", err)
		}
	}

	from := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 8, 7, 0, 0, 0, 0, time.UTC)
	got, err := repo.FindByGoalIDAndDateRange(goalA.ID, from, to)
	if err != nil {
		t.Fatalf("FindByGoalIDAndDateRange() returned unexpected error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("FindByGoalIDAndDateRange() returned %d tasks, want 2", len(got))
	}
	if got[0].ID != inRange1.ID || got[1].ID != inRange2.ID {
		t.Errorf("FindByGoalIDAndDateRange() = %+v, want [task-001, task-002] in date order", got)
	}
}

func TestDailyTaskRepository_UpdateDone(t *testing.T) {
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
	goal := entity.NewGoal("goal-001", workspace.ID, "Run a marathon", "detail", "condition", time.Now(), entity.ModeStrict, time.Now())
	if err := goalRepo.Save(goal); err != nil {
		t.Fatalf("goalRepo.Save() returned unexpected error: %v", err)
	}

	repo := NewDailyTaskRepository(db)
	date := time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)
	task := entity.NewDailyTask("task-001", goal.ID, date, "Run 5km", time.Now())
	if err := repo.Save(task); err != nil {
		t.Fatalf("Save() returned unexpected error: %v", err)
	}

	completedAt := time.Date(2026, 8, 17, 9, 30, 0, 0, time.UTC)
	if err := repo.UpdateDone(task.ID, true, &completedAt); err != nil {
		t.Fatalf("UpdateDone() returned unexpected error: %v", err)
	}

	got, err := repo.FindByDate(date)
	if err != nil {
		t.Fatalf("FindByDate() returned unexpected error: %v", err)
	}
	if len(got) != 1 || !got[0].Done {
		t.Fatalf("FindByDate() after UpdateDone = %+v, want Done=true", got)
	}
	if got[0].CompletedAt == nil || !got[0].CompletedAt.Equal(completedAt) {
		t.Errorf("CompletedAt = %v, want %v", got[0].CompletedAt, completedAt)
	}

	if err := repo.UpdateDone(task.ID, false, nil); err != nil {
		t.Fatalf("UpdateDone() (clearing) returned unexpected error: %v", err)
	}
	gotAfterClear, err := repo.FindByDate(date)
	if err != nil {
		t.Fatalf("FindByDate() returned unexpected error: %v", err)
	}
	if len(gotAfterClear) != 1 || gotAfterClear[0].Done {
		t.Fatalf("FindByDate() after clearing = %+v, want Done=false", gotAfterClear)
	}
	if gotAfterClear[0].CompletedAt != nil {
		t.Errorf("CompletedAt after clearing = %v, want nil", gotAfterClear[0].CompletedAt)
	}
}

func TestDailyTaskRepository_FindOldestPendingBefore(t *testing.T) {
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
	goal := entity.NewGoal("goal-001", workspace.ID, "Run a marathon", "detail", "condition", time.Now(), entity.ModeStrict, time.Now())
	if err := goalRepo.Save(goal); err != nil {
		t.Fatalf("goalRepo.Save() returned unexpected error: %v", err)
	}

	repo := NewDailyTaskRepository(db)
	older := entity.NewDailyTask("task-001", goal.ID, time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC), "Run 5km", time.Now())
	newer := entity.NewDailyTask("task-002", goal.ID, time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC), "Run 5km", time.Now())
	doneOne := entity.NewDailyTask("task-003", goal.ID, time.Date(2026, 8, 9, 0, 0, 0, 0, time.UTC), "Run 5km", time.Now())
	doneOne.Done = true
	for _, task := range []*entity.DailyTask{older, newer, doneOne} {
		if err := repo.Save(task); err != nil {
			t.Fatalf("Save() returned unexpected error: %v", err)
		}
	}

	before := time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)
	got, err := repo.FindOldestPendingBefore(goal.ID, before)
	if err != nil {
		t.Fatalf("FindOldestPendingBefore() returned unexpected error: %v", err)
	}
	if got == nil || got.ID != older.ID {
		t.Fatalf("FindOldestPendingBefore() = %+v, want task-001 (oldest not-done, ignoring the already-done task)", got)
	}

	none, err := repo.FindOldestPendingBefore(goal.ID, time.Date(2026, 8, 9, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("FindOldestPendingBefore() returned unexpected error: %v", err)
	}
	if none != nil {
		t.Errorf("FindOldestPendingBefore() with an early cutoff = %+v, want nil", none)
	}
}

func TestDailyTaskRepository_ShiftPendingForward(t *testing.T) {
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
	goal := entity.NewGoal("goal-001", workspace.ID, "Run a marathon", "detail", "condition", time.Now(), entity.ModeStrict, time.Now())
	if err := goalRepo.Save(goal); err != nil {
		t.Fatalf("goalRepo.Save() returned unexpected error: %v", err)
	}

	repo := NewDailyTaskRepository(db)
	missed := entity.NewDailyTask("task-001", goal.ID, time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC), "Run 5km", time.Now())
	future := entity.NewDailyTask("task-002", goal.ID, time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC), "Run 5km", time.Now())
	doneFuture := entity.NewDailyTask("task-003", goal.ID, time.Date(2026, 8, 11, 0, 0, 0, 0, time.UTC), "Run 5km", time.Now())
	doneFuture.Done = true
	before := entity.NewDailyTask("task-004", goal.ID, time.Date(2026, 8, 9, 0, 0, 0, 0, time.UTC), "Run 5km", time.Now())
	for _, task := range []*entity.DailyTask{missed, future, doneFuture, before} {
		if err := repo.Save(task); err != nil {
			t.Fatalf("Save() returned unexpected error: %v", err)
		}
	}

	if err := repo.ShiftPendingForward(goal.ID, missed.Date); err != nil {
		t.Fatalf("ShiftPendingForward() returned unexpected error: %v", err)
	}

	got, err := repo.FindByGoalIDAndDateRange(goal.ID, time.Date(2026, 8, 9, 0, 0, 0, 0, time.UTC), time.Date(2026, 8, 13, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("FindByGoalIDAndDateRange() returned unexpected error: %v", err)
	}

	byID := make(map[string]*entity.DailyTask)
	for _, task := range got {
		byID[task.ID] = task
	}

	if !byID[missed.ID].Date.Equal(time.Date(2026, 8, 11, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("missed task Date = %v, want 2026-08-11 (shifted +1 day)", byID[missed.ID].Date)
	}
	if !byID[future.ID].Date.Equal(time.Date(2026, 8, 13, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("future task Date = %v, want 2026-08-13 (shifted +1 day)", byID[future.ID].Date)
	}
	if !byID[doneFuture.ID].Date.Equal(time.Date(2026, 8, 11, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("done task Date = %v, want unchanged 2026-08-11 (done tasks are never shifted)", byID[doneFuture.ID].Date)
	}
	if !byID[before.ID].Date.Equal(time.Date(2026, 8, 9, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("earlier task Date = %v, want unchanged 2026-08-09 (before the shift's fromDate)", byID[before.ID].Date)
	}
}

func TestDailyTaskRepository_FindByDateAndWorkspaceID(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	db, err := Open(path)
	if err != nil {
		t.Fatalf("Open() returned unexpected error: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	workspaceRepo := NewWorkspaceRepository(db)
	wsA := entity.NewWorkSpace("workspace-a", DefaultUserID, "A", time.Now())
	wsB := entity.NewWorkSpace("workspace-b", DefaultUserID, "B", time.Now())
	for _, ws := range []*entity.WorkSpace{wsA, wsB} {
		if err := workspaceRepo.Save(ws); err != nil {
			t.Fatalf("Save() returned unexpected error: %v", err)
		}
	}

	goalRepo := NewGoalRepository(db)
	goalA := entity.NewGoal("goal-a", wsA.ID, "Aのgoal", "detail", "cond", time.Now(), entity.ModeStrict, time.Now())
	goalB := entity.NewGoal("goal-b", wsB.ID, "Bのgoal", "detail", "cond", time.Now(), entity.ModeStrict, time.Now())
	for _, g := range []*entity.Goal{goalA, goalB} {
		if err := goalRepo.Save(g); err != nil {
			t.Fatalf("Save() returned unexpected error: %v", err)
		}
	}

	repo := NewDailyTaskRepository(db)
	date := time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC)
	taskA := entity.NewDailyTask("task-a", goalA.ID, date, "Aのタスク", time.Now())
	taskB := entity.NewDailyTask("task-b", goalB.ID, date, "Bのタスク", time.Now())
	otherDate := entity.NewDailyTask("task-other-date", goalA.ID, date.AddDate(0, 0, 1), "別の日", time.Now())
	for _, task := range []*entity.DailyTask{taskA, taskB, otherDate} {
		if err := repo.Save(task); err != nil {
			t.Fatalf("Save() returned unexpected error: %v", err)
		}
	}

	gotA, err := repo.FindByDateAndWorkspaceID(date, wsA.ID)
	if err != nil {
		t.Fatalf("FindByDateAndWorkspaceID(A) returned unexpected error: %v", err)
	}
	if len(gotA) != 1 || gotA[0].ID != taskA.ID {
		t.Errorf("FindByDateAndWorkspaceID(A) = %+v, want only task-a", gotA)
	}

	gotB, err := repo.FindByDateAndWorkspaceID(date, wsB.ID)
	if err != nil {
		t.Fatalf("FindByDateAndWorkspaceID(B) returned unexpected error: %v", err)
	}
	if len(gotB) != 1 || gotB[0].ID != taskB.ID {
		t.Errorf("FindByDateAndWorkspaceID(B) = %+v, want only task-b", gotB)
	}
}
