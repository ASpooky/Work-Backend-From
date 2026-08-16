package sqlite

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
)

func TestConversationRepository_SaveAndFindByWorkspaceID(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	db, err := Open(path)
	if err != nil {
		t.Fatalf("Open() returned unexpected error: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	workspaceRepo := NewWorkspaceRepository(db)
	workspace := entity.NewWorkSpace("workspace-001", DefaultUserID, "private", time.Now())
	if err := workspaceRepo.Save(workspace); err != nil {
		t.Fatalf("workspaceRepo.Save() returned unexpected error: %v", err)
	}

	repo := NewConversationRepository(db)
	older := entity.NewConversation("conv-001", workspace.ID, "", "5km走れるようになりたい", time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC))
	newer := entity.NewConversation("conv-002", workspace.ID, "", "マラソン完走したい", time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC))
	for _, c := range []*entity.Conversation{older, newer} {
		if err := repo.Save(c); err != nil {
			t.Fatalf("Save() returned unexpected error: %v", err)
		}
	}

	got, err := repo.FindByWorkspaceID(workspace.ID)
	if err != nil {
		t.Fatalf("FindByWorkspaceID() returned unexpected error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("FindByWorkspaceID() returned %d conversations, want 2", len(got))
	}
	// Most-recently-updated first, for a chat-UI sidebar.
	if got[0].ID != newer.ID || got[1].ID != older.ID {
		t.Errorf("FindByWorkspaceID() = [%s, %s], want [%s, %s] (newest updated_at first)", got[0].ID, got[1].ID, newer.ID, older.ID)
	}
	if got[0].Title != newer.Title {
		t.Errorf("got[0].Title = %q, want %q", got[0].Title, newer.Title)
	}
}

func TestConversationRepository_GoalScopedConversations(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	db, err := Open(path)
	if err != nil {
		t.Fatalf("Open() returned unexpected error: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	workspaceRepo := NewWorkspaceRepository(db)
	workspace := entity.NewWorkSpace("workspace-001", DefaultUserID, "private", time.Now())
	if err := workspaceRepo.Save(workspace); err != nil {
		t.Fatalf("workspaceRepo.Save() returned unexpected error: %v", err)
	}

	repo := NewConversationRepository(db)
	general := entity.NewConversation("conv-general", workspace.ID, "", "5km走れるようになりたい", time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC))
	review := entity.NewConversation("conv-review", workspace.ID, "goal-001", "目標の見直し", time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC))
	for _, c := range []*entity.Conversation{general, review} {
		if err := repo.Save(c); err != nil {
			t.Fatalf("Save() returned unexpected error: %v", err)
		}
	}

	byWorkspace, err := repo.FindByWorkspaceID(workspace.ID)
	if err != nil {
		t.Fatalf("FindByWorkspaceID() returned unexpected error: %v", err)
	}
	if len(byWorkspace) != 1 || byWorkspace[0].ID != general.ID {
		t.Errorf("FindByWorkspaceID() = %+v, want only the goal-unscoped conversation", byWorkspace)
	}

	byGoal, err := repo.FindByGoalID("goal-001")
	if err != nil {
		t.Fatalf("FindByGoalID() returned unexpected error: %v", err)
	}
	if len(byGoal) != 1 || byGoal[0].ID != review.ID {
		t.Errorf("FindByGoalID() = %+v, want only the goal-scoped conversation", byGoal)
	}
}

func TestConversationRepository_Touch(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	db, err := Open(path)
	if err != nil {
		t.Fatalf("Open() returned unexpected error: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	workspaceRepo := NewWorkspaceRepository(db)
	workspace := entity.NewWorkSpace("workspace-001", DefaultUserID, "private", time.Now())
	if err := workspaceRepo.Save(workspace); err != nil {
		t.Fatalf("workspaceRepo.Save() returned unexpected error: %v", err)
	}

	repo := NewConversationRepository(db)
	created := time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC)
	conv := entity.NewConversation("conv-001", workspace.ID, "", "5km走れるようになりたい", created)
	if err := repo.Save(conv); err != nil {
		t.Fatalf("Save() returned unexpected error: %v", err)
	}

	touchedAt := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)
	if err := repo.Touch(conv.ID, touchedAt); err != nil {
		t.Fatalf("Touch() returned unexpected error: %v", err)
	}

	got, err := repo.FindByWorkspaceID(workspace.ID)
	if err != nil {
		t.Fatalf("FindByWorkspaceID() returned unexpected error: %v", err)
	}
	if len(got) != 1 || !got[0].UpdatedAt.Equal(touchedAt) {
		t.Fatalf("FindByWorkspaceID() after Touch() = %+v, want UpdatedAt=%v", got, touchedAt)
	}
	if !got[0].CreatedAt.Equal(created) {
		t.Errorf("CreatedAt = %v, want unchanged %v", got[0].CreatedAt, created)
	}
}

func TestConversationMessageRepository_SaveAndFindByConversationID(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	db, err := Open(path)
	if err != nil {
		t.Fatalf("Open() returned unexpected error: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	workspaceRepo := NewWorkspaceRepository(db)
	workspace := entity.NewWorkSpace("workspace-001", DefaultUserID, "private", time.Now())
	if err := workspaceRepo.Save(workspace); err != nil {
		t.Fatalf("workspaceRepo.Save() returned unexpected error: %v", err)
	}

	convRepo := NewConversationRepository(db)
	conv := entity.NewConversation("conv-001", workspace.ID, "", "5km走れるようになりたい", time.Now())
	if err := convRepo.Save(conv); err != nil {
		t.Fatalf("convRepo.Save() returned unexpected error: %v", err)
	}
	otherConv := entity.NewConversation("conv-002", workspace.ID, "", "別の会話", time.Now())
	if err := convRepo.Save(otherConv); err != nil {
		t.Fatalf("convRepo.Save() returned unexpected error: %v", err)
	}

	repo := NewConversationMessageRepository(db)
	first := entity.NewConversationMessage("msg-001", conv.ID, entity.ChatRoleUser, "5km走れるようになりたい", time.Date(2026, 8, 16, 9, 0, 0, 0, time.UTC))
	second := entity.NewConversationMessage("msg-002", conv.ID, entity.ChatRoleModel, "いつまでに達成したいですか？", time.Date(2026, 8, 16, 9, 0, 1, 0, time.UTC))
	other := entity.NewConversationMessage("msg-003", otherConv.ID, entity.ChatRoleUser, "別の会話のメッセージ", time.Now())
	for _, m := range []*entity.ConversationMessage{first, second, other} {
		if err := repo.Save(m); err != nil {
			t.Fatalf("Save() returned unexpected error: %v", err)
		}
	}

	got, err := repo.FindByConversationID(conv.ID)
	if err != nil {
		t.Fatalf("FindByConversationID() returned unexpected error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("FindByConversationID() returned %d messages, want 2", len(got))
	}
	if got[0].ID != first.ID || got[1].ID != second.ID {
		t.Errorf("FindByConversationID() = [%s, %s], want [%s, %s] in created_at order", got[0].ID, got[1].ID, first.ID, second.ID)
	}
	if got[0].Role != entity.ChatRoleUser || got[1].Role != entity.ChatRoleModel {
		t.Errorf("roles = [%s, %s], want [%s, %s]", got[0].Role, got[1].Role, entity.ChatRoleUser, entity.ChatRoleModel)
	}
}
