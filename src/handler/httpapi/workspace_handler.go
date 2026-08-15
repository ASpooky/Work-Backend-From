package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/ASpooky/Work-Backend-From/src/usecase/workspace"
)

type WorkspaceHandler struct {
	create *workspace.CreateWorkspaceUsecase
	list   *workspace.ListWorkspacesUsecase
}

func NewWorkspaceHandler(create *workspace.CreateWorkspaceUsecase, list *workspace.ListWorkspacesUsecase) *WorkspaceHandler {
	return &WorkspaceHandler{create: create, list: list}
}

type createWorkspaceRequest struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
}

func (h *WorkspaceHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	got, err := h.create.Execute(workspace.CreateWorkspaceInput{
		UserID: req.UserID,
		Name:   req.Name,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, got)
}

func (h *WorkspaceHandler) List(w http.ResponseWriter, r *http.Request) {
	got, err := h.list.Execute()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, got)
}
