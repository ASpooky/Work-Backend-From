package dailytask

// DeleteRepository is satisfied by DailyTaskRepository.Delete.
type DeleteRepository interface {
	Delete(id string) error
}

type DeleteDailyTaskUsecase struct {
	repo DeleteRepository
}

func NewDeleteDailyTaskUsecase(repo DeleteRepository) *DeleteDailyTaskUsecase {
	return &DeleteDailyTaskUsecase{repo: repo}
}

func (u *DeleteDailyTaskUsecase) Execute(id string) error {
	return u.repo.Delete(id)
}
