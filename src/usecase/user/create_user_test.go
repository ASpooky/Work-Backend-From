package user

import (
	"reflect"
	"testing"
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
)

type stubIDGenerator struct {
	id string
}

func (s stubIDGenerator) NewID() string {
	return s.id
}

type stubClock struct {
	now time.Time
}

func (s stubClock) Now() time.Time {
	return s.now
}

type mockRepository struct {
	saved *entity.User
}

func (m *mockRepository) Save(user *entity.User) error {
	m.saved = user
	return nil
}

func TestCreateUserUsecase_Execute(t *testing.T) {
	fixedID := "user-001"
	fixedNow := time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)

	repo := &mockRepository{}
	uc := NewCreateUserUsecase(repo, stubIDGenerator{id: fixedID}, stubClock{now: fixedNow})

	input := CreateUserInput{
		Name: "Taro",
	}

	got, err := uc.Execute(input)
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}

	want := entity.NewUser(fixedID, input.Name, fixedNow)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Execute() = %+v, want %+v", got, want)
	}

	if repo.saved == nil {
		t.Fatal("expected repository.Save to be called")
	}
	if !reflect.DeepEqual(repo.saved, want) {
		t.Errorf("repository saved = %+v, want %+v", repo.saved, want)
	}
}
