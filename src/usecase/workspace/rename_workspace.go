package workspace

type NameUpdater interface {
	UpdateName(id, name string) error
}

type RenameWorkspaceInput struct {
	ID   string
	Name string
}

type RenameWorkspaceUsecase struct {
	repo NameUpdater
}

func NewRenameWorkspaceUsecase(repo NameUpdater) *RenameWorkspaceUsecase {
	return &RenameWorkspaceUsecase{repo: repo}
}

func (u *RenameWorkspaceUsecase) Execute(input RenameWorkspaceInput) error {
	return u.repo.UpdateName(input.ID, input.Name)
}
