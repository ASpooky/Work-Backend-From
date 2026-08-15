package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ASpooky/Work-Backend-From/src/usecase/dailytask"
)

type DailyTaskHandler struct {
	create *dailytask.CreateDailyTaskUsecase
	list   *dailytask.ListDailyTasksUsecase
}

func NewDailyTaskHandler(create *dailytask.CreateDailyTaskUsecase, list *dailytask.ListDailyTasksUsecase) *DailyTaskHandler {
	return &DailyTaskHandler{create: create, list: list}
}

type createDailyTaskRequest struct {
	GoalID  string `json:"goal_id"`
	Date    string `json:"date"`
	Content string `json:"content"`
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

	got, err := h.list.Execute(dailytask.ListDailyTasksInput{Date: date})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, got)
}
