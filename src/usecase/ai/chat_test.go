package ai

import (
	"context"
	"strings"
	"testing"
	"time"

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

type stubClock struct {
	now time.Time
}

func (s stubClock) Now() time.Time {
	return s.now
}

func TestChatUsecase_Execute(t *testing.T) {
	completer := &stubChatCompleter{reply: "いいですね、期限はいつにしますか？"}
	fixedNow := time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC)
	uc := NewChatUsecase(completer, stubClock{now: fixedNow})

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
	if !strings.Contains(completer.capturedSystem, "2026-08-16") {
		t.Errorf("Chat() system instruction = %q, want it to mention today's date so the model can resolve relative dates like 3ヶ月後", completer.capturedSystem)
	}
	if len(completer.capturedMessages) != 1 || completer.capturedMessages[0].Content != messages[0].Content {
		t.Errorf("Chat() was called with messages = %+v, want %+v", completer.capturedMessages, messages)
	}
}
