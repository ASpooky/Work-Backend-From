package goal

type DeleteRepository interface {
	Delete(id string) error
}

type DeleteGoalUsecase struct {
	repo DeleteRepository
}

func NewDeleteGoalUsecase(repo DeleteRepository) *DeleteGoalUsecase {
	return &DeleteGoalUsecase{repo: repo}
}

func (u *DeleteGoalUsecase) Execute(id string) error {
	return u.repo.Delete(id)
}
