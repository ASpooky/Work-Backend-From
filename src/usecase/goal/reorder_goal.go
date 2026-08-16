package goal

// PriorityUpdater is satisfied by GoalRepository.UpdatePriority.
type PriorityUpdater interface {
	UpdatePriority(id string, priority int) error
}

type ReorderGoalInput struct {
	ID       string
	Priority int
}

// ReorderGoalUsecase sets a goal's manual sort position, separate from
// UpdateGoalUsecase since reordering is a list-display concern, not part of
// the "目標の見直し" edit flow.
type ReorderGoalUsecase struct {
	repo PriorityUpdater
}

func NewReorderGoalUsecase(repo PriorityUpdater) *ReorderGoalUsecase {
	return &ReorderGoalUsecase{repo: repo}
}

func (u *ReorderGoalUsecase) Execute(input ReorderGoalInput) error {
	return u.repo.UpdatePriority(input.ID, input.Priority)
}
