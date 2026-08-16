package ai

import (
	"context"
	"testing"

	"github.com/ASpooky/Work-Backend-From/src/entity"
)

type stubChatCompleter struct {
	capturedSystem   string
	capturedMessages []entity.ChatMessage
	reply            string
}

func (s *stubChatCompleter) Chat(ctx context.Context, systemInstruction string, messages []entity.ChatMessage) (string, error) {
	s.capturedSystem = systemInstruction
	s.capturedMessages = messages
	return s.reply, nil
}

func TestChatUsecase_Execute(t *testing.T) {
	completer := &stubChatCompleter{reply: "いいですね、期限はいつにしますか？"}
	uc := NewChatUsecase(completer)

	messages := []entity.ChatMessage{{Role: entity.ChatRoleUser, Content: "5km走れるようになりたい"}}
	got, err := uc.Execute(context.Background(), ChatInput{Messages: messages})
	if err != nil {
		t.Fatalf("Execute() returned unexpected error: %v", err)
	}
	if got != "いいですね、期限はいつにしますか？" {
		t.Errorf("Execute() = %q, want the stub reply", got)
	}
	if completer.capturedSystem == "" {
		t.Errorf("Chat() was called with an empty system instruction, want the coaching prompt")
	}
	if len(completer.capturedMessages) != 1 || completer.capturedMessages[0].Content != messages[0].Content {
		t.Errorf("Chat() was called with messages = %+v, want %+v", completer.capturedMessages, messages)
	}
}
