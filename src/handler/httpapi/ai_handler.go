package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/ASpooky/Work-Backend-From/src/entity"
	"github.com/ASpooky/Work-Backend-From/src/usecase/ai"
)

type AIHandler struct {
	chat *ai.ChatUsecase
	plan *ai.PlanGoalUsecase
}

func NewAIHandler(chat *ai.ChatUsecase, plan *ai.PlanGoalUsecase) *AIHandler {
	return &AIHandler{chat: chat, plan: plan}
}

type chatMessageRequest struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Messages []chatMessageRequest `json:"messages"`
}

func toChatMessages(req []chatMessageRequest) []entity.ChatMessage {
	messages := make([]entity.ChatMessage, len(req))
	for i, m := range req {
		messages[i] = entity.ChatMessage{Role: entity.ChatRole(m.Role), Content: m.Content}
	}
	return messages
}

func (h *AIHandler) Chat(w http.ResponseWriter, r *http.Request) {
	var req chatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	reply, err := h.chat.Execute(r.Context(), ai.ChatInput{Messages: toChatMessages(req.Messages)})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"reply": reply})
}

type plannedGoalResponse struct {
	Title                string `json:"title"`
	Detail               string `json:"detail"`
	AchievementCondition string `json:"achievement_condition"`
	EndDate              string `json:"end_date"`
	Mode                 string `json:"mode"`
}

type plannedTaskResponse struct {
	Content      string `json:"content"`
	StartDate    string `json:"start_date"`
	EndDate      string `json:"end_date"`
	RuleType     string `json:"rule_type"`
	IntervalDays int    `json:"interval_days,omitempty"`
	Weekdays     []int  `json:"weekdays,omitempty"`
}

type planResponse struct {
	Goal  plannedGoalResponse   `json:"goal"`
	Tasks []plannedTaskResponse `json:"tasks"`
}

func (h *AIHandler) Plan(w http.ResponseWriter, r *http.Request) {
	var req chatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	plan, err := h.plan.Execute(r.Context(), ai.PlanGoalInput{Messages: toChatMessages(req.Messages)})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := planResponse{
		Goal: plannedGoalResponse{
			Title:                plan.Goal.Title,
			Detail:               plan.Goal.Detail,
			AchievementCondition: plan.Goal.AchievementCondition,
			EndDate:              plan.Goal.EndDate.Format(dateLayout),
			Mode:                 string(plan.Goal.Mode),
		},
		Tasks: []plannedTaskResponse{},
	}
	for _, t := range plan.Tasks {
		tr := plannedTaskResponse{
			Content:      t.Content,
			StartDate:    t.StartDate.Format(dateLayout),
			EndDate:      t.EndDate.Format(dateLayout),
			RuleType:     string(t.Rule.Type),
			IntervalDays: t.Rule.IntervalDays,
		}
		for _, wd := range t.Rule.Weekdays {
			tr.Weekdays = append(tr.Weekdays, int(wd))
		}
		resp.Tasks = append(resp.Tasks, tr)
	}

	writeJSON(w, http.StatusOK, resp)
}
