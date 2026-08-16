package workspace

import (
	"testing"
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
)

func TestDeleteWorkspaceUsecase_Execute(t *testing.T) {
	repo := &mockRepository{findAll: []*entity.WorkSpace{
		entity.NewWorkSpace("workspace-001", "user-001", "仕事", time.Now()),
		entity.NewWorkSpace("workspace-002", "user-001", "プライベート", time.Now()),
	}}
	uc := NewDeleteWorkspaceUsecase(repo)

	if err := uc.Execute("workspace-001"); err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}

	if repo.deletedID != "workspace-001" {
		t.Errorf("Delete called with %q, want %q", repo.deletedID, "workspace-001")
	}
}

func TestDeleteWorkspaceUsecase_Execute_RefusesToDeleteTheLastWorkspace(t *testing.T) {
	repo := &mockRepository{findAll: []*entity.WorkSpace{
		entity.NewWorkSpace("workspace-001", "user-001", "唯一のworkspace", time.Now()),
	}}
	uc := NewDeleteWorkspaceUsecase(repo)

	err := uc.Execute("workspace-001")
	if err == nil {
		t.Fatal("Execute() deleting the only remaining workspace returned nil error, want non-nil")
	}
	if repo.deletedID != "" {
		t.Errorf("Delete was called (deletedID=%q), want it never called when only one workspace remains", repo.deletedID)
	}
}
