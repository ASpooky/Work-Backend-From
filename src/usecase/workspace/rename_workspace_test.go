package workspace

import "testing"

func TestRenameWorkspaceUsecase_Execute(t *testing.T) {
	repo := &mockRepository{}
	uc := NewRenameWorkspaceUsecase(repo)

	if err := uc.Execute(RenameWorkspaceInput{ID: "workspace-001", Name: "仕事"}); err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}

	if repo.updatedID != "workspace-001" || repo.updatedName != "仕事" {
		t.Errorf("UpdateName called with (%q, %q), want (%q, %q)", repo.updatedID, repo.updatedName, "workspace-001", "仕事")
	}
}
