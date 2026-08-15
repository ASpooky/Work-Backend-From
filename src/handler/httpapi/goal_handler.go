package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
	"github.com/ASpooky/Work-Backend-From/src/usecase/goal"
)

const dateLayout = "2006-01-02"

type GoalHandler struct {
	create *goal.CreateGoalUsecase
	list   *goal.ListGoalsUsecase
}

func NewGoalHandler(create *goal.CreateGoalUsecase, list *goal.ListGoalsUsecase) *GoalHandler {
	return &GoalHandler{create: create, list: list}
}

type createGoalRequest struct {
	WorkspaceID          string `json:"workspace_id"`
	Title                string `json:"title"`
	Detail               string `json:"detail"`
	AchievementCondition string `json:"achievement_condition"`
	EndDate              string `json:"end_date"`
	Mode                 string `json:"mode"`
}

func (h *GoalHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	endDate, err := time.Parse(dateLayout, req.EndDate)
	if err != nil {
		http.Error(w, "invalid end_date, expected YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	got, err := h.create.Execute(goal.CreateGoalInput{
		WorkspaceID:          req.WorkspaceID,
		Title:                req.Title,
		Detail:               req.Detail,
		AchievementCondition: req.AchievementCondition,
		EndDate:              endDate,
		Mode:                 entity.GoalMode(req.Mode),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, got)
}

func (h *GoalHandler) List(w http.ResponseWriter, r *http.Request) {
	workspaceID := r.URL.Query().Get("workspace_id")

	got, err := h.list.Execute(goal.ListGoalsInput{WorkspaceID: workspaceID})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, got)
}
