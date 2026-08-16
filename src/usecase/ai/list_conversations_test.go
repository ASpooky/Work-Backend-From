package ai

import (
	"testing"
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
)

type stubConversationLister struct {
	conversations []*entity.Conversation
}

func (s stubConversationLister) FindByWorkspaceID(workspaceID string) ([]*entity.Conversation, error) {
	return s.conversations, nil
}

func TestListConversationsUsecase_Execute(t *testing.T) {
	want := []*entity.Conversation{
		entity.NewConversation("conv-001", "workspace-001", "", "5km走れるようになりたい", time.Now()),
	}
	uc := NewListConversationsUsecase(stubConversationLister{conversations: want})

	got, err := uc.Execute("workspace-001")
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].ID != "conv-001" {
		t.Errorf("Execute() = %+v, want %+v", got, want)
	}
}

type stubGoalConversationLister struct {
	conversations []*entity.Conversation
}

func (s stubGoalConversationLister) FindByGoalID(goalID string) ([]*entity.Conversation, error) {
	return s.conversations, nil
}

func TestListGoalConversationsUsecase_Execute(t *testing.T) {
	want := []*entity.Conversation{
		entity.NewConversation("conv-001", "workspace-001", "goal-001", "目標の見直し", time.Now()),
	}
	uc := NewListGoalConversationsUsecase(stubGoalConversationLister{conversations: want})

	got, err := uc.Execute("goal-001")
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].ID != "conv-001" {
		t.Errorf("Execute() = %+v, want %+v", got, want)
	}
}

func TestGetConversationUsecase_Execute(t *testing.T) {
	messages := &spyMessageRepo{existing: []*entity.ConversationMessage{
		entity.NewConversationMessage("msg-001", "conv-001", entity.ChatRoleUser, "5km走れるようになりたい", time.Now()),
	}}
	uc := NewGetConversationUsecase(messages)

	got, err := uc.Execute("conv-001")
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].ID != "msg-001" {
		t.Errorf("Execute() = %+v, want the single existing message", got)
	}
}
