package workspace

import (
	"reflect"
	"testing"
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
)

func TestListWorkspacesUsecase_Execute(t *testing.T) {
	createdAt := time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)
	want := []*entity.WorkSpace{
		entity.NewWorkSpace("workspace-001", "user-001", "private", createdAt),
		entity.NewWorkSpace("workspace-002", "user-001", "work", createdAt),
	}

	repo := &mockRepository{findAll: want}
	uc := NewListWorkspacesUsecase(repo)

	got, err := uc.Execute()
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Execute() = %+v, want %+v", got, want)
	}
}
