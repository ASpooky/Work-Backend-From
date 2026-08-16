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

func TestGoalRepository_UpdateEndDate(t *testing.T) {
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
	if err := repo.UpdateEndDate(goal.ID, newEndDate); err != nil {
		t.Fatalf("UpdateEndDate() returned unexpected error: %v", err)
	}

	got, err := repo.FindByWorkspaceID(workspace.ID)
	if err != nil {
		t.Fatalf("FindByWorkspaceID() returned unexpected error: %v", err)
	}
	if len(got) != 1 || !got[0].EndDate.Equal(newEndDate) {
		t.Fatalf("FindByWorkspaceID() after UpdateEndDate = %+v, want EndDate=%v", got, newEndDate)
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

	if err := repo.UpdateDone(task.ID, true); err != nil {
		t.Fatalf("UpdateDone() returned unexpected error: %v", err)
	}

	got, err := repo.FindByDate(date)
	if err != nil {
		t.Fatalf("FindByDate() returned unexpected error: %v", err)
	}
	if len(got) != 1 || !got[0].Done {
		t.Fatalf("FindByDate() after UpdateDone = %+v, want Done=true", got)
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
