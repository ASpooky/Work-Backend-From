package dailytask

// MemoUpdater is satisfied by DailyTaskRepository.UpdateMemo.
type MemoUpdater interface {
	UpdateMemo(id string, memo *string) error
}

type UpdateTaskMemoInput struct {
	ID   string
	Memo *string // nil clears the memo
}

// UpdateTaskMemoUsecase is separate from UpdateDailyTaskDoneUsecase since a
// memo is editable at any time, independent of Done/CompletedAt.
type UpdateTaskMemoUsecase struct {
	repo MemoUpdater
}

func NewUpdateTaskMemoUsecase(repo MemoUpdater) *UpdateTaskMemoUsecase {
	return &UpdateTaskMemoUsecase{repo: repo}
}

func (u *UpdateTaskMemoUsecase) Execute(input UpdateTaskMemoInput) error {
	return u.repo.UpdateMemo(input.ID, input.Memo)
}
