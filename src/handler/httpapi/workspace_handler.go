package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/ASpooky/Work-Backend-From/src/usecase/workspace"
)

type WorkspaceHandler struct {
	create *workspace.CreateWorkspaceUsecase
	list   *workspace.ListWorkspacesUsecase
	rename *workspace.RenameWorkspaceUsecase
	delete *workspace.DeleteWorkspaceUsecase
}

func NewWorkspaceHandler(
	create *workspace.CreateWorkspaceUsecase,
	list *workspace.ListWorkspacesUsecase,
	rename *workspace.RenameWorkspaceUsecase,
	del *workspace.DeleteWorkspaceUsecase,
) *WorkspaceHandler {
	return &WorkspaceHandler{create: create, list: list, rename: rename, delete: del}
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

type renameWorkspaceRequest struct {
	Name string `json:"name"`
}

func (h *WorkspaceHandler) Rename(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req renameWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.rename.Execute(workspace.RenameWorkspaceInput{ID: id, Name: req.Name}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"id": id, "name": req.Name})
}

func (h *WorkspaceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.delete.Execute(id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
