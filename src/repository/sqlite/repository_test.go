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
