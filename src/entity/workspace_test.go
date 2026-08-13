package entity

import (
	"reflect"
	"testing"
	"time"
)

func TestNewWorkSpace(t *testing.T) {
	createdAt := time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)

	got := NewWorkSpace("workspace-001", "user-001", "private", createdAt)

	want := &WorkSpace{ID: "workspace-001", UserID: "user-001", Name: "private", CreatedAt: createdAt}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("NewWorkSpace() = %+v, want %+v", got, want)
	}
}
