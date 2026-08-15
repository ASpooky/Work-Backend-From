package dailytask

type DoneUpdater interface {
	UpdateDone(id string, done bool) error
}

type UpdateDailyTaskDoneInput struct {
	ID   string
	Done bool
}

type UpdateDailyTaskDoneUsecase struct {
	repo DoneUpdater
}

func NewUpdateDailyTaskDoneUsecase(repo DoneUpdater) *UpdateDailyTaskDoneUsecase {
	return &UpdateDailyTaskDoneUsecase{repo: repo}
}

func (u *UpdateDailyTaskDoneUsecase) Execute(input UpdateDailyTaskDoneInput) error {
	return u.repo.UpdateDone(input.ID, input.Done)
}
