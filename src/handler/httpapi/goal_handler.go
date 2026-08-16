package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ASpooky/Work-Backend-From/src/entity"
	"github.com/ASpooky/Work-Backend-From/src/usecase"
	"github.com/ASpooky/Work-Backend-From/src/usecase/goal"
)

const dateLayout = "2006-01-02"

type GoalHandler struct {
	create  *goal.CreateGoalUsecase
	list    *goal.ListGoalsUsecase
	get     *usecase.GetGoalStatsUsecase
	update  *goal.UpdateGoalUsecase
	delete  *goal.DeleteGoalUsecase
	reorder *goal.ReorderGoalUsecase
}

func NewGoalHandler(
	create *goal.CreateGoalUsecase,
	list *goal.ListGoalsUsecase,
	get *usecase.GetGoalStatsUsecase,
	update *goal.UpdateGoalUsecase,
	del *goal.DeleteGoalUsecase,
	reorder *goal.ReorderGoalUsecase,
) *GoalHandler {
	return &GoalHandler{create: create, list: list, get: get, update: update, delete: del, reorder: reorder}
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

type goalStatsResponse struct {
	Goal            *entity.Goal `json:"goal"`
	ScheduledCount  int          `json:"scheduled_count"`
	DoneCount       int          `json:"done_count"`
	AchievementRate float64      `json:"achievement_rate"`
	PostponeCount   int          `json:"postpone_count"`
	DaysRemaining   int          `json:"days_remaining"`
}

func (h *GoalHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	got, stats, err := h.get.Execute(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if got == nil {
		http.Error(w, "goal not found", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, goalStatsResponse{
		Goal:            got,
		ScheduledCount:  stats.ScheduledCount,
		DoneCount:       stats.DoneCount,
		AchievementRate: stats.AchievementRate,
		PostponeCount:   stats.PostponeCount,
		DaysRemaining:   stats.DaysRemaining,
	})
}

type updateGoalRequest struct {
	Title                string `json:"title"`
	Detail               string `json:"detail"`
	AchievementCondition string `json:"achievement_condition"`
	EndDate              string `json:"end_date"`
	Mode                 string `json:"mode"`
}

func (h *GoalHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req updateGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	endDate, err := time.Parse(dateLayout, req.EndDate)
	if err != nil {
		http.Error(w, "invalid end_date, expected YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	if err := h.update.Execute(goal.UpdateGoalInput{
		ID:                   id,
		Title:                req.Title,
		Detail:               req.Detail,
		AchievementCondition: req.AchievementCondition,
		EndDate:              endDate,
		Mode:                 entity.GoalMode(req.Mode),
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"id": id})
}

func (h *GoalHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.delete.Execute(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type reorderGoalRequest struct {
	Priority int `json:"priority"`
}

func (h *GoalHandler) Reorder(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req reorderGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.reorder.Execute(goal.ReorderGoalInput{ID: id, Priority: req.Priority}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"id": id, "priority": req.Priority})
}
