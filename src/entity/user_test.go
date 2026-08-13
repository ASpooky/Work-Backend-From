package entity

import (
	"reflect"
	"testing"
	"time"
)

func TestNewUser(t *testing.T) {
	createdAt := time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)

	got := NewUser("user-001", "Taro", createdAt)

	want := &User{ID: "user-001", Name: "Taro", CreatedAt: createdAt}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("NewUser() = %+v, want %+v", got, want)
	}
}
