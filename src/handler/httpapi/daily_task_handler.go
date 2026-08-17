package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ASpooky/Work-Backend-From/src/usecase/dailytask"
)

type DailyTaskHandler struct {
	create          *dailytask.CreateDailyTaskUsecase
	createRecurring *dailytask.CreateRecurringDailyTasksUsecase
	list            *dailytask.ListDailyTasksUsecase
	updateDone      *dailytask.UpdateDailyTaskDoneUsecase
	updateMemo      *dailytask.UpdateTaskMemoUsecase
	delete          *dailytask.DeleteDailyTaskUsecase
}

func NewDailyTaskHandler(
	create *dailytask.CreateDailyTaskUsecase,
	createRecurring *dailytask.CreateRecurringDailyTasksUsecase,
	list *dailytask.ListDailyTasksUsecase,
	updateDone *dailytask.UpdateDailyTaskDoneUsecase,
	updateMemo *dailytask.UpdateTaskMemoUsecase,
	del *dailytask.DeleteDailyTaskUsecase,
) *DailyTaskHandler {
	return &DailyTaskHandler{create: create, createRecurring: createRecurring, list: list, updateDone: updateDone, updateMemo: updateMemo, delete: del}
}

type updateDailyTaskDoneRequest struct {
	Done bool `json:"done"`
}

type updateTaskMemoRequest struct {
	Memo *string `json:"memo"`
}

type createDailyTaskRequest struct {
	GoalID  string `json:"goal_id"`
	Date    string `json:"date"`
	Content string `json:"content"`
}

type recurrenceRuleRequest struct {
	Type         string `json:"type"`
	IntervalDays int    `json:"interval_days"`
	Weekdays     []int  `json:"weekdays"`
}

type createRecurringDailyTasksRequest struct {
	GoalID    string                `json:"goal_id"`
	Content   string                `json:"content"`
	StartDate string                `json:"start_date"`
	EndDate   string                `json:"end_date"`
	Rule      recurrenceRuleRequest `json:"rule"`
}

func (h *DailyTaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createDailyTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	date, err := time.Parse(dateLayout, req.Date)
	if err != nil {
		http.Error(w, "invalid date, expected YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	got, err := h.create.Execute(dailytask.CreateDailyTaskInput{
		GoalID:  req.GoalID,
		Date:    date,
		Content: req.Content,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, got)
}

func (h *DailyTaskHandler) CreateRecurring(w http.ResponseWriter, r *http.Request) {
	var req createRecurringDailyTasksRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	startDate, err := time.Parse(dateLayout, req.StartDate)
	if err != nil {
		http.Error(w, "invalid start_date, expected YYYY-MM-DD", http.StatusBadRequest)
		return
	}
	endDate, err := time.Parse(dateLayout, req.EndDate)
	if err != nil {
		http.Error(w, "invalid end_date, expected YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	rule := dailytask.RecurrenceRule{
		Type:         dailytask.RecurrenceType(req.Rule.Type),
		IntervalDays: req.Rule.IntervalDays,
	}
	for _, wd := range req.Rule.Weekdays {
		rule.Weekdays = append(rule.Weekdays, time.Weekday(wd))
	}

	got, err := h.createRecurring.Execute(dailytask.CreateRecurringDailyTasksInput{
		GoalID:    req.GoalID,
		Content:   req.Content,
		StartDate: startDate,
		EndDate:   endDate,
		Rule:      rule,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, got)
}

func (h *DailyTaskHandler) UpdateDone(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req updateDailyTaskDoneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.updateDone.Execute(dailytask.UpdateDailyTaskDoneInput{ID: id, Done: req.Done}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"id": id, "done": req.Done})
}

func (h *DailyTaskHandler) UpdateMemo(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req updateTaskMemoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.updateMemo.Execute(dailytask.UpdateTaskMemoInput{ID: id, Memo: req.Memo}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"id": id, "memo": req.Memo})
}

func (h *DailyTaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.delete.Execute(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *DailyTaskHandler) List(w http.ResponseWriter, r *http.Request) {
	dateParam := r.URL.Query().Get("date")

	date := time.Now()
	if dateParam != "" {
		parsed, err := time.Parse(dateLayout, dateParam)
		if err != nil {
			http.Error(w, "invalid date, expected YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		date = parsed
	}

	got, err := h.list.Execute(dailytask.ListDailyTasksInput{
		Date:        date,
		WorkspaceID: r.URL.Query().Get("workspace_id"),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, got)
}
