package ai

import (
	"context"
	"testing"
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
)

type stubIDGenerator struct {
	ids []string
	i   int
}

func (s *stubIDGenerator) NewID() string {
	id := s.ids[s.i]
	s.i++
	return id
}

type spyConversationRepo struct {
	saved     []*entity.Conversation
	touchedID string
	touchedAt time.Time
}

func (s *spyConversationRepo) Save(c *entity.Conversation) error {
	s.saved = append(s.saved, c)
	return nil
}

func (s *spyConversationRepo) Touch(id string, updatedAt time.Time) error {
	s.touchedID = id
	s.touchedAt = updatedAt
	return nil
}

type spyMessageRepo struct {
	saved    []*entity.ConversationMessage
	existing []*entity.ConversationMessage
}

func (s *spyMessageRepo) Save(m *entity.ConversationMessage) error {
	s.saved = append(s.saved, m)
	return nil
}

func (s *spyMessageRepo) FindByConversationID(conversationID string) ([]*entity.ConversationMessage, error) {
	return s.existing, nil
}

func TestSendMessageUsecase_Execute_NewConversation(t *testing.T) {
	conversations := &spyConversationRepo{}
	messages := &spyMessageRepo{}
	completer := &stubChatCompleter{reply: "いつまでに達成したいですか？"}
	fixedNow := time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC)
	chat := NewChatUsecase(completer, stubClock{now: fixedNow})
	ids := &stubIDGenerator{ids: []string{"conv-001", "msg-user-001", "msg-model-001"}}

	uc := NewSendMessageUsecase(conversations, conversations, messages, chat, ids, stubClock{now: fixedNow})

	got, err := uc.Execute(context.Background(), SendMessageInput{
		WorkspaceID: "workspace-001",
		Content:     "5km走れるようになりたい",
	})
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}

	if got.ConversationID != "conv-001" {
		t.Errorf("ConversationID = %q, want %q", got.ConversationID, "conv-001")
	}
	if got.Reply != "いつまでに達成したいですか？" {
		t.Errorf("Reply = %q, want the stub reply", got.Reply)
	}

	if len(conversations.saved) != 1 {
		t.Fatalf("conversation saved %d times, want 1", len(conversations.saved))
	}
	if conversations.saved[0].Title == "" {
		t.Errorf("new conversation has an empty title, want it derived from the first message")
	}

	if len(messages.saved) != 2 {
		t.Fatalf("messages saved %d times, want 2 (user + model)", len(messages.saved))
	}
	if messages.saved[0].Role != entity.ChatRoleUser || messages.saved[0].Content != "5km走れるようになりたい" {
		t.Errorf("first saved message = %+v, want the user's content", messages.saved[0])
	}
	if messages.saved[1].Role != entity.ChatRoleModel || messages.saved[1].Content != "いつまでに達成したいですか？" {
		t.Errorf("second saved message = %+v, want the AI reply", messages.saved[1])
	}

	if conversations.touchedID != "conv-001" {
		t.Errorf("Touch() called with id = %q, want %q", conversations.touchedID, "conv-001")
	}
}

func TestSendMessageUsecase_Execute_ExistingConversation(t *testing.T) {
	existing := []*entity.ConversationMessage{
		entity.NewConversationMessage("msg-001", "conv-existing", entity.ChatRoleUser, "5km走れるようになりたい", time.Now()),
		entity.NewConversationMessage("msg-002", "conv-existing", entity.ChatRoleModel, "いつまでに達成したいですか？", time.Now()),
	}
	conversations := &spyConversationRepo{}
	messages := &spyMessageRepo{existing: existing}
	completer := &stubChatCompleter{reply: "3ヶ月後で必達ですね、承知しました。"}
	fixedNow := time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC)
	chat := NewChatUsecase(completer, stubClock{now: fixedNow})
	ids := &stubIDGenerator{ids: []string{"msg-user-002", "msg-model-002"}}

	uc := NewSendMessageUsecase(conversations, conversations, messages, chat, ids, stubClock{now: fixedNow})

	got, err := uc.Execute(context.Background(), SendMessageInput{
		WorkspaceID:    "workspace-001",
		ConversationID: "conv-existing",
		Content:        "3ヶ月後、必達で",
	})
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}

	if got.ConversationID != "conv-existing" {
		t.Errorf("ConversationID = %q, want %q (existing conversation, no new one created)", got.ConversationID, "conv-existing")
	}
	if len(conversations.saved) != 0 {
		t.Errorf("conversation Save() called %d times for an existing conversation, want 0", len(conversations.saved))
	}

	// The AI should see the full history plus the new message.
	if len(completer.capturedMessages) != 3 {
		t.Fatalf("Chat() received %d messages, want 3 (2 existing + 1 new)", len(completer.capturedMessages))
	}
	if completer.capturedMessages[2].Content != "3ヶ月後、必達で" {
		t.Errorf("last message sent to the AI = %q, want the new user content", completer.capturedMessages[2].Content)
	}
}
